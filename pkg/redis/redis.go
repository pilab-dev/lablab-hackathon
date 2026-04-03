package redis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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

// PushTick appends a tick to the pair's Redis list (LPUSH) and trims to maxLen (LTRIM)
func (c *Client) PushTick(ctx context.Context, pair string, bid, ask, last, volume float64, maxLen int) error {
	key := tickKey(pair)
	member := fmt.Sprintf("%f:%f:%f:%f", bid, ask, last, volume)

	pipe := c.rdb.Pipeline()
	pipe.LPush(ctx, key, member)
	pipe.LTrim(ctx, key, 0, int64(maxLen-1))
	_, err := pipe.Exec(ctx)
	return err
}

// GetLastTick returns the most recent tick for a pair
func (c *Client) GetLastTick(ctx context.Context, pair string) (bid, ask, last, volume float64, ok bool, err error) {
	val, err := c.rdb.LIndex(ctx, tickKey(pair), -1).Result()
	if err == redis.Nil {
		return 0, 0, 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, 0, 0, false, err
	}

	bid, ask, last, volume, err = parseTick(val)
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	return bid, ask, last, volume, true, nil
}

// GetPrevTick returns the second-most recent tick for a pair
func (c *Client) GetPrevTick(ctx context.Context, pair string) (bid, ask, last, volume float64, ok bool, err error) {
	val, err := c.rdb.LIndex(ctx, tickKey(pair), -2).Result()
	if err == redis.Nil {
		return 0, 0, 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, 0, 0, false, err
	}

	bid, ask, last, volume, err = parseTick(val)
	if err != nil {
		return 0, 0, 0, 0, false, err
	}
	return bid, ask, last, volume, true, nil
}

// GetWindow returns all ticks in the window for a pair (oldest first)
func (c *Client) GetWindow(ctx context.Context, pair string) ([]string, error) {
	return c.rdb.LRange(ctx, tickKey(pair), 0, -1).Result()
}

// WindowLen returns the number of ticks in the window
func (c *Client) WindowLen(ctx context.Context, pair string) (int64, error) {
	return c.rdb.LLen(ctx, tickKey(pair)).Result()
}

// RecordTrade stores a trade in a Redis sorted set for drawdown tracking
func (c *Client) RecordTrade(ctx context.Context, pair, action string, price, sizePct, pnl float64) error {
	key := tradeKey(pair)
	member := fmt.Sprintf("%f:%f:%f:%s", price, sizePct, pnl, action)
	score := float64(time.Now().UnixMilli())
	return c.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// TotalPnL sums all trade PnL across all pairs
func (c *Client) TotalPnL(ctx context.Context) (float64, error) {
	keys, err := c.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, key := range keys {
		members, err := c.rdb.ZRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, m := range members {
			parts := strings.SplitN(m, ":", 4)
			if len(parts) >= 3 {
				pnl := parseFloat(parts[2])
				total += pnl
			}
		}
	}
	return total, nil
}

// RecentPnL returns PnL over the last N trades across all pairs
func (c *Client) RecentPnL(ctx context.Context, n int) (float64, error) {
	keys, err := c.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	type timedPnL struct {
		score float64
		pnl   float64
	}
	var all []timedPnL

	for _, key := range keys {
		members, err := c.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, z := range members {
			member, _ := z.Member.(string)
			parts := strings.SplitN(member, ":", 4)
			if len(parts) >= 3 {
				all = append(all, timedPnL{score: z.Score, pnl: parseFloat(parts[2])})
			}
		}
	}

	// Sort by score descending (most recent first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].score > all[j].score
	})

	total := 0.0
	count := 0
	for _, tp := range all {
		if count >= n {
			break
		}
		total += tp.pnl
		count++
	}
	return total, nil
}

// TradeCount returns total number of recorded trades
func (c *Client) TradeCount(ctx context.Context) (int64, error) {
	keys, err := c.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, key := range keys {
		n, err := c.rdb.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}
		total += n
	}
	return total, nil
}

// LastTradeTime returns the timestamp of the most recent trade
func (c *Client) LastTradeTime(ctx context.Context) (time.Time, bool, error) {
	keys, err := c.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return time.Time{}, false, err
	}

	latest := float64(0)
	for _, key := range keys {
		members, err := c.rdb.ZRangeWithScores(ctx, key, -1, -1).Result()
		if err != nil || len(members) == 0 {
			continue
		}
		if members[0].Score > latest {
			latest = members[0].Score
		}
	}

	if latest == 0 {
		return time.Time{}, false, nil
	}

	sec := int64(latest / 1000)
	msec := int64(latest) % 1000
	return time.Unix(sec, msec*1e6), true, nil
}

// SetRateLimit sets a cooldown key with TTL (for rate limiting)
func (c *Client) SetRateLimit(ctx context.Context, pair string, ttl time.Duration) error {
	key := cooldownKey(pair)
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsRateLimited checks if a pair is in cooldown
func (c *Client) IsRateLimited(ctx context.Context, pair string) (bool, error) {
	_, err := c.rdb.Get(ctx, cooldownKey(pair)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// --- helpers ---

func tickKey(pair string) string     { return "ticks:" + pair }
func tradeKey(pair string) string    { return "trades:" + pair }
func cooldownKey(pair string) string { return "cooldown:" + pair }

func parseTick(val string) (bid, ask, last, volume float64, err error) {
	parts := strings.SplitN(val, ":", 4)
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid tick format: %s", val)
	}
	bid = parseFloat(parts[0])
	ask = parseFloat(parts[1])
	last = parseFloat(parts[2])
	volume = parseFloat(parts[3])
	return
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
