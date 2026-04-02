# Kraken AI Trader — Architecture Plan

## Vision

An autonomous AI trading agent that reads crypto/xStock market data and news sentiment,
then makes and executes trading decisions using a local LLM (Ollama) on Apple Silicon.

Built for the **lablab.ai AI Trading Agents Hackathon** (March 30 – April 12, 2026).
Competes in the **Kraken Challenge** — ranked by net PnL during the competition period.

## Hackathon Requirements

| Requirement | Our Approach |
|-------------|-------------|
| Uses Kraken CLI for market data + trades | Primary execution layer |
| AI-driven strategy analyzing signals | Ollama llama3.1:8b + news + OHLCV |
| Autonomous workflow | 5 goroutines, zero human input needed |
| Read-only API key for leaderboard verification | Provided at submission |
| Build in public (Social Engagement prize) | Twitter/X posts, architecture threads |
| Net PnL ranking | Bot runs 24/7 during competition |

## Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Core | Go 1.22+ | Goroutines, fast, single binary |
| Messaging | NATS + JetStream | High-throughput pub/sub, decoupled hot/cold paths |
| Trading CLI | kraken-cli (Rust binary) | Full Kraken API, paper + live trading, JSON output, MCP server |
| LLM | Ollama + llama3.1:8b | Local, fast on M4, no API costs, no rate limits |
| Embeddings | Ollama + nomic-embed-text | Local, no Google API dependency |
| Time-series DB | InfluxDB 2.7 (Docker) | Long-term OHLCV/Trades storage, written via JetStream |
| Vector DB | ChromaDB (Docker) | News embedding storage + similarity search |
| News | PRISM API (`/news/crypto`) + RSS fallback | Official partner API for better scoring |
| Signals | PRISM API (`/signals/summary`) | Momentum, breakout, and volume signals |
| Dashboard | Next.js 14 + Zustand + WebSockets | Professional UI, real-time reactive state |
| Social | Twitter/X API (optional) | Auto-post trade decisions for Social Engagement prize |

## System Architecture

```text
┌────────────────────────────────────────────────────────────────────────────┐
│                             NATS MESSAGING HUB                             │
│                                                                            │
│   [HOT PATH: Core Pub/Sub]                 [COLD PATH: JetStream]          │
│   Instant, in-memory routing               Persistent message streams      │
└─────────▲──────────────────▼───────────────────────▲─────────────────┬─────┘
          │                  │                       │                 │
     (Publishes)        (Subscribes)                 │            (Subscribes)
          │                  │                       │                 │
┌─────────┴────────┐ ┌───────┴──────────┐   ┌────────┴────────┐ ┌──────▼───────┐
│ Market Collector │ │ Decision Engine  │   │  News Crawler   │ │ Data Archiver│
│ (Kraken WS)      │ │ (State + Ollama) │   │  (PRISM/RSS)    │ │ (Go Worker)  │
└──────────────────┘ └───────┬──────────┘   └─────────────────┘ └──────┬───────┘
                             │                                         │
                        (Publishes)                              (Writes via HTTP)
                             │                                         │
┌──────────────────┐ ┌───────▼──────────┐                       ┌──────▼───────┐
│ PRISM API /      │ │ Trade Executor   │                       │  InfluxDB    │
│ Kraken CLI       │ │ (Kraken CLI)     │                       │  ChromaDB    │
└─────────▲────────┘ └───────┬──────────┘                       └──────────────┘
          │                  │
          └────(Network)─────┘
```

## Go Application — Components

```text
main()
 ├── 1: MarketDataCollector
 │    ├── Connects to kraken-cli WebSockets
 │    └── Publishes `market.tick.BTCUSD` to NATS
 │
 ├── 2: NewsCrawler
 │    ├── Fetches news/signals from PRISM and RSS every 5 min
 │    ├── Generates embeddings via Ollama
 │    ├── Stores in ChromaDB with metadata
 │    └── Publishes `news.crypto` to NATS
 │
 ├── 3: DecisionEngine (the brain)
 │    ├── Subscribes to NATS Core (`market.*`, `news.*`)
 │    ├── Triggers Ollama prompt on Sudden Spikes or High-Signal News
 │    └── Publishes `trade.decision.BTCUSD` to NATS
 │
 ├── 4: TradeExecutor
 │    ├── Subscribes to `trade.decision.*`
 │    ├── Executes via kraken-cli subprocess
 │    └── Publishes `trade.execution.BTCUSD` to NATS
 │
 ├── 5: DataArchiver (The Cold Path)
 │    ├── Subscribes to NATS JetStream (`MARKET`, `TRADING`)
 │    └── Batch writes to InfluxDB for historical storage
 │
 └── 6: Dashboard API
      ├── HTTP server on :8080 (REST endpoints for historical data)
      └── WebSocket /ws (Pushes NATS events live to Next.js Frontend)
```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose                           │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │   InfluxDB   │  │   ChromaDB   │                        │
│  │  (port 8086) │  │  (port 8000) │                        │
│  └──────┬───────┘  └──────┬───────┘                        │
└─────────┼─────────────────┼─────────────────────────────────┘
          │                 │
          │          ┌──────▼──────┐
          │          │  Ollama     │
          │          │  (port 11434)│
          │          │  llama3.1:8b │
          │          │  nomic-embed │
          │          └──────┬──────┘
          │                 │
          │          ┌──────▼──────┐
          │          │  kraken-cli │
          │          │  (binary)   │
          │          └──────┬──────┘
          │                 │
          └─────────────────┼─────────────────────────────
                            │
                     ┌──────▼──────┐
                     │   Kraken    │
                     │   Exchange  │
                     │   REST API  │
                     └─────────────┘
                            │
                     ┌──────▼──────┐
                     │   Twitter/X │  (Social Engagement)
                     │   (optional)│
                     └─────────────┘
```

## Go Application — Goroutine Layout

```
main()
 ├── goroutine 1: MarketDataCollector
 │    ├── Polls kraken ticker for each pair every 5s
 │    ├── Polls kraken ohlc (1m candles) every 60s
 │    └── Writes all data to InfluxDB
 │
 ├── goroutine 2: NewsCrawler
 │    ├── Fetches news from PRISM API and RSS feeds every 5 min
 │    ├── Generates embeddings via Ollama
 │    └── Stores in ChromaDB with metadata
 │
 ├── goroutine 3: DecisionEngine (the brain)
 │    ├── Runs every 30s (configurable)
 │    ├── Queries InfluxDB for recent OHLCV
 │    ├── Queries ChromaDB for recent news sentiment
 │    ├── Queries current portfolio state
 │    ├── Builds prompt with market state + news + portfolio
 │    ├── Sends to Ollama, parses JSON response
 │    └── Sends TradeDecision to channel
 │
 ├── goroutine 4: TradeExecutor
 │    ├── Receives decisions via channel
 │    ├── Risk checks before execution:
 │    │   - Max position size
 │    │   - Daily loss limit
 │    │   - Cooldown per pair
 │    │   - Max open positions
 │    ├── Executes via kraken-cli subprocess:
 │    │   - kraken order buy/sell ... -o json    (live)
 │    │   - kraken paper buy/sell ... -o json    (paper, dev only)
 │    ├── Logs result to InfluxDB
 │    └── Triggers social post (if enabled)
 │
 ├── goroutine 5: Dashboard/API
 │    ├── HTTP server on :8080
 │    ├── GET /status → portfolio, open orders, PnL
 │    ├── GET /decisions → recent AI decisions with reasoning
 │    ├── GET /performance → PnL chart data from InfluxDB
 │    ├── GET /leaderboard → current standing (if API available)
 │    └── WebSocket /ws → live price + trade stream
 │
 └── goroutine 6: SocialPoster (optional, for Social Engagement prize)
      ├── Receives trade events via channel
      ├── Formats tweet: "🤖 AI Trader just BOUGHT 0.01 BTC at $67,234"
      ├── Adds reasoning snippet
      ├── Posts to Twitter/X via API
      └── Logs engagement metrics
```

## Data Flow — One Decision Cycle

```
1. MarketDataCollector writes latest prices to InfluxDB (continuous)
2. NewsCrawler stores article embeddings in ChromaDB (every 5 min)
3. DecisionEngine wakes up (every 30s):
   a. Query InfluxDB: last 1h OHLCV for all pairs
   b. Query ChromaDB: similar news from last 24h, compute sentiment
   c. Query PRISM API for technical signals (momentum, breakout)
   d. Query kraken-cli: current portfolio balance + open positions
   e. Build prompt:
      ┌─────────────────────────────────────────────────┐
      │ You are an autonomous crypto trading agent.     │
      │                                                 │
      │ MARKET STATE:                                   │
      │ - BTC/USD: $67,234 (↑2.1% 1h, vol: 1,234 BTC)  │
      │ - ETH/USD: $3,456 (↓0.5% 1h, vol: 8,901 ETH)   │
      │                                                 │
      │ NEWS SENTIMENT (last 24h):                      │
      │ - Bullish: Fed rate cut expectations (CoinDesk) │
      │ - Bearish: Exchange hack report (CryptoSlate)   │
      │                                                 │
      │ PRISM AI SIGNALS:                               │
      │ - BTCUSD: Strong Momentum, Breakout Detected    │
      │                                                 │
      │ PORTFOLIO:                                      │
      │ - Balance: $8,432 USD, 0.05 BTC, 1.2 ETH       │
      │ - Open positions: 2                             │
      │ - Daily PnL: +$234 (+2.8%)                     │
      │                                                 │
      │ What should I do? Respond in JSON.              │
      └─────────────────────────────────────────────────┘
   f. POST to Ollama API → parse JSON response
   g. Risk checks pass? → Send TradeDecision to TradeExecutor
4. TradeExecutor:
   a. kraken order buy BTCUSD 0.01 --type market -o json
   b. Parse response, log to InfluxDB
   c. Emit event to dashboard WebSocket
   d. Trigger social post (if enabled)
```

## Trading Pairs (MVP)

| Pair | Asset Class | Priority | Rationale |
|------|------------|----------|-----------|
| BTCUSD | Crypto spot | P0 | Highest liquidity, best for PnL |
| ETHUSD | Crypto spot | P0 | Second most liquid, correlates with BTC |
| SOLUSD | Crypto spot | P1 | Higher volatility = more opportunities |
| AAPLx/USD | xStocks | P2 | Diversification, lower 24/7 activity |

## kraken-cli Integration

All commands use `-o json 2>/dev/null` pattern. Exit code 0 = success.

### Market Data (no auth needed)
```bash
kraken ticker BTCUSD -o json
kraken ohlc BTCUSD --interval 60 -o json
kraken orderbook BTCUSD --count 10 -o json
```

### Live Trading (auth required — THIS IS WHAT COUNTS FOR PnL)
```bash
export KRAKEN_API_KEY="your-key"
export KRAKEN_API_SECRET="your-secret"

kraken order buy BTCUSD 0.001 --type market -o json
kraken order sell BTCUSD 0.001 --type market -o json
kraken open-orders -o json
kraken balance -o json
kraken trades-history -o json
```

### Paper Trading (development/testing only)
```bash
kraken paper init --balance 10000 -o json
kraken paper buy BTCUSD 0.01 --type market -o json
kraken paper status -o json
```

## Ollama Integration

### Chat (Decision Making)
```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.1:8b",
  "stream": false,
  "messages": [
    {"role": "system", "content": "You are an autonomous crypto trading agent..."},
    {"role": "user", "content": "Market data: ..."}
  ],
  "format": {"type": "object", "properties": {...}}
}'
```

### Embeddings (News Processing)
```bash
curl http://localhost:11434/api/embeddings -d '{
  "model": "nomic-embed-text",
  "prompt": "Article text here..."
}'
```

## Project Structure

```
kraken-trader/
├── cmd/
│   └── trader/
│       └── main.go
├── internal/
│   ├── market/
│   │   ├── collector.go
│   │   └── types.go
│   ├── news/
│   │   ├── crawler.go
│   │   ├── embedder.go
│   │   └── chroma.go
│   ├── decision/
│   │   ├── engine.go
│   │   ├── prompt.go
│   │   └── parser.go
│   ├── execution/
│   │   ├── runner.go
│   │   ├── risk.go             # Risk checks before trade
│   │   └── live.go             # Live trading commands
│   ├── social/
│   │   └── poster.go           # Twitter/X integration
│   ├── storage/
│   │   ├── influxdb.go
│   │   └── models.go
│   └── api/
│       └── server.go
├── pkg/
│   └── kraken/
│       └── cli.go
├── configs/
│   └── docker-compose.yml
├── web/
│   └── index.html              # Dashboard
├── .env.example
├── go.mod
├── Makefile
├── PLAN.md
├── TODO.md
├── MVP.md
└── ROADMAP.md
```

## Communication Between Components

All Go components communicate via **NATS**. No direct memory sharing is required between domains.

```go
// NATS Message Formats

type TradeDecision struct {
    Pair       string    `json:"pair"`
    Side       string    `json:"side"`       // "buy" or "sell"
    Size       float64   `json:"size"`
    Type       string    `json:"type"`       // "market" or "limit"
    Price      float64   `json:"price"`      // for limit orders
    Confidence float64   `json:"confidence"` // 0.0 - 1.0
    Reasoning  string    `json:"reasoning"`
    Timestamp  time.Time `json:"timestamp"`
}
```

## Risk Management (Critical for PnL Ranking)

| Rule | Default | Configurable |
|------|---------|-------------|
| Max position size (% of portfolio) | 10% | Yes |
| Max open positions | 3 | Yes |
| Daily loss limit (% of portfolio) | 5% | Yes |
| Cooldown between trades (same pair) | 60s | Yes |
| Stop-loss (% below entry) | 5% | Yes |
| Take-profit (% above entry) | 10% | Yes |
| Confidence threshold | 0.6 | Yes |
| Max leverage | 1x (spot only) | No |

## Error Handling Strategy

| Error Source | Handling |
|-------------|----------|
| kraken-cli rate limit | Exponential backoff, max 5 retries |
| kraken-cli auth failure | Alert, stop trading, log |
| Ollama unavailable | Skip decision cycle, retry next interval |
| InfluxDB down | Buffer in memory, flush when recovered |
| ChromaDB down | Skip news sentiment, trade on market data only |
| Invalid LLM response | Log, skip, retry next cycle |
| Network disconnect | Pause trading, resume on reconnect |

## Trading Modes

| Mode | Description | Env Var |
|------|------------|---------|
| `live` | Real Kraken account (competition mode) | `TRADING_MODE=live` |
| `paper` | Sandbox (development only) | `TRADING_MODE=paper` |

**Default: `paper` for development.** Competition requires `TRADING_MODE=live` with real API keys.

## Social Engagement Strategy

For the Social Engagement prize, the bot can auto-post to Twitter/X:

| Event | Tweet Template |
|-------|---------------|
| Trade executed | "🤖 AI Trader just BOUGHT 0.01 BTC at $67,234. Reasoning: Bullish momentum + positive news sentiment. #KrakenCLI #AITrading" |
| Daily summary | "📊 Day 3: PnL +$234 (+2.8%). 12 trades made. Win rate: 67%. Running autonomous on local AI. #BuildInPublic" |
| Milestone | "🎯 $10,000 → $11,000! AI trading agent just hit 10% return. All decisions made by llama3.1 running locally on M4. #KrakenCLI" |

## Security Notes

- API keys stored in `.env`, never committed
- API key permissions: Query Funds + Modify Orders only (NO withdrawal)
- kraken-cli secrets passed via env vars, not CLI args
- Live mode requires explicit `TRADING_MODE=live` opt-in
- Daily loss limit prevents catastrophic losses
- Read-only API key provided to hackathon organizers for verification
