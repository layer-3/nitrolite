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
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// buildSignedSessionKeyStateReq creates a properly signed SubmitSessionKeyState request.
func buildSignedSessionKeyStateReq(t *testing.T, userAddress, sessionKey string, version uint64, applicationIDs, appSessionIDs []string, expiresAt time.Time, signer sign.Signer) rpc.AppSessionsV1SubmitSessionKeyStateRequest {
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

	return rpc.AppSessionsV1SubmitSessionKeyStateRequest{
		State: rpc.AppSessionKeyStateV1{
			UserAddress:    userAddress,
			SessionKey:     sessionKey,
			Version:        strconv.FormatUint(version, 10),
			ApplicationIDs: applicationIDs,
			AppSessionIDs:  appSessionIDs,
			ExpiresAt:      strconv.FormatInt(expiresAt.Unix(), 10),
			UserSig:        hexutil.Encode(sig),
		},
	}
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
		maxSignedUpdates: 16,
	}

	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	appIDs := []string{"0x1111111111111111111111111111111111111111111111111111111111111111"}
	sessionIDs := []string{"0x2222222222222222222222222222222222222222222222222222222222222222"}

	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, appIDs, sessionIDs, expiresAt, userSigner)

	mockStore.On("GetLastAppSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)
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

	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, appIDs, sessionIDs, expiresAt, userSigner)

	mockStore.On("GetLastAppSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)
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

func TestSubmitSessionKeyState_ExpiredExpiresAt(t *testing.T) {
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
			ExpiresAt:      strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10),
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
	assert.Contains(t, respErr.Error(), "expires_at must be in the future")
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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, []string{}, []string{}, expiresAt, userSigner)

	mockStore.On("GetLastAppSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)

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
	reqPayload := buildSignedSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{}, []string{}, expiresAt, differentSigner)

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
	assert.Contains(t, respErr.Error(), "signature does not match user_address")
}
