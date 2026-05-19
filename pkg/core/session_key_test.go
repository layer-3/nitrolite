package core

import (
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelSessionKeySignerV1(t *testing.T) {
	t.Parallel()
	// 1. Setup User Wallet
	userSigner, userAddress := createSigner(t)

	// 2. Setup Session Key
	sessionSigner, sessionKeyAddress := createSigner(t)

	// 3. Define Metadata
	version := uint64(1)
	assets := []string{"USDC", "WETH"}
	expiresAt := time.Now().Add(1 * time.Hour).Unix()

	// 4. Compute Metadata Hash
	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(userAddress, version, assets, expiresAt)
	require.NoError(t, err)

	// 5. Pack Data for Authorization (User signs this)
	packedAuthData, err := PackChannelKeyStateV1(sessionKeyAddress, metadataHash)
	require.NoError(t, err)

	// 6. User Signs Authorization
	authSig, err := userSigner.Sign(packedAuthData)
	require.NoError(t, err)
	authSigHex := hexutil.Encode(authSig)

	// 7. Create ChannelSessionKeySignerV1
	skSigner, err := NewChannelSessionKeySignerV1(sessionSigner, metadataHash.Hex(), authSigHex)
	require.NoError(t, err)
	assert.Equal(t, ChannelSignerType_SessionKey, skSigner.Type())

	// 8. Sign arbitrary data with Session Key Signer
	dataToSign := []byte("hello world")
	signature, err := skSigner.Sign(dataToSign)
	require.NoError(t, err)

	// 9. Verify Signature using ChannelSigValidator
	validator := NewChannelSigValidator(func(walletAddr, sessionKeyAddr, metadataHashStr string) (bool, error) {
		// Mock permission check
		if !strings.EqualFold(walletAddr, userAddress) {
			return false, nil
		}
		if !strings.EqualFold(sessionKeyAddr, sessionKeyAddress) {
			return false, nil
		}
		if metadataHashStr != metadataHash.Hex() {
			return false, nil
		}
		return true, nil
	})

	recoveredWallet, err := validator.Recover(dataToSign, signature)
	require.NoError(t, err)
	assert.Equal(t, strings.ToLower(userAddress), strings.ToLower(recoveredWallet))
}

func TestValidateChannelSessionKeyStateV1(t *testing.T) {
	t.Parallel()
	userSigner, userAddress := createSigner(t)
	sessionSigner, sessionKeyAddr := createSigner(t)

	version := uint64(1)
	assets := []string{"USDC"}
	expiresAt := time.Now().Add(1 * time.Hour)

	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(userAddress, version, assets, expiresAt.Unix())
	require.NoError(t, err)

	packed, err := PackChannelKeyStateV1(sessionKeyAddr, metadataHash)
	require.NoError(t, err)

	userSig, err := userSigner.Sign(packed)
	require.NoError(t, err)

	sessionKeySig, err := sessionSigner.Sign(packed)
	require.NoError(t, err)

	state := ChannelSessionKeyStateV1{
		UserAddress:   userAddress,
		SessionKey:    sessionKeyAddr,
		Version:       version,
		Assets:        assets,
		ExpiresAt:     expiresAt,
		UserSig:       hexutil.Encode(userSig),
		SessionKeySig: hexutil.Encode(sessionKeySig),
	}

	require.NoError(t, ValidateChannelSessionKeyStateV1(state))

	// Empty session_key_sig
	stateNoKeySig := state
	stateNoKeySig.SessionKeySig = ""
	err = ValidateChannelSessionKeyStateV1(stateNoKeySig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_key_sig is required")

	// user_sig signed by wrong wallet
	wrongSigner, _ := createSigner(t)
	wrongUserSig, err := wrongSigner.Sign(packed)
	require.NoError(t, err)
	stateWrongUser := state
	stateWrongUser.UserSig = hexutil.Encode(wrongUserSig)
	err = ValidateChannelSessionKeyStateV1(stateWrongUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match wallet")

	// session_key_sig signed by wrong key
	wrongKeySigner, _ := createSigner(t)
	wrongKeySig, err := wrongKeySigner.Sign(packed)
	require.NoError(t, err)
	stateWrongKey := state
	stateWrongKey.SessionKeySig = hexutil.Encode(wrongKeySig)
	err = ValidateChannelSessionKeyStateV1(stateWrongKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_key_sig does not match session_key")

	// Tampered version (hash mismatch on recover)
	stateTampered := state
	stateTampered.Version = 2
	assert.Error(t, ValidateChannelSessionKeyStateV1(stateTampered))
}

// TestValidateChannelSessionKeyStateV1_NoReplay verifies that signatures cannot be replayed
// across (wallet, session_key) pairs. session_key binds the packed payload and user_address
// binds the metadata hash, so substituting either dimension causes signature recovery to
// yield an unrelated address.
func TestValidateChannelSessionKeyStateV1_NoReplay(t *testing.T) {
	t.Parallel()
	userSignerA, userAddressA := createSigner(t)
	_, userAddressB := createSigner(t)

	sessionSignerA, sessionKeyAddrA := createSigner(t)
	_, sessionKeyAddrB := createSigner(t)

	version := uint64(1)
	assets := []string{"USDC"}
	expiresAt := time.Now().Add(1 * time.Hour)

	metadataHashA, err := GetChannelSessionKeyAuthMetadataHashV1(userAddressA, version, assets, expiresAt.Unix())
	require.NoError(t, err)
	packedA, err := PackChannelKeyStateV1(sessionKeyAddrA, metadataHashA)
	require.NoError(t, err)

	userSigA, err := userSignerA.Sign(packedA)
	require.NoError(t, err)
	sessionKeySigA, err := sessionSignerA.Sign(packedA)
	require.NoError(t, err)

	stateA := ChannelSessionKeyStateV1{
		UserAddress:   userAddressA,
		SessionKey:    sessionKeyAddrA,
		Version:       version,
		Assets:        assets,
		ExpiresAt:     expiresAt,
		UserSig:       hexutil.Encode(userSigA),
		SessionKeySig: hexutil.Encode(sessionKeySigA),
	}
	require.NoError(t, ValidateChannelSessionKeyStateV1(stateA))

	// Cross-session_key replay: substitute sessionKeyAddrB. packed bytes diverge, both
	// recoveries yield unrelated addresses.
	stateCrossKey := stateA
	stateCrossKey.SessionKey = sessionKeyAddrB
	err = ValidateChannelSessionKeyStateV1(stateCrossKey)
	require.Error(t, err)

	// Cross-wallet replay: substitute userAddressB. metadataHash diverges, packed bytes
	// diverge, both recoveries yield unrelated addresses.
	stateCrossUser := stateA
	stateCrossUser.UserAddress = userAddressB
	err = ValidateChannelSessionKeyStateV1(stateCrossUser)
	require.Error(t, err)
}

func TestGenerateSessionKeyStateIDV1(t *testing.T) {
	t.Parallel()
	userAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	sessionKey := common.HexToAddress("0x2222222222222222222222222222222222222222")
	version := uint64(1)

	id1, err1 := GenerateSessionKeyStateIDV1(userAddr.String(), sessionKey.String(), version)
	require.NoError(t, err1)
	assert.NotEmpty(t, id1)

	// Same inputs -> Same ID
	id2, err2 := GenerateSessionKeyStateIDV1(userAddr.String(), sessionKey.String(), version)
	require.NoError(t, err2)
	assert.Equal(t, id1, id2)

	// Different version -> Different ID
	id3, err3 := GenerateSessionKeyStateIDV1(userAddr.String(), sessionKey.String(), version+1)
	require.NoError(t, err3)
	assert.NotEqual(t, id1, id3)
}

// TestPackChannelKeyStateV1_Typehash verifies the SessionKeyAuthTypehash matches the
// Solidity constant SESSION_KEY_AUTH_TYPEHASH in SessionKeyValidator.sol so that
// off-chain authorization payloads are accepted on-chain.
func TestPackChannelKeyStateV1_Typehash(t *testing.T) {
	t.Parallel()
	expected := common.HexToHash("0x251773da8b8949935ef07284d20cc8605ad7d6f4cf6b5e040ce07dae857f0b6c")
	assert.Equal(t, expected, SessionKeyAuthTypehash())
}

func TestPackChannelKeyStateV1(t *testing.T) {
	t.Parallel()
	sessionKey := "0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF"
	metadataHash := common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	packed, err := PackChannelKeyStateV1(sessionKey, metadataHash)
	require.NoError(t, err)
	assert.Len(t, packed, 96, "payload must be 96 bytes: typehash || padded address || metadataHash")

	expected := common.FromHex("0x251773da8b8949935ef07284d20cc8605ad7d6f4cf6b5e040ce07dae857f0b6c000000000000000000000000deadbeefdeadbeefdeadbeefdeadbeefdeadbeefabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	assert.Equal(t, expected, packed)
}

func TestEncodeDecodeChannelSessionKeySignature(t *testing.T) {
	t.Parallel()
	skAuth := channelSessionKeyAuthorization{
		SessionKey:    common.HexToAddress("0xSessionKey"),
		MetadataHash:  [32]byte{1, 2, 3},
		AuthSignature: []byte{4, 5, 6},
	}
	skSignature := []byte{7, 8, 9}

	encoded, err := encodeChannelSessionKeySignature(skAuth, skSignature)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decodedAuth, decodedSig, err := decodeChannelSessionKeySignature(encoded)
	require.NoError(t, err)

	assert.Equal(t, skAuth.SessionKey, decodedAuth.SessionKey)
	assert.Equal(t, skAuth.MetadataHash, decodedAuth.MetadataHash)
	assert.Equal(t, skAuth.AuthSignature, decodedAuth.AuthSignature)
	assert.Equal(t, skSignature, decodedSig)
}

func createSigner(t *testing.T) (sign.Signer, string) {
	t.Helper()
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	pkHex := hexutil.Encode(crypto.FromECDSA(pk))

	rawSigner, err := sign.NewEthereumRawSigner(pkHex)
	require.NoError(t, err)

	msgSigner, err := sign.NewEthereumMsgSignerFromRaw(rawSigner)
	require.NoError(t, err)

	return msgSigner, rawSigner.PublicKey().Address().String()
}
