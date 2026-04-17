package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannel_TableName(t *testing.T) {
	channel := Channel{}
	assert.Equal(t, "channels", channel.TableName())
}

func TestDBStore_CreateChannel(t *testing.T) {
	t.Run("Success - Create home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xhomechannel123").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, "0xhomechannel123", dbChannel.ChannelID)
		assert.Equal(t, "0xuser123", dbChannel.UserWallet)
		assert.Equal(t, core.ChannelTypeHome, dbChannel.Type)
		assert.Equal(t, uint64(1), dbChannel.BlockchainID)
		assert.Equal(t, "0xtoken123", dbChannel.Token)
		assert.Equal(t, uint32(86400), dbChannel.ChallengeDuration)
		assert.Equal(t, uint64(1), dbChannel.Nonce)
		assert.Equal(t, core.ChannelStatusOpen, dbChannel.Status)
		assert.Equal(t, uint64(0), dbChannel.StateVersion)
		assert.False(t, dbChannel.CreatedAt.IsZero())
		assert.False(t, dbChannel.UpdatedAt.IsZero())
	})

	t.Run("Success - Create escrow channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xescrowchannel456",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xescrowchannel456").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, core.ChannelTypeEscrow, dbChannel.Type)
		assert.Equal(t, uint64(137), dbChannel.BlockchainID)
	})

	t.Run("Error - Duplicate channel ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xchannel789",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Try to create again with same ID
		err = store.CreateChannel(channel)
		assert.Error(t, err)
	})

	t.Run("Success - Create channel with challenge expiration", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour)
		channel := core.Channel{
			ChannelID:          "0xchannel999",
			UserWallet:         "0xuser123",
			Asset:              "usdc",
			Type:               core.ChannelTypeHome,
			BlockchainID:       1,
			TokenAddress:       "0xtoken123",
			ChallengeDuration:  86400,
			ChallengeExpiresAt: &expiresAt,
			Nonce:              1,
			Status:             core.ChannelStatusChallenged,
			StateVersion:       1,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xchannel999").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, core.ChannelStatusChallenged, dbChannel.Status)
		assert.NotNil(t, dbChannel.ChallengeExpiresAt)
	})
}

func TestDBStore_GetChannelByID(t *testing.T) {
	t.Run("Success - Get existing channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		require.NoError(t, store.CreateChannel(channel))

		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "0xhomechannel123", result.ChannelID)
		assert.Equal(t, "0xuser123", result.UserWallet)
		assert.Equal(t, core.ChannelTypeHome, result.Type)
		assert.Equal(t, uint64(1), result.BlockchainID)
		assert.Equal(t, "0xtoken123", result.TokenAddress)
	})

	t.Run("No channel found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetChannelByID("0xnonexistent")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_GetActiveHomeChannel(t *testing.T) {
	t.Run("Success - Get active home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the home channel
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

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, homeChannelID, result.ChannelID)
		assert.Equal(t, core.ChannelTypeHome, result.Type)
		assert.Equal(t, core.ChannelStatusOpen, result.Status)
	})

	t.Run("No active home channel - user has no state", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetActiveHomeChannel("0xnonexistent", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - channel is closed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create closed channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the closed channel
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

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - channel is escrow type", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xescrowchannel123"

		// Create escrow channel (not home)
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the escrow channel as home
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

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - state has no home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create state without home channel
		state := core.State{
			ID:         "state1",
			Asset:      "USDC",
			UserWallet: "0xuser123",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}
		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_CheckOpenChannel(t *testing.T) {
	t.Run("Success - Has open channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create open home channel
		channel := core.Channel{
			ChannelID:             homeChannelID,
			UserWallet:            "0xuser123",
			Asset:                 "usdc",
			Type:                  core.ChannelTypeHome,
			BlockchainID:          1,
			TokenAddress:          "0xtoken123",
			ChallengeDuration:     86400,
			Nonce:                 1,
			ApprovedSigValidators: "0x2",
			Status:                core.ChannelStatusOpen,
			StateVersion:          0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the channel
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

		approvedSigValidators, hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.True(t, hasOpenChannel)
		assert.Equal(t, "0x2", approvedSigValidators)
	})

	t.Run("No open channel - user not found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		approvedSigValidators, hasOpenChannel, err := store.CheckOpenChannel("0xnonexistent", "USDC")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
		assert.Equal(t, "", approvedSigValidators)
	})

	t.Run("No open channel - channel is closed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create closed channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the closed channel
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

		approvedSigValidators, hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
		assert.Equal(t, "", approvedSigValidators)
	})

	t.Run("No open channel - wrong asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create open home channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state for USDC
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

		// Check for different asset
		approvedSigValidators, hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "ETH")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
		assert.Equal(t, "", approvedSigValidators)
	})
}

func TestDBStore_UpdateChannel(t *testing.T) {
	t.Run("Success - Update channel status and version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update channel
		channel.Status = core.ChannelStatusClosed
		channel.StateVersion = 5

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, core.ChannelStatusClosed, result.Status)
		assert.Equal(t, uint64(5), result.StateVersion)
	})

	t.Run("Success - Update challenge expiration", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update with challenge expiration
		expiresAt := time.Now().Add(24 * time.Hour)
		channel.Status = core.ChannelStatusChallenged
		channel.ChallengeExpiresAt = &expiresAt

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, core.ChannelStatusChallenged, result.Status)
		assert.NotNil(t, result.ChallengeExpiresAt)
	})

	t.Run("Success - Update blockchain and token", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update blockchain and token
		channel.BlockchainID = 137
		channel.TokenAddress = "0xnewtoken456"

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, uint64(137), result.BlockchainID)
		assert.Equal(t, "0xnewtoken456", result.TokenAddress)
	})

	t.Run("Error - Update non-existent channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xnonexistent",
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}

		err := store.UpdateChannel(channel)
		require.NoError(t, err) // GORM doesn't return error for update with no rows affected
	})
}

func TestDBStore_GetUserChannels(t *testing.T) {
	t.Run("Success - Get all channels for user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		ch1 := core.Channel{
			ChannelID:         "0xchannel_a",
			UserWallet:        "0xuser_gc",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken1",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      1,
		}
		ch2 := core.Channel{
			ChannelID:         "0xchannel_b",
			UserWallet:        "0xuser_gc",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken1",
			ChallengeDuration: 86400,
			Nonce:             2,
			Status:            core.ChannelStatusClosed,
			StateVersion:      3,
		}
		require.NoError(t, store.CreateChannel(ch1))
		require.NoError(t, store.CreateChannel(ch2))

		channels, total, err := store.GetUserChannels("0xuser_gc", nil, nil, nil, 100, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 2)
		assert.Equal(t, uint32(2), total)
	})

	t.Run("Success - Filter by status", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_open", UserWallet: "0xuser_sf",
			Asset: "usdc", Type: core.ChannelTypeHome, BlockchainID: 1,
			TokenAddress: "0xt", ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen, StateVersion: 0,
		}))
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_closed", UserWallet: "0xuser_sf",
			Asset: "usdc", Type: core.ChannelTypeHome, BlockchainID: 1,
			TokenAddress: "0xt", ChallengeDuration: 86400, Nonce: 2,
			Status: core.ChannelStatusClosed, StateVersion: 1,
		}))

		status := core.ChannelStatusClosed
		channels, total, err := store.GetUserChannels("0xuser_sf", &status, nil, nil, 100, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, uint32(1), total)
		assert.Equal(t, core.ChannelStatusClosed, channels[0].Status)
	})

	t.Run("Success - Filter by asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_usdc", UserWallet: "0xuser_af",
			Asset: "usdc", Type: core.ChannelTypeHome, BlockchainID: 1,
			TokenAddress: "0xt1", ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen, StateVersion: 0,
		}))
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_weth", UserWallet: "0xuser_af",
			Asset: "weth", Type: core.ChannelTypeHome, BlockchainID: 1,
			TokenAddress: "0xt2", ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen, StateVersion: 0,
		}))

		asset := "usdc"
		channels, total, err := store.GetUserChannels("0xuser_af", nil, &asset, nil, 100, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, uint32(1), total)
		assert.Equal(t, "usdc", channels[0].Asset)
	})

	t.Run("Success - Filter by channel type", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_home", UserWallet: "0xuser_tf",
			Asset: "usdc", Type: core.ChannelTypeHome, BlockchainID: 1,
			TokenAddress: "0xt", ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen, StateVersion: 0,
		}))
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xch_escrow", UserWallet: "0xuser_tf",
			Asset: "usdc", Type: core.ChannelTypeEscrow, BlockchainID: 1,
			TokenAddress: "0xt", ChallengeDuration: 86400, Nonce: 2,
			Status: core.ChannelStatusOpen, StateVersion: 0,
		}))

		homeType := core.ChannelTypeHome
		channels, total, err := store.GetUserChannels("0xuser_tf", nil, nil, &homeType, 100, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, uint32(1), total)
		assert.Equal(t, core.ChannelTypeHome, channels[0].Type)
	})

	t.Run("Success - Pagination limits results", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		for i := 0; i < 5; i++ {
			require.NoError(t, store.CreateChannel(core.Channel{
				ChannelID: fmt.Sprintf("0xch_pg_%d", i), UserWallet: "0xuser_pg",
				Asset: "usdc", Type: core.ChannelTypeHome, BlockchainID: 1,
				TokenAddress: "0xt", ChallengeDuration: 86400, Nonce: uint64(i),
				Status: core.ChannelStatusOpen, StateVersion: 0,
			}))
		}

		channels, total, err := store.GetUserChannels("0xuser_pg", nil, nil, nil, 2, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 2)
		assert.Equal(t, uint32(5), total)
	})

	t.Run("Success - Empty result for unknown user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channels, total, err := store.GetUserChannels("0xnonexistent", nil, nil, nil, 100, 0)
		require.NoError(t, err)
		assert.Len(t, channels, 0)
		assert.Equal(t, uint32(0), total)
	})
}

func TestDBStore_GetChannelsCountByLabels(t *testing.T) {
	t.Run("no channels returns empty", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		results, err := store.GetChannelsCountByLabels()
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("counts grouped by asset and status", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		for i := 0; i < 3; i++ {
			require.NoError(t, store.CreateChannel(core.Channel{
				ChannelID: fmt.Sprintf("0xchannel_open_usdc_%d", i), UserWallet: "0xuser1", Asset: "usdc",
				Type: core.ChannelTypeHome, BlockchainID: 1, ChallengeDuration: 86400, Nonce: uint64(i + 1),
				Status: core.ChannelStatusOpen,
			}))
		}
		for i := 0; i < 2; i++ {
			require.NoError(t, store.CreateChannel(core.Channel{
				ChannelID: fmt.Sprintf("0xchannel_closed_usdc_%d", i), UserWallet: "0xuser1", Asset: "usdc",
				Type: core.ChannelTypeHome, BlockchainID: 1, ChallengeDuration: 86400, Nonce: uint64(i + 10),
				Status: core.ChannelStatusClosed,
			}))
		}
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xchannel_open_eth_0", UserWallet: "0xuser2", Asset: "eth",
			Type: core.ChannelTypeHome, BlockchainID: 1, ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen,
		}))

		results, err := store.GetChannelsCountByLabels()
		require.NoError(t, err)

		countMap := make(map[string]uint64)
		for _, r := range results {
			countMap[r.Asset+"/"+r.Status.String()] = r.Count
		}

		assert.Equal(t, uint64(3), countMap["usdc/"+core.ChannelStatusOpen.String()])
		assert.Equal(t, uint64(2), countMap["usdc/"+core.ChannelStatusClosed.String()])
		assert.Equal(t, uint64(1), countMap["eth/"+core.ChannelStatusOpen.String()])
	})

	t.Run("reflects status transitions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xchannel1", UserWallet: "0xuser1", Asset: "usdc",
			Type: core.ChannelTypeHome, BlockchainID: 1, ChallengeDuration: 86400, Nonce: 1,
			Status: core.ChannelStatusOpen,
		}))
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID: "0xchannel2", UserWallet: "0xuser1", Asset: "usdc",
			Type: core.ChannelTypeHome, BlockchainID: 1, ChallengeDuration: 86400, Nonce: 2,
			Status: core.ChannelStatusOpen,
		}))

		// Transition one channel to closed
		ch1, err := store.GetChannelByID("0xchannel1")
		require.NoError(t, err)
		ch1.Status = core.ChannelStatusClosed
		require.NoError(t, store.UpdateChannel(*ch1))

		results, err := store.GetChannelsCountByLabels()
		require.NoError(t, err)

		countMap := make(map[string]uint64)
		for _, r := range results {
			countMap[r.Asset+"/"+r.Status.String()] = r.Count
		}

		assert.Equal(t, uint64(1), countMap["usdc/"+core.ChannelStatusOpen.String()])
		assert.Equal(t, uint64(1), countMap["usdc/"+core.ChannelStatusClosed.String()])
	})
}
