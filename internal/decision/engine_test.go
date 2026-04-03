package decision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewEngine_DefaultProvider(t *testing.T) {
	engine := NewEngine("", "", "", nil, nil, nil, nil, nil)
	if engine.provider != "ollama" {
		t.Errorf("expected default provider 'ollama', got %q", engine.provider)
	}
}

func TestNewEngine_DefaultBaseURL_Ollama(t *testing.T) {
	engine := NewEngine("ollama", "", "", nil, nil, nil, nil, nil)
	if engine.baseURL != "http://localhost:11434" {
		t.Errorf("expected default ollama baseURL 'http://localhost:11434', got %q", engine.baseURL)
	}
}

func TestNewEngine_DefaultBaseURL_EmptyProviderFallback(t *testing.T) {
	// Empty provider defaults to "ollama", so baseURL should default to ollama's URL
	engine := NewEngine("", "", "", nil, nil, nil, nil, nil)
	if engine.baseURL != "http://localhost:11434" {
		t.Errorf("expected default baseURL 'http://localhost:11434', got %q", engine.baseURL)
	}
}

func TestNewEngine_DefaultBaseURL_LMStudio(t *testing.T) {
	engine := NewEngine("lmstudio", "", "", nil, nil, nil, nil, nil)
	if engine.baseURL != "http://localhost:1234" {
		t.Errorf("expected default lmstudio baseURL 'http://localhost:1234', got %q", engine.baseURL)
	}
}

func TestNewEngine_DefaultModel(t *testing.T) {
	engine := NewEngine("", "", "", nil, nil, nil, nil, nil)
	if engine.model != "llama3.1:8b" {
		t.Errorf("expected default model 'llama3.1:8b', got %q", engine.model)
	}
}

func TestNewEngine_ExplicitValues(t *testing.T) {
	engine := NewEngine("lmstudio", "http://myhost:1234", "my-model", nil, nil, nil, nil, nil)
	if engine.provider != "lmstudio" {
		t.Errorf("expected provider 'lmstudio', got %q", engine.provider)
	}
	if engine.baseURL != "http://myhost:1234" {
		t.Errorf("expected baseURL 'http://myhost:1234', got %q", engine.baseURL)
	}
	if engine.model != "my-model" {
		t.Errorf("expected model 'my-model', got %q", engine.model)
	}
}

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

func TestEngine_CallLMStudio_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status 500, got: %v", err)
	}
}

func TestEngine_CallLMStudio_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": []}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("expected 'no choices' error, got: %v", err)
	}
}

func TestEngine_CallLMStudio_InvalidResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for invalid response body, got nil")
	}
}

func TestEngine_CallLMStudio_InvalidLLMContentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// The outer JSON is valid but the content string is not valid JSON
		w.Write([]byte(`{"choices": [{"message": {"content": "not json"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for invalid LLM content JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse LLM decisions JSON") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestEngine_CallLMStudio_NetworkError(t *testing.T) {
	// Use a server that is immediately closed so connections fail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

func TestEngine_CallLMStudio_VerifiesRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.3, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\"}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngine_CallOllama_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to mention status 400, got: %v", err)
	}
}

func TestEngine_CallOllama_InvalidResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{bad json`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for invalid response body, got nil")
	}
}

func TestEngine_CallOllama_InvalidLLMContentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "not valid json at all"}}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for invalid LLM content JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse LLM decisions JSON") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestEngine_CallOllama_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

func TestEngine_ProcessLLMResponse_SizingQUARTER(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"buy\", \"confidence\": 0.4, \"sizing\": \"QUARTER\", \"reasoning\": {\"primary_signal\": \"Weak signal\", \"supporting\": []}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].SizePct != 2.5 {
		t.Errorf("expected QUARTER size_pct 2.5, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_SizingSKIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"hold\", \"confidence\": 0.2, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"No signal\", \"supporting\": []}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].SizePct != 0.0 {
		t.Errorf("expected SKIP size_pct 0.0, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_SizingFULL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"buy\", \"confidence\": 0.9, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"Strong signal\"}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].SizePct != 10.0 {
		t.Errorf("expected FULL size_pct 10.0, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_EmptyAction(t *testing.T) {
	// When action is empty string, processLLMResponse returns an empty decisions slice
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"\", \"confidence\": 0.1, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"No action\"}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected 0 decisions for empty action, got %d", len(decisions))
	}
}

func TestEngine_ProcessLLMResponse_ReasoningConcatenated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"action\": \"sell\", \"confidence\": 0.7, \"sizing\": \"HALF\", \"reasoning\": {\"primary_signal\": \"Overbought\", \"supporting\": [\"RSI > 70\"]}}"
				}
			}]
		}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	expected := "Overbought; RSI > 70"
	if decisions[0].Reasoning != expected {
		t.Errorf("expected reasoning %q, got %q", expected, decisions[0].Reasoning)
	}
}

func TestEngine_CallLMStudio_VerifiesRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if reqBody["stream"] != false {
			t.Errorf("expected stream=false, got %v", reqBody["stream"])
		}
		temp, ok := reqBody["temperature"].(float64)
		if !ok || temp != 0.1 {
			t.Errorf("expected temperature=0.1, got %v", reqBody["temperature"])
		}
		if reqBody["model"] != "custom-model" {
			t.Errorf("expected model='custom-model', got %v", reqBody["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"HALF\", \"reasoning\": {\"primary_signal\": \"ok\"}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "custom-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngine_CallOllama_VerifiesRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if reqBody["stream"] != false {
			t.Errorf("expected stream=false, got %v", reqBody["stream"])
		}
		if reqBody["format"] != "json" {
			t.Errorf("expected format='json', got %v", reqBody["format"])
		}
		if reqBody["model"] != "llama3.1:8b" {
			t.Errorf("expected model='llama3.1:8b', got %v", reqBody["model"])
		}
		options, ok := reqBody["options"].(map[string]interface{})
		if !ok {
			t.Errorf("expected options map, got %T", reqBody["options"])
		} else if options["temperature"] != 0.1 {
			t.Errorf("expected temperature=0.1, got %v", options["temperature"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "{\"action\": \"buy\", \"confidence\": 0.8, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"ok\"}}"}}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "BTCUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngine_CallLMStudio_PairPropagatedToDecision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"buy\", \"confidence\": 0.8, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"ok\"}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "XBTUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Pair != "XBTUSD" {
		t.Errorf("expected pair 'XBTUSD', got %q", decisions[0].Pair)
	}
}

func TestEngine_CallOllama_PairPropagatedToDecision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "{\"action\": \"sell\", \"confidence\": 0.9, \"sizing\": \"HALF\", \"reasoning\": {\"primary_signal\": \"ok\"}}"}}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	decisions, err := engine.callOllama(context.Background(), "ETHUSD", "Test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Pair != "ETHUSD" {
		t.Errorf("expected pair 'ETHUSD', got %q", decisions[0].Pair)
	}
}

func TestEngine_CallLMStudio_ContentTruncatedInErrorLog(t *testing.T) {
	// Content longer than 100 chars should not cause a panic during error logging
	longInvalidContent := strings.Repeat("x", 200)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"content": longInvalidContent}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "Test prompt")
	if err == nil {
		t.Fatal("expected error for invalid JSON content, got nil")
	}
}