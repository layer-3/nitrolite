package database

import (
	"testing"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppLedgerEntryV1_TableName(t *testing.T) {
	entry := AppLedgerEntryV1{}
	assert.Equal(t, "app_ledger_v1", entry.TableName())
}

func TestDBStore_RecordLedgerEntry(t *testing.T) {
	t.Run("Success - Record positive amount (credit)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		amount := decimal.NewFromInt(1000)
		err := store.RecordLedgerEntry("0xuser123", "session123", "USDC", amount)
		require.NoError(t, err)

		// Verify entry was recorded
		var entry AppLedgerEntryV1
		err = db.Where("account_id = ? AND asset_symbol = ?", "session123", "USDC").First(&entry).Error
		require.NoError(t, err)

		assert.Equal(t, "session123", entry.AccountID)
		assert.Equal(t, "USDC", entry.AssetSymbol)
		assert.True(t, decimal.NewFromInt(1000).Equal(entry.Credit))
		assert.True(t, decimal.Zero.Equal(entry.Debit))
		assert.False(t, entry.CreatedAt.IsZero())
	})

	t.Run("Success - Record negative amount (debit)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		amount := decimal.NewFromInt(-500)
		err := store.RecordLedgerEntry("0xuser456", "session456", "ETH", amount)
		require.NoError(t, err)

		// Verify entry was recorded
		var entry AppLedgerEntryV1
		err = db.Where("account_id = ? AND asset_symbol = ?", "session456", "ETH").First(&entry).Error
		require.NoError(t, err)

		assert.Equal(t, "session456", entry.AccountID)
		assert.Equal(t, "ETH", entry.AssetSymbol)
		assert.True(t, decimal.Zero.Equal(entry.Credit))
		assert.True(t, decimal.NewFromInt(500).Equal(entry.Debit)) // Absolute value
	})

	t.Run("Success - Zero amount (no entry)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		amount := decimal.Zero
		err := store.RecordLedgerEntry("0xuser789", "session789", "USDC", amount)
		require.NoError(t, err)

		// Verify no entry was created
		var count int64
		err = db.Model(&AppLedgerEntryV1{}).Where("account_id = ?", "session789").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Success - Multiple entries for same account", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Record multiple entries
		err := store.RecordLedgerEntry("0xuser100", "session100", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)

		err = store.RecordLedgerEntry("0xuser100", "session100", "USDC", decimal.NewFromInt(-300))
		require.NoError(t, err)

		err = store.RecordLedgerEntry("0xuser100", "session100", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)

		// Verify all entries were created
		var count int64
		err = db.Model(&AppLedgerEntryV1{}).Where("account_id = ? AND asset_symbol = ?", "session100", "USDC").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

func TestDBStore_GetAppSessionBalances(t *testing.T) {
	t.Run("Success - Get balances for single asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Record entries
		err := store.RecordLedgerEntry("0xuser123", "session123", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser123", "session123", "USDC", decimal.NewFromInt(-200))
		require.NoError(t, err)

		balances, err := store.GetAppSessionBalances("session123")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.True(t, decimal.NewFromInt(800).Equal(balances["USDC"])) // 1000 - 200
	})

	t.Run("Success - Get balances for multiple assets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Record entries for different assets
		err := store.RecordLedgerEntry("0xuser456", "session456", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser456", "session456", "ETH", decimal.NewFromInt(5))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser456", "session456", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)

		balances, err := store.GetAppSessionBalances("session456")
		require.NoError(t, err)

		assert.Len(t, balances, 2)
		assert.True(t, decimal.NewFromInt(1500).Equal(balances["USDC"]))
		assert.True(t, decimal.NewFromInt(5).Equal(balances["ETH"]))
	})

	t.Run("Success - Balance with credits and debits", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Record mixed entries
		err := store.RecordLedgerEntry("0xuser789", "session789", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser789", "session789", "USDC", decimal.NewFromInt(-300))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser789", "session789", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser789", "session789", "USDC", decimal.NewFromInt(-200))
		require.NoError(t, err)

		balances, err := store.GetAppSessionBalances("session789")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.True(t, decimal.NewFromInt(1000).Equal(balances["USDC"])) // 1000 - 300 + 500 - 200
	})

	t.Run("No entries found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		balances, err := store.GetAppSessionBalances("nonexistent")
		require.NoError(t, err)
		assert.Empty(t, balances)
	})

	t.Run("Success - Zero balance", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Record entries that cancel out
		err := store.RecordLedgerEntry("0xuser100", "session100", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser100", "session100", "USDC", decimal.NewFromInt(-500))
		require.NoError(t, err)

		balances, err := store.GetAppSessionBalances("session100")
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.True(t, decimal.Zero.Equal(balances["USDC"]))
	})
}

func TestDBStore_GetParticipantAllocations(t *testing.T) {
	t.Run("Success - Get allocations for single participant", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create app session with participant
		session := app.AppSessionV1{
			SessionID:     "session123",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session))

		// Record ledger entries for participant
		err := store.RecordLedgerEntry("0xuser123", "session123", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser123", "session123", "USDC", decimal.NewFromInt(-200))
		require.NoError(t, err)

		allocations, err := store.GetParticipantAllocations("session123")
		require.NoError(t, err)

		assert.Len(t, allocations, 1)
		assert.Contains(t, allocations, "0xuser123")
		assert.True(t, decimal.NewFromInt(800).Equal(allocations["0xuser123"]["USDC"]))
	})

	t.Run("Success - Get allocations for multiple participants", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create app session with participants
		session := app.AppSessionV1{
			SessionID:     "session456",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 50,
				},
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 50,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session))

		// Record ledger entries for participants
		err := store.RecordLedgerEntry("0xuser123", "session456", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser456", "session456", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)

		allocations, err := store.GetParticipantAllocations("session456")
		require.NoError(t, err)

		assert.Len(t, allocations, 2)
		assert.Contains(t, allocations, "0xuser123")
		assert.Contains(t, allocations, "0xuser456")
		assert.True(t, decimal.NewFromInt(1000).Equal(allocations["0xuser123"]["USDC"]))
		assert.True(t, decimal.NewFromInt(500).Equal(allocations["0xuser456"]["USDC"]))
	})

	t.Run("Success - Multiple assets per participant", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create app session with participant
		session := app.AppSessionV1{
			SessionID:     "session789",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session))

		// Record ledger entries for multiple assets
		err := store.RecordLedgerEntry("0xuser123", "session789", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser123", "session789", "ETH", decimal.NewFromInt(5))
		require.NoError(t, err)

		allocations, err := store.GetParticipantAllocations("session789")
		require.NoError(t, err)

		assert.Len(t, allocations, 1)
		assert.Contains(t, allocations, "0xuser123")
		assert.Len(t, allocations["0xuser123"], 2)
		assert.True(t, decimal.NewFromInt(1000).Equal(allocations["0xuser123"]["USDC"]))
		assert.True(t, decimal.NewFromInt(5).Equal(allocations["0xuser123"]["ETH"]))
	})

	t.Run("No allocations found - session has no participants with balances", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create app session with participant but no ledger entries
		session := app.AppSessionV1{
			SessionID:     "session100",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session))

		allocations, err := store.GetParticipantAllocations("session100")
		require.NoError(t, err)
		assert.Empty(t, allocations)
	})

	t.Run("No allocations found - nonexistent session", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		allocations, err := store.GetParticipantAllocations("nonexistent")
		require.NoError(t, err)
		assert.Empty(t, allocations)
	})

	t.Run("Success - Excludes non-participant entries", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create app session with one participant
		session := app.AppSessionV1{
			SessionID:     "session200",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session))

		// Record ledger entries for both participant and non-participant
		err := store.RecordLedgerEntry("0xuser123", "session200", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xotheruser", "session200", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)

		allocations, err := store.GetParticipantAllocations("session200")
		require.NoError(t, err)

		// Should only include participant wallet
		assert.Len(t, allocations, 1)
		assert.Contains(t, allocations, "0xuser123")
		assert.NotContains(t, allocations, "0xotheruser")
		assert.True(t, decimal.NewFromInt(1000).Equal(allocations["0xuser123"]["USDC"]))
	})

	t.Run("Success - Isolates allocations by app session", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create two app sessions with the same participant
		session1 := app.AppSessionV1{
			SessionID:     "session300",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session1))

		session2 := app.AppSessionV1{
			SessionID:     "session400",
			ApplicationID: "poker",
			Nonce:         2,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
		}
		require.NoError(t, store.CreateAppSession(session2))

		// Record different amounts for the same wallet in different sessions
		err := store.RecordLedgerEntry("0xuser123", "session300", "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)
		err = store.RecordLedgerEntry("0xuser123", "session400", "USDC", decimal.NewFromInt(500))
		require.NoError(t, err)

		// Get allocations for session1 - should only see session1 amounts
		allocations1, err := store.GetParticipantAllocations("session300")
		require.NoError(t, err)
		assert.Len(t, allocations1, 1)
		assert.Contains(t, allocations1, "0xuser123")
		assert.True(t, decimal.NewFromInt(1000).Equal(allocations1["0xuser123"]["USDC"]))

		// Get allocations for session2 - should only see session2 amounts
		allocations2, err := store.GetParticipantAllocations("session400")
		require.NoError(t, err)
		assert.Len(t, allocations2, 1)
		assert.Contains(t, allocations2, "0xuser123")
		assert.True(t, decimal.NewFromInt(500).Equal(allocations2["0xuser123"]["USDC"]))
	})
}
