package database

import (
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricID(t *testing.T) {
	t.Run("deterministic ID", func(t *testing.T) {
		id1, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		id2, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		assert.Equal(t, id1, id2)
	})

	t.Run("different labels produce different IDs", func(t *testing.T) {
		id1, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		id2, err := getMetricID("test_metric", "key1", "val2")
		require.NoError(t, err)

		assert.NotEqual(t, id1, id2)
	})

	t.Run("different names produce different IDs", func(t *testing.T) {
		id1, err := getMetricID("metric_a")
		require.NoError(t, err)

		id2, err := getMetricID("metric_b")
		require.NoError(t, err)

		assert.NotEqual(t, id1, id2)
	})

	t.Run("no labels", func(t *testing.T) {
		id, err := getMetricID("metric_no_labels")
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("ID starts with 0x", func(t *testing.T) {
		id, err := getMetricID("test")
		require.NoError(t, err)
		assert.Equal(t, "0x", id[:2])
	})
}

func TestGetLifetimeMetricLastTimestamp(t *testing.T) {
	t.Run("no metrics returns zero time", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		ts, err := store.GetLifetimeMetricLastTimestamp("nonexistent")
		require.NoError(t, err)
		assert.True(t, ts.IsZero())
	})

	t.Run("returns most recent timestamp", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		ts1 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
		ts2 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
		ts3 := time.Now().Truncate(time.Second)

		db.Create(&LifespanMetric{ID: "id-a", Name: "my_metric", Value: decimal.NewFromInt(1), LastTimestamp: ts1})
		db.Create(&LifespanMetric{ID: "id-b", Name: "my_metric", Value: decimal.NewFromInt(2), LastTimestamp: ts3})
		db.Create(&LifespanMetric{ID: "id-c", Name: "my_metric", Value: decimal.NewFromInt(3), LastTimestamp: ts2})

		latest, err := store.GetLifetimeMetricLastTimestamp("my_metric")
		require.NoError(t, err)
		assert.Equal(t, ts3.UTC(), latest.UTC())
	})

	t.Run("scoped to metric name", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		tsOld := time.Now().Add(-time.Hour).Truncate(time.Second)
		tsNew := time.Now().Truncate(time.Second)

		db.Create(&LifespanMetric{ID: "id-1", Name: "metric_a", Value: decimal.NewFromInt(1), LastTimestamp: tsOld})
		db.Create(&LifespanMetric{ID: "id-2", Name: "metric_b", Value: decimal.NewFromInt(1), LastTimestamp: tsNew})

		latest, err := store.GetLifetimeMetricLastTimestamp("metric_a")
		require.NoError(t, err)
		assert.Equal(t, tsOld.UTC(), latest.UTC())
	})
}

func TestCountActiveUsers(t *testing.T) {
	t.Run("no data returns only ALL with zero", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		require.Len(t, results, 1)
		assert.Equal(t, "ALL", results[0].Label)
		assert.Equal(t, uint64(0), results[0].Count)
	})

	t.Run("counts distinct users per asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		now := time.Now()
		db.Create(&UserBalance{UserWallet: "0xuser1", Asset: "USDC", Balance: decimal.NewFromInt(100), UpdatedAt: now})
		db.Create(&UserBalance{UserWallet: "0xuser2", Asset: "USDC", Balance: decimal.NewFromInt(200), UpdatedAt: now})
		db.Create(&UserBalance{UserWallet: "0xuser1", Asset: "ETH", Balance: decimal.NewFromInt(50), UpdatedAt: now})

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		require.Len(t, results, 3)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(2), countByLabel["USDC"])
		assert.Equal(t, uint64(1), countByLabel["ETH"])
		assert.Equal(t, uint64(2), countByLabel["ALL"])
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		old := time.Now().Add(-48 * time.Hour)
		recent := time.Now()
		db.Create(&UserBalance{UserWallet: "0xold", Asset: "USDC", Balance: decimal.NewFromInt(100), UpdatedAt: old})
		db.Create(&UserBalance{UserWallet: "0xnew", Asset: "USDC", Balance: decimal.NewFromInt(200), UpdatedAt: recent})

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(1), countByLabel["USDC"])
		assert.Equal(t, uint64(1), countByLabel["ALL"])
	})
}

func TestCountActiveAppSessions(t *testing.T) {
	t.Run("no data returns empty", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("counts sessions per application", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		now := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s3", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(2), countByLabel["app1"])
		assert.Equal(t, uint64(1), countByLabel["app2"])
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		old := time.Now().Add(-48 * time.Hour)
		recent := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: old})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: recent})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(1), countByLabel["app1"])
	})

	t.Run("multiple applications with mixed statuses", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := &DBStore{db: db}

		now := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusClosed, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s3", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s4", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s5", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusClosed, Nonce: 3, UpdatedAt: now})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		// CountActiveAppSessions counts all sessions regardless of status
		assert.Equal(t, uint64(2), countByLabel["app1"])
		assert.Equal(t, uint64(3), countByLabel["app2"])
	})
}
