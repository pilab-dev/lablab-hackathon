package strategy

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kraken-trader/internal/features"
	"kraken-trader/internal/state"
	"kraken-trader/internal/tracker"
)

func setupTestStrategy(t *testing.T) (*tracker.StateTracker, *features.FeatureEngine, *state.MemoryManager, context.Context) {
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
	keys, _ = rdb.Keys(ctx, "trades:*").Result()
	for _, k := range keys {
		rdb.Del(ctx, k)
	}

	st := tracker.NewStateTracker(rdb, 10)
	fe := features.NewFeatureEngine(st)
	sm := state.NewMemoryManager()

	t.Cleanup(func() {
		rdb.Close()
	})
	return st, fe, sm, ctx
}

func TestNewStrategyEngine(t *testing.T) {
	_, fe, sm, _ := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)
	assert.NotNil(t, se)
}

func TestStrategyEngine_Evaluate_NoData(t *testing.T) {
	_, fe, sm, ctx := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	_, ok, err := se.Evaluate(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestStrategyEngine_Evaluate_LowLiquidity(t *testing.T) {
	st, fe, sm, ctx := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 110, 3.0, 105, 100)
	require.NoError(t, err)

	result, ok, err := se.Evaluate(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, Hold, result.Signal)
	assert.Contains(t, result.Reasons[0], "spread too wide")
}

func TestStrategyEngine_Evaluate_BullishMomentum(t *testing.T) {
	st, fe, sm, ctx := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.5, 120)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 101.0, 140)
	require.NoError(t, err)

	result, ok, err := se.Evaluate(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NotEqual(t, Hold, result.Signal)
	assert.Greater(t, result.Confidence, 0.0)
	assert.NotEmpty(t, result.Reasons)
}

func TestStrategyEngine_Evaluate_BearishMomentum(t *testing.T) {
	st, fe, sm, ctx := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 101.0, 100)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.5, 80)
	require.NoError(t, err)
	err = st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.0, 60)
	require.NoError(t, err)

	result, ok, err := se.Evaluate(ctx, "BTCUSD")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NotEqual(t, Hold, result.Signal)
}

func TestStrategyEngine_scoreToSignal(t *testing.T) {
	_, fe, sm, _ := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	tests := []struct {
		score    float64
		expected SignalStrength
	}{
		{0.8, StrongBuy},
		{0.5, Buy},
		{0.2, WeakBuy},
		{0.0, Hold},
		{-0.2, WeakSell},
		{-0.5, Sell},
		{-0.8, StrongSell},
	}

	for _, tt := range tests {
		result := se.scoreToSignal(tt.score)
		assert.Equal(t, tt.expected, result)
	}
}

func TestStrategyEngine_scorePRISM(t *testing.T) {
	_, fe, sm, _ := setupTestStrategy(t)

	se := NewStrategyEngine(fe, sm)

	tests := []struct {
		name     string
		snap     state.PairState
		expected float64
	}{
		{
			name: "all bullish",
			snap: state.PairState{
				MomentumSignal: "bullish",
				BreakoutSignal: "up",
				VolumeSignal:   "strong",
			},
			expected: 1.0,
		},
		{
			name: "all bearish",
			snap: state.PairState{
				MomentumSignal: "bearish",
				BreakoutSignal: "down",
				VolumeSignal:   "weak",
			},
			expected: -0.9,
		},
		{
			name: "neutral",
			snap: state.PairState{
				MomentumSignal: "",
				BreakoutSignal: "",
				VolumeSignal:   "",
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := se.scorePRISM(tt.snap)
			assert.InDelta(t, tt.expected, score, 0.001)
		})
	}
}
