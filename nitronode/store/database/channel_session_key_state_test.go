package database

import (
	"strings"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAsset1 = "usdc"
	testAsset2 = "weth"
	testAsset3 = "wbtc"
)

func TestChannelSessionKeyStateV1_TableName(t *testing.T) {
	assert.Equal(t, "channel_session_key_states_v1", ChannelSessionKeyStateV1{}.TableName())
}

func TestChannelSessionKeyAssetV1_TableName(t *testing.T) {
	assert.Equal(t, "channel_session_key_assets_v1", ChannelSessionKeyAssetV1{}.TableName())
}

func TestDBStore_StoreChannelSessionKeyState(t *testing.T) {
	t.Run("Success - Store state with assets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1, testAsset2},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig123",
		}

		err := store.StoreChannelSessionKeyState(state)
		require.NoError(t, err)

		// Verify via GetLastChannelSessionKeyStates
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, testUser1, results[0].UserAddress)
		assert.Equal(t, testSessionKey, results[0].SessionKey)
		assert.Equal(t, uint64(1), results[0].Version)
		assert.Len(t, results[0].Assets, 2)
		assert.Contains(t, results[0].Assets, testAsset1)
		assert.Contains(t, results[0].Assets, testAsset2)
		assert.Equal(t, "0xsig123", results[0].UserSig)
	})

	t.Run("Success - Store state with no assets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig123",
		}

		err := store.StoreChannelSessionKeyState(state)
		require.NoError(t, err)

		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Empty(t, results[0].Assets)
	})

	t.Run("Success - Addresses and assets are lowercased", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			SessionKey:  "0xFEDCBA0987654321FEDCBA0987654321FEDCBA09",
			Version:     1,
			Assets:      []string{"0xAA00AA00AA00AA00AA00AA00AA00AA00AA00AA00"},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}

		err := store.StoreChannelSessionKeyState(state)
		require.NoError(t, err)

		// Query with mixed case - should still find it
		results, err := store.GetLastChannelSessionKeyStates("0xAbCdEf1234567890AbCdEf1234567890AbCdEf12", nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", results[0].UserAddress)
		assert.Equal(t, "0xfedcba0987654321fedcba0987654321fedcba09", results[0].SessionKey)
		assert.Contains(t, results[0].Assets, "0xaa00aa00aa00aa00aa00aa00aa00aa00aa00aa00")
	})

	t.Run("Error - Duplicate version for same user and session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig123",
		}

		err := store.StoreChannelSessionKeyState(state)
		require.NoError(t, err)

		// Same user, session key, and version should fail
		err = store.StoreChannelSessionKeyState(state)
		assert.Error(t, err)
	})

	t.Run("Success - Multiple versions for same user and session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			Assets:      []string{testAsset1, testAsset2},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		// Should return only the latest version
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, uint64(2), results[0].Version)
		assert.Len(t, results[0].Assets, 2)
	})

	t.Run("Error - Duplicate asset in same state is rejected", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1, testAsset1}, // duplicate
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}

		err := store.StoreChannelSessionKeyState(state)
		assert.Error(t, err)
	})
}

func TestDBStore_GetLastChannelSessionKeyStates(t *testing.T) {
	t.Run("Success - Returns latest state per session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Session key A: versions 1 and 2
		stateA1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA1))

		stateA2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     2,
			Assets:      []string{testAsset1, testAsset2},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA2))

		// Session key B: version 1
		stateB1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			Assets:      []string{testAsset3},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigB1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateB1))

		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Find results by session key
		var keyA, keyB *core.ChannelSessionKeyStateV1
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
		assert.Len(t, keyA.Assets, 2)

		require.NotNil(t, keyB)
		assert.Equal(t, uint64(1), keyB.Version)
		assert.Len(t, keyB.Assets, 1)
	})

	t.Run("Success - Filter by session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		stateA := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA))

		stateB := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateB))

		sessionKey := testKeyA
		results, err := store.GetLastChannelSessionKeyStates(testUser1, &sessionKey)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, testKeyA, results[0].SessionKey)
	})

	t.Run("Returns all keys including expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Non-expired key
		stateA := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA))

		// Expired key
		stateB := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateB))

		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		// Both keys returned — caller is responsible for checking expiration
		assert.Len(t, results, 2)
	})

	t.Run("Returns empty for non-existent user", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		results, err := store.GetLastChannelSessionKeyStates("0x0000000000000000000000000000000000000099", nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Does not return other users' keys", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser2,
			SessionKey:  testKeyB,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, testUser1, results[0].UserAddress)
	})
}

func TestDBStore_GetLastChannelSessionKeyVersion(t *testing.T) {
	t.Run("Success - Returns latest version number", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     5,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v5",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		version, err := store.GetLastChannelSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(5), version)
	})

	t.Run("Returns 0 for non-existent key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		version, err := store.GetLastChannelSessionKeyVersion("0x0000000000000000000000000000000000000099", "0x0000000000000000000000000000000000000098")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), version)
	})

	t.Run("Returns version even when expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     3,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsig",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		version, err := store.GetLastChannelSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), version)
	})

	t.Run("Returns latest version even if expired (newer always supersedes)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store expired version 2 (higher version but expired)
		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		// Store non-expired version 1
		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		version, err := store.GetLastChannelSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		// v2 is returned because newer version always supersedes
		assert.Equal(t, uint64(2), version)
	})
}

func TestDBStore_ValidateChannelSessionKeyForAsset(t *testing.T) {
	// Helper to compute metadata hash for a given state
	computeMetadataHash := func(t *testing.T, version uint64, assets []string, expiresAt time.Time) string {
		t.Helper()
		hash, err := core.GetChannelSessionKeyAuthMetadataHashV1(version, assets, expiresAt.Unix())
		require.NoError(t, err)
		return strings.ToLower(hash.Hex())
	}

	t.Run("Success - Valid session key, asset, and metadata hash", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1, testAsset2}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		metadataHash := computeMetadataHash(t, 1, assets, expiresAt)

		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, metadataHash)
		require.NoError(t, err)
		assert.True(t, valid)

		// Also valid for second asset
		valid, err = store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset2, metadataHash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Failure - Asset not in allowed list", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		metadataHash := computeMetadataHash(t, 1, assets, expiresAt)

		// testAsset2 is not in the allowed list
		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset2, metadataHash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Failure - Wrong metadata hash", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		wrongHash := "0x0000000000000000000000000000000000000000000000000000000000000000"

		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, wrongHash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Failure - Expired session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(-1 * time.Hour).UTC()
		assets := []string{testAsset1}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		metadataHash := computeMetadataHash(t, 1, assets, expiresAt)

		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, metadataHash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Failure - Non-latest version is not validated", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Version 1: allows testAsset1
		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		// Version 2: allows only testAsset2 (removes testAsset1)
		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			Assets:      []string{testAsset2},
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		// Use v1 metadata hash with testAsset1 - should fail because v2 is latest
		metadataHashV1 := computeMetadataHash(t, 1, []string{testAsset1}, expiresAt)
		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, metadataHashV1)
		require.NoError(t, err)
		assert.False(t, valid)

		// Use v2 metadata hash with testAsset2 - should succeed
		metadataHashV2 := computeMetadataHash(t, 2, []string{testAsset2}, expiresAt)
		valid, err = store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset2, metadataHashV2)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Failure - Non-existent session key", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, "0x0000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Failure - Wrong user address", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		metadataHash := computeMetadataHash(t, 1, assets, expiresAt)

		// Use testUser2 instead of testUser1
		valid, err := store.ValidateChannelSessionKeyForAsset(testUser2, testSessionKey, testAsset1, metadataHash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Success - Case insensitive matching", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1}

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      assets,
			ExpiresAt:   expiresAt,
			UserSig:     "0xsig123",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		metadataHash := computeMetadataHash(t, 1, assets, expiresAt)

		// Use uppercase addresses
		valid, err := store.ValidateChannelSessionKeyForAsset(
			strings.ToUpper(testUser1),
			strings.ToUpper(testSessionKey),
			strings.ToUpper(testAsset1),
			strings.ToUpper(metadataHash),
		)
		require.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestDBStore_ChannelSessionKeyState_ForeignRelations(t *testing.T) {
	t.Run("Join table records reference correct parent ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1, testAsset2},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state))

		// Retrieve the parent state's generated ID
		var parentState ChannelSessionKeyStateV1
		err := db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&parentState).Error
		require.NoError(t, err)
		require.NotEmpty(t, parentState.ID)

		// Verify asset join table records reference the parent ID
		var assetRecords []ChannelSessionKeyAssetV1
		err = db.Where("session_key_state_id = ?", parentState.ID).Find(&assetRecords).Error
		require.NoError(t, err)
		assert.Len(t, assetRecords, 2)

		assetAddrs := []string{assetRecords[0].Asset, assetRecords[1].Asset}
		assert.Contains(t, assetAddrs, testAsset1)
		assert.Contains(t, assetAddrs, testAsset2)
	})

	t.Run("Each version maintains independent asset sets", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Version 1: authorized for asset1
		state1 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v1",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state1))

		// Version 2: authorized for asset2 and asset3 (completely different set)
		state2 := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     2,
			Assets:      []string{testAsset2, testAsset3},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig_v2",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(state2))

		// Get both parent IDs
		var parentV1, parentV2 ChannelSessionKeyStateV1
		err := db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&parentV1).Error
		require.NoError(t, err)
		err = db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 2).First(&parentV2).Error
		require.NoError(t, err)
		assert.NotEqual(t, parentV1.ID, parentV2.ID)

		// Verify v1 has its own assets
		var v1Assets []ChannelSessionKeyAssetV1
		err = db.Where("session_key_state_id = ?", parentV1.ID).Find(&v1Assets).Error
		require.NoError(t, err)
		assert.Len(t, v1Assets, 1)
		assert.Equal(t, testAsset1, v1Assets[0].Asset)

		// Verify v2 has its own assets
		var v2Assets []ChannelSessionKeyAssetV1
		err = db.Where("session_key_state_id = ?", parentV2.ID).Find(&v2Assets).Error
		require.NoError(t, err)
		assert.Len(t, v2Assets, 2)

		v2AssetAddrs := []string{v2Assets[0].Asset, v2Assets[1].Asset}
		assert.Contains(t, v2AssetAddrs, testAsset2)
		assert.Contains(t, v2AssetAddrs, testAsset3)
	})

	t.Run("Preloaded relations correct in batch query", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Key A: 2 assets
		stateA := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			Assets:      []string{testAsset1, testAsset2},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA))

		// Key B: 1 asset
		stateB := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			Assets:      []string{testAsset3},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateB))

		// GetLastChannelSessionKeyStates returns both — verify preloaded relations are correct
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		var resultA, resultB *core.ChannelSessionKeyStateV1
		for i := range results {
			switch results[i].SessionKey {
			case testKeyA:
				resultA = &results[i]
			case testKeyB:
				resultB = &results[i]
			}
		}

		require.NotNil(t, resultA)
		assert.Len(t, resultA.Assets, 2)
		assert.Contains(t, resultA.Assets, testAsset1)
		assert.Contains(t, resultA.Assets, testAsset2)

		require.NotNil(t, resultB)
		assert.Len(t, resultB.Assets, 1)
		assert.Contains(t, resultB.Assets, testAsset3)
	})

	t.Run("Same asset can be used by different session key states", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Key A authorized for asset1
		stateA := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyA,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigA",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateA))

		// Key B also authorized for asset1 (same asset, different session key)
		stateB := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testKeyB,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsigB",
		}
		require.NoError(t, store.StoreChannelSessionKeyState(stateB))

		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, r := range results {
			assert.Contains(t, r.Assets, testAsset1)
		}
	})
}
