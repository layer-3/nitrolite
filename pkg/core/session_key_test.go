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
	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(version, assets, expiresAt)
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

func TestValidateChannelSessionKeyAuthSigV1(t *testing.T) {
	t.Parallel()
	// 1. Setup User Wallet
	userSigner, userAddress := createSigner(t)

	// 2. Setup Session Key
	// We just need address for validation logic, not the signer itself unless we sign with it (which we don't for auth sig)
	// But let's use createSigner for consistency
	_, sessionKeyAddr := createSigner(t)

	// 3. Define State
	version := uint64(1)
	assets := []string{"USDC"}
	expiresAt := time.Now().Add(1 * time.Hour)

	// 4. Create valid signature
	metadataHash, err := GetChannelSessionKeyAuthMetadataHashV1(version, assets, expiresAt.Unix())
	require.NoError(t, err)

	packed, err := PackChannelKeyStateV1(sessionKeyAddr, metadataHash)
	require.NoError(t, err)

	authSig, err := userSigner.Sign(packed)
	require.NoError(t, err)

	state := ChannelSessionKeyStateV1{
		UserAddress: userAddress,
		SessionKey:  sessionKeyAddr,
		Version:     version,
		Assets:      assets,
		ExpiresAt:   expiresAt,
		UserSig:     hexutil.Encode(authSig),
	}

	// 5. Validate
	err = ValidateChannelSessionKeyAuthSigV1(state)
	require.NoError(t, err)

	// 6. Test Invalid Signature (wrong signer)
	wrongSigner, _ := createSigner(t)
	wrongSig, err := wrongSigner.Sign(packed)
	require.NoError(t, err)

	state.UserSig = hexutil.Encode(wrongSig)
	err1 := ValidateChannelSessionKeyAuthSigV1(state)
	require.Error(t, err1)
	assert.Contains(t, err1.Error(), "does not match wallet")

	// 7. Test Invalid Signature (wrong data)
	state.UserSig = hexutil.Encode(authSig)                    // Reset sig
	state.Version = 2                                          // Change data
	assert.Error(t, ValidateChannelSessionKeyAuthSigV1(state)) // Hash mismatch leads to recover address mismatch
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
