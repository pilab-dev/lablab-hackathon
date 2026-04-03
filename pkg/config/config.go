package config

import (
	"fmt"
	"strings"
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
	TradePairs          []string      `mapstructure:"TRADE_PAIRS"`
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

	// Logging
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// PRISM API
	PrismAPIKey string `mapstructure:"PRISM_API_KEY"`
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

	// Parse comma-separated strings
	if tradePairsStr := v.GetString("TRADE_PAIRS"); tradePairsStr != "" {
		cfg.TradePairs = strings.Split(tradePairsStr, ",")
		for i, pair := range cfg.TradePairs {
			cfg.TradePairs[i] = strings.TrimSpace(pair)
		}
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	v.SetDefault("TRADING_MODE", "paper")
	v.SetDefault("TRADE_PAIRS", []string{"BTCUSD", "ETHUSD"})
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
	v.SetDefault("LOG_LEVEL", "info")
}
