package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"kraken-trader/internal/api"
	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
	"kraken-trader/internal/messaging"
	"kraken-trader/internal/news"
	"kraken-trader/internal/repository"
	"kraken-trader/internal/state"
	"kraken-trader/internal/storage"
	"kraken-trader/pkg/config"
	"kraken-trader/pkg/kraken"
)

func runRun(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	log.Info().Msg("Starting Kraken Trader...")
	log.Info().Str("mode", cfg.TradingMode).Msg("Configuration loaded")

	// ── Context with Graceful Shutdown ──────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// ── Initialize Components ──────────────────────────────────────

	// Kraken CLI
	krakenClient := kraken.NewClient("")
	log.Info().Msg("Kraken CLI client ready")

	// InfluxDB Storage
	var dbClient *storage.Client
	if cfg.InfluxDBURL != "" {
		dbClient, err = storage.NewClient(
			cfg.InfluxDBURL,
			cfg.InfluxDBToken,
			cfg.InfluxDBOrg,
			cfg.InfluxDBBucket,
		)
		if err != nil {
			log.Warn().Err(err).Msg("InfluxDB not available")
		} else {
			defer dbClient.Close()
			log.Info().Msg("InfluxDB connected")
		}
	}

	// NATS Messaging
	natsClient, err := messaging.NewNATSClient(cfg.NATSURL)
	if err != nil {
		log.Warn().Err(err).Msg("NATS not available")
		natsClient = nil
	} else {
		defer natsClient.Close()
		// Create JetStream stream for persisted trade data
		if err := natsClient.EnsureStream(ctx, "TRADING", []string{
			"trade.decision.*",
			"trade.execution.*",
		}, jetstream.FileStorage); err != nil {
			log.Warn().Err(err).Msg("Could not create JetStream stream")
		}
	}

	// In-Memory State Manager
	stateMgr := state.NewMemoryManager()
	log.Info().Msg("Memory manager initialized")

	// PRISM API Client
	prismClient := news.NewPrismClient(cfg.PrismAPIKey)
	log.Info().Msg("PRISM client ready")

	// ChromaDB + Embedder
	chromaClient := news.NewChromaClient(cfg.ChromaURL)
	if err := chromaClient.Initialize(ctx, cfg.ChromaCollection); err != nil {
		log.Warn().Err(err).Msg("ChromaDB not available")
		chromaClient = nil
	} else {
		log.Info().Msg("ChromaDB connected")
	}

	embedder := news.NewEmbedder(cfg.OllamaURL, cfg.OllamaEmbedModel)
	log.Info().Msg("Embedder ready")

	// SQLite Repository for prompts
	promptRepo, err := repository.NewSQLiteRepository(cfg.SQLitePath)
	if err != nil {
		log.Warn().Err(err).Msg("SQLite not available, prompts will not be stored")
		promptRepo = nil
	}

	// Decision Engine
	engine := decision.NewEngine(
		cfg.OllamaURL,
		cfg.OllamaModel,
		stateMgr,
		chromaClient,
		embedder,
		promptRepo,
	)
	log.Info().Msg("Decision engine initialized")

	// ── Start Data Pipelines ───────────────────────────────────────

	// Market Data Collector (WebSocket via kraken CLI)
	collector := market.NewCollector(krakenClient, dbClient, stateMgr, natsClient, promptRepo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		collector.Start(ctx)
	}()

	// HTTP API Server for subscription management
	router := gin.Default()
	apiServer := api.NewServer(collector, engine, krakenClient)
	api.RegisterHandlersWithOptions(router, apiServer, api.GinServerOptions{})

	// Serve embedded OpenAPI spec and Swagger UI
	router.GET("/openapi.json", func(c *gin.Context) {
		swagger, err := api.GetSwagger()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get swagger"})
			return
		}
		c.JSON(http.StatusOK, swagger)
	})
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})
	router.GET("/swagger/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
    <title>Kraken Trader API</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: "#swagger-ui",
                deepLinking: true
            });
        };
    </script>
</body>
</html>`)
	})

	apiAddr := fmt.Sprintf(":%d", cfg.APIPort)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Int("port", cfg.APIPort).Msg("Starting Market API server")
		if err := http.ListenAndServe(apiAddr, router); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("API server error")
		}
	}()

	// PRISM Signal Polling (every 60s)
	wg.Add(1)
	go func() {
		defer wg.Done()
		pollSignals(ctx, prismClient, stateMgr, promptRepo)
	}()

	// Balance Polling (every 30s)
	wg.Add(1)
	go func() {
		defer wg.Done()
		pollBalance(ctx, krakenClient, stateMgr)
	}()

	// News Ingestion + Embedding (every 120s)
	wg.Add(1)
	go func() {
		defer wg.Done()
		pollNews(ctx, prismClient, chromaClient, embedder, stateMgr)
	}()

	// NATS JetStream Consumer for trade decisions (persist to InfluxDB)
	if natsClient != nil && dbClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			persistDecisions(ctx, natsClient, dbClient)
		}()
	}

	// Decision Loop — triggered by price alerts
	wg.Add(1)
	go func() {
		defer wg.Done()
		runDecisionLoop(ctx, stateMgr, engine, natsClient, cfg.TradingMode)
	}()

	// ── Block until Shutdown Signal ────────────────────────────────
	log.Info().Msg("Kraken Trader running. Press Ctrl+C to stop.")
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("Received signal, shutting down...")
	cancel()

	// Wait for goroutines with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("All goroutines stopped cleanly")
	case <-time.After(10 * time.Second):
		log.Warn().Msg("Shutdown timeout — exiting")
	}

	return nil
}

// pollSignals periodically fetches PRISM technical signals and stores them in memory
func pollSignals(ctx context.Context, prism *news.PrismClient, stateMgr *state.MemoryManager, repo repository.Repository) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial fetch
	fetchAndStoreSignals(ctx, prism, stateMgr, repo)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fetchAndStoreSignals(ctx, prism, stateMgr, repo)
		}
	}
}

func fetchAndStoreSignals(ctx context.Context, prism *news.PrismClient, stateMgr *state.MemoryManager, repo repository.Repository) {
	subs, err := repo.GetActiveSubscriptions(ctx)
	if err != nil || len(subs) == 0 {
		log.Debug().Err(err).Msg("No subscriptions found for signal polling")
		return
	}

	pairs := make([]string, len(subs))
	for i, s := range subs {
		pairs[i] = s.Symbol
	}
	// Extract base symbols by trimming common quote currencies
	symbols := make([]string, len(pairs))
	for i, p := range pairs {
		// Trim common quote currencies (USD, USDT, EUR, etc.)
		symbol := strings.TrimSuffix(p, "USD")
		symbol = strings.TrimSuffix(symbol, "USDT")
		symbol = strings.TrimSuffix(symbol, "EUR")
		symbol = strings.TrimSuffix(symbol, "BTC")
		if symbol == "" {
			// Fallback if no suffix matched
			symbol = p
		}
		symbols[i] = symbol
	}

	summaries, err := prism.FetchSignalSummary(ctx, symbols)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch PRISM signals")
		return
	}

	log.Trace().Interface("signals", summaries).Msg("PRISM signals response")

	for _, s := range summaries {
		stateMgr.UpdateSignal(s.Symbol, s.OverallSignal, s.Direction, s.Strength)
	}
	log.Info().Int("symbols", len(summaries)).Msg("Updated PRISM signals")
}

// pollNews periodically fetches crypto news, embeds it, and stores in ChromaDB
func pollNews(ctx context.Context, prism *news.PrismClient, chroma *news.ChromaClient, embedder *news.Embedder, stateMgr *state.MemoryManager) {
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()

	fetchAndStoreNews(ctx, prism, chroma, embedder, stateMgr)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fetchAndStoreNews(ctx, prism, chroma, embedder, stateMgr)
		}
	}
}

func fetchAndStoreNews(ctx context.Context, prism *news.PrismClient, chroma *news.ChromaClient, embedder *news.Embedder, stateMgr *state.MemoryManager) {
	articles, err := prism.FetchCryptoNews(ctx, 10, 30)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch PRISM news")
		return
	}

	log.Trace().Interface("articles", articles).Int("count", len(articles)).Msg("PRISM news response")

	if len(articles) == 0 {
		return
	}

	// Update in-memory news for the LLM prompt
	memArticles := make([]state.NewsArticle, len(articles))
	for i, a := range articles {
		memArticles[i] = state.NewsArticle{
			ID:        a.ID,
			Title:     a.Title,
			Summary:   a.Summary,
			Source:    a.Source,
			Timestamp: a.Timestamp,
		}
	}
	stateMgr.UpdateNews(memArticles)

	// Embed and store in ChromaDB for semantic search
	if chroma != nil && embedder != nil {
		embeddings, err := embedder.BatchEmbed(ctx, articles)
		if err != nil {
			log.Error().Err(err).Msg("Failed to embed articles")
			return
		}
		if err := chroma.AddEmbeddings(ctx, articles, embeddings); err != nil {
			log.Error().Err(err).Msg("Failed to store embeddings in ChromaDB")
		}
		log.Info().Int("articles", len(articles)).Msg("Stored news articles in ChromaDB")
	}
}

// persistDecisions consumes JetStream trade decisions and writes them to InfluxDB
func persistDecisions(ctx context.Context, natsClient *messaging.NATSClient, db *storage.Client) {
	err := natsClient.ConsumePersisted(ctx, "TRADING", "influxdb-writer", func(msg jetstream.Msg) {
		var trade storage.TradePoint
		if err := json.Unmarshal(msg.Data(), &trade); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal trade decision")
			if err := msg.Term(); err != nil {
				log.Error().Err(err).Msg("Failed to terminate message")
			}
			return
		}
		measurement, tags, fields, ts := trade.ToPointData()
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := db.WritePoint(dbCtx, measurement, tags, fields, ts); err != nil {
			log.Error().Err(err).Msg("Failed to write trade to InfluxDB")
			if err := msg.Nak(); err != nil {
				log.Error().Err(err).Msg("Failed to nak message")
			}
			return
		}
		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Msg("Failed to ack message")
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("JetStream consumer error")
	}
}

// runDecisionLoop listens for price alerts and triggers the LLM decision engine
func runDecisionLoop(ctx context.Context, stateMgr *state.MemoryManager, engine *decision.Engine, natsClient *messaging.NATSClient, mode string) {
	log.Info().Msg("Decision loop started — waiting for price alerts")

	for {
		select {
		case <-ctx.Done():
			return
		case pair := <-stateMgr.PriceAlertCh:
			log.Info().Str("pair", pair).Msg("Price alert triggered — running decision engine")

			decisions, err := engine.Decide(ctx, []string{pair})
			if err != nil {
				log.Error().Err(err).Msg("Decision engine error")
				continue
			}

			for _, d := range decisions {
				log.Info().
					Str("mode", mode).
					Str("action", d.Action).
					Str("pair", d.Pair).
					Float64("confidence", d.Confidence).
					Str("reasoning", d.Reasoning).
					Msg("Trade decision")

				// Publish decision to NATS JetStream
				if natsClient != nil {
					data, _ := json.Marshal(d)
					pubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					if err := natsClient.PublishPersisted(pubCtx, "trade.decision."+d.Pair, data); err != nil {
						log.Error().Err(err).Msg("Failed to publish decision to NATS")
					}
					cancel()
				}
			}
		}
	}
}

// pollBalance periodically fetches account balance from Kraken and stores in memory
func pollBalance(ctx context.Context, krakenClient *kraken.Client, stateMgr *state.MemoryManager) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	fetchAndStoreBalance(ctx, krakenClient, stateMgr)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fetchAndStoreBalance(ctx, krakenClient, stateMgr)
		}
	}
}

func fetchAndStoreBalance(ctx context.Context, krakenClient *kraken.Client, stateMgr *state.MemoryManager) {
	balances, err := krakenClient.GetBalance(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to fetch balance")
		return
	}

	stateMgr.UpdateBalance(balances)
	log.Info().Int("assets", len(balances)).Msg("Updated account balances")
}
