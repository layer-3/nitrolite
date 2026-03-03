package app_session_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/clearnode/metrics"
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func TestCreateAppSession_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create a real test wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  12345,
	}
	sessionData := `{"test": "data"}`
	sig1 := wallet1.SignCreateRequest(t, appDef, sessionData)

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 1, // Only need 1 signature
			Nonce:  "12345",
		},
		QuorumSigs:  []string{sig1},
		SessionData: sessionData,
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.MatchedBy(func(session any) bool {
		return true // Accept any app session for now
	})).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Parse the response
	var resp rpc.AppSessionsV1CreateAppSessionResponse
	err = ctx.Response.Payload.Translate(&resp)
	require.NoError(t, err)

	// Verify response fields
	assert.NotEmpty(t, resp.AppSessionID)
	assert.Equal(t, "1", resp.Version)
	assert.Equal(t, app.AppSessionStatusOpen.String(), resp.Status)

	// Verify all mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_QuorumWithMultipleSignatures(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create real test wallets for participant1 and participant2
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := wallet2.Address
	participant3 := "0x3333333333333333333333333333333333333333"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 2},
			{WalletAddress: participant2, SignatureWeight: 1},
			{WalletAddress: participant3, SignatureWeight: 1},
		},
		Quorum: 3,
		Nonce:  12345,
	}
	sig1 := wallet1.SignCreateRequest(t, appDef, "")
	sig2 := wallet2.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 2, // Weight of 2
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1, // Weight of 1
				},
				{
					WalletAddress:   participant3,
					SignatureWeight: 1, // Weight of 1
				},
			},
			Quorum: 3, // Need total weight of 3
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			sig1, // participant1 (weight 2)
			sig2, // participant2 (weight 1)
		},
		SessionData: "",
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Verify all mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_ZeroNonce(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "0", // Zero nonce - invalid
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about nonce
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_QuorumExceedsTotalWeights(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 5, // Total weights = 2, but quorum = 5
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0x1234567890abcdef"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about quorum
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum")
	assert.Contains(t, err.Error(), "weights")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_NoSignatures(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{}, // Empty signatures
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signatures
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no signatures")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_SignatureFromNonParticipant(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create a wallet that is NOT a participant
	nonParticipantWallet := NewTestAppSessionWallet(t)
	participant1 := "0x1111111111111111111111111111111111111111"

	// Build the app.AppDefinitionV1 for signing (with participant1 only)
	appDef := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  12345,
	}
	// Sign with the non-participant wallet
	sig := nonParticipantWallet.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{sig},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about non-participant
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-participant")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_QuorumNotMet(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create a real wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"
	participant3 := "0x3333333333333333333333333333333333333333"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
			{WalletAddress: participant3, SignatureWeight: 1},
		},
		Quorum: 3,
		Nonce:  12345,
	}
	// Sign with only one wallet (gives weight 1, but need 3)
	sig1 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant3,
					SignatureWeight: 1,
				},
			},
			Quorum: 3, // Need all 3
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			sig1, // Only one signature, need 3 total weight
		},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about quorum not met
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum not met")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_DuplicateSignatures(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create a real wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum: 2,
		Nonce:  12345,
	}
	// Sign TWICE with the same wallet
	sig1 := wallet1.SignCreateRequest(t, appDef, "")
	sig2 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
				{
					WalletAddress:   participant2,
					SignatureWeight: 1,
				},
			},
			Quorum: 2, // Need both participants
			Nonce:  "12345",
		},
		QuorumSigs: []string{
			sig1, // participant1
			sig2, // participant1 again (duplicate)
		},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error - duplicate signatures shouldn't count twice
	// Should fail with "quorum not met" since only 1 weight achieved, need 2
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum not met")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_InvalidSignatureHex(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{"not-valid-hex"}, // Invalid hex string
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signature decoding
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode signature")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_SignatureRecoveryFailure(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{
					WalletAddress:   participant1,
					SignatureWeight: 1,
				},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		// Send a signature with the 0xA1 wallet prefix but invalid ECDSA data
		QuorumSigs: []string{"0xa100000000"},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	// Verify response contains error about signature recovery
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recover user wallet")

	// Verify mocks were called
	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_AppNotRegistered(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		Application: "unregistered-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  12345,
	}
	sig1 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "unregistered-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{sig1},
	}

	// GetApp returns nil (not found)
	mockStore.On("GetApp", "unregistered-app").Return(nil, nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")

	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_OwnerSigRequired(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		Application: "restricted-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  12345,
	}
	sig1 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "restricted-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{sig1},
		// No OwnerSig provided
	}

	// App requires approval (CreationApprovalNotRequired = false)
	mockStore.On("GetApp", "restricted-app").Return(&app.AppInfoV1{
		App: app.AppV1{
			ID:                          "restricted-app",
			OwnerWallet:                 "0xowneraddr",
			CreationApprovalNotRequired: false,
		},
	}, nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner_sig is required")

	mockStore.AssertExpectations(t)
}

func TestCreateAppSession_OwnerSigSuccess(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create participant and owner wallets
	wallet1 := NewTestAppSessionWallet(t)
	ownerWallet := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		Application: "restricted-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
		},
		Quorum: 1,
		Nonce:  12345,
	}
	sessionData := `{"game": "poker"}`

	// Participant signs for quorum
	sig1 := wallet1.SignCreateRequest(t, appDef, sessionData)
	// Owner signs for approval
	ownerSig := ownerWallet.SignCreateRequest(t, appDef, sessionData)

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "restricted-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs:  []string{sig1},
		SessionData: sessionData,
		OwnerSig:    ownerSig,
	}

	// App requires approval — owner wallet matches
	mockStore.On("GetApp", "restricted-app").Return(&app.AppInfoV1{
		App: app.AppV1{
			ID:                          "restricted-app",
			OwnerWallet:                 ownerWallet.Address,
			CreationApprovalNotRequired: false,
		},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.MatchedBy(func(session any) bool {
		return true
	})).Return(nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	assert.NotNil(t, ctx.Response)

	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	var resp rpc.AppSessionsV1CreateAppSessionResponse
	err = ctx.Response.Payload.Translate(&resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.AppSessionID)
	assert.Equal(t, "1", resp.Version)
	assert.Equal(t, app.AppSessionStatusOpen.String(), resp.Status)

	mockStore.AssertExpectations(t)
}
