package kraken

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// DefaultTimeout is the default time limit for kraken-cli commands
const DefaultTimeout = 30 * time.Second

// ErrorEnvelope matches the standard kraken-cli JSON error structure
type ErrorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Client wraps the kraken-cli executable
type Client struct {
	binPath string
}

// NewClient creates a new Kraken CLI wrapper.
// If binPath is empty, it assumes "kraken" is in the system PATH.
func NewClient(binPath string) *Client {
	if binPath == "" {
		binPath = "kraken"
	}
	return &Client{
		binPath: binPath,
	}
}

// RunRaw executes a kraken-cli command and returns the raw JSON bytes.
// It automatically appends "-o json" to the arguments.
func (c *Client) RunRaw(ctx context.Context, args ...string) ([]byte, error) {
	// Ensure we always request JSON output
	fullArgs := append(args, "-o", "json")

	cmd := exec.CommandContext(ctx, c.binPath, fullArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// If the command failed, or even if it succeeded but we want to check for API errors
	if err != nil {
		// Try to parse the stdout as a structured error envelope
		if stdout.Len() > 0 {
			var errEnv ErrorEnvelope
			if jsonErr := json.Unmarshal(stdout.Bytes(), &errEnv); jsonErr == nil && errEnv.Error != "" {
				return nil, fmt.Errorf("kraken-cli error [%s]: %s", errEnv.Error, errEnv.Message)
			}
		}
		// Fallback error reporting
		return nil, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// Run executes a kraken-cli command and unmarshals the JSON output into the target interface.
func (c *Client) Run(ctx context.Context, target interface{}, args ...string) error {
	out, err := c.RunRaw(ctx, args...)
	if err != nil {
		return err
	}

	// Unmarshal into the provided struct
	if err := json.Unmarshal(out, target); err != nil {
		return fmt.Errorf("failed to parse kraken-cli json response: %w\nRaw output: %s", err, string(out))
	}

	return nil
}

// RunStream executes a long-running kraken-cli command (like WebSockets)
// and streams the NDJSON output line-by-line to a callback function.
func (c *Client) RunStream(ctx context.Context, callback func([]byte), args ...string) error {
	fullArgs := append(args, "-o", "json")

	cmd := exec.CommandContext(ctx, c.binPath, fullArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start streaming command: %w", err)
	}

	// Read line by line asynchronously
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Bytes()
			// Skip empty lines
			if len(bytes.TrimSpace(line)) > 0 {
				callback(line)
			}
		}

		// When scanner ends, wait for the command to finish
		cmd.Wait()
	}()

	return nil
}

type BalanceResponse struct {
	Error  interface{}        `json:"error"`
	Result map[string]float64 `json:"result,omitempty"`
}

func (c *Client) GetBalance(ctx context.Context) (map[string]float64, error) {
	var resp BalanceResponse
	if err := c.Run(ctx, &resp, "balance"); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("balance error: %v", resp.Error)
	}

	return resp.Result, nil
}
