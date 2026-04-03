package tracker

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Tick represents a single market data point
type Tick struct {
	Bid       float64
	BidVolume float64
	Ask       float64
	AskVolume float64
	Last      float64
	Volume    float64
}

// TradeRecord stores executed trades for drawdown tracking
type TradeRecord struct {
	Time    time.Time
	Pair    string
	Action  string
	Price   float64
	SizePct float64
	PnL     float64
}

// StateTracker manages per-pair tick windows backed by Redis
type StateTracker struct {
	rdb    *redis.Client
	maxLen int
}

// NewStateTracker creates a Redis-backed state tracker
func NewStateTracker(rdb *redis.Client, maxWindow int) *StateTracker {
	return &StateTracker{
		rdb:    rdb,
		maxLen: maxWindow,
	}
}

// PushTick adds a tick to the pair's Redis list (LPUSH + LTRIM)
func (st *StateTracker) PushTick(ctx context.Context, pair string, bid, bidVol, ask, askVol, last, volume float64) error {
	key := tickKey(pair)
	member := encodeTick(bid, bidVol, ask, askVol, last, volume)

	pipe := st.rdb.Pipeline()
	pipe.LPush(ctx, key, member)
	pipe.LTrim(ctx, key, 0, int64(st.maxLen-1))
	_, err := pipe.Exec(ctx)
	return err
}

// GetLastTick returns the most recent tick for a pair
func (st *StateTracker) GetLastTick(ctx context.Context, pair string) (Tick, bool, error) {
	val, err := st.rdb.LIndex(ctx, tickKey(pair), 0).Result()
	if err == redis.Nil {
		return Tick{}, false, nil
	}
	if err != nil {
		return Tick{}, false, err
	}

	tick, err := decodeTick(val)
	if err != nil {
		return Tick{}, false, err
	}
	return tick, true, nil
}

// GetPrevTick returns the second-most recent tick
func (st *StateTracker) GetPrevTick(ctx context.Context, pair string) (Tick, bool, error) {
	val, err := st.rdb.LIndex(ctx, tickKey(pair), 1).Result()
	if err == redis.Nil {
		return Tick{}, false, nil
	}
	if err != nil {
		return Tick{}, false, err
	}

	tick, err := decodeTick(val)
	if err != nil {
		return Tick{}, false, err
	}
	return tick, true, nil
}

// Window returns all ticks in the window (oldest first)
func (st *StateTracker) Window(ctx context.Context, pair string) ([]Tick, error) {
	vals, err := st.rdb.LRange(ctx, tickKey(pair), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	ticks := make([]Tick, 0, len(vals))
	for _, v := range vals {
		t, err := decodeTick(v)
		if err != nil {
			continue
		}
		ticks = append(ticks, t)
	}
	return ticks, nil
}

// WindowLen returns the number of ticks in the window
func (st *StateTracker) WindowLen(ctx context.Context, pair string) (int64, error) {
	return st.rdb.LLen(ctx, tickKey(pair)).Result()
}

// SMA calculates the simple moving average of Last prices
func (st *StateTracker) SMA(ctx context.Context, pair string) (float64, bool, error) {
	ticks, err := st.Window(ctx, pair)
	if err != nil {
		return 0, false, err
	}
	if len(ticks) == 0 {
		return 0, false, nil
	}

	sum := 0.0
	for _, t := range ticks {
		sum += t.Last
	}
	return sum / float64(len(ticks)), true, nil
}

// Volatility calculates the standard deviation of Last prices
func (st *StateTracker) Volatility(ctx context.Context, pair string) (float64, bool, error) {
	ticks, err := st.Window(ctx, pair)
	if err != nil {
		return 0, false, err
	}
	n := len(ticks)
	if n < 2 {
		return 0, false, nil
	}

	mean := 0.0
	for _, t := range ticks {
		mean += t.Last
	}
	mean /= float64(n)

	variance := 0.0
	for _, t := range ticks {
		diff := t.Last - mean
		variance += diff * diff
	}
	variance /= float64(n)

	return stdDev(variance), true, nil
}

// PriceChangePct returns the percentage change between oldest and newest tick
func (st *StateTracker) PriceChangePct(ctx context.Context, pair string) (float64, bool, error) {
	ticks, err := st.Window(ctx, pair)
	if err != nil {
		return 0, false, err
	}
	if len(ticks) < 2 {
		return 0, false, nil
	}

	newest := ticks[0].Last
	oldest := ticks[len(ticks)-1].Last

	if oldest == 0 {
		return 0, false, nil
	}
	return ((newest - oldest) / oldest) * 100, true, nil
}

// VolumeDelta returns the difference between the last two volume readings
func (st *StateTracker) VolumeDelta(ctx context.Context, pair string) (float64, bool, error) {
	ticks, err := st.Window(ctx, pair)
	if err != nil {
		return 0, false, err
	}
	if len(ticks) < 2 {
		return 0, false, nil
	}

	return ticks[0].Volume - ticks[1].Volume, true, nil
}

// VolumeSurge returns the ratio of current volume to average volume in the window
func (st *StateTracker) VolumeSurge(ctx context.Context, pair string) (float64, bool, error) {
	ticks, err := st.Window(ctx, pair)
	if err != nil {
		return 0, false, err
	}
	if len(ticks) < 2 {
		return 0, false, nil
	}

	currentVol := ticks[0].Volume

	sum := 0.0
	for _, t := range ticks[1:] {
		sum += t.Volume
	}
	avgVol := sum / float64(len(ticks)-1)

	if avgVol == 0 {
		return 0, false, nil
	}

	return currentVol / avgVol, true, nil
}

// RecordTrade stores a trade in a Redis sorted set for drawdown tracking
func (st *StateTracker) RecordTrade(ctx context.Context, tr TradeRecord) error {
	key := tradeKey(tr.Pair)
	member := fmt.Sprintf("%.6f:%.6f:%.6f:%s", tr.Price, tr.SizePct, tr.PnL, tr.Action)
	score := float64(time.Now().UnixMilli())
	return st.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// TotalPnL returns the sum of all recorded trade PnL across all pairs
func (st *StateTracker) TotalPnL(ctx context.Context) (float64, error) {
	keys, err := st.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, key := range keys {
		members, err := st.rdb.ZRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, m := range members {
			parts := strings.SplitN(m, ":", 4)
			if len(parts) >= 3 {
				pnl, _ := strconv.ParseFloat(parts[2], 64)
				total += pnl
			}
		}
	}
	return total, nil
}

// RecentPnL returns PnL over the last N trades across all pairs
func (st *StateTracker) RecentPnL(ctx context.Context, n int) (float64, error) {
	keys, err := st.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	type timedPnL struct {
		score float64
		pnl   float64
	}
	var all []timedPnL

	for _, key := range keys {
		members, err := st.rdb.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, z := range members {
			member, _ := z.Member.(string)
			parts := strings.SplitN(member, ":", 4)
			if len(parts) >= 3 {
				pnl, _ := strconv.ParseFloat(parts[2], 64)
				all = append(all, timedPnL{score: z.Score, pnl: pnl})
			}
		}
	}

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
func (st *StateTracker) TradeCount(ctx context.Context) (int64, error) {
	keys, err := st.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, key := range keys {
		n, err := st.rdb.ZCard(ctx, key).Result()
		if err != nil {
			continue
		}
		total += n
	}
	return total, nil
}

// LastTradeTime returns the timestamp of the most recent trade
func (st *StateTracker) LastTradeTime(ctx context.Context) (time.Time, bool, error) {
	keys, err := st.rdb.Keys(ctx, "trades:*").Result()
	if err != nil {
		return time.Time{}, false, err
	}

	latest := float64(0)
	for _, key := range keys {
		members, err := st.rdb.ZRangeWithScores(ctx, key, -1, -1).Result()
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

// SetRateLimit sets a cooldown key with TTL
func (st *StateTracker) SetRateLimit(ctx context.Context, pair string, ttl time.Duration) error {
	return st.rdb.Set(ctx, cooldownKey(pair), "1", ttl).Err()
}

// IsRateLimited checks if a pair is in cooldown
func (st *StateTracker) IsRateLimited(ctx context.Context, pair string) (bool, error) {
	_, err := st.rdb.Get(ctx, cooldownKey(pair)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// --- helpers ---

func tickKey(pair string) string     { return "ticks:" + pair }
func tradeKey(pair string) string    { return "trades:" + pair }
func cooldownKey(pair string) string { return "cooldown:" + pair }

func encodeTick(bid, bidVol, ask, askVol, last, volume float64) string {
	return fmt.Sprintf("%.6f:%.6f:%.6f:%.6f:%.6f:%.6f", bid, bidVol, ask, askVol, last, volume)
}

func decodeTick(val string) (Tick, error) {
	parts := strings.SplitN(val, ":", 6)
	if len(parts) != 6 {
		return Tick{}, fmt.Errorf("invalid tick format: %s", val)
	}
	bid, _ := strconv.ParseFloat(parts[0], 64)
	bidVol, _ := strconv.ParseFloat(parts[1], 64)
	ask, _ := strconv.ParseFloat(parts[2], 64)
	askVol, _ := strconv.ParseFloat(parts[3], 64)
	last, _ := strconv.ParseFloat(parts[4], 64)
	volume, _ := strconv.ParseFloat(parts[5], 64)
	return Tick{Bid: bid, BidVolume: bidVol, Ask: ask, AskVolume: askVol, Last: last, Volume: volume}, nil
}

func stdDev(variance float64) float64 {
	if variance < 0 {
		return 0
	}
	return math.Sqrt(variance)
}
