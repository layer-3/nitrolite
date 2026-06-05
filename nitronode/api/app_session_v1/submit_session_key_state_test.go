package app_session_v1

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
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// buildSignedSessionKeyStateReq creates a properly signed SubmitSessionKeyState request.
// signer signs the wallet UserSig; keySigner signs the SessionKeySig over the same packed
// bytes. Pass nil for keySigner to omit the field (for negative-path tests).
func buildSignedSessionKeyStateReq(t *testing.T, userAddress, sessionKey string, version uint64, applicationIDs, appSessionIDs []string, expiresAt time.Time, signer, keySigner sign.Signer) rpc.AppSessionsV1SubmitSessionKeyStateRequest {
	t.Helper()

	if applicationIDs == nil {
		applicationIDs = []string{}
	}
	if appSessionIDs == nil {
		appSessionIDs = []string{}
	}

	coreState := app.AppSessionKeyStateV1{
		UserAddress:    strings.ToLower(userAddress),
		SessionKey:     strings.ToLower(sessionKey),
		Version:        version,
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      expiresAt,
	}

	packed, err := app.PackAppSessionKeyStateV1(coreState)
	require.NoError(t, err)

	sig, err := signer.Sign(packed)
	require.NoError(t, err)

	state := rpc.AppSessionKeyStateV1{
		UserAddress:    userAddress,
		SessionKey:     sessionKey,
		Version:        strconv.FormatUint(version, 10),
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      strconv.FormatInt(expiresAt.Unix(), 10),
		UserSig:        hexutil.Encode(sig),
	}

	if keySigner != nil {
		keySig, err := keySigner.Sign(packed)
		require.NoError(t, err)
		state.SessionKeySig = hexutil.Encode(keySig)
	}

	return rpc.AppSessionsV1SubmitSessionKeyStateRequest{State: state}
}

func TestSubmitSessionKeyState_Success(t *testing.T) {
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
		maxParticipants:  32,
		maxSessionData:   1024,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	appIDs := []string{"0x1111111111111111111111111111111111111111111111111111111111111111"}
	sessionIDs := []string{"0x2222222222222222222222222222222222222222222222222222222222222222"}

	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, appIDs, sessionIDs, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(0, time.Time{}, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
	mockStore.AssertExpectations(t)
}

func TestSubmitSessionKeyState_ApplicationIDsExceedsMax(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeyAddress := "0x3333333333333333333333333333333333333333"

	// 3 application_ids exceeds max of 2
	appIDs := []string{
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333333333333333333333333333",
	}

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    userAddress,
			SessionKey:     sessionKeyAddress,
			Version:        "1",
			ApplicationIDs: appIDs,
			AppSessionIDs:  []string{},
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "application_ids array exceeds maximum length of 2")
}

func TestSubmitSessionKeyState_AppSessionIDsExceedsMax(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 2,
	}

	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeyAddress := "0x3333333333333333333333333333333333333333"

	// 3 app_session_ids exceeds max of 2
	sessionIDs := []string{
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333333333333333333333333333",
	}

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    userAddress,
			SessionKey:     sessionKeyAddress,
			Version:        "1",
			ApplicationIDs: []string{},
			AppSessionIDs:  sessionIDs,
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app_session_ids array exceeds maximum length of 2")
}

func TestSubmitSessionKeyState_AtMaxLimit(t *testing.T) {
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
	appIDs := []string{
		"0x1111111111111111111111111111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222222222222222222222222222",
	}
	sessionIDs := []string{
		"0x3333333333333333333333333333333333333333333333333333333333333333",
		"0x4444444444444444444444444444444444444444444444444444444444444444",
	}

	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, appIDs, sessionIDs, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(0, time.Time{}, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}

func TestSubmitSessionKeyState_InvalidUserAddress(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    "not-an-address",
			SessionKey:     "0x3333333333333333333333333333333333333333",
			Version:        "1",
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{},
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "invalid user_address")
}

func TestSubmitSessionKeyState_RevokeWithPastExpiresAt(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, nil, nil, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(0, time.Time{}, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}

// Covers the typical revocation path: an active key (latestVersion > 0, prev expires_at in
// the future) is deactivated by submitting version+1 with a past expires_at. The per-user
// cap check is short-circuited because the previous state was already active (revoke
// decreases the active count), so CountSessionKeysForUser must not be called.
func TestSubmitSessionKeyState_RevokeExistingActiveKey(t *testing.T) {
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
		maxSessionKeysPerUser: 5,
	}

	expiresAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 2, nil, nil, expiresAt, userSigner, sessionKeySigner)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(1, prevActiveExpiresAt, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
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
func TestSubmitSessionKeyState_ReactivateAfterRevoke_BelowCapAllowed(t *testing.T) {
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
		maxSessionKeysPerUser: 5,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, nil, nil, expiresAt, userSigner, sessionKeySigner)

	prevRevokedExpiresAt := time.Now().Add(-time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(2, prevRevokedExpiresAt, nil)
	mockStore.On("CountSessionKeysForUser", userAddress).Return(4, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}

// Reactivating a revoked key when the user is already at the per-user cap must be rejected.
// Without this gate a user at the cap can revoke key A, register fresh key B into the freed
// slot, then re-submit key A with a future expires_at and end up above the cap.
func TestSubmitSessionKeyState_ReactivateAfterRevoke_AtCapRejected(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, nil, nil, expiresAt, userSigner, sessionKeySigner)

	prevRevokedExpiresAt := time.Now().Add(-time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(2, prevRevokedExpiresAt, nil)
	mockStore.On("CountSessionKeysForUser", userAddress).Return(3, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session key limit of 3")
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "StoreAppSessionKeyState", mock.Anything)
}

func TestSubmitSessionKeyState_RejectsNegativeExpiresAt(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    "0x1111111111111111111111111111111111111111",
			SessionKey:     "0x3333333333333333333333333333333333333333",
			Version:        "1",
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{},
			ExpiresAt:      "-1",
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "expires_at must be non-negative")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_MissingUserSig(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    "0x1111111111111111111111111111111111111111",
			SessionKey:     "0x3333333333333333333333333333333333333333",
			Version:        "1",
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{},
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "user_sig is required")
}

func TestSubmitSessionKeyState_VersionMismatch(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, []string{}, []string{}, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(0, time.Time{}, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), fmt.Sprintf("expected version %d, got %d", 1, 3))
	mockStore.AssertExpectations(t)
}

func TestSubmitSessionKeyState_RejectsWhenAtUserCap(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, nil, nil, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(0, time.Time{}, nil)
	mockStore.On("CountSessionKeysForUser", userAddress).Return(3, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session key limit of 3")
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "StoreAppSessionKeyState", mock.Anything)
}

func TestSubmitSessionKeyState_AllowsUpdateForExistingKeyAtCap(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 5, nil, nil, expiresAt, userSigner, sessionKeySigner)

	prevActiveExpiresAt := time.Now().Add(24 * time.Hour)
	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).Return(4, prevActiveExpiresAt, nil)
	mockStore.On("StoreAppSessionKeyState", mock.AnythingOfType("app.AppSessionKeyStateV1")).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	assert.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "CountSessionKeysForUser", mock.Anything)
}

func TestSubmitSessionKeyState_RejectsNonLowercaseApplicationID(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeyAddress := "0x3333333333333333333333333333333333333333"

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    userAddress,
			SessionKey:     sessionKeyAddress,
			Version:        "1",
			ApplicationIDs: []string{"App-1"},
			AppSessionIDs:  []string{},
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "application_ids must be lowercase, got: App-1")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_RejectsNonLowercaseAppSessionID(t *testing.T) {
	mockStore := new(MockStore)
	handler := &Handler{
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 10,
	}

	userSigner := NewMockSigner()
	userAddress := strings.ToLower(userSigner.PublicKey().Address().String())
	sessionKeyAddress := "0x3333333333333333333333333333333333333333"

	reqPayload := rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    userAddress,
			SessionKey:     sessionKeyAddress,
			Version:        "1",
			ApplicationIDs: []string{},
			AppSessionIDs:  []string{"Session-ABC"},
			ExpiresAt:      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			UserSig:        "0xdeadbeef",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app_session_ids must be lowercase, got: Session-ABC")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_SignatureMismatch(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{}, []string{}, expiresAt, differentSigner, sessionKeySigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "user_sig does not match user_address")
}

func TestSubmitSessionKeyState_RejectsUserAddressEqualsSessionKey(t *testing.T) {
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
	// Use the wallet as its own session key — must be rejected outright.
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, userAddress, 1, nil, nil, expiresAt, userSigner, userSigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key must differ from user_address")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_RejectsMissingSessionKeySig(t *testing.T) {
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
	// keySigner=nil → SessionKeySig field stays empty in the request.
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, nil, nil, expiresAt, userSigner, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key_sig is required")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_RejectsMismatchedSessionKeySig(t *testing.T) {
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
	// SessionKeySig produced by an unrelated key — declared session_key won't match the recovered address.
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, nil, nil, expiresAt, userSigner, otherSigner)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key_sig does not match session_key")
	mockStore.AssertNotCalled(t, "LockSessionKeyState", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmitSessionKeyState_RejectsForeignOwner(t *testing.T) {
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, nil, nil, expiresAt, userSigner, sessionKeySigner)

	mockStore.On("LockSessionKeyState", userAddress, sessionKeyAddress, database.SessionKeyKindAppSession).
		Return(0, time.Time{}, database.ErrSessionKeyNotAllowed)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)
	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1SubmitSessionKeyStateMethod.String(), payload),
	}

	handler.SubmitSessionKeyState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "session_key not allowed")
	mockStore.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "StoreAppSessionKeyState", mock.Anything)
}
