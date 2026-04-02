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

// Close gracefully shuts down the InfluxDB client
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
