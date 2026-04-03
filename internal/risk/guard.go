package risk

import (
	"context"
	"fmt"
	"time"

	"kraken-trader/internal/tracker"
)

// RiskConfig holds risk management parameters
type RiskConfig struct {
	MaxDrawdownPct       float64
	MaxPositionSizePct   float64
	MaxTradesPerMinute   int
	MinConfidence        float64
	MaxSlippagePct       float64
	CooldownDuration     time.Duration
	MaxSpreadForEntryPct float64 // Spread threshold above which confidence is capped
	SpreadConfidenceCap  float64 // Max confidence when spread exceeds threshold
	MaxVolatilityPct     float64 // Volatility percentage above which trading is restricted
	MaxVolumeSurge       float64 // Volume surge ratio above which trading is restricted
}

// DefaultRiskConfig returns conservative defaults
func DefaultRiskConfig() RiskConfig {
	return RiskConfig{
		MaxDrawdownPct:       5.0,
		MaxPositionSizePct:   10.0,
		MaxTradesPerMinute:   5,
		MinConfidence:        0.7,
		MaxSlippagePct:       0.05,
		CooldownDuration:     30 * time.Second,
		MaxSpreadForEntryPct: 0.1,
		SpreadConfidenceCap:  0.3,
		MaxVolatilityPct:     2.0,
		MaxVolumeSurge:       5.0,
	}
}

// RiskAssessment holds the result of a risk check
type RiskAssessment struct {
	Approved      bool     `json:"approved"`
	Reasons       []string `json:"reasons"`
	SlippagePct   float64  `json:"slippage_pct"`
	MaxSizePct    float64  `json:"max_size_pct"`
	AdjConfidence float64  `json:"adj_confidence"`
	SpreadPct     float64  `json:"spread_pct"`
	VolatilityPct float64  `json:"volatility_pct"`
}

// RiskGuard validates trades against risk parameters
type RiskGuard struct {
	cfg     RiskConfig
	tracker *tracker.StateTracker
}

// NewRiskGuard creates a risk guard with the given configuration
func NewRiskGuard(cfg RiskConfig, st *tracker.StateTracker) *RiskGuard {
	return &RiskGuard{
		cfg:     cfg,
		tracker: st,
	}
}

// Check evaluates whether a trade is safe to execute
func (rg *RiskGuard) Check(ctx context.Context, pair string, action string, sizePct, confidence, midPrice float64) *RiskAssessment {
	reasons := []string{}
	approved := true
	maxSize := rg.cfg.MaxPositionSizePct
	adjConfidence := confidence

	// 1. Confidence check
	if confidence < rg.cfg.MinConfidence {
		approved = false
		reasons = append(reasons, fmt.Sprintf("confidence %.2f below threshold %.2f", confidence, rg.cfg.MinConfidence))
	}

	// 2. Drawdown check
	totalPnL, err := rg.tracker.TotalPnL(ctx)
	if err == nil && totalPnL < 0 {
		drawdownPct := -totalPnL
		if drawdownPct > rg.cfg.MaxDrawdownPct {
			approved = false
			reasons = append(reasons, fmt.Sprintf("drawdown %.2f%% exceeds max %.2f%%", drawdownPct, rg.cfg.MaxDrawdownPct))
		}
	}

	// 3. Rate limiting — cooldown between trades
	count, err := rg.tracker.TradeCount(ctx)
	if err == nil && count > 0 {
		lastTradeTime, hasLastTrade, _ := rg.tracker.LastTradeTime(ctx)
		if hasLastTrade && time.Since(lastTradeTime) < rg.cfg.CooldownDuration {
			approved = false
			reasons = append(reasons, fmt.Sprintf("cooldown active, last trade %v ago", time.Since(lastTradeTime).Round(time.Second)))
		}
	}

	// 4. Slippage estimation from spread
	current, hasCurrent, _ := rg.tracker.GetLastTick(ctx, pair)
	spreadPct := 0.0
	if hasCurrent && midPrice > 0 {
		spread := current.Ask - current.Bid
		spreadPct = (spread / midPrice) * 100
		if spreadPct > rg.cfg.MaxSlippagePct {
			approved = false
			reasons = append(reasons, fmt.Sprintf("estimated slippage %.4f%% exceeds max %.4f%%", spreadPct, rg.cfg.MaxSlippagePct))
		}
	}

	// 5. Spread-based confidence capping
	if spreadPct > rg.cfg.MaxSpreadForEntryPct {
		if adjConfidence > rg.cfg.SpreadConfidenceCap {
			adjConfidence = rg.cfg.SpreadConfidenceCap
			reasons = append(reasons, fmt.Sprintf("confidence capped at %.2f due to wide spread (%.4f%%)", rg.cfg.SpreadConfidenceCap, spreadPct))
		}
	}

	// 6. Volatility-based trading restriction
	volatilityPct := 0.0
	if hasCurrent && midPrice > 0 {
		ticks, err := rg.tracker.Window(ctx, pair)
		if err == nil && len(ticks) >= 2 {
			prices := make([]float64, len(ticks))
			for i, t := range ticks {
				prices[i] = t.Last
			}
			vol := calcVolatility(prices)
			volatilityPct = (vol / midPrice) * 100
			if volatilityPct > rg.cfg.MaxVolatilityPct {
				approved = false
				reasons = append(reasons, fmt.Sprintf("volatility %.2f%% exceeds max %.2f%%", volatilityPct, rg.cfg.MaxVolatilityPct))
			}
		}
	}

	// 7. Volume surge restriction
	if hasCurrent {
		volSurge, _, _ := rg.tracker.VolumeSurge(ctx, pair)
		if volSurge > rg.cfg.MaxVolumeSurge {
			approved = false
			reasons = append(reasons, fmt.Sprintf("volume surge %.2fx exceeds max %.2fx", volSurge, rg.cfg.MaxVolumeSurge))
		}
	}

	// 8. Position size limit
	if sizePct > rg.cfg.MaxPositionSizePct {
		maxSize = rg.cfg.MaxPositionSizePct
		reasons = append(reasons, fmt.Sprintf("size capped from %.2f%% to %.2f%%", sizePct, maxSize))
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "all risk checks passed")
	}

	return &RiskAssessment{
		Approved:      approved,
		Reasons:       reasons,
		SlippagePct:   spreadPct,
		MaxSizePct:    maxSize,
		AdjConfidence: adjConfidence,
		SpreadPct:     spreadPct,
		VolatilityPct: volatilityPct,
	}
}

func calcVolatility(prices []float64) float64 {
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

	if variance < 0 {
		return 0
	}

	z := variance
	for i := 0; i < 20; i++ {
		if z == 0 {
			break
		}
		z -= (z*z - variance) / (2 * z)
	}
	return z
}
