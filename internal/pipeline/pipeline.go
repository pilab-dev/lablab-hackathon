package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"kraken-trader/internal/features"
	"kraken-trader/internal/messaging"
	"kraken-trader/internal/risk"
	"kraken-trader/internal/strategy"
	"kraken-trader/internal/tracker"

	"github.com/rs/zerolog/log"
)

// PipelineResult holds the final output after all stages
type PipelineResult struct {
	Pair       string                   `json:"pair"`
	Action     string                   `json:"action"`
	SizePct    float64                  `json:"size_pct"`
	Confidence float64                  `json:"confidence"`
	Strategy   *strategy.StrategyResult `json:"strategy"`
	Risk       *risk.RiskAssessment     `json:"risk"`
	Features   *features.PairFeatures   `json:"features"`
}

// Pipeline executes the full DATAFLOW: NATS -> Features -> Strategy -> Risk -> Execute
type Pipeline struct {
	features *features.FeatureEngine
	strategy *strategy.StrategyEngine
	risk     *risk.RiskGuard
	tracker  *tracker.StateTracker
	nats     *messaging.NATSClient
}

// NewPipeline creates a trading pipeline with all stages
func NewPipeline(
	fe *features.FeatureEngine,
	se *strategy.StrategyEngine,
	rg *risk.RiskGuard,
	st *tracker.StateTracker,
	nats *messaging.NATSClient,
) *Pipeline {
	return &Pipeline{
		features: fe,
		strategy: se,
		risk:     rg,
		tracker:  st,
		nats:     nats,
	}
}

// ProcessTick ingests a raw tick, updates state, and runs the pipeline
func (p *Pipeline) ProcessTick(ctx context.Context, pair string, bid, bidVol, ask, askVol, last, volume float64) {
	// Stage 0: Ingest — store tick in Redis
	if p.tracker != nil {
		if err := p.tracker.PushTick(ctx, pair, bid, bidVol, ask, askVol, last, volume); err != nil {
			log.Error().Err(err).Str("pair", pair).Msg("Pipeline: failed to push tick to Redis")
			return
		}
	} else {
		log.Debug().Str("pair", pair).Msg("Pipeline: skipping tick ingest — no state tracker")
	}

	// Stage 1: Feature Engineering
	feat, ok, err := p.features.Compute(ctx, pair)
	if err != nil || !ok {
		return
	}

	// Stage 2: Strategy Evaluation
	result, ok, err := p.strategy.Evaluate(ctx, pair)
	if err != nil || !ok {
		return
	}

	// Stage 3: Risk Guard
	assessment := p.risk.Check(ctx, pair, string(result.Signal), 10.0, result.Confidence, feat.MidPrice)

	// Stage 4: Publish result
	pipelineResult := PipelineResult{
		Pair:       pair,
		Action:     string(result.Signal),
		SizePct:    assessment.MaxSizePct,
		Confidence: result.Confidence,
		Strategy:   result,
		Risk:       assessment,
		Features:   feat,
	}

	if assessment.Approved {
		pipelineResult.Action = "EXECUTE"
		log.Info().
			Str("pair", pair).
			Str("signal", string(result.Signal)).
			Float64("confidence", result.Confidence).
			Float64("max_size", assessment.MaxSizePct).
			Msg("Pipeline: trade APPROVED")
	} else {
		pipelineResult.Action = "REJECTED"
		log.Debug().
			Str("pair", pair).
			Strs("reasons", assessment.Reasons).
			Msg("Pipeline: trade REJECTED")
	}

	// Publish to NATS for downstream consumers (API, logging, execution)
	if p.nats != nil {
		data, err := json.Marshal(pipelineResult)
		if err != nil {
			log.Error().Err(err).Msg("Pipeline: failed to marshal result")
			return
		}
		subject := fmt.Sprintf("pipeline.result.%s", pair)
		if err := p.nats.Publish(subject, data); err != nil {
			log.Error().Err(err).Msg("Pipeline: failed to publish result")
		}
	}
}

// ProcessAll runs the pipeline for all given pairs without ingesting new ticks
func (p *Pipeline) ProcessAll(ctx context.Context, pairs []string) []PipelineResult {
	results := make([]PipelineResult, 0, len(pairs))

	for _, pair := range pairs {
		feat, ok, err := p.features.Compute(ctx, pair)
		if err != nil || !ok {
			continue
		}

		result, ok, err := p.strategy.Evaluate(ctx, pair)
		if err != nil || !ok {
			continue
		}

		assessment := p.risk.Check(ctx, pair, string(result.Signal), 10.0, result.Confidence, feat.MidPrice)

		finalAction := "REJECTED"
		if assessment.Approved {
			finalAction = "EXECUTE"
		}

		results = append(results, PipelineResult{
			Pair:       pair,
			Action:     finalAction,
			SizePct:    assessment.MaxSizePct,
			Confidence: result.Confidence,
			Strategy:   result,
			Risk:       assessment,
			Features:   feat,
		})
	}

	return results
}

// GetPipelineStatus returns the current state of all pipeline components
func (p *Pipeline) GetPipelineStatus(ctx context.Context, pairs []string) map[string]map[string]interface{} {
	status := make(map[string]map[string]interface{})

	for _, pair := range pairs {
		entry := make(map[string]interface{})

		if p.tracker != nil {
			tickLen, err := p.tracker.WindowLen(ctx, pair)
			if err == nil {
				entry["ticks_in_window"] = tickLen
			}

			sma, _, _ := p.tracker.SMA(ctx, pair)
			entry["sma"] = sma

			vol, _, _ := p.tracker.Volatility(ctx, pair)
			entry["volatility"] = vol

			changePct, _, _ := p.tracker.PriceChangePct(ctx, pair)
			entry["price_change_pct"] = changePct
		}

		if feat, ok, _ := p.features.Compute(ctx, pair); ok {
			entry["features"] = feat
		}

		if result, ok, _ := p.strategy.Evaluate(ctx, pair); ok {
			entry["strategy"] = result
		}

		status[pair] = entry
	}

	return status
}

// RecordTrade logs a completed trade for drawdown tracking
func (p *Pipeline) RecordTrade(ctx context.Context, pair, action string, price, sizePct, pnl float64) {
	if p.tracker == nil {
		log.Warn().Str("pair", pair).Msg("Pipeline: cannot record trade — no state tracker")
		return
	}
	if err := p.tracker.RecordTrade(ctx, tracker.TradeRecord{
		Pair:    pair,
		Action:  action,
		Price:   price,
		SizePct: sizePct,
		PnL:     pnl,
	}); err != nil {
		log.Error().Err(err).Str("pair", pair).Msg("Pipeline: failed to record trade")
	}
}

// GetTotalPnL returns the cumulative PnL of all recorded trades
func (p *Pipeline) GetTotalPnL(ctx context.Context) float64 {
	if p.tracker == nil {
		return 0
	}
	total, _ := p.tracker.TotalPnL(ctx)
	return total
}

// GetTradeCount returns the number of recorded trades
func (p *Pipeline) GetTradeCount(ctx context.Context) int {
	if p.tracker == nil {
		return 0
	}
	count, _ := p.tracker.TradeCount(ctx)
	return int(count)
}
