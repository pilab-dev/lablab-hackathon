package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"kraken-trader/pkg/blockchain/bindings/agentregistry"
	"kraken-trader/pkg/blockchain/bindings/hackathonvault"
	"kraken-trader/pkg/blockchain/bindings/reputationregistry"
	"kraken-trader/pkg/blockchain/bindings/riskrouter"
	"kraken-trader/pkg/blockchain/bindings/validationregistry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	rpc       *ethclient.Client
	chainID   *big.Int
	auth      *bind.TransactOpts
	agentAuth *bind.TransactOpts

	operatorAddr common.Address
	agentAddr    common.Address
	agentId      *big.Int

	agentRegistry      *agentregistry.Agentregistry
	hackathonVault     *hackathonvault.Hackathonvault
	riskRouter         *riskrouter.Riskrouter
	validationRegistry *validationregistry.Validationregistry
	reputationRegistry *reputationregistry.Reputationregistry
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.RPCURL == "" {
		return nil, fmt.Errorf("SEPOLIA_RPC_URL is required")
	}

	if cfg.OperatorPK == "" {
		return nil, fmt.Errorf("OPERATOR_PRIVATE_KEY is required")
	}

	if cfg.AgentPK == "" {
		return nil, fmt.Errorf("AGENT_PRIVATE_KEY is required")
	}

	operatorKey, err := crypto.HexToECDSA(cfg.OperatorPK)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operator private key: %w", err)
	}

	agentKey, err := crypto.HexToECDSA(cfg.AgentPK)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent private key: %w", err)
	}

	if cfg.AgentID != "" {
		_, ok := new(big.Int).SetString(cfg.AgentID, 10)
		if !ok {
			return nil, fmt.Errorf("failed to parse AGENT_ID: %s", cfg.AgentID)
		}
	}

	rpc, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Sepolia RPC: %w", err)
	}

	rpcChainID, err := rpc.ChainID(ctx)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to get chain ID from RPC: %w", err)
	}

	if cfg.ChainID > 0 {
		expectedChainID := new(big.Int).SetUint64(cfg.ChainID)
		if rpcChainID.Cmp(expectedChainID) != 0 {
			rpc.Close()
			return nil, fmt.Errorf("RPC chain ID %s does not match expected chain ID %d", rpcChainID.String(), cfg.ChainID)
		}
	}

	operatorAuth, err := bind.NewKeyedTransactorWithChainID(operatorKey, rpcChainID)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create operator transactor: %w", err)
	}

	agentAuth, err := bind.NewKeyedTransactorWithChainID(agentKey, rpcChainID)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create agent transactor: %w", err)
	}

	if cfg.GasLimit > 0 {
		operatorAuth.GasLimit = cfg.GasLimit
		agentAuth.GasLimit = cfg.GasLimit
	}

	if cfg.GasPriceGwei > 0 {
		gasPrice := new(big.Int).Mul(big.NewInt(int64(cfg.GasPriceGwei)), big.NewInt(1e9))
		operatorAuth.GasPrice = gasPrice
		agentAuth.GasPrice = gasPrice
	}

	operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)
	agentAddr := crypto.PubkeyToAddress(agentKey.PublicKey)

	var agentId *big.Int
	if cfg.AgentID != "" {
		var ok bool
		agentId, ok = new(big.Int).SetString(cfg.AgentID, 10)
		if !ok {
			rpc.Close()
			return nil, fmt.Errorf("failed to parse AGENT_ID: %s", cfg.AgentID)
		}
	}

	agentRegistryAddr := common.HexToAddress(cfg.Contracts.AgentRegistry)
	hackathonVaultAddr := common.HexToAddress(cfg.Contracts.HackathonVault)
	riskRouterAddr := common.HexToAddress(cfg.Contracts.RiskRouter)
	validationRegistryAddr := common.HexToAddress(cfg.Contracts.ValidationRegistry)
	reputationRegistryAddr := common.HexToAddress(cfg.Contracts.ReputationRegistry)

	agentRegistry, err := agentregistry.NewAgentregistry(agentRegistryAddr, rpc)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create AgentRegistry binding: %w", err)
	}

	hackathonVault, err := hackathonvault.NewHackathonvault(hackathonVaultAddr, rpc)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create HackathonVault binding: %w", err)
	}

	riskRouter, err := riskrouter.NewRiskrouter(riskRouterAddr, rpc)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create RiskRouter binding: %w", err)
	}

	validationRegistry, err := validationregistry.NewValidationregistry(validationRegistryAddr, rpc)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create ValidationRegistry binding: %w", err)
	}

	reputationRegistry, err := reputationregistry.NewReputationregistry(reputationRegistryAddr, rpc)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("failed to create ReputationRegistry binding: %w", err)
	}

	return &Client{
		rpc:                rpc,
		chainID:            rpcChainID,
		auth:               operatorAuth,
		agentAuth:          agentAuth,
		operatorAddr:       operatorAddr,
		agentAddr:          agentAddr,
		agentId:            agentId,
		agentRegistry:      agentRegistry,
		hackathonVault:     hackathonVault,
		riskRouter:         riskRouter,
		validationRegistry: validationRegistry,
		reputationRegistry: reputationRegistry,
	}, nil
}

func (c *Client) Close() {
	if c.rpc != nil {
		c.rpc.Close()
	}
}

func (c *Client) OperatorAddress() common.Address {
	return c.operatorAddr
}

func (c *Client) AgentAddress() common.Address {
	return c.agentAddr
}

func (c *Client) AgentID() *big.Int {
	return c.agentId
}

func (c *Client) RPC() *ethclient.Client {
	return c.rpc
}

func (c *Client) ChainID() *big.Int {
	return c.chainID
}

func (c *Client) OperatorAuth() *bind.TransactOpts {
	return c.auth
}

func (c *Client) AgentAuth() *bind.TransactOpts {
	return c.agentAuth
}

func (c *Client) AgentRegistry() *agentregistry.Agentregistry {
	return c.agentRegistry
}

func (c *Client) HackathonVault() *hackathonvault.Hackathonvault {
	return c.hackathonVault
}

func (c *Client) RiskRouter() *riskrouter.Riskrouter {
	return c.riskRouter
}

func (c *Client) ValidationRegistry() *validationregistry.Validationregistry {
	return c.validationRegistry
}

func (c *Client) ReputationRegistry() *reputationregistry.Reputationregistry {
	return c.reputationRegistry
}
