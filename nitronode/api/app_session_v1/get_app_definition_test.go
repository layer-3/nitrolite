package app_session_v1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetAppDefinition_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		useStoreInTx: func(fn StoreTxHandler) error {
			return fn(mockStore)
		},
		signer:        mockSigner,
		nodeAddress:   mockSigner.PublicKey().Address().String(),
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		metrics:       metrics.NewNoopRuntimeMetricExporter(),
	}

	// Test data
	sessionID := "session123"
	participant1 := "0x1234567890123456789012345678901234567890"
	participant2 := "0x9876543210987654321098765432109876543210"

	session := &app.AppSessionV1{
		SessionID:     sessionID,
		ApplicationID: "game",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum:      2,
		Nonce:       1,
		Status:      app.AppSessionStatusClosed,
		Version:     1,
		SessionData: "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockStore.On("GetAppSession", sessionID).Return(session, nil)

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppDefinitionRequest{
		AppSessionID: sessionID,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_definition",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppDefinition(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.AppSessionsV1GetAppDefinitionResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, "game", response.Definition.Application)
	assert.Len(t, response.Definition.Participants, 2)
	assert.Equal(t, participant1, response.Definition.Participants[0].WalletAddress)
	assert.Equal(t, uint8(1), response.Definition.Participants[0].SignatureWeight)
	assert.Equal(t, participant2, response.Definition.Participants[1].WalletAddress)
	assert.Equal(t, uint8(1), response.Definition.Participants[1].SignatureWeight)
	assert.Equal(t, uint8(2), response.Definition.Quorum)
	assert.Equal(t, "1", response.Definition.Nonce)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

func TestGetAppDefinition_NotFound(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		useStoreInTx: func(fn StoreTxHandler) error {
			return fn(mockStore)
		},
		signer:        mockSigner,
		nodeAddress:   mockSigner.PublicKey().Address().String(),
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		metrics:       metrics.NewNoopRuntimeMetricExporter(),
	}

	// Test data
	sessionID := "nonexistent"

	// Mock expectations - return nil (not found)
	mockStore.On("GetAppSession", sessionID).Return(nil, nil)

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppDefinitionRequest{
		AppSessionID: sessionID,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_definition",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppDefinition(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "app_session_not_found")

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}
