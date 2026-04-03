package tracker

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestTracker(t *testing.T) (*StateTracker, context.Context) {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	// Clean up test keys
	keys, _ := rdb.Keys(ctx, "ticks:*").Result()
	for _, k := range keys {
		rdb.Del(ctx, k)
	}
	keys, _ = rdb.Keys(ctx, "trades:*").Result()
	for _, k := range keys {
		rdb.Del(ctx, k)
	}

	st := NewStateTracker(rdb, 10)
	t.Cleanup(func() {
		rdb.Close()
	})
	return st, ctx
}

func TestStateTracker_PushAndGetTick(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 50)
	require.NoError(t, err)

	tick, ok, err := st.GetLastTick(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 100.0, tick.Bid, 0.001)
	assert.InDelta(t, 5.0, tick.BidVolume, 0.001)
	assert.InDelta(t, 101.0, tick.Ask, 0.001)
	assert.InDelta(t, 3.0, tick.AskVolume, 0.001)
	assert.InDelta(t, 100.5, tick.Last, 0.001)
	assert.InDelta(t, 50.0, tick.Volume, 0.001)
}

func TestStateTracker_PrevTick(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 101, 4.0, 102, 2.5, 101.5, 60)
	require.NoError(t, err)

	tick, ok, err := st.GetPrevTick(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 100.5, tick.Last, 0.001)
}

func TestStateTracker_WindowTrims(t *testing.T) {
	st, ctx := setupTestTracker(t)

	for i := 0; i < 15; i++ {
		err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, float64(100+i), 50)
		require.NoError(t, err)
	}

	ticks, err := st.Window(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.Len(t, ticks, 10)
	assert.InDelta(t, 114.0, ticks[0].Last, 0.001)
	assert.InDelta(t, 105.0, ticks[9].Last, 0.001)
}

func TestStateTracker_WindowLen(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 50)
	require.NoError(t, err)

	n, err := st.WindowLen(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestStateTracker_SMA(t *testing.T) {
	st, ctx := setupTestTracker(t)

	_, ok, err := st.SMA(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.False(t, ok)

	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 10, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 20, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 30, 50)
	require.NoError(t, err)

	sma, ok, err := st.SMA(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 20.0, sma, 0.001)
}

func TestStateTracker_Volatility(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
	require.NoError(t, err)

	vol, ok, err := st.Volatility(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 0.0, vol, 0.001)

	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 200, 50)
	require.NoError(t, err)

	vol, ok, err = st.Volatility(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, vol, 0.0)
}

func TestStateTracker_PriceChangePct(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 110, 50)
	require.NoError(t, err)

	change, ok, err := st.PriceChangePct(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 10.0, change, 0.01)
}

func TestStateTracker_VolumeDelta(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 101, 150)
	require.NoError(t, err)

	delta, ok, err := st.VolumeDelta(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 50.0, delta, 0.001)
}

func TestStateTracker_RecordTrade(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 50000, SizePct: 5.0, PnL: 100})
	require.NoError(t, err)

	count, err := st.TradeCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestStateTracker_TotalPnL(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 50000, SizePct: 5.0, PnL: 100})
	require.NoError(t, err)
	err = st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "sell", Price: 50100, SizePct: 5.0, PnL: -50})
	require.NoError(t, err)

	total, err := st.TotalPnL(ctx)
	require.NoError(t, err)
	assert.InDelta(t, 50.0, total, 0.001)
}

func TestStateTracker_RecentPnL(t *testing.T) {
	st, ctx := setupTestTracker(t)

	err := st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 50000, SizePct: 5.0, PnL: 100})
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "sell", Price: 50100, SizePct: 5.0, PnL: -50})
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 50200, SizePct: 5.0, PnL: 200})
	require.NoError(t, err)

	recent, err := st.RecentPnL(ctx, 2)
	require.NoError(t, err)
	assert.InDelta(t, 150.0, recent, 0.001)
}

func TestStateTracker_LastTradeTime(t *testing.T) {
	st, ctx := setupTestTracker(t)

	_, ok, err := st.LastTradeTime(ctx)
	require.NoError(t, err)
	assert.False(t, ok)

	err = st.RecordTrade(ctx, TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 50000, SizePct: 5.0, PnL: 100})
	require.NoError(t, err)

	tradeTime, ok, err := st.LastTradeTime(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.False(t, tradeTime.IsZero())
}

func TestStateTracker_RateLimit(t *testing.T) {
	st, ctx := setupTestTracker(t)

	limited, err := st.IsRateLimited(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.False(t, limited)

	err = st.SetRateLimit(ctx, "BTCUSD", 1000000000)
	require.NoError(t, err)

	limited, err = st.IsRateLimited(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, limited)
}
