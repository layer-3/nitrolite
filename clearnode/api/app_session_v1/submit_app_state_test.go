package app_session_v1

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubmitAppState_OperateIntent_NoRedistribution_Success(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: `{"state":"initial"}`,
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	// Build the core app state update for signing (with lowercased participant addresses)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(100)},
			{Participant: participant2, Asset: "USDC", Amount: decimal.NewFromInt(50)},
		},
		SessionData: `{"state":"updated"}`,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant2, Asset: "USDC", Amount: "50"},
			},
			SessionData: `{"state":"updated"}`,
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.SessionData == `{"state":"updated"}` && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_WithRedistribution_Success(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	// Current allocations: p1=100, p2=50 (total=150)
	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	// Build the core app state update for signing
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(75)},
			{Participant: participant2, Asset: "USDC", Amount: decimal.NewFromInt(75)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	// New allocations: p1=75, p2=75 (total=150) - redistribution!
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "75"}, // -25
				{Participant: participant2, Asset: "USDC", Amount: "75"}, // +25
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	// Expect ledger entries for the redistribution
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-25)).Return(nil).Once()
	mockStore.On("RecordLedgerEntry", participant2, appSessionID, "USDC", decimal.NewFromInt(25)).Return(nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := strings.ToLower(mockSigner.PublicKey().Address().String())

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		nodeAddress,
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	// Build the core app state update for signing
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(60)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60"}, // Withdraw 40
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-40)).Return(nil)

	// Mock expectations for channel state issuance (issueReleaseReceiverState)
	homeChannelID := "0xHomeChannel"
	existingUserState := core.State{
		Asset:         "USDC",
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			UserBalance: decimal.NewFromInt(200),
			UserNetFlow: decimal.NewFromInt(200),
		},
	}
	mockStore.On("LockUserState", participant1, "USDC").Return(decimal.Zero, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(existingUserState, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil)
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil)

	var capturedState core.State
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		capturedState = state
		return true
	})).Return(nil)

	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusOpen
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Verify node signature on stored channel state
	require.NotNil(t, capturedState.NodeSig, "Node signature should be present on stored state")
	VerifyNodeSignature(t, nodeAddress, []byte("packed"), *capturedState.NodeSig)

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_ReceiverWithEscrowLock_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	// Build the core app state update for signing
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(60)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60"},
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-40)).Return(nil)

	// Mock expectations for channel state issuance (issueReleaseReceiverState)
	homeChannelID := "0xHomeChannel"
	existingUserState := core.State{
		Asset:         "USDC",
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			UserBalance: decimal.NewFromInt(200),
			UserNetFlow: decimal.NewFromInt(200),
		},
	}

	// Last signed state has an active escrow channel
	escrowChannelID := "0xEscrowChannel456"
	lastSignedState := core.State{
		Asset:           "USDC",
		UserWallet:      participant1,
		Epoch:           1,
		Version:         1,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
	}

	mockStore.On("LockUserState", participant1, "USDC").Return(decimal.Zero, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(existingUserState, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(lastSignedState, nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because participant has an active escrow lock
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error when participant has active escrow lock")
	assert.Contains(t, respErr.Error(), "last signed state is a lock with escrow channel")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_CloseIntent_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := strings.ToLower(mockSigner.PublicKey().Address().String())

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		nodeAddress,
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	// Build the core app state update for signing
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(100)},
			{Participant: participant2, Asset: "USDC", Amount: decimal.NewFromInt(50)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentClose,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant2, Asset: "USDC", Amount: "50"},
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)

	homeChannelID := "0xHomeChannel"
	existingUserState1 := core.State{
		Asset:         "USDC",
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			UserBalance: decimal.NewFromInt(200),
			UserNetFlow: decimal.NewFromInt(200),
		},
	}
	existingUserState2 := core.State{
		Asset:         "USDC",
		UserWallet:    participant2,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			UserBalance: decimal.NewFromInt(100),
			UserNetFlow: decimal.NewFromInt(100),
		},
	}

	// Mock expectations for fund release and channel state issuance on close
	// Participant 1: 100 USDC
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-100)).Return(nil)
	mockStore.On("LockUserState", participant1, "USDC").Return(decimal.Zero, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(existingUserState1, nil)
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil)
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil)

	// Participant 2: 50 USDC
	mockStore.On("RecordLedgerEntry", participant2, appSessionID, "USDC", decimal.NewFromInt(-50)).Return(nil)
	mockStore.On("LockUserState", participant2, "USDC").Return(decimal.Zero, nil)
	mockStore.On("GetLastUserState", participant2, "USDC", false).Return(existingUserState2, nil)
	mockStore.On("GetLastUserState", participant2, "USDC", true).Return(nil, nil)

	// Capture stored states to verify node signatures
	var capturedStates []core.State
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		capturedStates = append(capturedStates, state)
		return true
	})).Return(nil)

	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.Status == app.AppSessionStatusClosed
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Verify node signatures on all stored channel states
	require.Len(t, capturedStates, 2, "Expected 2 stored states for 2 participants")
	for _, state := range capturedStates {
		require.NotNil(t, state.NodeSig, "Node signature should be present on stored state for %s", state.UserWallet)
		VerifyNodeSignature(t, nodeAddress, []byte("packed"), *state.NodeSig)
	}

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_CloseIntent_AllocationMismatch_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	// Build the core app state update for signing (with mismatched amount)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(50)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentClose,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "50"}, // Mismatch: trying to close with different amount
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because allocations don't match current state
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for close with allocation mismatch")
	assert.Contains(t, respErr.Error(), "close intent requires allocations to match current state")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_MissingAllocation_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
		participant2: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	// Build the core app state update for signing (only participant1 allocation - missing participant2)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(150)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "150"}, // Only one participant - missing participant2
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()

	// Map iteration order is non-deterministic, so participant1 might be processed before the participant2 missing error
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(50)).Return(nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because participant2 allocation is missing
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for operate with missing allocation")
	assert.Contains(t, respErr.Error(), "operate intent missing allocation")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_MissingAllocation_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
			"DAI":  decimal.NewFromInt(50),
		},
	}

	// Build the core app state update for signing (missing DAI allocation)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(60)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60"}, // Missing DAI allocation
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetAssetDecimals", "DAI").Return(uint8(18), nil).Maybe()
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil).Maybe()

	// Map iteration order is non-deterministic, so USDC might be processed before the DAI missing error
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", decimal.NewFromInt(-40)).Return(nil).Maybe()
	mockStore.On("LockUserState", participant1, "USDC").Return(decimal.Zero, nil).Maybe()
	mockStore.On("GetLastUserState", participant1, "USDC", false).Return(nil, nil).Maybe()
	mockStore.On("GetLastUserState", participant1, "USDC", true).Return(nil, nil).Maybe()
	mockStore.On("StoreUserState", mock.Anything).Return(nil).Maybe()
	mockStore.On("RecordTransaction", mock.Anything).Return(nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because DAI allocation is missing
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for withdraw with missing allocation")
	assert.Contains(t, respErr.Error(), "withdraw intent missing allocation")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_DepositIntent_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      "2",
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "deposit intent must use submit_deposit_state endpoint")
}

func TestSubmitAppState_ClosedSession_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	existingSession := &app.AppSessionV1{
		SessionID: appSessionID,
		Status:    app.AppSessionStatusClosed, // Already closed
		Version:   1,
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app session is already closed")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_InvalidVersion_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Status:        app.AppSessionStatusOpen,
		Version:       5, // Current version is 5
		Participants: []app.AppParticipantV1{
			{WalletAddress: "0x1111111111111111111111111111111111111111", SignatureWeight: 1},
		},
		Quorum: 1,
	}

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "10", // Wrong version
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "invalid app session version")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_SessionNotFound_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations:  []rpc.AppAllocationV1{},
			SessionData:  "",
		},
		QuorumSigs: []string{"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef00"},
	}

	// Mock expectations - session not found
	mockStore.On("GetAppSession", appSessionID).Return(nil, errors.New("not found"))

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "app session not found")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_InvalidDecimalPrecision_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(100),
	}

	// Build the core app state update for signing (with invalid precision amount)
	invalidAmount, _ := decimal.NewFromString("100.1234567")
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: invalidAmount},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	// Create amount with too many decimal places (7 decimals for USDC which has 6)
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100.1234567"}, // 7 decimal places
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because of invalid decimal precision
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for invalid decimal precision")
	assert.Contains(t, respErr.Error(), "invalid amount for allocation with asset USDC")
	assert.Contains(t, respErr.Error(), "amount exceeds maximum decimal precision")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_WithdrawIntent_InvalidDecimalPrecision_Rejected(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := NewHandler(
		storeTxProvider,
		mockAssetStore,
		&MockActionGateway{},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"USDC": decimal.NewFromInt(100),
		},
	}

	// Build the core app state update for signing (with invalid precision amount)
	invalidAmount, _ := decimal.NewFromString("60.1234567")
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: invalidAmount},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	// Create amount with too many decimal places (7 decimals for USDC which has 6)
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentWithdraw,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "60.1234567"}, // 7 decimal places, withdrawing 40
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	// RecordLedgerEntry will be called before validation, but then validation will fail
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "USDC", mock.Anything).Return(nil).Maybe()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should fail because of invalid decimal precision
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for invalid decimal precision")
	assert.Contains(t, respErr.Error(), "invalid withdraw amount for allocation with asset USDC")
	assert.Contains(t, respErr.Error(), "amount exceeds maximum decimal precision")

	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_RedistributeToNewParticipant_Success(t *testing.T) {
	// Test redistributing funds to a participant who didn't have any allocation before
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := wallet2.Address

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 50},
			{WalletAddress: participant2, SignatureWeight: 50},
		},
		Quorum:  100,
		Version: 1,
		Status:  app.AppSessionStatusOpen,
	}

	// Current state: participant1 has 0.015 WETH, participant2 has nothing
	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {
			"WETH": decimal.NewFromFloat(0.015),
		},
	}

	sessionBalances := map[string]decimal.Decimal{
		"WETH": decimal.NewFromFloat(0.015),
	}

	// Build the core app state update for signing
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "WETH", Amount: decimal.NewFromFloat(0.01)},
			{Participant: participant2, Asset: "WETH", Amount: decimal.NewFromFloat(0.005)},
		},
		SessionData: "",
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdateCore)

	// New state: redistribute 0.015 WETH to participant1=0.01, participant2=0.005
	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "WETH", Amount: "0.01"},
				{Participant: participant2, Asset: "WETH", Amount: "0.005"},
			},
			SessionData: "",
		},
		QuorumSigs: []string{sig1, sig2},
	}

	// Mock expectations
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "WETH").Return(uint8(18), nil)

	// Expect ledger entries:
	// 1. participant1 WETH: 0.01 - 0.015 = -0.005 (sending to participant2)
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, "WETH", decimal.NewFromFloat(-0.005)).Return(nil)
	// 2. participant2 WETH: 0.005 - 0 = 0.005 (receiving from participant1)
	mockStore.On("RecordLedgerEntry", participant2, appSessionID, "WETH", decimal.NewFromFloat(0.005)).Return(nil)

	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == appSessionID && session.Version == 2
	})).Return(nil)

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	// Execute
	handler.SubmitAppState(ctx)

	// Assert - should succeed
	require.NotNil(t, ctx.Response)
	require.Nil(t, ctx.Response.Error(), "Expected no error for valid redistribution to new participant")

	mockStore.AssertExpectations(t)
	mockAssetStore.AssertExpectations(t)
}

// TestSubmitAppState_AppRegistryDisabled verifies that when appRegistryEnabled=false,
// app lookup and AllowAction are skipped but the operate intent still succeeds.
func TestSubmitAppState_AppRegistryDisabled(t *testing.T) {
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
		&MockActionGateway{Err: errors.New("should not be called")},
		mockSigner,
		core.NewStateAdvancerV1(mockAssetStore),
		mockStatePacker,
		"0xNode",
		false, // appRegistryEnabled=false
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:      5,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: `{"state":"initial"}`,
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {"USDC": decimal.NewFromInt(100)},
		participant2: {"USDC": decimal.NewFromInt(50)},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(150),
	}

	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(100)},
			{Participant: participant2, Asset: "USDC", Amount: decimal.NewFromInt(50)},
		},
		SessionData: `{"state":"updated"}`,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant2, Asset: "USDC", Amount: "50"},
			},
			SessionData: `{"state":"updated"}`,
		},
		QuorumSigs: []string{sig1},
	}

	// NO GetApp mock — it should not be called
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.Version == 2 && session.SessionData == `{"state":"updated"}`
	})).Return(nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	handler.SubmitAppState(ctx)

	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Strict: GetApp must NOT have been called
	mockStore.AssertNotCalled(t, "GetApp", mock.Anything)
	mockStore.AssertExpectations(t)
}

func TestSubmitAppState_OperateIntent_DuplicateAllocation_Rejected(t *testing.T) {
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
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	appSessionID := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	wallet1 := NewTestAppSessionWallet(t)
	participant1 := wallet1.Address
	participant2 := "0x2222222222222222222222222222222222222222"

	existingSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 5},
			{WalletAddress: participant2, SignatureWeight: 5},
		},
		Quorum:  5,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	currentAllocations := map[string]map[string]decimal.Decimal{
		participant1: {"USDC": decimal.NewFromInt(50)},
		participant2: {"USDC": decimal.NewFromInt(50)},
	}

	// Craft a malicious payload with duplicate (participant, asset) entries.
	// The first entry inflates the sum to pass balance validation while
	// the second overwrites the per-participant map to zero out balances.
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(100)}, // inflates sum
			{Participant: participant1, Asset: "USDC", Amount: decimal.NewFromInt(0)},   // duplicate overwrites to 0
			{Participant: participant2, Asset: "USDC", Amount: decimal.NewFromInt(0)},
		},
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdateCore)

	reqPayload := rpc.AppSessionsV1SubmitAppStateRequest{
		AppStateUpdate: rpc.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentOperate,
			Version:      "2",
			Allocations: []rpc.AppAllocationV1{
				{Participant: participant1, Asset: "USDC", Amount: "100"},
				{Participant: participant1, Asset: "USDC", Amount: "0"}, // duplicate
				{Participant: participant2, Asset: "USDC", Amount: "0"},
			},
		},
		QuorumSigs: []string{sig1},
	}

	sessionBalances := map[string]decimal.Decimal{
		"USDC": decimal.NewFromInt(100),
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingSession, nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(currentAllocations, nil)
	mockStore.On("GetAppSessionBalances", appSessionID).Return(sessionBalances, nil)
	mockAssetStore.On("GetAssetDecimals", "USDC").Return(uint8(6), nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitAppStateMethod), payload),
	}

	handler.SubmitAppState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "expected error for duplicate allocation")
	assert.Contains(t, respErr.Error(), "duplicate allocation")

	// RecordLedgerEntry must never be called — the request should be rejected before ledger writes
	mockStore.AssertNotCalled(t, "RecordLedgerEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
