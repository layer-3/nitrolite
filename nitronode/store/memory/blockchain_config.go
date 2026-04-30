package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/layer-3/nitrolite/pkg/core"
	"gopkg.in/yaml.v3"
)

const (
	defaultBlockStep    = uint64(10000)
	blockchainsFileName = "blockchains.yaml"
)

var (
	blockchainNameRegex  = regexp.MustCompile(`^[a-z][a-z_]+[a-z]$`)
	contractAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// BlockchainsConfig represents the root configuration structure for all blockchain settings.
// It contains default contract addresses that apply to all blockchains unless overridden,
// and a list of individual blockchain configurations.
type BlockchainsConfig struct {
	Blockchains []BlockchainConfig `yaml:"blockchains"`
}

// BlockchainConfig represents configuration for a single blockchain.
// It includes connection details, contract addresses, and scanning parameters.
type BlockchainConfig struct {
	// Name is the blockchain identifier (e.g., "polygon_amoy", "base_sepolia")
	// Must match pattern: lowercase letters and underscores only
	Name string `yaml:"name"`
	// ID is the chain ID used for RPC validation
	ID uint64 `yaml:"id"`
	// TODO: blockchains must not be disabled in prod deployment
	Disabled bool `yaml:"disabled"`
	// BlockStep defines the block range for scanning (default: 10000)
	BlockStep uint64 `yaml:"block_step"`
	// ChannelHubAddress is the address of the ChannelHub contract on this blockchain
	ChannelHubAddress string `yaml:"channel_hub_address"`
	// ChannelHubSigValidators maps validator IDs to the addresses of signature validators for the ChannelHub contract on this blockchain
	ChannelHubSigValidators map[uint8]string `yaml:"channel_hub_sig_validators"`
	// LockingContractAddress is the address of the locking contract on this blockchain
	LockingContractAddress string `yaml:"locking_contract_address"`
}

// LoadEnabledBlockchains loads and validates blockchain configurations from a YAML file.
// It reads from <configDirPath>/blockchains.yaml, validates all settings,
// verifies RPC connections, and returns a map of enabled blockchains indexed by chain ID.
//
// The function performs the following validations:
// - Contract addresses format (0x + 40 hex chars)
// - Blockchain names (lowercase with underscores)
// - RPC endpoint availability and chain ID matching
// - Required contract addresses (using defaults when not specified)
func LoadEnabledBlockchains(configDirPath string) (map[uint64]BlockchainConfig, error) {
	blockchainsPath := filepath.Join(configDirPath, blockchainsFileName)
	f, err := os.Open(blockchainsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg BlockchainsConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	if err := verifyBlockchainsConfig(&cfg); err != nil {
		return nil, err
	}

	return getEnabledBlockchains(&cfg), nil
}

func verifyBlockchainsConfig(cfg *BlockchainsConfig) error {
	for i, bc := range cfg.Blockchains {
		if bc.Disabled {
			continue
		}

		if !blockchainNameRegex.MatchString(bc.Name) {
			return fmt.Errorf("invalid blockchain name '%s', should match snake_case format", bc.Name)
		}

		if bc.ChannelHubAddress == "" && bc.LockingContractAddress == "" {
			return fmt.Errorf("blockchain '%s' must specify at least one of channel_hub_address or locking_contract_address", bc.Name)
		}

		if bc.ChannelHubAddress != "" && !contractAddressRegex.MatchString(bc.ChannelHubAddress) {
			return fmt.Errorf("invalid channel hub address '%s' for blockchain '%s'", bc.ChannelHubAddress, bc.Name)
		}

		if bc.LockingContractAddress != "" && !contractAddressRegex.MatchString(bc.LockingContractAddress) {
			return fmt.Errorf("invalid locking contract address '%s' for blockchain '%s'", bc.LockingContractAddress, bc.Name)
		}

		if bc.BlockStep == 0 {
			cfg.Blockchains[i].BlockStep = defaultBlockStep
		}

		if bc.ChannelHubAddress != "" && len(core.ChannelSignerTypes) > 1 {
			for _, channelSignerType := range core.ChannelSignerTypes[1:] {
				validatorAddress, ok := bc.ChannelHubSigValidators[uint8(channelSignerType)]
				if !ok {
					return fmt.Errorf("blockchain '%s' must specify a signature validator address for channel signer type %d", bc.Name, channelSignerType)
				}
				if !contractAddressRegex.MatchString(validatorAddress) {
					return fmt.Errorf("invalid signature validator address '%s' for channel signer type %d on blockchain '%s'", validatorAddress, channelSignerType, bc.Name)
				}
			}
		}
	}

	return nil
}

// getEnabledBlockchains returns a map of enabled blockchains indexed by their chain ID.
// Only blockchains with enabled=true are included in the result.
func getEnabledBlockchains(cfg *BlockchainsConfig) map[uint64]BlockchainConfig {
	enabledBlockchains := make(map[uint64]BlockchainConfig)
	for _, bc := range cfg.Blockchains {
		if !bc.Disabled {
			enabledBlockchains[bc.ID] = bc
		}
	}
	return enabledBlockchains
}
