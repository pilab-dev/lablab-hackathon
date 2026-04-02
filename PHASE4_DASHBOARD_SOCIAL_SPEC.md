# 🚀 Phase 4 Technical Spec: Dashboard & Social Automation

*Targeted at Senior Developers building the UI and "Hype Machine" layer of the trading bot.*

Phase 4 is all about **Presentation** and **Social Engagement**. The AI could be a genius, but if judges can't see its reasoning or if it doesn't generate buzz, it won't win. 

This phase bridges the high-speed Go backend (NATS) to a Next.js frontend (via WebSockets) and to Twitter/X (via API).

---

## 1. The Real-Time Dashboard (Next.js 14)

The dashboard is the visual proof that your bot is actually trading and thinking. It must be fast, dark-themed (for that "pro trader" look), and reactive.

### A. The Go WebSocket Server (`internal/api/server.go`)
Before Next.js can show anything, Go needs to serve it.
*   **Concept:** Go listens to NATS and forwards events to connected browsers.
*   **Implementation:** Use `gorilla/websocket`.
*   **Channels:**
    *   `ws.WriteJSON(msg)` for every `market.tick.*` (Price updates).
    *   `ws.WriteJSON(msg)` for every `trade.execution.*` (New trades).

### B. State Management in Next.js (Zustand)
React state (`useState`) is too slow for high-frequency market ticks.
*   **Solution:** Use Zustand. It allows components to subscribe to specific store changes without re-rendering the whole page.
*   **Implementation (`web/lib/store.ts`):**
    ```typescript
    import { create } from 'zustand'

    interface TraderState {
      prices: Record<string, number>;
      portfolioValue: number;
      recentTrades: any[];
      updatePrice: (pair: string, price: number) => void;
      addTrade: (trade: any) => void;
    }

    export const useTraderStore = create<TraderState>((set) => ({
      prices: {},
      portfolioValue: 0,
      recentTrades: [],
      updatePrice: (pair, price) => set((state) => ({ 
        prices: { ...state.prices, [pair]: price } 
      })),
      addTrade: (trade) => set((state) => ({ 
        recentTrades: [trade, ...state.recentTrades].slice(0, 20) 
      })),
    }))
    ```

### C. The "AI Transparency" Component
Judges want to see the LLM's brain.
*   **UI Requirement:** Create a `<DecisionFeed />` component.
*   **Data:** Display the `Pair`, `Action`, `Confidence`, and crucially, the `Reasoning` string returned by Llama 3.1.
*   **Styling:** Make the reasoning text prominent. Use Tailwind: `text-emerald-400` for buys, `text-rose-400` for sells.

---

## 2. The Social Poster (`internal/social/poster.go`)

This is the engine for the $1,200 Social Engagement prize. It converts NATS events into tweets.

### A. Integration Setup
*   **API:** Use the official Twitter/X API (v2).
*   **Go Library:** `github.com/g8rswimmer/go-twitter/v2` is a standard choice.
*   **Trigger:** Subscribe to the `trade.execution.*` NATS subject.

### B. Message Templates (The "Hype" format)
To maximize the Lablab algorithm, tweets must have specific structures.

**Template 1: The Trade Execution**
```go
tweet := fmt.Sprintf("🤖 AI Agent just executed a %s on %s.\n\nReasoning: %s\n\n#KrakenCLI #AITrading #BuildInPublic @krakenfx @lablabai @Surgexyz_",
    strings.ToUpper(trade.Action), trade.Pair, trade.Reasoning)
```

**Template 2: The Daily PnL Recap**
*   **Trigger:** Run via a cron job (or `time.Ticker`) at 23:59 UTC.
*   **Content:** Net PnL, Win Rate, and the ERC-8004 cryptographic signature hash to prove authenticity.

### C. Rate Limiting & Safety
*   **Rule:** Twitter will ban you if you spam. 
*   **Implementation:** 
    *   Limit to max 1 tweet per 15 minutes.
    *   If the bot executes 5 trades in a minute, batch them into a single "Thread" or just tweet the most confident one.

---

## 3. Recommended Implementation Workflow

### Step 1: The Go API Scaffold
1.  Add `github.com/gofiber/fiber/v2` or standard `net/http` to your Go project.
2.  Create REST endpoints for historical data: `GET /api/portfolio`, `GET /api/trades`.
3.  Implement the `/ws` endpoint that upgrades the HTTP connection and listens to NATS.

### Step 2: Next.js Initialization
1.  Run `npx create-next-app@latest web` inside the repo.
2.  Install `zustand`, `lucide-react` (icons), and `recharts` (for PnL graphs).
3.  Setup `shadcn/ui` for quick, professional-looking components (Cards, Tables).

### Step 3: The WebSocket Hook
Create `web/hooks/useLiveSync.ts` to manage the WebSocket connection, handle reconnects, and push data into the Zustand store.

### Step 4: The Social Worker
1.  Create `internal/social/poster.go`.
2.  Write a simple `PostTweet(text string)` function.
3.  Add a boolean flag in `.env` (`ENABLE_SOCIAL_POSTING=false`). **Keep this false during local dev!**

---

## 🎨 UI/UX Cheat Sheet for Hackathons:
*   **Dark Mode Only:** It looks more "quant-like". (`bg-slate-950` with `border-slate-800`).
*   **Flash Updates:** When a price changes, briefly flash the background green or red. It makes the app feel "alive" during the demo video.
*   **Show the Math:** If you calculate a confidence score of 0.85, show it as an 85% progress bar. 
*   **Logos:** Put the Kraken and Strykr (PRISM) logos in the header. Judges look for partner technology usage.
