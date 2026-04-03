package features

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kraken-trader/internal/tracker"
)

func setupTestEngine(t *testing.T) (*tracker.StateTracker, context.Context) {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available")
	}

	keys, _ := rdb.Keys(ctx, "ticks:*").Result()
	for _, k := range keys {
		rdb.Del(ctx, k)
	}

	st := tracker.NewStateTracker(rdb, 10)
	t.Cleanup(func() {
		rdb.Close()
	})
	return st, ctx
}

func TestNewFeatureEngine(t *testing.T) {
	st, _ := setupTestEngine(t)
	fe := NewFeatureEngine(st)
	assert.NotNil(t, fe)
}

func TestFeatureEngine_Compute_NoData(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	_, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestFeatureEngine_Compute_SingleTick(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 67010, 67005, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "BTCUSD", feat.Pair)
	assert.InDelta(t, 10.0, feat.Spread, 0.001)
	assert.InDelta(t, 67005.0, feat.MidPrice, 0.001)
	assert.InDelta(t, 0.0, feat.Momentum, 0.001)
	assert.InDelta(t, 0.0, feat.MomentumPct, 0.001)
	assert.InDelta(t, 0.0, feat.VolumeDelta, 0.001)
	assert.Equal(t, "neutral", feat.Trend)
}

func TestFeatureEngine_Compute_TwoTicks(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 67010, 67005, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 67010, 67020, 67015, 120)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 10.0, feat.Momentum, 0.001)
	assert.Greater(t, feat.MomentumPct, 0.0)
	assert.InDelta(t, 20.0, feat.VolumeDelta, 0.001)
}

func TestFeatureEngine_Compute_SpreadPct(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 67013.4, 67006.7, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 13.4, feat.Spread, 0.001)
	assert.InDelta(t, 67006.7, feat.MidPrice, 0.001)

	expectedSpreadPct := (13.4 / 67006.7) * 100
	assert.InDelta(t, expectedSpreadPct, feat.SpreadPct, 0.001)
}

func TestFeatureEngine_Compute_TrendClassification(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 100, 101, 100.5, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 101, 101.0, 110)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, []string{"bullish", "neutral"}, feat.Trend)
}

func TestFeatureEngine_Compute_LiquidityClassification(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 100, 100.005, 100.0025, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "high", feat.Liquidity)

	err = st.PushTick(ctx, "ETHUSD", 100, 100.03, 100.015, 100)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "ETHUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "medium", feat2.Liquidity)
}

func TestFeatureEngine_ComputeAll(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 67010, 67005, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "ETHUSD", 3500, 3505, 3502.5, 200)
	require.NoError(t, err)

	results := fe.ComputeAll(ctx, []string{"BTCUSD", "ETHUSD", "NONEXISTENT"})

	assert.Len(t, results, 2)
	assert.Equal(t, "BTCUSD", results[0].Pair)
	assert.Equal(t, "ETHUSD", results[1].Pair)
}

func TestFeatureEngine_Compute_SMA(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 100, 101, 100, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 101, 110, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 101, 120, 50)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 110.0, feat.SMA, 0.001)
}

func TestFeatureEngine_Compute_Volatility(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	for i := 0; i < 3; i++ {
		err := st.PushTick(ctx, "BTCUSD", 100, 101, 100, 50)
		require.NoError(t, err)
	}

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 0.0, feat.Volatility, 0.001)

	err = st.PushTick(ctx, "BTCUSD", 100, 101, 200, 50)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, feat2.Volatility, 0.0)
}
