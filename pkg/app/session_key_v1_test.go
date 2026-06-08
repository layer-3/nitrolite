package app

import (
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a signer
func createTestSigner(t *testing.T) (sign.Signer, string) {
	t.Helper()
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))

	rawSigner, err1 := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err1)

	msgSigner, err2 := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err2)

	return msgSigner, rawSigner.PublicKey().Address().String()
}

func TestGenerateSessionKeyStateIDV1(t *testing.T) {
	t.Parallel()
	userAddr := "0x1111111111111111111111111111111111111111"
	sessionKey := "0x2222222222222222222222222222222222222222"
	version := uint64(1)

	id1, err1 := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version)
	require.NoError(t, err1)
	assert.NotEmpty(t, id1)

	id2, err2 := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version)
	require.NoError(t, err2)
	assert.Equal(t, id1, id2)

	id3, err3 := GenerateSessionKeyStateIDV1(userAddr, sessionKey, version+1)
	require.NoError(t, err3)
	assert.NotEqual(t, id1, id3)
}

func TestValidateAppSessionKeyStateV1(t *testing.T) {
	t.Parallel()
	userSigner, userAddress := createTestSigner(t)
	sessionSigner, sessionKeyAddr := createTestSigner(t)

	version := uint64(1)
	appSessionIDs := []string{
		"0x1111111111111111111111111111111111111111111111111111111111111111",
	}
	applicationIDs := []string{
		"0x2222222222222222222222222222222222222222222222222222222222222222",
	}
	expiresAt := time.Now().Add(1 * time.Hour)

	baseState := AppSessionKeyStateV1{
		UserAddress:    userAddress,
		SessionKey:     sessionKeyAddr,
		Version:        version,
		AppSessionIDs:  appSessionIDs,
		ApplicationIDs: applicationIDs,
		ExpiresAt:      expiresAt,
	}

	packed, err := PackAppSessionKeyStateV1(baseState)
	require.NoError(t, err)

	userSig, err := userSigner.Sign(packed)
	require.NoError(t, err)
	sessionKeySig, err := sessionSigner.Sign(packed)
	require.NoError(t, err)

	state := baseState
	state.UserSig = hexutil.Encode(userSig)
	state.SessionKeySig = hexutil.Encode(sessionKeySig)

	require.NoError(t, ValidateAppSessionKeyStateV1(state))

	// Empty session_key_sig
	stateNoKeySig := state
	stateNoKeySig.SessionKeySig = ""
	err = ValidateAppSessionKeyStateV1(stateNoKeySig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_key_sig is required")

	// user_sig signed by wrong wallet
	wrongSigner, _ := createTestSigner(t)
	wrongUserSig, err := wrongSigner.Sign(packed)
	require.NoError(t, err)
	stateWrongUser := state
	stateWrongUser.UserSig = hexutil.Encode(wrongUserSig)
	err = ValidateAppSessionKeyStateV1(stateWrongUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_sig does not match user_address")

	// session_key_sig signed by wrong key
	wrongKeySigner, _ := createTestSigner(t)
	wrongKeySig, err := wrongKeySigner.Sign(packed)
	require.NoError(t, err)
	stateWrongKey := state
	stateWrongKey.SessionKeySig = hexutil.Encode(wrongKeySig)
	err = ValidateAppSessionKeyStateV1(stateWrongKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_key_sig does not match session_key")

	// Tampered version (hash mismatch on recover)
	stateTampered := state
	stateTampered.Version = 2
	assert.Error(t, ValidateAppSessionKeyStateV1(stateTampered))

	// Cross-wallet replay: substitute a different user_address. Packed bytes diverge so
	// neither recovery yields the matching address.
	_, otherUser := createTestSigner(t)
	stateCrossUser := state
	stateCrossUser.UserAddress = otherUser
	assert.Error(t, ValidateAppSessionKeyStateV1(stateCrossUser))

	// Cross-session-key replay: substitute a different session_key.
	_, otherKey := createTestSigner(t)
	stateCrossKey := state
	stateCrossKey.SessionKey = otherKey
	assert.Error(t, ValidateAppSessionKeyStateV1(stateCrossKey))
}

func TestValidateAppSessionKeyStateUserSigV1(t *testing.T) {
	t.Parallel()
	userSigner, userAddress := createTestSigner(t)
	_, sessionKeyAddr := createTestSigner(t)

	baseState := AppSessionKeyStateV1{
		UserAddress: userAddress,
		SessionKey:  sessionKeyAddr,
		Version:     1,
		ExpiresAt:   time.Now().Add(-time.Hour), // revocation
	}

	packed, err := PackAppSessionKeyStateV1(baseState)
	require.NoError(t, err)
	userSig, err := userSigner.Sign(packed)
	require.NoError(t, err)

	state := baseState
	state.UserSig = hexutil.Encode(userSig)

	// Valid user_sig alone passes — no session_key_sig required on the revocation path.
	require.NoError(t, ValidateAppSessionKeyStateUserSigV1(state))

	// A missing session_key_sig is irrelevant here (the field stays empty above).
	// user_sig signed by the wrong wallet is rejected.
	wrongSigner, _ := createTestSigner(t)
	wrongUserSig, err := wrongSigner.Sign(packed)
	require.NoError(t, err)
	stateWrongUser := state
	stateWrongUser.UserSig = hexutil.Encode(wrongUserSig)
	err = ValidateAppSessionKeyStateUserSigV1(stateWrongUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_sig does not match user_address")

	// Tampered version diverges the packed bytes, so recovery no longer matches.
	stateTampered := state
	stateTampered.Version = 2
	assert.Error(t, ValidateAppSessionKeyStateUserSigV1(stateTampered))
}

func TestPackAppSessionKeyStateV1(t *testing.T) {
	t.Parallel()
	expiresAt := time.Unix(1739812234, 0)
	state := AppSessionKeyStateV1{
		UserAddress:    "0x1111111111111111111111111111111111111111",
		SessionKey:     "0x2222222222222222222222222222222222222222",
		Version:        1,
		ApplicationIDs: []string{"0x00000000000000000000000000000000000000000000000000000000000000a1"},
		AppSessionIDs:  []string{"0x00000000000000000000000000000000000000000000000000000000000000b1"},
		ExpiresAt:      expiresAt,
		UserSig:        "0xSig",
	}

	packed, err := PackAppSessionKeyStateV1(state)
	require.NoError(t, err)
	assert.NotEmpty(t, packed)
	assert.Len(t, packed, 32)

	// Strengthen assertion: validate content by comparing against pre-calculated hash
	// This ensures that if the packing logic changes, the test will fail.
	expectedHash := "0x6d404fa628918dbe4abec4ae2808c7ea01dc880ad4ad392ca2d0c4ce21f706c1"
	assert.Equal(t, expectedHash, hexutil.Encode(packed))
}

func TestPackAppSessionKeyStateV1_NoIDCollision(t *testing.T) {
	t.Parallel()
	base := AppSessionKeyStateV1{
		UserAddress: "0x1111111111111111111111111111111111111111",
		SessionKey:  "0x2222222222222222222222222222222222222222",
		Version:     1,
		ExpiresAt:   time.Unix(1739812234, 0),
	}

	// Human-readable (non-hex) IDs must each produce a distinct packed hash.
	ids := []string{"app-1", "app-2", "trading", "gaming", "defi"}
	hashes := make(map[string]string)
	for _, id := range ids {
		s := base
		s.ApplicationIDs = []string{id}
		packed, err := PackAppSessionKeyStateV1(s)
		require.NoError(t, err)
		h := hexutil.Encode(packed)
		if prev, seen := hashes[h]; seen {
			t.Fatalf("collision: %q and %q produced the same hash %s", prev, id, h)
		}
		hashes[h] = id
	}

	// Same uniqueness requirement holds for AppSessionIDs.
	sessionHashes := make(map[string]string)
	for _, id := range ids {
		s := base
		s.AppSessionIDs = []string{id}
		packed, err := PackAppSessionKeyStateV1(s)
		require.NoError(t, err)
		h := hexutil.Encode(packed)
		if prev, seen := sessionHashes[h]; seen {
			t.Fatalf("appSessionID collision: %q and %q produced the same hash %s", prev, id, h)
		}
		sessionHashes[h] = id
	}

	// An empty ID and a whitespace ID must also be distinct.
	sEmpty := base
	sEmpty.ApplicationIDs = []string{""}
	sSpace := base
	sSpace.ApplicationIDs = []string{" "}
	hashEmpty, err := PackAppSessionKeyStateV1(sEmpty)
	require.NoError(t, err)
	hashSpace, err := PackAppSessionKeyStateV1(sSpace)
	require.NoError(t, err)
	assert.NotEqual(t, hexutil.Encode(hashEmpty), hexutil.Encode(hashSpace))
}

func TestAppSessionSignerV1(t *testing.T) {
	t.Parallel()
	baseSigner, _ := createTestSigner(t)
	data := []byte("hello")

	t.Run("WalletSigner", func(t *testing.T) {
		t.Parallel()
		signer, err := NewAppSessionWalletSignerV1(baseSigner)
		require.NoError(t, err)

		sig, err := signer.Sign(data)
		require.NoError(t, err)
		assert.Equal(t, byte(AppSessionSignerTypeV1_Wallet), sig[0])
	})

	t.Run("SessionKeySigner", func(t *testing.T) {
		t.Parallel()
		signer, err := NewAppSessionKeySignerV1(baseSigner)
		require.NoError(t, err)

		sig, err := signer.Sign(data)
		require.NoError(t, err)
		assert.Equal(t, byte(AppSessionSignerTypeV1_SessionKey), sig[0])
	})

	t.Run("InvalidType", func(t *testing.T) {
		t.Parallel()
		_, err := newAppSessionSignerV1(0xFF, baseSigner)
		assert.Error(t, err)
	})
}

func TestAppSessionKeyValidatorV1(t *testing.T) {
	t.Parallel()
	userSigner, userAddr := createTestSigner(t)
	sessionSigner, sessionKeyAddr := createTestSigner(t)
	data := []byte("hello")

	// Setup validator
	validator := NewAppSessionKeySigValidatorV1(func(skAddr string) (string, error) {
		if strings.EqualFold(skAddr, sessionKeyAddr) {
			return userAddr, nil
		}
		return "", assert.AnError
	})

	t.Run("VerifyWalletSignature", func(t *testing.T) {
		t.Parallel()
		signer, err := NewAppSessionWalletSignerV1(userSigner)
		require.NoError(t, err)
		sig, err := signer.Sign(data)
		require.NoError(t, err)
		require.NoError(t, validator.Verify(userAddr, data, sig))

		recovered, err := validator.Recover(data, sig)
		require.NoError(t, err)
		require.Equal(t, strings.ToLower(userAddr), strings.ToLower(recovered))
	})

	t.Run("VerifySessionKeySignature", func(t *testing.T) {
		t.Parallel()
		signer, err := NewAppSessionKeySignerV1(sessionSigner)
		require.NoError(t, err)
		sig, err1 := signer.Sign(data)
		require.NoError(t, err1)
		require.NoError(t, validator.Verify(userAddr, data, sig))

		recovered, err2 := validator.Recover(data, sig)
		require.NoError(t, err2)
		require.Equal(t, strings.ToLower(userAddr), strings.ToLower(recovered))
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		t.Parallel()
		// Too short
		assert.Error(t, validator.Verify(userAddr, data, []byte{0x00}))
		// Unknown type
		require.Error(t, validator.Verify(userAddr, data, []byte{0xFF, 0x01}))
	})

	t.Run("WrongOwner", func(t *testing.T) {
		t.Parallel()
		signer, err := NewAppSessionWalletSignerV1(sessionSigner) // Signed by session key but claims to be wallet
		require.NoError(t, err)
		sig, err1 := signer.Sign(data)
		require.NoError(t, err1)

		// Should recover sessionKeyAddr, which != userAddr
		err2 := validator.Verify(userAddr, data, sig)
		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "invalid signature")
	})
}

func TestAppSessionSignerTypeV1_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "wallet", AppSessionSignerTypeV1_Wallet.String())
	assert.Equal(t, "session_key", AppSessionSignerTypeV1_SessionKey.String())
	assert.Equal(t, "unknown(255)", AppSessionSignerTypeV1(255).String())
}
