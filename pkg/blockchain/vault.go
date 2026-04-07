package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rs/zerolog/log"
)

func (c *Client) HasClaimed(ctx context.Context, agentId *big.Int) (bool, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "HasClaimed").Logger()
	logger.Debug().Str("agent_id", agentId.String()).Msg("Checking vault claim status")

	hasClaimed, err := c.hackathonVault.HasClaimed(&bind.CallOpts{Context: ctx}, agentId)
	if err != nil {
		return false, fmt.Errorf("failed to check vault claim status: %w", err)
	}

	logger.Debug().Str("agent_id", agentId.String()).Bool("has_claimed", hasClaimed).Msg("Vault claim status retrieved")
	return hasClaimed, nil
}

func (c *Client) ClaimAllocation(ctx context.Context, agentId *big.Int) error {
	logger := log.With().Str("module", "blockchain").Str("method", "ClaimAllocation").Logger()
	logger.Info().Str("agent_id", agentId.String()).Msg("Claiming vault allocation")

	tx, err := c.hackathonVault.ClaimAllocation(c.auth, agentId)
	if err != nil {
		return fmt.Errorf("failed to submit claim allocation transaction: %w", err)
	}

	logger.Info().Str("tx_hash", tx.Hash().Hex()).Msg("Claim allocation transaction submitted, waiting for confirmation")

	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for claim allocation transaction: %w", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("claim allocation transaction failed (status: %d)", receipt.Status)
	}

	logger.Info().Str("agent_id", agentId.String()).Str("tx_hash", tx.Hash().Hex()).Msg("Vault allocation claimed successfully")
	return nil
}

func (c *Client) GetVaultBalance(ctx context.Context, agentId *big.Int) (*big.Int, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "GetVaultBalance").Logger()
	logger.Debug().Str("agent_id", agentId.String()).Msg("Fetching agent vault balance")

	balance, err := c.hackathonVault.GetBalance(&bind.CallOpts{Context: ctx}, agentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault balance: %w", err)
	}

	logger.Debug().Str("agent_id", agentId.String()).Str("balance", balance.String()).Msg("Vault balance retrieved")
	return balance, nil
}

func (c *Client) GetTotalVaultBalance(ctx context.Context) (*big.Int, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "GetTotalVaultBalance").Logger()
	logger.Debug().Msg("Fetching total vault balance")

	balance, err := c.hackathonVault.TotalVaultBalance(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to get total vault balance: %w", err)
	}

	logger.Debug().Str("balance", balance.String()).Msg("Total vault balance retrieved")
	return balance, nil
}

func (c *Client) GetUnallocatedBalance(ctx context.Context) (*big.Int, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "GetUnallocatedBalance").Logger()
	logger.Debug().Msg("Fetching unallocated vault balance")

	balance, err := c.hackathonVault.UnallocatedBalance(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to get unallocated vault balance: %w", err)
	}

	logger.Debug().Str("balance", balance.String()).Msg("Unallocated vault balance retrieved")
	return balance, nil
}

func (c *Client) GetAllocationPerTeam(ctx context.Context) (*big.Int, error) {
	logger := log.With().Str("module", "blockchain").Str("method", "GetAllocationPerTeam").Logger()
	logger.Debug().Msg("Fetching allocation per team")

	allocation, err := c.hackathonVault.AllocationPerTeam(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation per team: %w", err)
	}

	logger.Debug().Str("allocation", allocation.String()).Msg("Allocation per team retrieved")
	return allocation, nil
}
