package database

import (
	"strings"
	"testing"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreContractEvent(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	store := NewDBStore(db)

	event := core.BlockchainEvent{
		ContractAddress: "0x1234567890123456789012345678901234567890",
		BlockchainID:    1,
		Name:            "HomeChannelCreated",
		BlockNumber:     100,
		TransactionHash: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		LogIndex:        5,
	}

	err := store.StoreContractEvent(event)
	require.NoError(t, err)

	// Verify the event was stored
	var storedEvent ContractEvent
	err = db.Where("transaction_hash = ? AND log_index = ?", event.TransactionHash, event.LogIndex).First(&storedEvent).Error
	require.NoError(t, err)

	assert.Equal(t, event.ContractAddress, storedEvent.ContractAddress)
	assert.Equal(t, event.BlockchainID, storedEvent.BlockchainID)
	assert.Equal(t, event.Name, storedEvent.Name)
	assert.Equal(t, event.BlockNumber, storedEvent.BlockNumber)
	assert.Equal(t, event.TransactionHash, storedEvent.TransactionHash)
	assert.Equal(t, event.LogIndex, storedEvent.LogIndex)
}

func TestGetLatestContractEventBlockNumber(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	store := NewDBStore(db)

	contractAddress := "0x1234567890123456789012345678901234567890"
	blockchainID := uint64(1)

	t.Run("no events returns zero", func(t *testing.T) {
		block, err := store.GetLatestContractEventBlockNumber(contractAddress, blockchainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), block)
	})

	t.Run("returns max block number across multiple events", func(t *testing.T) {
		events := []core.BlockchainEvent{
			{ContractAddress: contractAddress, BlockchainID: blockchainID, Name: "E1", BlockNumber: 100, TransactionHash: "0xaaa", LogIndex: 0},
			{ContractAddress: contractAddress, BlockchainID: blockchainID, Name: "E2", BlockNumber: 200, TransactionHash: "0xbbb", LogIndex: 0},
			{ContractAddress: contractAddress, BlockchainID: blockchainID, Name: "E3", BlockNumber: 150, TransactionHash: "0xccc", LogIndex: 0},
		}
		for _, ev := range events {
			require.NoError(t, store.StoreContractEvent(ev))
		}

		block, err := store.GetLatestContractEventBlockNumber(contractAddress, blockchainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(200), block)
	})

	t.Run("different contract returns zero", func(t *testing.T) {
		block, err := store.GetLatestContractEventBlockNumber("0x9999999999999999999999999999999999999999", blockchainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), block)
	})

	t.Run("different blockchain returns zero", func(t *testing.T) {
		block, err := store.GetLatestContractEventBlockNumber(contractAddress, 999)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), block)
	})
}

// TestGetContractEventBlockHash_TrimsPaddedEmpty is the regression guard for the
// reorg-reconciliation bug: block_hash was a CHAR(66) NOT NULL DEFAULT '' column,
// so legacy/pre-migration rows read back as a 66-space-padded string rather than "".
// The reconciler's empty-hash guard compares against "", so an untrimmed padded
// value slipped through, got fed to common.HexToHash (-> zero hash), and made every
// stored block look reorged on every chain. The getters now TrimSpace so a padded
// empty collapses to "" the guard recognizes, while real 0x… hashes pass through.
func TestGetContractEventBlockHash_TrimsPaddedEmpty(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	store := NewDBStore(db)

	contractAddress := "0x1234567890123456789012345678901234567890"
	blockchainID := uint64(1)
	paddedEmpty := strings.Repeat(" ", 66) // mimics CHAR(66) padding of ''
	realHash := "0x" + strings.Repeat("ab", 32)

	t.Run("latest padded-empty hash trims to empty string", func(t *testing.T) {
		require.NoError(t, store.StoreContractEvent(core.BlockchainEvent{
			ContractAddress: contractAddress, BlockchainID: blockchainID, Name: "E1",
			BlockNumber: 100, BlockHash: paddedEmpty, TransactionHash: "0xaaa", LogIndex: 0,
		}))

		num, hash, err := store.GetLatestContractEventBlockHashAndNumber(contractAddress, blockchainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), num)
		assert.Equal(t, "", hash, "padded-empty block_hash must trim to \"\" so the reconciler guard fires")
	})

	t.Run("previous padded-empty hash trims to empty string", func(t *testing.T) {
		// A newer row sits above the padded-empty one so the "below" query returns the legacy row.
		require.NoError(t, store.StoreContractEvent(core.BlockchainEvent{
			ContractAddress: contractAddress, BlockchainID: blockchainID, Name: "E2",
			BlockNumber: 200, BlockHash: realHash, TransactionHash: "0xbbb", LogIndex: 0,
		}))

		num, hash, err := store.GetPreviousDistinctBlockHash(contractAddress, blockchainID, 200)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), num)
		assert.Equal(t, "", hash, "padded-empty block_hash must trim to \"\" mid-walk")
	})

	t.Run("real hash is preserved (not over-trimmed)", func(t *testing.T) {
		num, hash, err := store.GetLatestContractEventBlockHashAndNumber(contractAddress, blockchainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(200), num)
		assert.Equal(t, realHash, hash)
	})
}

func TestIsContractEventProcessed(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	store := NewDBStore(db)

	// Store a known event
	ev := core.BlockchainEvent{
		ContractAddress: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BlockchainID:    1,
		Name:            "TestEvent",
		BlockNumber:     500,
		TransactionHash: "0xAbCdEf1234567890AbCdEf1234567890AbCdEf1234567890AbCdEf1234567890",
		LogIndex:        3,
	}
	require.NoError(t, store.StoreContractEvent(ev))

	t.Run("existing event returns true", func(t *testing.T) {
		present, err := store.IsContractEventProcessed(ev.TransactionHash, 3, 1)
		require.NoError(t, err)
		assert.True(t, present)
	})

	t.Run("case-insensitive txHash match", func(t *testing.T) {
		// Query with uppercase — stored value was lowercased by StoreContractEvent
		present, err := store.IsContractEventProcessed("0xABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890", 3, 1)
		require.NoError(t, err)
		assert.True(t, present)
	})

	t.Run("wrong log index returns false", func(t *testing.T) {
		present, err := store.IsContractEventProcessed(ev.TransactionHash, 4, 1)
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("wrong blockchain returns false", func(t *testing.T) {
		present, err := store.IsContractEventProcessed(ev.TransactionHash, 3, 2)
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("wrong txHash returns false", func(t *testing.T) {
		present, err := store.IsContractEventProcessed("0x0000000000000000000000000000000000000000000000000000000000000000", 3, 1)
		require.NoError(t, err)
		assert.False(t, present)
	})
}
