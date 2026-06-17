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

// allAssetsMemoryStore returns a MockMemoryStore that treats every asset as supported, so
// validateSessionKeyAssets passes for arbitrary symbols in these tests.
func allAssetsMemoryStore() *MockMemoryStore {
	m := new(MockMemoryStore)
	m.On("GetAssetDecimals", mock.AnythingOfType("string")).Return(uint8(6), nil)
	return m
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	assets := []string{"usdc"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, time.Time{}, nil)
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	// 3 assets exceeds max of 2
	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"usdc", "eth", "btc"},
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

func TestChannelSubmitSessionKeyState_RejectsDuplicateAssets(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	// Two identical assets must be rejected as a duplicate.
	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"usdc", "usdc"},
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
	assert.Contains(t, respErr.Error(), "duplicate asset")
}

func TestChannelSubmitSessionKeyState_RejectsNonCanonicalAsset(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	// Assets must be submitted lowercased; a non-canonical casing is rejected.
	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"USDC"},
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
	assert.Contains(t, respErr.Error(), "non-canonical asset")
}

func TestChannelSubmitSessionKeyState_RejectsUnsupportedAsset(t *testing.T) {
	mockStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockMemoryStore.On("GetAssetDecimals", "foo").Return(uint8(0), fmt.Errorf("asset 'foo' is not supported"))
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      mockMemoryStore,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"foo"},
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
	assert.Contains(t, respErr.Error(), "unsupported asset")
	mockMemoryStore.AssertExpectations(t)
}

// An empty assets list authorizes no usable asset, so it is only meaningful as a
// revocation/deactivation state (expires_at <= now). Submitting an empty list with a future
// expires_at must be rejected so it can never become an active current version that
// authorizes nothing. The allowed past-expires_at revoke companion is
// TestChannelSubmitSessionKeyState_RevokeExistingActiveKey.
func TestChannelSubmitSessionKeyState_RejectsEmptyAssetsWithFutureExpiresAt(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      allAssetsMemoryStore(),
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
	assert.Contains(t, respErr.Error(), "assets must be non-empty unless expires_at is in the past")
	// The submission is rejected before any store interaction.
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// Exactly at max (2) should pass validation
	assets := []string{"usdc", "eth"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, time.Time{}, nil)
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
		memoryStore:      allAssetsMemoryStore(),
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

// A version-1 revoke has no prior delegation to deactivate; allowing it would let a wallet
// seed a permanent (session_key, kind) ownership claim for a key it never proved possession
// of. It must be rejected before LockSessionKeyState runs, so no seed row is ever written.
func TestChannelSubmitSessionKeyState_RevokeFirstSubmit_Rejected(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	assets := []string{"usdc"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner, sessionKeySigner)

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
	assert.Contains(t, respErr.Error(), "no prior delegation")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "StoreChannelSessionKeyState", mock.Anything)
}

// Covers the typical revocation path: an active key (latestVersion > 0, prev expires_at in
// the future) is deactivated by submitting version+1 with a past expires_at. The per-user
// cap check is short-circuited because the previous state was already active (revoke
// decreases the active count), so CountSessionKeysForUser must not be called.
func TestChannelSubmitSessionKeyState_RevokeExistingActiveKey(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:           allAssetsMemoryStore(),
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 5,
	}

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	assets := []string{"usdc"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 2, assets, expiresAt, userSigner, sessionKeySigner)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(1, prevActiveExpiresAt, nil)
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

// Covers the re-activation path: after a revoke (latestVersion > 0, prev expires_at in the
// past), submitting version+1 with a future expires_at re-activates the slot — i.e. the
// active count goes from N-1 back to N. Because the previous latest state was inactive, the
// per-user cap MUST be re-checked here so a user at the cap cannot revoke→register-new→
// reactivate to exceed it.
func TestChannelSubmitSessionKeyState_ReactivateAfterRevoke_BelowCapAllowed(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:           allAssetsMemoryStore(),
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 5,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	assets := []string{"usdc"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, assets, expiresAt, userSigner, sessionKeySigner)

	prevRevokedExpiresAt := time.Now().Add(-time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(2, prevRevokedExpiresAt, nil)
	mockStore.On("CountSessionKeysForUser", userAddress).Return(4, nil)
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

// Reactivating a revoked key when the user is already at the per-user cap must be rejected.
// Without this gate a user at the cap can revoke key A, register fresh key B into the freed
// slot, then re-submit key A with a future expires_at and end up above the cap.
func TestChannelSubmitSessionKeyState_ReactivateAfterRevoke_AtCapRejected(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeySigner := NewMockSigner()
	sessionKeyAddress := strings.ToLower(sessionKeySigner.PublicKey().Address().String())

	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:           allAssetsMemoryStore(),
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 3,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	assets := []string{"usdc"}

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, assets, expiresAt, userSigner, sessionKeySigner)

	prevRevokedExpiresAt := time.Now().Add(-time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(2, prevRevokedExpiresAt, nil)
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

func TestChannelSubmitSessionKeyState_RejectsNegativeExpiresAt(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		memoryStore:      allAssetsMemoryStore(),
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: "0x1111111111111111111111111111111111111111",
			SessionKey:  "0x3333333333333333333333333333333333333333",
			Version:     "1",
			Assets:      []string{"usdc"},
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	// Submit version 3 when latest is 0 (expects 1)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, []string{"usdc"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, time.Time{}, nil)

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
		memoryStore:           allAssetsMemoryStore(),
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 3,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"usdc"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(0, time.Time{}, nil)
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
		memoryStore:           allAssetsMemoryStore(),
		metrics:               metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs:      10,
		maxSessionKeysPerUser: 3,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// Existing key at version 4: submit version 5. Cap must NOT block updates.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 5, []string{"usdc"}, expiresAt, userSigner, sessionKeySigner)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(4, prevActiveExpiresAt, nil)
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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	// Sign with differentSigner but claim userAddress
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"usdc"}, expiresAt, differentSigner, sessionKeySigner)

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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, userAddress, 1, []string{"usdc"}, expiresAt, userSigner, userSigner)

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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// keySigner=nil → SessionKeySig field stays empty.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"usdc"}, expiresAt, userSigner, nil)

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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	// SessionKeySig from a key that does not match the declared session_key.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"usdc"}, expiresAt, userSigner, otherSigner)

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

// Wallet-only revocation: a submit with a past expires_at and NO session_key_sig is accepted
// on the strength of user_sig alone. This is the core remediation — a lost, unavailable, or
// uncooperative session key can no longer veto revocation of its own delegation.
func TestChannelSubmitSessionKeyState_RevokeUserSigOnly_Succeeds(t *testing.T) {
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

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	// keySigner=nil → SessionKeySig stays empty; the revocation path must not require it.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 2, []string{}, expiresAt, userSigner, nil)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(1, prevActiveExpiresAt, nil)
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

// On the revocation path a present-but-mismatched session_key_sig is ignored, not validated.
// The same signature would be rejected on the active path (see RejectsMismatchedSessionKeySig).
func TestChannelSubmitSessionKeyState_RevokeIgnoresMismatchedSessionKeySig(t *testing.T) {
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

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	// SessionKeySig signed by an unrelated key — would fail the active path, ignored on revoke.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 2, []string{}, expiresAt, userSigner, otherSigner)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).Return(1, prevActiveExpiresAt, nil)
	// The ignored session_key_sig must be cleared before persisting so stored revocation
	// rows never retain unverified client input.
	mockStore.On("StoreChannelSessionKeyState", mock.MatchedBy(func(s core.ChannelSessionKeyStateV1) bool {
		return s.SessionKeySig == ""
	})).Return(nil)

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

// Even on the revocation path the wallet's user_sig must be valid: a revoke signed by a key
// other than the user_address is rejected, so revocation stays a wallet-only right (not anyone's).
func TestChannelSubmitSessionKeyState_RevokeInvalidUserSig_Rejected(t *testing.T) {
	mockStore := new(MockStore)
	userSigner := NewMockSigner()
	differentSigner := NewMockSigner()
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

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	// user_sig produced by a key other than userAddress; no session_key_sig.
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 2, []string{}, expiresAt, differentSigner, nil)

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
		memoryStore:      allAssetsMemoryStore(),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{"usdc"}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindChannel).
		Return(0, time.Time{}, database.ErrSessionKeyNotAllowed)

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
