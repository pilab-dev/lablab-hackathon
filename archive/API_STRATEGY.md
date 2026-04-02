# Kraken AI Trader — API Integrations & Data Strategy

## Core APIs
1. **Kraken CLI (Primary Execution & Market Data)**
   - Used for all live trades (to count for PnL).
   - Used for primary OHLCV and order book data.
2. **PRISM / Strykr API (Hackathon Partner - Signal & News Enrichment)**
   - As an official tech partner, integrating PRISM API increases our "Application of Technology" score.
   - Endpoint: `GET /signals/summary?symbols=BTC,ETH` -> AI-generated signals for momentum, volume, breakout.
   - Endpoint: `GET /news/crypto` -> Aggregated crypto news.
   - This can replace or augment our RSS crawler.

## Dashboard Updates
1. Add a **PRISM Signals** pane to show the current technical sentiment.
2. Ensure the Kraken Logo and Strykr Logo are visible on the dashboard to show partner integration.
