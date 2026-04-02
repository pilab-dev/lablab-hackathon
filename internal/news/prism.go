package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// PaginationInfo represents pagination metadata from PRISM API
type PaginationInfo struct {
	Total      int     `json:"total"`
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
	PrevCursor *string `json:"prev_cursor"`
}

// NewsResponse represents the paginated PRISM news API response
type NewsResponse struct {
	Object     string         `json:"object"`
	Data       []NewsItem     `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	RequestID  string         `json:"request_id"`
}

// NewsItem represents a single news item in the PRISM API response
type NewsItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Source      string `json:"source"`
	PublishedAt string `json:"published_at"`
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
		bb, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("failed to decode prism signals: %w, body: %+v", err, string(bb))
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

// FetchCryptoNews gets the latest crypto news from PRISM with pagination support
func (c *PrismClient) FetchCryptoNews(ctx context.Context, limit int) ([]NewsArticle, error) {
	var allArticles []NewsArticle
	cursor := ""

	for {
		url := fmt.Sprintf("%s/news/crypto?limit=%d", c.baseURL, limit)
		if cursor != "" {
			url += fmt.Sprintf("&cursor=%s", cursor)
		}

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

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("prism api returned status: %d", resp.StatusCode)
		}

		var newsResp NewsResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode prism news: %w", err)
		}
		resp.Body.Close()

		for _, item := range newsResp.Data {
			ts, _ := time.Parse(time.RFC3339, item.PublishedAt)
			allArticles = append(allArticles, NewsArticle{
				ID:        item.ID,
				Title:     item.Title,
				Summary:   item.Description,
				URL:       item.URL,
				Source:    item.Source,
				Timestamp: ts,
			})
		}

		if !newsResp.Pagination.HasMore || newsResp.Pagination.NextCursor == nil {
			break
		}

		cursor = *newsResp.Pagination.NextCursor
	}

	return allArticles, nil
}
