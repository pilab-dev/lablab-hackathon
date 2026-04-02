# Kraken AI Trader — MVP Definition

## Hackathon Reality Check

This is **not a demo**. This is a **live PnL-ranked competition** running March 30 – April 12, 2026.

The bot must run **autonomously 24/7** on a real Kraken account. Net PnL determines ranking.

**Success criteria:** Positive PnL at the end of the competition period, with transparent AI decision-making.

---

## MVP Feature Set (Must Have)

### 1. Market Data Collection
- [x] Poll BTCUSD and ETHUSD prices every 5 seconds via kraken-cli
- [x] Store OHLCV data in InfluxDB
- [ ] Basic price change detection (up/down/flat)

### 2. News Sentiment
- [x] Fetch 3 RSS feeds every 5 minutes
- [x] Generate embeddings via Ollama
- [x] Store in ChromaDB
- [ ] Simple sentiment classification (bullish/bearish/neutral)

### 3. AI Decision Making
- [x] Decision engine runs every 30 seconds
- [x] Queries market data from InfluxDB
- [x] Queries recent news from ChromaDB
- [x] Sends structured prompt to Ollama llama3.1:8b
- [x] Parses JSON response with trade decisions
- [ ] Confidence threshold filtering (skip < 0.6)

### 4. Trade Execution (LIVE — this is what counts)
- [x] Live trading via kraken-cli with real API keys
- [x] Market orders (buy/sell)
- [ ] Trade cooldown (60s minimum between trades on same pair)
- [ ] Log all trades to InfluxDB
- [ ] Risk checks before every trade

### 5. Risk Management (Critical for PnL)
- [ ] Max position size: 10% of portfolio
- [ ] Max open positions: 3
- [ ] Daily loss limit: 5% of portfolio (circuit breaker)
- [ ] Stop-loss: 5% below entry
- [ ] Take-profit: 10% above entry

### 6. Dashboard
- [x] Simple HTML page
- [x] Current portfolio value and PnL
- [x] Recent decisions table with reasoning
- [x] Trade history
- [ ] Live price ticker via WebSocket

### 7. Social Engagement (Optional but valuable)
- [ ] Auto-post trades to Twitter/X
- [ ] Daily PnL summary posts
- [ ] Build-in-public thread

---

## MVP Scope Boundaries

### IN Scope
- **Live trading** on real Kraken account (small amounts)
- 2 crypto pairs: BTCUSD, ETHUSD
- 1 decision cycle every 30 seconds
- Market orders only
- Basic news sentiment (bullish/bearish/neutral)
- Simple HTML dashboard
- Local Ollama on Mac M4
- Risk management (stop-loss, daily limit, position sizing)

### OUT of Scope (Post-MVP)
- xStocks trading (AAPLx, TSLAx)
- Limit orders
- Technical indicators (RSI, MACD) — add later if PnL is poor
- Backtesting
- Multiple LLM comparison
- Futures trading
- Telegram/Discord notifications
- Advanced risk management (trailing stops, hedging)

---

## Competition Strategy

### Phase 1: Days 1-3 (Setup + Paper Test)
- Build the full pipeline
- Run on paper trading for 48 hours
- Tune prompt, confidence threshold, risk parameters
- Verify all components work together

### Phase 2: Days 4-5 (Go Live Small)
- Switch to live trading with minimal position sizes
- Start with 0.001 BTC per trade (~$67)
- Monitor closely for first 24 hours
- Adjust risk parameters based on real behavior

### Phase 3: Days 6-14 (Optimize)
- Let it run autonomously
- Monitor PnL dashboard daily
- Tweak prompt if losing
- Adjust position sizing based on confidence
- Post to Twitter/X for Social Engagement prize

### Phase 4: Days 15-12 (Final Push)
- Lock in winning strategy
- Don't over-trade in final days
- Prepare submission materials

---

## Submission Requirements

| Item | Status |
|------|--------|
| Project Title | Kraken AI Trader |
| Short Description | Autonomous AI trading agent using local LLM + Kraken CLI |
| Long Description | See README.md |
| Technology Tags | Go, Ollama, Kraken CLI, InfluxDB, ChromaDB |
| Cover Image | Architecture diagram |
| Video Presentation | 3-min demo of dashboard + live trading |
| Slide Presentation | Architecture, strategy, results |
| Public GitHub Repo | This repo |
| Demo Application | Dashboard at localhost:8080 (or deployed) |
| Read-only Kraken API Key | For leaderboard verification |
| early.surge.xyz Registration | **REQUIRED for prize eligibility** |

---

## Pre-Competition Checklist

- [ ] Register project at early.surge.xyz (credentials in hackathon page)
- [ ] Create Kraken account with read-only + trade API keys
- [ ] API key permissions: Query Funds + Modify Orders (NO withdrawal)
- [ ] kraken-cli installed and tested
- [ ] Ollama running with llama3.1:8b and nomic-embed-text
- [ ] Docker services (InfluxDB, ChromaDB) running
- [ ] Bot tested on paper trading for 48+ hours
- [ ] Risk parameters tuned and documented
- [ ] Twitter/X account ready for Social Engagement posts
- [ ] Dashboard accessible (localhost or deployed)
- [ ] .env configured with real API keys (never committed)
- [ ] Kill switch tested (can stop all trading immediately)

---

## Risk Parameters (Starting Values)

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Initial capital | Whatever is in Kraken account | Start small |
| Max position size | 10% of portfolio | Limit exposure |
| Max open positions | 3 | Diversify without over-committing |
| Daily loss limit | 5% of portfolio | Circuit breaker |
| Stop-loss | 5% below entry | Limit per-trade loss |
| Take-profit | 10% above entry | Lock in gains |
| Confidence threshold | 0.6 | Skip uncertain trades |
| Trade cooldown | 60s | Prevent over-trading |
| Decision interval | 30s | Enough time for market to move |

---

## Judging Criteria Alignment

| Criteria | How We Address It |
|----------|------------------|
| **Application of Technology** | Go + Ollama + kraken-cli + InfluxDB + ChromaDB, all integrated |
| **Presentation** | Live dashboard, transparent AI reasoning, social posts |
| **Business Value** | Autonomous trading that actually makes money (PnL-ranked) |
| **Originality** | Local AI making decisions, news sentiment + market data fusion |
| **Trading Performance** | Net PnL — the primary ranking metric |
| **Social Engagement** | Build-in-public posts on Twitter/X |
