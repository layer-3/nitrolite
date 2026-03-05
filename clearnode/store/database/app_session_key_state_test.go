package database

import (
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid hex addresses for deterministic ID generation via common.HexToAddress.
const (
	testUser1      = "0x1111111111111111111111111111111111111111"
	testUser2      = "0x2222222222222222222222222222222222222222"
	testSessionKey = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testKeyA       = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testKeyB       = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testApp1       = "0xcccccccccccccccccccccccccccccccccccccccc"
	testApp2       = "0xdddddddddddddddddddddddddddddddddddddd"
	testApp3       = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	testSess1      = "0xaaaa000000000000000000000000000000000001"
	testSess2      = "0xaaaa000000000000000000000000000000000002"
)

func TestAppSessionKeyStateV1_TableName(t *testing.T) {
	assert.Equal(t, "app_session_key_states_v1", AppSessionKeyStateV1{}.TableName())
}

func TestAppSessionKeyApplicationV1_TableName(t *testing.T) {
	assert.Equal(t, "app_session_key_applications_v1", AppSessionKeyApplicationV1{}.TableName())
}

func TestAppSessionKeyAppSessionIDV1_TableName(t *testing.T) {
	assert.Equal(t, "app_session_key_app_sessions_v1", AppSessionKeyAppSessionIDV1{}.TableName())
}

func TestDBStore_StoreAppSessionKeyState(t *testing.T) {
	t.Run("Success - Store session key state with application IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1, testApp2},
			AppSessionIDs:  []string{},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig123",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		// Verify via GetLastAppSessionKeyState
		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testUser1, result.UserAddress)
		assert.Equal(t, testSessionKey, result.SessionKey)
		assert.Equal(t, uint64(1), result.Version)
		assert.Len(t, result.ApplicationIDs, 2)
		assert.Contains(t, result.ApplicationIDs, testApp1)
		assert.Contains(t, result.ApplicationIDs, testApp2)
		assert.Equal(t, "0xsig123", result.UserSig)
	})

	t.Run("Success - Store session key state with app session IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{testSess1, testSess2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig123",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.AppSessionIDs, 2)
		assert.Contains(t, result.AppSessionIDs, testSess1)
		assert.Contains(t, result.AppSessionIDs, testSess2)
	})

	t.Run("Success - Store session key state with both application and session IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			AppSessionIDs:  []string{testSess1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig123",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.ApplicationIDs, 1)
		assert.Len(t, result.AppSessionIDs, 1)
	})

	t.Run("Success - Store with no application or session IDs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig123",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.ApplicationIDs)
		assert.Empty(t, result.AppSessionIDs)
	})

	t.Run("Success - Addresses are lowercased", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			SessionKey:     "0xFEDCBA0987654321FEDCBA0987654321FEDCBA09",
			Version:        1,
			ApplicationIDs: []string{"0xAA00AA00AA00AA00AA00AA00AA00AA00AA00AA00"},
			AppSessionIDs:  []string{"0xBB00BB00BB00BB00BB00BB00BB00BB00BB00BB00"},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		// Query with mixed case - should still find it
		result, err := store.GetLastAppSessionKeyState("0xAbCdEf1234567890AbCdEf1234567890AbCdEf12", "0xFeDcBa0987654321FeDcBa0987654321FeDcBa09")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", result.UserAddress)
		assert.Equal(t, "0xfedcba0987654321fedcba0987654321fedcba09", result.SessionKey)
		assert.Contains(t, result.ApplicationIDs, "0xaa00aa00aa00aa00aa00aa00aa00aa00aa00aa00")
		assert.Contains(t, result.AppSessionIDs, "0xbb00bb00bb00bb00bb00bb00bb00bb00bb00bb00")
	})

	t.Run("Error - Duplicate version for same user and session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig123",
		}

		err := store.StoreAppSessionKeyState(state)
		require.NoError(t, err)

		// Same user, session key, and version should fail
		err = store.StoreAppSessionKeyState(state)
		assert.Error(t, err)
	})
}

func TestDBStore_GetLastAppSessionKeyState(t *testing.T) {
	t.Run("Success - Returns latest version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store version 1
		state1 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		// Store version 2
		state2 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, "0xsig_v2", result.UserSig)
	})

	t.Run("Returns nil for non-existent key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetLastAppSessionKeyState("0x0000000000000000000000000000000000000099", "0x0000000000000000000000000000000000000098")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Returns expired key (newer version always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(1), result.Version)
	})

	t.Run("Returns latest version even if expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store expired version 2 (higher version but expired)
		state2 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		// Store non-expired version 1
		state1 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		// v2 is returned because newer version always supersedes, even if expired
		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, "0xsig_v2", result.UserSig)
	})
}

func TestDBStore_GetLastAppSessionKeyVersion(t *testing.T) {
	t.Run("Success - Returns latest version number", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state1 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		state2 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     5,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v5",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		version, err := store.GetLastAppSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(5), version)
	})

	t.Run("Returns 0 for non-existent key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		version, err := store.GetLastAppSessionKeyVersion("0x0000000000000000000000000000000000000099", "0x0000000000000000000000000000000000000098")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), version)
	})

	t.Run("Returns version even when expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     3,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsig",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		version, err := store.GetLastAppSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), version)
	})
}

func TestDBStore_GetLastAppSessionKeyStates(t *testing.T) {
	t.Run("Success - Returns latest state per session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Session key A: versions 1 and 2
		stateA1 := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyA,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigA1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA1))

		stateA2 := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyA,
			Version:        2,
			ApplicationIDs: []string{testApp1, testApp2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigA2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA2))

		// Session key B: version 1
		stateB1 := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyB,
			Version:        1,
			ApplicationIDs: []string{testApp3},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigB1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB1))

		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		assert.Len(t, results, 2)

		// Find results by session key
		var keyA, keyB *app.AppSessionKeyStateV1
		for i := range results {
			switch results[i].SessionKey {
			case testKeyA:
				keyA = &results[i]
			case testKeyB:
				keyB = &results[i]
			}
		}

		require.NotNil(t, keyA)
		assert.Equal(t, uint64(2), keyA.Version) // Latest version
		assert.Len(t, keyA.ApplicationIDs, 2)

		require.NotNil(t, keyB)
		assert.Equal(t, uint64(1), keyB.Version)
	})

	t.Run("Success - Filter by session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		stateA := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA))

		stateB := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB))

		sessionKey := testKeyA
		results, err := store.GetLastAppSessionKeyStates(testUser1, &sessionKey)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, testKeyA, results[0].SessionKey)
	})

	t.Run("Returns all keys including expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Non-expired key
		stateA := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA))

		// Expired key
		stateB := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB))

		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		// Both keys returned — caller is responsible for checking expiration
		assert.Len(t, results, 2)
	})

	t.Run("Returns empty for non-existent user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		results, err := store.GetLastAppSessionKeyStates("0x0000000000000000000000000000000000000099", nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Does not return other users' keys", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state1 := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		state2 := app.AppSessionKeyStateV1{
			UserAddress: testUser2,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, testUser1, results[0].UserAddress)
	})
}

func TestDBStore_AppSessionKeyState_ForeignRelations(t *testing.T) {
	t.Run("Join table records reference correct parent ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1, testApp2},
			AppSessionIDs:  []string{testSess1, testSess2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		// Retrieve the parent state's generated ID
		var parentState AppSessionKeyStateV1
		err := db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&parentState).Error
		require.NoError(t, err)
		require.NotEmpty(t, parentState.ID)

		// Verify application join table records reference the parent ID
		var appRecords []AppSessionKeyApplicationV1
		err = db.Where("session_key_state_id = ?", parentState.ID).Find(&appRecords).Error
		require.NoError(t, err)
		assert.Len(t, appRecords, 2)

		appIDs := []string{appRecords[0].ApplicationID, appRecords[1].ApplicationID}
		assert.Contains(t, appIDs, testApp1)
		assert.Contains(t, appIDs, testApp2)

		// Verify app session join table records reference the parent ID
		var sessRecords []AppSessionKeyAppSessionIDV1
		err = db.Where("session_key_state_id = ?", parentState.ID).Find(&sessRecords).Error
		require.NoError(t, err)
		assert.Len(t, sessRecords, 2)

		sessIDs := []string{sessRecords[0].AppSessionID, sessRecords[1].AppSessionID}
		assert.Contains(t, sessIDs, testSess1)
		assert.Contains(t, sessIDs, testSess2)
	})

	t.Run("Each version maintains independent relation sets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Version 1: authorized for app1 and sess1
		state1 := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			AppSessionIDs:  []string{testSess1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig_v1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		// Version 2: authorized for app2, app3 and sess2 (completely different set)
		state2 := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        2,
			ApplicationIDs: []string{testApp2, testApp3},
			AppSessionIDs:  []string{testSess2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig_v2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		// Get both parent IDs
		var parentV1, parentV2 AppSessionKeyStateV1
		err := db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&parentV1).Error
		require.NoError(t, err)
		err = db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 2).First(&parentV2).Error
		require.NoError(t, err)
		assert.NotEqual(t, parentV1.ID, parentV2.ID)

		// Verify v1 has its own application IDs
		var v1Apps []AppSessionKeyApplicationV1
		err = db.Where("session_key_state_id = ?", parentV1.ID).Find(&v1Apps).Error
		require.NoError(t, err)
		assert.Len(t, v1Apps, 1)
		assert.Equal(t, testApp1, v1Apps[0].ApplicationID)

		// Verify v2 has its own application IDs
		var v2Apps []AppSessionKeyApplicationV1
		err = db.Where("session_key_state_id = ?", parentV2.ID).Find(&v2Apps).Error
		require.NoError(t, err)
		assert.Len(t, v2Apps, 2)

		v2AppIDs := []string{v2Apps[0].ApplicationID, v2Apps[1].ApplicationID}
		assert.Contains(t, v2AppIDs, testApp2)
		assert.Contains(t, v2AppIDs, testApp3)

		// Verify v1 has its own session IDs
		var v1Sess []AppSessionKeyAppSessionIDV1
		err = db.Where("session_key_state_id = ?", parentV1.ID).Find(&v1Sess).Error
		require.NoError(t, err)
		assert.Len(t, v1Sess, 1)
		assert.Equal(t, testSess1, v1Sess[0].AppSessionID)

		// Verify v2 has its own session IDs
		var v2Sess []AppSessionKeyAppSessionIDV1
		err = db.Where("session_key_state_id = ?", parentV2.ID).Find(&v2Sess).Error
		require.NoError(t, err)
		assert.Len(t, v2Sess, 1)
		assert.Equal(t, testSess2, v2Sess[0].AppSessionID)
	})

	t.Run("Preloaded relations do not leak between states", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Key A with app1
		stateA := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyA,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			AppSessionIDs:  []string{testSess1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigA",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA))

		// Key B with app2 and app3
		stateB := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyB,
			Version:        1,
			ApplicationIDs: []string{testApp2, testApp3},
			AppSessionIDs:  []string{testSess2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigB",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB))

		// Retrieve both states and verify no cross-contamination
		resultA, err := store.GetLastAppSessionKeyState(testUser1, testKeyA)
		require.NoError(t, err)
		require.NotNil(t, resultA)
		assert.Len(t, resultA.ApplicationIDs, 1)
		assert.Contains(t, resultA.ApplicationIDs, testApp1)
		assert.Len(t, resultA.AppSessionIDs, 1)
		assert.Contains(t, resultA.AppSessionIDs, testSess1)

		resultB, err := store.GetLastAppSessionKeyState(testUser1, testKeyB)
		require.NoError(t, err)
		require.NotNil(t, resultB)
		assert.Len(t, resultB.ApplicationIDs, 2)
		assert.Contains(t, resultB.ApplicationIDs, testApp2)
		assert.Contains(t, resultB.ApplicationIDs, testApp3)
		assert.Len(t, resultB.AppSessionIDs, 1)
		assert.Contains(t, resultB.AppSessionIDs, testSess2)
	})

	t.Run("Preloaded relations correct in batch query", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Key A: 2 apps, 1 session
		stateA := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyA,
			Version:        1,
			ApplicationIDs: []string{testApp1, testApp2},
			AppSessionIDs:  []string{testSess1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigA",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA))

		// Key B: 1 app, 2 sessions
		stateB := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyB,
			Version:        1,
			ApplicationIDs: []string{testApp3},
			AppSessionIDs:  []string{testSess1, testSess2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigB",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB))

		// GetLastAppSessionKeyStates returns both — verify preloaded relations are correct
		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		var resultA, resultB *app.AppSessionKeyStateV1
		for i := range results {
			switch results[i].SessionKey {
			case testKeyA:
				resultA = &results[i]
			case testKeyB:
				resultB = &results[i]
			}
		}

		require.NotNil(t, resultA)
		assert.Len(t, resultA.ApplicationIDs, 2)
		assert.Contains(t, resultA.ApplicationIDs, testApp1)
		assert.Contains(t, resultA.ApplicationIDs, testApp2)
		assert.Len(t, resultA.AppSessionIDs, 1)
		assert.Contains(t, resultA.AppSessionIDs, testSess1)

		require.NotNil(t, resultB)
		assert.Len(t, resultB.ApplicationIDs, 1)
		assert.Contains(t, resultB.ApplicationIDs, testApp3)
		assert.Len(t, resultB.AppSessionIDs, 2)
		assert.Contains(t, resultB.AppSessionIDs, testSess1)
		assert.Contains(t, resultB.AppSessionIDs, testSess2)
	})

	t.Run("Duplicate application ID in same state is rejected", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1, testApp1}, // duplicate
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig",
		}

		err := store.StoreAppSessionKeyState(state)
		assert.Error(t, err)
	})

	t.Run("Duplicate app session ID in same state is rejected", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress:   testUser1,
			SessionKey:    testSessionKey,
			Version:       1,
			AppSessionIDs: []string{testSess1, testSess1}, // duplicate
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			UserSig:       "0xsig",
		}

		err := store.StoreAppSessionKeyState(state)
		assert.Error(t, err)
	})

	t.Run("Same application ID can be used by different session key states", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Key A authorized for app1
		stateA := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyA,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigA",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateA))

		// Key B also authorized for app1 (same app, different session key)
		stateB := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testKeyB,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsigB",
		}
		require.NoError(t, store.StoreAppSessionKeyState(stateB))

		resultA, err := store.GetLastAppSessionKeyState(testUser1, testKeyA)
		require.NoError(t, err)
		require.NotNil(t, resultA)
		assert.Contains(t, resultA.ApplicationIDs, testApp1)

		resultB, err := store.GetLastAppSessionKeyState(testUser1, testKeyB)
		require.NoError(t, err)
		require.NotNil(t, resultB)
		assert.Contains(t, resultB.ApplicationIDs, testApp1)
	})
}

func TestDBStore_GetAppSessionKeyOwner(t *testing.T) {
	t.Run("Success - Find owner by app session ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create an app session that the session key references
		appSession := app.AppSessionV1{
			SessionID:     testSess1,
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{WalletAddress: testUser1, SignatureWeight: 100},
			},
			SessionData: `{}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, store.CreateAppSession(appSession))

		// Store session key with app session ID
		state := app.AppSessionKeyStateV1{
			UserAddress:   testUser1,
			SessionKey:    testSessionKey,
			Version:       1,
			AppSessionIDs: []string{testSess1},
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			UserSig:       "0xsig123",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		owner, err := store.GetAppSessionKeyOwner(testSessionKey, testSess1)
		require.NoError(t, err)
		assert.Equal(t, testUser1, owner)
	})

	t.Run("Success - Find owner by application ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create an app session with a specific application
		appSession := app.AppSessionV1{
			SessionID:     testSess1,
			ApplicationID: testApp1,
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{WalletAddress: testUser1, SignatureWeight: 100},
			},
			SessionData: `{}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, store.CreateAppSession(appSession))

		// Store session key with application ID matching the app session's application
		state := app.AppSessionKeyStateV1{
			UserAddress:    testUser1,
			SessionKey:     testSessionKey,
			Version:        1,
			ApplicationIDs: []string{testApp1},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			UserSig:        "0xsig123",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		owner, err := store.GetAppSessionKeyOwner(testSessionKey, testSess1)
		require.NoError(t, err)
		assert.Equal(t, testUser1, owner)
	})

	t.Run("Error - No matching session key found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		_, err := store.GetAppSessionKeyOwner("0x0000000000000000000000000000000000000099", "0x0000000000000000000000000000000000000098")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active session key found")
	})

	t.Run("Error - Expired session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create an app session
		appSession := app.AppSessionV1{
			SessionID:     testSess1,
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{WalletAddress: testUser1, SignatureWeight: 100},
			},
			SessionData: `{}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, store.CreateAppSession(appSession))

		// Store expired session key
		state := app.AppSessionKeyStateV1{
			UserAddress:   testUser1,
			SessionKey:    testSessionKey,
			Version:       1,
			AppSessionIDs: []string{testSess1},
			ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired
			UserSig:       "0xsig123",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state))

		_, err := store.GetAppSessionKeyOwner(testSessionKey, testSess1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active session key found")
	})

	t.Run("Success - Returns owner from latest version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create an app session
		appSession := app.AppSessionV1{
			SessionID:     testSess1,
			ApplicationID: "poker",
			Nonce:         1,
			Participants: []app.AppParticipantV1{
				{WalletAddress: testUser1, SignatureWeight: 100},
			},
			SessionData: `{}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, store.CreateAppSession(appSession))

		// Store version 1 with the app session ID
		state1 := app.AppSessionKeyStateV1{
			UserAddress:   testUser1,
			SessionKey:    testSessionKey,
			Version:       1,
			AppSessionIDs: []string{testSess1},
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			UserSig:       "0xsig_v1",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state1))

		// Store version 2 with the app session ID
		state2 := app.AppSessionKeyStateV1{
			UserAddress:   testUser1,
			SessionKey:    testSessionKey,
			Version:       2,
			AppSessionIDs: []string{testSess1},
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			UserSig:       "0xsig_v2",
		}
		require.NoError(t, store.StoreAppSessionKeyState(state2))

		owner, err := store.GetAppSessionKeyOwner(testSessionKey, testSess1)
		require.NoError(t, err)
		assert.Equal(t, testUser1, owner)
	})
}
