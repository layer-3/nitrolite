package app_session_v1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetAppSessions_SuccessWithParticipant(t *testing.T) {
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
	participant := "0x1234567890123456789012345678901234567890"
	participant2 := "0x9876543210987654321098765432109876543210"

	sessions := []app.AppSessionV1{
		{
			SessionID:     "session1",
			ApplicationID: "game",
			Participants: []app.AppParticipantV1{
				{WalletAddress: participant, SignatureWeight: 1},
				{WalletAddress: participant2, SignatureWeight: 1},
			},
			Quorum:      2,
			Nonce:       1,
			Status:      app.AppSessionStatusClosed,
			Version:     1,
			SessionData: "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			SessionID:     "session2",
			ApplicationID: "betting",
			Participants: []app.AppParticipantV1{
				{WalletAddress: participant, SignatureWeight: 1},
			},
			Quorum:      1,
			Nonce:       2,
			Status:      app.AppSessionStatusOpen,
			Version:     5,
			SessionData: "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	metadata := core.PaginationMetadata{
		Page:       1,
		PerPage:    10,
		TotalCount: 2,
		PageCount:  1,
	}

	// Mock expectations
	mockStore.On("GetAppSessions", (*string)(nil), &participant, app.AppSessionStatusVoid, &core.PaginationParams{}).Return(sessions, metadata, nil)
	mockStore.On("GetParticipantAllocations", "session1").Return(map[string]map[string]decimal.Decimal{}, nil)
	mockStore.On("GetParticipantAllocations", "session2").Return(map[string]map[string]decimal.Decimal{}, nil)

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{
		Participant: &participant,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_sessions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppSessions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.AppSessionsV1GetAppSessionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.AppSessions, 2)

	// Verify first session
	assert.Equal(t, "session1", response.AppSessions[0].AppSessionID)
	assert.Equal(t, "closed", response.AppSessions[0].Status)
	assert.Len(t, response.AppSessions[0].AppDefinitionV1.Participants, 2)
	assert.Equal(t, participant, response.AppSessions[0].AppDefinitionV1.Participants[0].WalletAddress)
	assert.Equal(t, uint8(2), response.AppSessions[0].AppDefinitionV1.Quorum)
	assert.Equal(t, "1", response.AppSessions[0].Version)
	assert.NotNil(t, response.AppSessions[0].SessionData)

	// Verify second session
	assert.Equal(t, "session2", response.AppSessions[1].AppSessionID)
	assert.Equal(t, "open", response.AppSessions[1].Status)
	assert.Len(t, response.AppSessions[1].AppDefinitionV1.Participants, 1)
	assert.Equal(t, uint8(1), response.AppSessions[1].AppDefinitionV1.Quorum)
	assert.Equal(t, "5", response.AppSessions[1].Version)
	assert.Nil(t, response.AppSessions[1].SessionData)

	// Verify metadata
	assert.Equal(t, uint32(1), response.Metadata.Page)
	assert.Equal(t, uint32(10), response.Metadata.PerPage)
	assert.Equal(t, uint32(2), response.Metadata.TotalCount)
	assert.Equal(t, uint32(1), response.Metadata.PageCount)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

func TestGetAppSessions_SuccessWithAppSessionID(t *testing.T) {
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
	sessionID := "session1"
	participant := "0x1234567890123456789012345678901234567890"

	sessions := []app.AppSessionV1{
		{
			SessionID:     sessionID,
			ApplicationID: "game",
			Participants: []app.AppParticipantV1{
				{WalletAddress: participant, SignatureWeight: 1},
			},
			Quorum:      1,
			Nonce:       1,
			Status:      app.AppSessionStatusClosed,
			Version:     1,
			SessionData: "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	metadata := core.PaginationMetadata{
		Page:       1,
		PerPage:    10,
		TotalCount: 1,
		PageCount:  1,
	}

	// Mock expectations
	mockStore.On("GetAppSessions", &sessionID, (*string)(nil), app.AppSessionStatusVoid, &core.PaginationParams{}).Return(sessions, metadata, nil)
	mockStore.On("GetParticipantAllocations", "session1").Return(map[string]map[string]decimal.Decimal{}, nil)

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{
		AppSessionID: &sessionID,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_sessions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppSessions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.AppSessionsV1GetAppSessionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.AppSessions, 1)
	assert.Equal(t, sessionID, response.AppSessions[0].AppSessionID)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

func TestGetAppSessions_MissingRequiredParams(t *testing.T) {
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

	// Create RPC request without app_session_id or participant
	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_sessions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppSessions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "either app_session_id or participant must be provided")

	// Verify no store calls were made
	mockStore.AssertExpectations(t)
}

func TestGetAppSessions_WithStatusFilter(t *testing.T) {
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
	participant := "0x1234567890123456789012345678901234567890"
	status := "open"

	sessions := []app.AppSessionV1{
		{
			SessionID:     "session1",
			ApplicationID: "game",
			Participants: []app.AppParticipantV1{
				{WalletAddress: participant, SignatureWeight: 1},
			},
			Quorum:      1,
			Nonce:       1,
			Status:      app.AppSessionStatusOpen,
			Version:     1,
			SessionData: "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	metadata := core.PaginationMetadata{
		Page:       1,
		PerPage:    10,
		TotalCount: 1,
		PageCount:  1,
	}

	// Mock expectations
	mockStore.On("GetAppSessions", (*string)(nil), &participant, app.AppSessionStatusOpen, &core.PaginationParams{}).Return(sessions, metadata, nil)
	mockStore.On("GetParticipantAllocations", "session1").Return(map[string]map[string]decimal.Decimal{}, nil)

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{
		Participant: &participant,
		Status:      &status,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_sessions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppSessions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.AppSessionsV1GetAppSessionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.AppSessions, 1)
	assert.Equal(t, "open", response.AppSessions[0].Status)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

func TestGetAppSessions_StoreError(t *testing.T) {
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
	participant := "0x1234567890123456789012345678901234567890"

	// Mock expectations - return error
	mockStore.On("GetAppSessions", (*string)(nil), &participant, app.AppSessionStatusVoid, &core.PaginationParams{}).Return(nil, core.PaginationMetadata{}, fmt.Errorf("database error"))

	// Create RPC request
	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{
		Participant: &participant,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "app_sessions.v1.get_app_sessions",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAppSessions(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "failed to retrieve app sessions")

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

// TestGetAppSessions_NormalizesParticipant verifies the participant filter is normalized
// before being passed to the store.
func TestGetAppSessions_NormalizesParticipant(t *testing.T) {
	mockStore := new(MockStore)

	handler := &Handler{
		useStoreInTx: func(fn StoreTxHandler) error { return fn(mockStore) },
		metrics:      metrics.NewNoopRuntimeMetricExporter(),
	}

	canonicalParticipant := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	mixedCaseParticipant := "0xABCDEFabcdefABCDEFabcdefABCDEFabcdefABCD"

	mockStore.On("GetAppSessions", (*string)(nil), &canonicalParticipant, app.AppSessionStatusVoid, &core.PaginationParams{}).
		Return([]app.AppSessionV1{}, core.PaginationMetadata{}, nil)

	reqPayload := rpc.AppSessionsV1GetAppSessionsRequest{Participant: &mixedCaseParticipant}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "app_sessions.v1.get_app_sessions", Payload: payload},
	}

	handler.GetAppSessions(ctx)

	require.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}
