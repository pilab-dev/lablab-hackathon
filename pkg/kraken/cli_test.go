package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func TestClientTimeoutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	t.Cleanup(cancel)

	// Wait for context to expire
	<-ctx.Done()

	t.Setenv("GO_WANT_KRAKEN_HELPER_PROCESS", "1")
	t.Setenv("KRAKEN_HELPER_STDOUT", "{}\n")
	t.Setenv("KRAKEN_HELPER_STDERR", "")
	t.Setenv("KRAKEN_HELPER_EXIT_CODE", "0")

	c := NewClient(os.Args[0])
	_, err := c.RunRaw(ctx, "-test.run=TestHelperProcess", "--", "__kraken_cli_test_helper__")

	if err == nil {
		t.Error("Expected error due to context cancellation/timeout, but got nil")
	}
}

type commandResponseCase struct {
	Name                string         `json:"name"`
	Fn                  string         `json:"fn"`
	Stdout              string         `json:"stdout"`
	Stderr              string         `json:"stderr"`
	ExitCode            int            `json:"exit_code"`
	ExpectErrorContains string         `json:"expect_error_contains"`
	ExpectMap           map[string]any `json:"expect_map"`
}

func assertMapContains(got map[string]any, want map[string]any) bool {
	for wantKey, wantVal := range want {
		gotVal, ok := got[wantKey]
		if !ok {
			return false
		}
		wantMap, wantIsMap := wantVal.(map[string]any)
		gotMap, gotIsMap := gotVal.(map[string]any)
		if wantIsMap && gotIsMap {
			if !assertMapContains(gotMap, wantMap) {
				return false
			}
		} else if fmt.Sprintf("%v", gotVal) != fmt.Sprintf("%v", wantVal) {
			return false
		}
	}
	return true
}

func TestCommandResponseCasesFromJSON(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "command_response_cases.json"))
	if err != nil {
		t.Fatalf("failed to read command response cases json: %v", err)
	}

	var cases []commandResponseCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal command response cases json: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found in command_response_cases.json")
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Setenv("GO_WANT_KRAKEN_HELPER_PROCESS", "1")
			t.Setenv("KRAKEN_HELPER_STDOUT", tc.Stdout)
			t.Setenv("KRAKEN_HELPER_STDERR", tc.Stderr)
			t.Setenv("KRAKEN_HELPER_EXIT_CODE", fmt.Sprintf("%d", tc.ExitCode))

			c := NewClient(os.Args[0])

			switch tc.Fn {
			case "run_raw":
				_, err := c.RunRaw(context.Background(), "-test.run=TestHelperProcess", "--", "__kraken_cli_test_helper__")
				if tc.ExpectErrorContains != "" {
					if err == nil || !strings.Contains(err.Error(), tc.ExpectErrorContains) {
						t.Fatalf("expected error containing %q, got: %v", tc.ExpectErrorContains, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			case "run":
				var got map[string]any
				err := c.Run(context.Background(), &got, "-test.run=TestHelperProcess", "--", "__kraken_cli_test_helper__")
				if tc.ExpectErrorContains != "" {
					if err == nil || !strings.Contains(err.Error(), tc.ExpectErrorContains) {
						t.Fatalf("expected error containing %q, got: %v", tc.ExpectErrorContains, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if tc.ExpectMap != nil {
					if !assertMapContains(got, tc.ExpectMap) {
						t.Fatalf("map does not contain expected values: want %v, got %v", tc.ExpectMap, got)
					}
				}
			default:
				t.Fatalf("unknown fn %q", tc.Fn)
			}
		})
	}
}

// TestHelperProcess is used to mock the kraken-cli binary.
// It is invoked by setting Client.binPath = os.Args[0] and passing "__kraken_cli_test_helper__" as the first arg.
//
// Based on https://pkg.go.dev/os/exec#hdr-Testing_External_Commands
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_KRAKEN_HELPER_PROCESS") != "1" {
		return
	}
	if len(os.Args) < 2 {
		os.Exit(2)
	}

	// Find the marker arg. exec.CommandContext will pass: <testbin> <args...> "-o" "json"
	isHelper := false
	for _, a := range os.Args[1:] {
		if a == "__kraken_cli_test_helper__" {
			isHelper = true
			break
		}
	}
	if !isHelper {
		return
	}

	_, _ = os.Stdout.WriteString(os.Getenv("KRAKEN_HELPER_STDOUT"))
	_, _ = os.Stderr.WriteString(os.Getenv("KRAKEN_HELPER_STDERR"))

	exitCodeStr := os.Getenv("KRAKEN_HELPER_EXIT_CODE")
	if exitCodeStr == "" {
		os.Exit(0)
	}
	var exitCode int
	if _, err := fmt.Sscanf(exitCodeStr, "%d", &exitCode); err != nil {
		os.Exit(2)
	}
	os.Exit(exitCode)
}
