package kraken

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// We won't test the actual kraken binary here to avoid CI/CD issues and missing dependencies,
// but we will test the error unmarshaling logic.
func TestErrorEnvelope_Unmarshal(t *testing.T) {
	rawJSON := `{"error": "rate_limit", "message": "API rate limit exceeded"}`

	var env ErrorEnvelope
	if err := json.Unmarshal([]byte(rawJSON), &env); err != nil {
		t.Fatalf("Failed to unmarshal error envelope: %v", err)
	}

	if env.Error != "rate_limit" {
		t.Errorf("Expected error 'rate_limit', got '%s'", env.Error)
	}

	if !strings.Contains(env.Message, "limit exceeded") {
		t.Errorf("Unexpected message content: %s", env.Message)
	}
}

// Ensure the client can be instantiated
func TestNewClient(t *testing.T) {
	c := NewClient("")
	if c.binPath != "kraken" {
		t.Errorf("Expected default binPath 'kraken', got '%s'", c.binPath)
	}

	c2 := NewClient("/usr/local/bin/kraken")
	if c2.binPath != "/usr/local/bin/kraken" {
		t.Errorf("Expected custom binPath, got '%s'", c2.binPath)
	}
}

// We can test context timeout handling concept, though mocking exec.Command is complex in Go
// without abstracting the exec.Cmd creation. The context integration is standard lib functionality.
func TestClientTimeoutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	<-ctx.Done()

	c := NewClient("sleep") // fake command
	_, err := c.RunRaw(ctx, "1")

	if err == nil {
		t.Error("Expected error due to context cancellation/timeout, but got nil")
	}
}
