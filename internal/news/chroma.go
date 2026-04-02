package news

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChromaClient interacts with a ChromaDB instance via REST
type ChromaClient struct {
	baseURL      string
	collectionID string
	httpClient   *http.Client
}

// NewChromaClient creates a ChromaDB client
func NewChromaClient(baseURL string) *ChromaClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &ChromaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Initialize setup the collection, getting its ID
func (c *ChromaClient) Initialize(ctx context.Context, collectionName string) error {
	// First try to get the collection to see if it exists
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, collectionName)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := c.httpClient.Do(req)
	
	if err == nil && resp.StatusCode == http.StatusOK {
		var result struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			c.collectionID = result.ID
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()
	}

	// If it doesn't exist, create it
	createUrl := fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	payload := map[string]string{"name": collectionName}
	jsonData, _ := json.Marshal(payload)
	
	req, err = http.NewRequestWithContext(ctx, "POST", createUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("chroma returned status %d creating collection", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	c.collectionID = result.ID
	return nil
}

// AddEmbeddings inserts articles into the vector database
func (c *ChromaClient) AddEmbeddings(ctx context.Context, articles []NewsArticle, embeddings [][]float32) error {
	if len(articles) == 0 {
		return nil
	}
	if c.collectionID == "" {
		return fmt.Errorf("collection not initialized")
	}

	var ids []string
	var metadatas []map[string]interface{}
	var documents []string

	for _, a := range articles {
		ids = append(ids, a.ID)
		documents = append(documents, a.Summary)
		metadatas = append(metadatas, map[string]interface{}{
			"title":     a.Title,
			"source":    a.Source,
			"url":       a.URL,
			"timestamp": a.Timestamp.Unix(), // ChromaDB metadatas prefer basic types
		})
	}

	payload := map[string]interface{}{
		"ids":        ids,
		"embeddings": embeddings,
		"metadatas":  metadatas,
		"documents":  documents,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add embeddings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("chroma returned status %d on add", resp.StatusCode)
	}

	return nil
}

// QueryResult represents a matching news article from the vector DB
type QueryResult struct {
	ID       string
	Distance float32
	Title    string
	Summary  string
	Source   string
}

// QuerySimilar finds the n most similar articles to a given embedding vector
func (c *ChromaClient) QuerySimilar(ctx context.Context, queryEmbedding []float32, nResults int) ([]QueryResult, error) {
	if c.collectionID == "" {
		return nil, fmt.Errorf("collection not initialized")
	}

	payload := map[string]interface{}{
		"query_embeddings": [][]float32{queryEmbedding},
		"n_results":        nResults,
		"include":          []string{"metadatas", "documents", "distances"},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chroma query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chroma returned status %d on query", resp.StatusCode)
	}

	// Chroma query returns arrays of arrays since you can pass multiple query embeddings
	var result struct {
		IDs       [][]string                   `json:"ids"`
		Distances [][]float32                  `json:"distances"`
		Metadatas [][]map[string]interface{}   `json:"metadatas"`
		Documents [][]string                   `json:"documents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var matches []QueryResult
	if len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i := range result.IDs[0] {
			title, _ := result.Metadatas[0][i]["title"].(string)
			source, _ := result.Metadatas[0][i]["source"].(string)
			
			matches = append(matches, QueryResult{
				ID:       result.IDs[0][i],
				Distance: result.Distances[0][i],
				Title:    title,
				Summary:  result.Documents[0][i],
				Source:   source,
			})
		}
	}

	return matches, nil
}
