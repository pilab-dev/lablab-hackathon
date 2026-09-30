package blockchain

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testIntent(wallet common.Address) TradeIntent {
	return TradeIntent{
		AgentId:         big.NewInt(42),
		AgentWallet:     wallet,
		Pair:            "BTCUSD",
		Action:          "buy",
		AmountUsdScaled: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18)),
		MaxSlippageBps:  big.NewInt(100),
		Nonce:           big.NewInt(0),
		Deadline:        big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
	}
}

func TestSignTradeIntent(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	sig, err := SignTradeIntent(key, testIntent(crypto.PubkeyToAddress(key.PublicKey)))
	require.NoError(t, err)
	assert.Len(t, sig, 65, "signature should be 65 bytes")
	assert.Contains(t, []byte{27, 28}, sig[64], "v should be 27 or 28")
}

func TestSignTradeIntent_NilKey(t *testing.T) {
	_, err := SignTradeIntent(nil, testIntent(common.Address{}))
	assert.Error(t, err)
}

func TestRecoverSigner(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	intent := testIntent(crypto.PubkeyToAddress(key.PublicKey))

	hash, err := HashTradeIntent(intent)
	require.NoError(t, err)

	sig, err := SignTradeIntent(key, intent)
	require.NoError(t, err)

	recovered, err := RecoverSigner(hash, sig)
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), recovered)
}

func TestRecoverSigner_AcceptsZeroOneRecoveryID(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	intent := testIntent(crypto.PubkeyToAddress(key.PublicKey))

	hash, err := HashTradeIntent(intent)
	require.NoError(t, err)
	sig, err := SignTradeIntent(key, intent)
	require.NoError(t, err)

	sig[64] -= 27
	recovered, err := RecoverSigner(hash, sig)
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), recovered)
}

func TestRecoverSigner_InvalidSignature(t *testing.T) {
	var hash [32]byte

	_, err := RecoverSigner(hash, make([]byte, 64))
	assert.Error(t, err, "short signature should be rejected")

	sig := make([]byte, 65)
	sig[64] = 29
	_, err = RecoverSigner(hash, sig)
	assert.Error(t, err, "recovery id 29 should be rejected")
}

func TestRecoverSigner_TamperedIntent(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	intent := testIntent(crypto.PubkeyToAddress(key.PublicKey))

	sig, err := SignTradeIntent(key, intent)
	require.NoError(t, err)

	intent.AmountUsdScaled = new(big.Int).Mul(intent.AmountUsdScaled, big.NewInt(10))
	hash, err := HashTradeIntent(intent)
	require.NoError(t, err)

	recovered, err := RecoverSigner(hash, sig)
	require.NoError(t, err)
	assert.NotEqual(t, crypto.PubkeyToAddress(key.PublicKey), recovered)
}

func TestHashTradeIntent_InvalidFields(t *testing.T) {
	intent := testIntent(common.Address{})
	intent.Nonce = nil
	_, err := HashTradeIntent(intent)
	assert.Error(t, err, "nil uint256 field should be rejected")

	intent = testIntent(common.Address{})
	intent.MaxSlippageBps = big.NewInt(-1)
	_, err = HashTradeIntent(intent)
	assert.Error(t, err, "negative uint256 field should be rejected")
}

// TestHashTradeIntent_MatchesRiskRouterDigest rebuilds the digest exactly as
// RiskRouter.submitTradeIntent does (_hashTypedDataV4 over TRADE_INTENT_TYPEHASH)
// without going through apitypes.
func TestHashTradeIntent_MatchesRiskRouterDigest(t *testing.T) {
	// RiskRouter.TRADE_INTENT_TYPEHASH() on Sepolia.
	typeHash := common.HexToHash("0x907ba001521509472ca63148609f92d473fdfc999a6dae92c9ab7f1e6db0833c")
	require.Equal(t, typeHash, crypto.Keccak256Hash([]byte(
		"TradeIntent(uint256 agentId,address agentWallet,string pair,string action,"+
			"uint256 amountUsdScaled,uint256 maxSlippageBps,uint256 nonce,uint256 deadline)")))

	word := func(v *big.Int) []byte { return common.LeftPadBytes(v.Bytes(), 32) }

	domainSeparator := crypto.Keccak256(
		crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")),
		crypto.Keccak256([]byte("RiskRouter")),
		crypto.Keccak256([]byte("1")),
		word(big.NewInt(11155111)),
		common.LeftPadBytes(common.HexToAddress("0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC").Bytes(), 32),
	)

	intent := testIntent(common.HexToAddress("0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"))
	intent.Nonce = big.NewInt(7)

	structHash := crypto.Keccak256(
		typeHash.Bytes(),
		word(intent.AgentId),
		common.LeftPadBytes(intent.AgentWallet.Bytes(), 32),
		crypto.Keccak256([]byte(intent.Pair)),
		crypto.Keccak256([]byte(intent.Action)),
		word(intent.AmountUsdScaled),
		word(intent.MaxSlippageBps),
		word(intent.Nonce),
		word(intent.Deadline),
	)

	want := crypto.Keccak256([]byte{0x19, 0x01}, domainSeparator, structHash)

	got, err := HashTradeIntent(intent)
	require.NoError(t, err)
	assert.Equal(t, want, got[:])
}

// TestSignTradeIntent_KnownVector checks the signature against one produced
// independently by ethers v6 Wallet.signTypedData for the same domain, types,
// message and key.
func TestSignTradeIntent_KnownVector(t *testing.T) {
	key, err := crypto.HexToECDSA("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)

	intent := TradeIntent{
		AgentId:         big.NewInt(42),
		AgentWallet:     common.HexToAddress("0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"),
		Pair:            "BTCUSD",
		Action:          "buy",
		AmountUsdScaled: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18)),
		MaxSlippageBps:  big.NewInt(100),
		Nonce:           big.NewInt(7),
		Deadline:        big.NewInt(1767225600),
	}
	require.Equal(t, intent.AgentWallet, crypto.PubkeyToAddress(key.PublicKey))

	hash, err := HashTradeIntent(intent)
	require.NoError(t, err)
	assert.Equal(t, "0x6802934c5cb9a25f51f0bc70a04561d311b2b2bdad3a169ea30427a93a56d315", hexutil.Encode(hash[:]))

	sig, err := SignTradeIntent(key, intent)
	require.NoError(t, err)
	assert.Equal(t,
		"0x359bcd9090f59efd1ce2f4d0082382610420b39e926278d25a7a10d0871456582a42c4cebde00323bc2e8dccf4ecc2d533f183a733f35fe0dc566c241130c0131c",
		hexutil.Encode(sig))
}

func TestHashCheckpoint(t *testing.T) {
	agentId := big.NewInt(42)
	amount := big.NewInt(1e18)
	price := new(big.Int).Mul(big.NewInt(50000), big.NewInt(1e18))

	hash1 := HashCheckpoint(agentId, 1234567890, "buy", "BTCUSD", amount, price, common.Hash{})
	hash2 := HashCheckpoint(agentId, 1234567890, "buy", "BTCUSD", amount, price, common.Hash{})
	assert.Equal(t, hash1, hash2, "same inputs should produce same hash")

	assert.NotEqual(t, hash1, HashCheckpoint(agentId, 1234567890, "sell", "BTCUSD", amount, price, common.Hash{}))
	assert.NotEqual(t, hash1, HashCheckpoint(agentId, 1234567891, "buy", "BTCUSD", amount, price, common.Hash{}))
	assert.NotEqual(t, hash1, HashCheckpoint(agentId, 1234567890, "buy", "BTCUSD", amount, price, common.HexToHash("0x01")))
}

func TestHashCheckpoint_NilAmountsEncodeAsZero(t *testing.T) {
	zero := new(big.Int)
	assert.Equal(t,
		HashCheckpoint(zero, 1, "buy", "BTCUSD", zero, zero, common.Hash{}),
		HashCheckpoint(nil, 1, "buy", "BTCUSD", nil, nil, common.Hash{}),
	)
}

func TestHashCheckpoint_NegativeAmountPanics(t *testing.T) {
	assert.Panics(t, func() {
		HashCheckpoint(big.NewInt(1), 1, "buy", "BTCUSD", big.NewInt(-1), big.NewInt(1), common.Hash{})
	})
}
