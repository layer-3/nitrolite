package app_session_v1

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// assertSuccess checks if the RPC context has a successful response
func assertSuccess(t *testing.T, ctx *rpc.Context) {
	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)
}

// assertError checks if the RPC context has an error response with the expected message
func assertError(t *testing.T, ctx *rpc.Context, expectedMessage string) {
	require.NotNil(t, ctx.Response)
	err := ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedMessage)
}

func TestRebalanceAppSessions_Success_TwoSessions(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create test wallets with real keys
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1.Address, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     5,
		SessionData: `{"data":"session1"}`,
	}

	session2 := &app.AppSessionV1{
		SessionID:     sessionID2,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet2.Address, SignatureWeight: 10},
		},
		Quorum:      10,
		Status:      app.AppSessionStatusOpen,
		Version:     3,
		SessionData: `{"data":"session2"}`,
	}

	// Session 1: currently has 200 USDC, will have 100 USDC (loses 100)
	currentAllocations1 := map[string]map[string]decimal.Decimal{
		wallet1.Address: {
			"USDC": decimal.NewFromInt(200),
		},
	}

	// Session 2: currently has 50 USDC, will have 150 USDC (gains 100)
	currentAllocations2 := map[string]map[string]decimal.Decimal{
		wallet2.Address: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	// Build app state updates for signing
	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      6,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1.Address, Asset: "USDC", Amount: decimal.NewFromInt(100)},
		},
		SessionData: `{"data":"session1_updated"}`,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      4,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2.Address, Asset: "USDC", Amount: decimal.NewFromInt(150)},
		},
		SessionData: `{"data":"session2_updated"}`,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "6",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet1.Address, Asset: "USDC", Amount: "100"},
					},
					SessionData: `{"data":"session1_updated"}`,
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "4",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet2.Address, Asset: "USDC", Amount: "150"},
					},
					SessionData: `{"data":"session2_updated"}`,
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	// Mock expectations for session 1
	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID1 &&
			session.Version == 6 &&
			session.SessionData == `{"data":"session1_updated"}`
	})).Return(nil).Once()

	// Mock expectations for session 2
	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID2 &&
			session.Version == 4 &&
			session.SessionData == `{"data":"session2_updated"}`
	})).Return(nil).Once()

	// Mock ledger entry and transaction recording
	mockStore.On("RecordLedgerEntry", wallet1.Address, sessionID1, "USDC", decimal.NewFromInt(-100)).Return(nil)
	mockStore.On("RecordLedgerEntry", wallet2.Address, sessionID2, "USDC", decimal.NewFromInt(100)).Return(nil)
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeRebalance && tx.Asset == "USDC"
	})).Return(nil).Twice()

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	assertSuccess(t, ctx)
	mockStore.AssertExpectations(t)

	// Verify response contains batch_id
	var response rpc.AppSessionsV1RebalanceAppSessionsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.BatchID)
}

func TestRebalanceAppSessions_Success_MultiAsset(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create test wallets with real keys
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1.Address, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	session2 := &app.AppSessionV1{
		SessionID:     sessionID2,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet2.Address, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	// Session 1: 200 USDC, 1 ETH -> 100 USDC, 1.5 ETH (loses 100 USDC, gains 0.5 ETH)
	currentAllocations1 := map[string]map[string]decimal.Decimal{
		wallet1.Address: {
			"USDC": decimal.NewFromInt(200),
			"ETH":  decimal.NewFromInt(1),
		},
	}

	// Session 2: 50 USDC, 2 ETH -> 150 USDC, 1.5 ETH (gains 100 USDC, loses 0.5 ETH)
	currentAllocations2 := map[string]map[string]decimal.Decimal{
		wallet2.Address: {
			"USDC": decimal.NewFromInt(50),
			"ETH":  decimal.NewFromInt(2),
		},
	}

	// Build app state updates for signing
	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1.Address, Asset: "USDC", Amount: decimal.NewFromInt(100)},
			{Participant: wallet1.Address, Asset: "ETH", Amount: decimal.RequireFromString("1.5")},
		},
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2.Address, Asset: "USDC", Amount: decimal.NewFromInt(150)},
			{Participant: wallet2.Address, Asset: "ETH", Amount: decimal.RequireFromString("1.5")},
		},
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet1.Address, Asset: "USDC", Amount: "100"},
						{Participant: wallet1.Address, Asset: "ETH", Amount: "1.5"},
					},
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet2.Address, Asset: "USDC", Amount: "150"},
						{Participant: wallet2.Address, Asset: "ETH", Amount: "1.5"},
					},
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	// Mock expectations
	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID1 && s.Version == 2
	})).Return(nil).Once()

	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID2 && s.Version == 2
	})).Return(nil).Once()

	// Ledger entries
	mockStore.On("RecordLedgerEntry", wallet1.Address, sessionID1, "USDC", decimal.NewFromInt(-100)).Return(nil)
	mockStore.On("RecordLedgerEntry", wallet1.Address, sessionID1, "ETH", decimal.RequireFromString("0.5")).Return(nil)
	mockStore.On("RecordLedgerEntry", wallet2.Address, sessionID2, "USDC", decimal.NewFromInt(100)).Return(nil)
	mockStore.On("RecordLedgerEntry", wallet2.Address, sessionID2, "ETH", decimal.RequireFromString("-0.5")).Return(nil)
	mockStore.On("RecordTransaction", mock.Anything).Return(nil).Times(4) // 2 assets x 2 sessions

	// Create RPC context
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	assertSuccess(t, ctx)
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_InsufficientSessions(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)

	appStateUpdate := app.AppStateUpdateV1{
		AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig1},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "rebalancing requires at least 2 sessions")
}

func TestRebalanceAppSessions_Error_InvalidIntent(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
		Intent:       app.AppStateUpdateIntentOperate, // Wrong intent
		Version:      2,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: "0x2222222222222222222222222222222222222222222222222222222222222222",
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x1111111111111111111111111111111111111111111111111111111111111111",
					Intent:       app.AppStateUpdateIntentOperate, // Wrong intent
					Version:      "2",
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: "0x2222222222222222222222222222222222222222222222222222222222222222",
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "all updates must have 'rebalance' intent")
}

func TestRebalanceAppSessions_Error_DuplicateSession(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID := "0x1111111111111111111111111111111111111111111111111111111111111111"

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID, // Duplicate
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      3,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID, // Duplicate
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "3",
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "duplicate session in rebalance")
}

func TestRebalanceAppSessions_Error_ConservationViolation(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	// Create test wallets with real keys
	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1.Address, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	session2 := &app.AppSessionV1{
		SessionID:     sessionID2,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet2.Address, SignatureWeight: 10},
		},
		Quorum:  10,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	currentAllocations1 := map[string]map[string]decimal.Decimal{
		wallet1.Address: {
			"USDC": decimal.NewFromInt(200),
		},
	}

	currentAllocations2 := map[string]map[string]decimal.Decimal{
		wallet2.Address: {
			"USDC": decimal.NewFromInt(50),
		},
	}

	// Build app state updates for signing
	// Session 1 loses 100 USDC, Session 2 gains 200 USDC (not conserved!)
	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1.Address, Asset: "USDC", Amount: decimal.NewFromInt(100)},
		},
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2.Address, Asset: "USDC", Amount: decimal.NewFromInt(250)}, // Conservation violation
		},
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet1.Address, Asset: "USDC", Amount: "100"},
					},
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
					Allocations: []rpc.AppAllocationV1{
						{Participant: wallet2.Address, Asset: "USDC", Amount: "250"}, // Conservation violation
					},
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	// Mock expectations
	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID1 && s.Version == 2
	})).Return(nil).Once()

	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(s app.AppSessionV1) bool {
		return s.SessionID == sessionID2 && s.Version == 2
	})).Return(nil).Once()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "conservation violation")
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_SessionNotFound(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	// Mock first session returns nil (not found)
	mockStore.On("GetAppSession", sessionID1).Return(nil, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "app session not found")
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_ClosedSession(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Status:        app.AppSessionStatusClosed, // Closed
		Version:       1,
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1.Address, SignatureWeight: 1},
			{WalletAddress: wallet2.Address, SignatureWeight: 1},
		},
		Quorum: 1,
	}

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "already closed")
	mockStore.AssertExpectations(t)
}

func TestRebalanceAppSessions_Error_InvalidVersion(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		true,
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Status:        app.AppSessionStatusOpen,
		Version:       5, // Current version is 5
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1.Address, SignatureWeight: 1},
			{WalletAddress: wallet2.Address, SignatureWeight: 1},
		},
		Quorum: 1,
	}

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      10, // Wrong version (should be 6)
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      2,
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "10", // Wrong version (should be 6)
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "2",
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	mockStore.On("GetApp", mock.Anything).Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil)
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	// Execute
	handler.RebalanceAppSessions(ctx)

	// Assert
	// Error case
	assertError(t, ctx, "invalid version")
	mockStore.AssertExpectations(t)
}

// TestRebalanceAppSessions_AppRegistryDisabled verifies that when appRegistryEnabled=false,
// app lookup and AllowAction are skipped but rebalance still succeeds.
func TestRebalanceAppSessions_AppRegistryDisabled(t *testing.T) {
	mockStore := new(MockStore)

	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		storeTxProvider,
		nil,
		&MockActionGateway{},
		nil,
		nil,
		nil,
		"0xNode",
		false, // appRegistryEnabled=false
		metrics.NewNoopRuntimeMetricExporter(),
		32, 1024, 256, 16,
	)

	wallet1 := NewTestAppSessionWallet(t)
	wallet2 := NewTestAppSessionWallet(t)

	sessionID1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	sessionID2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

	session1 := &app.AppSessionV1{
		SessionID:     sessionID1,
		ApplicationID: "test-app",
		Participants:  []app.AppParticipantV1{{WalletAddress: wallet1.Address, SignatureWeight: 10}},
		Quorum:        10,
		Status:        app.AppSessionStatusOpen,
		Version:       5,
	}

	session2 := &app.AppSessionV1{
		SessionID:     sessionID2,
		ApplicationID: "test-app",
		Participants:  []app.AppParticipantV1{{WalletAddress: wallet2.Address, SignatureWeight: 10}},
		Quorum:        10,
		Status:        app.AppSessionStatusOpen,
		Version:       3,
	}

	currentAllocations1 := map[string]map[string]decimal.Decimal{
		wallet1.Address: {"USDC": decimal.NewFromInt(200)},
	}
	currentAllocations2 := map[string]map[string]decimal.Decimal{
		wallet2.Address: {"USDC": decimal.NewFromInt(50)},
	}

	appStateUpdate1 := app.AppStateUpdateV1{
		AppSessionID: sessionID1,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      6,
		Allocations:  []app.AppAllocationV1{{Participant: wallet1.Address, Asset: "USDC", Amount: decimal.NewFromInt(100)}},
	}
	sig1 := wallet1.SignAppStateUpdate(t, appStateUpdate1)

	appStateUpdate2 := app.AppStateUpdateV1{
		AppSessionID: sessionID2,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      4,
		Allocations:  []app.AppAllocationV1{{Participant: wallet2.Address, Asset: "USDC", Amount: decimal.NewFromInt(150)}},
	}
	sig2 := wallet2.SignAppStateUpdate(t, appStateUpdate2)

	reqPayload := rpc.AppSessionsV1RebalanceAppSessionsRequest{
		SignedUpdates: []rpc.SignedAppStateUpdateV1{
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID1,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "6",
					Allocations:  []rpc.AppAllocationV1{{Participant: wallet1.Address, Asset: "USDC", Amount: "100"}},
				},
				QuorumSigs: []string{sig1},
			},
			{
				AppStateUpdate: rpc.AppStateUpdateV1{
					AppSessionID: sessionID2,
					Intent:       app.AppStateUpdateIntentRebalance,
					Version:      "4",
					Allocations:  []rpc.AppAllocationV1{{Participant: wallet2.Address, Asset: "USDC", Amount: "150"}},
				},
				QuorumSigs: []string{sig2},
			},
		},
	}

	// NO GetApp mock — it should not be called
	mockStore.On("GetAppSession", sessionID1).Return(session1, nil)
	mockStore.On("GetParticipantAllocations", sessionID1).Return(currentAllocations1, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID1 && session.Version == 6
	})).Return(nil).Once()

	mockStore.On("GetAppSession", sessionID2).Return(session2, nil)
	mockStore.On("GetParticipantAllocations", sessionID2).Return(currentAllocations2, nil)
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == sessionID2 && session.Version == 4
	})).Return(nil).Once()

	mockStore.On("RecordLedgerEntry", wallet1.Address, sessionID1, "USDC", decimal.NewFromInt(-100)).Return(nil)
	mockStore.On("RecordLedgerEntry", wallet2.Address, sessionID2, "USDC", decimal.NewFromInt(100)).Return(nil)
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeRebalance && tx.Asset == "USDC"
	})).Return(nil).Twice()

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "app_sessions.v1.rebalance_app_sessions", payload),
	}

	handler.RebalanceAppSessions(ctx)

	assertSuccess(t, ctx)

	// Strict: GetApp must NOT have been called
	mockStore.AssertNotCalled(t, "GetApp", mock.Anything)
	mockStore.AssertExpectations(t)
}
