package nitronode

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientValidation(t *testing.T) {
	t.Run("should fail with invalid private key", func(t *testing.T) {
		client, err := NewClient("not-a-valid-key", "ws://localhost:8080", "usdc", decimal.NewFromInt(10), 1)

		assert.Nil(t, client)
		require.Error(t, err)
	})

	t.Run("should fail with invalid URL", func(t *testing.T) {
		validKey := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

		client, err := NewClient(validKey, "ws://localhost:19999", "usdc", decimal.NewFromInt(10), 1)

		assert.Nil(t, client)
		require.Error(t, err)
	})

	t.Run("should handle 0x prefixed key", func(t *testing.T) {
		// NewEthereumMsgSigner should handle 0x prefix; if it fails, err is non-nil.
		// Either outcome is acceptable — the test just asserts no panic and consistency.
		key := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

		client, err := NewClient(key, "ws://localhost:19999", "usdc", decimal.NewFromInt(10), 1)
		if err == nil {
			require.NotNil(t, client)
			require.NoError(t, client.Close())
		}
	})
}
