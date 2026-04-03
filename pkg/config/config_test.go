package config

import (
	"os"
	"testing"
)

func TestLoadConfig_LLMProviderDefault(t *testing.T) {
	// Ensure env vars are not set so defaults kick in
	os.Unsetenv("LLM_PROVIDER")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMProvider != "ollama" {
		t.Errorf("expected default LLMProvider 'ollama', got %q", cfg.LLMProvider)
	}
}

func TestLoadConfig_LLMModelDefault(t *testing.T) {
	os.Unsetenv("LLM_MODEL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMModel != "llama3.1:8b" {
		t.Errorf("expected default LLMModel 'llama3.1:8b', got %q", cfg.LLMModel)
	}
}

func TestLoadConfig_LMStudioURLDefault(t *testing.T) {
	os.Unsetenv("LMSTUDIO_URL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LMStudioURL != "http://localhost:1234" {
		t.Errorf("expected default LMStudioURL 'http://localhost:1234', got %q", cfg.LMStudioURL)
	}
}

func TestLoadConfig_LLMProviderFromEnv(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "lmstudio")
	defer os.Unsetenv("LLM_PROVIDER")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMProvider != "lmstudio" {
		t.Errorf("expected LLMProvider 'lmstudio', got %q", cfg.LLMProvider)
	}
}

func TestLoadConfig_LLMModelFromEnv(t *testing.T) {
	os.Setenv("LLM_MODEL", "mistral:7b")
	defer os.Unsetenv("LLM_MODEL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMModel != "mistral:7b" {
		t.Errorf("expected LLMModel 'mistral:7b', got %q", cfg.LLMModel)
	}
}

func TestLoadConfig_LMStudioURLFromEnv(t *testing.T) {
	os.Setenv("LMSTUDIO_URL", "http://myhost:5678")
	defer os.Unsetenv("LMSTUDIO_URL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LMStudioURL != "http://myhost:5678" {
		t.Errorf("expected LMStudioURL 'http://myhost:5678', got %q", cfg.LMStudioURL)
	}
}

func TestLoadConfig_AllNewLLMFieldsPresent(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "lmstudio")
	os.Setenv("LLM_MODEL", "phi-3")
	os.Setenv("LMSTUDIO_URL", "http://studio:9000")
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("LLM_MODEL")
		os.Unsetenv("LMSTUDIO_URL")
	}()

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMProvider != "lmstudio" {
		t.Errorf("expected LLMProvider 'lmstudio', got %q", cfg.LLMProvider)
	}
	if cfg.LLMModel != "phi-3" {
		t.Errorf("expected LLMModel 'phi-3', got %q", cfg.LLMModel)
	}
	if cfg.LMStudioURL != "http://studio:9000" {
		t.Errorf("expected LMStudioURL 'http://studio:9000', got %q", cfg.LMStudioURL)
	}
}

// Regression: existing Ollama fields must not be affected by the new LLM fields
func TestLoadConfig_OllamaFieldsUnaffected(t *testing.T) {
	os.Unsetenv("OLLAMA_URL")
	os.Unsetenv("OLLAMA_MODEL")
	os.Unsetenv("OLLAMA_EMBED_MODEL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("expected OllamaURL 'http://localhost:11434', got %q", cfg.OllamaURL)
	}
	if cfg.OllamaModel != "llama3.1:8b" {
		t.Errorf("expected OllamaModel 'llama3.1:8b', got %q", cfg.OllamaModel)
	}
	if cfg.OllamaEmbedModel != "nomic-embed-text" {
		t.Errorf("expected OllamaEmbedModel 'nomic-embed-text', got %q", cfg.OllamaEmbedModel)
	}
}