package decision

import (
	"encoding/json"
	"fmt"
	"time"
)

const SystemPrompt = `You are a high-frequency quantitative trading analyst. Your task is to evaluate the provided market indicators and sentiment data, then produce a structured trading decision in strict JSON format.

## Output Format

Respond only with a single valid JSON object. No markdown, no explanation, no extra text.

Sizing logic:
- confidence >= 0.75: FULL
- confidence 0.50-0.74: HALF
- confidence 0.30-0.49: QUARTER
- confidence < 0.30: SKIP

If action is HOLD, set sizing to SKIP and set entry_price, stop_loss, and take_profit to null.`

type LLMInput struct {
	Symbol       string     `json:"symbol"`
	Timestamp    string     `json:"timestamp"`
	Price        float64    `json:"price"`
	SpreadPct    float64    `json:"spread_pct"`
	MicroPrice   float64    `json:"micro_price"`
	OrderBookImb float64    `json:"order_book_imbalance"`
	Indicators   Indicators `json:"indicators"`
	Sentiment    Sentiment  `json:"sentiment"`
	Momentum     Momentum   `json:"momentum"`
}

type Indicators struct {
	RSI           float64 `json:"rsi"`
	MACDSignal    float64 `json:"macd_signal"`
	MACDHistogram float64 `json:"macd_histogram"`
	EMA20         float64 `json:"ema_20"`
	EMA50         float64 `json:"ema_50"`
	VolumeRatio   float64 `json:"volume_ratio"`
}

type Sentiment struct {
	Score  float64 `json:"score"`
	Source string  `json:"source"`
}

type Momentum struct {
	Score        float64 `json:"score"`
	LookbackBars int     `json:"lookback_bars"`
}

type LLMOutput struct {
	Action     string    `json:"action"`
	Confidence float64   `json:"confidence"`
	Sizing     string    `json:"sizing"`
	EntryPrice *float64  `json:"entry_price"`
	StopLoss   *float64  `json:"stop_loss"`
	TakeProfit *float64  `json:"take_profit"`
	Reasoning  Reasoning `json:"reasoning"`
	Risk       Risk      `json:"risk"`
}

type Reasoning struct {
	PrimarySignal string   `json:"primary_signal"`
	Supporting    []string `json:"supporting"`
	Conflicts     []string `json:"conflicts"`
	RulesApplied  []string `json:"rules_applied"`
}

type Risk struct {
	RewardRatio  float64 `json:"reward_ratio"`
	Invalidation string  `json:"invalidation"`
}

func BuildLLMInput(
	symbol string,
	price, spreadPct float64,
	microPrice, orderBookImb float64,
	rsi, macdSignal, macdHistogram, ema20, ema50, volumeRatio float64,
	sentimentScore float64,
	sentimentSource string,
	momentumScore float64,
	lookbackBars int,
) LLMInput {
	return LLMInput{
		Symbol:       symbol,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Price:        price,
		SpreadPct:    spreadPct,
		MicroPrice:   microPrice,
		OrderBookImb: orderBookImb,
		Indicators: Indicators{
			RSI:           rsi,
			MACDSignal:    macdSignal,
			MACDHistogram: macdHistogram,
			EMA20:         ema20,
			EMA50:         ema50,
			VolumeRatio:   volumeRatio,
		},
		Sentiment: Sentiment{
			Score:  sentimentScore,
			Source: sentimentSource,
		},
		Momentum: Momentum{
			Score:        momentumScore,
			LookbackBars: lookbackBars,
		},
	}
}

func BuildUserPromptFromInput(input LLMInput) (string, error) {
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM input: %w", err)
	}

	return fmt.Sprintf("## Input Data\n\n%s\n\nApply the decision rules and respond with the required JSON output.", string(inputJSON)), nil
}

func PRISMSignalToRSI(signal string) float64 {
	switch signal {
	case "strong_bullish":
		return 75.0
	case "bullish":
		return 65.0
	case "neutral":
		return 50.0
	case "bearish":
		return 35.0
	case "strong_bearish":
		return 25.0
	default:
		return 50.0
	}
}

func PRISMSignalToMACD(signal, direction string) (float64, float64) {
	var macdSignal, histogram float64
	switch direction {
	case "up":
		macdSignal = 0.5
		histogram = 0.3
	case "down":
		macdSignal = -0.5
		histogram = -0.3
	default:
		macdSignal = 0.0
		histogram = 0.0
	}

	switch signal {
	case "strong_bullish":
		macdSignal *= 2.0
		histogram *= 2.0
	case "strong_bearish":
		macdSignal *= 2.0
		histogram *= 2.0
	}

	return macdSignal, histogram
}

func PRISMStrengthToVolumeRatio(strength string) float64 {
	switch strength {
	case "strong":
		return 1.5
	case "moderate":
		return 1.0
	case "weak":
		return 0.5
	default:
		return 1.0
	}
}

func PRISMToSentimentScore(signal string) float64 {
	switch signal {
	case "strong_bullish":
		return 0.8
	case "bullish":
		return 0.5
	case "neutral":
		return 0.0
	case "bearish":
		return -0.5
	case "strong_bearish":
		return -0.8
	default:
		return 0.0
	}
}

func PRISMToMomentumScore(signal string) float64 {
	switch signal {
	case "strong_bullish":
		return 0.8
	case "bullish":
		return 0.4
	case "neutral":
		return 0.0
	case "bearish":
		return -0.4
	case "strong_bearish":
		return -0.8
	default:
		return 0.0
	}
}
