# Kraken AI Trader — Next.js Dashboard & Sync Docs

## Tech Stack Overview
- **Frontend Framework:** Next.js 14 (App Router)
- **Styling:** Tailwind CSS + shadcn/ui
- **State Management:** Zustand (for live WebSocket data)
- **Data Fetching:** SWR or React Query (for historical REST data)
- **Charts:** Recharts
- **Go Backend API:** Fiber or standard `net/http` with `gorilla/websocket`

---

## 1. Directory Structure

The Next.js dashboard will live in the `web/` directory at the root of the Go project.

```text
kraken-trader/
├── cmd/
├── internal/
├── web/                   <-- Next.js Application
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx       (Main Dashboard)
│   ├── components/
│   │   ├── ui/            (shadcn components)
│   │   ├── charts/        (Recharts components)
│   │   ├── live-ticker.tsx
│   │   ├── decision-feed.tsx
│   ├── lib/
│   │   ├── store.ts       (Zustand store for WS data)
│   │   ├── utils.ts
│   ├── package.json
```

---

## 2. API & Sync Flow (Go ↔ Next.js)

The Go backend serves both REST endpoints (for initial load) and a WebSocket (for live updates).

### Step 1: Initial Load (REST)
When the Next.js page loads, it fetches the historical context.
- `GET /api/portfolio`: Returns current equity, fiat balance, and open positions.
- `GET /api/decisions`: Returns the last 50 LLM decisions and trade executions.
- `GET /api/chart/pnl`: Returns daily/hourly PnL snapshots for the Recharts graph.

### Step 2: Live Sync (WebSockets)
Next.js opens a WebSocket connection to `ws://localhost:8080/ws`.
The Go backend subscribes to **NATS** (`market.tick.*` and `trade.execution.*`) and forwards these events down the WebSocket to the browser.

#### WebSocket Message Formats (JSON)
```json
// Event: Market Tick
{
  "type": "TICK",
  "data": {
    "pair": "BTCUSD",
    "price": 67250.00,
    "change_24h": 2.5
  }
}

// Event: Trade Executed
{
  "type": "TRADE",
  "data": {
    "pair": "BTCUSD",
    "action": "buy",
    "size": 0.01,
    "price": 67250.00,
    "reasoning": "Bullish breakout detected by PRISM...",
    "pnl_impact": 0.00
  }
}
```

---

## 3. State Management (Zustand)

We use Zustand in Next.js to handle the high-frequency WebSocket updates without re-rendering the entire DOM tree.

```typescript
import { create } from 'zustand'

interface TraderState {
  prices: Record<string, number>;
  portfolioValue: number;
  recentTrades: Trade[];
  updatePrice: (pair: string, price: number) => void;
  addTrade: (trade: Trade) => void;
}

export const useTraderStore = create<TraderState>((set) => ({
  prices: {},
  portfolioValue: 10000,
  recentTrades: [],
  updatePrice: (pair, price) => 
    set((state) => ({ prices: { ...state.prices, [pair]: price } })),
  addTrade: (trade) => 
    set((state) => ({ recentTrades: [trade, ...state.recentTrades].slice(0, 50) })),
}))
```

---

## 4. WebSocket Hook Integration

A custom React hook manages the connection to the Go backend.

```typescript
// hooks/useLiveSync.ts
import { useEffect } from 'react';
import { useTraderStore } from '@/lib/store';

export function useLiveSync() {
  const { updatePrice, addTrade } = useTraderStore();

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === 'TICK') {
        updatePrice(msg.data.pair, msg.data.price);
      } else if (msg.type === 'TRADE') {
        addTrade(msg.data);
      }
    };

    return () => ws.close();
  }, []);
}
```

---

## 5. UI/UX Recommendations for the Hackathon

1. **Terminal Aesthetic:** Use `bg-slate-950` with text colors like `text-green-400` and `text-emerald-500`.
2. **Flash Effects:** When a `TICK` event updates a price, flash the background of that specific component (green for up, red for down) for 300ms using CSS transitions.
3. **Reasoning Transparency:** Make the LLM's `reasoning` field highly visible. It's the core innovation of the project.
4. **Partner Logos:** Ensure Kraken and Strykr (PRISM) logos are visible on the dashboard to hit the grading criteria for tech utilization.