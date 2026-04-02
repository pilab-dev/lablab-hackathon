# Kraken AI Trader — TODO List

## Legend

| Symbol | Meaning |
|--------|---------|
| `[ ]` | Not started |
| `[~]` | In progress |
| `[x]` | Done |
| `P0` | Must have for competition |
| `P1` | Should have |
| `P2` | Nice to have |

## Hackathon: March 30 – April 12, 2026
## Kraken Challenge — Ranked by Net PnL

---

## Phase 0: Immediate (Today)

### Registration & Setup
- [ ] Register project at early.surge.xyz (REQUIRED for prize eligibility)
  - Username: admin
  - Password: JBRv2xWG7AzwVrLz88
- [ ] Create Kraken account (if not exists)
- [ ] Generate API keys: Query Funds + Modify Orders (NO withdrawal)
- [ ] Install kraken-cli: `curl -LsSf https://github.com/krakenfx/kraken-cli/releases/latest/download/kraken-cli-installer.sh | sh`
- [ ] Verify: `kraken ticker BTCUSD -o json`
- [ ] Install Ollama: `brew install ollama`
- [ ] Pull models: `ollama pull llama3.1:8b && ollama pull nomic-embed-text`
- [ ] Create Twitter/X account for Social Engagement prize
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** nothing

### Project Scaffolding
- [ ] `go mod init kraken-trader`
- [ ] Create directory structure (see PLAN.md)
- [ ] Add `Makefile` with common targets
- [ ] Add `.gitignore`
- [ ] Add `.env.example` → copy to `.env` and fill in API keys
- [ ] Create `configs/docker-compose.yml`
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** nothing

### Docker Services
- [ ] Add NATS with JetStream to `docker-compose.yml`
- [ ] `make docker-up` — start InfluxDB + ChromaDB + NATS
- [ ] Verify InfluxDB: `curl http://localhost:8086/ping`
- [ ] Verify ChromaDB: `curl http://localhost:8000/api/v1/heartbeat`
- [ ] Verify NATS: `nc -zv localhost 4222`
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** docker-compose.yml

---

## Phase 1: Core Infrastructure (Days 1-2)

### Go — kraken-cli Subprocess Wrapper
- [ ] Create `pkg/kraken/cli.go`
- [ ] Implement `RunCommand(cmd string, args ...string) ([]byte, error)`
- [ ] Implement `RunJSON(cmd string, args ...string) (interface{}, error)`
- [ ] Add timeout handling (default 30s)
- [ ] Add error parsing from JSON error envelopes
- [ ] Write tests with mock exec
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** kraken-cli installed

### Go — InfluxDB Client
- [ ] Add `github.com/influxdata/influxdb-client-go/v2` dependency
- [ ] Create `internal/storage/influxdb.go`
- [ ] Implement `NewClient(dsn, token, org, bucket) (*Client, error)`
- [ ] Implement `WritePoint(measurement, tags, fields)`
- [ ] Implement `QueryFlux(flux) ([]Point, error)`
- [ ] Create `internal/storage/models.go` with measurement structs
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** InfluxDB running

### Go — NATS Client
- [ ] Add `github.com/nats-io/nats.go` dependency
- [ ] Create `pkg/messaging/nats.go`
- [ ] Implement `Connect(url)` and setup JetStream context
- [ ] Implement `CreateStreams()` for MARKET, INTELLIGENCE, TRADING
- [ ] Implement Pub/Sub wrappers
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** NATS running

### Go — Market Data Collector
- [ ] Update `internal/market/collector.go` to use WebSockets
- [ ] Parse `kraken ws ticker` output
- [ ] Publish directly to NATS Core (`market.tick.{pair}`) instead of channel/RAM
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** kraken wrapper, NATS client

---

## Phase 2: News & Sentiment (Days 2-3)

### News & Signals Crawler
- [ ] Add `github.com/mmcdole/gofeed` dependency
- [ ] Create `internal/news/crawler.go`
- [ ] Implement PRISM API `/news/crypto` fetcher
- [ ] Implement PRISM API `/signals/summary` fetcher
- [ ] Define fallback RSS feed URLs (CoinDesk, Cointelegraph)
- [ ] Implement `FetchFeeds() ([]Article, error)`
- [ ] Publish news to NATS (`news.crypto`)
- [ ] Publish signals to NATS (`signal.prism.{pair}`)
- [ ] Run every 5 minutes
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** nothing

### Embedding Pipeline
- [ ] Create `internal/news/embedder.go`
- [ ] Implement `GenerateEmbedding(text string) ([]float32, error)` via Ollama API
- [ ] Implement `BatchEmbed(articles []Article) ([]Embedding, error)`
- [ ] Add retry logic for Ollama failures
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** Ollama running

### ChromaDB Integration
- [ ] Create `internal/news/chroma.go`
- [ ] Implement `NewClient(url string) (*Client, error)`
- [ ] Implement `CreateCollection(name string)`
- [ ] Implement `AddEmbeddings(collection, ids, embeddings, metadatas)`
- [ ] Implement `QuerySimilar(collection, embedding, nResults, whereFilter)`
- [ ] Implement sentiment extraction from metadata
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** ChromaDB running, embedder

---

## Phase 3: Decision Engine & Trading (Days 3-4)

### Decision Engine
- [ ] Create `internal/decision/prompt.go`
- [ ] Write system prompt template with JSON schema
- [ ] Create `internal/decision/engine.go`
- [ ] Implement `BuildPrompt(marketData, news, portfolio) string`
- [ ] Implement `CallOllama(prompt string) (string, error)` via HTTP
- [ ] Create `internal/decision/parser.go`
- [ ] Implement `ParseResponse(raw string) ([]TradeDecision, error)`
- [ ] Add JSON schema validation
- [ ] Add confidence threshold filter (skip decisions < 0.6)
- [ ] Wire up: query InfluxDB + ChromaDB → build prompt → call Ollama → parse
- [ ] Run on ticker (every 30s)
- [ ] Send decisions to `decisionCh` channel
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** market collector, news crawler, Ollama

### Risk Management (Critical for PnL)
- [ ] Create `internal/execution/risk.go`
- [ ] Implement `CheckPositionSize(portfolio, decision) bool`
- [ ] Implement `CheckDailyLoss(tradesToday, limit) bool`
- [ ] Implement `CheckCooldown(pair, lastTradeTime) bool`
- [ ] Implement `CheckMaxOpenPositions(openPositions, max) bool`
- [ ] Implement stop-loss tracking
- [ ] Implement take-profit tracking
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** decision engine

### Trade Executor (LIVE MODE — this counts for PnL)
- [ ] Create `internal/execution/runner.go`
- [ ] Implement `NewExecutor(mode string) (*Executor, error)`
- [ ] Create `internal/execution/live.go`
- [ ] Implement `LiveBuy(pair, size, orderType, price)`
- [ ] Implement `LiveSell(pair, size, orderType, price)`
- [ ] Implement `LiveBalance()`
- [ ] Implement `LiveOpenOrders()`
- [ ] Consume `decisionCh` in a loop
- [ ] Apply risk checks before execution
- [ ] Execute trades, log results to InfluxDB
- [ ] Add trade cooldown (min 60s between trades on same pair)
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** kraken wrapper, decision engine, risk checks

### Main Orchestrator
- [ ] Create `cmd/trader/main.go`
- [ ] Load config from `.env`
- [ ] Initialize all components
- [ ] Start all goroutines
- [ ] Wire channels together
- [ ] Add signal handler (SIGINT/SIGTERM) for graceful shutdown
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** all components above

---

## Phase 4: Dashboard & Social (Days 4-5)

### Dashboard API
- [ ] Create `internal/api/server.go`
- [ ] Implement `GET /status` → portfolio summary + PnL
- [ ] Implement `GET /decisions` → last 20 AI decisions
- [ ] Implement `GET /decisions/{id}` → single decision with reasoning
- [ ] Implement `GET /performance` → PnL over time (from InfluxDB)
- [ ] Implement `GET /health` → liveness check
- [ ] Implement WebSocket `/ws` → live price + trade stream
- [ ] **Owner:** ___ **Priority:** P1 **Depends on:** InfluxDB, trade executor

### Next.js Dashboard Frontend
- [ ] `npx create-next-app@latest web` (TypeScript, Tailwind, App Router)
- [ ] Install dependencies: `zustand recharts lucide-react`
- [ ] Add shadcn/ui: `npx shadcn-ui@latest init`
- [ ] Setup WebSocket hook (`hooks/useLiveSync.ts`) to connect to Go backend
- [ ] Create Zustand store (`lib/store.ts`) for live data
- [ ] Create `LiveTicker` component (flashing green/red on update)
- [ ] Create `DecisionFeed` component showing LLM reasoning
- [ ] **Owner:** ___ **Priority:** P1 **Depends on:** API server

### Social Poster (For Social Engagement Prize)
- [ ] Create `internal/social/poster.go`
- [ ] Set up Twitter/X API credentials
- [ ] Implement `PostTrade(decision TradeDecision) error`
- [ ] Implement `PostDailySummary(stats) error`
- [ ] Format tweets with hashtags: #KrakenCLI #AITrading #BuildInPublic
- [ ] Tag @krakenfx @lablabai @Surgexyz_
- [ ] **Owner:** ___ **Priority:** P1 **Depends on:** trade executor

---

## Phase 5: Paper Testing (Days 5-6)

### Paper Trading Validation
- [ ] Switch to `TRADING_MODE=paper`
- [ ] Run bot for 24-48 hours on paper trading
- [ ] Monitor PnL, win rate, trade frequency
- [ ] Tune prompt based on results
- [ ] Adjust risk parameters:
  - Confidence threshold
  - Position sizing
  - Stop-loss / take-profit levels
  - Trade cooldown
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** all components

### Edge Case Testing
- [ ] Test with Ollama stopped → graceful degradation
- [ ] Test with InfluxDB stopped → buffered writes
- [ ] Test with no internet → clean error messages
- [ ] Test rapid price swings → bot response
- [ ] Test Kraken rate limit → backoff behavior
- [ ] Test daily loss limit → circuit breaker triggers
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** everything working

---

## Phase 6: Go Live (Day 7+)

### Live Trading Launch
- [ ] Switch to `TRADING_MODE=live`
- [ ] Start with minimal position sizes (0.001 BTC)
- [ ] Monitor closely for first 24 hours
- [ ] Adjust risk parameters based on real behavior
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** paper testing passed

### Social Engagement Campaign
- [ ] Enable auto-posting to Twitter/X
- [ ] Post architecture thread
- [ ] Post daily PnL updates
- [ ] Share interesting AI decisions
- [ ] Tag @krakenfx @lablabai @Surgexyz_
- [ ] **Owner:** ___ **Priority:** P1 **Depends on:** live trading running

### Monitoring & Optimization (Days 8-14)
- [ ] Daily PnL review
- [ ] Prompt tuning if losing
- [ ] Position sizing adjustments
- [ ] News source evaluation
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** live trading running

---

## Phase 7: Submission (Before April 12)

### Submission Materials
- [ ] Write README.md
- [ ] Record 3-min demo video
- [ ] Create slide presentation (architecture, strategy, results)
- [ ] Prepare cover image (architecture diagram)
- [ ] Submit at lablab.ai
- [ ] Provide read-only Kraken API key for leaderboard verification
- [ ] **Owner:** ___ **Priority:** P0 **Depends on:** everything working

---

## Work Split (2 People)

| Person | Focus Areas |
|--------|------------|
| **Person A** | Infrastructure, kraken wrapper, market collector, trade executor, risk management, main orchestrator |
| **Person B** | PRISM API integration, embeddings, ChromaDB, decision engine, dashboard, social poster |

Overlap days 3-4 for integration work.

---

## Blockers & Dependencies

```
Phase 0: kraken-cli + Ollama installed → everything else
Phase 1: kraken wrapper + InfluxDB → Market collector
Phase 2: Ollama running → Embedder → ChromaDB → News pipeline
Phase 3: Market collector + News pipeline + Ollama → Decision engine
         Decision engine + kraken wrapper + risk checks → Trade executor
         All components → Main orchestrator
Phase 4: Trade executor + InfluxDB → Dashboard
         Trade executor → Social poster
Phase 5: Everything → Paper testing (48h)
Phase 6: Paper testing passed → Live trading
Phase 7: Everything → Submission
```
