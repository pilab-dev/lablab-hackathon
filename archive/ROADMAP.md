# Kraken AI Trader — Roadmap

Post-MVP improvements organized by category. Each phase builds on the previous.

---

## Phase 1: Bot Tuning (Week 1-2)

### Prompt Engineering
- [ ] A/B test different system prompts
- [ ] Add few-shot examples to prompt (show good trade examples)
- [ ] Experiment with temperature settings (0.1 vs 0.3 vs 0.7)
- [ ] Add market regime detection to prompt (bull/bear/sideways)
- [ ] Chain-of-thought prompting for better reasoning
- [ ] Prompt versioning — track which prompt version made which trades

### Decision Engine Improvements
- [ ] Add confidence scoring calibration
- [ ] Implement voting: run 3 decisions, take majority
- [ ] Add cooldown per pair (no re-entry within X minutes)
- [ ] Maximum positions limit (max 3 open at once)
- [ ] Anti-churn logic: don't buy/sell same pair within N cycles
- [ ] Decision logging with full prompt + response for analysis

### Risk Management
- [ ] Stop-loss: auto-sell if position drops X%
- [ ] Take-profit: auto-sell if position gains X%
- [ ] Position sizing: allocate % of portfolio based on confidence
- [ ] Max drawdown circuit breaker: stop trading if portfolio drops X%
- [ ] Daily loss limit: stop trading after losing $X in a day
- [ ] Trailing stop-loss

---

## Phase 2: Data Enrichment (Week 2-3)

### Technical Indicators
- [ ] RSI (Relative Strength Index)
- [ ] MACD (Moving Average Convergence Divergence)
- [ ] Bollinger Bands
- [ ] EMA/SMA crossovers
- [ ] Volume analysis
- [ ] Feed indicators into prompt alongside raw prices

### More Data Sources
- [ ] Order book depth analysis
- [ ] Funding rates (for futures)
- [ ] Whale alert tracking (large transactions)
- [ ] Fear & Greed Index
- [ ] On-chain metrics (active addresses, transaction volume)
- [ ] Social sentiment from Twitter/X

### News Pipeline Improvements
- [ ] More RSS sources (Bloomberg Crypto, Decrypt, The Block)
- [ ] News deduplication across sources
- [ ] Source credibility weighting
- [ ] Real-time news via WebSocket (if available)
- [ ] Event extraction (Fed meetings, ETF approvals, hacks)
- [ ] Time-decay weighting (older news = less influence)

---

## Phase 3: Model Improvements (Week 3-4)

### LLM Experimentation
- [ ] Test multiple models:
  - llama3.1:8b (baseline)
  - llama3.1:70b (quantized, if RAM allows)
  - qwen2.5:7b (strong reasoning)
  - mistral:7b (fast)
  - deepseek-coder:7b (structured output)
- [ ] Model comparison dashboard (which model trades best)
- [ ] Ensemble: multiple models vote on decisions
- [ ] Fine-tune on historical trade data (if feasible)

### Embedding Improvements
- [ ] Test embedding models:
  - nomic-embed-text (current)
  - mxbai-embed-large
  - all-minilm (faster, smaller)
- [ ] Hybrid search: embedding + keyword
- [ ] News categorization (regulation, tech, market, security)

### Context Window Optimization
- [ ] Summarize old news instead of raw text
- [ ] Compress market data (only send significant changes)
- [ ] Hierarchical prompts: quick scan → deep analysis
- [ ] Token usage monitoring and optimization

---

## Phase 4: Backtesting (Week 4-5)

### Backtesting Engine
- [ ] Historical data ingestion (download OHLCV from Kraken)
- [ ] Replay engine: feed historical data to decision engine
- [ ] Simulated execution with realistic slippage
- [ ] Performance metrics:
  - Total return
  - Sharpe ratio
  - Max drawdown
  - Win rate
  - Profit factor
  - Average trade duration
- [ ] Compare against buy-and-hold baseline

### Parameter Optimization
- [ ] Grid search over decision intervals
- [ ] Optimize confidence thresholds
- [ ] Test different position sizing strategies
- [ ] Compare prompt variants via backtest
- [ ] Walk-forward optimization

### Strategy Comparison
- [ ] AI strategy vs buy-and-hold
- [ ] AI strategy vs simple moving average crossover
- [ ] AI strategy vs RSI-based strategy
- [ ] Visual comparison chart

---

## Phase 5: Advanced Trading (Week 5-6)

### More Trading Pairs
- [ ] Add xStocks: AAPLx/USD, TSLAx/USD, SPYx/USD
- [ ] Add more crypto: SOL/USD, XRP/USD, DOGE/USD
- [ ] Forex pairs: EUR/USD, GBP/USD
- [ ] Dynamic pair selection based on volume/volatility

### Order Types
- [ ] Limit orders with AI-determined prices
- [ ] Stop orders
- [ ] OCO (One-Cancels-Other) orders
- [ ] Batch orders (multiple at once)
- [ ] DCA (Dollar Cost Averaging) mode

### Futures Trading
- [ ] Perpetual futures support
- [ ] Leverage management
- [ ] Funding rate awareness
- [ ] Long/short strategies
- [ ] Hedging: long spot, short futures

### Live Trading
- [ ] Live mode with real API keys
- [ ] Start with tiny positions ($1-5 per trade)
- [ ] Gradual position size increase
- [ ] Kill switch (immediate cancel all + close positions)
- [ ] Real-time P&L alerts

---

## Phase 6: Infrastructure & UX (Week 6-7)

### Dashboard Upgrade
- [ ] React-based SPA
- [ ] Real-time charts (TradingView lightweight charts)
- [ ] Portfolio allocation pie chart
- [ ] Trade history with filters
- [ ] Decision explorer (see full prompt + response)
- [ ] Settings page (adjust parameters live)
- [ ] Dark mode

### Notifications
- [ ] Telegram bot for trade alerts
- [ ] Discord webhook integration
- [ ] Email summaries (daily/weekly)
- [ ] Push notifications for large moves
- [ ] SMS alerts for critical events

### Deployment
- [ ] Dockerize the Go application
- [ ] Kubernetes deployment manifest
- [ ] Health checks and readiness probes
- [ ] Metrics export (Prometheus)
- [ ] Grafana dashboards
- [ ] Log aggregation (Loki or ELK)

### Performance
- [ ] Connection pooling for InfluxDB
- [ ] Batch writes to InfluxDB
- [ ] ChromaDB index optimization
- [ ] Ollama request caching
- [ ] Profile and optimize hot paths

---

## Phase 7: Research & Experimentation (Ongoing)

### Multi-Agent Architecture
- [ ] Separate agents for different roles:
  - Analyst agent: processes news and market data
  - Trader agent: makes decisions
  - Risk agent: approves/rejects trades
  - Portfolio agent: manages allocations
- [ ] Agent communication via message bus
- [ ] Debate mechanism: agents argue before decision

### Reinforcement Learning
- [ ] Reward function design (profit - risk penalty)
- [ ] State space: market data + news + portfolio
- [ ] Action space: buy/sell/hold + size
- [ ] Compare RL agent vs LLM agent
- [ ] Hybrid: RL for sizing, LLM for direction

### Market Making
- [ ] Bid-ask spread optimization
- [ ] Inventory management
- [ ] Adverse selection protection
- [ ] Compare vs directional trading

### Sentiment Analysis
- [ ] Fine-tune small model for crypto sentiment
- [ ] Compare LLM sentiment vs dedicated sentiment model
- [ ] Sentiment momentum (rate of change)
- [ ] Cross-source sentiment divergence detection

---

## Bot Tweaking Guide

### Quick Wins (30 min each)

| Tweak | Expected Impact | Risk |
|-------|----------------|------|
| Increase decision interval (30s → 60s) | Fewer, more considered trades | Low |
| Raise confidence threshold (0.6 → 0.7) | Fewer false signals | Low |
| Add trade cooldown (60s → 300s) | Reduce churn | Low |
| Change temperature (0.3 → 0.1) | More consistent decisions | Low |
| Add stop-loss at 5% | Limit losses | Medium |

### Medium Effort (2-4 hours each)

| Tweak | Expected Impact | Risk |
|-------|----------------|------|
| Add RSI to prompt | Better entry/exit timing | Low |
| Position sizing by confidence | Better risk-adjusted returns | Medium |
| News source weighting | Higher quality sentiment | Low |
| Multi-model voting | More robust decisions | Medium |
| Backtest prompt variants | Find best prompt | Low |

### Deep Dives (1-2 days each)

| Tweak | Expected Impact | Risk |
|-------|----------------|------|
| Fine-tune embedding model | Better news retrieval | High |
| Multi-agent architecture | More sophisticated decisions | High |
| Reinforcement learning | Optimal strategy discovery | High |
| Live trading with real money | Actual profits (or losses) | Very High |

---

## Performance Benchmarks to Track

| Metric | Target | Current |
|--------|--------|---------|
| Decision latency | < 5s | TBD |
| Trade execution latency | < 2s | TBD |
| Uptime | > 99% | TBD |
| Win rate | > 55% | TBD |
| Sharpe ratio | > 1.0 | TBD |
| Max drawdown | < 10% | TBD |
| Trades per day | 10-30 | TBD |
| News articles processed/hour | 50+ | TBD |
