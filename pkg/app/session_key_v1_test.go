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
	expectedHash := "0x9fedfbcd577c5e677b95b1273e38f52ffdeee096e98f731c5455e4c73e0274aa"
	assert.Equal(t, expectedHash, hexutil.Encode(packed))
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
