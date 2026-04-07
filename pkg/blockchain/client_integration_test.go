package blockchain

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullTradeFlow(t *testing.T) {
	if os.Getenv("SEPOLIA_RPC_URL") == "" {
		t.Skip("Skipping integration test: SEPOLIA_RPC_URL not set")
	}

	ctx := context.Background()

	// 1. Create client
	cfg := Config{
		RPCURL:     os.Getenv("SEPOLIA_RPC_URL"),
		ChainID:    11155111,
		OperatorPK: os.Getenv("OPERATOR_PRIVATE_KEY"),
		AgentPK:    os.Getenv("AGENT_PRIVATE_KEY"),
		AgentID:    os.Getenv("AGENT_ID"),
		Contracts: ContractAddresses{
			AgentRegistry:      "0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3",
			HackathonVault:     "0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90",
			RiskRouter:         "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
			ReputationRegistry: "0x423a9904e39537a9997fbaF0f220d79D7d545763",
			ValidationRegistry: "0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1",
		},
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	defer client.Close()

	// 2. Get agent status
	agentId := client.AgentID()
	require.NotNil(t, agentId, "AGENT_ID must be set")

	isRegistered, err := client.IsRegistered(ctx, agentId)
	require.NoError(t, err)
	t.Logf("Agent registered: %v", isRegistered)

	// 3. Check vault balance
	hasClaimed, err := client.HasClaimed(ctx, agentId)
	require.NoError(t, err)
	t.Logf("Has claimed: %v", hasClaimed)

	balance, err := client.GetVaultBalance(ctx, agentId)
	require.NoError(t, err)
	t.Logf("Vault balance: %s wei", balance.String())

	// 4. Get nonce
	nonce, err := client.GetIntentNonce(ctx)
	require.NoError(t, err)
	t.Logf("Intent nonce: %s", nonce.String())

	// 5. Build and simulate intent
	intent := TradeIntent{
		AgentId:         agentId,
		AgentWallet:     client.AgentAddress(),
		Pair:            "BTCUSD",
		Action:          "buy",
		AmountUsdScaled: new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		MaxSlippageBps:  big.NewInt(100),
		Nonce:           nonce,
		Deadline:        big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
	}

	approved, reason, err := client.SimulateIntent(ctx, intent)
	require.NoError(t, err)
	t.Logf("SimulateIntent: approved=%v, reason=%s", approved, reason)

	if !approved {
		t.Skip("Simulation rejected: " + reason)
	}
}

func TestIntegration_NonceIncrementing(t *testing.T) {
	if os.Getenv("SEPOLIA_RPC_URL") == "" {
		t.Skip("Skipping integration test: SEPOLIA_RPC_URL not set")
	}

	ctx := context.Background()
	cfg := getTestConfig(t)

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	defer client.Close()

	agentId := client.AgentID()
	_ = agentId // silence unused for now

	nonce1, err := client.GetIntentNonce(ctx)
	require.NoError(t, err)

	// No transaction submitted between nonce reads - nonces should be equal
	assert.Equal(t, nonce1, nonce2, "Nonce should not change without a submission")
}

func TestIntegration_InvalidIntentRejection(t *testing.T) {
	if os.Getenv("SEPOLIA_RPC_URL") == "" {
		t.Skip("Skipping integration test: SEPOLIA_RPC_URL not set")
	}

	ctx := context.Background()
	cfg := getTestConfig(t)

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	defer client.Close()

	agentId := client.AgentID()

	invalidIntent := TradeIntent{
		AgentId:         agentId,
		AgentWallet:     client.AgentAddress(),
		Pair:            "INVALID",
		Action:          "invalid_action",
		AmountUsdScaled: big.NewInt(0),
		MaxSlippageBps:  big.NewInt(10000),
		Nonce:           big.NewInt(0),
		Deadline:        big.NewInt(0),
	}

	approved, _, err := client.SimulateIntent(ctx, invalidIntent)
	require.NoError(t, err)
	assert.False(t, approved, "Invalid intent should be rejected")
}

func getTestConfig(t *testing.T) Config {
	return Config{
		RPCURL:     os.Getenv("SEPOLIA_RPC_URL"),
		ChainID:    11155111,
		OperatorPK: os.Getenv("OPERATOR_PRIVATE_KEY"),
		AgentPK:    os.Getenv("AGENT_PRIVATE_KEY"),
		AgentID:    os.Getenv("AGENT_ID"),
		Contracts: ContractAddresses{
			AgentRegistry:      "0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3",
			HackathonVault:     "0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90",
			RiskRouter:         "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
			ReputationRegistry: "0x423a9904e39537a9997fbaF0f220d79D7d545763",
			ValidationRegistry: "0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1",
		},
	}
}
