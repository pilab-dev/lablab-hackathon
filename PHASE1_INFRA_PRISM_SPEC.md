# 🏗️ Phase 1 Technical Spec: Core Infrastructure & PRISM Data

*Targeted at Senior Developers building the foundation: Data Ingestion, Messaging, and Partner Integrations.*

Before the "Brain" (Phase 2) can make decisions, it needs high-quality data and a robust nervous system to deliver it. This phase focuses on the **Strykr/PRISM API** (our market intelligence), the **Kraken CLI** wrapper (our market data source), and **NATS** (our message bus).

---

## 1. The Strykr / PRISM API Integration (The "Alpha")

Using the PRISM API is crucial not just for trading performance, but because **Strykr is a hackathon partner**. Utilizing their API boosts your "Application of Technology" judging score.

### A. What is PRISM?
PRISM provides AI-generated technical signals and curated crypto news. Instead of manually calculating moving averages or RSI in Go, we outsource this to PRISM and feed the results directly into our Llama 3.1 Brain.

### B. The Endpoints to Implement (`internal/news/prism.go`)
You need to wrap two specific REST endpoints:

1.  **Technical Signals (`GET /signals/summary?symbols=BTC,ETH`)**
    *   **Returns:** JSON containing `momentum`, `breakout`, and `volume` signals.
    *   **Interpretation:** 
        *   *Momentum:* Is the trend accelerating? (Bullish/Bearish/Neutral)
        *   *Breakout:* Is it crossing a key resistance/support line?
        *   *Volume:* Is there enough trading volume to sustain the move?
    *   **Usage:** Save this directly into your `state.MemoryManager` so the LLM Prompt Builder can inject it into the prompt.

2.  **Curated News (`GET /news/crypto`)**
    *   **Returns:** The latest high-impact crypto news articles.
    *   **Usage:** This replaces the need to scrape Twitter or RSS feeds. Pass these articles into your `news.Embedder` to store in ChromaDB for semantic search.

### C. Authentication
*   Ensure your `.env` contains `PRISM_API_KEY`.
*   Pass it via the `X-API-Key` HTTP header in your `net/http` client.

---

## 2. The Kraken CLI Subprocess Wrapper (`pkg/kraken/cli.go`)

**Hackathon Rule:** You *must* use the Kraken CLI for market data and execution, not the raw Kraken REST APIs.

### A. The Subprocess Pattern
Because the Kraken CLI is a Rust binary, you must call it from Go using `os/exec`.

```go
// Example Wrapper implementation concept
func RunKrakenCmd(ctx context.Context, args ...string) ([]byte, error) {
    // Force JSON output to easily parse in Go
    fullArgs := append(args, "-o", "json")
    
    cmd := exec.CommandContext(ctx, "kraken", fullArgs...)
    
    // Inject API keys from your environment securely
    cmd.Env = append(os.Environ(), 
        "KRAKEN_API_KEY="+os.Getenv("KRAKEN_API_KEY"),
        "KRAKEN_API_SECRET="+os.Getenv("KRAKEN_API_SECRET"),
    )
    
    return cmd.Output() // Returns stdout. Make sure to capture stderr for debugging!
}
```

### B. Core CLI Commands to Wrap
*   **Market Data:** `kraken ticker BTCUSD` (Poll this every 5 seconds).
*   **Account State:** `kraken balance` and `kraken open-orders`.
*   **Trading (Phase 3):** `kraken order buy BTCUSD 0.01 --type market`.

---

## 3. NATS & JetStream (The Nervous System)

Do not use Go Channels (`chan`) to pass data between your major components. If a channel blocks or a goroutine crashes, you lose data. Use NATS.

### A. The "Hot Path" (NATS Core)
Use standard NATS Pub/Sub for things that need sub-millisecond latency and don't matter if they are dropped (e.g., live price ticks).
*   **Subject:** `market.tick.BTCUSD`
*   **Publisher:** The Kraken CLI polling loop.
*   **Subscriber:** The Next.js Dashboard WebSocket and the `state.MemoryManager`.

### B. The "Cold Path" (NATS JetStream)
Use JetStream for things that *must* be saved (decisions, trade executions). JetStream acts like a lightweight Kafka.
*   **Stream:** `TRADING`
*   **Subjects:** `trade.decision.*`, `trade.execution.*`
*   **Benefit:** If your InfluxDB container crashes, JetStream holds the trades in memory. When InfluxDB comes back, a Go worker reads from the stream and saves them. Zero data loss.

---

## 4. Local Infrastructure Setup (Docker Compose)

Your `docker-compose.yml` is the backbone of your local testing.

1.  **NATS:** Use the official `nats:latest` image. **Crucial:** You must enable JetStream by passing the command `-js` to the container.
2.  **ChromaDB:** Used for storing the embedded PRISM news articles (`chromadb/chroma:latest`).
3.  **InfluxDB:** Used for long-term storage of price ticks and trade PnL (`influxdb:2.7`).

### Startup Workflow for a Dev:
1.  `docker-compose up -d`
2.  Start Ollama locally: `ollama run llama3.1:8b`
3.  Run the Go orchestrator: `go run cmd/trader/main.go`

---
*By treating PRISM as your "Signal Oracle" and Kraken CLI as your "Execution Tool", you perfectly align with the hackathon's technical requirements while offloading the hardest math to your partners.*
