package news

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"kraken-trader/internal/state"
)

// Crawler manages fetching news and signals and updating the state manager & ChromaDB
type Crawler struct {
	prismClient  *PrismClient
	embedder     *Embedder
	chromaClient *ChromaClient
	stateMgr     *state.MemoryManager
	feedParser   *gofeed.Parser
	rssUrls      []string
}

// NewCrawler initializes the news and signals crawler
func NewCrawler(prismClient *PrismClient, embedder *Embedder, chromaClient *ChromaClient, stateMgr *state.MemoryManager) *Crawler {
	return &Crawler{
		prismClient:  prismClient,
		embedder:     embedder,
		chromaClient: chromaClient,
		stateMgr:     stateMgr,
		feedParser:   gofeed.NewParser(),
		rssUrls: []string{
			"https://www.coindesk.com/arc/outboundfeeds/rss/",
			"https://cointelegraph.com/rss",
		},
	}
}

// Start begins the periodic polling for news and signals
func (c *Crawler) Start(ctx context.Context) {
	log.Info().Msg("Starting News & Signals Crawler...")

	// Initialize ChromaDB Collection
	if c.chromaClient != nil {
		if err := c.chromaClient.Initialize(ctx, "crypto_news"); err != nil {
			log.Error().Err(err).Msg("Failed to initialize ChromaDB collection")
		} else {
			log.Info().Msg("ChromaDB initialized for news embeddings")
		}
	}

	// Fast tick for signals (e.g., every 30 seconds)
	signalTicker := time.NewTicker(30 * time.Second)
	defer signalTicker.Stop()

	// Slow tick for news (e.g., every 5 minutes)
	newsTicker := time.NewTicker(5 * time.Minute)
	defer newsTicker.Stop()

	// Do initial fetches immediately
	c.fetchSignals(ctx)
	c.fetchNews(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("News Crawler stopping...")
			return
		case <-signalTicker.C:
			c.fetchSignals(ctx)
		case <-newsTicker.C:
			c.fetchNews(ctx)
		}
	}
}

func (c *Crawler) fetchSignals(ctx context.Context) {
	if c.prismClient == nil || c.prismClient.apiKey == "" {
		return // Skip if PRISM is not configured
	}

	// For the hackathon, we fetch BTC and ETH signals
	summaries, err := c.prismClient.FetchSignalSummary(ctx, []string{"BTC", "ETH"})
	if err != nil {
		log.Error().Err(err).Msg("Error fetching PRISM signals")
		return
	}

	for _, s := range summaries {
		// Update the RAM state manager so the LLM Prompt Builder has instant access
		c.stateMgr.UpdateSignal(s.Symbol, s.OverallSignal, s.Direction, s.Strength)
	}
}

func (c *Crawler) fetchNews(ctx context.Context) {
	var allArticles []NewsArticle

	// Try PRISM first
	if c.prismClient != nil && c.prismClient.apiKey != "" {
		prismArticles, err := c.prismClient.FetchCryptoNews(ctx, 10, 0)
		if err == nil {
			allArticles = append(allArticles, prismArticles...)
		} else {
			log.Warn().Err(err).Msg("PRISM news fetch failed, falling back to RSS")
		}
	}

	// Fallback/Augment with RSS Feeds
	for _, url := range c.rssUrls {
		feed, err := c.feedParser.ParseURLWithContext(url, ctx)
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("Error parsing RSS feed")
			continue
		}

		for i, item := range feed.Items {
			if i >= 5 { // Only take top 5 per feed to avoid context bloat
				break
			}

			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(item.Link)))
			pubDate := time.Now()
			if item.PublishedParsed != nil {
				pubDate = *item.PublishedParsed
			}

			allArticles = append(allArticles, NewsArticle{
				ID:        hash,
				Title:     item.Title,
				Summary:   item.Description,
				URL:       item.Link,
				Source:    feed.Title,
				Timestamp: pubDate,
			})
		}
	}

	if len(allArticles) == 0 {
		return
	}

	// 1. Update the RAM state manager for quick dashboard/live access
	var stateArticles []state.NewsArticle
	for _, a := range allArticles {
		stateArticles = append(stateArticles, state.NewsArticle{
			ID:        a.ID,
			Title:     a.Title,
			Summary:   a.Summary,
			Source:    a.Source,
			Timestamp: a.Timestamp,
		})
	}
	c.stateMgr.UpdateNews(stateArticles)

	// 2. Embed and Store in ChromaDB for Semantic Search
	if c.embedder != nil && c.chromaClient != nil {
		embeddings, err := c.embedder.BatchEmbed(ctx, allArticles)
		if err != nil {
			log.Error().Err(err).Msg("Error embedding news articles")
			return
		}

		err = c.chromaClient.AddEmbeddings(ctx, allArticles, embeddings)
		if err != nil {
			log.Error().Err(err).Msg("Error storing embeddings in ChromaDB")
		} else {
			log.Info().Int("articles", len(allArticles)).Msg("Successfully embedded and stored articles in ChromaDB")
		}
	}
}
