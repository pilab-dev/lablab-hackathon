package decision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"kraken-trader/internal/features"
	"kraken-trader/internal/news"
	"kraken-trader/internal/repository"
	"kraken-trader/internal/state"
)

// TradeDecision matches the LLM JSON output schema
type TradeDecision struct {
	Pair       string  `json:"pair"`
	Action     string  `json:"action"` // "buy", "sell", "hold"
	SizePct    float64 `json:"size_pct"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// Engine is the brain of the bot, powered by Llama 3.1
type Engine struct {
	provider     string
	baseURL      string
	model        string
	stateMgr     *state.MemoryManager
	chromaClient *news.ChromaClient
	embedder     *news.Embedder
	httpClient   *http.Client
	repo         repository.Repository
	featEngine   *features.FeatureEngine
}

// NewEngine creates a new Llama-based decision engine
func NewEngine(provider, baseURL, model string, stateMgr *state.MemoryManager, chroma *news.ChromaClient, embedder *news.Embedder, repo repository.Repository, featEngine *features.FeatureEngine) *Engine {
	if provider == "" {
		provider = "ollama"
	}
	if baseURL == "" {
		if provider == "ollama" {
			baseURL = "http://localhost:11434"
		} else {
			baseURL = "http://localhost:1234"
		}
	}
	if model == "" {
		model = "llama3.1:8b"
	}
	return &Engine{
		provider:     provider,
		baseURL:      baseURL,
		model:        model,
		stateMgr:     stateMgr,
		chromaClient: chroma,
		embedder:     embedder,
		repo:         repo,
		featEngine:   featEngine,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SetFeatureEngine allows injecting the feature engine after initialization
func (e *Engine) SetFeatureEngine(featEngine *features.FeatureEngine) {
	e.featEngine = featEngine
}

// Decide runs the full inference loop for a set of pairs
func (e *Engine) Decide(ctx context.Context, pairs []string) ([]TradeDecision, error) {
	allDecisions := []TradeDecision{}

	for _, pair := range pairs {
		snap, ok := e.stateMgr.GetMarketSnapshot(pair)
		if !ok {
			continue
		}

		var feat *features.PairFeatures
		if e.featEngine != nil {
			f, hasFeat, _ := e.featEngine.Compute(ctx, pair)
			if hasFeat {
				feat = f
			}
		}

		price := snap.Last
		spreadPct := 0.0
		microPrice := 0.0
		orderBookImb := 0.0
		if feat != nil {
			spreadPct = feat.SpreadPct
			microPrice = feat.MicroPrice
			orderBookImb = feat.OrderBookImbalance
		} else if snap.Ask > 0 && snap.Bid > 0 {
			spreadPct = ((snap.Ask - snap.Bid) / ((snap.Ask + snap.Bid) / 2.0)) * 100
		}

		rsi := PRISMSignalToRSI(snap.MomentumSignal)

		macdSignal, macdHistogram := PRISMSignalToMACD(snap.MomentumSignal, snap.BreakoutSignal)

		ema20 := 0.0
		ema50 := 0.0
		var volumeRatio float64
		if feat != nil {
			ema20 = feat.SMA
			ema50 = 0
			volumeRatio = feat.VolumeSurge
		} else {
			volumeRatio = PRISMStrengthToVolumeRatio(snap.VolumeSignal)
		}

		sentimentScore := PRISMToSentimentScore(snap.MomentumSignal)
		sentimentSource := "PRISM:" + pair

		momentumScore := PRISMToMomentumScore(snap.MomentumSignal)
		lookbackBars := 20

		input := BuildLLMInput(
			pair, price, spreadPct,
			microPrice, orderBookImb,
			rsi, macdSignal, macdHistogram, ema20, ema50, volumeRatio,
			sentimentScore, sentimentSource,
			momentumScore, lookbackBars,
		)

		userPrompt, err := BuildUserPromptFromInput(input)
		if err != nil {
			log.Error().Err(err).Str("pair", pair).Msg("failed to build prompt")
			continue
		}

		var decisions []TradeDecision
		if e.provider == "lmstudio" {
			decisions, err = e.callLMStudio(ctx, pair, userPrompt)
		} else {
			decisions, err = e.callOllama(ctx, pair, userPrompt)
		}

		if err != nil {
			log.Error().Err(err).Str("pair", pair).Msg("LLM decision failed")
			continue
		}

		allDecisions = append(allDecisions, decisions...)
	}

	return allDecisions, nil
}

func (e *Engine) callOllama(ctx context.Context, pair string, userPrompt string) ([]TradeDecision, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	messages := []message{
		{Role: "system", Content: SystemPrompt},
	}
	messages = append(messages, message{Role: "user", Content: userPrompt})

	reqBody := map[string]interface{}{
		"model":    e.model,
		"messages": messages,
		"stream":   false,
		"format":   "json",
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/chat", e.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return e.processLLMResponse(ctx, pair, messages, ollamaResp.Message.Content)
}

func (e *Engine) callLMStudio(ctx context.Context, pair string, userPrompt string) ([]TradeDecision, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	messages := []message{
		{Role: "system", Content: SystemPrompt},
	}
	messages = append(messages, message{Role: "user", Content: userPrompt})

	reqBody := map[string]interface{}{
		"model":    e.model,
		"messages": messages,
		"stream":   false,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"temperature": 0.1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/chat/completions", e.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lmstudio chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lmstudio returned status %d", resp.StatusCode)
	}

	var lmsResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&lmsResp); err != nil {
		return nil, fmt.Errorf("failed to decode lmstudio response: %w", err)
	}

	if len(lmsResp.Choices) == 0 {
		return nil, fmt.Errorf("lmstudio returned no choices")
	}

	return e.processLLMResponse(ctx, pair, messages, lmsResp.Choices[0].Message.Content)
}

func (e *Engine) processLLMResponse(ctx context.Context, pair string, messages interface{}, content string) ([]TradeDecision, error) {
	var llmOut LLMOutput
	if err := json.Unmarshal([]byte(content), &llmOut); err != nil {
		preview := content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.Debug().Str("raw_output_preview", preview).Int("raw_output_length", len(content)).Msg("Raw LLM output")
		return nil, fmt.Errorf("failed to parse LLM decisions JSON: %w", err)
	}

	sizePct := 0.0
	switch llmOut.Sizing {
	case "FULL":
		sizePct = 10.0
	case "HALF":
		sizePct = 5.0
	case "QUARTER":
		sizePct = 2.5
	case "SKIP":
		sizePct = 0.0
	}

	reasoning := llmOut.Reasoning.PrimarySignal
	if len(llmOut.Reasoning.Supporting) > 0 {
		reasoning += "; " + llmOut.Reasoning.Supporting[0]
	}

	decisions := []TradeDecision{}
	if llmOut.Action != "" {
		decisions = append(decisions, TradeDecision{
			Pair:       pair,
			Action:     llmOut.Action,
			SizePct:    sizePct,
			Confidence: llmOut.Confidence,
			Reasoning:  reasoning,
		})

		if e.repo != nil {
			rawPrompt, _ := json.Marshal(messages)
			rec := repository.PromptRecord{
				Type:       "decision",
				Pair:       pair,
				RawPrompt:  string(rawPrompt),
				RawAnswer:  content,
				Answer:     content,
				Action:     llmOut.Action,
				SizePct:    sizePct,
				Confidence: llmOut.Confidence,
				Success:    true,
			}
			_ = e.repo.SavePrompt(ctx, rec)
		}
	}

	return decisions, nil
}

func (e *Engine) GetPrompts(ctx context.Context, limit int) ([]repository.PromptRecord, error) {
	if e.repo == nil {
		return nil, nil
	}
	return e.repo.GetPromptsList(ctx, limit)
}
