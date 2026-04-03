package storage

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// Client wraps the InfluxDB v2 Go client
type Client struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
	org      string
	bucket   string
}

// NewClient initializes a new InfluxDB client connection
func NewClient(serverURL, authToken, org, bucket string) (*Client, error) {
	if serverURL == "" || authToken == "" {
		return nil, fmt.Errorf("server URL and auth token are required for InfluxDB")
	}

	// Create a new client using the server URL and the token
	client := influxdb2.NewClient(serverURL, authToken)

	// Verify connection by fetching health status
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ready, err := client.Health(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}
	if ready.Status != "pass" {
		client.Close()
		return nil, fmt.Errorf("InfluxDB health check failed, status: %s", ready.Status)
	}

	// Get blocking write client
	writeAPI := client.WriteAPIBlocking(org, bucket)
	// Get query client
	queryAPI := client.QueryAPI(org)

	return &Client{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		org:      org,
		bucket:   bucket,
	}, nil
}

// WritePoint writes a single data point to InfluxDB synchronously
func (c *Client) WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	p := influxdb2.NewPoint(measurement, tags, fields, ts)
	err := c.writeAPI.WritePoint(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to write point to InfluxDB: %w", err)
	}
	return nil
}

// HistoricalData represents OHLC data for a time period
type HistoricalData struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// GetHistoricalData queries aggregated OHLC data for a trading pair
func (c *Client) GetHistoricalData(ctx context.Context, pair string, timeframe string, limit int) ([]HistoricalData, error) {
	// Map timeframe to Flux window period
	var window string
	switch timeframe {
	case "1h":
		window = "1h"
	case "4h":
		window = "4h"
	case "1d":
		window = "1d"
	case "1w":
		window = "7d"
	default:
		window = "1d" // default to daily
	}

	// Flux query to aggregate ticker data into OHLC candles
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -30d)
		|> filter(fn: (r) => r["_measurement"] == "ticker")
		|> filter(fn: (r) => r["pair"] == "%s")
		|> filter(fn: (r) => r["_field"] == "last" or r["_field"] == "volume")
		|> aggregateWindow(every: %s, fn: (tables=<-, column) =>
			tables
			|> reduce(
				identity: {open: 0.0, high: 0.0, low: 999999.0, close: 0.0, volume: 0.0},
				fn: (r, accumulator) => ({
					open: if accumulator.open == 0.0 then r._value else accumulator.open,
					high: max(x: r._value, y: accumulator.high),
					low: min(x: r._value, y: accumulator.low),
					close: r._value,
					volume: accumulator.volume + (if r._field == "volume" then r._value else 0.0)
				})
			)
		)
		|> sort(columns: ["_time"], desc: false)
		|> limit(n: %d)
	`, c.bucket, pair, window, limit)

	result, err := c.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical data: %w", err)
	}
	defer result.Close()

	var data []HistoricalData
	for result.Next() {
		record := result.Record()

		// Get OHLC values from the aggregated record
		values := record.Values()
		timestamp := record.Time()

		dataPoint := HistoricalData{
			Timestamp: timestamp,
			Open:      values["open"].(float64),
			High:      values["high"].(float64),
			Low:       values["low"].(float64),
			Close:     values["close"].(float64),
			Volume:    values["volume"].(float64),
		}

		data = append(data, dataPoint)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("error reading query results: %w", result.Err())
	}

	return data, nil
}

// Close gracefully shuts down the InfluxDB client
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

type TradeRecord struct {
	Pair       string
	Action     string
	OrderType  string
	Size       float64
	Price      float64
	Cost       float64
	Fee        float64
	Mode       string
	Reasoning  string
	Confidence float64
	Timestamp  time.Time
}

func (c *Client) GetTradeHistory(ctx context.Context, limit int) ([]TradeRecord, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -30d)
		|> filter(fn: (r) => r["_measurement"] == "trades")
		|> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: %d)
	`, c.bucket, limit)

	result, err := c.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query trade history: %w", err)
	}
	defer result.Close()

	var trades []TradeRecord
	for result.Next() {
		record := result.Record()
		values := record.Values()

		trade := TradeRecord{
			Timestamp: record.Time(),
		}

		if v, ok := values["pair"].(string); ok {
			trade.Pair = v
		}
		if v, ok := values["action"].(string); ok {
			trade.Action = v
		}
		if v, ok := values["type"].(string); ok {
			trade.OrderType = v
		}
		if v, ok := values["mode"].(string); ok {
			trade.Mode = v
		}
		if v, ok := values["size"].(float64); ok {
			trade.Size = v
		}
		if v, ok := values["price"].(float64); ok {
			trade.Price = v
		}
		if v, ok := values["cost"].(float64); ok {
			trade.Cost = v
		}
		if v, ok := values["fee"].(float64); ok {
			trade.Fee = v
		}
		if v, ok := values["confidence"].(float64); ok {
			trade.Confidence = v
		}
		if v, ok := values["reasoning"].(string); ok {
			trade.Reasoning = v
		}

		trades = append(trades, trade)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("error reading query results: %w", result.Err())
	}

	return trades, nil
}
