package database

import (
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

func TestDBStore_LockSessionKeyState(t *testing.T) {
	t.Run("Seeds row at version=0 on first call", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		v, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v)

		// Second call returns the same seeded row.
		v2, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v2)
	})

	t.Run("Returns latest version after a successful submit", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		v, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v)
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

		channelV, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindChannel)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), channelV)

		// App-session pointer for the same (user, session_key) is unaffected.
		appV, err := store.LockSessionKeyState(testUser1, testSessionKey, SessionKeyKindAppSession)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), appV)
	})

	t.Run("Lowercases user_address and session_key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		_, err := store.LockSessionKeyState(
			"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
			SessionKeyKindAppSession,
		)
		require.NoError(t, err)

		// Lower-case query returns the same row (no duplicate seeded).
		v, err := store.LockSessionKeyState(
			"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			SessionKeyKindAppSession,
		)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), v)

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
		_, err := store.LockSessionKeyState(testUser1, testKeyA, SessionKeyKindAppSession)
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
