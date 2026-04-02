package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
	"kraken-trader/internal/messaging"
	"kraken-trader/internal/news"
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
	log.Info().Str("mode", cfg.TradingMode).Strs("pairs", cfg.TradePairs).Msg("Configuration loaded")

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
		}); err != nil {
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

	// Decision Engine
	engine := decision.NewEngine(
		cfg.OllamaURL,
		cfg.OllamaModel,
		stateMgr,
		chromaClient,
		embedder,
	)
	log.Info().Msg("Decision engine initialized")

	// ── Start Data Pipelines ───────────────────────────────────────

	// Market Data Collector (WebSocket via kraken CLI)
	collector := market.NewCollector(krakenClient, dbClient, stateMgr, cfg.TradePairs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		collector.Start(ctx)
	}()

	// PRISM Signal Polling (every 60s)
	wg.Add(1)
	go func() {
		defer wg.Done()
		pollSignals(ctx, prismClient, stateMgr, cfg.TradePairs)
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
func pollSignals(ctx context.Context, prism *news.PrismClient, stateMgr *state.MemoryManager, pairs []string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial fetch
	fetchAndStoreSignals(ctx, prism, stateMgr, pairs)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fetchAndStoreSignals(ctx, prism, stateMgr, pairs)
		}
	}
}

func fetchAndStoreSignals(ctx context.Context, prism *news.PrismClient, stateMgr *state.MemoryManager, pairs []string) {
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

	for _, s := range summaries {
		stateMgr.UpdateSignal(s.Symbol, s.Momentum, s.Breakout, s.Volume)
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
	articles, err := prism.FetchCryptoNews(ctx, 10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch PRISM news")
		return
	}

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
			return
		}
		measurement, tags, fields, ts := trade.ToPointData()
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := db.WritePoint(dbCtx, measurement, tags, fields, ts); err != nil {
			log.Error().Err(err).Msg("Failed to write trade to InfluxDB")
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
