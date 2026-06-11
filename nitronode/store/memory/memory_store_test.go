package memory

import (
	"errors"
	"testing"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBlockchainID        uint64 = 137
	testEnabledNoTokensID   uint64 = 480 // ChannelHub-enabled chain with no configured tokens
	testTokenAddress               = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	testUnknownTokenAddress        = "0x000000000000000000000000000000000000dead"
)

func newTestStore() *MemoryStoreV1 {
	return &MemoryStoreV1{
		tokenAssets: map[uint64]map[string]core.Asset{
			testBlockchainID: {testTokenAddress: {Symbol: "usdc"}},
		},
		tokenDecimals: map[uint64]map[string]uint8{
			testBlockchainID: {testTokenAddress: 6},
		},
	}
}

func TestMemoryStoreV1_GetTokenAsset(t *testing.T) {
	ms := newTestStore()

	t.Run("configured token returns asset", func(t *testing.T) {
		asset, err := ms.GetTokenAsset(testBlockchainID, testTokenAddress)
		require.NoError(t, err)
		assert.Equal(t, "usdc", asset)
	})

	t.Run("unknown token on configured chain wraps ErrTokenNotSupported", func(t *testing.T) {
		_, err := ms.GetTokenAsset(testBlockchainID, testUnknownTokenAddress)
		require.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrTokenNotSupported))
	})

	// MF3-H03 regression: a ChannelHub-enabled chain with no configured tokens
	// must also surface ErrTokenNotSupported so the reactor skips the event
	// instead of stopping the listener.
	t.Run("chain with no configured tokens wraps ErrTokenNotSupported", func(t *testing.T) {
		_, err := ms.GetTokenAsset(testEnabledNoTokensID, testTokenAddress)
		require.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrTokenNotSupported))
	})
}

func TestMemoryStoreV1_GetTokenDecimals(t *testing.T) {
	ms := newTestStore()

	t.Run("configured token returns decimals", func(t *testing.T) {
		decimals, err := ms.GetTokenDecimals(testBlockchainID, testTokenAddress)
		require.NoError(t, err)
		assert.Equal(t, uint8(6), decimals)
	})

	t.Run("unknown token on configured chain wraps ErrTokenNotSupported", func(t *testing.T) {
		_, err := ms.GetTokenDecimals(testBlockchainID, testUnknownTokenAddress)
		require.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrTokenNotSupported))
	})

	// MF3-H03 regression: see GetTokenAsset counterpart above.
	t.Run("chain with no configured tokens wraps ErrTokenNotSupported", func(t *testing.T) {
		_, err := ms.GetTokenDecimals(testEnabledNoTokensID, testTokenAddress)
		require.Error(t, err)
		assert.True(t, errors.Is(err, core.ErrTokenNotSupported))
	})
}
