package database

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainAction_TableName(t *testing.T) {
	action := BlockchainAction{}
	assert.Equal(t, "blockchain_actions", action.TableName())
}

func TestDBStore_ScheduleCheckpoint(t *testing.T) {
	t.Run("Success - Schedule checkpoint action", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// First create a state to reference
		state := core.State{
			ID:         "0x1234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x1234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
			},
		}
		require.NoError(t, store.StoreUserState(state))

		err := store.ScheduleCheckpoint(state.ID, 0)
		require.NoError(t, err)

		// Verify action was created
		var action BlockchainAction
		err = db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeCheckpoint, action.Type)
		assert.Equal(t, state.ID, action.StateID)
		assert.Equal(t, BlockchainActionStatusPending, action.Status)
		assert.Equal(t, uint8(0), action.Retries)
		assert.Empty(t, action.Error)
		assert.False(t, action.CreatedAt.IsZero())
		assert.False(t, action.UpdatedAt.IsZero())
	})
}

func TestDBStore_ScheduleInitiateEscrowWithdrawal(t *testing.T) {
	t.Run("Success - Schedule initiate escrow withdrawal", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create a state
		state := core.State{
			ID:         "0x2234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x2234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
			},
		}
		require.NoError(t, store.StoreUserState(state))

		err := store.ScheduleInitiateEscrowWithdrawal(state.ID, 0)
		require.NoError(t, err)

		// Verify action was created
		var action BlockchainAction
		err = db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeInitiateEscrowWithdrawal, action.Type)
		assert.Equal(t, state.ID, action.StateID)
		assert.Equal(t, BlockchainActionStatusPending, action.Status)
	})
}

func TestDBStore_ScheduleFinalizeEscrowDeposit(t *testing.T) {
	t.Run("Success - Schedule finalize escrow deposit", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x3234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0x3234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(100),
			},
		}
		require.NoError(t, store.StoreUserState(state))

		err := store.ScheduleFinalizeEscrowDeposit(state.ID, 0)
		require.NoError(t, err)

		var action BlockchainAction
		err = db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeFinalizeEscrowDeposit, action.Type)
		assert.Equal(t, state.ID, action.StateID)
	})
}

func TestDBStore_ScheduleFinalizeEscrowWithdrawal(t *testing.T) {
	t.Run("Success - Schedule finalize escrow withdrawal", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x4234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0x4234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(200),
			},
		}
		require.NoError(t, store.StoreUserState(state))

		err := store.ScheduleFinalizeEscrowWithdrawal(state.ID, 0)
		require.NoError(t, err)

		var action BlockchainAction
		err = db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeFinalizeEscrowWithdrawal, action.Type)
		assert.Equal(t, state.ID, action.StateID)
	})
}

func TestDBStore_Fail(t *testing.T) {
	t.Run("Success - Mark action as failed and increment retry count", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x5234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x5234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(300),
			},
		}
		require.NoError(t, store.StoreUserState(state))
		require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))

		var action BlockchainAction
		err := db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		initialRetries := action.Retries

		err = store.Fail(action.ID, "test error message")
		require.NoError(t, err)

		// Verify action was updated
		err = db.Where("id = ?", action.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, BlockchainActionStatusFailed, action.Status)
		assert.Equal(t, "test error message", action.Error)
		assert.Equal(t, initialRetries+1, action.Retries)
	})
}

func TestDBStore_FailNoRetry(t *testing.T) {
	t.Run("Success - Mark action as failed without incrementing retry count", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x6234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x6234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(400),
			},
		}
		require.NoError(t, store.StoreUserState(state))
		require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))

		var action BlockchainAction
		err := db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		initialRetries := action.Retries

		err = store.FailNoRetry(action.ID, "fatal error")
		require.NoError(t, err)

		// Verify action was updated
		err = db.Where("id = ?", action.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, BlockchainActionStatusFailed, action.Status)
		assert.Equal(t, "fatal error", action.Error)
		assert.Equal(t, initialRetries, action.Retries) // Should not increment
	})
}

func TestDBStore_RecordAttempt(t *testing.T) {
	t.Run("Success - Record attempt and increment retry count", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x7234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0x7234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(150),
			},
		}
		require.NoError(t, store.StoreUserState(state))
		require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))

		var action BlockchainAction
		err := db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		initialRetries := action.Retries

		err = store.RecordAttempt(action.ID, "temporary network error")
		require.NoError(t, err)

		// Verify action was updated
		err = db.Where("id = ?", action.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, BlockchainActionStatusPending, action.Status) // Still pending
		assert.Equal(t, "temporary network error", action.Error)
		assert.Equal(t, initialRetries+1, action.Retries)
	})
}

func TestDBStore_Complete(t *testing.T) {
	t.Run("Success - Mark action as completed with transaction hash", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x8234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x8234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(600),
			},
		}
		require.NoError(t, store.StoreUserState(state))
		require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))

		var action BlockchainAction
		err := db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		txHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

		err = store.Complete(action.ID, txHash)
		require.NoError(t, err)

		// Verify action was updated
		err = db.Where("id = ?", action.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, BlockchainActionStatusCompleted, action.Status)
		assert.Equal(t, txHash, action.TxHash)
		assert.Empty(t, action.Error) // Error should be cleared
	})

	t.Run("Success - Clears previous error on completion", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "0x9234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0x9234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(250),
			},
		}
		require.NoError(t, store.StoreUserState(state))
		require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))

		var action BlockchainAction
		err := db.Where("state_id = ?", state.ID).First(&action).Error
		require.NoError(t, err)

		// First record an error
		err = store.RecordAttempt(action.ID, "some error")
		require.NoError(t, err)

		// Then complete it
		txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		err = store.Complete(action.ID, txHash)
		require.NoError(t, err)

		// Verify error was cleared
		err = db.Where("id = ?", action.ID).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, BlockchainActionStatusCompleted, action.Status)
		assert.Empty(t, action.Error)
	})
}

func TestDBStore_GetActions(t *testing.T) {
	t.Run("Success - Get pending actions ordered by creation time", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create multiple states and actions
		state1 := core.State{
			ID:         "0xa234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0xa234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(100)},
		}
		state2 := core.State{
			ID:         "0xb234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0xb234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(200)},
		}
		state3 := core.State{
			ID:         "0xc234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0xc234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(300)},
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))
		require.NoError(t, store.StoreUserState(state3))

		require.NoError(t, store.ScheduleCheckpoint(state1.ID, 0))
		require.NoError(t, store.ScheduleInitiateEscrowWithdrawal(state2.ID, 0))
		require.NoError(t, store.ScheduleFinalizeEscrowDeposit(state3.ID, 0))

		// Get all pending actions
		actions, err := store.GetActions(0, 0)
		require.NoError(t, err)

		assert.Len(t, actions, 3)
		// Verify they're ordered by created_at ASC
		assert.True(t, actions[0].CreatedAt.Before(actions[1].CreatedAt) || actions[0].CreatedAt.Equal(actions[1].CreatedAt))
		assert.True(t, actions[1].CreatedAt.Before(actions[2].CreatedAt) || actions[1].CreatedAt.Equal(actions[2].CreatedAt))
	})

	t.Run("Success - Limit results", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create 5 states and actions
		for i := 0; i < 5; i++ {
			state := core.State{
				ID:         common.BigToHash(common.Big1).Hex() + string(rune('0'+i)),
				Asset:      "USDC",
				UserWallet: "0xd234567890123456789012345678901234567890",
				Epoch:      uint64(i + 1),
				Version:    1,
				Transition: core.Transition{},
				HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(int64(100 * (i + 1)))},
			}
			require.NoError(t, store.StoreUserState(state))
			require.NoError(t, store.ScheduleCheckpoint(state.ID, 0))
		}

		// Get only 2 actions
		actions, err := store.GetActions(2, 0)
		require.NoError(t, err)

		assert.Len(t, actions, 2)
	})

	t.Run("Success - Excludes completed and failed actions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create states
		state1 := core.State{
			ID:         "0xe234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0xe234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(100)},
		}
		state2 := core.State{
			ID:         "0xf234567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: "0xf234567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(200)},
		}
		state3 := core.State{
			ID:         "0x0334567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: "0x0334567890123456789012345678901234567890",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{UserBalance: decimal.NewFromInt(300)},
		}

		require.NoError(t, store.StoreUserState(state1))
		require.NoError(t, store.StoreUserState(state2))
		require.NoError(t, store.StoreUserState(state3))

		require.NoError(t, store.ScheduleCheckpoint(state1.ID, 0))
		require.NoError(t, store.ScheduleCheckpoint(state2.ID, 0))
		require.NoError(t, store.ScheduleCheckpoint(state3.ID, 0))

		// Get action IDs
		var action1, action2 BlockchainAction
		require.NoError(t, db.Where("state_id = ?", state1.ID).First(&action1).Error)
		require.NoError(t, db.Where("state_id = ?", state2.ID).First(&action2).Error)

		// Mark first as completed and second as failed
		require.NoError(t, store.Complete(action1.ID, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"))
		require.NoError(t, store.Fail(action2.ID, "some error"))

		// Get pending actions - should only return the third one
		actions, err := store.GetActions(0, 0)
		require.NoError(t, err)

		assert.Len(t, actions, 1)
		assert.Equal(t, state3.ID, actions[0].StateID)
		assert.Equal(t, BlockchainActionStatusPending, actions[0].Status)
	})

	t.Run("Success - Empty result when no pending actions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		actions, err := store.GetActions(0, 0)
		require.NoError(t, err)

		assert.Empty(t, actions)
	})
}
