# ERC-8004 Integration — Task Breakdown

## Shared Context (for all tasks)

**Shared Contracts (Sepolia, Chain ID: 11155111):**

| Contract | Address |
|----------|---------|
| AgentRegistry | `0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3` |
| HackathonVault | `0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90` |
| RiskRouter | `0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC` |
| ReputationRegistry | `0x423a9904e39537a9997fbaF0f220d79D7d545763` |
| ValidationRegistry | `0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1` |

**Key contract ABIs (extracted from Etherscan):**

### AgentRegistry
```
registerAgent(string name, string description, string[] capabilities, string metadataURI, address agentWallet, address operatorWallet) → uint256 agentId
isRegistered(uint256 agentId) → bool
getAgent(uint256 agentId) → (address owner, address operatorWallet, address agentWallet, string name, string description, uint256 registeredAt)
ownerOf(uint256 tokenId) → address
```

### HackathonVault
```
claimAllocation(uint256 agentId)
getBalance(uint256 agentId) → uint256
totalVaultBalance() → uint256
unallocatedBalance() → uint256
hasClaimed(uint256 agentId) → bool
allocationPerTeam() → uint256  // 0.05 ETH = 50000000000000000 wei
```

### RiskRouter
```
submitTradeIntent((uint256 agentId, address agentWallet, string pair, string action, uint256 amountUsdScaled, uint256 maxSlippageBps, uint256 nonce, uint256 deadline) intent, bytes signature)
getIntentNonce(address agentWallet) → uint256
simulateIntent((...) intent) → (bool approved, string reason)
```

### ValidationRegistry
```
postEIP712Attestation(uint256 agentId, bytes32 checkpointHash, uint8 score, string notes)
postAttestation(uint256 agentId, bytes32 checkpointHash, uint8 score, uint8 proofType, bytes proof, string notes)
getAttestations(uint256 agentId) → Attestation[]
getAverageValidationScore(uint256 agentId) → uint256
```

### ReputationRegistry
```
submitFeedback(uint256 agentId, uint8 score, bytes32 outcomeRef, string comment, uint8 feedbackType)
getAverageScore(uint256 agentId) → uint256
getFeedbackHistory(uint256 agentId) → FeedbackEntry[]
hasRated(uint256 agentId, address rater) → bool
```

**EIP-712 Domain for RiskRouter:**
```
Name: "RiskRouter"
Version: "1"
ChainId: 11155111
VerifyingContract: 0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC
```

**EIP-712 TradeIntent type:**
```
TradeIntent([
  {name: "agentId", type: "uint256"},
  {name: "agentWallet", type: "address"},
  {name: "pair", type: "string"},
  {name: "action", type: "string"},
  {name: "amountUsdScaled", type: "uint256"},
  {name: "maxSlippageBps", type: "uint256"},
  {name: "nonce", type: "uint256"},
  {name: "deadline", type: "uint256"}
])
```

**Project conventions (from AGENTS.md):**
- Go 1.26+, zerolog logging, testify for tests
- Clean architecture: interfaces where used, constructor injection
- camelCase for unexported, PascalCase for exported
- Error wrapping with `fmt.Errorf("...: %w", err)`
- Context passed as first arg for I/O operations

---

## Task 1: Add go-ethereum dependency and update configuration

**Files to create/modify:**
- `go.mod` (add dependency)
- `.env.example` (add blockchain config section)
- `pkg/config/config.go` (add blockchain fields)

**What to do:**

1. Run `go get github.com/ethereum/go-ethereum@latest` to add the dependency
2. Add to `.env.example`:
```env
# ERC-8004 Blockchain Configuration
SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
SEPOLIA_CHAIN_ID=11155111
OPERATOR_PRIVATE_KEY=
AGENT_PRIVATE_KEY=
AGENT_ID=
AGENT_REGISTRY_ADDRESS=0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3
HACKATHON_VAULT_ADDRESS=0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90
RISK_ROUTER_ADDRESS=0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC
REPUTATION_REGISTRY_ADDRESS=0x423a9904e39537a9997fbaF0f220d79D7d545763
VALIDATION_REGISTRY_ADDRESS=0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1
GAS_LIMIT=300000
GAS_PRICE_GWEI=2
```
3. In `pkg/config/config.go`, add a `Blockchain` struct and fields to the main config struct. Load all env vars with viper.

**Acceptance criteria:**
- `go.mod` contains `github.com/ethereum/go-ethereum`
- `.env.example` has all blockchain env vars documented
- `pkg/config/config.go` exposes all blockchain settings via the config struct
- `make check` passes (fmt, vet, lint)

---

## Task 2: Extract contract ABIs and generate Go bindings

**Files to create:**
- `pkg/blockchain/abi/AgentRegistry.json`
- `pkg/blockchain/abi/HackathonVault.json`
- `pkg/blockchain/abi/RiskRouter.json`
- `pkg/blockchain/abi/ValidationRegistry.json`
- `pkg/blockchain/abi/ReputationRegistry.json`
- `pkg/blockchain/bindings/agentregistry/agentregistry.go`
- `pkg/blockchain/bindings/hackathonvault/hackathonvault.go`
- `pkg/blockchain/bindings/riskrouter/riskrouter.go`
- `pkg/blockchain/bindings/validationregistry/validationregistry.go`
- `pkg/blockchain/bindings/reputationregistry/reputationregistry.go`

**What to do:**

1. Create `pkg/blockchain/abi/` directory
2. Extract ABIs from Etherscan (already verified contracts). Save each as a JSON file.
3. Install abigen: `go install github.com/ethereum/go-ethereum/cmd/abigen@latest`
4. Generate bindings for each contract:
```bash
abigen --abi=pkg/blockchain/abi/AgentRegistry.json --pkg=agentregistry --out=pkg/blockchain/bindings/agentregistry/agentregistry.go
abigen --abi=pkg/blockchain/abi/HackathonVault.json --pkg=hackathonvault --out=pkg/blockchain/bindings/hackathonvault/hackathonvault.go
abigen --abi=pkg/blockchain/abi/RiskRouter.json --pkg=riskrouter --out=pkg/blockchain/bindings/riskrouter/riskrouter.go
abigen --abi=pkg/blockchain/abi/ValidationRegistry.json --pkg=validationregistry --out=pkg/blockchain/bindings/validationregistry/validationregistry.go
abigen --abi=pkg/blockchain/abi/ReputationRegistry.json --pkg=reputationregistry --out=pkg/blockchain/bindings/reputationregistry/reputationregistry.go
```

**Acceptance criteria:**
- All 5 ABI JSON files exist in `pkg/blockchain/abi/`
- All 5 binding Go files exist in `pkg/blockchain/bindings/`
- `go build ./...` succeeds with no errors
- Bindings compile and expose typed methods for all contract functions

---

## Task 3: Create blockchain client core (`pkg/blockchain/client.go` + `config.go`)

**Files to create:**
- `pkg/blockchain/client.go`
- `pkg/blockchain/config.go`

**What to do:**

Create `pkg/blockchain/config.go`:
```go
type ContractAddresses struct {
    AgentRegistry      string
    HackathonVault     string
    RiskRouter         string
    ReputationRegistry string
    ValidationRegistry string
}

type Config struct {
    RPCURL     string
    ChainID    uint64
    OperatorPK string
    AgentPK    string
    AgentID    string
    Contracts  ContractAddresses
    GasLimit   uint64
    GasPriceGwei uint64
}
```

Create `pkg/blockchain/client.go`:
```go
type Client struct {
    rpc     *ethclient.Client
    chainID *big.Int
    auth    *bind.TransactOpts  // operator wallet transact opts
    agentAuth *bind.TransactOpts // agent wallet transact opts

    operatorAddr common.Address
    agentAddr    common.Address
    agentId      *big.Int

    // Contract bindings
    agentRegistry      *agentregistry.Agentregistry
    hackathonVault     *hackathonvault.Hackathonvault
    riskRouter         *riskrouter.Riskrouter
    validationRegistry *validationregistry.Validationregistry
    reputationRegistry *reputationregistry.Reputationregistry
}

func NewClient(ctx context.Context, cfg Config) (*Client, error)
func (c *Client) Close()
func (c *Client) OperatorAddress() common.Address
func (c *Client) AgentAddress() common.Address
func (c *Client) AgentID() *big.Int
func (c *Client) RPC() *ethclient.Client
```

Constructor should:
1. Connect to Sepolia via `ethclient.Dial(cfg.RPCURL)`
2. Parse operator and agent private keys with `crypto.HexToECDSA()`
3. Create `bind.NewKeyedTransactorWithChainID()` for both wallets
4. Set gas price and limit on transact opts
5. Instantiate all 5 contract bindings with `bind.NewBoundContract()`
6. If `cfg.AgentID` is set, parse and store it
7. Verify connection by calling `client.ChainID(ctx)`

**Acceptance criteria:**
- `NewClient()` connects to Sepolia and initializes all contract bindings
- Both operator and agent wallet addresses are derivable
- `Close()` properly closes the RPC connection
- Constructor returns clear errors for invalid keys, unreachable RPC, wrong chain ID
- Unit tests with mock ethclient (or skip if RPC unavailable)

---

## Task 4: Implement agent registration and vault operations

**Files to create:**
- `pkg/blockchain/registration.go`
- `pkg/blockchain/vault.go`

**What to do — `registration.go`:**

```go
func (c *Client) IsRegistered(ctx context.Context, agentId *big.Int) (bool, error)
func (c *Client) RegisterAgent(ctx context.Context, name, description string, capabilities []string, metadataURI string, agentWallet, operatorWallet common.Address) (*big.Int, error)
func (c *Client) GetAgent(ctx context.Context, agentId *big.Int) (*AgentInfo, error)
```

Where `AgentInfo` is:
```go
type AgentInfo struct {
    ID             *big.Int
    Owner          common.Address
    OperatorWallet common.Address
    AgentWallet    common.Address
    Name           string
    Description    string
    RegisteredAt   *big.Int
}
```

`RegisterAgent()` should:
1. Call `c.agentRegistry.RegisterAgent(c.auth, name, description, capabilities, metadataURI, agentWallet, operatorWallet)`
2. Wait for transaction receipt with `bind.WaitMined()`
3. Parse `AgentRegistered` event to extract the new agentId
4. Return the agentId

**What to do — `vault.go`:**

```go
func (c *Client) HasClaimed(ctx context.Context, agentId *big.Int) (bool, error)
func (c *Client) ClaimAllocation(ctx context.Context, agentId *big.Int) error
func (c *Client) GetVaultBalance(ctx context.Context, agentId *big.Int) (*big.Int, error)
func (c *Client) GetTotalVaultBalance(ctx context.Context) (*big.Int, error)
func (c *Client) GetUnallocatedBalance(ctx context.Context) (*big.Int, error)
func (c *Client) GetAllocationPerTeam(ctx context.Context) (*big.Int, error)
```

`ClaimAllocation()` should:
1. Call `c.hackathonVault.ClaimAllocation(c.auth, agentId)`
2. Wait for receipt
3. Return error if tx reverted

**Acceptance criteria:**
- `IsRegistered()` correctly queries `agentRegistry.isRegistered()`
- `RegisterAgent()` submits tx, waits for receipt, extracts agentId from event
- `ClaimAllocation()` submits tx and waits for confirmation
- All vault read methods return correct values
- Errors include context about which operation failed and why
- Tests for read methods (mock contract calls)

---

## Task 5: Implement EIP-712 signing utilities

**Files to create:**
- `pkg/blockchain/eip712.go`
- `pkg/blockchain/eip712_test.go`

**What to do:**

Create the EIP-712 domain and type definitions matching the RiskRouter contract:

```go
var RiskRouterDomain = apitypes.TypedDataDomain{
    Name:              "RiskRouter",
    Version:           "1",
    ChainId:           math.NewHexOrDecimal256(int64(11155111)),
    VerifyingContract: "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC",
}

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
```

Implement:
```go
func SignTradeIntent(key *ecdsa.PrivateKey, intent TradeIntent) ([]byte, error)
func HashTradeIntent(intent TradeIntent) ([32]byte, error)
func RecoverSigner(hash [32]byte, sig []byte) (common.Address, error)
func HashCheckpoint(agentId *big.Int, timestamp uint64, action, pair string, amountUsdScaled, priceUsdScaled *big.Int, reasoningHash common.Hash) common.Hash
```

`SignTradeIntent()` should:
1. Build `apitypes.TypedData` with domain, types, and message
2. Call `apitypes.TypedData.HashStruct()` to get the hash
3. Sign with `crypto.Sign(hash, key)`
4. Return the 65-byte signature (v, r, s)

`HashCheckpoint()` should:
1. Use `crypto.Keccak256Hash()` on the ABI-encoded checkpoint data
2. Return the 32-byte hash for posting to ValidationRegistry

**Acceptance criteria:**
- `SignTradeIntent()` produces a valid 65-byte ECDSA signature
- `RecoverSigner()` recovers the correct address from hash + signature
- `HashCheckpoint()` produces deterministic hashes for same inputs
- Unit tests verify signature recovery matches the signing key
- Unit tests verify hash determinism
- Signature format matches what RiskRouter expects (v=27/28 or v=0/1)

---

## Task 6: Implement trade submission to RiskRouter

**Files to create:**
- `pkg/blockchain/trade.go`
- `pkg/blockchain/trade_test.go`

**What to do:**

Define the TradeIntent struct:
```go
type TradeIntent struct {
    AgentId         *big.Int
    AgentWallet     common.Address
    Pair            string
    Action          string  // "buy" or "sell"
    AmountUsdScaled *big.Int  // amount * 1e18
    MaxSlippageBps  *big.Int  // basis points, e.g. 100 = 1%
    Nonce           *big.Int
    Deadline        *big.Int  // unix timestamp
}

type TradeResult struct {
    Approved   bool
    IntentHash common.Hash
    TxHash     common.Hash
    Reason     string
    GasUsed    uint64
}
```

Implement:
```go
func (c *Client) GetIntentNonce(ctx context.Context) (*big.Int, error)
func (c *Client) SignAndSubmitTradeIntent(ctx context.Context, intent TradeIntent) (*TradeResult, error)
func (c *Client) SimulateIntent(ctx context.Context, intent TradeIntent) (bool, string, error)
```

`SignAndSubmitTradeIntent()` should:
1. Sign the intent with agent wallet using `SignTradeIntent()`
2. Call `c.riskRouter.SubmitTradeIntent(c.agentAuth, intentStruct, signature)`
3. Wait for transaction receipt with 60s timeout
4. Parse `TradeApproved` or `TradeRejected` event from receipt logs
5. Return `TradeResult` with approval status, tx hash, gas used, and reason

`SimulateIntent()` should:
1. Call `c.riskRouter.SimulateIntent(nil, intentStruct)` (view call, no auth needed)
2. Return (approved, reason)

**Acceptance criteria:**
- `GetIntentNonce()` returns current nonce from RiskRouter for agent wallet
- `SignAndSubmitTradeIntent()` correctly signs, submits, and parses events
- `SimulateIntent()` performs a dry-run call
- Transaction receipt parsing handles both approval and rejection
- Timeout of 60s on waiting for receipt
- Errors include tx hash for debugging
- Unit tests with mocked RiskRouter binding

---

## Task 7: Implement validation checkpoint posting

**Files to create:**
- `pkg/blockchain/validation.go`

**What to do:**

```go
type Checkpoint struct {
    AgentId         *big.Int
    Timestamp       uint64
    Action          string
    Pair            string
    AmountUsdScaled *big.Int
    PriceUsdScaled  *big.Int
    ReasoningHash   common.Hash
}

func (c *Client) PostCheckpoint(ctx context.Context, agentId *big.Int, checkpointHash common.Hash, score uint8, notes string) (common.Hash, error)
func (c *Client) GetValidationScore(ctx context.Context, agentId *big.Int) (*big.Int, error)
func (c *Client) GetAttestations(ctx context.Context, agentId *big.Int) ([]Attestation, error)
```

`PostCheckpoint()` should:
1. Call `c.validationRegistry.PostEIP712Attestation(c.auth, agentId, checkpointHash, score, notes)`
2. Wait for transaction receipt
3. Return the tx hash

`GetValidationScore()` should:
1. Call `c.validationRegistry.GetAverageValidationScore(nil, agentId)`
2. Return the score

**Acceptance criteria:**
- `PostCheckpoint()` submits attestation and returns tx hash
- `GetValidationScore()` returns average score (0 if no attestations)
- `GetAttestations()` returns list of all attestations for an agent
- Score validation (0-100 range)
- Unit tests with mocked ValidationRegistry binding

---

## Task 8: Implement reputation registry interactions

**Files to create:**
- `pkg/blockchain/reputation.go`

**What to do:**

```go
type FeedbackEntry struct {
    Rater        common.Address
    Score        uint8
    OutcomeRef   common.Hash
    Comment      string
    Timestamp    *big.Int
    FeedbackType uint8
}

func (c *Client) SubmitFeedback(ctx context.Context, agentId *big.Int, score uint8, outcomeRef common.Hash, comment string, feedbackType uint8) (common.Hash, error)
func (c *Client) GetReputationScore(ctx context.Context, agentId *big.Int) (*big.Int, error)
func (c *Client) GetFeedbackHistory(ctx context.Context, agentId *big.Int) ([]FeedbackEntry, error)
func (c *Client) HasRated(ctx context.Context, agentId *big.Int, rater common.Address) (bool, error)
```

`SubmitFeedback()` should:
1. Call `c.reputationRegistry.SubmitFeedback(c.auth, agentId, score, outcomeRef, comment, feedbackType)`
2. Wait for receipt
3. Return tx hash

**Acceptance criteria:**
- All methods correctly call the ReputationRegistry contract
- `GetReputationScore()` returns 0 when no feedback exists
- `HasRated()` prevents double-rating
- Unit tests with mocked ReputationRegistry binding

---

## Task 9: Create trade executor (`internal/executor/`)

**Files to create:**
- `internal/executor/executor.go`
- `internal/executor/interfaces.go`
- `internal/executor/executor_test.go`

**What to do:**

Create `internal/executor/interfaces.go`:
```go
package executor

import (
    "context"
    "kraken-trader/internal/decision"
)

type BlockchainClient interface {
    AgentID() *big.Int
    AgentAddress() common.Address
    GetIntentNonce(ctx context.Context) (*big.Int, error)
    SignAndSubmitTradeIntent(ctx context.Context, intent blockchain.TradeIntent) (*blockchain.TradeResult, error)
    SimulateIntent(ctx context.Context, intent blockchain.TradeIntent) (bool, string, error)
    PostCheckpoint(ctx context.Context, agentId *big.Int, checkpointHash common.Hash, score uint8, notes string) (common.Hash, error)
    GetVaultBalance(ctx context.Context, agentId *big.Int) (*big.Int, error)
    GetReputationScore(ctx context.Context, agentId *big.Int) (*big.Int, error)
}

type MarketDataProvider interface {
    GetTicker(ctx context.Context, pair string) (*Ticker, error)
}

type PortfolioProvider interface {
    GetPortfolioValue() float64
}

type TradeRecorder interface {
    RecordTrade(ctx context.Context, trade TradeRecord) error
}
```

Create `internal/executor/executor.go`:
```go
type TradeExecutor struct {
    bc       BlockchainClient
    kraken   MarketDataProvider
    portfolio PortfolioProvider
    recorder TradeRecorder
    config   Config
}

type Config struct {
    MaxSlippageBps  uint64
    IntentDeadline  time.Duration
    MinConfidence   float64
    TradingMode     string // "paper" or "live"
}

func New(bc BlockchainClient, kraken MarketDataProvider, portfolio PortfolioProvider, recorder TradeRecorder, cfg Config) *TradeExecutor

func (e *TradeExecutor) Execute(ctx context.Context, d *decision.TradeDecision) (*ExecutionResult, error)
func (e *TradeExecutor) PostCheckpoint(ctx context.Context, d *decision.TradeDecision, price float64) error
```

`Execute()` flow:
1. Validate: confidence >= MinConfidence, action != "hold"
2. Get portfolio value from PortfolioProvider
3. Calculate USD amount: `portfolioValue * (d.SizePct / 100)`
4. Scale to 18 decimals: `amountUsdScaled = amount * 1e18`
5. Fetch nonce from RiskRouter
6. Build TradeIntent with deadline = now + IntentDeadline
7. Call `bc.SignAndSubmitTradeIntent()`
8. If approved and TradingMode == "live": execute Kraken order
9. Record trade via TradeRecorder
10. Return ExecutionResult

**Acceptance criteria:**
- `Execute()` follows the full flow: validate → size → sign → submit → record
- Paper mode skips Kraken order execution but still submits on-chain
- Live mode executes Kraken order after on-chain approval
- Cooldown enforcement (no duplicate trades on same pair within cooldown)
- `PostCheckpoint()` hashes the decision and posts to ValidationRegistry
- All errors are wrapped with context
- Unit tests with mock dependencies

---

## Task 10: Add blockchain state persistence (SQLite + InfluxDB)

**Files to create/modify:**
- `internal/repository/agent_identity.go` (new)
- `internal/repository/trade_record.go` (modify existing)
- `internal/storage/influxdb.go` (add blockchain measurement)

**What to do:**

Create `internal/repository/agent_identity.go`:
```go
type AgentIdentity struct {
    ID              uint      `gorm:"primaryKey"`
    AgentID         string    `gorm:"uniqueIndex"`
    OperatorAddress string
    AgentAddress    string
    Name            string
    Description     string
    RegisteredAt    time.Time
    HasClaimed      bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type AgentIdentityRepository struct {
    db *gorm.DB
}

func NewAgentIdentityRepository(db *gorm.DB) *AgentIdentityRepository
func (r *AgentIdentityRepository) Get() (*AgentIdentity, error)
func (r *AgentIdentityRepository) Save(identity *AgentIdentity) error
func (r *AgentIdentityRepository) UpdateClaimed(agentID string, claimed bool) error
```

Modify trade record model to add blockchain fields:
```go
type TradeRecord struct {
    // Existing fields...
    Pair       string
    Action     string
    SizePct    float64
    Confidence float64
    Reasoning  string
    CreatedAt  time.Time
    // New blockchain fields
    TxHash          string
    IntentHash      string
    OnChainApproved bool
    CheckpointHash  string
    GasUsed         uint64
    BlockNumber     uint64
}
```

Add InfluxDB measurement `blockchain_trades`:
```go
type BlockchainTrade struct {
    Time           time.Time
    AgentID        string `influxdb:"tag"`
    Pair           string `influxdb:"tag"`
    Action         string `influxdb:"tag"`
    TxHash         string `influxdb:"tag"`
    IntentHash     string
    AmountUsdScaled string
    GasUsed        uint64
    Approved       bool
    CheckpointHash string
}
```

**Acceptance criteria:**
- Auto-migration creates `agent_identities` table
- `AgentIdentityRepository` CRUD operations work
- Trade records persist blockchain fields
- InfluxDB writes `blockchain_trades` measurement
- Existing tests still pass

---

## Task 11: Wire blockchain client into application startup

**Files to modify:**
- `cmd/trader/run.go`
- `cmd/trader/main.go` (possibly)

**What to do:**

In `runRun()` in `cmd/trader/run.go`, after existing component initialization:

1. Read blockchain config from viper
2. If `OPERATOR_PRIVATE_KEY` and `AGENT_PRIVATE_KEY` are set:
   - Create `blockchain.Config` from env vars
   - Call `blockchain.NewClient(ctx, bcCfg)`
   - Check if agent is registered; if not and auto-register is enabled, register
   - Check if allocation claimed; if not and auto-claim is enabled, claim
   - Create `executor.TradeExecutor` with the blockchain client
   - Pass executor to `runDecisionLoop()` and `runPipeline()`
3. If blockchain keys are NOT set:
   - Log warning: "Blockchain execution disabled — set OPERATOR_PRIVATE_KEY and AGENT_PRIVATE_KEY"
   - Continue in current paper-only mode

Modify `runDecisionLoop()` signature to accept an optional `*executor.TradeExecutor`:
```go
func runDecisionLoop(ctx context.Context, engine *decision.Engine, memMgr *state.MemoryManager, pub *messaging.Publisher, tradeExec *executor.TradeExecutor, cfg Config) {
    for pair := range memMgr.PriceAlertCh {
        decisions, err := engine.Decide(ctx, []string{pair})
        // ... existing error handling ...

        for _, d := range decisions {
            if d.Action != "hold" && d.Confidence >= cfg.ConfidenceThreshold {
                if tradeExec != nil {
                    result, err := tradeExec.Execute(ctx, d)
                    if err != nil {
                        log.Error().Err(err).Str("pair", pair).Msg("Trade execution failed")
                    } else {
                        log.Info().Str("pair", pair).Str("action", d.Action).
                            Bool("approved", result.Approved).
                            Str("tx_hash", result.TxHash.Hex()).
                            Msg("Trade executed")

                        // Post checkpoint
                        if result.Approved {
                            price := getCurrentPrice(memMgr, pair)
                            tradeExec.PostCheckpoint(ctx, d, price)
                        }
                    }
                }
                // ... existing NATS publish ...
            }
        }
    }
}
```

**Acceptance criteria:**
- App starts normally when blockchain keys are NOT set (backward compatible)
- App initializes blockchain client when keys ARE set
- Decision loop calls executor.Execute() when tradeExec is non-nil
- Graceful shutdown closes blockchain RPC connection
- No panics on nil executor (graceful degradation)

---

## Task 12: Add blockchain API endpoints

**Files to modify:**
- `internal/market/api.yaml` (add endpoints)
- `internal/api/implementation.go` (add handlers)

**What to do:**

Add to `api.yaml`:
```yaml
/api/blockchain/status:
  get:
    summary: Get blockchain agent status
    responses:
      200:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/BlockchainStatus'

/api/blockchain/register:
  post:
    summary: Register agent on-chain
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/RegisterRequest'
    responses:
      200:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterResponse'

/api/blockchain/claim:
  post:
    summary: Claim vault allocation
    responses:
      200:
        content:
          application/json:
            schema:
              type: object
              properties:
                success:
                  type: boolean
                tx_hash:
                  type: string

/api/blockchain/trades:
  get:
    summary: List on-chain trade history
    parameters:
      - name: limit
        in: query
        schema:
          type: integer
          default: 20
    responses:
      200:
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: '#/components/schemas/BlockchainTrade'

/api/blockchain/simulate:
  post:
    summary: Simulate a trade intent (dry-run)
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/SimulateRequest'
    responses:
      200:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SimulateResponse'
```

Add schemas:
```yaml
BlockchainStatus:
  type: object
  properties:
    connected:
      type: boolean
    agent_id:
      type: string
    operator_address:
      type: string
    agent_address:
      type: string
    is_registered:
      type: boolean
    has_claimed:
      type: boolean
    vault_balance:
      type: string
    reputation_score:
      type: string
    nonce:
      type: string

RegisterRequest:
  type: object
  required: [name, description]
  properties:
    name:
      type: string
    description:
      type: string
    capabilities:
      type: array
      items:
        type: string
    metadata_uri:
      type: string

RegisterResponse:
  type: object
  properties:
    success:
      type: boolean
    agent_id:
      type: string
    tx_hash:
      type: string

BlockchainTrade:
  type: object
  properties:
    pair:
      type: string
    action:
      type: string
    size_pct:
      type: number
    confidence:
      type: number
    tx_hash:
      type: string
    intent_hash:
      type: string
    approved:
      type: boolean
    gas_used:
      type: integer
    created_at:
      type: string
      format: date-time

SimulateRequest:
  type: object
  required: [pair, action, size_pct]
  properties:
    pair:
      type: string
    action:
      type: string
    size_pct:
      type: number

SimulateResponse:
  type: object
  properties:
    approved:
      type: boolean
    reason:
      type: string
    intent_hash:
      type: string
```

Implement handlers in `internal/api/implementation.go`:
```go
func (s *Server) GetBlockchainStatus(w http.ResponseWriter, r *http.Request)
func (s *Server) RegisterAgent(w http.ResponseWriter, r *http.Request)
func (s *Server) ClaimAllocation(w http.ResponseWriter, r *http.Request)
func (s *Server) GetBlockchainTrades(w http.ResponseWriter, r *http.Request)
func (s *Server) SimulateTradeIntent(w http.ResponseWriter, r *http.Request)
```

Each handler should:
1. Check if blockchain client is initialized
2. If not, return 503 Service Unavailable
3. Call the appropriate blockchain client method
4. Return JSON response

**Acceptance criteria:**
- `GET /api/blockchain/status` returns full agent status
- `POST /api/blockchain/register` registers agent and returns agentId
- `POST /api/blockchain/claim` claims allocation
- `GET /api/blockchain/trades` returns trade history from SQLite
- `POST /api/blockchain/simulate` dry-runs an intent against RiskRouter
- All endpoints return 503 when blockchain is not configured
- API code regenerated with `make generate`
- `make check` passes

---

## Task 13: Add CLI commands for agent lifecycle

**Files to create:**
- `cmd/trader/agent.go`

**What to do:**

Use Cobra to add `agent` subcommands. Add to `cmd/trader/main.go`:
```go
var agentCmd = &cobra.Command{
    Use:   "agent",
    Short: "Manage ERC-8004 agent identity",
}

var agentRegisterCmd = &cobra.Command{
    Use:   "register",
    Short: "Register agent on AgentRegistry",
    RunE:  runAgentRegister,
}

var agentClaimCmd = &cobra.Command{
    Use:   "claim",
    Short: "Claim vault allocation from HackathonVault",
    RunE:  runAgentClaim,
}

var agentStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show agent on-chain status",
    RunE:  runAgentStatus,
}

var agentReputationCmd = &cobra.Command{
    Use:   "reputation",
    Short: "Show agent reputation score",
    RunE:  runAgentReputation,
}
```

Implement each `runAgent*` function:
- `runAgentRegister`: Connect to blockchain, call `RegisterAgent()`, save to SQLite, print agentId
- `runAgentClaim`: Connect, call `ClaimAllocation()`, update SQLite
- `runAgentStatus`: Connect, query all contracts, print formatted table
- `runAgentReputation`: Connect, query ReputationRegistry + ValidationRegistry, print scores

Example output for `agent status`:
```
Agent Status
============
Agent ID:        42
Operator:        0x1234...5678
Agent Wallet:    0xabcd...ef01
Registered:      Yes
Claimed:         Yes
Vault Balance:   0.05 ETH
Reputation:      0 (no feedback yet)
Validation Score: 0 (no attestations yet)
Intent Nonce:    0
```

**Acceptance criteria:**
- `go run cmd/trader/main.go agent register --name "MyBot" --description "AI trader"` works
- `go run cmd/trader/main.go agent claim` claims allocation
- `go run cmd/trader/main.go agent status` prints formatted status
- `go run cmd/trader/main.go agent reputation` prints scores
- All commands handle missing keys gracefully with clear error messages
- Commands save state to SQLite after successful on-chain operations

---

## Task 14: Add error handling, retry logic, and circuit breaker

**Files to create/modify:**
- `pkg/blockchain/retry.go` (new)
- `pkg/blockchain/client.go` (modify — add circuit breaker state)

**What to do:**

Create `pkg/blockchain/retry.go`:
```go
type RetryConfig struct {
    MaxRetries   int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func WithRetry[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error)
```

Implement exponential backoff retry:
1. Call `fn()`
2. If error, sleep `delay`, multiply by `Multiplier`, cap at `MaxDelay`
3. Retry up to `MaxRetries` times
4. Return last error if all retries exhausted

Add circuit breaker to `Client`:
```go
type CircuitBreaker struct {
    mu            sync.Mutex
    failures      int
    threshold     int
    openUntil     time.Time
    cooldown      time.Duration
}

func (cb *CircuitBreaker) Allow() bool
func (cb *CircuitBreaker) RecordSuccess()
func (cb *CircuitBreaker) RecordFailure()
```

Circuit breaker logic:
1. Track consecutive failures
2. After `threshold` failures, open circuit for `cooldown` duration
3. While open, reject all calls immediately with `ErrCircuitOpen`
4. After cooldown, allow one call through (half-open)
5. If success, close circuit; if failure, re-open

Integrate into `SignAndSubmitTradeIntent()`:
1. Check `circuitBreaker.Allow()` before submitting
2. On success: `RecordSuccess()`
3. On failure: `RecordFailure()`
4. On `ErrCircuitOpen`: log warning, return error without attempting

**Acceptance criteria:**
- `WithRetry()` correctly implements exponential backoff
- Circuit breaker opens after 3 consecutive failures
- Circuit breaker closes after successful call in half-open state
- Trade executor handles circuit open gracefully (skips trade, logs warning)
- Unit tests for retry logic (mock fn that fails N times then succeeds)
- Unit tests for circuit breaker state transitions

---

## Task 15: Write integration tests

**Files to create:**
- `pkg/blockchain/client_integration_test.go`
- `internal/executor/executor_integration_test.go`

**What to do:**

Create integration tests that require `SEPOLIA_RPC_URL` and wallet keys to be set:

```go
func TestIntegration_FullTradeFlow(t *testing.T) {
    if os.Getenv("SEPOLIA_RPC_URL") == "" {
        t.Skip("Skipping integration test: SEPOLIA_RPC_URL not set")
    }

    // 1. Create client
    // 2. Register agent (or use existing)
    // 3. Claim allocation
    // 4. Build TradeIntent
    // 5. Sign and submit
    // 6. Verify approval
    // 7. Post checkpoint
    // 8. Verify attestation
    // 9. Check reputation
}
```

Test scenarios:
1. **Full trade flow**: Register → Claim → Submit Intent → Approve → Checkpoint
2. **Simulate before submit**: Simulate returns same result as actual submit
3. **Nonce incrementing**: Two consecutive submits use incrementing nonces
4. **Invalid intent rejection**: Submit intent with invalid action, verify rejection
5. **Circuit breaker**: Simulate 3 failures, verify circuit opens
6. **Retry logic**: Mock RPC that fails twice then succeeds

**Acceptance criteria:**
- Tests are skipped when env vars not set (no CI breakage)
- Full trade flow test passes on Sepolia with real contracts
- All error paths are tested
- `make test` passes (integration tests skipped)
- `make test-integration` runs with Docker + env vars

---

## Task Execution Order

Tasks must be completed in this order due to dependencies:

```
Task 1 (deps/config)
  └── Task 2 (ABI bindings)
        └── Task 3 (client core)
              ├── Task 4 (registration/vault)
              ├── Task 5 (EIP-712 signing)
              │     └── Task 6 (trade submission)
              ├── Task 7 (validation checkpoints)
              └── Task 8 (reputation)
                    └── Task 9 (trade executor)
                          ├── Task 10 (persistence)
                          └── Task 11 (app wiring)
                                ├── Task 12 (API endpoints)
                                └── Task 13 (CLI commands)
                                      └── Task 14 (error handling)
                                            └── Task 15 (integration tests)
```

**Parallelizable after Task 3:**
- Tasks 4, 5, 7, 8 can be done in parallel (they all depend only on the client core)
- Tasks 12, 13 can be done in parallel (both depend on Task 11)

---

## AI Coding Agent Prompt

When delegating a task to an AI coding agent, use this prompt template:

```
You are working on the ERC-8004 integration for a Go-based crypto trading bot.

## Project Context
- Repository: /Users/paalgyula/Workspace/pilab/kraken-trader
- Language: Go 1.26+
- Entry point: cmd/trader/main.go
- Read AGENTS.md for coding conventions (imports, naming, error handling, testing, logging)
- The project uses zerolog, testify, cobra, viper, gorm, NATS, InfluxDB

## ERC-8004 Shared Contracts (Sepolia, Chain ID: 11155111)
- AgentRegistry: 0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3
- HackathonVault: 0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90
- RiskRouter: 0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC
- ReputationRegistry: 0x423a9904e39537a9997fbaF0f220d79D7d545763
- ValidationRegistry: 0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1

Full contract source code is verified on Etherscan Sepolia.

## Your Task

[INSERT SPECIFIC TASK NUMBER AND DESCRIPTION FROM ABOVE]

## Key Requirements
1. Follow the project's code conventions from AGENTS.md
2. Use zerolog for logging with module context
3. Wrap all errors with fmt.Errorf("context: %w", err)
4. Pass context.Context as first argument for I/O
5. Write tests using testify (assert + require)
6. Use table-driven tests where applicable
7. Do NOT add comments unless they explain non-obvious logic
8. Run `make check` (fmt, vet, lint) before declaring done
9. Run `make test` to verify existing tests still pass

## Acceptance Criteria

[INSERT ACCEPTANCE CRITERIA FROM THE SPECIFIC TASK]

## Important Notes
- The go-ethereum library is already added as a dependency
- Contract ABIs and bindings are in pkg/blockchain/abi/ and pkg/blockchain/bindings/
- The blockchain.Client struct is already defined in pkg/blockchain/client.go
- Do NOT modify existing files outside the task scope
- If you need to reference contract ABIs, they match the Etherscan verified source
```
