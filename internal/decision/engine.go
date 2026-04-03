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
	repo         repository.Repository
}

// NewEngine creates a new Llama-based decision engine
func NewEngine(ollamaURL, model string, stateMgr *state.MemoryManager, chroma *news.ChromaClient, embedder *news.Embedder, repo repository.Repository) *Engine {
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
		repo:         repo,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
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
	for _, ex := range FewShotExamples {
		messages = append(messages, message{Role: ex.Role, Content: ex.Content})
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

	var result llmResponse
	if err := json.Unmarshal([]byte(ollamaResp.Message.Content), &result); err != nil {
		preview := ollamaResp.Message.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.Debug().Str("raw_output_preview", preview).Int("raw_output_length", len(ollamaResp.Message.Content)).Msg("Raw LLM output")
		return nil, fmt.Errorf("failed to parse LLM decisions JSON: %w", err)
	}

	if e.repo != nil {
		rawPrompt, _ := json.Marshal(messages)
		for _, d := range result.Decisions {
			rec := repository.PromptRecord{
				Type:       "decision",
				Pair:       d.Pair,
				RawPrompt:  string(rawPrompt),
				RawAnswer:  ollamaResp.Message.Content,
				Answer:     ollamaResp.Message.Content,
				Action:     d.Action,
				SizePct:    d.SizePct,
				Confidence: d.Confidence,
				Success:    true,
			}
			_ = e.repo.SavePrompt(ctx, rec)
		}
	}

	return result.Decisions, nil
}

func (e *Engine) GetPrompts(ctx context.Context, limit int) ([]repository.PromptRecord, error) {
	if e.repo == nil {
		return nil, nil
	}
	return e.repo.GetPromptsList(ctx, limit)
}
