package storage

import "time"

// TickerPoint represents a single ticker price measurement
type TickerPoint struct {
	Pair      string
	Ask       float64
	Bid       float64
	Last      float64
	Volume    float64
	Timestamp time.Time
}

// OHLCVPoint represents a single candle measurement
type OHLCVPoint struct {
	Pair      string
	Interval  int // e.g., 1 for 1-minute, 60 for 1-hour
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time
}

// TradePoint represents a trade execution logged to the database
type TradePoint struct {
	Pair       string
	Action     string // "buy", "sell"
	OrderType  string // "market", "limit"
	Size       float64
	Price      float64
	Cost       float64
	Fee        float64
	Mode       string  // "paper", "live"
	Reasoning  string  // Brief explanation from LLM
	Confidence float64 // LLM confidence score
	Timestamp  time.Time
}

// TickerToFields converts a TickerPoint to InfluxDB fields and tags
func (t *TickerPoint) ToPointData() (string, map[string]string, map[string]interface{}, time.Time) {
	tags := map[string]string{
		"pair": t.Pair,
	}
	fields := map[string]interface{}{
		"ask":    t.Ask,
		"bid":    t.Bid,
		"last":   t.Last,
		"volume": t.Volume,
	}
	return "ticker", tags, fields, t.Timestamp
}

// TradeToFields converts a TradePoint to InfluxDB fields and tags
func (t *TradePoint) ToPointData() (string, map[string]string, map[string]interface{}, time.Time) {
	tags := map[string]string{
		"pair":   t.Pair,
		"action": t.Action,
		"mode":   t.Mode,
		"type":   t.OrderType,
	}
	fields := map[string]interface{}{
		"size":       t.Size,
		"price":      t.Price,
		"cost":       t.Cost,
		"fee":        t.Fee,
		"confidence": t.Confidence,
		"reasoning":  t.Reasoning,
	}
	return "trades", tags, fields, t.Timestamp
}
