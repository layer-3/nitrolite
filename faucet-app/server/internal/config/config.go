package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/shopspring/decimal"
)

// Config holds all runtime configuration for the faucet server.
type Config struct {
	ServerPort string `env:"SERVER_PORT" env-default:"8080" env-description:"HTTP server port"`

	OwnerPrivateKey  string `env:"OWNER_PRIVATE_KEY" env-required:"true" env-description:"Private key for faucet owner wallet (without 0x prefix)"`
	NitronodeURL     string `env:"NITRONODE_URL" env-required:"true" env-description:"Nitronode WebSocket URL"`
	TokenSymbol      string `env:"TOKEN_SYMBOL" env-required:"true" env-description:"Token symbol to distribute (e.g., usdc, weth)"`
	StandardTipAmount string `env:"STANDARD_TIP_AMOUNT" env-required:"true" env-description:"Default amount to send per request"`
	MinTransferCount  int    `env:"MIN_TRANSFER_COUNT" env-required:"true" env-description:"Number of transfers a server should have a balance for to operate"`
	CooldownPeriod    string `env:"COOLDOWN_PERIOD" env-required:"true" env-description:"Cooldown between requests per wallet/IP (e.g. 24h, 1h30m)"`
	TrustedProxies    string `env:"TRUSTED_PROXIES" env-default:"" env-description:"Comma-separated trusted proxy IPs; empty means direct exposure only"`

	LogLevel string `env:"LOG_LEVEL" env-default:"info" env-description:"Logging level (debug, info, warn, error)"`

	// Parsed values (set after loading)
	StandardTipAmountDecimal decimal.Decimal
	CooldownPeriodDuration   time.Duration
	TrustedProxyList         []string
}

// Load reads configuration from a .env file (if present) or environment variables.
// A missing .env file is not an error; any other read or validation failure is.
func Load() (*Config, error) {
	var config Config

	if _, statErr := os.Stat(".env"); statErr == nil {
		if err := cleanenv.ReadConfig(".env", &config); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to stat .env file: %w", statErr)
	} else {
		if err := cleanenv.ReadEnv(&config); err != nil {
			return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
		}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate parses and range-checks all fields that require post-load processing.
func (c *Config) Validate() error {
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

	if c.MinTransferCount <= 0 {
		return fmt.Errorf("MIN_TRANSFER_COUNT must be a positive integer")
	}

	if c.TrustedProxies != "" {
		for _, p := range strings.Split(c.TrustedProxies, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				c.TrustedProxyList = append(c.TrustedProxyList, trimmed)
			}
		}
	}

	return nil
}
