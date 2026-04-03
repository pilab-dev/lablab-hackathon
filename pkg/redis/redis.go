package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Client wraps go-redis with convenience methods for the trading pipeline
type Client struct {
	rdb *redis.Client
}

// NewClient creates and pings a Redis connection
func NewClient(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	log.Info().Str("addr", addr).Int("db", db).Msg("Connected to Redis")
	return &Client{rdb: rdb}, nil
}

// Conn returns the underlying redis.Client for advanced usage
func (c *Client) Conn() *redis.Client {
	return c.rdb
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}
