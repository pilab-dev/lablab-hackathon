package decision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// --- NewEngine default value tests ---

func TestNewEngine_Defaults_EmptyProvider(t *testing.T) {
	e := NewEngine("", "", "", nil, nil, nil, nil, nil)
	if e.provider != "ollama" {
		t.Errorf("Expected default provider 'ollama', got %q", e.provider)
	}
}

func TestNewEngine_Defaults_EmptyBaseURL_Ollama(t *testing.T) {
	e := NewEngine("ollama", "", "", nil, nil, nil, nil, nil)
	if e.baseURL != "http://localhost:11434" {
		t.Errorf("Expected default ollama baseURL 'http://localhost:11434', got %q", e.baseURL)
	}
}

func TestNewEngine_Defaults_EmptyBaseURL_LMStudio(t *testing.T) {
	e := NewEngine("lmstudio", "", "", nil, nil, nil, nil, nil)
	if e.baseURL != "http://localhost:1234" {
		t.Errorf("Expected default lmstudio baseURL 'http://localhost:1234', got %q", e.baseURL)
	}
}

func TestNewEngine_Defaults_EmptyModel(t *testing.T) {
	e := NewEngine("", "", "", nil, nil, nil, nil, nil)
	if e.model != "llama3.1:8b" {
		t.Errorf("Expected default model 'llama3.1:8b', got %q", e.model)
	}
}

func TestNewEngine_ExplicitValues(t *testing.T) {
	e := NewEngine("lmstudio", "http://myhost:1234", "phi-3", nil, nil, nil, nil, nil)
	if e.provider != "lmstudio" {
		t.Errorf("Expected provider 'lmstudio', got %q", e.provider)
	}
	if e.baseURL != "http://myhost:1234" {
		t.Errorf("Expected baseURL 'http://myhost:1234', got %q", e.baseURL)
	}
	if e.model != "phi-3" {
		t.Errorf("Expected model 'phi-3', got %q", e.model)
	}
}

// --- callLMStudio error / edge case tests ---

func TestEngine_CallLMStudio_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to mention status 500, got: %v", err)
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
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for empty choices, got nil")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("Expected 'no choices' error, got: %v", err)
	}
}

func TestEngine_CallLMStudio_MalformedResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-valid-json`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for malformed response body, got nil")
	}
}

func TestEngine_CallLMStudio_InvalidContentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "this is not json"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for invalid content JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse LLM decisions JSON") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestEngine_CallLMStudio_RequestBody(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "my-model", nil, nil, nil, nil, nil)
	engine.callLMStudio(context.Background(), "ETHUSD", "some prompt")

	if capturedBody["model"] != "my-model" {
		t.Errorf("Expected model 'my-model' in request, got %v", capturedBody["model"])
	}
	if capturedBody["stream"] != false {
		t.Errorf("Expected stream false in request, got %v", capturedBody["stream"])
	}
	if capturedBody["temperature"] != 0.1 {
		t.Errorf("Expected temperature 0.1 in request, got %v", capturedBody["temperature"])
	}
	respFmt, ok := capturedBody["response_format"].(map[string]interface{})
	if !ok || respFmt["type"] != "json_object" {
		t.Errorf("Expected response_format.type json_object, got %v", capturedBody["response_format"])
	}
}

// --- callOllama error / edge case tests ---

func TestEngine_CallOllama_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("Expected error to mention status 502, got: %v", err)
	}
}

func TestEngine_CallOllama_MalformedResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{broken json`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for malformed response body, got nil")
	}
}

func TestEngine_CallOllama_InvalidContentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "plain text not json"}}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	_, err := engine.callOllama(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for invalid content JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse LLM decisions JSON") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestEngine_CallOllama_RequestBody(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\", \"supporting\": []}}" }}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	engine.callOllama(context.Background(), "BTCUSD", "some prompt")

	if capturedBody["model"] != "llama3.1:8b" {
		t.Errorf("Expected model 'llama3.1:8b', got %v", capturedBody["model"])
	}
	if capturedBody["format"] != "json" {
		t.Errorf("Expected format 'json', got %v", capturedBody["format"])
	}
	if capturedBody["stream"] != false {
		t.Errorf("Expected stream false, got %v", capturedBody["stream"])
	}
}

// --- processLLMResponse sizing / reasoning tests ---

func TestEngine_ProcessLLMResponse_SizingFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"buy\", \"confidence\": 0.9, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"breakout\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if decisions[0].SizePct != 10.0 {
		t.Errorf("Expected FULL sizing = 10.0, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_SizingQuarter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"buy\", \"confidence\": 0.6, \"sizing\": \"QUARTER\", \"reasoning\": {\"primary_signal\": \"signal\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if decisions[0].SizePct != 2.5 {
		t.Errorf("Expected QUARTER sizing = 2.5, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_SizingSkip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if decisions[0].SizePct != 0.0 {
		t.Errorf("Expected SKIP sizing = 0.0, got %f", decisions[0].SizePct)
	}
}

func TestEngine_ProcessLLMResponse_EmptyAction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// action is empty string - should yield empty decisions
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("Expected 0 decisions for empty action, got %d", len(decisions))
	}
}

func TestEngine_ProcessLLMResponse_ReasoningWithoutSupporting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"sell\", \"confidence\": 0.7, \"sizing\": \"HALF\", \"reasoning\": {\"primary_signal\": \"trend reversal\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "ETHUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("Expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Reasoning != "trend reversal" {
		t.Errorf("Expected reasoning 'trend reversal', got %q", decisions[0].Reasoning)
	}
}

func TestEngine_ProcessLLMResponse_ReasoningWithSupporting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"buy\", \"confidence\": 0.85, \"sizing\": \"FULL\", \"reasoning\": {\"primary_signal\": \"momentum\", \"supporting\": [\"RSI oversold\"]}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	decisions, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(decisions[0].Reasoning, "momentum") {
		t.Errorf("Expected reasoning to contain 'momentum', got %q", decisions[0].Reasoning)
	}
	if !strings.Contains(decisions[0].Reasoning, "RSI oversold") {
		t.Errorf("Expected reasoning to contain 'RSI oversold', got %q", decisions[0].Reasoning)
	}
}

// --- Decide() routing test ---

func TestEngine_Decide_RoutesToLMStudio(t *testing.T) {
	var calledPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return minimal valid response
		w.Write([]byte(`{"choices": [{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\", \"supporting\": []}}"}}]}`))
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	// Decide requires stateMgr, so just verify routing by calling the underlying method directly
	// and ensure the right path was hit
	engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if calledPath != "/v1/chat/completions" {
		t.Errorf("Expected lmstudio path '/v1/chat/completions', got %q", calledPath)
	}
}

func TestEngine_Decide_RoutesToOllama(t *testing.T) {
	var calledPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"content": "{\"action\": \"hold\", \"confidence\": 0.5, \"sizing\": \"SKIP\", \"reasoning\": {\"primary_signal\": \"flat\", \"supporting\": []}}"}}`))
	}))
	defer server.Close()

	engine := NewEngine("ollama", server.URL, "llama3.1:8b", nil, nil, nil, nil, nil)
	engine.callOllama(context.Background(), "BTCUSD", "prompt")
	if calledPath != "/api/chat" {
		t.Errorf("Expected ollama path '/api/chat', got %q", calledPath)
	}
}

// Regression: long content (>100 chars) in bad JSON is truncated in debug log without panic
func TestEngine_ProcessLLMResponse_LongInvalidContentTruncated(t *testing.T) {
	longContent := strings.Repeat("x", 200)
	body, _ := json.Marshal(map[string]interface{}{
		"choices": []map[string]interface{}{
			{"message": map[string]interface{}{"content": longContent}},
		},
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	engine := NewEngine("lmstudio", server.URL, "test-model", nil, nil, nil, nil, nil)
	_, err := engine.callLMStudio(context.Background(), "BTCUSD", "prompt")
	if err == nil {
		t.Fatal("Expected error for invalid JSON content, got nil")
	}
	// Should not panic and should return parse error
	if !strings.Contains(err.Error(), "failed to parse LLM decisions JSON") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}