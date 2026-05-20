package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/shopspring/decimal"
)

type Config struct {
	ServerPort string `env:"SERVER_PORT" env-default:"8080" env-description:"HTTP server port"`

	OwnerPrivateKey  string `env:"OWNER_PRIVATE_KEY" env-required:"true" env-description:"Private key for faucet owner wallet (without 0x prefix)"`
	ClearnodeURL     string `env:"CLEARNODE_URL" env-required:"true" env-description:"Clearnode WebSocket URL"`
	TokenSymbol       string `env:"TOKEN_SYMBOL" env-required:"true" env-description:"Token symbol to distribute (e.g., usdc, weth)"`
	StandardTipAmount string `env:"STANDARD_TIP_AMOUNT" env-required:"true" env-description:"Default amount to send per request"`
	MinTransferCount  int    `env:"MIN_TRANSFER_COUNT" env-required:"true" env-description:"Number of transfers a server should have a balance for to operate"`
	CooldownPeriod    string `env:"COOLDOWN_PERIOD" env-required:"true" env-description:"Cooldown between requests per wallet/IP (e.g. 24h, 1h30m)"`

	LogLevel string `env:"LOG_LEVEL" env-default:"info" env-description:"Logging level (debug, info, warn, error)"`

	// Parsed values (set after loading)
	StandardTipAmountDecimal decimal.Decimal
	CooldownPeriodDuration   time.Duration
}

func Load() (*Config, error) {
	var config Config

	// Try to read from .env file first, then from environment variables
	if err := cleanenv.ReadConfig(".env", &config); err != nil {
		// If .env doesn't exist, try to read from environment variables only
		if err := cleanenv.ReadEnv(&config); err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	// Parse the decimal amount
	amount, err := decimal.NewFromString(c.StandardTipAmount)
	if err != nil {
		return fmt.Errorf("STANDARD_TIP_AMOUNT must be a valid decimal number: %w", err)
	}

	if amount.IsZero() || amount.IsNegative() {
		return fmt.Errorf("STANDARD_TIP_AMOUNT must be a positive number")
	}

	c.StandardTipAmountDecimal = amount

	d, err := time.ParseDuration(c.CooldownPeriod)
	if err != nil {
		return fmt.Errorf("COOLDOWN_PERIOD must be a valid duration (e.g. 24h, 1h30m): %w", err)
	}
	if d <= 0 {
		return fmt.Errorf("COOLDOWN_PERIOD must be positive")
	}
	c.CooldownPeriodDuration = d

	return nil
}
