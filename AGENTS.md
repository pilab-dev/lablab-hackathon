# AGENTS.md — Developer Guide for AI Agents

This document provides essential information for AI agents working in this repository.

## Development Workflow: API-First Approach

**Always follow this order:**

1. **Define OpenAPI Spec** — Write `api.yaml` first with all endpoints, request/response schemas
2. **Generate Mappings** — Create handler functions and types from the spec (e.g., `internal/api/`)
3. **Implement Logic** — Write the actual business logic to fulfill the API contracts

> Never start with implementation. The OpenAPI spec is the source of truth.

---

## Project Overview

- **Language:** Go 1.26+
- **Type:** Autonomous crypto trading bot (lablab.ai AI Trading Agents Hackathon)
- **Architecture:** NATS messaging, Ollama (Llama 3.1), Kraken CLI, InfluxDB, ChromaDB
- **Entry Point:** `cmd/trader/main.go`

---

## Build & Development Commands

### Running the Project

```bash
make run              # Run the trader bot
make run-with-env    # Run with .env file loaded
make build           # Build binary to bin/kraken-trader
make clean           # Remove build artifacts
```

### Testing

```bash
make test            # Run all tests with coverage
make test-integration # Run integration tests (requires Docker)
```

**Run a single test:**
```bash
go test -v ./pkg/kraken -run TestNewClient
go test -v ./internal/decision -run TestEngine
```

### Linting & Code Quality

```bash
make lint            # Run golangci-lint
make fmt            # Format code (go fmt)
make vet            # Run go vet
make check          # Run fmt, vet, lint
make generate       # Generate API code from OpenAPI spec (uses //go:generate)
```

### Generating API Code

API definitions are in `internal/market/api.yaml`. After editing the spec:
```bash
go generate ./internal/api     # Regenerate from //go:generate directive
# or
make generate                  # Via Makefile
```

### Docker Services

```bash
make docker-up      # Start InfluxDB & ChromaDB
make docker-down    # Stop services
make docker-logs    # Follow logs
```

---

## Code Style Guidelines

### Imports

Group imports in the standard Go order:
1. Standard library (`bufio`, `context`, `encoding/json`, etc.)
2. External packages (`github.com/...`)
3. Internal packages (`kraken-trader/internal/...`)

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "time"

    "kraken-trader/internal/news"
    "kraken-trader/internal/state"
)
```

### Naming Conventions

- **Variables/functions:** camelCase (`binPath`, `RunRaw`)
- **Exported types/constants:** PascalCase (`Client`, `DefaultTimeout`)
- **Struct fields:** PascalCase (JSON tags handle JSON serialization)
- **Acronyms:** Keep uppercase (e.g., `API`, `URL`, `JSON`)
- **Booleans:** Use `is/has/should` prefixes where appropriate

### Error Handling

- Use `fmt.Errorf` with `%w` for wrapped errors
- Return errors up the call stack; handle at top levels
- Include context in error messages: `"failed to parse kraken-cli json response: %w"`
- Check errors explicitly; avoid bare `if err != nil`

```go
if err := json.Unmarshal(out, target); err != nil {
    return fmt.Errorf("failed to parse kraken-cli json response: %w\nRaw output: %s", err, string(out))
}
```

### Context Usage

- Pass `context.Context` as first argument to functions that perform I/O
- Use `context.Background()` for top-level entry points
- Use `context.WithTimeout` for operations with deadlines
- Always respect context cancellation in loops

### HTTP Clients

- Create `http.Client` with explicit timeouts (e.g., `60 * time.Second` for LLM inference)
- Close response bodies with `defer resp.Body.Close()`
- Check status codes explicitly

### Testing

- Test files: `*_test.go` in same package as code
- Use **stretchr/testify** (`github.com/stretchr/testify/assert` and `require`)
- Prefer table-driven tests for multiple cases
- Mock external dependencies (e.g., exec.Command)
- Test error paths, not just happy path

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
    c := NewClient("")
    assert.Equal(t, "kraken", c.binPath, "Expected default binPath")
}

func TestErrorEnvelope_Unmarshal(t *testing.T) {
    rawJSON := `{"error": "rate_limit", "message": "API rate limit exceeded"}`
    var env ErrorEnvelope
    err := json.Unmarshal([]byte(rawJSON), &env)
    require.NoError(t, err)
    assert.Equal(t, "rate_limit", env.Error)
}
```

### Clean Architecture

- Define **interfaces** in the layer that uses the dependency (e.g., `internal/decision/interfaces.go`)
- Implement concrete types in the appropriate layer (`internal/storage/`, `pkg/kraken/`)
- Dependencies point inward: domain → use cases → adapters → infrastructure
- Use constructor injection for all dependencies

```go
// Define interface where it's used (e.g., internal/decision/interfaces.go)
type MarketDataProvider interface {
    GetTicker(ctx context.Context, pair string) (*Ticker, error)
    GetOHLC(ctx context.Context, pair string, interval int) ([]OHLC, error)
}

// Use interface in business logic (e.g., internal/decision/engine.go)
type Engine struct {
    marketProvider MarketDataProvider
}

func NewEngine(mp MarketDataProvider) *Engine {
    return &Engine{marketProvider: mp}
}
```

### Mocking with gomock

- Use **go.uber.org/mock** (`go.uber.org/mock/gomock`) for generating mocks
- Generate mocks with: `go generate ./...` (add `//go:generate` directives)
- Store generated mocks in `mocks/` directory

```go
//go:generate mockgen -destination=mocks/market_provider.go -package=mocks . MarketDataProvider

type MarketDataProvider interface {
    GetTicker(ctx context.Context, pair string) (*Ticker, error)
}

// Usage in tests
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockProvider := NewMockMarketDataProvider(ctrl)
mockProvider.EXPECT().GetTicker(gomock.Any(), "BTCUSD").Return(&Ticker{Price: 50000}, nil)
```

### JSON Struct Tags

- Use explicit JSON tags for all exported struct fields
- Keep JSON keys lowercase: `json:"pair"`
- Group tags compactly: `json:"pair"`

```go
type TradeDecision struct {
    Pair       string  `json:"pair"`
    Action     string  `json:"action"`
    SizePct    float64 `json:"size_pct"`
    Confidence float64 `json:"confidence"`
    Reasoning  string  `json:"reasoning"`
}
```

### Logging

- Use **zerolog** for structured logging with pretty console output
- Configure with caller (file:line) for debugging: `zerolog.Caller().Logger().Output(os.Stdout)`
- Use level-based logging: `log.Debug().Msgf("failed to parse: %v", err)`
- Create module-specific loggers: `log.With().Str("module", "kraken").Logger()`

```go
import "github.com/rs/zerolog"

var log = zerolog.Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

func init() {
    zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// Usage
log.Debug().Str("binPath", c.binPath).Msg("executing command")
log.Info().Err(err).Str("url", url).Msg("request failed")
```

### Constants

- Group related constants; use iota for enum-like values
- Document exported constants with comments

---

## Project Structure

```
cmd/trader/          # Application entry point
cmd/trader/frontend/ # Embedded frontend build (auto-generated)
internal/api/        # OpenAPI generated types and handlers
internal/decision/   # LLM decision engine
internal/messaging/  # NATS messaging client
internal/news/       # News crawler, ChromaDB, embedder
internal/market/     # Market data collection
internal/repository/ # SQLite repositories
internal/state/      # In-memory state management
internal/storage/    # InfluxDB client & models
frontend/            # React + MUI frontend source
pkg/kraken/         # Kraken CLI wrapper
pkg/logger/         # Zerolog logger wrapper
pkg/config/         # Viper configuration
```

---

## Frontend Development

The frontend is a React + MUI application served as embedded static content.

### Directory Structure
```
frontend/
├── src/
│   ├── client/          # Orval-generated API client (auto-generated)
│   │   ├── api.ts       # API functions
│   │   └── models/      # TypeScript types
│   ├── hooks/           # React Query hooks
│   ├── pages/           # Page components
│   ├── theme/           # MUI theme config
│   ├── App.tsx          # Router + navigation
│   └── main.tsx         # App entry
├── orval.config.ts      # API client generation config
└── package.json
```

### Commands
```bash
make generate-client  # Generate TypeScript API client from OAS
make frontend-deps   # Install dependencies (pnpm)
make frontend-build  # Build production bundle
make frontend-copy  # Copy build to cmd/trader/frontend
make frontend       # All of the above (deps + generate + build + copy)
make build          # Builds Go binary with embedded frontend
```

### API-First Frontend Workflow

**IMPORTANT: Always follow this order for API changes:**

1. **Update OAS** — Edit `internal/market/api.yaml` with new endpoints/schemas
2. **Regenerate Go server** — `make generate` (regenerates `internal/api/generated.go`)
3. **Regenerate TS client** — `make generate-client` (regenerates `frontend/src/client/`)
4. **Update frontend** — Use generated types/functions in React components
5. **Rebuild** — `make build`

### Development
```bash
cd frontend && pnpm dev  # Vite dev server (proxies /api to localhost:8081)
```

### Key Points
- API client is auto-generated by **orval** from the OpenAPI spec
- All API calls use `/api` prefix (relative URL)
- Types are shared between Go backend and TypeScript frontend via OAS

---

## API Routes (prefixed with /api)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/subscriptions` | GET | List subscriptions |
| `/api/subscriptions` | POST | Add subscription |
| `/api/subscriptions/:symbol` | DELETE | Remove subscription |
| `/api/subscriptions/detail` | GET | List with details |
| `/api/assets` | GET | List tradable assets |
| `/api/ticker/:symbol` | GET | Get ticker data |
| `/api/prompts` | GET | List prompt history |
| `/api/loglevel` | GET | Get log level |
| `/api/loglevel` | POST | Set log level |
| `/api/swagger/` | GET | Swagger UI |

---

## External Dependencies

- **Kraken CLI:** Must be installed in PATH (`kraken` command)
- **Ollama:** Must be running with `llama3.1:8b` and `nomic-embed-text` models
- **Docker:** For InfluxDB and ChromaDB services

---

## Data Persistence

- **SQLite** (`trader.db`): Stores subscriptions and LLM prompts/responses
- **InfluxDB**: Stores ticker price data for dashboards

Subscriptions are restored from SQLite on startup with `created_at` and `last_data` timestamps.

---

## Important Files

- `Makefile` — All build/lint/test commands
- `go.mod` / `go.sum` — Dependency management
- `.env.example` — Environment template (copy to `.env`)
- `PHASE*_SPEC.md` — Feature specifications