package database

import (
	"testing"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_TableName(t *testing.T) {
	state := State{}
	assert.Equal(t, "channel_states", state.TableName())
}

func TestDBStore_StoreUserState(t *testing.T) {
	t.Run("Success - Store new state", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		state := core.State{
			ID:            "state123",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type:      core.TransitionTypeHomeDeposit,
				AccountID: homeChannelID,
				Amount:    decimal.NewFromInt(1000),
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.NewFromInt(1000),
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		err := store.StoreUserState(state)
		require.NoError(t, err)

		// Verify state was stored
		var dbState State
		err = db.Where("id = ?", "state123").First(&dbState).Error
		require.NoError(t, err)

		assert.Equal(t, "state123", dbState.ID)
		assert.Equal(t, "USDC", dbState.Asset)
		assert.Equal(t, "0xuser123", dbState.UserWallet)
		assert.Equal(t, uint64(1), dbState.Epoch)
		assert.Equal(t, uint64(1), dbState.Version)
		assert.Equal(t, &homeChannelID, dbState.HomeChannelID)
		assert.True(t, dbState.HomeUserBalance.Equal(decimal.NewFromInt(1000)))
		assert.True(t, dbState.HomeUserNetFlow.Equal(decimal.NewFromInt(1000)))
		assert.NotNil(t, dbState.UserSig)
		assert.NotNil(t, dbState.NodeSig)
	})

	t.Run("Success - Store state with escrow ledger", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		state := core.State{
			ID:              "state456",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         2,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition: core.Transition{
				Type:      core.TransitionTypeEscrowDeposit,
				AccountID: escrowChannelID,
				Amount:    decimal.NewFromInt(500),
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.NewFromInt(-500),
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			EscrowLedger: &core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.NewFromInt(500),
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		err := store.StoreUserState(state)
		require.NoError(t, err)

		// Verify state was stored
		var dbState State
		err = db.Where("id = ?", "state456").First(&dbState).Error
		require.NoError(t, err)

		assert.Equal(t, "state456", dbState.ID)
		assert.Equal(t, &escrowChannelID, dbState.EscrowChannelID)
		assert.True(t, dbState.EscrowUserBalance.Equal(decimal.NewFromInt(500)))
		assert.True(t, dbState.EscrowUserNetFlow.Equal(decimal.NewFromInt(500)))
	})

	t.Run("Error - Duplicate state ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		state := core.State{
			ID:            "state789",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		err := store.StoreUserState(state)
		require.NoError(t, err)

		// Try to store again with same ID
		err = store.StoreUserState(state)
		assert.Error(t, err)
	})
}

func TestDBStore_GetLastUserState(t *testing.T) {
	t.Run("Success - Get last state unsigned", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create multiple states with different versions
		state1 := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		state2 := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       2,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(2000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))

		// Get last state
		result, err := store.GetLastUserState("0xuser123", "USDC", false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})

	t.Run("Success - Get last signed state only", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create unsigned state with higher version
		state1 := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       3,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(3000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		// Create signed state with lower version
		state2 := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       2,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(2000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))

		// Get last signed state should return state2
		result, err := store.GetLastUserState("0xuser123", "USDC", true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})

	t.Run("No state found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetLastUserState("0xnonexistent", "USDC", false)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Success - Get last state by epoch and version ordering", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create states with different epochs
		state1 := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       5,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		state2 := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         2,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(2000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))

		// Get last state - should prioritize higher epoch
		result, err := store.GetLastUserState("0xuser123", "USDC", false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(2), result.Epoch)
		assert.Equal(t, uint64(1), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})
}

func TestDBStore_GetLastStateByChannelID(t *testing.T) {
	t.Run("Success - Get by home channel ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetLastStateByChannelID(homeChannelID, false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, &homeChannelID, result.HomeChannelID)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})

	t.Run("Success - Get by escrow channel ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create escrow channel
		escrowChannel := core.Channel{
			ChannelID:    escrowChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeEscrow,
			BlockchainID: 2,
			TokenAddress: "0xtoken456",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(escrowChannel))

		state := core.State{
			ID:              "state2",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         2,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition:      core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			EscrowLedger: &core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetLastStateByChannelID(escrowChannelID, false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, &escrowChannelID, result.EscrowChannelID)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
		assert.Equal(t, uint64(2), result.EscrowLedger.BlockchainID)
		assert.Equal(t, "0xtoken456", result.EscrowLedger.TokenAddress)
	})

	t.Run("Success - Get signed state only", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Unsigned state
		state1 := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       2,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		// Signed state
		state2 := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))

		result, err := store.GetLastStateByChannelID(homeChannelID, true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(1), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})

	t.Run("No state found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetLastStateByChannelID("0xnonexistent", false)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_GetStateByChannelIDAndVersion(t *testing.T) {
	t.Run("Success - Get specific version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create multiple versions
		state1 := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		state2 := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       2,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(2000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))

		// Get version 1
		result, err := store.GetStateByChannelIDAndVersion(homeChannelID, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, uint64(1), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)

		// Get version 2
		result, err = store.GetStateByChannelIDAndVersion(homeChannelID, 2)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
	})

	t.Run("Success - Get by escrow channel ID and version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create escrow channel
		escrowChannel := core.Channel{
			ChannelID:    escrowChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeEscrow,
			BlockchainID: 2,
			TokenAddress: "0xtoken456",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(escrowChannel))

		state := core.State{
			ID:              "state1",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         5,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition:      core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			EscrowLedger: &core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetStateByChannelIDAndVersion(escrowChannelID, 5)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, uint64(5), result.Version)
		assert.Equal(t, uint64(1), result.HomeLedger.BlockchainID)
		assert.Equal(t, "0xtoken123", result.HomeLedger.TokenAddress)
		assert.Equal(t, uint64(2), result.EscrowLedger.BlockchainID)
		assert.Equal(t, "0xtoken456", result.EscrowLedger.TokenAddress)
	})

	t.Run("No state found - version not exists", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetStateByChannelIDAndVersion(homeChannelID, 999)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No state found - channel not exists", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetStateByChannelIDAndVersion("0xnonexistent", 1)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
