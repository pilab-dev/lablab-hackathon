package features

import (
	"context"
	"math"

	"kraken-trader/internal/tracker"
)

// PairFeatures holds all derived metrics for a single trading pair
type PairFeatures struct {
	Pair        string  `json:"pair"`
	Spread      float64 `json:"spread"`
	SpreadPct   float64 `json:"spread_pct"`
	MidPrice    float64 `json:"mid_price"`
	Momentum    float64 `json:"momentum"`
	MomentumPct float64 `json:"momentum_pct"`
	VolumeDelta float64 `json:"volume_delta"`
	Volatility  float64 `json:"volatility"`
	SMA         float64 `json:"sma"`
	Trend       string  `json:"trend"`
	Liquidity   string  `json:"liquidity"`
}

// FeatureEngine computes derived metrics from raw tick data
type FeatureEngine struct {
	tracker *tracker.StateTracker
}

// NewFeatureEngine creates a feature engine backed by a state tracker
func NewFeatureEngine(st *tracker.StateTracker) *FeatureEngine {
	return &FeatureEngine{tracker: st}
}

// Compute calculates all features for a given pair
func (fe *FeatureEngine) Compute(ctx context.Context, pair string) (*PairFeatures, bool, error) {
	current, hasCurrent, err := fe.tracker.GetLastTick(ctx, pair)
	if err != nil {
		return nil, false, err
	}
	if !hasCurrent {
		return nil, false, nil
	}

	prev, hasPrev, err := fe.tracker.GetPrevTick(ctx, pair)
	if err != nil {
		return nil, false, err
	}

	spread := current.Ask - current.Bid
	midPrice := (current.Ask + current.Bid) / 2.0

	spreadPct := 0.0
	if midPrice > 0 {
		spreadPct = (spread / midPrice) * 100
	}

	momentum := 0.0
	momentumPct := 0.0
	if hasPrev && prev.Last > 0 {
		momentum = current.Last - prev.Last
		momentumPct = ((current.Last - prev.Last) / prev.Last) * 100
	}

	volDelta := 0.0
	if hasPrev {
		volDelta = current.Volume - prev.Volume
	}

	sma, _, _ := fe.tracker.SMA(ctx, pair)
	volatility, _, _ := fe.tracker.Volatility(ctx, pair)
	priceChangePct, _, _ := fe.tracker.PriceChangePct(ctx, pair)

	trend := fe.classifyTrend(momentumPct, priceChangePct)
	liquidity := fe.classifyLiquidity(spreadPct)

	return &PairFeatures{
		Pair:        pair,
		Spread:      spread,
		SpreadPct:   spreadPct,
		MidPrice:    midPrice,
		Momentum:    momentum,
		MomentumPct: momentumPct,
		VolumeDelta: volDelta,
		Volatility:  volatility,
		SMA:         sma,
		Trend:       trend,
		Liquidity:   liquidity,
	}, true, nil
}

// ComputeAll calculates features for all tracked pairs
func (fe *FeatureEngine) ComputeAll(ctx context.Context, pairs []string) []*PairFeatures {
	results := make([]*PairFeatures, 0, len(pairs))
	for _, pair := range pairs {
		if f, ok, _ := fe.Compute(ctx, pair); ok {
			results = append(results, f)
		}
	}
	return results
}

func (fe *FeatureEngine) classifyTrend(shortMomentum, windowChange float64) string {
	if shortMomentum > 0.05 && windowChange > 0.5 {
		return "strong_bullish"
	}
	if shortMomentum > 0.01 && windowChange > 0.1 {
		return "bullish"
	}
	if shortMomentum < -0.05 && windowChange < -0.5 {
		return "strong_bearish"
	}
	if shortMomentum < -0.01 && windowChange < -0.1 {
		return "bearish"
	}
	return "neutral"
}

func (fe *FeatureEngine) classifyLiquidity(spreadPct float64) string {
	if spreadPct < 0.01 {
		return "high"
	}
	if spreadPct < 0.05 {
		return "medium"
	}
	return "low"
}

// SpreadPctFromTick calculates spread percentage from raw bid/ask
func SpreadPctFromTick(bid, ask float64) float64 {
	mid := (bid + ask) / 2.0
	if mid == 0 {
		return 0
	}
	return ((ask - bid) / mid) * 100
}

// VolatilityFromPrices calculates standard deviation from a slice of prices
func VolatilityFromPrices(prices []float64) float64 {
	n := len(prices)
	if n < 2 {
		return 0
	}

	mean := 0.0
	for _, p := range prices {
		mean += p
	}
	mean /= float64(n)

	variance := 0.0
	for _, p := range prices {
		diff := p - mean
		variance += diff * diff
	}
	variance /= float64(n)

	return math.Sqrt(variance)
}

// SMAFromPrices calculates simple moving average from a slice of prices
func SMAFromPrices(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}
