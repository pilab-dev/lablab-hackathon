package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PrismClient interacts with the Strykr PRISM API
type PrismClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// SignalSummary maps to the PRISM GET /signals/summary response
type SignalSummary struct {
	Symbol   string `json:"symbol"`
	Momentum string `json:"momentum"`
	Breakout string `json:"breakout"`
	Volume   string `json:"volume"`
}

// NewsArticle represents a standard news item
type NewsArticle struct {
	ID        string
	Title     string
	Summary   string
	URL       string
	Source    string
	Timestamp time.Time
}

// NewPrismClient creates a new PRISM API client
func NewPrismClient(apiKey string) *PrismClient {
	return &PrismClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		baseURL:    "https://api.prismapi.ai",
	}
}

// FetchSignalSummary gets technical signals for specific pairs
func (c *PrismClient) FetchSignalSummary(ctx context.Context, symbols []string) ([]SignalSummary, error) {
	url := fmt.Sprintf("%s/signals/summary?symbols=%s", c.baseURL, strings.Join(symbols, ","))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prism api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prism api returned status: %d", resp.StatusCode)
	}

	// PRISM returns a map of symbol -> signals, we need to unmarshal the dynamic keys
	var rawResp map[string]struct {
		Momentum string `json:"momentum"`
		Breakout string `json:"breakout"`
		Volume   string `json:"volume"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode prism signals: %w", err)
	}

	var summaries []SignalSummary
	for sym, data := range rawResp {
		summaries = append(summaries, SignalSummary{
			Symbol:   sym,
			Momentum: data.Momentum,
			Breakout: data.Breakout,
			Volume:   data.Volume,
		})
	}

	return summaries, nil
}

// FetchCryptoNews gets the latest crypto news from PRISM
func (c *PrismClient) FetchCryptoNews(ctx context.Context, limit int) ([]NewsArticle, error) {
	url := fmt.Sprintf("%s/news/crypto?limit=%d", c.baseURL, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prism api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prism api returned status: %d", resp.StatusCode)
	}

	// Based on typical PRISM responses, assuming an array of articles
	var rawArticles []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Source      string `json:"source"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawArticles); err != nil {
		return nil, fmt.Errorf("failed to decode prism news: %w", err)
	}

	var articles []NewsArticle
	for _, raw := range rawArticles {
		ts, _ := time.Parse(time.RFC3339, raw.PublishedAt) // Will fallback to zero time if error
		articles = append(articles, NewsArticle{
			ID:        raw.ID,
			Title:     raw.Title,
			Summary:   raw.Description,
			URL:       raw.URL,
			Source:    raw.Source,
			Timestamp: ts,
		})
	}

	return articles, nil
}
