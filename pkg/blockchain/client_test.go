package blockchain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_MissingRPCURL(t *testing.T) {
	cfg := Config{
		OperatorPK: "abc123",
		AgentPK:    "def456",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SEPOLIA_RPC_URL is required")
}

func TestNewClient_MissingOperatorPK(t *testing.T) {
	cfg := Config{
		RPCURL:  "http://localhost:8545",
		AgentPK: "def456",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPERATOR_PRIVATE_KEY is required")
}

func TestNewClient_MissingAgentPK(t *testing.T) {
	cfg := Config{
		RPCURL:     "http://localhost:8545",
		OperatorPK: "abc123",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AGENT_PRIVATE_KEY is required")
}

func TestNewClient_InvalidOperatorKey(t *testing.T) {
	cfg := Config{
		RPCURL:     "http://localhost:8545",
		OperatorPK: "xyz-not-hex",
		AgentPK:    "def456",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse operator private key")
}

func TestNewClient_InvalidAgentID(t *testing.T) {
	cfg := Config{
		RPCURL:     "http://localhost:8545",
		OperatorPK: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		AgentPK:    "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
		AgentID:    "not-a-number",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse AGENT_ID")
}

func TestNewClient_ValidAgentID(t *testing.T) {
	cfg := Config{
		RPCURL:     "http://localhost:8545",
		OperatorPK: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		AgentPK:    "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
		AgentID:    "42",
	}

	_, err := NewClient(context.Background(), cfg)
	require.Error(t, err)
}

func TestClient_Close_NilRPC(t *testing.T) {
	c := &Client{}
	c.Close()
}
