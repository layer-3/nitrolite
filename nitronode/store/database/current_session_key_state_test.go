package database

import (
	"errors"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentSessionKeyStateV1_TableName(t *testing.T) {
	assert.Equal(t, "current_session_key_states_v1", CurrentSessionKeyStateV1{}.TableName())
}

// TestCurrentSessionKeyStateV1_UniqueKeyKindConstraint pins the (session_key, kind) uniqueness
// invariant at the database layer on every supported dialect. Postgres gets it from migration
// 20260508000000; sqlite gets it from the uniqueIndex gorm tag via AutoMigrate. Without the
// tag, sqlite would silently accept two pointer rows for the same key/kind under different
// wallets, breaking LockSessionKeyState's read-first-then-check ownership flow.
func TestCurrentSessionKeyStateV1_UniqueKeyKindConstraint(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	now := time.Now().UTC()
	first := CurrentSessionKeyStateV1{
		UserAddress: testUser1,
		SessionKey:  testSessionKey,
		Kind:        SessionKeyKindAppSession,
		Version:     1,
		UpdatedAt:   now,
	}
	require.NoError(t, db.Create(&first).Error)

	// Foreign wallet attempting the same (session_key, kind) must be rejected at the
	// database layer, not just by application logic.
	collision := CurrentSessionKeyStateV1{
		UserAddress: testUser2,
		SessionKey:  testSessionKey,
		Kind:        SessionKeyKindAppSession,
		Version:     1,
		UpdatedAt:   now,
	}
	err := db.Create(&collision).Error
	require.Error(t, err)

	// Same (session_key) under a different kind is allowed — the constraint is composite.
	otherKind := CurrentSessionKeyStateV1{
		UserAddress: testUser2,
		SessionKey:  testSessionKey,
		Kind:        SessionKeyKindChannel,
		Version:     1,
		UpdatedAt:   now,
	}
	require.NoError(t, db.Create(&otherKind).Error)
}

func TestDBStore_LockSessionKeyState(t *testing.T) {
	t.Run("Seeds row at version=0 on first call", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		v, expiresAt, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v)
		assert.True(t, expiresAt.IsZero(), "expires_at must be zero when no history row exists")

		// Second call returns the same seeded row.
		v2, expiresAt2, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v2)
		assert.True(t, expiresAt2.IsZero())
	})

	t.Run("Returns latest version and expires_at after a successful submit", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		futureExpiry := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   futureExpiry,
			UserSig:     "0xsig",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		v, expiresAt, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v)
		assert.WithinDuration(t, futureExpiry, expiresAt, time.Second)
	})

	t.Run("Returns past expires_at for a revoked latest version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// Active v1 followed by a revoke at v2 with past expires_at.
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig1",
		}))
		pastExpiry := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			ExpiresAt:   pastExpiry,
			UserSig:     "0xsig2",
		}))

		v, expiresAt, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), v)
		assert.WithinDuration(t, pastExpiry, expiresAt, time.Second)
		assert.True(t, expiresAt.Before(time.Now()), "revoked latest must surface a past expires_at")
	})

	t.Run("Channel and app_session kinds are independent", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// Submit channel session key v1.
		require.NoError(t, store.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))

		channelV, channelExpiresAt, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindChannel)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), channelV)
		assert.False(t, channelExpiresAt.IsZero())

		// App-session pointer for the same (user, session_key) is unaffected.
		appV, appExpiresAt, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), appV)
		assert.True(t, appExpiresAt.IsZero())
	})

	t.Run("Foreign wallet trying to claim an already-owned (session_key, kind) is rejected", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// User1 owns the session key for the app-session kind.
		_, _, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)

		// User2 attempts to lock the same (session_key, kind) — must surface the generic
		// not-allowed sentinel without leaking that the key belongs to someone else.
		_, _, err = store.LockSessionKeyState(testUser2, testSessionKey, SessionKeyKindAppSession)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrSessionKeyNotAllowed))
	})

	t.Run("Same (user, session_key) across both kinds is allowed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		_, _, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindChannel)
		require.NoError(t, err)
		_, _, err = store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
	})

	t.Run("Lowercases user_address and session_key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		_, _, err := store.LockSessionKeyState(
			"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
			SessionKeyKindAppSession,
		)
		require.NoError(t, err)

		// Lower-case query returns the same row (no duplicate seeded).
		v, expiresAt, err := store.LockSessionKeyState(
			"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			SessionKeyKindAppSession,
		)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v)
		assert.True(t, expiresAt.IsZero())

		count, err := store.CountSessionKeysForUser("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		require.NoError(t, err)
		// Seeded-only rows (version=0) must not count toward the cap.
		assert.Equal(t, uint32(0), count)
	})
}

func TestDBStore_CountSessionKeysForUser(t *testing.T) {
	t.Run("Counts only rows with version > 0", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// Seed-only row (Lock with no submit) must not be counted.
		_, _, err := store.LockSessionKeyState(testUser1, testKeyA, SessionKeyKindAppSession)
		require.NoError(t, err)

		count, err := store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(0), count)

		// Real submit -> counted.
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))
		count, err = store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(1), count)
	})

	t.Run("Counts across both kinds for the same user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))
		require.NoError(t, store.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))

		count, err := store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(2), count)
	})

	t.Run("Does not count keys belonging to a different user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))

		count, err := store.CountSessionKeysForUser(testUser2)
		require.NoError(t, err)
		assert.Equal(t, uint32(0), count)
	})

	t.Run("Revoked or expired keys do not count against the cap", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		// Active app-session key.
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))

		// Revoked app-session key (submit at the next version with expires_at in the past).
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(-time.Hour),
			UserSig:     "0xsig",
		}))

		// Active channel key for the same user.
		require.NoError(t, store.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}))

		// Revoked channel key.
		require.NoError(t, store.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(-time.Hour),
			UserSig:     "0xsig",
		}))

		count, err := store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(2), count, "only active keys (1 app + 1 channel) should count")
	})

	t.Run("Rotating an existing key out via past expires_at frees the slot", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_active",
		}))

		count, err := store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(1), count)

		// Submit version 2 with past expires_at as a revoke.
		require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     2,
			ExpiresAt:   time.Now().Add(-time.Hour),
			UserSig:     "0xsig_revoke",
		}))

		count, err = store.CountSessionKeysForUser(testUser1)
		require.NoError(t, err)
		assert.Equal(t, uint32(0), count, "revoke must free the slot")
	})
}

func TestDBStore_CurrentPointer_VersionMonotonic(t *testing.T) {
	// Out-of-order writers must not regress the pointer (EXCLUDED.version > current.version).
	db, cleanup := SetupTestDB(t)
	defer cleanup()
	store := NewDBStore(db)

	require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
		UserAddress: testUser1,
		SessionKey:  testSessionKey,
		Version:     1,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		UserSig:     "0xsig1",
	}))
	require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
		UserAddress: testUser1,
		SessionKey:  testSessionKey,
		Version:     3,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		UserSig:     "0xsig3",
	}))

	// Pointer reflects version 3.
	v, err := store.GetLastAppSessionKeyVersion(testUser1, testSessionKey)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), v)

	// A late-arriving v2 must not regress the pointer back to 2 (the EXCLUDED.version > current
	// guard is what protects this; the history row insert itself is allowed).
	require.NoError(t, store.StoreAppSessionKeyState(app.AppSessionKeyStateV1{
		UserAddress: testUser1,
		SessionKey:  testSessionKey,
		Version:     2,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		UserSig:     "0xsig2",
	}))
	v, err = store.GetLastAppSessionKeyVersion(testUser1, testSessionKey)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), v)
}
