package strategy

import (
	"context"
	"fmt"
	"time"

	"kraken-trader/internal/features"
	"kraken-trader/internal/state"
)

// SignalStrength represents the combined signal strength
type SignalStrength string

const (
	StrongBuy  SignalStrength = "strong_buy"
	Buy        SignalStrength = "buy"
	WeakBuy    SignalStrength = "weak_buy"
	Hold       SignalStrength = "hold"
	WeakSell   SignalStrength = "weak_sell"
	Sell       SignalStrength = "sell"
	StrongSell SignalStrength = "strong_sell"
)

// StrategyResult holds the output of the strategy engine
type StrategyResult struct {
	Pair       string         `json:"pair"`
	Signal     SignalStrength `json:"signal"`
	Score      float64        `json:"score"`
	Confidence float64        `json:"confidence"`
	Reasons    []string       `json:"reasons"`
	Timestamp  time.Time      `json:"timestamp"`
}

// StrategyEngine combines market features with PRISM sentiment
type StrategyEngine struct {
	features *features.FeatureEngine
	stateMgr *state.MemoryManager
}

// NewStrategyEngine creates a new strategy engine
func NewStrategyEngine(fe *features.FeatureEngine, sm *state.MemoryManager) *StrategyEngine {
	return &StrategyEngine{
		features: fe,
		stateMgr: sm,
	}
}

// Evaluate runs the strategy logic for a single pair
func (se *StrategyEngine) Evaluate(ctx context.Context, pair string) (*StrategyResult, bool, error) {
	feat, ok, err := se.features.Compute(ctx, pair)
	if err != nil || !ok {
		return nil, false, err
	}

	score := 0.0
	reasons := []string{}
	confidence := 0.0

	// 1. Spread check — skip if liquidity too low
	if feat.Liquidity == "low" {
		return &StrategyResult{
			Pair:       pair,
			Signal:     Hold,
			Score:      0,
			Confidence: 1.0,
			Reasons:    []string{"spread too wide, low liquidity"},
			Timestamp:  time.Now(),
		}, true, nil
	}

	// 2. Momentum scoring (-1 to +1)
	momentumScore := 0.0
	if feat.Trend == "strong_bullish" {
		momentumScore = 1.0
		reasons = append(reasons, "strong bullish momentum")
	} else if feat.Trend == "bullish" {
		momentumScore = 0.6
		reasons = append(reasons, "bullish momentum")
	} else if feat.Trend == "strong_bearish" {
		momentumScore = -1.0
		reasons = append(reasons, "strong bearish momentum")
	} else if feat.Trend == "bearish" {
		momentumScore = -0.6
		reasons = append(reasons, "bearish momentum")
	} else {
		reasons = append(reasons, "neutral momentum")
	}

	// 3. Volume confirmation
	if feat.VolumeDelta > 0 && momentumScore > 0 {
		momentumScore *= 1.2
		reasons = append(reasons, "volume confirms upward move")
	} else if feat.VolumeDelta > 0 && momentumScore < 0 {
		momentumScore *= 1.1
		reasons = append(reasons, "volume confirms downward move")
	} else if feat.VolumeDelta < 0 {
		momentumScore *= 0.8
		reasons = append(reasons, "low volume, weak signal")
	}

	// 4. Mean reversion check
	if feat.SMA > 0 {
		deviation := (feat.MidPrice - feat.SMA) / feat.SMA
		if deviation > 0.02 && momentumScore < 0 {
			momentumScore -= 0.2
			reasons = append(reasons, "price above SMA, overextended")
		} else if deviation < -0.02 && momentumScore > 0 {
			momentumScore += 0.2
			reasons = append(reasons, "price below SMA, potential bounce")
		}
	}

	// 5. PRISM sentiment overlay
	snap, hasSnap := se.stateMgr.GetMarketSnapshot(pair)
	if hasSnap {
		sentimentScore := se.scorePRISM(snap)
		momentumScore += sentimentScore * 0.3
		if sentimentScore > 0.3 {
			reasons = append(reasons, fmt.Sprintf("PRISM sentiment bullish (%s)", snap.MomentumSignal))
		} else if sentimentScore < -0.3 {
			reasons = append(reasons, fmt.Sprintf("PRISM sentiment bearish (%s)", snap.MomentumSignal))
		}
	}

	// Clamp score to [-1, 1]
	if momentumScore > 1.0 {
		momentumScore = 1.0
	}
	if momentumScore < -1.0 {
		momentumScore = -1.0
	}

	score = momentumScore
	confidence = se.calcConfidence(feat, hasSnap)

	return &StrategyResult{
		Pair:       pair,
		Signal:     se.scoreToSignal(score),
		Score:      score,
		Confidence: confidence,
		Reasons:    reasons,
		Timestamp:  time.Now(),
	}, true, nil
}

func (se *StrategyEngine) scorePRISM(snap state.PairState) float64 {
	score := 0.0

	switch snap.MomentumSignal {
	case "bullish", "strong_bullish":
		score += 0.4
	case "bearish", "strong_bearish":
		score -= 0.4
	}

	switch snap.BreakoutSignal {
	case "up":
		score += 0.3
	case "down":
		score -= 0.3
	}

	switch snap.VolumeSignal {
	case "strong":
		score += 0.3
	case "weak":
		score -= 0.2
	}

	return score
}

func (se *StrategyEngine) calcConfidence(feat *features.PairFeatures, hasPRISM bool) float64 {
	conf := 0.5

	if feat.Liquidity == "high" {
		conf += 0.15
	} else if feat.Liquidity == "medium" {
		conf += 0.05
	}

	if feat.Volatility > 0 {
		relVol := feat.Volatility / feat.MidPrice
		if relVol < 0.01 {
			conf += 0.1
		} else if relVol > 0.05 {
			conf -= 0.15
		}
	}

	if hasPRISM {
		conf += 0.1
	}

	if conf > 1.0 {
		conf = 1.0
	}
	if conf < 0.0 {
		conf = 0.0
	}

	return conf
}

func (se *StrategyEngine) scoreToSignal(score float64) SignalStrength {
	switch {
	case score >= 0.7:
		return StrongBuy
	case score >= 0.3:
		return Buy
	case score >= 0.1:
		return WeakBuy
	case score <= -0.7:
		return StrongSell
	case score <= -0.3:
		return Sell
	case score <= -0.1:
		return WeakSell
	default:
		return Hold
	}
}

// EvaluateAll runs strategy for multiple pairs
func (se *StrategyEngine) EvaluateAll(ctx context.Context, pairs []string) ([]*StrategyResult, error) {
	results := make([]*StrategyResult, 0, len(pairs))
	for _, pair := range pairs {
		r, ok, err := se.Evaluate(ctx, pair)
		if err != nil {
			return results, err
		}
		if ok {
			results = append(results, r)
		}
	}
	return results, nil
}
