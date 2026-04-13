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

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// buildSignedChannelSessionKeyStateReq creates a properly signed ChannelsV1SubmitSessionKeyState request.
func buildSignedChannelSessionKeyStateReq(t *testing.T, userAddress, sessionKey string, version uint64, assets []string, expiresAt time.Time, signer sign.Signer) rpc.ChannelsV1SubmitSessionKeyStateRequest {
	t.Helper()

	if assets == nil {
		assets = []string{}
	}

	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(version, assets, expiresAt.Unix())
	require.NoError(t, err)

	packed, err := core.PackChannelKeyStateV1(strings.ToLower(sessionKey), metadataHash)
	require.NoError(t, err)

	sig, err := signer.Sign(packed)
	require.NoError(t, err)

	return rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: rpc.ChannelSessionKeyStateV1{
			UserAddress: userAddress,
			SessionKey:  sessionKey,
			Version:     strconv.FormatUint(version, 10),
			Assets:      assets,
			ExpiresAt:   strconv.FormatInt(expiresAt.Unix(), 10),
			UserSig:     hexutil.Encode(sig),
		},
	}
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

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner)

	mockStore.On("GetLastChannelSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)
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

	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, assets, expiresAt, userSigner)

	mockStore.On("GetLastChannelSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)
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

func TestChannelSubmitSessionKeyState_ExpiredExpiresAt(t *testing.T) {
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
			ExpiresAt:   strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10),
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
	assert.Contains(t, respErr.Error(), "expires_at must be in the future")
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
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 3, []string{}, expiresAt, userSigner)

	mockStore.On("GetLastChannelSessionKeyVersion", userAddress, sessionKeyAddress).Return(uint64(0), nil)

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
	reqPayload := buildSignedChannelSessionKeyStateReq(t, userAddress, sessionKeyAddress, 1, []string{}, expiresAt, differentSigner)

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
