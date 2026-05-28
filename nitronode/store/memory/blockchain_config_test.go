package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainConfig_verifyVariables(t *testing.T) {
	tcs := []struct {
		name             string
		cfg              BlockchainsConfig
		expectedErrorStr string
		assertFunc       func(t *testing.T, blockchains []BlockchainConfig)
	}{
		{
			name: "valid config",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:                      1,
						Name:                    "ethereum",
						ChannelHubAddress:       "0x1111111111111111111111111111111111111111",
						BlockStep:               10,
						ChannelHubSigValidators: map[uint8]string{1: "0x3333333333333333333333333333333333333333"},
					},
					{
						ID:                     11155111,
						Name:                   "ethereum_sepolia",
						LockingContractAddress: "0x2222222222222222222222222222222222222222",
					},
				},
			},
			expectedErrorStr: "",
			assertFunc: func(t *testing.T, blockchains []BlockchainConfig) {
				require.Len(t, blockchains, 2)

				ethCfg := blockchains[0]
				assert.Equal(t, "ethereum", ethCfg.Name)
				assert.Equal(t, uint64(1), ethCfg.ID)
				assert.Equal(t, "0x1111111111111111111111111111111111111111", ethCfg.ChannelHubAddress)
				assert.False(t, ethCfg.Disabled)
				assert.Equal(t, uint64(10), ethCfg.BlockStep)

				sepoliaCfg := blockchains[1]
				assert.Equal(t, "ethereum_sepolia", sepoliaCfg.Name)
				assert.Equal(t, uint64(11155111), sepoliaCfg.ID)
				assert.Equal(t, "0x2222222222222222222222222222222222222222", sepoliaCfg.LockingContractAddress)
				assert.False(t, sepoliaCfg.Disabled)
				assert.Equal(t, defaultBlockStep, sepoliaCfg.BlockStep)
			},
		},
		{
			name: "invalid name 1",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						Name: "Invalid Name!",
						ID:   1,
					},
				},
			},
			expectedErrorStr: "invalid blockchain name 'Invalid Name!', should match snake_case format",
		},
		{
			name: "invalid name 2",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						Name: "_foo_",
						ID:   1,
					},
				},
			},
			expectedErrorStr: "invalid blockchain name '_foo_', should match snake_case format",
		},
		{
			name: "disabled blockchain",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:                      1,
						Name:                    "ethereum",
						Disabled:                false,
						ChannelHubAddress:       "0x1111111111111111111111111111111111111111",
						ChannelHubSigValidators: map[uint8]string{1: "0x3333333333333333333333333333333333333333"},
					},
					{
						ID:       11155111,
						Name:     "_ethereum_sepolia_",
						Disabled: true,
					},
				},
			},
			expectedErrorStr: "",
			assertFunc: func(t *testing.T, blockchains []BlockchainConfig) {
				require.Len(t, blockchains, 2)

				ethCfg := blockchains[0]
				assert.Equal(t, "ethereum", ethCfg.Name)
				assert.Equal(t, uint64(1), ethCfg.ID)

				sepoliaCfg := blockchains[1]
				assert.Equal(t, "_ethereum_sepolia_", sepoliaCfg.Name)
				assert.Equal(t, uint64(11155111), sepoliaCfg.ID)
			},
		},
		{
			name: "invalid channel hub address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:                1,
						Name:              "ethereum",
						ChannelHubAddress: "0x0000s00000000000000000000000000000000001",
					},
				},
			},
			expectedErrorStr: "invalid channel hub address '0x0000s00000000000000000000000000000000001' for blockchain 'ethereum'",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyBlockchainsConfig(&tc.cfg)
			if tc.expectedErrorStr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorStr, err.Error())
				return
			}

			require.NoError(t, err)
			if tc.assertFunc != nil {
				tc.assertFunc(t, tc.cfg.Blockchains)
			}
		})
	}
}
