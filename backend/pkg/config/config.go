package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Exchange  ExchangeConfig         `yaml:"exchange"`
	Portfolio PortfolioConfig        `yaml:"portfolio"`
	Trading   TradingConfig          `yaml:"trading"`
}

// ExchangeConfig represents exchange configuration
type ExchangeConfig struct {
	Type          string             `yaml:"type"` // "virtual" or "binance"
	InitialPrices map[string]float64 `yaml:"initial_prices"`
	Binance       BinanceConfig      `yaml:"binance"`
}

// BinanceConfig represents Binance-specific configuration
type BinanceConfig struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
	UseTestnet bool  `yaml:"use_testnet"` // Use testnet for safe testing
}

// PortfolioConfig represents portfolio configuration
type PortfolioConfig struct {
	InitialBalances map[string]float64 `yaml:"initial_balances"`
}

// TradingConfig represents trading configuration
type TradingConfig struct {
	Symbols           []string `yaml:"symbols"`
	UpdateIntervalSec int      `yaml:"update_interval_sec"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Exchange.Type == "" {
		config.Exchange.Type = "virtual"
	}

	if config.Trading.UpdateIntervalSec == 0 {
		config.Trading.UpdateIntervalSec = 5
	}

	// Override with environment variables if present
	if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
		config.Exchange.Binance.APIKey = apiKey
	}
	if secretKey := os.Getenv("BINANCE_SECRET_KEY"); secretKey != "" {
		config.Exchange.Binance.SecretKey = secretKey
	}

	return &config, nil
}

// LoadOrDefault loads configuration or returns default config if file doesn't exist
func LoadOrDefault(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	return Load(path)
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Exchange: ExchangeConfig{
			Type: "virtual",
			InitialPrices: map[string]float64{
				"BTCUSDT": 45000.0,
				"ETHUSDT": 3000.0,
			},
		},
		Portfolio: PortfolioConfig{
			InitialBalances: map[string]float64{
				"USDT": 10000.0,
			},
		},
		Trading: TradingConfig{
			Symbols:           []string{"BTCUSDT", "ETHUSDT"},
			UpdateIntervalSec: 5,
		},
	}
}
