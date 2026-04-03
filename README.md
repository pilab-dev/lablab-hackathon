# 🦑 Kraken AI Trader

Autonomous AI trading agent for the **lablab.ai AI Trading Agents Hackathon**.

## 🏆 The Winning Playbook
We are competing for the **Kraken PnL Rank** and the **Social Engagement Prize**. Our strategy is documented in these four pillars:

1.  **[GO_FOR_GOLD.md](./GO_FOR_GOLD.md)**: Master strategy, PnL Alpha, and the Innovation play (ERC-8004).
2.  **[HYPE_MACHINE.md](./HYPE_MACHINE.md)**: Social engagement and "Build in Public" plan for the $1,200 prize.
3.  **[TECHNICAL_BLUEPRINT.md](./TECHNICAL_BLUEPRINT.md)**: NATS architecture, Llama 3.1 Brain, and Kraken CLI integration.
4.  **[EXECUTION_LOG.md](./EXECUTION_LOG.md)**: Live tasks, roadmap, and competition milestones.

---

## Setup Instructions

### 1. Install Prerequisites

| Tool | Version | Installation |
|------|---------|---------------|
| Go | 1.26+ | `brew install go` |
| Docker | Latest | `brew install --cask docker` |
| Kraken CLI | Latest | `brew install kraken-cli` |
| Ollama | Latest | `curl -s https://ollama.com/install.sh | sh` |

### 2. Configure Kraken CLI

```bash
# Set up API credentials (get from https://www.kraken.com/u/settings/api)
kraken auth set

# Verify authentication
kraken balance -o json
```

### 3. Start Ollama Models

```bash
# Start Ollama server
ollama serve &

# Pull required models
ollama pull llama3.1:8b
ollama pull nomic-embed-text
```

### 4. Configure Environment

```bash
# Copy example config
cp .env.example .env

# Edit .env with your API keys:
# - KRAKEN_API_KEY / KRAKEN_API_SECRET (or use kraken auth)
# - PRISM_API_KEY (get from https://prismapp.io)
# - OLLAMA_URL=http://localhost:11434
```

### 5. Start Docker Services

```bash
make docker-up
```

This starts:
- **InfluxDB** (http://localhost:8086) - metrics storage
- **ChromaDB** (http://localhost:8000) - vector embeddings
- **NATS** (localhost:4222) - message broker

### 6. Build & Run

```bash
# Build
make build

# Run
make run-with-env
```

Or for development with auto-reload:
```bash
go run ./cmd/trader
```

---

## Quick Start (Dev)
1.  **Prerequisites:** Go 1.26+, Docker, Ollama (llama3.1:8b), Kraken CLI.
2.  **Setup:** `cp .env.example .env` and fill in your keys.
3.  **Run:** `make docker-up` then `go run ./cmd/trader`.

---

## Data Persistence

- **SQLite** (`trader.db`): Stores subscriptions and LLM prompts/responses
- **InfluxDB**: Stores ticker price data for dashboards

Subscriptions are restored from SQLite on startup with `created_at` and `last_data` timestamps.

---

## API Endpoints (Port 8081)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/subscriptions` | List active subscriptions (symbols only) |
| GET | `/subscriptions/detail` | List subscriptions with created_at, last_data timestamps |
| POST | `/subscriptions` | Add subscription (body: `{"symbol": "BTC/USD"}`) |
| DELETE | `/subscriptions/{symbol}` | Remove subscription |
| GET | `/prompts` | List last 20 prompts with raw prompt/response |
| GET | `/loglevel` | Get current console log level |
| POST | `/loglevel` | Set console log level (body: `{"level": "trace\|debug\|info\|warn\|error"}`) |

---

## 🏗️ Architecture Summary
- **Brain:** Local Llama 3.1 (Ollama)
- **Signals:** PRISM AI + News Sentiment
- **Execution:** Kraken CLI (MCP-ready)
- **Messaging:** NATS JetStream
- **Verification:** ERC-8004 Performance Signing

---
*Archived research and old plans can be found in the [archive/](./archive/) directory.*
