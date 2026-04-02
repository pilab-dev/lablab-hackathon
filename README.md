# 🦑 Kraken AI Trader

Autonomous AI trading agent for the **lablab.ai AI Trading Agents Hackathon**.

## 🏆 The Winning Playbook
We are competing for the **Kraken PnL Rank** and the **Social Engagement Prize**. Our strategy is documented in these four pillars:

1.  **[GO_FOR_GOLD.md](./GO_FOR_GOLD.md)**: Master strategy, PnL Alpha, and the Innovation play (ERC-8004).
2.  **[HYPE_MACHINE.md](./HYPE_MACHINE.md)**: Social engagement and "Build in Public" plan for the $1,200 prize.
3.  **[TECHNICAL_BLUEPRINT.md](./TECHNICAL_BLUEPRINT.md)**: NATS architecture, Llama 3.1 Brain, and Kraken CLI integration.
4.  **[EXECUTION_LOG.md](./EXECUTION_LOG.md)**: Live tasks, roadmap, and competition milestones.

---

## Quick Start (Dev)
1.  **Prerequisites:** Go 1.22+, Docker, Ollama (llama3.1:8b), Kraken CLI.
2.  **Setup:** `cp .env.example .env` and fill in your keys.
3.  **Run:** `make docker-up` then `go run cmd/trader/main.go`.

---

## 🏗️ Architecture Summary
- **Brain:** Local Llama 3.1 (Ollama)
- **Signals:** PRISM AI + News Sentiment
- **Execution:** Kraken CLI (MCP-ready)
- **Messaging:** NATS JetStream
- **Verification:** ERC-8004 Performance Signing

---
*Archived research and old plans can be found in the [archive/](./archive/) directory.*
