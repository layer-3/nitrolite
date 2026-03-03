package database

import (
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppV1_TableName(t *testing.T) {
	a := AppV1{}
	assert.Equal(t, "apps_v1", a.TableName())
}

func TestDBStore_CreateApp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		entry := app.AppV1{
			ID:                          "test-app",
			OwnerWallet:                 "0x1111111111111111111111111111111111111111",
			Metadata:                    "0xabcdef",
			Version:                     1,
			CreationApprovalNotRequired: true,
		}

		err := store.CreateApp(entry)
		require.NoError(t, err)

		// Verify app was created
		var dbApp AppV1
		err = db.Where("id = ?", "test-app").First(&dbApp).Error
		require.NoError(t, err)

		assert.Equal(t, "test-app", dbApp.ID)
		assert.Equal(t, "0x1111111111111111111111111111111111111111", dbApp.OwnerWallet)
		assert.Equal(t, "0xabcdef", dbApp.Metadata)
		assert.Equal(t, uint64(1), dbApp.Version)
		assert.True(t, dbApp.CreationApprovalNotRequired)
	})

	t.Run("Duplicate ID error", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		entry := app.AppV1{
			ID:          "test-app",
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0xabcdef",
			Version:     1,
		}

		err := store.CreateApp(entry)
		require.NoError(t, err)

		// Try to create again with same ID
		err = store.CreateApp(entry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create app")
	})

	t.Run("Stores lowercase ID and wallet", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		entry := app.AppV1{
			ID:          "My-App",
			OwnerWallet: "0xABCD1234567890ABCDEF1234567890ABCDEF1234",
			Metadata:    "0x00",
			Version:     1,
		}

		err := store.CreateApp(entry)
		require.NoError(t, err)

		var dbApp AppV1
		err = db.Where("id = ?", "my-app").First(&dbApp).Error
		require.NoError(t, err)

		assert.Equal(t, "my-app", dbApp.ID)
		assert.Equal(t, "0xabcd1234567890abcdef1234567890abcdef1234", dbApp.OwnerWallet)
	})
}

func TestDBStore_GetApp(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		entry := app.AppV1{
			ID:                          "test-app",
			OwnerWallet:                 "0x1111111111111111111111111111111111111111",
			Metadata:                    "0xabcdef",
			Version:                     1,
			CreationApprovalNotRequired: true,
		}
		require.NoError(t, store.CreateApp(entry))

		result, err := store.GetApp("test-app")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "test-app", result.App.ID)
		assert.Equal(t, "0x1111111111111111111111111111111111111111", result.App.OwnerWallet)
		assert.Equal(t, "0xabcdef", result.App.Metadata)
		assert.Equal(t, uint64(1), result.App.Version)
		assert.True(t, result.App.CreationApprovalNotRequired)
	})

	t.Run("Not found returns nil", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetApp("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Case-insensitive lookup", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		entry := app.AppV1{
			ID:          "test-app",
			OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata:    "0x00",
			Version:     1,
		}
		require.NoError(t, store.CreateApp(entry))

		// Look up with different casing
		result, err := store.GetApp("Test-App")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test-app", result.App.ID)
	})
}

func TestDBStore_GetApps(t *testing.T) {
	t.Run("No filter", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-1", OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata: "0x01", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-2", OwnerWallet: "0x2222222222222222222222222222222222222222",
			Metadata: "0x02", Version: 1,
		}))

		// Small delay to ensure different created_at times
		time.Sleep(10 * time.Millisecond)

		pagination := &core.PaginationParams{}
		apps, metadata, err := store.GetApps(nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 2)
		assert.Equal(t, uint32(2), metadata.TotalCount)
	})

	t.Run("Filter by appID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-1", OwnerWallet: "0x1111111111111111111111111111111111111111",
			Metadata: "0x01", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-2", OwnerWallet: "0x2222222222222222222222222222222222222222",
			Metadata: "0x02", Version: 1,
		}))

		appID := "app-1"
		pagination := &core.PaginationParams{}
		apps, metadata, err := store.GetApps(&appID, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "app-1", apps[0].App.ID)
	})

	t.Run("Filter by ownerWallet", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		owner := "0x1111111111111111111111111111111111111111"
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-1", OwnerWallet: owner, Metadata: "0x01", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-2", OwnerWallet: "0x2222222222222222222222222222222222222222",
			Metadata: "0x02", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-3", OwnerWallet: owner, Metadata: "0x03", Version: 1,
		}))

		pagination := &core.PaginationParams{}
		apps, metadata, err := store.GetApps(nil, &owner, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 2)
		assert.Equal(t, uint32(2), metadata.TotalCount)
		for _, a := range apps {
			assert.Equal(t, owner, a.App.OwnerWallet)
		}
	})

	t.Run("Combined filters", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		owner := "0x1111111111111111111111111111111111111111"
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-1", OwnerWallet: owner, Metadata: "0x01", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-2", OwnerWallet: owner, Metadata: "0x02", Version: 1,
		}))
		require.NoError(t, store.CreateApp(app.AppV1{
			ID: "app-3", OwnerWallet: "0x2222222222222222222222222222222222222222",
			Metadata: "0x03", Version: 1,
		}))

		appID := "app-1"
		pagination := &core.PaginationParams{}
		apps, metadata, err := store.GetApps(&appID, &owner, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "app-1", apps[0].App.ID)
		assert.Equal(t, owner, apps[0].App.OwnerWallet)
	})

	t.Run("Pagination", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		for i := 1; i <= 3; i++ {
			require.NoError(t, store.CreateApp(app.AppV1{
				ID:          "app-" + string(rune(i+'0')),
				OwnerWallet: "0x1111111111111111111111111111111111111111",
				Metadata:    "0x00",
				Version:     1,
			}))
			time.Sleep(10 * time.Millisecond)
		}

		limit := uint32(2)
		offset := uint32(0)
		pagination := &core.PaginationParams{Limit: &limit, Offset: &offset}

		apps, metadata, err := store.GetApps(nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 2)
		assert.Equal(t, uint32(3), metadata.TotalCount)
		assert.Equal(t, uint32(1), metadata.Page)
		assert.Equal(t, uint32(2), metadata.PerPage)

		// Second page
		offset = 2
		pagination.Offset = &offset
		apps, metadata, err = store.GetApps(nil, nil, pagination)
		require.NoError(t, err)

		assert.Len(t, apps, 1)
		assert.Equal(t, uint32(2), metadata.Page)
	})

	t.Run("Empty results", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		appID := "nonexistent"
		pagination := &core.PaginationParams{}
		apps, metadata, err := store.GetApps(&appID, nil, pagination)
		require.NoError(t, err)

		assert.Empty(t, apps)
		assert.Equal(t, uint32(0), metadata.TotalCount)
	})
}
