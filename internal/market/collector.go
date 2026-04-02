package market

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"kraken-trader/internal/state"
	"kraken-trader/internal/storage"
	"kraken-trader/pkg/kraken"
)

// WSTickerData represents the JSON object streamed by `kraken ws ticker`
// The output of `kraken ws ticker` looks like this:
// {"channel":"ticker","type":"update","data":[{"symbol":"BTC/USD","bid":67234.0,"ask":67234.1,"last":67234.1,"volume":1234.56}]}
type WSTickerData struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol string  `json:"symbol"`
		Bid    float64 `json:"bid"`
		Ask    float64 `json:"ask"`
		Last   float64 `json:"last"`
		Volume float64 `json:"volume"`
	} `json:"data"`
}

// Collector manages the polling of market data from Kraken
type Collector struct {
	cli   *kraken.Client
	db    *storage.Client
	state *state.MemoryManager
	pairs []string
}

// NewCollector initializes a new market data collector using WebSockets
func NewCollector(cli *kraken.Client, db *storage.Client, stateMgr *state.MemoryManager, pairs []string) *Collector {
	// Format pairs for the WebSocket (e.g., BTC/USD instead of BTCUSD)
	wsPairs := make([]string, len(pairs))
	for i, p := range pairs {
		if !strings.Contains(p, "/") {
			wsPairs[i] = p[:3] + "/" + p[3:] // Hacky, but works for BTCUSD -> BTC/USD
		} else {
			wsPairs[i] = p
		}
	}

	return &Collector{
		cli:   cli,
		db:    db,
		state: stateMgr,
		pairs: wsPairs,
	}
}

// Start begins the WebSocket stream. It blocks until the context is canceled.
func (c *Collector) Start(ctx context.Context) {
	log.Printf("Starting WebSocket Market Data Collector for pairs: %v", c.pairs)

	args := append([]string{"ws", "ticker"}, c.pairs...)

	// Start the continuous stream. This callback fires for every line of JSON received.
	err := c.cli.RunStream(ctx, c.handleWSTick, args...)
	if err != nil {
		log.Printf("WebSocket stream failed or context canceled: %v", err)
	}
}

func (c *Collector) handleWSTick(line []byte) {
	var tick WSTickerData
	if err := json.Unmarshal(line, &tick); err != nil {
		// Ignore parse errors from heartbeat/system messages
		return
	}

	// Only process valid ticker updates
	if tick.Channel != "ticker" || tick.Type != "update" || len(tick.Data) == 0 {
		return
	}

	ts := time.Now()

	for _, d := range tick.Data {
		pair := strings.ReplaceAll(d.Symbol, "/", "") // Convert BTC/USD back to BTCUSD

		// 1. HOT PATH: Update In-Memory State instantly
		if c.state != nil {
			c.state.UpdateTick(pair, d.Bid, d.Ask, d.Last)
		}

		// 2. COLD PATH: Async Write to InfluxDB for Dashboards/Backtesting
		if c.db != nil {
			point := storage.TickerPoint{
				Pair:      pair,
				Ask:       d.Ask,
				Bid:       d.Bid,
				Last:      d.Last,
				Volume:    d.Volume,
				Timestamp: ts,
			}

			// Fire and forget (don't block the WebSocket reader)
			go func(p storage.TickerPoint) {
				// Use a quick timeout context for the DB write
				dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				measurement, tags, fields, t := p.ToPointData()
				if err := c.db.WritePoint(dbCtx, measurement, tags, fields, t); err != nil {
					log.Printf("Error writing ticker to DB for %s: %v", p.Pair, err)
				}
			}(point)
		}
	}
}
