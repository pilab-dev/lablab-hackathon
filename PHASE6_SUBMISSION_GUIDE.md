# 🏁 Phase 6 Technical Spec: Submission & Hackathon Finale

*Targeted at the Team Lead / Senior Developer finalizing the Lablab.ai hackathon submission.*

Winning a hackathon isn't just about having the best code; it's about proving it. Phase 6 covers the exact steps required to submit your project by **April 12, 2026, at 9:00 AM PDT**, ensuring you hit all judging criteria for the Kraken Challenge and the Surge (ERC-8004) track.

---

## 1. The Submission Package Checklist

Before hitting "Submit" on the Lablab.ai platform, ensure you have the following assets ready.

### A. The Required Assets
- [ ] **Project Title:** "Kraken AI Trader" (or your specific team name).
- [ ] **Short Description:** "An autonomous, Llama 3.1-powered trading agent executing high-conviction trades via Kraken CLI, with on-chain performance verification."
- [ ] **Public GitHub Repo:** Ensure your repo is public. Make sure `.env` files and `kraken-cli` secret keys are **NOT** committed. Check git history to be safe.
- [ ] **Demo Application URL:** Since this is a backend bot, your "Demo URL" should link to a deployed version of your Next.js Dashboard. If you can't deploy it, link to a comprehensive README section explaining how to run it locally.
- [ ] **Cover Image:** Create a clean, professional architecture diagram (e.g., using Excalidraw) showing the flow from PRISM -> LLM -> NATS -> Kraken CLI.

---

## 2. The Pitch Video (3 Minutes Max)

Judges have to review hundreds of projects. Your video must be dense with signal and light on fluff.

### Video Structure
1. **The Hook (0:00 - 0:30):** State the problem (Trading bots are black boxes) and your solution (A transparent, LLM-reasoned agent using PRISM signals). Show the live dashboard immediately.
2. **The Tech Stack (0:30 - 1:30):** Explain the architecture. Mention the **Kraken CLI (MCP Server)**, the **NATS messaging bus** for high speed, and **Llama 3.1 running locally**. *Judges love seeing the tech partner tools used properly.*
3. **The Proof (1:30 - 2:30):** Show the bot executing a trade. Point out the LLM's reasoning on the dashboard. Briefly show the ERC-8004 cryptographic signature confirming the trade's authenticity.
4. **The Results (2:30 - 3:00):** Display your final PnL chart and your automated Twitter/X feed. End with a strong closing statement.

---

## 3. The PnL Verification (Kraken Leaderboard)

To be eligible for the Kraken PnL prize, you must prove your trades were real.

1. **Read-Only API Key:** Go to your Kraken account and generate a new API key with **strictly "Query Funds" and "Query Open Orders & Trades" permissions**.
2. **Submission:** Provide this Read-Only key in the designated submission field on the Lablab/Surge platform. This allows the judges to audit your net PnL mathematically.
3. **The Surge Registry:** Ensure your project is officially registered at `early.surge.xyz` (as per the hackathon rules) to be eligible for the prize pool.

---

## 4. The Presentation Slides

Your slide deck should reinforce the video but provide more technical depth for the judges who want to dig in.

*   **Slide 1: Title & Vision.**
*   **Slide 2: The Alpha (Strategy).** How PRISM signals + LLM sentiment creates an edge.
*   **Slide 3: Architecture Diagram.** Visualizing Go, NATS, and Kraken CLI.
*   **Slide 4: Risk Management.** Highlighting your daily loss limits and position sizing (proves you understand real trading).
*   **Slide 5: Business Value / Innovation.** Explain the ERC-8004 integration and how verifiable performance is the future of AI agents.
*   **Slide 6: Team & PnL Results.**

---

## 5. Final Code Cleanup

A clean codebase scores points for "Application of Technology".
*   Run `go fmt ./...` and `go vet ./...`.
*   Ensure your `README.md` has clear `make docker-up` and `go run` instructions.
*   Make sure your `HYPE_MACHINE.md` and `GO_FOR_GOLD.md` playbooks are visible in the repo to show your strategic thinking.

**Good luck. You have the architecture and the strategy to win.**
