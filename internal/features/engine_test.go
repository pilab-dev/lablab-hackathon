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

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67010, 3.0, 67005, 100)
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

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67010, 3.0, 67005, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 67010, 4.0, 67020, 2.5, 67015, 120)
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

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67013.4, 3.0, 67006.7, 100)
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

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 101.0, 110)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, []string{"bullish", "neutral"}, feat.Trend)
}

func TestFeatureEngine_Compute_LiquidityClassification(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.005, 3.0, 100.0025, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "high", feat.Liquidity)

	err = st.PushTick(ctx, "ETHUSD", 100, 5.0, 100.03, 3.0, 100.015, 100)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "ETHUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "medium", feat2.Liquidity)
}

func TestFeatureEngine_ComputeAll(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67010, 3.0, 67005, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "ETHUSD", 3500, 2.0, 3505, 1.5, 3502.5, 200)
	require.NoError(t, err)

	results := fe.ComputeAll(ctx, []string{"BTCUSD", "ETHUSD", "NONEXISTENT"})

	assert.Len(t, results, 2)
	assert.Equal(t, "BTCUSD", results[0].Pair)
	assert.Equal(t, "ETHUSD", results[1].Pair)
}

func TestFeatureEngine_Compute_SMA(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 110, 50)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 120, 50)
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
		err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100, 50)
		require.NoError(t, err)
	}

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 0.0, feat.Volatility, 0.001)

	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 200, 50)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, feat2.Volatility, 0.0)
}

func TestFeatureEngine_Compute_MicroPrice(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67010, 3.0, 67005, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Greater(t, feat.MicroPrice, 0.0)

	expectedMicroPrice := (67000*3.0 + 67010*5.0) / (5.0 + 3.0)
	assert.InDelta(t, expectedMicroPrice, feat.MicroPrice, 0.001)
}

func TestFeatureEngine_Compute_OrderBookImbalance(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	err := st.PushTick(ctx, "BTCUSD", 67000, 5.0, 67010, 5.0, 67005, 100)
	require.NoError(t, err)

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 0.0, feat.OrderBookImbalance, 0.001)

	err = st.PushTick(ctx, "BTCUSD", 67000, 7.5, 67010, 2.5, 67005, 150)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)

	expectedOBI := 0.5
	assert.InDelta(t, expectedOBI, feat2.OrderBookImbalance, 0.001)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()
	rdb.Del(ctx, "ticks:OBI_TEST")

	st2 := tracker.NewStateTracker(rdb, 10)
	err = st2.PushTick(ctx, "OBI_TEST", 67000, 7.5, 67010, 2.5, 67005, 150)
	require.NoError(t, err)

	fe2 := NewFeatureEngine(st2)
	feat3, ok, err := fe2.Compute(ctx, "OBI_TEST")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 0.5, feat3.OrderBookImbalance, 0.001)
}

func TestFeatureEngine_Compute_VolumeSurge(t *testing.T) {
	st, ctx := setupTestEngine(t)
	fe := NewFeatureEngine(st)

	for i := 0; i < 5; i++ {
		err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 50)
		require.NoError(t, err)
	}

	feat, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 1.0, feat.VolumeSurge, 0.01)

	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 101, 3.0, 100.5, 150)
	require.NoError(t, err)

	feat2, ok, err := fe.Compute(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 3.0, feat2.VolumeSurge, 0.01)
}

func TestCalcMicroPrice(t *testing.T) {
	tests := []struct {
		name     string
		bid      float64
		ask      float64
		bidVol   float64
		askVol   float64
		expected float64
	}{
		{
			name:     "equal volumes",
			bid:      100.0,
			ask:      102.0,
			bidVol:   50.0,
			askVol:   50.0,
			expected: 101.0,
		},
		{
			name:     "higher bid volume pulls micro-price up",
			bid:      100.0,
			ask:      102.0,
			bidVol:   75.0,
			askVol:   25.0,
			expected: 101.5,
		},
		{
			name:     "higher ask volume pulls micro-price down",
			bid:      100.0,
			ask:      102.0,
			bidVol:   25.0,
			askVol:   75.0,
			expected: 100.5,
		},
		{
			name:     "zero volumes falls back to mid-price",
			bid:      100.0,
			ask:      102.0,
			bidVol:   0.0,
			askVol:   0.0,
			expected: 101.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcMicroPrice(tt.bid, tt.ask, tt.bidVol, tt.askVol)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestCalcOrderBookImbalance(t *testing.T) {
	tests := []struct {
		name     string
		bidVol   float64
		askVol   float64
		expected float64
	}{
		{
			name:     "equal volumes",
			bidVol:   50.0,
			askVol:   50.0,
			expected: 0.0,
		},
		{
			name:     "higher bid volume",
			bidVol:   75.0,
			askVol:   25.0,
			expected: 0.5,
		},
		{
			name:     "higher ask volume",
			bidVol:   25.0,
			askVol:   75.0,
			expected: -0.5,
		},
		{
			name:     "zero volumes",
			bidVol:   0.0,
			askVol:   0.0,
			expected: 0.0,
		},
		{
			name:     "all bid volume",
			bidVol:   100.0,
			askVol:   0.0,
			expected: 1.0,
		},
		{
			name:     "all ask volume",
			bidVol:   0.0,
			askVol:   100.0,
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcOrderBookImbalance(tt.bidVol, tt.askVol)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}
