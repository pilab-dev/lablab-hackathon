package config

import (
	"os"
	"testing"
)

func TestLoadConfig_LLMProviderDefault(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMProvider != "ollama" {
		t.Errorf("Expected default LLMProvider 'ollama', got %q", cfg.LLMProvider)
	}
}

func TestLoadConfig_LLMModelDefault(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMModel != "llama3.1:8b" {
		t.Errorf("Expected default LLMModel 'llama3.1:8b', got %q", cfg.LLMModel)
	}
}

func TestLoadConfig_LMStudioURLDefault(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LMStudioURL != "http://localhost:1234" {
		t.Errorf("Expected default LMStudioURL 'http://localhost:1234', got %q", cfg.LMStudioURL)
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
		t.Errorf("Expected LLMProvider 'lmstudio' from env, got %q", cfg.LLMProvider)
	}
}

func TestLoadConfig_LLMModelFromEnv(t *testing.T) {
	os.Setenv("LLM_MODEL", "phi-3-mini")
	defer os.Unsetenv("LLM_MODEL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LLMModel != "phi-3-mini" {
		t.Errorf("Expected LLMModel 'phi-3-mini' from env, got %q", cfg.LLMModel)
	}
}

func TestLoadConfig_LMStudioURLFromEnv(t *testing.T) {
	os.Setenv("LMSTUDIO_URL", "http://myserver:5678")
	defer os.Unsetenv("LMSTUDIO_URL")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.LMStudioURL != "http://myserver:5678" {
		t.Errorf("Expected LMStudioURL 'http://myserver:5678' from env, got %q", cfg.LMStudioURL)
	}
}

// Verify pre-existing Ollama fields are still present alongside new LLM fields
func TestLoadConfig_OllamaFieldsCoexistWithLLMFields(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("Expected OllamaURL default 'http://localhost:11434', got %q", cfg.OllamaURL)
	}
	if cfg.OllamaModel != "llama3.1:8b" {
		t.Errorf("Expected OllamaModel default 'llama3.1:8b', got %q", cfg.OllamaModel)
	}
	if cfg.LLMProvider == "" {
		t.Error("LLMProvider should not be empty")
	}
	if cfg.LMStudioURL == "" {
		t.Error("LMStudioURL should not be empty")
	}
}