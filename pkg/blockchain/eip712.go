package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// RiskRouterDomain is the EIP-712 domain of the RiskRouter contract on Sepolia.
// It matches the values returned by the contract's eip712Domain() view.
var RiskRouterDomain = apitypes.TypedDataDomain{
	Name:              "RiskRouter",
	Version:           "1",
	ChainId:           math.NewHexOrDecimal256(int64(11155111)),
	VerifyingContract: "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
}

// TradeIntentTypes are the EIP-712 type definitions for a RiskRouter TradeIntent.
var TradeIntentTypes = apitypes.Types{
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	},
	"TradeIntent": {
		{Name: "agentId", Type: "uint256"},
		{Name: "agentWallet", Type: "address"},
		{Name: "pair", Type: "string"},
		{Name: "action", Type: "string"},
		{Name: "amountUsdScaled", Type: "uint256"},
		{Name: "maxSlippageBps", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "deadline", Type: "uint256"},
	},
}

// TradeIntent mirrors the RiskRouter.TradeIntent struct.
type TradeIntent struct {
	AgentId         *big.Int
	AgentWallet     common.Address
	Pair            string
	Action          string
	AmountUsdScaled *big.Int
	MaxSlippageBps  *big.Int
	Nonce           *big.Int
	Deadline        *big.Int
}

func tradeIntentTypedData(intent TradeIntent) apitypes.TypedData {
	return apitypes.TypedData{
		Types:       TradeIntentTypes,
		PrimaryType: "TradeIntent",
		Domain:      RiskRouterDomain,
		Message: apitypes.TypedDataMessage{
			"agentId":         intent.AgentId,
			"agentWallet":     intent.AgentWallet.Hex(),
			"pair":            intent.Pair,
			"action":          intent.Action,
			"amountUsdScaled": intent.AmountUsdScaled,
			"maxSlippageBps":  intent.MaxSlippageBps,
			"nonce":           intent.Nonce,
			"deadline":        intent.Deadline,
		},
	}
}

// HashTradeIntent returns the EIP-712 digest of the intent:
// keccak256("\x19\x01" || domainSeparator || hashStruct(intent)).
//
// This is the value RiskRouter recovers the signer from (_hashTypedDataV4),
// so it is the hash to sign and to pass to RecoverSigner. The bare struct
// hash is not enough: a signature over it recovers to a different address
// on-chain and the intent is rejected with "Invalid signature".
func HashTradeIntent(intent TradeIntent) ([32]byte, error) {
	var out [32]byte
	digest, _, err := apitypes.TypedDataAndHash(tradeIntentTypedData(intent))
	if err != nil {
		return out, fmt.Errorf("failed to hash trade intent: %w", err)
	}
	copy(out[:], digest)
	return out, nil
}

// SignTradeIntent signs the intent's EIP-712 digest with key and returns a
// 65-byte [R || S || V] signature with V in {27, 28}, the form expected by
// OpenZeppelin's ECDSA.recover used in RiskRouter.
func SignTradeIntent(key *ecdsa.PrivateKey, intent TradeIntent) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("signing key is nil")
	}

	hash, err := HashTradeIntent(intent)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(hash[:], key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign trade intent: %w", err)
	}

	sig[crypto.RecoveryIDOffset] += 27
	return sig, nil
}

// RecoverSigner returns the address that produced sig over hash. It accepts
// signatures with V in {27, 28} (as returned by SignTradeIntent) or {0, 1}.
func RecoverSigner(hash [32]byte, sig []byte) (common.Address, error) {
	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("invalid signature length: got %d, want %d", len(sig), crypto.SignatureLength)
	}

	normalized := make([]byte, crypto.SignatureLength)
	copy(normalized, sig)
	if normalized[crypto.RecoveryIDOffset] >= 27 {
		normalized[crypto.RecoveryIDOffset] -= 27
	}
	if normalized[crypto.RecoveryIDOffset] > 1 {
		return common.Address{}, fmt.Errorf("invalid signature recovery id: %d", sig[crypto.RecoveryIDOffset])
	}

	pub, err := crypto.SigToPub(hash[:], normalized)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover signer: %w", err)
	}
	return crypto.PubkeyToAddress(*pub), nil
}

var checkpointArgs = func() abi.Arguments {
	mustType := func(t string) abi.Type {
		typ, err := abi.NewType(t, "", nil)
		if err != nil {
			panic(err)
		}
		return typ
	}
	return abi.Arguments{
		{Type: mustType("uint256")}, // agentId
		{Type: mustType("uint64")},  // timestamp
		{Type: mustType("string")},  // action
		{Type: mustType("string")},  // pair
		{Type: mustType("uint256")}, // amountUsdScaled
		{Type: mustType("uint256")}, // priceUsdScaled
		{Type: mustType("bytes32")}, // reasoningHash
	}
}()

// HashCheckpoint returns keccak256(abi.encode(agentId, timestamp, action, pair,
// amountUsdScaled, priceUsdScaled, reasoningHash)), the checkpoint hash posted
// to the ValidationRegistry. Nil big.Int values are encoded as zero; negative
// values cannot be ABI-encoded as uint256 and cause a panic.
func HashCheckpoint(agentId *big.Int, timestamp uint64, action, pair string, amountUsdScaled, priceUsdScaled *big.Int, reasoningHash common.Hash) common.Hash {
	encoded, err := checkpointArgs.Pack(
		orZero(agentId),
		timestamp,
		action,
		pair,
		orZero(amountUsdScaled),
		orZero(priceUsdScaled),
		[32]byte(reasoningHash),
	)
	if err != nil {
		// Only reachable for negative integers, which cannot represent a
		// valid on-chain checkpoint.
		panic(fmt.Sprintf("failed to encode checkpoint: %v", err))
	}
	return crypto.Keccak256Hash(encoded)
}

func orZero(v *big.Int) *big.Int {
	if v == nil {
		return new(big.Int)
	}
	return v
}
