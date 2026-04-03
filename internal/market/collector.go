package market

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"kraken-trader/internal/messaging"
	"kraken-trader/internal/state"
	"kraken-trader/internal/storage"
	"kraken-trader/pkg/kraken"

	"github.com/rs/zerolog/log"
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

// WSStatus represents the status message from Kraken WebSocket
// {"channel":"status","data":[{"api_version":"v2","connection_id":...,"system":"online","version":"2.0.10"}],"type":"update"}
type WSStatus struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		APIVersion   string `json:"api_version"`
		ConnectionID int64  `json:"connection_id"`
		System       string `json:"system"`
		Version      string `json:"version"`
	} `json:"data"`
}

// WSSubscribeResponse represents the response to a subscribe request
// {"method":"subscribe","result":{"channel":"ticker","event_trigger":"trades","snapshot":true,"symbol":"BTC/USD"},"success":true,...}
type WSSubscribeResponse struct {
	Method  string `json:"method"`
	Success bool   `json:"success"`
	Result  struct {
		Channel      string `json:"channel"`
		Symbol       string `json:"symbol"`
		EventTrigger string `json:"event_trigger"`
		Snapshot     bool   `json:"snapshot"`
	} `json:"result"`
	TimeIn  time.Time `json:"time_in"`
	TimeOut time.Time `json:"time_out"`
}

// Collector manages the polling of market data from Kraken
type Collector struct {
	cli   *kraken.Client
	db    *storage.Client
	state *state.MemoryManager
	nats  *messaging.NATSClient
	pairs []string

	subscriptions map[string]bool
	subsMu        sync.RWMutex
}

// NewCollector initializes a new market data collector using WebSockets
func NewCollector(cli *kraken.Client, db *storage.Client, stateMgr *state.MemoryManager, natsClient *messaging.NATSClient, pairs []string) *Collector {
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
		cli:           cli,
		db:            db,
		state:         stateMgr,
		nats:          natsClient,
		pairs:         wsPairs,
		subscriptions: make(map[string]bool),
	}
}

// Start begins the WebSocket stream. It blocks until the context is canceled.
func (c *Collector) Start(ctx context.Context) {
	log.Info().Strs("pairs", c.pairs).Msg("Starting WebSocket Market Data Collector")

	args := append([]string{"ws", "ticker"}, c.pairs...)

	// Start the continuous stream. This callback fires for every line of JSON received.
	err := c.cli.RunStream(ctx, c.handleWSTick, args...)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket stream failed or context canceled")
	}
}

// WSMessage is a generic wrapper to determine message type before parsing
type WSMessage struct {
	Channel string          `json:"channel"`
	Method  string          `json:"method"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Result  json.RawMessage `json:"result"`
}

func (c *Collector) handleWSTick(line []byte) {
	var msg WSMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		log.Error().Str("data", string(line)).Err(err).Msg("Failed to unmarshal message")
		return
	}

	switch {
	case msg.Channel == "status":
		c.handleStatus(line)
	case msg.Method == "subscribe":
		c.handleSubscribe(line)
	case msg.Channel == "heartbeat":
		return
	case msg.Channel == "ticker" && (msg.Type == "update" || msg.Type == "snapshot"):
		c.handleTicker(line)
	default:
		log.Debug().Str("data", string(line)).Msg("Unknown message type")
	}
}

func (c *Collector) handleStatus(line []byte) {
	var status struct {
		Channel string `json:"channel"`
		Type    string `json:"type"`
		Data    []struct {
			APIVersion string `json:"api_version"`
			System     string `json:"system"`
			Version    string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(line, &status); err != nil {
		return
	}
	if len(status.Data) > 0 {
		d := status.Data[0]
		log.Info().Str("system", d.System).Str("version", d.Version).Msg("WebSocket status")
		c.publishToNATS("market.status", map[string]string{
			"system":  d.System,
			"version": d.Version,
		})
	}
}

func (c *Collector) handleSubscribe(line []byte) {
	var subResp struct {
		Method  string `json:"method"`
		Success bool   `json:"success"`
		Result  struct {
			Channel string `json:"channel"`
			Symbol  string `json:"symbol"`
		} `json:"result"`
	}
	if err := json.Unmarshal(line, &subResp); err != nil {
		return
	}
	if subResp.Success {
		log.Info().Str("channel", subResp.Result.Channel).Str("symbol", subResp.Result.Symbol).Msg("Subscribed to channel")
		c.subsMu.Lock()
		c.subscriptions[subResp.Result.Symbol] = true
		c.subsMu.Unlock()
		c.publishToNATS("market.subscribed", map[string]string{
			"channel": subResp.Result.Channel,
			"symbol":  subResp.Result.Symbol,
		})
	} else {
		log.Warn().Str("channel", subResp.Result.Channel).Str("symbol", subResp.Result.Symbol).Msg("Subscription failed")
		c.publishToNATS("market.subscribe_failed", map[string]string{
			"channel": subResp.Result.Channel,
			"symbol":  subResp.Result.Symbol,
		})
	}
}

func (c *Collector) handleTicker(line []byte) {
	var tick WSTickerData
	if err := json.Unmarshal(line, &tick); err != nil {
		log.Error().Str("data", string(line)).Err(err).Msg("Failed to unmarshal ticker")
		return
	}
	if len(tick.Data) == 0 {
		return
	}

	ts := time.Now()

	for _, d := range tick.Data {
		pair := strings.ReplaceAll(d.Symbol, "/", "")

		if c.state != nil {
			c.state.UpdateTick(pair, d.Bid, d.Ask, d.Last)
		}

		if c.db != nil {
			point := storage.TickerPoint{
				Pair:      pair,
				Ask:       d.Ask,
				Bid:       d.Bid,
				Last:      d.Last,
				Volume:    d.Volume,
				Timestamp: ts,
			}
			go func(p storage.TickerPoint) {
				dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				measurement, tags, fields, t := p.ToPointData()
				if err := c.db.WritePoint(dbCtx, measurement, tags, fields, t); err != nil {
					log.Error().Err(err).Str("pair", p.Pair).Msg("Error writing ticker to DB")
				}
			}(point)
		}

		c.publishToNATS("market.ticker", map[string]interface{}{
			"symbol": d.Symbol,
			"bid":    d.Bid,
			"ask":    d.Ask,
			"last":   d.Last,
			"volume": d.Volume,
		})
	}
}

func (c *Collector) publishToNATS(subject string, data interface{}) {
	if c.nats == nil {
		return
	}
	payload, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("Failed to marshal NATS payload")
		return
	}
	if err := c.nats.Publish(subject, payload); err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("Failed to publish to NATS")
	}
}

func (c *Collector) formatSymbol(symbol string) string {
	if !strings.Contains(symbol, "/") && len(symbol) >= 6 {
		return symbol[:3] + "/" + symbol[3:]
	}
	return symbol
}

func (c *Collector) normalizeSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "/", "")
}

func (c *Collector) ListSubscriptions() []string {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()

	result := make([]string, 0, len(c.subscriptions))
	for sym := range c.subscriptions {
		result = append(result, sym)
	}
	return result
}

func (c *Collector) IsSubscribed(symbol string) bool {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()
	return c.subscriptions[symbol]
}

func (c *Collector) AddSubscription(symbol string) bool {
	c.subsMu.Lock()
	defer c.subsMu.Unlock()

	if c.subscriptions[symbol] {
		return false
	}
	c.subscriptions[symbol] = true
	return true
}

func (c *Collector) RemoveSubscription(symbol string) bool {
	c.subsMu.Lock()
	defer c.subsMu.Unlock()

	if !c.subscriptions[symbol] {
		return false
	}
	delete(c.subscriptions, symbol)
	return true
}
