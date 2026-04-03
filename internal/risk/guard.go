package risk

import (
	"context"
	"fmt"
	"time"

	"kraken-trader/internal/tracker"
)

// RiskConfig holds risk management parameters
type RiskConfig struct {
	MaxDrawdownPct     float64
	MaxPositionSizePct float64
	MaxTradesPerMinute int
	MinConfidence      float64
	MaxSlippagePct     float64
	CooldownDuration   time.Duration
}

// DefaultRiskConfig returns conservative defaults
func DefaultRiskConfig() RiskConfig {
	return RiskConfig{
		MaxDrawdownPct:     5.0,
		MaxPositionSizePct: 10.0,
		MaxTradesPerMinute: 5,
		MinConfidence:      0.7,
		MaxSlippagePct:     0.05,
		CooldownDuration:   30 * time.Second,
	}
}

// RiskAssessment holds the result of a risk check
type RiskAssessment struct {
	Approved    bool     `json:"approved"`
	Reasons     []string `json:"reasons"`
	SlippagePct float64  `json:"slippage_pct"`
	MaxSizePct  float64  `json:"max_size_pct"`
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
	if hasCurrent && midPrice > 0 {
		spread := current.Ask - current.Bid
		slippagePct := (spread / midPrice) * 100
		if slippagePct > rg.cfg.MaxSlippagePct {
			approved = false
			reasons = append(reasons, fmt.Sprintf("estimated slippage %.4f%% exceeds max %.4f%%", slippagePct, rg.cfg.MaxSlippagePct))
		}
	}

	// 5. Position size limit
	if sizePct > rg.cfg.MaxPositionSizePct {
		maxSize = rg.cfg.MaxPositionSizePct
		reasons = append(reasons, fmt.Sprintf("size capped from %.2f%% to %.2f%%", sizePct, maxSize))
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "all risk checks passed")
	}

	return &RiskAssessment{
		Approved:    approved,
		Reasons:     reasons,
		SlippagePct: 0,
		MaxSizePct:  maxSize,
	}
}
