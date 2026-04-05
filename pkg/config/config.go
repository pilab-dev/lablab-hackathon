package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the Kraken Trader application
type Config struct {
	// Kraken API Keys
	KrakenAPIKey    string `mapstructure:"KRAKEN_API_KEY"`
	KrakenAPISecret string `mapstructure:"KRAKEN_API_SECRET"`

	// Trading Configuration
	TradingMode         string        `mapstructure:"TRADING_MODE"`
	TradeInterval       time.Duration `mapstructure:"TRADE_INTERVAL"`
	ConfidenceThreshold float64       `mapstructure:"CONFIDENCE_THRESHOLD"`
	TradeCooldown       time.Duration `mapstructure:"TRADE_COOLDOWN"`

	// Ollama Configuration
	OllamaURL        string `mapstructure:"OLLAMA_URL"`
	OllamaModel      string `mapstructure:"OLLAMA_MODEL"`
	OllamaEmbedModel string `mapstructure:"OLLAMA_EMBED_MODEL"`

	// InfluxDB Configuration
	InfluxDBURL    string `mapstructure:"INFLUXDB_URL"`
	InfluxDBToken  string `mapstructure:"INFLUXDB_TOKEN"`
	InfluxDBOrg    string `mapstructure:"INFLUXDB_ORG"`
	InfluxDBBucket string `mapstructure:"INFLUXDB_BUCKET"`

	// ChromaDB Configuration
	ChromaURL        string `mapstructure:"CHROMA_URL"`
	ChromaCollection string `mapstructure:"CHROMA_COLLECTION"`

	// NATS Configuration
	NATSURL string `mapstructure:"NATS_URL"`

	// Dashboard
	DashboardPort int `mapstructure:"DASHBOARD_PORT"`

	// HTTP API Server
	APIPort int `mapstructure:"API_PORT"`

	// SQLite Database
	SQLitePath string `mapstructure:"SQLITE_PATH"`

	// Logging
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// PRISM API
	PrismAPIKey string `mapstructure:"PRISM_API_KEY"`

	// Redis Configuration
	RedisURL      string `mapstructure:"REDIS_URL"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// ERC-8004 Blockchain Configuration
	SepoliaRPCURL          string `mapstructure:"SEPOLIA_RPC_URL"`
	SepoliaChainID         uint64 `mapstructure:"SEPOLIA_CHAIN_ID"`
	OperatorPrivateKey     string `mapstructure:"OPERATOR_PRIVATE_KEY"`
	AgentPrivateKey        string `mapstructure:"AGENT_PRIVATE_KEY"`
	AgentID                string `mapstructure:"AGENT_ID"`
	AgentRegistryAddr      string `mapstructure:"AGENT_REGISTRY_ADDRESS"`
	HackathonVaultAddr     string `mapstructure:"HACKATHON_VAULT_ADDRESS"`
	RiskRouterAddr         string `mapstructure:"RISK_ROUTER_ADDRESS"`
	ReputationRegistryAddr string `mapstructure:"REPUTATION_REGISTRY_ADDRESS"`
	ValidationRegistryAddr string `mapstructure:"VALIDATION_REGISTRY_ADDRESS"`
	GasLimit               uint64 `mapstructure:"GAS_LIMIT"`
	GasPriceGwei           uint64 `mapstructure:"GAS_PRICE_GWEI"`
}

// LoadConfig reads configuration from environment variables and config file
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Bind environment variables
	v.AutomaticEnv()

	// Read config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	v.SetDefault("TRADING_MODE", "paper")
	v.SetDefault("TRADE_INTERVAL", "30s")
	v.SetDefault("CONFIDENCE_THRESHOLD", 0.6)
	v.SetDefault("TRADE_COOLDOWN", "60s")
	v.SetDefault("OLLAMA_URL", "http://localhost:11434")
	v.SetDefault("OLLAMA_MODEL", "llama3.1:8b")
	v.SetDefault("OLLAMA_EMBED_MODEL", "nomic-embed-text")
	v.SetDefault("INFLUXDB_URL", "http://localhost:8086")
	v.SetDefault("INFLUXDB_TOKEN", "admin123")
	v.SetDefault("INFLUXDB_ORG", "kraken-trader")
	v.SetDefault("INFLUXDB_BUCKET", "market-data")
	v.SetDefault("CHROMA_URL", "http://localhost:8000")
	v.SetDefault("CHROMA_COLLECTION", "news-embeddings")
	v.SetDefault("NATS_URL", "nats://localhost:4222")
	v.SetDefault("DASHBOARD_PORT", 8080)
	v.SetDefault("API_PORT", 8081)
	v.SetDefault("SQLITE_PATH", "./trader.db")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("REDIS_URL", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("SEPOLIA_RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com")
	v.SetDefault("SEPOLIA_CHAIN_ID", 11155111)
	v.SetDefault("GAS_LIMIT", 300000)
	v.SetDefault("GAS_PRICE_GWEI", 2)
	v.SetDefault("AGENT_REGISTRY_ADDRESS", "0x97b07dDc405B0c28B17559aFFE63BdB3632d0ca3")
	v.SetDefault("HACKATHON_VAULT_ADDRESS", "0x0E7CD8ef9743FEcf94f9103033a044caBD45fC90")
	v.SetDefault("RISK_ROUTER_ADDRESS", "0xd6A6952545FF6E6E6681c2d15C59f9EB8F40FdBC")
	v.SetDefault("REPUTATION_REGISTRY_ADDRESS", "0x423a9904e39537a9997fbaF0f220d79D7d545763")
	v.SetDefault("VALIDATION_REGISTRY_ADDRESS", "0x92bF63E5C7Ac6980f237a7164Ab413BE226187F1")
}
