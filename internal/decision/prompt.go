package decision

import (
	"fmt"
	"strings"
)

const SystemPrompt = `
You are an autonomous, high-conviction crypto trading agent competing in a PnL-ranked hackathon.
Your goal is to maximize profit while strictly adhering to risk management rules.

RULES:
1. You may ONLY output valid JSON. No conversational filler.
2. Actions must be one of: "buy", "sell", "hold".
3. "size_pct" must be between 0.0 and 10.0 (max 10% of portfolio per trade).
4. "confidence" must be between 0.0 and 1.0.
5. If you are not highly confident (> 0.7), you MUST choose "hold".
6. Base decisions ONLY on the provided Market Data, PRISM Signals, and News Sentiment.
7. Keep "reasoning" concise and factual (under 2 sentences).

OUTPUT SCHEMA:
{
  "decisions": [
    {
      "pair": "BTCUSD",
      "action": "buy",
      "size_pct": 5.0,
      "confidence": 0.85,
      "reasoning": "Reason for the trade..."
    }
  ]
}
`

// FewShotExamples provides context to the LLM on how to behave in different scenarios.
var FewShotExamples = []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}{
	{
		Role:    "user",
		Content: "Market: BTCUSD $67k (+2% 1h). PRISM: Strong Momentum. News: 'Institutional adoption spikes'.",
	},
	{
		Role: "assistant",
		Content: `{"decisions": [{"pair": "BTCUSD", "action": "buy", "size_pct": 10.0, "confidence": 0.95, "reasoning": "Strong upward momentum and bullish institutional news provide a high-conviction entry."}]}`,
	},
	{
		Role:    "user",
		Content: "Market: ETHUSD $3400 (-0.5% 1h). PRISM: Neutral. News: Mixed sentiment.",
	},
	{
		Role: "assistant",
		Content: `{"decisions": [{"pair": "ETHUSD", "action": "hold", "size_pct": 0.0, "confidence": 0.80, "reasoning": "Market is consolidating with neutral signals; waiting for clear direction."}]}`,
	},
}

// BuildUserPrompt constructs the dynamic part of the prompt with real-time context.
func BuildUserPrompt(marketState string, signals []string, news []string, portfolio string) string {
	return fmt.Sprintf(`
CURRENT STATE:
%s

PRISM SIGNALS:
%s

RECENT RELEVANT NEWS:
%s

PORTFOLIO & RISK:
%s

Provide your trading decisions in the required JSON format.
`, marketState, strings.Join(signals, "\n"), strings.Join(news, "\n"), portfolio)
}
