# 📝 EXECUTION_LOG — Tasks & Roadmap

## Phase 0: Foundation (COMPLETED)
- [x] Architecture Design (NATS-based)
- [x] Tech Stack Selection (Go, Ollama, Kraken CLI)
- [x] Documentation restructures (Winning Playbook)

## Phase 1: Infrastructure (April 2 - April 3)
- [ ] **Registration:** Register at early.surge.xyz (**CRITICAL**)
- [ ] **Docker:** Spin up NATS, InfluxDB, ChromaDB.
- [ ] **Wrappers:** Go subprocess wrapper for `kraken-cli`.
- [ ] **Messaging:** NATS stream configuration for `MARKET`, `TRADING`, `INTELLIGENCE`.

## Phase 2: Intelligence (April 3 - April 4)
- [ ] **PRISM:** Integrate PRISM API for technical signals.
- [ ] **News:** Embedding pipeline with Ollama + ChromaDB.
- [ ] **Brain:** Llama 3.1 Decision Engine with JSON output mode.
- [ ] **Prompt Tuning:** Refine the system prompt with few-shot examples.

## Phase 3: Trading & Risk (April 4 - April 5)
- [ ] **Risk:** Implement position sizing and daily loss limits.
- [ ] **Executor:** Live trading logic via `kraken-cli`.
- [ ] **Validation:** ERC-8004 registry integration for performance signing.

## Phase 4: Testing & Tuning (April 5 - April 6)
- [ ] **Paper Test:** 48h run in `TRADING_MODE=paper`.
- [ ] **Optimization:** Tune confidence thresholds and trade frequency.

## Phase 5: LIVE & HYPE (April 6 - April 12)
- [ ] **Go Live:** Switch to `TRADING_MODE=live`.
- [ ] **Hype:** Enable `SocialPoster` for automated Twitter/X updates.
- [ ] **Threads:** Post architecture and "Why ERC-8004" threads on social.

## Phase 6: Submission (April 12)
- [ ] **Video:** 3-min demo of Dashboard + Live Trade.
- [ ] **Slides:** Final presentation deck.
- [ ] **PnL Audit:** Ensure Kraken has the read-only API key for the leaderboard.
