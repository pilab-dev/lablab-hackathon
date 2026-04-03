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
	Symbol        string  `json:"symbol"`
	OverallSignal string  `json:"overall_signal"`
	Direction     string  `json:"direction"`
	Strength      string  `json:"strength"`
	BullishScore  int     `json:"bullish_score"`
	BearishScore  int     `json:"bearish_score"`
	CurrentPrice  float64 `json:"current_price"`
	Timestamp     string  `json:"timestamp"`
}

type prismSignalResponse struct {
	Object     string          `json:"object"`
	Data       []SignalSummary `json:"data"`
	Pagination PaginationInfo  `json:"pagination"`
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

func (c *PrismClient) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	var lastErr error
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Check if it's a transient error that might be worth retrying
			errMsg := err.Error()
			if strings.Contains(errMsg, "http2: server sent GOAWAY") ||
				strings.Contains(errMsg, "connection reset by peer") ||
				strings.Contains(errMsg, "timeout") {
				// Retry on transient network issues
			} else {
				return nil, err // unexpected error
			}
		} else {
			if resp.StatusCode == http.StatusOK {
				return resp, nil
			}

			lastErr = fmt.Errorf("prism api returned status: %d", resp.StatusCode)

			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				// Handle Retry-After if present for 429
				if resp.StatusCode == http.StatusTooManyRequests {
					if ra := resp.Header.Get("Retry-After"); ra != "" {
						if d, err := time.ParseDuration(ra + "s"); err == nil {
							backoff = d
						}
					}
				}
				resp.Body.Close()
			} else {
				// Other status codes (400, 401, 403, 404) should not be retried
				return resp, nil
			}
		}

		if i == maxRetries {
			break
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			if backoff < 1*time.Minute { // Cap backoff
				backoff *= 2
			}
		}
	}

	return nil, lastErr
}

// FetchSignalSummary gets technical signals for specific pairs
func (c *PrismClient) FetchSignalSummary(ctx context.Context, symbols []string) ([]SignalSummary, error) {
	url := fmt.Sprintf("%s/signals/summary?symbols=%s", c.baseURL, strings.Join(symbols, ","))

	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prism api returned status: %d", resp.StatusCode)
	}

	// PRISM returns {"object":"list","data":[...],"pagination":{...}}
	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prism response body: %w", err)
	}

	var signalResp prismSignalResponse
	if err := json.Unmarshal(bb, &signalResp); err != nil {
		return nil, fmt.Errorf("failed to decode prism signals: %w, body: %s", err, string(bb))
	}

	return signalResp.Data, nil
}

// FetchCryptoNews gets the latest crypto news from PRISM with pagination support
func (c *PrismClient) FetchCryptoNews(ctx context.Context, perPage int, totalLimit int) ([]NewsArticle, error) {
	const maxPages = 100
	var allArticles []NewsArticle
	cursor := ""
	pageCount := 0

	for {
		pageCount++
		if pageCount >= maxPages {
			break
		}

		url := fmt.Sprintf("%s/news/crypto?limit=%d", c.baseURL, perPage)
		if cursor != "" {
			url += fmt.Sprintf("&cursor=%s", cursor)
		}

		resp, err := c.doRequest(ctx, "GET", url)
		if err != nil {
			return nil, err
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

		if totalLimit > 0 && len(allArticles) >= totalLimit {
			allArticles = allArticles[:totalLimit]
			break
		}

		if !newsResp.Pagination.HasMore || newsResp.Pagination.NextCursor == nil {
			break
		}

		cursor = *newsResp.Pagination.NextCursor
	}

	return allArticles, nil
}
