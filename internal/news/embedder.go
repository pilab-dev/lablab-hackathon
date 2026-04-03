package news

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Embedder uses Ollama to generate text embeddings
type Embedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewEmbedder creates a new Ollama embedder
func NewEmbedder(baseURL, model string) *Embedder {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &Embedder{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// GenerateEmbedding calls Ollama to convert text into a vector
func (e *Embedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbeddingRequest{
		Model:  e.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", e.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return result.Embedding, nil
}

// BatchEmbed takes a slice of articles and embeds their content
func (e *Embedder) BatchEmbed(ctx context.Context, articles []NewsArticle) ([][]float32, error) {
	var embeddings [][]float32

	// We do this sequentially to avoid overloading the local M4 Mac with too many concurrent model calls.
	for _, article := range articles {
		// Embed the title and summary together for better semantic meaning
		text := fmt.Sprintf("%s. %s", article.Title, article.Summary)
		emb, err := e.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed article %s: %w", article.ID, err)
		}
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}
