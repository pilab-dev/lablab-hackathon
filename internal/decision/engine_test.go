package decision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine_CallLMStudio(t *testing.T) {
	// Mock LM Studio server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path /v1/chat/completions, got %s", r.URL.Path)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}

		// Verify JSON mode is requested
		respFormat, ok := reqBody["response_format"].(map[string]interface{})
		if !ok || respFormat["type"] != "json_object" {
			t.Errorf("Expected response_format type json_object, got %v", respFormat)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"buy\", \"confidence\": 0.8, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"Strong momentum\", \"supporting\": [\"RSI is low\"]}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")

	if err != nil {
		t.Fatalf("callLMStudio failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("Expected 1 decision, got %d", len(decisions))
	}

	if decisions[0].Action != "buy" {
		t.Errorf("Expected action buy, got %s", decisions[0].Action)
	}

	if decisions[0].Confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", decisions[0].Confidence)
	}
}

func TestEngine_CallOllama(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path /api/chat, got %s", r.URL.Path)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}

		// Verify format json is requested
		if reqBody["format"] != "json" {
			t.Errorf("Expected format json, got %v", reqBody["format"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": {
				"content": "{\"action\": \"sell\", \"confidence\": 0.9, \"sizing\": \"HALF\", \"reasoning\": {\"primary_signal\": \"Overbought\", \"supporting\": [\"RSI is high\"]}}"
			}
		}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	decisions, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")

	if err != nil {
		t.Fatalf("callOllama failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("Expected 1 decision, got %d", len(decisions))
	}

	if decisions[0].Action != "sell" {
		t.Errorf("Expected action sell, got %s", decisions[0].Action)
	}

	if decisions[0].SizePct != 5.0 {
		t.Errorf("Expected size_pct 5.0, got %f", decisions[0].SizePct)
	}
}
