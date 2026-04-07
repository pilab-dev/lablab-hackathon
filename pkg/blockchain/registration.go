package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type AgentInfo struct {
	ID             *big.Int
	Owner          common.Address
	OperatorWallet common.Address
	AgentWallet    common.Address
	Name           string
	Description    string
	RegisteredAt   *big.Int
}

func (c *Client) IsRegistered(ctx context.Context, agentId *big.Int) (bool, error) {
	if agentId == nil {
		return false, fmt.Errorf("agentId cannot be nil")
	}
	logger := log.With().Str("module", "blockchain").Str("method", "IsRegistered").Logger()
	logger.Debug().Str("agent_id", agentId.String()).Msg("Checking agent registration status")

	isRegistered, err := c.agentRegistry.IsRegistered(&bind.CallOpts{Context: ctx}, agentId)
	if err != nil {
		return false, fmt.Errorf("failed to check agent registration: %w", err)
	}

	logger.Debug().Str("agent_id", agentId.String()).Bool("registered", isRegistered).Msg("Agent registration status retrieved")
	return isRegistered, nil
}

func (c *Client) RegisterAgent(ctx context.Context, name, description string, capabilities []string, metadataURI string, agentWallet, operatorWallet common.Address) (*big.Int, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "RegisterAgent").Logger()
	logger.Info().Str("agent_wallet", agentWallet.Hex()).Str("operator_wallet", operatorWallet.Hex()).Msg("Registering new agent")

	tx, err := c.agentRegistry.RegisterAgent(c.auth, name, description, capabilities, metadataURI, agentWallet, operatorWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to submit register agent transaction: %w", err)
	}

	logger.Info().Str("tx_hash", tx.Hash().Hex()).Msg("Agent registration transaction submitted, waiting for confirmation")

	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for register agent transaction: %w", err)
	}

	if receipt.Status != 1 {
		return nil, fmt.Errorf("agent registration transaction failed (status: %d)", receipt.Status)
	}

	// Parse AgentRegistered event from receipt
	for _, logEntry := range receipt.Logs {
		event, err := c.agentRegistry.ParseAgentRegistered(*logEntry)
		if err == nil {
			logger.Info().Str("agent_id", event.AgentId.String()).Str("tx_hash", tx.Hash().Hex()).Msg("Agent registered successfully")
			return event.AgentId, nil
		}
	}

	return nil, fmt.Errorf("AgentRegistered event not found in transaction receipt")
}

func (c *Client) GetAgent(ctx context.Context, agentId *big.Int) (*AgentInfo, error) {
	if agentId == nil {
		return nil, fmt.Errorf("agentId cannot be nil")
	}
	logger := log.With().Str("module", "blockchain").Str("method", "GetAgent").Logger()
	logger.Debug().Str("agent_id", agentId.String()).Msg("Fetching agent info")

	owner, operatorWallet, agentWallet, name, description, registeredAt, err := c.agentRegistry.GetAgent(&bind.CallOpts{Context: ctx}, agentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}

	agent := &AgentInfo{
		ID:             agentId,
		Owner:          owner,
		OperatorWallet: operatorWallet,
		AgentWallet:    agentWallet,
		Name:           name,
		Description:    description,
		RegisteredAt:   registeredAt,
	}

	logger.Debug().Str("agent_id", agentId.String()).Str("name", name).Msg("Agent info retrieved successfully")
	return agent, nil
}
