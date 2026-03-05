package database

import (
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

func TestGetLatestEvent(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	store := NewDBStore(db)

	contractAddress := "0x1234567890123456789012345678901234567890"
	networkID := uint64(1)

	t.Run("no events returns empty event", func(t *testing.T) {
		event, err := store.GetLatestEvent(contractAddress, networkID)
		require.NoError(t, err)
		assert.Equal(t, core.BlockchainEvent{}, event)
	})

	t.Run("returns latest event", func(t *testing.T) {
		// Store multiple events
		events := []core.BlockchainEvent{
			{
				ContractAddress: contractAddress,
				BlockchainID:    networkID,
				Name:            "Event1",
				BlockNumber:     100,
				TransactionHash: "0xaaa",
				LogIndex:        1,
			},
			{
				ContractAddress: contractAddress,
				BlockchainID:    networkID,
				Name:            "Event2",
				BlockNumber:     100,
				TransactionHash: "0xbbb",
				LogIndex:        2,
			},
			{
				ContractAddress: contractAddress,
				BlockchainID:    networkID,
				Name:            "Event3",
				BlockNumber:     150,
				TransactionHash: "0xccc",
				LogIndex:        0,
			},
		}

		for _, ev := range events {
			err := store.StoreContractEvent(ev)
			require.NoError(t, err)
		}

		// Get latest event
		latestEvent, err := store.GetLatestEvent(contractAddress, networkID)
		require.NoError(t, err)

		// Should return the event with highest block number
		assert.Equal(t, uint64(150), latestEvent.BlockNumber)
		assert.Equal(t, uint32(0), latestEvent.LogIndex)
		assert.Equal(t, "Event3", latestEvent.Name)
		assert.Equal(t, contractAddress, latestEvent.ContractAddress)
		assert.Equal(t, networkID, latestEvent.BlockchainID)
	})

	t.Run("different contract returns empty event", func(t *testing.T) {
		differentContract := "0x9999999999999999999999999999999999999999"
		event, err := store.GetLatestEvent(differentContract, networkID)
		require.NoError(t, err)
		assert.Equal(t, core.BlockchainEvent{}, event)
	})

	t.Run("different network returns empty event", func(t *testing.T) {
		differentNetwork := uint64(999)
		event, err := store.GetLatestEvent(contractAddress, differentNetwork)
		require.NoError(t, err)
		assert.Equal(t, core.BlockchainEvent{}, event)
	})

	t.Run("returns highest log index when same block", func(t *testing.T) {
		contractAddr := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		chainID := uint64(42)

		// Store events in same block with different log indices
		events := []core.BlockchainEvent{
			{
				ContractAddress: contractAddr,
				BlockchainID:    chainID,
				Name:            "EventA",
				BlockNumber:     200,
				TransactionHash: "0x111",
				LogIndex:        5,
			},
			{
				ContractAddress: contractAddr,
				BlockchainID:    chainID,
				Name:            "EventB",
				BlockNumber:     200,
				TransactionHash: "0x222",
				LogIndex:        10,
			},
			{
				ContractAddress: contractAddr,
				BlockchainID:    chainID,
				Name:            "EventC",
				BlockNumber:     200,
				TransactionHash: "0x333",
				LogIndex:        3,
			},
		}

		for _, ev := range events {
			err := store.StoreContractEvent(ev)
			require.NoError(t, err)
		}

		// Get latest event - should return highest log index for the block
		latestEvent, err := store.GetLatestEvent(contractAddr, chainID)
		require.NoError(t, err)

		assert.Equal(t, uint64(200), latestEvent.BlockNumber)
		assert.Equal(t, uint32(10), latestEvent.LogIndex)
		assert.Equal(t, "EventB", latestEvent.Name)
	})
}
