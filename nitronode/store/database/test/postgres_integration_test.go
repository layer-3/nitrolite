package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/nitronode/store/database"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgreSQL connection string should be provided via environment variable
// Example: POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=nitrolite_test port=5432 sslmode=disable"
func getPostgresDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN environment variable not set, skipping PostgreSQL integration tests")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to PostgreSQL")

	return db
}

func TestPostgres_ChannelOperations(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	t.Run("Create and retrieve channel", func(t *testing.T) {
		channel := core.Channel{
			ChannelID:         "0x1234567890123456789012345678901234567890123456789012345678901234",
			UserWallet:        "0x1234567890123456789012345678901234567890",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		retrieved, err := store.GetChannelByID(channel.ChannelID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, channel.ChannelID, retrieved.ChannelID)
		assert.Equal(t, channel.UserWallet, retrieved.UserWallet)
		assert.Equal(t, channel.Type, retrieved.Type)
		assert.Equal(t, channel.BlockchainID, retrieved.BlockchainID)
	})

	t.Run("Update channel", func(t *testing.T) {
		channel := core.Channel{
			ChannelID:         "0x2234567890123456789012345678901234567890123456789012345678901234",
			UserWallet:        "0x2234567890123456789012345678901234567890",
			Asset:             "usdc",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		channel.Status = core.ChannelStatusClosed
		channel.StateVersion = 5
		err = store.UpdateChannel(channel)
		require.NoError(t, err)

		retrieved, err := store.GetChannelByID(channel.ChannelID)
		require.NoError(t, err)
		assert.Equal(t, core.ChannelStatusClosed, retrieved.Status)
		assert.Equal(t, uint64(5), retrieved.StateVersion)
	})
}

func TestPostgres_StateOperations(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	wallet := "0x3234567890123456789012345678901234567890"
	asset := "USDC"

	t.Run("Store and retrieve state", func(t *testing.T) {
		state := core.State{
			ID:         "0x3334567890123456789012345678901234567890123456789012345678901234",
			Asset:      asset,
			UserWallet: wallet,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.NewFromInt(100),
				NodeBalance: decimal.NewFromInt(500),
				NodeNetFlow: decimal.NewFromInt(50),
			},
		}

		err := store.StoreUserState(state, "")
		require.NoError(t, err)

		retrieved, err := store.GetLastUserState(wallet, asset, false)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, state.ID, retrieved.ID)
		assert.Equal(t, state.UserWallet, retrieved.UserWallet)
		assert.Equal(t, state.Asset, retrieved.Asset)
		assert.True(t, state.HomeLedger.UserBalance.Equal(retrieved.HomeLedger.UserBalance))
		core.ValidateDecimalPrecision(state.HomeLedger.NodeBalance, 0)
		core.ValidateDecimalPrecision(state.HomeLedger.NodeNetFlow, 0)
		core.ValidateDecimalPrecision(state.HomeLedger.UserBalance, 0)
		core.ValidateDecimalPrecision(state.HomeLedger.UserNetFlow, 0)
	})

	t.Run("Store multiple states and retrieve latest", func(t *testing.T) {
		wallet2 := "0x4234567890123456789012345678901234567890"

		state1 := core.State{
			ID:         "0x4334567890123456789012345678901234567890123456789012345678901234",
			Asset:      asset,
			UserWallet: wallet2,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
			},
		}

		state2 := core.State{
			ID:         "0x4434567890123456789012345678901234567890123456789012345678901234",
			Asset:      asset,
			UserWallet: wallet2,
			Epoch:      1,
			Version:    2,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(900),
			},
		}

		state3 := core.State{
			ID:         "0x4534567890123456789012345678901234567890123456789012345678901234",
			Asset:      asset,
			UserWallet: wallet2,
			Epoch:      2,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(850),
			},
		}

		require.NoError(t, store.StoreUserState(state1, ""))
		require.NoError(t, store.StoreUserState(state2, ""))
		require.NoError(t, store.StoreUserState(state3, ""))

		retrieved, err := store.GetLastUserState(wallet2, asset, false)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		// Should get state3 (highest epoch, then highest version)
		assert.Equal(t, state3.ID, retrieved.ID)
		assert.Equal(t, uint64(2), retrieved.Epoch)
		assert.Equal(t, uint64(1), retrieved.Version)
	})
}

func TestPostgres_TransactionOperations(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	wallet := "0x5234567890123456789012345678901234567890"

	t.Run("Record and retrieve transactions", func(t *testing.T) {
		tx1 := core.Transaction{
			ID:          "0x5534567890123456789012345678901234567890123456789012345678901234",
			TxType:      core.TransactionTypeHomeDeposit,
			Asset:       "USDC",
			FromAccount: wallet,
			ToAccount:   "0x5634567890123456789012345678901234567890123456789012345678901234",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		err := store.RecordTransaction(tx1, "")
		require.NoError(t, err)

		transactions, metadata, err := store.GetUserTransactions(wallet, nil, nil, nil, nil, &core.PaginationParams{})
		require.NoError(t, err)

		assert.Len(t, transactions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, tx1.ID, transactions[0].ID)
		assert.True(t, tx1.Amount.Equal(transactions[0].Amount))
	})

	t.Run("Filter transactions by asset", func(t *testing.T) {
		wallet2 := "0x6234567890123456789012345678901234567890"

		tx1 := core.Transaction{
			ID:          "0x6534567890123456789012345678901234567890123456789012345678901234",
			TxType:      core.TransactionTypeHomeDeposit,
			Asset:       "USDC",
			FromAccount: wallet2,
			ToAccount:   "0x6634567890123456789012345678901234567890123456789012345678901234",
			Amount:      decimal.NewFromInt(1000),
			CreatedAt:   time.Now(),
		}

		tx2 := core.Transaction{
			ID:          "0x6734567890123456789012345678901234567890123456789012345678901234",
			TxType:      core.TransactionTypeHomeDeposit,
			Asset:       "ETH",
			FromAccount: wallet2,
			ToAccount:   "0x6834567890123456789012345678901234567890123456789012345678901234",
			Amount:      decimal.NewFromInt(5),
			CreatedAt:   time.Now(),
		}

		require.NoError(t, store.RecordTransaction(tx1, ""))
		require.NoError(t, store.RecordTransaction(tx2, ""))

		assetFilter := "USDC"
		transactions, _, err := store.GetUserTransactions(wallet2, &assetFilter, nil, nil, nil, &core.PaginationParams{})
		require.NoError(t, err)

		assert.Len(t, transactions, 1)
		assert.Equal(t, "USDC", transactions[0].Asset)
	})
}

func TestPostgres_AppSessionOperations(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	t.Run("Create and retrieve app session", func(t *testing.T) {
		session := app.AppSessionV1{
			SessionID:     "0x7234567890123456789012345678901234567890123456789012345678901234",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0x7334567890123456789012345678901234567890",
					SignatureWeight: 50,
				},
				{
					WalletAddress:   "0x7434567890123456789012345678901234567890",
					SignatureWeight: 50,
				},
			},
			SessionData: `{"game":"holdem","stakes":"1/2"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.CreateAppSession(session)
		require.NoError(t, err)

		retrieved, err := store.GetAppSession(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, session.SessionID, retrieved.SessionID)
		assert.Equal(t, session.ApplicationID, retrieved.ApplicationID)
		assert.Len(t, retrieved.Participants, 2)
	})

	t.Run("Get app sessions by participant", func(t *testing.T) {
		participant := "0x7534567890123456789012345678901234567890"

		session := app.AppSessionV1{
			SessionID:     "0x7634567890123456789012345678901234567890123456789012345678901234",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   participant,
					SignatureWeight: 100,
				},
			},
			SessionData: `{"game":"holdem"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session))

		sessions, metadata, err := store.GetAppSessions(nil, &participant, app.AppSessionStatusVoid, &core.PaginationParams{})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(sessions), 1)
		assert.GreaterOrEqual(t, metadata.TotalCount, uint32(1))
	})

	t.Run("Update app session", func(t *testing.T) {
		session := app.AppSessionV1{
			SessionID:     "0x7734567890123456789012345678901234567890123456789012345678901234",
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0x7834567890123456789012345678901234567890",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"game":"holdem"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session))

		session.Status = app.AppSessionStatusClosed
		session.SessionData = `{"game":"holdem","result":"completed"}`
		session.Version = 2

		err := store.UpdateAppSession(session)
		require.NoError(t, err)

		retrieved, err := store.GetAppSession(session.SessionID)
		require.NoError(t, err)
		assert.Equal(t, app.AppSessionStatusClosed, retrieved.Status)
		assert.Equal(t, uint64(2), retrieved.Version)
	})
}

func TestPostgres_AppLedgerOperations(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	t.Run("Record ledger entries and get balances", func(t *testing.T) {
		sessionID := "0x8234567890123456789012345678901234567890123456789012345678901234"
		wallet := "0x8334567890123456789012345678901234567890"

		err := store.RecordLedgerEntry(wallet, sessionID, "USDC", decimal.NewFromInt(1000))
		require.NoError(t, err)

		err = store.RecordLedgerEntry(wallet, sessionID, "USDC", decimal.NewFromInt(-200))
		require.NoError(t, err)

		balances, err := store.GetAppSessionBalances(sessionID)
		require.NoError(t, err)

		assert.Len(t, balances, 1)
		assert.True(t, decimal.NewFromInt(800).Equal(balances["USDC"]))
	})

	t.Run("Get participant allocations", func(t *testing.T) {
		sessionID := "0x8434567890123456789012345678901234567890123456789012345678901234"
		wallet1 := "0x8534567890123456789012345678901234567890"
		wallet2 := "0x8634567890123456789012345678901234567890"

		// Create session with participants
		session := app.AppSessionV1{
			SessionID:     sessionID,
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{WalletAddress: wallet1, SignatureWeight: 50},
				{WalletAddress: wallet2, SignatureWeight: 50},
			},
			SessionData: `{"game":"holdem"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, store.CreateAppSession(session))

		// Record ledger entries
		require.NoError(t, store.RecordLedgerEntry(wallet1, wallet1, "USDC", decimal.NewFromInt(1000)))
		require.NoError(t, store.RecordLedgerEntry(wallet2, wallet2, "USDC", decimal.NewFromInt(500)))

		allocations, err := store.GetParticipantAllocations(sessionID)
		require.NoError(t, err)

		assert.Len(t, allocations, 2)
		assert.True(t, decimal.NewFromInt(1000).Equal(allocations[wallet1]["USDC"]))
		assert.True(t, decimal.NewFromInt(500).Equal(allocations[wallet2]["USDC"]))
	})
}

func TestPostgres_UserBalances(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	wallet := "0x9234567890123456789012345678901234567890"

	t.Run("Get user balances across multiple assets", func(t *testing.T) {
		state1 := core.State{
			ID:         "0x9334567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: wallet,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
			},
		}

		state2 := core.State{
			ID:         "0x9434567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: wallet,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(5),
			},
		}

		// Lock user states before storing
		_, err := store.LockUserState(wallet, "USDC")
		require.NoError(t, err)
		_, err = store.LockUserState(wallet, "ETH")
		require.NoError(t, err)

		require.NoError(t, store.StoreUserState(state1, ""))
		require.NoError(t, store.StoreUserState(state2, ""))

		balances, err := store.GetUserBalances(wallet)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(balances), 2)

		// Find USDC and ETH balances
		var usdcBalance, ethBalance *decimal.Decimal
		for _, b := range balances {
			if b.Asset == "USDC" {
				usdcBalance = &b.Balance
			}
			if b.Asset == "ETH" {
				ethBalance = &b.Balance
			}
		}

		require.NotNil(t, usdcBalance)
		require.NotNil(t, ethBalance)
		assert.True(t, decimal.NewFromInt(1000).Equal(*usdcBalance))
		assert.True(t, decimal.NewFromInt(5).Equal(*ethBalance))
	})
}

func TestPostgres_DecimalPrecision(t *testing.T) {
	db := getPostgresDB(t)
	store := database.NewDBStore(db)

	t.Run("Store and retrieve large decimal values", func(t *testing.T) {
		wallet := "0xa234567890123456789012345678901234567890"

		// Test with a large number with 18 decimal places
		largeBalance := decimal.RequireFromString("123456789012345678.123456789012345678")
		largeNetFlow := decimal.RequireFromString("999999999999999999")

		state := core.State{
			ID:         "0xa334567890123456789012345678901234567890123456789012345678901234",
			Asset:      "USDC",
			UserWallet: wallet,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: largeBalance,
				UserNetFlow: largeNetFlow,
			},
		}

		err := store.StoreUserState(state, "")
		require.NoError(t, err)

		retrieved, err := store.GetLastUserState(wallet, "USDC", false)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.True(t, largeBalance.Equal(retrieved.HomeLedger.UserBalance))
		assert.True(t, largeNetFlow.Equal(retrieved.HomeLedger.UserNetFlow))
	})

	t.Run("Store and retrieve negative net flow", func(t *testing.T) {
		wallet := "0xb234567890123456789012345678901234567890"
		negativeNetFlow := decimal.RequireFromString("-12345678901234567")

		state := core.State{
			ID:         "0xb334567890123456789012345678901234567890123456789012345678901234",
			Asset:      "ETH",
			UserWallet: wallet,
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(500),
				UserNetFlow: negativeNetFlow,
			},
		}

		err := store.StoreUserState(state, "")
		require.NoError(t, err)

		retrieved, err := store.GetLastUserState(wallet, "ETH", false)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.True(t, negativeNetFlow.Equal(retrieved.HomeLedger.UserNetFlow))
	})
}

// PrintCleanupSQL prints SQL commands to clean up the test data
func TestPostgres_PrintCleanupSQL(t *testing.T) {
	fmt.Print(`
\n=== CLEANUP SQL COMMANDS ===
Run these commands to clean up the test database:

-- Delete all test data
DELETE FROM blockchain_actions;
DELETE FROM app_ledger_v1;
DELETE FROM app_session_participants_v1;
DELETE FROM app_sessions_v1;
DELETE FROM transactions;
DELETE FROM channel_states;
DELETE FROM channels;

-- Or drop and recreate all tables (use with caution!)
DROP TABLE IF EXISTS blockchain_actions CASCADE;
DROP TABLE IF EXISTS app_ledger_v1 CASCADE;
DROP TABLE IF EXISTS app_session_participants_v1 CASCADE;
DROP TABLE IF EXISTS app_sessions_v1 CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS channel_states CASCADE;
DROP TABLE IF EXISTS channels CASCADE;
DROP TABLE IF EXISTS contract_events CASCADE;
DROP TABLE IF EXISTS session_keys CASCADE;
DROP TABLE IF EXISTS rpc_store CASCADE;

-- Then re-run the migrations to recreate the schema
=== END CLEANUP SQL ===\n
`)
}
