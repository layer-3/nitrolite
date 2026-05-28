package stress

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/layer-3/nitrolite/pkg/sign"
)

// Config holds all stress test settings, read from environment variables.
type Config struct {
	WsURL        string        `env:"STRESS_WS_URL,required"`
	PrivateKey   string        `env:"STRESS_PRIVATE_KEY"`
	Connections  int           `env:"STRESS_CONNECTIONS" env-default:"10"`
	Timeout      time.Duration `env:"STRESS_TIMEOUT" env-default:"10m"`
	MaxErrorRate float64       `env:"STRESS_MAX_ERROR_RATE" env-default:"0.01"`
}

// ReadConfig reads the stress test configuration from environment variables.
// If STRESS_PRIVATE_KEY is not set, an ephemeral key is generated.
func ReadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read stress test config: %w", err)
	}
	if cfg.WsURL == "" {
		return nil, fmt.Errorf("STRESS_WS_URL is required")
	}
	if cfg.PrivateKey == "" {
		key, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
		}
		cfg.PrivateKey = hex.EncodeToString(ethcrypto.FromECDSA(key))
	}
	return &cfg, nil
}

// WalletAddress derives the wallet address from the configured private key.
// Returns empty string if no private key was explicitly configured.
func (c *Config) WalletAddress() (string, error) {
	signer, err := sign.NewEthereumRawSigner(c.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to derive wallet address: %w", err)
	}
	return signer.PublicKey().Address().String(), nil
}
