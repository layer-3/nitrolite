package channel_v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/nitronode/store/database"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// buildSignedChannelSessionKeyStateReq creates a properly signed ChannelsV1SubmitSessionKeyState request.
// Both signer (wallet UserSig) and keySigner (SessionKeySig) sign over the same
// PackChannelKeyStateV1 payload. session_key is bound into the metadata hash, so a signature
// minted for one key cannot be replayed as ownership of another. Pass nil for keySigner to
// leave SessionKeySig empty for negative-path tests.
func buildSignedChannelSessionKeyStateReq(t *testing.T, userAddress, sessionKey string, version uint64, assets []string, expiresAt time.Time, signer, keySigner sign.Signer) rpc.ChannelsV1SubmitSessionKeyStateRequest {
	t.Helper()

	if assets == nil {
		assets = []string{}
	}

	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(strings.ToLower(userAddress), version, assets, expiresAt.Unix())
	require.NoError(t, err)

	packed, err := core.PackChannelKeyStateV1(strings.ToLower(sessionKey), metadataHash)
	require.NoError(t, err)

	sig, err := signer.Sign(packed)
	require.NoError(t, err)

	state := rpc.ChannelSessionKeyStateV1{
		UserAddress: userAddress,
		SessionKey:  sessionKey,
		Version:     strconv.FormatUint(version, 10),
		Assets:      assets,
		ExpiresAt:   strconv.FormatInt(expiresAt.Unix(), 10),
		UserSig:     hexutil.Encode(sig),
	}

	if keySigner != nil {
		keySig, err := keySigner.Sign(packed)
		require.NoError(t, err)
		state.SessionKeySig = hexutil.Encode(keySig)
	}

	return rpc.ChannelsV1SubmitSessionKeyStateRequest{State: state}
}

func TestChannelSubmitSessionKeyState_Success(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	assets := []string{"USDC"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, nil)
	mockStore.On("StoreChannelSessionKeyState", mock.AnythingOfType("core.ChannelSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
	mockStore.AssertExpectations(t)
}

func TestChannelSubmitSessionKeyState_AssetsExceedsMax(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	// 3 assets exceeds max of 2
	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"USDC", "ETH", "BTC"},
			ExpiresAt:   strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:     "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "assets array exceeds maximum length of 2")
}

func TestChannelSubmitSessionKeyState_AtMaxLimit(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// Exactly at max (2) should pass validation
	assets := []string{"USDC", "ETH"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, nil)
	mockStore.On("StoreChannelSessionKeyState", mock.AnythingOfType("core.ChannelSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}

func TestChannelSubmitSessionKeyState_InvalidUserAddress(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "not-an-address",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{},
			ExpiresAt:   strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:     "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "invalid user_address")
}

func TestChannelSubmitSessionKeyState_RevokeWithPastExpiresAt(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	// expires_at in the past expresses a revoke: the same monotonic version sequence
	// is preserved, the auth path filters expires_at > now so the key is deactivated.
	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	assets := []string{"USDC"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, nil)
	mockStore.On("StoreChannelSessionKeyState", mock.AnythingOfType("core.ChannelSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}

func TestChannelSubmitSessionKeyState_RejectsNegativeExpiresAt(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{},
			ExpiresAt:   "-1",
			UserSig:     "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "expires_at must be non-negative")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestChannelSubmitSessionKeyState_MissingUserSig(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{},
			ExpiresAt:   strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:     "",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "user_sig is required")
}

func TestChannelSubmitSessionKeyState_VersionMismatch(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	// Submit version 3 when latest is 0 (expects 1)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, []string{}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), fmt.Sprintf("expected version %d, got %d", 1, 3))
	mockStore.AssertExpectations(t)
}

func TestChannelSubmitSessionKeyState_RejectsWhenAtUserCap(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 3,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"USDC"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, nil)
	mockStore.On("CountSessionKeysForUser", userAddress).Return(3, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session key limit of 3")
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "StoreChannelSessionKeyState", mock.Anything)
}

func TestChannelSubmitSessionKeyState_AllowsUpdateForExistingKeyAtCap(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 3,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// Existing key at version 4: submit version 5. Cap must NOT block updates.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 5, []string{"USDC"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(4, nil)
	mockStore.On("StoreChannelSessionKeyState", mock.AnythingOfType("core.ChannelSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "CountSessionKeysForUser", mock.Anything)
}

func TestChannelSubmitSessionKeyState_SignatureMismatch(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	differentSigner := NewMockSigner() // sign with a different key
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	// Sign with differentSigner but claim userAddress
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{}, expiresAt, differentSigner, sessionKeySigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "does not match wallet")
}

func TestChannelSubmitSessionKeyState_RejectsUserAddressEqualsSessionKey(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, userAddress, 1, []string{"USDC"}, expiresAt, userSigner, userSigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key must differ from user_address")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestChannelSubmitSessionKeyState_RejectsMissingSessionKeySig(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// keySigner=nil → SessionKeySig field stays empty.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"USDC"}, expiresAt, userSigner, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key_sig is required")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestChannelSubmitSessionKeyState_RejectsMismatchedSessionKeySig(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())
	otherSigner := NewMockSigner()

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// SessionKeySig from a key that does not match the declared session_key.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"USDC"}, expiresAt, userSigner, otherSigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key_sig does not match session_key")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestChannelSubmitSessionKeyState_RejectsForeignOwner(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"USDC"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).
		Return(0, database.ErrSessionKeyNotAllowed)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.ChannelsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key not allowed")
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "StoreChannelSessionKeyState", mock.Anything)
}
