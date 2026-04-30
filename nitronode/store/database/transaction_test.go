package database

import (
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction_TableName(t *testing.T) {
	tx := Transaction{}
	assert.Equal(t, "transactions", tx.TableName())
}

func TestDBStore_RecordTransaction(t *testing.T) {
	t.Run("Success - Record deposit transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		senderStateID := "state123"
		tx := core.Transaction{
			ID:                 "tx123",
			Asset:              "USDC",
			TxType:             core.TransactionTypeHomeDeposit,
			FromAccount:        "0xuser123",
			ToAccount:          "0xhomechannel123",
			SenderNewStateID:   &senderStateID,
			ReceiverNewStateID: nil,
			Amount:             decimal.NewFromInt(1000),
			CreatedAt:          time.Now(),
		}

		err := store.RecordTransaction(tx, "")
		require.NoError(t, err)

		// Verify transaction was recorded
		var dbTx Transaction
		err = db.Where("id = ?", "tx123").First(&dbTx).Error
		require.NoError(t, err)

		assert.Equal(t, "tx123", dbTx.ID)
		assert.Equal(t, "USDC", dbTx.AssetSymbol)
		assert.Equal(t, core.TransactionTypeHomeDeposit, dbTx.Type)
		assert.Equal(t, "0xuser123", dbTx.FromAccount)
		assert.Equal(t, "0xhomechannel123", dbTx.ToAccount)
		assert.NotNil(t, dbTx.SenderNewStateID)
		assert.Equal(t, "state123", *dbTx.SenderNewStateID)
		assert.Nil(t, dbTx.ReceiverNewStateID)
		assert.Equal(t, decimal.NewFromInt(1000), dbTx.Amount)
		assert.False(t, dbTx.CreatedAt.IsZero())
	})

	t.Run("Success - Record transfer transaction with both state IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		senderStateID := "senderstate456"
		receiverStateID := "receiverstate789"
		tx := core.Transaction{
			ID:                 "tx456",
			Asset:              "USDC",
			TxType:             core.TransactionTypeTransfer,
			FromAccount:        "0xuser123",
			ToAccount:          "0xuser456",
			SenderNewStateID:   &senderStateID,
			ReceiverNewStateID: &receiverStateID,
			Amount:             decimal.NewFromInt(500),
			CreatedAt:          time.Now(),
		}

		err := store.RecordTransaction(tx, "")
		require.NoError(t, err)

		// Verify transaction was recorded
		var dbTx Transaction
		err = db.Where("id = ?", "tx456").First(&dbTx).Error
		require.NoError(t, err)

		assert.Equal(t, "tx456", dbTx.ID)
		assert.Equal(t, core.TransactionTypeTransfer, dbTx.Type)
		assert.NotNil(t, dbTx.SenderNewStateID)
		assert.Equal(t, "senderstate456", *dbTx.SenderNewStateID)
		assert.NotNil(t, dbTx.ReceiverNewStateID)
		assert.Equal(t, "receiverstate789", *dbTx.ReceiverNewStateID)
	})

	t.Run("Error - Duplicate transaction ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		tx := core.Transaction{
			ID:          "tx789",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		err := store.RecordTransaction(tx, "")
		require.NoError(t, err)

		// Try to record again with same ID
		err = store.RecordTransaction(tx, "")
		assert.Error(t, err)
	})
}

func TestDBStore_GetUserTransactions(t *testing.T) {
	t.Run("Success - Get all transactions for user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create multiple transactions
		tx1 := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		}

		tx2 := core.Transaction{
			ID:          "tx2",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser123",
			ToAccount:   "0xuser456",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		}

		tx3 := core.Transaction{
			ID:          "tx3",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser456",
			ToAccount:   "0xuser123",
			Amount:      decimal.NewFromInt(200),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))
		require.NoError(t, store.RecordTransaction(tx3, ""))

		// Get all transactions for user123
		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", nil, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 3)
		assert.Equal(t, uint32(3), metadata.TotalCount)
		// Should be ordered by created_at DESC
		assert.Equal(t, "tx3", transactions[0].ID)
		assert.Equal(t, "tx2", transactions[1].ID)
		assert.Equal(t, "tx1", transactions[2].ID)
	})

	t.Run("Success - Filter by asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create transactions with different assets
		tx1 := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		tx2 := core.Transaction{
			ID:          "tx2",
			Asset:       "ETH",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel456",
			Amount:      decimal.NewFromInt(2),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))

		// Filter by USDC
		asset := "USDC"
		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", &asset, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "tx1", transactions[0].ID)
		assert.Equal(t, "USDC", transactions[0].Asset)
	})

	t.Run("Success - Filter by transaction type", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create transactions with different types
		tx1 := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		tx2 := core.Transaction{
			ID:          "tx2",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser123",
			ToAccount:   "0xuser456",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))

		// Filter by deposit type
		txType := core.TransactionTypeHomeDeposit
		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", nil, &txType, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "tx1", transactions[0].ID)
		assert.Equal(t, core.TransactionTypeHomeDeposit, transactions[0].TxType)
	})

	t.Run("Success - Filter by time range", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		baseTime := time.Now()

		// Create transactions at different times
		tx1 := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   baseTime.Add(-3 * time.Hour),
		}

		tx2 := core.Transaction{
			ID:          "tx2",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser123",
			ToAccount:   "0xuser456",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   baseTime.Add(-1 * time.Hour),
		}

		tx3 := core.Transaction{
			ID:          "tx3",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeWithdrawal,
			FromAccount: "0xhomechannel123",
			ToAccount:   "0xuser123",
			Amount:      decimal.NewFromInt(200),
			CreatedAt:   baseTime.Add(-1 * time.Second),
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))
		require.NoError(t, store.RecordTransaction(tx3, ""))

		// Filter from 2 hours ago to now
		fromTime := uint64(baseTime.Add(-2 * time.Hour).Unix())
		toTime := uint64(baseTime.Unix())
		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", nil, nil, &fromTime, &toTime, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 2) // Should get tx2 and tx3
		assert.Equal(t, uint32(2), metadata.TotalCount)
	})

	t.Run("Success - Pagination", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create 5 transactions
		for i := 1; i <= 5; i++ {
			txID := "tx" + string(rune(i+'0'))
			tx := core.Transaction{
				ID:          txID,
				Asset:       "USDC",
				TxType:      core.TransactionTypeTransfer,
				FromAccount: "0xuser123",
				ToAccount:   "0xuser456",
				Amount:      decimal.NewFromInt(int64(i * 100)),
				CreatedAt:   time.Now().Add(time.Duration(i) * time.Minute),
			}
			require.NoError(t, store.RecordTransaction(tx, ""))
		}

		// Get first page (2 items)
		limit := uint32(2)
		offset := uint32(0)
		pagination := &core.PaginationParams{
			Limit:  &limit,
			Offset: &offset,
		}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", nil, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 2)
		assert.Equal(t, uint32(5), metadata.TotalCount)
		assert.Equal(t, uint32(1), metadata.Page)
		assert.Equal(t, uint32(2), metadata.PerPage)
		assert.Equal(t, uint32(3), metadata.PageCount) // ceil(5/2) = 3

		// Get second page
		offset = 2
		pagination.Offset = &offset
		transactions, metadata, err = store.GetUserTransactions("0xuser123", nil, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 2)
		assert.Equal(t, uint32(2), metadata.Page)
	})

	t.Run("Success - Combined filters", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		baseTime := time.Now()

		// Create various transactions
		tx1 := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   baseTime.Add(-2 * time.Hour),
		}

		tx2 := core.Transaction{
			ID:          "tx2",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel123",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   baseTime,
		}

		tx3 := core.Transaction{
			ID:          "tx3",
			Asset:       "ETH",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuser123",
			ToAccount:   "0xhomechannel456",
			Amount:      decimal.NewFromInt(2),
			CreatedAt:   baseTime,
		}

		tx4 := core.Transaction{
			ID:          "tx4",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser123",
			ToAccount:   "0xuser456",
			Amount:      decimal.NewFromInt(100),
			CreatedAt:   baseTime,
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))
		require.NoError(t, store.RecordTransaction(tx3, ""))
		require.NoError(t, store.RecordTransaction(tx4, ""))

		// Filter: USDC + Deposit + from 1 hour ago
		asset := "USDC"
		txType := core.TransactionTypeHomeDeposit
		fromTime := uint64(baseTime.Add(-1 * time.Hour).Unix())
		pagination := &core.PaginationParams{}

		transactions, metadata, err := store.GetUserTransactions("0xuser123", &asset, &txType, &fromTime, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 1) // Should only get tx2
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "tx2", transactions[0].ID)
	})

	t.Run("No transactions found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xnonexistent", nil, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Empty(t, transactions)
		assert.Equal(t, uint32(0), metadata.TotalCount)
	})

	t.Run("Success - Get transactions as receiver", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Transaction where user123 is receiver
		tx := core.Transaction{
			ID:          "tx1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeTransfer,
			FromAccount: "0xuser456",
			ToAccount:   "0xuser123",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx, ""))

		pagination := &core.PaginationParams{}
		transactions, metadata, err := store.GetUserTransactions("0xuser123", nil, nil, nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, transactions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "tx1", transactions[0].ID)
		assert.Equal(t, "0xuser123", transactions[0].ToAccount)
	})

	t.Run("ApplicationID is persisted when provided", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		tx := core.Transaction{
			ID:          "txapp1",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xuserapp",
			ToAccount:   "0xchannelapp",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx, "my-app"))

		var dbTx Transaction
		require.NoError(t, db.Where("id = ?", "txapp1").First(&dbTx).Error)
		require.NotNil(t, dbTx.ApplicationID)
		assert.Equal(t, "my-app", *dbTx.ApplicationID)
	})

	t.Run("ApplicationID is NULL when empty", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		tx := core.Transaction{
			ID:          "txnoapp",
			Asset:       "USDC",
			TxType:      core.TransactionTypeHomeDeposit,
			FromAccount: "0xusernoapp",
			ToAccount:   "0xchannelnoapp",
			Amount:      decimal.NewFromInt(500),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx, ""))

		var dbTx Transaction
		require.NoError(t, db.Where("id = ?", "txnoapp").First(&dbTx).Error)
		assert.Nil(t, dbTx.ApplicationID)
	})
}
