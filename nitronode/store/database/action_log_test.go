package database

import (
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordAction(t *testing.T) {
	t.Run("records action successfully", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		err := store.RecordAction("0xUser123", core.GatedActionTransfer)
		require.NoError(t, err)

		count, err := store.GetUserActionCount("0xuser123", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)
	})

	t.Run("records multiple actions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))

		count, err := store.GetUserActionCount("0xuser", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), count)
	})

	t.Run("normalizes wallet to lowercase", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xABCDEF", core.GatedActionTransfer))

		count, err := store.GetUserActionCount("0xabcdef", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)
	})
}

func TestGetUserActionCount(t *testing.T) {
	t.Run("returns zero for no actions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		count, err := store.GetUserActionCount("0xuser", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})

	t.Run("filters by gated action", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionAppSessionOperation))

		transferCount, err := store.GetUserActionCount("0xuser", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), transferCount)

		opCount, err := store.GetUserActionCount("0xuser", core.GatedActionAppSessionOperation, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), opCount)
	})

	t.Run("filters by wallet", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xuser1", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser1", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser2", core.GatedActionTransfer))

		count, err := store.GetUserActionCount("0xuser1", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), count)

		count, err = store.GetUserActionCount("0xuser2", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		oldEntry := ActionLogEntryV1{
			ID:          [16]byte{1},
			UserWallet:  "0xuser",
			GatedAction: core.GatedActionTransfer.ID(),
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		}
		require.NoError(t, db.Create(&oldEntry).Error)
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))

		count, err := store.GetUserActionCount("0xuser", core.GatedActionTransfer, time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)

		count, err = store.GetUserActionCount("0xuser", core.GatedActionTransfer, 3*time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), count)
	})
}

func TestGetUserActionCounts(t *testing.T) {
	t.Run("returns empty map for no actions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		counts, err := store.GetUserActionCounts("0xuser", time.Hour)
		require.NoError(t, err)
		assert.Empty(t, counts)
	})

	t.Run("groups by gated action", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionAppSessionOperation))

		counts, err := store.GetUserActionCounts("0xuser", time.Hour)
		require.NoError(t, err)

		assert.Equal(t, uint64(2), counts[core.GatedActionTransfer])
		assert.Equal(t, uint64(1), counts[core.GatedActionAppSessionOperation])
	})

	t.Run("filters by wallet", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.RecordAction("0xuser1", core.GatedActionTransfer))
		require.NoError(t, store.RecordAction("0xuser2", core.GatedActionTransfer))

		counts, err := store.GetUserActionCounts("0xuser1", time.Hour)
		require.NoError(t, err)
		assert.Len(t, counts, 1)
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		oldEntry := ActionLogEntryV1{
			ID:          [16]byte{1},
			UserWallet:  "0xuser",
			GatedAction: core.GatedActionTransfer.ID(),
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		}
		require.NoError(t, db.Create(&oldEntry).Error)
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))

		counts, err := store.GetUserActionCounts("0xuser", time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), counts[core.GatedActionTransfer])

		counts, err = store.GetUserActionCounts("0xuser", 3*time.Hour)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), counts[core.GatedActionTransfer])
	})

	t.Run("skips unknown gated action IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// Insert an entry with an unknown gated action ID directly
		entry := ActionLogEntryV1{
			ID:          [16]byte{99},
			UserWallet:  "0xuser",
			GatedAction: 255, // unknown ID
			CreatedAt:   time.Now(),
		}
		require.NoError(t, db.Create(&entry).Error)
		require.NoError(t, store.RecordAction("0xuser", core.GatedActionTransfer))

		counts, err := store.GetUserActionCounts("0xuser", time.Hour)
		require.NoError(t, err)
		assert.Len(t, counts, 1)
		assert.Equal(t, uint64(1), counts[core.GatedActionTransfer])
	})
}

func TestGetAppCount(t *testing.T) {
	t.Run("returns zero for no apps", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		count, err := store.GetAppCount("0xowner")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})

	t.Run("counts apps for owner", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, db.Create(&AppV1{ID: "app1", OwnerWallet: "0xowner1", Metadata: "{}"}).Error)
		require.NoError(t, db.Create(&AppV1{ID: "app2", OwnerWallet: "0xowner1", Metadata: "{}"}).Error)
		require.NoError(t, db.Create(&AppV1{ID: "app3", OwnerWallet: "0xowner2", Metadata: "{}"}).Error)

		count, err := store.GetAppCount("0xowner1")
		require.NoError(t, err)
		assert.Equal(t, uint64(2), count)

		count, err = store.GetAppCount("0xowner2")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)
	})

	t.Run("normalizes wallet to lowercase", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, db.Create(&AppV1{ID: "app1", OwnerWallet: "0xabcdef", Metadata: "{}"}).Error)

		count, err := store.GetAppCount("0xABCDEF")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), count)
	})
}
