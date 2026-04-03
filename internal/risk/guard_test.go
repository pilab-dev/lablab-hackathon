package risk

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kraken-trader/internal/tracker"
)

func setupTestGuard(t *testing.T) (*tracker.StateTracker, context.Context) {
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
	keys, _ = rdb.Keys(ctx, "cooldown:*").Result()
	for _, k := range keys {
		rdb.Del(ctx, k)
	}

	st := tracker.NewStateTracker(rdb, 10)
	t.Cleanup(func() {
		rdb.Close()
	})
	return st, ctx
}

func TestDefaultRiskConfig(t *testing.T) {
	cfg := DefaultRiskConfig()
	assert.Equal(t, 5.0, cfg.MaxDrawdownPct)
	assert.Equal(t, 10.0, cfg.MaxPositionSizePct)
	assert.Equal(t, 5, cfg.MaxTradesPerMinute)
	assert.InDelta(t, 0.7, cfg.MinConfidence, 0.001)
	assert.InDelta(t, 0.05, cfg.MaxSlippagePct, 0.001)
	assert.Equal(t, 30*time.Second, cfg.CooldownDuration)
}

func TestRiskGuard_Check_AllPassed(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := DefaultRiskConfig()
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 5.0, 0.9, 100.005)
	assert.True(t, result.Approved)
	assert.Contains(t, result.Reasons[0], "all risk checks passed")
	assert.InDelta(t, 10.0, result.MaxSizePct, 0.001)
}

func TestRiskGuard_Check_LowConfidence(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := DefaultRiskConfig()
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 5.0, 0.5, 100.005)
	assert.False(t, result.Approved)
	assert.Contains(t, result.Reasons[0], "confidence")
}

func TestRiskGuard_Check_Drawdown(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := DefaultRiskConfig()
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)

	err = st.RecordTrade(ctx, tracker.TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 100, SizePct: 10, PnL: -6.0})
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 5.0, 0.9, 100.005)
	assert.False(t, result.Approved)
	assert.Contains(t, result.Reasons[0], "drawdown")
}

func TestRiskGuard_Check_Cooldown(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := RiskConfig{
		MaxDrawdownPct:     5.0,
		MaxPositionSizePct: 10.0,
		MaxTradesPerMinute: 5,
		MinConfidence:      0.5,
		MaxSlippagePct:     0.1,
		CooldownDuration:   1 * time.Minute,
	}
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)

	err = st.RecordTrade(ctx, tracker.TradeRecord{Pair: "BTCUSD", Action: "buy", Price: 100, SizePct: 10, PnL: 1.0})
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 5.0, 0.9, 100.005)
	assert.False(t, result.Approved)
	assert.Contains(t, result.Reasons[0], "cooldown")
}

func TestRiskGuard_Check_PositionSizeCap(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := DefaultRiskConfig()
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.01, 3.0, 100.005, 100)
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 15.0, 0.9, 100.005)
	assert.True(t, result.Approved)
	assert.InDelta(t, 10.0, result.MaxSizePct, 0.001)
}

func TestRiskGuard_Check_Slippage(t *testing.T) {
	st, ctx := setupTestGuard(t)
	cfg := RiskConfig{
		MaxDrawdownPct:     5.0,
		MaxPositionSizePct: 10.0,
		MaxTradesPerMinute: 5,
		MinConfidence:      0.5,
		MaxSlippagePct:     0.01,
		CooldownDuration:   0,
	}
	rg := NewRiskGuard(cfg, st)

	err := st.PushTick(ctx, "BTCUSD", 100, 5.0, 100.05, 3.0, 100.025, 100)
	require.NoError(t, err)

	result := rg.Check(ctx, "BTCUSD", "buy", 5.0, 0.9, 100.025)
	assert.False(t, result.Approved)
	assert.Contains(t, result.Reasons[0], "slippage")
}
