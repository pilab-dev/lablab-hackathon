package executor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pilab-dev/lablab-hackathon/internal/decision"
	"github.com/pilab-dev/lablab-hackathon/pkg/blockchain"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ExecutorFlow(t *testing.T) {
	if os.Getenv("SEPOLIA_RPC_URL") == "" {
		t.Skip("Skipping integration test: SEPOLIA_RPC_URL not set")
	}

	ctx := context.Background()

	// Setup with real blockchain client
	cfg := blockchain.Config{
		RPCURL:     os.Getenv("SEPOLIA_RPC_URL"),
		ChainID:    11155111,
		OperatorPK: os.Getenv("OPERATOR_PRIVATE_KEY"),
		AgentPK:    os.Getenv("AGENT_PRIVATE_KEY"),
		AgentID:    os.Getenv("AGENT_ID"),
		Contracts: blockchain.ContractAddresses{
			AgentRegistry:      "0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3",
			HackathonVault:     "0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90",
			RiskRouter:         "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
			ReputationRegistry: "0x423a9904e39537a9997fbaF0f220d79D7d545763",
			ValidationRegistry: "0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1",
		},
	}

	bcClient, err := blockchain.NewClient(ctx, cfg)
	require.NoError(t, err)
	defer bcClient.Close()

	executor := New(
		bcClient,
		nil, // mockMarketData
		nil, // mockPortfolio
		nil, // mockRecorder
		Config{
			MaxSlippageBps:  100,
			IntentDeadline:  1 * time.Hour,
			MinConfidence:   0.5,
			TradingMode:     "paper",
		},
	)

	// Execute a trade decision
	decision := &decision.TradeDecision{
		Pair:       "BTCUSD",
		Action:     "buy",
		SizePct:    10.0,
		Confidence: 0.8,
		Reasoning:  "Integration test trade",
	}

	result, err := executor.Execute(ctx, decision)
	require.NoError(t, err)

	if result != nil {
		t.Logf("Trade result: approved=%v, tx_hash=%s", result.Approved, result.TxHash.Hex())
	}
}

func TestIntegration_CircuitBreakerOpensOnFailures(t *testing.T) {
	// This test requires a mock RPC that fails 3 times
	// In practice, this would need a more sophisticated mock
	t.Skip("Manual test: requires controlled failure injection")
}
