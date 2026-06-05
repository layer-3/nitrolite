package app_session_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestCreateAppSession_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create a real test wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create real test wallets for participant1 and participant2
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := wallet2.Address
	participant3 := "0x3333333333333333333333333333333333333333"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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
	// Zero-padded values must also be rejected: strconv.ParseUint accepts "00",
	// "000", etc. and yields 0, which used to bypass the raw-string "0" check.
	cases := []string{"0", "00", "000"}
	for _, nonce := range cases {
		t.Run("nonce="+nonce, func(t *testing.T) {
			mockStore := new(MockStore)

			storeTxProvider := func(fn StoreTxHandler) error {
				return fn(mockStore)
			}

			mockSigner := NewMockChannelSigner()
			mockAssetStore := new(MockAssetStore)
			mockStatePacker := new(MockStatePacker)

			handler := NewHandler(
				storeTxProvider,
				mockAssetStore,
				&MockActionGateway{},
				mockSigner,
				core.NewStateAdvancerV1(mockAssetStore),
				mockStatePacker,
				"0xnode",
				true,
				metrics.NewNoopRuntimeMetricExporter(),
				32, 1024, 256, 16, 100,
			)

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
					Nonce:  nonce,
				},
				QuorumSigs: []string{"0x1234567890abcdef"},
			}

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
			assert.Contains(t, err.Error(), "nonce")

			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreateAppSession_QuorumExceedsTotalWeights(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create a wallet that is NOT a participant
	nonParticipantWallet := NewTestAppSessionWallet(t)
	participant1 := "0x1111111111111111111111111111111111111111"

	// Build the app.AppDefinitionV1 for signing (with participant1 only)
	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create a real wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"
	participant3 := "0x3333333333333333333333333333333333333333"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create a real wallet for participant1
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	// Build the app.AppDefinitionV1 for signing
	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		ApplicationID: "unregistered-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		ApplicationID: "restricted-app",
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

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Create participant and owner wallets
	wallet1 := NewTestAppSessionWallet(t)
	ownerWallet := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	appDef := app.AppDefinitionV1{
		ApplicationID: "restricted-app",
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

// TestCreateAppSession_AppRegistryDisabled verifies that when appRegistryEnabled=false,
// app lookup, owner signature validation, and AllowAction are all skipped.
func TestCreateAppSession_AppRegistryDisabled(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		false, // appRegistryEnabled=false
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
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
				{WalletAddress: participant1, SignatureWeight: 1},
				{WalletAddress: participant2, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs:  []string{sig1},
		SessionData: sessionData,
	}

	// Only CreateAppSession should be called — NO GetApp, NO AllowAction
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

	// Strict: GetApp and AllowAction must NOT have been called
	mockStore.AssertNotCalled(t, "GetApp", mock.Anything)
	mockStore.AssertExpectations(t)
}

// TestCreateAppSession_TotalWeightsOver255 verifies that session creation succeeds when the
// real sum of participant weights exceeds 255 but the quorum is still achievable. With a
// uint8 accumulator the sum wraps modulo 256, which can make a valid quorum look unreachable
// at creation time (e.g. weights 200+200=400 wraps to 144, and quorum=200 falsely appears
// to exceed total). The accumulator must be at least uint16 to avoid this.
func TestCreateAppSession_TotalWeightsOver255(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Two participants each with weight 200; real total = 400, wraps to 144 in uint8.
	// Quorum = 200 is achievable (wallet1 alone covers it) but 200 > 144 would have
	// triggered a false rejection before the fix.
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 200},
			{WalletAddress: participant2, SignatureWeight: 200},
		},
		Quorum: 200,
		Nonce:  12345,
	}
	sig1 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 200},
				{WalletAddress: participant2, SignatureWeight: 200},
			},
			Quorum: 200,
			Nonce:  "12345",
		},
		QuorumSigs: []string{sig1},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error: total weights 400 with quorum 200 must be accepted, got: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
	mockStore.AssertExpectations(t)
}

// TestCreateAppSession_DuplicateParticipantAcrossCases verifies that two participant
// addresses that differ only in letter case are detected as duplicates. Without address
// normalization the duplicate-check map would key on the raw representation and accept
// the same wallet twice.
func TestCreateAppSession_DuplicateParticipantAcrossCases(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	lower := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	upper := "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: lower, SignatureWeight: 1},
				{WalletAddress: upper, SignatureWeight: 1},
			},
			Quorum: 1,
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0xdeadbeef"},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "duplicate participant address")
	mockStore.AssertNotCalled(t, "GetApp", mock.Anything)
	mockStore.AssertNotCalled(t, "CreateAppSession", mock.Anything)
}

// TestCreateAppSession_TotalWeightsWrapToZero tests the 128+128=256 case where uint8 wraps
// to exactly 0 — the most damaging overflow (any quorum > 0 appears unreachable).
func TestCreateAppSession_TotalWeightsWrapToZero(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockSigner := NewMockChannelSigner()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// 128+128=256 wraps to 0 in uint8 — any quorum > 0 would appear unreachable.
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	appDef := app.AppDefinitionV1{
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 128},
			{WalletAddress: participant2, SignatureWeight: 128},
		},
		Quorum: 128,
		Nonce:  12345,
	}
	sig1 := wallet1.SignCreateRequest(t, appDef, "")

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 128},
				{WalletAddress: participant2, SignatureWeight: 128},
			},
			Quorum: 128,
			Nonce:  "12345",
		},
		QuorumSigs: []string{sig1},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", CreationApprovalNotRequired: true},
	}, nil).Once()
	mockStore.On("CreateAppSession", mock.Anything).Return(nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("total weights 256 (128+128) with quorum 128 must be accepted, got: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
	mockStore.AssertExpectations(t)
}

// TestCreateAppSession_QuorumExceedsTotalWeights_Rejected verifies that a quorum genuinely
// larger than the real total weight is rejected. Uses small weights (100+100=200) because
// Quorum is uint8 and cannot exceed 255, so this guard cannot be exercised with total > 255.
func TestCreateAppSession_QuorumExceedsTotalWeights_Rejected(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		new(MockAssetStore),
		&MockActionGateway{},
		NewMockChannelSigner(),
		core.NewStateAdvancerV1(new(MockAssetStore)),
		new(MockStatePacker),
		"0xnode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16, 100,
	)

	// Real total = 200+200 = 400; quorum = 255 (max uint8 but still < 400, so valid).
	// quorum cannot exceed 255 because the wire type is uint8 — so we can't test quorum=401.
	// Instead test that quorum=255 (which is < 400) is accepted.
	// To test actual rejection: use quorum=255 with total weights=100+100=200 (uint8-range).
	participant1 := "0x1111111111111111111111111111111111111111"
	participant2 := "0x2222222222222222222222222222222222222222"

	reqPayload := rpc.AppSessionsV1CreateAppSessionRequest{
		Definition: rpc.AppDefinitionV1{
			Application: "test-app",
			Participants: []rpc.AppParticipantV1{
				{WalletAddress: participant1, SignatureWeight: 100},
				{WalletAddress: participant2, SignatureWeight: 100},
			},
			Quorum: 255, // quorum (255) > real total (200) → must be rejected
			Nonce:  "12345",
		},
		QuorumSigs: []string{"0xdeadbeef"},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1CreateAppSessionMethod), payload),
	}

	handler.CreateAppSession(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.Error(t, respErr)
	assert.Contains(t, respErr.Error(), "quorum")
	mockStore.AssertExpectations(t)
}
