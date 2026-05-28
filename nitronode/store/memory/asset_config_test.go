package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssetsConfig_verifyVariables tests the validation logic for asset configuration
func TestAssetsConfig_verifyVariables(t *testing.T) {
	// Test missing asset symbol
	t.Run("missing asset symbol", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Symbol: "", // Missing symbol
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.Error(t, err)
		assert.Equal(t, "missing asset symbol for asset[0]", err.Error())
	})

	// Test missing token address
	t.Run("missing token address", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:                  "USD Coin",
					Symbol:                "USDC",
					SuggestedBlockchainID: 1,
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "", // Missing address
						},
					},
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.Error(t, err)
		assert.Equal(t, "missing USD Coin token address for blockchain with id 1", err.Error())
	})

	// Test invalid token address
	t.Run("invalid token address", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:                  "USD Coin",
					Symbol:                "USDC",
					SuggestedBlockchainID: 1,
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "0xinvalid", // Invalid address
						},
					},
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.Error(t, err)
		assert.Equal(t, "invalid USD Coin token address '0xinvalid' for blockchain with id 1", err.Error())
	})

	// Test asset decimals exceeding token decimals
	t.Run("asset decimals exceed token decimals", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:                  "USD Coin",
					Symbol:                "USDC",
					Decimals:              8,
					SuggestedBlockchainID: 1,
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
							Decimals:     6,
						},
					},
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.Error(t, err)
		assert.Equal(t, "asset USDC decimals (8) must not exceed token USDC decimals (6) on blockchain 1", err.Error())
	})

	// Test asset decimals equal to token decimals (should pass)
	t.Run("asset decimals equal to token decimals", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:                  "USD Coin",
					Symbol:                "USDC",
					Decimals:              6,
					SuggestedBlockchainID: 1,
					Tokens: []TokenConfig{
						{
							Name:         "USD Coin",
							Symbol:       "USDC",
							BlockchainID: 1,
							Address:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
							Decimals:     6,
						},
					},
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.NoError(t, err)
	})

	// Test custom symbol for token (inherits from asset when empty)
	t.Run("custom symbol for token", func(t *testing.T) {
		cfg := AssetsConfig{
			Assets: []AssetConfig{
				{
					Name:                  "USD Coin",
					Symbol:                "USDC",
					SuggestedBlockchainID: 137,
					Tokens: []TokenConfig{
						{
							Name:         "Bridged USDC",
							Symbol:       "", // Should inherit "USDC" from asset
							BlockchainID: 137,
							Address:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
						},
					},
				},
			},
		}
		err := verifyAssetsConfig(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "USDC", cfg.Assets[0].Tokens[0].Symbol)
	})
}
