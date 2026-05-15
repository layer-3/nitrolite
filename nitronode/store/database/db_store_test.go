package database

import (
	"testing"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBStore_GetUserBalances(t *testing.T) {
	t.Run("Success - Get balances for single asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create state with balance
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

		// Lock user state before storing (ensures row exists in user_balances)
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		balances, err := store.GetUserBalances("0xuser123")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.Equal(t, "USDC", balances[0].Asset)
		assert.Equal(t, decimal.NewFromInt(1000), balances[0].Balance)
	})

	t.Run("Success - Get balances for multiple assets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create state for USDC
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

		// Create state for ETH
		state2 := core.State{
			ID:            "state2",
			Asset:         "ETH",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(5),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		// Lock user states before storing (ensures rows exist in user_balances)
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)
		_, err = store.LockUserState("0xuser123", "ETH")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state1, ""))
		require.NoError(t, store.StoreUserState(state2, ""))

		balances, err := store.GetUserBalances("0xuser123")
		require.NoError(t, err)

		assert.Len(t, balances, 2)

		// Find balances by asset
		var usdcBalance, ethBalance *core.BalanceEntry
		for i := range balances {
			switch balances[i].Asset {
			case "USDC":
				usdcBalance = &balances[i]
			case "ETH":
				ethBalance = &balances[i]
			}
		}

		require.NotNil(t, usdcBalance)
		require.NotNil(t, ethBalance)
		assert.Equal(t, decimal.NewFromInt(1000), usdcBalance.Balance)
		assert.Equal(t, decimal.NewFromInt(5), ethBalance.Balance)
	})

	t.Run("Success - Returns latest version for each asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create multiple versions for same asset
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

		// Lock user state before storing (ensures row exists in user_balances)
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state1, ""))
		require.NoError(t, store.StoreUserState(state2, ""))

		balances, err := store.GetUserBalances("0xuser123")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.Equal(t, "USDC", balances[0].Asset)
		assert.Equal(t, decimal.NewFromInt(2000), balances[0].Balance) // Latest version
	})

	t.Run("Success - Returns latest epoch for each asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

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
				UserBalance: decimal.NewFromInt(3000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		// Lock user state before storing (ensures row exists in user_balances)
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state1, ""))
		require.NoError(t, store.StoreUserState(state2, ""))

		balances, err := store.GetUserBalances("0xuser123")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.Equal(t, decimal.NewFromInt(3000), balances[0].Balance) // Latest epoch
	})

	t.Run("No balances found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		balances, err := store.GetUserBalances("0xnonexistent")
		require.NoError(t, err)
		assert.Empty(t, balances)
	})
}

func TestDBStore_EnsureNoOngoingStateTransitions(t *testing.T) {
	t.Run("No previous state - No error", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		err := store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	t.Run("HomeDeposit - Versions match", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create channel
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
			StateVersion:      1, // Matches state version
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create signed state
		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeHomeDeposit,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	t.Run("HomeDeposit - Versions mismatch", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create channel with different version
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
			StateVersion:      0, // Doesn't match state version
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create signed state
		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeHomeDeposit,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "home deposit is still ongoing")
	})

	t.Run("MutualLock - Both versions match", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create home channel
		homeChannel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      2, // Matches state version
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create escrow channel
		escrowChannel := core.Channel{
			ChannelID:         escrowChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      2, // Matches state version
		}
		require.NoError(t, store.CreateChannel(escrowChannel))

		// Create signed state
		state := core.State{
			ID:              "state1",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         2,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeMutualLock,
			},
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
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	t.Run("MutualLock - Home channel version mismatch", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create home channel with mismatched version
		homeChannel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      1, // Doesn't match state version
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Create escrow channel
		escrowChannel := core.Channel{
			ChannelID:         escrowChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      2,
		}
		require.NoError(t, store.CreateChannel(escrowChannel))

		// Create signed state
		state := core.State{
			ID:              "state1",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         2,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeMutualLock,
			},
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
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mutual lock is still ongoing")
	})

	t.Run("EscrowWithdraw - Versions match", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		escrowChannelID := "0xescrowchannel456"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create escrow channel
		escrowChannel := core.Channel{
			ChannelID:         escrowChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      4, // Matches state version
		}
		require.NoError(t, store.CreateChannel(escrowChannel))

		// Create signed state
		state := core.State{
			ID:              "state1",
			Asset:           "USDC",
			UserWallet:      "0xuser123",
			Epoch:           1,
			Version:         4,
			HomeChannelID:   &homeChannelID,
			EscrowChannelID: &escrowChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeEscrowWithdraw,
			},
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
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	t.Run("Unsigned state - No validation", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create channel with mismatched version
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

		// Create unsigned state
		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeHomeDeposit,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			// No signatures
		}

		// Lock user state before storing
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state, ""))

		// Should not error because there's no signed state
		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	const wallet = "0xuser123"
	const asset = "USDC"
	const homeChannelID = "0xhomechannel123"
	const escrowChannelID = "0xescrowchannel456"
	userSig := "0xusersig"
	nodeSig := "0xnodesig"

	newHomeChannel := func(version uint64) core.Channel {
		return core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        wallet,
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      version,
		}
	}

	newEscrowChannel := func(version uint64) core.Channel {
		return core.Channel{
			ChannelID:         escrowChannelID,
			UserWallet:        wallet,
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      version,
		}
	}

	newSignedState := func(version uint64, transitionType core.TransitionType, withEscrow bool) core.State {
		hc := homeChannelID
		state := core.State{
			ID:            "state1",
			Asset:         asset,
			UserWallet:    wallet,
			Epoch:         1,
			Version:       version,
			HomeChannelID: &hc,
			Transition:    core.Transition{Type: transitionType},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}
		if withEscrow {
			ec := escrowChannelID
			state.EscrowChannelID = &ec
			state.EscrowLedger = &core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			}
		}
		return state
	}

	storeState := func(t *testing.T, store DatabaseStore, state core.State) {
		t.Helper()
		_, err := store.LockUserState(wallet, asset)
		require.NoError(t, err)
		require.NoError(t, store.StoreUserState(state, ""))
	}

	t.Run("HomeDeposit - home channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		storeState(t, store, newSignedState(1, core.TransitionTypeHomeDeposit, false))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "home deposit is still ongoing")
	})

	t.Run("HomeWithdrawal - home channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		storeState(t, store, newSignedState(1, core.TransitionTypeHomeWithdrawal, false))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "home withdrawal is still ongoing")
	})

	t.Run("MutualLock - home channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(newEscrowChannel(2)))
		storeState(t, store, newSignedState(2, core.TransitionTypeMutualLock, true))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutual lock is still ongoing")
	})

	t.Run("MutualLock - escrow channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(newHomeChannel(2)))
		storeState(t, store, newSignedState(2, core.TransitionTypeMutualLock, true))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutual lock is still ongoing")
	})

	t.Run("EscrowLock - escrow channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(newHomeChannel(1)))
		storeState(t, store, newSignedState(1, core.TransitionTypeEscrowLock, true))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow lock is still ongoing")
	})

	t.Run("EscrowWithdraw - escrow channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(newHomeChannel(4)))
		storeState(t, store, newSignedState(4, core.TransitionTypeEscrowWithdraw, true))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow withdrawal is still ongoing")
	})

	t.Run("Migrate - home channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		storeState(t, store, newSignedState(1, core.TransitionTypeMigrate, false))

		err := store.EnsureNoOngoingStateTransitions(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "home chain migration is still ongoing")
	})
}

func TestDBStore_UpdateStateSigsIfMissing(t *testing.T) {
	t.Run("Backfills user_sig when null and unblocks gate", func(t *testing.T) {
		// This is the wedge-recovery path: a node-only state was checkpointed on chain
		// (e.g. recipient submitted a transfer_receive state directly). After the event
		// reactor backfills user_sig, EnsureNoOngoingStateTransitions must see a fully
		// signed row at the on-chain version and pass.
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		nodeSig := "0xnodesig"

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
			StateVersion:      2,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Node-only state at version 2; gate would normally skip this row and find
		// nothing else, returning nil. To exercise the wedge, also seed an older
		// bilateral state at version 1.
		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		bilateralUserSig := "0xprior"
		bilateralNodeSig := "0xpriornode"
		bilateral := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeHomeDeposit,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &bilateralUserSig,
			NodeSig: &bilateralNodeSig,
		}
		require.NoError(t, store.StoreUserState(bilateral, ""))

		nodeOnly := core.State{
			ID:            "state2",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       2,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeTransferReceive,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(750),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			NodeSig: &nodeSig,
		}
		require.NoError(t, store.StoreUserState(nodeOnly, ""))

		// Pre-backfill: gate sees bilateral row at version 1, channel.state_version is 2 → mismatch.
		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.Error(t, err)

		// Backfill the user signature recovered from the on-chain event.
		recoveredSig := "0xrecovered"
		require.NoError(t, store.UpdateStateSigsIfMissing(homeChannelID, 2, recoveredSig, ""))

		got, err := store.GetStateByID("state2")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.UserSig)
		assert.Equal(t, recoveredSig, *got.UserSig)

		// Post-backfill: gate sees the now-bilateral row at version 2, matches channel state_version.
		err = store.EnsureNoOngoingStateTransitions("0xuser123", "USDC")
		require.NoError(t, err)
	})

	t.Run("Idempotent on replay - existing sig preserved", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xexisting"
		nodeSig := "0xnodesig"

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
			StateVersion:      1,
		}
		require.NoError(t, store.CreateChannel(channel))

		_, err := store.LockUserState("0xuser123", "USDC")
		require.NoError(t, err)

		state := core.State{
			ID:            "state1",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition: core.Transition{
				Type: core.TransitionTypeHomeDeposit,
			},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
			NodeSig: &nodeSig,
		}
		require.NoError(t, store.StoreUserState(state, ""))

		// Replayed event would carry a different (or any) sig; existing one must not be overwritten.
		require.NoError(t, store.UpdateStateSigsIfMissing(homeChannelID, 1, "0xshould-not-overwrite", "0xshould-not-overwrite-node"))

		got, err := store.GetStateByID("state1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.UserSig)
		assert.Equal(t, userSig, *got.UserSig)
	})

	t.Run("Empty sig is no-op", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		require.NoError(t, store.UpdateStateSigsIfMissing(homeChannelID, 1, "", ""))
	})

	t.Run("Unknown version returns no error", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		require.NoError(t, store.UpdateStateSigsIfMissing(homeChannelID, 99, "0xanything", "0xanything-node"))
	})

	t.Run("Backfills node_sig when user_sig already present", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      1,
		}))

		userSig := "0xusersigonly"
		state := core.State{
			ID:            "state-user-only",
			Asset:         "USDC",
			UserWallet:    "0xuser123",
			Epoch:         1,
			Version:       1,
			HomeChannelID: &homeChannelID,
			Transition:    core.Transition{Type: core.TransitionTypeHomeDeposit},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(100),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: &userSig,
		}
		require.NoError(t, store.StoreUserState(state, ""))

		require.NoError(t, store.UpdateStateSigsIfMissing(homeChannelID, 1, "0xshould-not-overwrite-user", "0xrecoverednode"))

		got, err := store.GetStateByID("state-user-only")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.UserSig)
		assert.Equal(t, userSig, *got.UserSig, "existing user_sig must be preserved")
		require.NotNil(t, got.NodeSig)
		assert.Equal(t, "0xrecoverednode", *got.NodeSig)
	})
}

func TestDBStore_SumUnsignedReceiverStateAmountsAfterVersion(t *testing.T) {
	const wallet = "0xuser123"
	const asset = "USDC"
	const homeChannelID = "0xhomechannel123"

	setupChannel := func(t *testing.T, store DatabaseStore) {
		t.Helper()
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        wallet,
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusChallenged,
			StateVersion:      5,
		}))
	}

	storeState := func(t *testing.T, store DatabaseStore, version uint64, transitionType core.TransitionType, amount decimal.Decimal, hasNodeSig bool) {
		t.Helper()
		channelIDLocal := homeChannelID
		s := core.State{
			ID:            core.GetStateID(wallet, asset, 1, version),
			Asset:         asset,
			UserWallet:    wallet,
			Epoch:         1,
			Version:       version,
			HomeChannelID: &channelIDLocal,
			Transition: core.Transition{
				Type:      transitionType,
				TxID:      "0xtx",
				AccountID: "0xacct",
				Amount:    amount,
			},
			HomeLedger: core.Ledger{
				UserBalance: amount,
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}
		userSig := "0xusersig"
		s.UserSig = &userSig
		if hasNodeSig {
			nodeSig := "0xnodesig"
			s.NodeSig = &nodeSig
		}
		require.NoError(t, store.StoreUserState(s, ""))
	}

	t.Run("Sums only unsigned receiver states strictly above version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)
		setupChannel(t, store)

		// Below the cutoff — must be ignored.
		storeState(t, store, 5, core.TransitionTypeTransferReceive, decimal.NewFromInt(10), false)
		// At the cutoff — must be ignored (strictly greater than minVersion).
		storeState(t, store, 6, core.TransitionTypeTransferReceive, decimal.NewFromInt(20), false)
		// Above cutoff, unsigned receive — included.
		storeState(t, store, 7, core.TransitionTypeTransferReceive, decimal.NewFromInt(30), false)
		// Above cutoff, unsigned release — included.
		storeState(t, store, 8, core.TransitionTypeRelease, decimal.NewFromInt(40), false)
		// Above cutoff, but already node-signed — must be ignored.
		storeState(t, store, 9, core.TransitionTypeTransferReceive, decimal.NewFromInt(100), true)
		// Above cutoff, unsigned but wrong transition type — must be ignored.
		storeState(t, store, 10, core.TransitionTypeHomeDeposit, decimal.NewFromInt(500), false)

		total, err := store.SumUnsignedReceiverStateAmountsAfterVersion(homeChannelID, 6)
		require.NoError(t, err)
		assert.True(t, decimal.NewFromInt(70).Equal(total), "want 70, got %s", total.String())
	})

	t.Run("Returns zero when no matches", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)
		setupChannel(t, store)

		total, err := store.SumUnsignedReceiverStateAmountsAfterVersion(homeChannelID, 0)
		require.NoError(t, err)
		assert.True(t, total.IsZero())
	})
}

func TestDBStore_HasSignedFinalize(t *testing.T) {
	const wallet = "0xuser123"
	const asset = "USDC"
	const homeChannelID = "0xhomechannel123"

	setupChannel := func(t *testing.T, store DatabaseStore) {
		t.Helper()
		require.NoError(t, store.CreateChannel(core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        wallet,
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusChallenged,
			StateVersion:      5,
		}))
	}

	storeState := func(t *testing.T, store DatabaseStore, version uint64, transitionType core.TransitionType, hasNodeSig bool) {
		t.Helper()
		channelIDLocal := homeChannelID
		s := core.State{
			ID:            core.GetStateID(wallet, asset, 1, version),
			Asset:         asset,
			UserWallet:    wallet,
			Epoch:         1,
			Version:       version,
			HomeChannelID: &channelIDLocal,
			Transition: core.Transition{
				Type:      transitionType,
				TxID:      "0xtx",
				AccountID: "0xacct",
				Amount:    decimal.Zero,
			},
		}
		userSig := "0xusersig"
		s.UserSig = &userSig
		if hasNodeSig {
			nodeSig := "0xnodesig"
			s.NodeSig = &nodeSig
		}
		require.NoError(t, store.StoreUserState(s, ""))
	}

	t.Run("True when node-signed Finalize exists", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)
		setupChannel(t, store)

		storeState(t, store, 6, core.TransitionTypeTransferReceive, true)
		storeState(t, store, 7, core.TransitionTypeFinalize, true)

		got, err := store.HasSignedFinalize(homeChannelID)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("False when no Finalize state exists", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)
		setupChannel(t, store)

		storeState(t, store, 6, core.TransitionTypeTransferReceive, true)
		storeState(t, store, 7, core.TransitionTypeTransferReceive, true)

		got, err := store.HasSignedFinalize(homeChannelID)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("False when Finalize exists without node sig", func(t *testing.T) {
		// Pins the documented invariant: a Finalize row whose node_sig is NULL must not
		// be treated as node-signed, even though SubmitState today never stores one in
		// that shape. Guards against a regression that drops the node_sig predicate or
		// a future path that writes Finalize rows ahead of the node signature.
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)
		setupChannel(t, store)

		storeState(t, store, 7, core.TransitionTypeFinalize, false)

		got, err := store.HasSignedFinalize(homeChannelID)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestDBStore_EnsureNoOngoingEscrowOperation(t *testing.T) {
	const wallet = "0xuser123"
	const asset = "USDC"
	const homeChannelID = "0xhomechannel123"
	const escrowChannelID = "0xescrowchannel456"
	const userSig = "0xusersig"
	const nodeSig = "0xnodesig"

	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        wallet,
		Asset:             "usdc",
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xtoken123",
		ChallengeDuration: 86400,
		Nonce:             1,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	newEscrowChannel := func(version uint64) core.Channel {
		return core.Channel{
			ChannelID:         escrowChannelID,
			UserWallet:        wallet,
			Asset:             "usdc",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      version,
		}
	}

	newSignedState := func(version uint64, transitionType core.TransitionType, withEscrow bool) core.State {
		state := core.State{
			ID:            "state1",
			Asset:         asset,
			UserWallet:    wallet,
			Epoch:         1,
			Version:       version,
			HomeChannelID: ptr(homeChannelID),
			Transition:    core.Transition{Type: transitionType},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			UserSig: ptr(userSig),
			NodeSig: ptr(nodeSig),
		}
		if withEscrow {
			state.EscrowChannelID = ptr(escrowChannelID)
			state.EscrowLedger = &core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			}
		}
		return state
	}

	storeState := func(t *testing.T, store DatabaseStore, state core.State) {
		t.Helper()
		_, err := store.LockUserState(wallet, asset)
		require.NoError(t, err)
		require.NoError(t, store.StoreUserState(state, ""))
	}

	t.Run("No previous state - allow", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.NoError(t, err)
	})

	t.Run("Non-escrow transition (TransferSend) - allow", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))

		storeState(t, store, newSignedState(1, core.TransitionTypeTransferSend, false))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.NoError(t, err)
	})

	t.Run("EscrowLock - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(1)))

		storeState(t, store, newSignedState(1, core.TransitionTypeEscrowLock, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow lock is still ongoing")
	})

	t.Run("MutualLock - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(1)))

		storeState(t, store, newSignedState(1, core.TransitionTypeMutualLock, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutual lock is still ongoing")
	})

	t.Run("EscrowDeposit - chain caught up - allow", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(2)))

		storeState(t, store, newSignedState(2, core.TransitionTypeEscrowDeposit, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.NoError(t, err)
	})

	t.Run("EscrowDeposit - chain not synced - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(1)))

		storeState(t, store, newSignedState(2, core.TransitionTypeEscrowDeposit, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow deposit finalization is still ongoing")
	})

	t.Run("EscrowWithdraw - chain caught up - allow", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(3)))

		storeState(t, store, newSignedState(3, core.TransitionTypeEscrowWithdraw, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.NoError(t, err)
	})

	t.Run("EscrowWithdraw - chain not synced - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(2)))

		storeState(t, store, newSignedState(3, core.TransitionTypeEscrowWithdraw, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow withdrawal finalization is still ongoing")
	})

	t.Run("EscrowDeposit - escrow channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))

		storeState(t, store, newSignedState(2, core.TransitionTypeEscrowDeposit, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow deposit finalization is still ongoing")
	})

	t.Run("EscrowWithdraw - escrow channel missing from DB - block", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))

		storeState(t, store, newSignedState(3, core.TransitionTypeEscrowWithdraw, true))

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escrow withdrawal finalization is still ongoing")
	})

	t.Run("Unsigned state - ignored", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)
		require.NoError(t, store.CreateChannel(homeChannel))
		require.NoError(t, store.CreateChannel(newEscrowChannel(1)))

		state := newSignedState(2, core.TransitionTypeEscrowLock, true)
		state.UserSig = nil
		state.NodeSig = nil
		storeState(t, store, state)

		err := store.EnsureNoOngoingEscrowOperation(wallet, asset)
		require.NoError(t, err)
	})
}

func ptr[T any](v T) *T {
	return &v
}
