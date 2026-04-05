# ERC-8004 Integration Plan

## Overview

Integrate the ERC-8004 shared hackathon contracts to enable on-chain trade submission, validation, and reputation tracking. Currently, the decision engine produces trade signals but never executes them. This plan adds a blockchain execution layer that submits signed TradeIntents to the RiskRouter on Sepolia testnet.

## Shared Contracts (Sepolia, Chain ID: 11155111)

| Contract | Address |
|----------|---------|
| AgentRegistry | `0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3` |
| HackathonVault | `0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90` |
| RiskRouter | `0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC` |
| ReputationRegistry | `0x423a9904e39537a9997fbaF0f220d79D7d545763` |
| ValidationRegistry | `0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1` |

## Full Flow

```
Market Data (Kraken WS) → Decision Engine (Ollama LLM) → TradeDecision
  → TradeExecutor (new)
    → 1. Sign TradeIntent with EIP-712 (agentWallet)
    → 2. Submit to RiskRouter.submitTradeIntent()
    → 3. Wait for TradeApproved/TradeRejected event
    → 4. Post checkpoint to ValidationRegistry.postEIP712Attestation()
    → 5. Record trade in InfluxDB + SQLite
```

---

## Phase 1: Dependencies & Configuration

### 1.1 Add go-ethereum dependency

```bash
go get github.com/ethereum/go-ethereum@latest
```

This provides:
- `ethclient` — JSON-RPC client for Sepolia
- `accounts/abi/bind` — Generated contract bindings
- `common` — Address, Hash types
- `crypto` — ECDSA signing, keccak256
- `core/types` — Transactions

### 1.2 Update `.env.example` with blockchain config

```env
# ERC-8004 Blockchain Configuration
SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
SEPOLIA_CHAIN_ID=11155111

# Operator wallet (owns ERC-721, pays gas for registration/claim)
OPERATOR_PRIVATE_KEY=

# Agent wallet (hot wallet, signs TradeIntents)
AGENT_PRIVATE_KEY=

# Agent ID (obtained after registration, saved to DB)
AGENT_ID=

# Contract addresses (pre-filled with shared hackathon addresses)
AGENT_REGISTRY_ADDRESS=0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3
HACKATHON_VAULT_ADDRESS=0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90
RISK_ROUTER_ADDRESS=0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC
REPUTATION_REGISTRY_ADDRESS=0x423a9904e39537a9997fbaF0f220d79D7d545763
VALIDATION_REGISTRY_ADDRESS=0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1

# Gas configuration
GAS_LIMIT=300000
GAS_PRICE_GWEI=2
```

### 1.3 Update `pkg/config/` to load blockchain settings

Add fields to the config struct for all new env vars.

---

## Phase 2: Contract Bindings

### 2.1 Generate Go bindings from ABIs

Use `abigen` to generate type-safe Go bindings for each contract:

```bash
# Install abigen
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# Generate bindings for each contract
abigen --abi=contracts/AgentRegistry.abi --pkg=agentregistry --out=pkg/blockchain/agentregistry/agentregistry.go
abigen --abi=contracts/HackathonVault.abi --pkg=hackathonvault --out=pkg/blockchain/hackathonvault/hackathonvault.go
abigen --abi=contracts/RiskRouter.abi --pkg=riskrouter --out=pkg/blockchain/riskrouter/riskrouter.go
abigen --abi=contracts/ValidationRegistry.abi --pkg=validationregistry --out=pkg/blockchain/validationregistry/validationregistry.go
abigen --abi=contracts/ReputationRegistry.abi --pkg=reputationregistry --out=pkg/blockchain/reputationregistry/reputationregistry.go
```

ABIs will be extracted from the verified Etherscan source code we've already reviewed.

### 2.2 Alternatively: Manual bindings

If abigen is problematic, create lightweight wrapper structs with manual ABI encoding. Given the contracts are simple, this may be cleaner:

```go
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
```

---

## Phase 3: Blockchain Client Package (`pkg/blockchain/`)

### 3.1 `pkg/blockchain/client.go` — Core client

```go
type Client struct {
    rpc          *ethclient.Client
    chainID      *big.Int
    operatorKey  *ecdsa.PrivateKey
    agentKey     *ecdsa.PrivateKey
    operatorAddr common.Address
    agentAddr    common.Address
    agentId      *big.Int

    // Contract instances
    agentRegistry      *AgentRegistry
    hackathonVault     *HackathonVault
    riskRouter         *RiskRouter
    validationRegistry *ValidationRegistry
    reputationRegistry *ReputationRegistry
}
```

Key methods:
- `NewClient(cfg Config) (*Client, error)` — Connect to Sepolia, load wallets
- `Close()` — Close RPC connection
- `OperatorAddress() common.Address`
- `AgentAddress() common.Address`
- `AgentID() *big.Int`

### 3.2 `pkg/blockchain/registration.go` — Agent lifecycle

```go
func (c *Client) IsRegistered(ctx context.Context) (bool, error)
func (c *Client) Register(ctx context.Context, name, description string, capabilities []string, uri string) (agentId *big.Int, err error)
func (c *Client) HasClaimed(ctx context.Context) (bool, error)
func (c *Client) ClaimAllocation(ctx context.Context) error
func (c *Client) GetVaultBalance(ctx context.Context) (*big.Int, error)
func (c *Client) GetIntentNonce(ctx context.Context) (*big.Int, error)
func (c *Client) GetReputationScore(ctx context.Context) (*big.Int, error)
```

### 3.3 `pkg/blockchain/trade.go` — Trade submission

```go
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

type TradeResult struct {
    Approved    bool
    IntentHash  common.Hash
    TxHash      common.Hash
    Reason      string
}

func (c *Client) SignTradeIntent(intent TradeIntent) ([]byte, error)
func (c *Client) SubmitTradeIntent(ctx context.Context, intent TradeIntent) (*TradeResult, error)
func (c *Client) SimulateIntent(ctx context.Context, intent TradeIntent) (bool, string, error)
```

EIP-712 domain for RiskRouter:
```go
var RiskRouterDomain = apitypes.TypedDataDomain{
    Name:              "RiskRouter",
    Version:           "1",
    ChainId:           math.NewHexOrDecimal256(11155111),
    VerifyingContract: "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
}
```

EIP-712 types:
```json
{
  "TradeIntent": [
    {"name": "agentId", "type": "uint256"},
    {"name": "agentWallet", "type": "address"},
    {"name": "pair", "type": "string"},
    {"name": "action", "type": "string"},
    {"name": "amountUsdScaled", "type": "uint256"},
    {"name": "maxSlippageBps", "type": "uint256"},
    {"name": "nonce", "type": "uint256"},
    {"name": "deadline", "type": "uint256"}
  ]
}
```

### 3.4 `pkg/blockchain/validation.go` — Checkpoint posting

```go
type Checkpoint struct {
    AgentId        *big.Int
    Timestamp      uint64
    Action         string
    Pair           string
    AmountUsdScaled *big.Int
    PriceUsdScaled *big.Int
    ReasoningHash  common.Hash
}

func (c *Client) PostCheckpoint(ctx context.Context, agentId *big.Int, checkpointHash common.Hash, score uint8, notes string) (common.Hash, error)
func (c *Client) GetValidationScore(ctx context.Context, agentId *big.Int) (*big.Int, error)
```

### 3.5 `pkg/blockchain/eip712.go` — Signing utilities

```go
func SignTradeIntent(key *ecdsa.PrivateKey, domain apitypes.TypedDataDomain, intent TradeIntent) ([]byte, error)
func HashCheckpoint(checkpoint Checkpoint) common.Hash
```

---

## Phase 4: Trade Executor (`internal/executor/`)

### 4.1 `internal/executor/executor.go` — Trade execution bridge

This is the critical missing piece. It connects the decision engine to blockchain execution.

```go
type TradeExecutor struct {
    blockchain   *blockchain.Client
    kraken       *kraken.Client
    state        *state.MemoryManager
    storage      *storage.Client
    repo         *repository.Repository
    config       Config
}

type Config struct {
    MaxSlippageBps  uint64
    IntentDeadline  time.Duration
    MinConfidence   float64
}

func (e *TradeExecutor) Execute(ctx context.Context, decision *decision.TradeDecision) error
func (e *TradeExecutor) PostCheckpoint(ctx context.Context, decision *decision.TradeDecision, price float64) error
```

**Execution flow in `Execute()`:**
1. Validate decision (confidence threshold, cooldown)
2. Convert decision sizing to USD amount (using portfolio value from MemoryManager)
3. Fetch current nonce from RiskRouter
4. Build TradeIntent struct
5. Sign with agent wallet (EIP-712)
6. Call `riskRouter.submitTradeIntent(intent, signature)`
7. Wait for transaction receipt
8. Parse `TradeApproved` / `TradeRejected` event
9. If approved: execute actual trade via Kraken CLI (or simulate in paper mode)
10. Record trade in InfluxDB + SQLite
11. Post checkpoint to ValidationRegistry
12. Return result

### 4.2 `internal/executor/interfaces.go`

```go
type DecisionProvider interface {
    Decide(ctx context.Context, pairs []string) ([]*TradeDecision, error)
}

type MarketDataProvider interface {
    GetTicker(ctx context.Context, pair string) (*Ticker, error)
}
```

---

## Phase 5: Integration with Existing Flow

### 5.1 Modify `cmd/trader/run.go` — Wire up blockchain client

In `runRun()` (line 44), after initializing existing components:

```go
// Initialize blockchain client
bcCfg := blockchain.Config{
    RPCURL:     viper.GetString("SEPOLIA_RPC_URL"),
    ChainID:    viper.GetUint64("SEPOLIA_CHAIN_ID"),
    OperatorPK: viper.GetString("OPERATOR_PRIVATE_KEY"),
    AgentPK:    viper.GetString("AGENT_PRIVATE_KEY"),
    AgentID:    viper.GetString("AGENT_ID"),
    Contracts:  blockchain.ContractAddresses{...},
}
bcClient, err := blockchain.NewClient(ctx, bcCfg)

// Initialize trade executor
executor := executor.NewTradeExecutor(bcClient, krakenClient, memMgr, influxClient, repo, executorCfg)
```

### 5.2 Modify `runDecisionLoop()` — Add execution after decision

Current flow (run.go:500-537):
```
Price Alert → Decide() → Log → Publish to NATS → [END]
```

New flow:
```
Price Alert → Decide() → executor.Execute() → Log result → Publish to NATS → PostCheckpoint()
```

```go
func runDecisionLoop(...) {
    for pair := range memMgr.PriceAlertCh {
        decisions, err := engine.Decide(ctx, pair)
        if err != nil { ... }

        for _, d := range decisions {
            if d.Action != "hold" && d.Confidence >= cfg.ConfidenceThreshold {
                result, err := tradeExecutor.Execute(ctx, d)
                if err != nil {
                    log.Error().Err(err).Msg("Trade execution failed")
                } else if result.Approved {
                    // Post checkpoint after successful trade
                    price := getCurrentPrice(pair)
                    tradeExecutor.PostCheckpoint(ctx, d, price)
                }
            }
        }
    }
}
```

### 5.3 Modify `internal/pipeline/pipeline.go` — Add blockchain execution

In `ProcessTick()` (line 55), after `assessment.Approved`:

```go
if assessment.Approved {
    decision := &decision.TradeDecision{
        Pair:       pair,
        Action:     string(result.Signal),
        SizePct:    assessment.MaxSizePct,
        Confidence: assessment.AdjConfidence,
    }
    tradeResult, err := p.executor.Execute(ctx, decision)
    if err != nil {
        log.Error().Err(err).Msg("Pipeline trade execution failed")
    }
    p.result.Action = "EXECUTED"
}
```

---

## Phase 6: State Persistence

### 6.1 SQLite — Store agent identity

Add to `internal/repository/`:

```go
type AgentIdentity struct {
    ID              uint `gorm:"primaryKey"`
    AgentID         string
    OperatorAddress string
    AgentAddress    string
    Name            string
    RegisteredAt    time.Time
    HasClaimed      bool
}
```

Migration: Add `agent_identities` table. On startup, check if agent is registered; if not, register and save.

### 6.2 SQLite — Store trade records with on-chain data

Extend existing trade model:

```go
type TradeRecord struct {
    // Existing fields...
    Pair       string
    Action     string
    SizePct    float64
    // New blockchain fields...
    TxHash         string
    IntentHash     string
    OnChainApproved bool
    CheckpointHash  string
    GasUsed        uint64
}
```

### 6.3 InfluxDB — Store blockchain metrics

Add new measurement `blockchain_trades`:
- `tx_hash`, `intent_hash`, `agent_id`, `pair`, `action`, `amount_usd_scaled`, `gas_used`, `approved`, `checkpoint_hash`

---

## Phase 7: API Endpoints

Add to `api.yaml` and regenerate:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/blockchain/status` | GET | Agent registration status, vault balance, reputation score |
| `/api/blockchain/register` | POST | Register agent on-chain |
| `/api/blockchain/claim` | POST | Claim vault allocation |
| `/api/blockchain/trades` | GET | List on-chain trade history |
| `/api/blockchain/nonce` | GET | Current intent nonce |
| `/api/blockchain/simulate` | POST | Simulate a trade intent (dry-run) |

---

## Phase 8: CLI Commands

Add Cobra subcommands for agent lifecycle:

```bash
# Register agent
go run cmd/trader/main.go agent register --name "KrakenTrader" --description "AI trading bot"

# Claim allocation
go run cmd/trader/main.go agent claim

# Check status
go run cmd/trader/main.go agent status

# Check reputation
go run cmd/trader/main.go agent reputation
```

---

## Phase 9: Paper Mode vs Live Mode

### Paper Mode (default)
- Decisions are made normally
- TradeIntent is signed and submitted to RiskRouter (on-chain record)
- Actual Kraken order is NOT placed
- Trade is recorded as "simulated" in InfluxDB

### Live Mode (`TRADING_MODE=live`)
- Full flow: Decision → RiskRouter approval → Kraken order execution → Checkpoint
- Requires `KRAKEN_API_KEY` and `KRAKEN_API_SECRET`

---

## Phase 10: Error Handling & Resilience

### Key failure modes:
1. **RPC connection lost** — Retry with exponential backoff, fallback to paper mode
2. **Transaction reverted** — Log reason, skip trade, continue
3. **Insufficient gas** — Alert, skip trade
4. **Nonce mismatch** — Re-fetch nonce, retry once
5. **RiskRouter rejection** — Log reason, do NOT execute on Kraken

### Cooldown on failures:
- After 3 consecutive failures, pause blockchain execution for 5 minutes
- Continue market data collection and decision making (log only)

---

## File Structure (New/Modified)

```
pkg/blockchain/
├── client.go              # Core client, connection management
├── config.go              # Blockchain configuration
├── agent_registry.go      # AgentRegistry interactions
├── hackathon_vault.go     # HackathonVault interactions
├── risk_router.go         # RiskRouter interactions + EIP-712 signing
├── validation_registry.go # ValidationRegistry interactions
├── reputation_registry.go # ReputationRegistry interactions
├── eip712.go              # EIP-712 domain + signing utilities
├── types.go               # Shared types (TradeIntent, Checkpoint, etc.)
└── client_test.go         # Tests

internal/executor/
├── executor.go            # Trade execution bridge
├── interfaces.go          # Provider interfaces
└── executor_test.go       # Tests

cmd/trader/
├── agent.go               # New: agent lifecycle CLI commands
└── run.go                 # Modified: wire up blockchain client + executor

internal/api/
├── api.yaml               # Modified: add blockchain endpoints
├── implementation.go      # Modified: add blockchain handlers
└── generated.go           # Auto-regenerated

internal/repository/
├── agent_identity.go      # New: agent identity model
└── trade_record.go        # Modified: add blockchain fields

.env.example               # Modified: add blockchain config
pkg/config/                # Modified: load blockchain settings
```

---

## Implementation Order

1. **Phase 1** — Dependencies + config (30 min)
2. **Phase 2** — Contract bindings (1 hour)
3. **Phase 3** — Blockchain client package (3 hours)
4. **Phase 4** — Trade executor (2 hours)
5. **Phase 5** — Integration with decision loop (1 hour)
6. **Phase 6** — State persistence (1 hour)
7. **Phase 7** — API endpoints (1 hour)
8. **Phase 8** — CLI commands (1 hour)
9. **Phase 9** — Paper/live mode toggle (30 min)
10. **Phase 10** — Error handling (1 hour)

**Estimated total: ~12 hours**

---

## Judging Criteria Alignment

| Criteria | How we score |
|----------|-------------|
| Risk-adjusted PnL | TradeExecutor records PnL per trade, tracks drawdown |
| Drawdown control | RiskRouter enforces 5% max, we track in Redis StateTracker |
| Validation quality | Every trade posts checkpoint to ValidationRegistry with reasoning |
| Reputation score | Accumulates via ReputationRegistry from validator feedback |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Sepolia RPC rate limits | Use public node with retry logic, consider Infura/Alchemy fallback |
| Gas price spikes | Configurable gas price, skip trades if gas > threshold |
| EIP-712 signature mismatch | Extensive unit tests against contract typehash |
| Agent not registered | Auto-register on startup if not already registered |
| Vault underfunded | Check balance before claiming, alert if insufficient |
| Transaction timeout | 60s timeout per tx, retry with higher gas |
