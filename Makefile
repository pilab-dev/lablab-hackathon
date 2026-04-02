.PHONY: help setup docker-up docker-down test run clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Initialize Go module and download dependencies
	go mod init kraken-trader
	go mod tidy

docker-up: ## Start Docker services (InfluxDB, ChromaDB)
	docker compose -f configs/docker-compose.yml up -d

docker-down: ## Stop Docker services
	docker compose -f configs/docker-compose.yml down

docker-logs: ## Show Docker service logs
	docker compose -f configs/docker-compose.yml logs -f

ollama-setup: ## Pull required Ollama models
	ollama pull llama3.1:8b
	ollama pull nomic-embed-text

ollama-serve: ## Start Ollama server
	ollama serve

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests (requires Docker services)
	go test -v -race -tags=integration ./...

run: ## Run the trader bot
	go run cmd/trader/main.go

run-with-env: ## Run the trader bot with .env file
	set -a && . .env && set +a && go run cmd/trader/main.go

build: ## Build the binary
	go build -o bin/kraken-trader ./cmd/trader

clean: ## Remove build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

check: fmt vet lint ## Run all checks

kraken-test: ## Test kraken-cli connection
	kraken ticker BTCUSD -o json

kraken-paper-init: ## Initialize paper trading with $10,000
	kraken paper init --balance 10000 -o json

kraken-paper-status: ## Check paper trading status
	kraken paper status

influxdb-shell: ## Open InfluxDB shell
	docker exec -it kraken-trader-influxdb influx shell -t admin123 -o kraken-trader -b market-data

chroma-shell: ## Open ChromaDB shell (curl test)
	curl http://localhost:8000/api/v1/heartbeat
