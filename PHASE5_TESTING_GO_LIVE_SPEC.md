# 🧪 Phase 5 Technical Spec: Paper Testing & Go-Live Strategy

*Targeted at Senior Developers managing the transition from development to a live, PnL-ranked environment.*

You have built the Brain (LLM), the Muscle (Executor), and the Face (Dashboard). Phase 5 is about ensuring the bot actually makes money without blowing up the account. This phase bridges the gap between theoretical code and live market dynamics.

---

## 1. Paper Trading Validation (The 48-Hour Crucible)

Before risking real capital, the bot must prove itself in a simulated environment using real-time data.

### A. Environment Configuration
*   **Mode:** Ensure `.env` has `TRADING_MODE=paper`.
*   **Kraken Paper Engine:** The `kraken-cli` supports a `--paper` flag or a dedicated paper environment. If the CLI doesn't natively simulate fills well enough, your `PaperTrader` implementation in Go should track virtual balances in `state.PortfolioState` based on the exact market price at the moment of the decision.

### B. What to Monitor
During the 48-hour paper run, monitor the following via your Next.js dashboard and InfluxDB logs:
1.  **Win Rate:** Are >55% of trades profitable after simulated fees?
2.  **Trade Frequency (Churn):** Is it trading 50 times an hour? If so, your LLM prompt is too sensitive or your cooldown is broken. Target: 10-20 high-conviction trades per day.
3.  **Drawdown:** What is the maximum peak-to-trough drop in portfolio value? (Must stay above your Circuit Breaker limit).

---

## 2. Tuning the Llama 3.1 Brain

The LLM is deterministic given the same prompt (with `temperature=0.1`), but market conditions change.

### A. Adjusting the "Confidence Threshold"
*   **Observation:** The bot is losing money on trades with 0.75 confidence.
*   **Action:** Raise the hard-coded threshold in `risk.go` to `0.85`.

### B. Prompt Engineering Tweaks
*   **Observation:** The bot is buying at the top of green candles right before a pullback.
*   **Action:** Add a rule to the `SystemPrompt`: *"Do not buy if the asset has surged more than 3% in the last hour without a consolidation period."*
*   **Observation:** The bot is ignoring PRISM signals.
*   **Action:** Update the few-shot examples in `prompt.go` to explicitly show the LLM making a decision *because* of a PRISM Momentum signal.

---

## 3. The "Go-Live" Sequence

Switching to real money requires a strict protocol to prevent catastrophic API errors or logic loops.

### A. Kraken Account Setup
1.  Create a fresh API Key on Kraken specifically for this bot.
2.  **Permissions:** Select `Query Funds`, `Query Open Orders & Trades`, `Create & Modify Orders`, and `Cancel & Reject Orders`.
3.  **CRITICAL:** Do **NOT** enable `Withdraw Funds`.

### B. The Deployment Checklist
- [ ] Wallet funded with starting capital (e.g., $100 - $500).
- [ ] `.env` updated with `TRADING_MODE=live` and the real API keys.
- [ ] NATS streams purged (`make reset-nats`) to prevent the bot from reading old paper-trading decisions and trying to execute them live.
- [ ] Social Poster enabled (`ENABLE_SOCIAL_POSTING=true`).

### C. The Kill Switch
You need a way to stop the bot instantly if it goes rogue.
*   **Implementation:** Expose a hidden endpoint on your Go API (e.g., `POST /api/emergency-stop` with an auth token).
*   **Action:** When called, the Go backend must:
    1. Send a `CancelAllOrders` command via `kraken-cli`.
    2. Optional: Send a `MarketSell` for all open crypto positions to return to USD.
    3. Call `os.Exit(1)` or sleep infinitely.

---

## 4. Live Monitoring & The "Don't Touch It" Rule

Once the bot is live and ranked on the Lablab/Kraken leaderboard:
1.  **Do not restart the bot casually.** NATS JetStream will recover state, but restarting during an active trade can mess up local portfolio tracking.
2.  **Let the Circuit Breaker do its job.** If the bot is down 3%, trust your 5% hard limit. Human interference often ruins algorithmic strategies.
3.  **Focus on Hype:** While the bot trades, your job is to post screenshots of the dashboard and the LLM's reasoning to Twitter/X to win the Social Engagement prize.
