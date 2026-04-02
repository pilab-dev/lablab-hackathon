# Kraken AI Trader — NATS Architecture & Flow

To achieve a professional-grade, highly optimized, and decoupled system, we use **NATS** for the Hot Path (in-memory routing) and **NATS JetStream** for the Cold Path (persistence, replay, and archiving to InfluxDB).

## Why NATS?
1. **Decoupling:** Components don't need to know about each other. The Collector just publishes prices; the Decision Engine just listens.
2. **Hot/Cold Split:** NATS Core provides sub-millisecond pub/sub (Hot Path). NATS JetStream automatically persists these messages for historical analysis and dashboard loading (Cold Path).
3. **Resilience:** If the database (InfluxDB) goes down, JetStream queues the data until it returns. No data loss.

---

## 1. System Architecture Diagram

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

---

## 2. NATS Subjects & JetStream Configuration

We organize data into functional topics (Subjects). JetStream captures specific subjects into Streams.

### Subjects (NATS Core)
| Subject Pattern | Publisher | Description |
|-----------------|-----------|-------------|
| `market.tick.{pair}` | Market Collector | Live price updates from Kraken WS |
| `news.crypto` | News Crawler | New articles from PRISM / RSS |
| `signal.prism.{pair}`| News Crawler | Technical signals from PRISM API |
| `trade.decision.{pair}`| Decision Engine | LLM-generated trade decisions |
| `trade.execution.{pair}`| Trade Executor| Actual completed trades (success/fail) |

### Streams (NATS JetStream - Cold Path)
| Stream Name | Subjects | Retention Policy | Purpose |
|-------------|----------|------------------|---------|
| `MARKET` | `market.tick.*` | Limits: 24 hours | Replay recent history on boot |
| `INTELLIGENCE`| `news.*`, `signal.*` | Limits: 7 days | Store context for backtesting |
| `TRADING` | `trade.*` | WorkQueue / Infinite | PnL tracking, dashboard load, auditing |

---

## 3. Component Data Flow Diagrams

### Flow A: Market Data (The High-Frequency Loop)
```text
[Kraken WebSockets]
       │
       ▼ (NDJSON)
[Market Collector] ───(publishes `market.tick.BTCUSD`)───> [NATS Core]
                                                               │
        ┌──────────────────────────────────────────────────────┴──┐
        ▼                                                         ▼
[Decision Engine]                                          [NATS JetStream `MARKET`]
(Updates in-memory RAM state,                                     │
 checks for sudden spikes/triggers)                               ▼
                                                           [Data Archiver]
                                                           (Writes batch to InfluxDB)
```

### Flow B: The AI Decision & Execution Loop
```text
[Decision Engine] ──(Detects trigger: Price Spike + Bullish Signal)──┐
                                                                     │
        ┌─(Queries Ollama via RAM State)─────────────────────────────┘
        │
        ▼ (LLM JSON Response)
[Decision Engine] ───(publishes `trade.decision.BTCUSD`)───> [NATS Core]
                                                               │
        ┌──────────────────────────────────────────────────────┴──┐
        ▼                                                         ▼
[Trade Executor]                                           [NATS JetStream `TRADING`]
(Validates risk limits,                                           │
 executes via Kraken CLI)                                         ▼
        │                                                  [Data Archiver]
        ▼                                                  (Writes to InfluxDB)
(Executes Trade)
        │
        └───(publishes `trade.execution.BTCUSD`)───> [NATS Core] ---> [Dashboard API]
```

### Flow C: The Dashboard Sync (Go Backend to Next.js)
```text
[NATS JetStream] (History) & [NATS Core] (Live)
       │
       ▼
[Go API Server]
       │
       ├── GET /api/portfolio  (Fetches historical states from JetStream/InfluxDB)
       ├── GET /api/decisions  (Fetches historical trades from JetStream)
       │
       └── WebSocket /ws
             │ (Pushes `market.tick.*` and `trade.execution.*` events live)
             ▼
[Next.js Frontend] (Updates Zustand state, triggers React re-renders)
```

---

## 4. Why This Architecture Wins Hackathons
1. **Fault Tolerance:** You can restart the Decision Engine or the API Server, and they will instantly catch up on missed trades by reading from JetStream.
2. **Microservices Ready:** The architecture natively supports splitting the bot into multiple containers (Collector Container, Engine Container, API Container) without changing code.
3. **"Look under the hood" appeal:** Judges love seeing professional-grade messaging queues (NATS) instead of messy shared memory or direct DB polling.
