package decision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"kraken-trader/internal/news"
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

type llmResponse struct {
	Decisions []TradeDecision `json:"decisions"`
}

// Engine is the brain of the bot, powered by Llama 3.1
type Engine struct {
	ollamaURL    string
	model        string
	stateMgr     *state.MemoryManager
	chromaClient *news.ChromaClient
	embedder     *news.Embedder
	httpClient   *http.Client
}

// NewEngine creates a new Llama-based decision engine
func NewEngine(ollamaURL, model string, stateMgr *state.MemoryManager, chroma *news.ChromaClient, embedder *news.Embedder) *Engine {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1:8b"
	}
	return &Engine{
		ollamaURL:    ollamaURL,
		model:        model,
		stateMgr:     stateMgr,
		chromaClient: chroma,
		embedder:     embedder,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // LLM inference can take time on local hardware
		},
	}
}

// Decide runs the full inference loop for a set of pairs
func (e *Engine) Decide(ctx context.Context, pairs []string) ([]TradeDecision, error) {
	// 1. Gather Context
	marketCtx := ""
	signalsCtx := []string{}
	for _, p := range pairs {
		snap, ok := e.stateMgr.GetMarketSnapshot(p)
		if ok {
			marketCtx += fmt.Sprintf("- %s: Price $%.2f, Bid $%.2f, Ask $%.2f\n", p, snap.Last, snap.Bid, snap.Ask)
			signalsCtx = append(signalsCtx, fmt.Sprintf("- %s: Momentum: %s, Breakout: %s, Vol: %s",
				p, snap.MomentumSignal, snap.BreakoutSignal, snap.VolumeSignal))
		}
	}

	// 2. Semantic News Search (ChromaDB)
	// We search for news matching the current market "vibe" (e.g. "bitcoin price action")
	newsCtx := []string{}
	if e.chromaClient != nil && e.embedder != nil {
		query := "bitcoin ethereum market price action sentiment"
		emb, err := e.embedder.GenerateEmbedding(ctx, query)
		if err == nil {
			matches, err := e.chromaClient.QuerySimilar(ctx, emb, 5)
			if err == nil {
				for _, m := range matches {
					newsCtx = append(newsCtx, fmt.Sprintf("- [%s] %s: %s", m.Source, m.Title, m.Summary))
				}
			}
		}
	}

	// 3. Build Prompt
	userPrompt := BuildUserPrompt(marketCtx, signalsCtx, newsCtx, "Portfolio: $10,000 USD, Positions: 0")

	// 4. Call Ollama
	return e.callOllama(ctx, userPrompt)
}

func (e *Engine) callOllama(ctx context.Context, userPrompt string) ([]TradeDecision, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	messages := []message{
		{Role: "system", Content: SystemPrompt},
	}
	// Add Few-Shot Examples
	for _, ex := range FewShotExamples {
		messages = append(messages, message{Role: ex.Role, Content: ex.Content})
	}
	// Add Dynamic User Prompt
	messages = append(messages, message{Role: "user", Content: userPrompt})

	reqBody := map[string]interface{}{
		"model":    e.model,
		"messages": messages,
		"stream":   false,
		"format":   "json", // Force JSON mode
		"options": map[string]interface{}{
			"temperature": 0.1, // Keep it deterministic for trading
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/chat", e.ollamaURL)
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

	// 5. Parse JSON Decisions
	var result llmResponse
	if err := json.Unmarshal([]byte(ollamaResp.Message.Content), &result); err != nil {
		log.Debug().Str("raw_output", ollamaResp.Message.Content).Msg("Raw LLM output")
		return nil, fmt.Errorf("failed to parse LLM decisions JSON: %w", err)
	}

	return result.Decisions, nil
}
