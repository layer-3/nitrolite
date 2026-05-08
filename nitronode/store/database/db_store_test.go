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
}

func TestDBStore_UpdateStateUserSigIfMissing(t *testing.T) {
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
		require.NoError(t, store.UpdateStateUserSigIfMissing(homeChannelID, 2, recoveredSig))

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
		require.NoError(t, store.UpdateStateUserSigIfMissing(homeChannelID, 1, "0xshould-not-overwrite"))

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
		require.NoError(t, store.UpdateStateUserSigIfMissing(homeChannelID, 1, ""))
	})

	t.Run("Unknown version returns no error", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		require.NoError(t, store.UpdateStateUserSigIfMissing(homeChannelID, 99, "0xanything"))
	})
}
