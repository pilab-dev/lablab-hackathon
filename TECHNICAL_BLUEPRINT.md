# 🛠️ TECHNICAL_BLUEPRINT — System Architecture

## The Stack
- **Core:** Go 1.22+ (High-concurrency trading loop)
- **Messaging:** NATS + JetStream (Decoupled, fault-tolerant data flow)
- **Decision:** Ollama + Llama 3.1:8b (Local inference on Apple Silicon)
- **Data:** PRISM API (Partner signals) + Kraken CLI (Market data/Trades)
- **Storage:** InfluxDB (Time-series) + ChromaDB (Vector news memory)
- **Identity:** ERC-8004 (On-chain reputation and performance validation)
- **Dashboard:** Next.js 14 + WebSockets (Real-time monitoring)

---

## 1. The Autonomous Loop
1. **Collector:** Polls Kraken CLI for prices and PRISM API for technical signals.
2. **Brain:** Decision Engine pulls Market State + News + Signals into a dynamic prompt for Llama 3.1.
3. **Validator:** Risk module checks the Decision against Portfolio limits (Stop-loss, Max Position).
4. **Executor:** Kraken CLI executes the trade; results are published to NATS.
5. **Broadcaster:** Dashboard UI and Social Poster update instantly via NATS subscribers.

## 2. The MCP Edge (Kraken CLI)
We leverage the **Kraken CLI's MCP (Model Context Protocol) Server**.
- This allows our Go backend to treat Kraken as a standardized "Tool" for the LLM.
- Future-proofing: Makes the bot compatible with any MCP-enabled agent framework.

## 3. Risk Safeguards (The "Circuit Breakers")
- **Daily Loss Limit:** If total portfolio value drops 5% in 24h, the bot cancels all orders and stops trading.
- **Confidence Floor:** Decisions with < 70% confidence are discarded.
- **Connectivity Watchdog:** If NATS or Kraken CLI becomes unresponsive, the bot enters "Safe Mode" (liquidates to fiat or holds current positions).

## 4. ERC-8004 Performance Logging
- Every trade execution generates a payload signed by the bot's private key.
- This payload includes: `Timestamp`, `Pair`, `Price`, `PnL`, and `LLM_Reasoning_Hash`.
- This ensures our leaderboard position is backed by **verifiable cryptographic proof**.
