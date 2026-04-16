package app_session_v1

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestSubmitDepositState_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: true,
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	// Test data - create one key for both app session and channel state signing
	userRawSigner := NewMockSigner()
	channelWalletSigner, _ := core.NewChannelDefaultSigner(userRawSigner)
	appWalletSigner, _ := app.NewAppSessionWalletSignerV1(userRawSigner)
	participant1 := strings.ToLower(userRawSigner.PublicKey().Address().String())
	participant2 := "0x2222222222222222222222222222222222222222"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	depositAmount := decimal.NewFromInt(100)
	appSessionID := "0xAppSession123"

	// Create existing app session
	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{
				WalletAddress:   participant1,
				SignatureWeight: 1,
			},
			{
				WalletAddress:   participant2,
				SignatureWeight: 1,
			},
		},
		Quorum:      1,
		Nonce:       12345,
		Status:      app.AppSessionStatusOpen,
		Version:     1,
		SessionData: "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create user's current state (before deposit)
	currentUserState := core.State{
		ID: core.GetStateID(participant1, asset, 1, 1),
		Transition: core.Transition{
			Type: core.TransitionTypeVoid,
		},
		Asset:         asset,
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming user state (with commit transition)
	incomingUserState := currentUserState.NextState()

	_, err := incomingUserState.ApplyCommitTransition(appSessionID, depositAmount)
	require.NoError(t, err)

	// Sign the incoming user state with channel wallet signer (adds 0x01 prefix)
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(*incomingUserState)
	userSig, _ := channelWalletSigner.Sign(packedUserState)
	userSigStr := userSig.String()
	incomingUserState.UserSig = &userSigStr

	// Create app state update and sign with app wallet signer (includes 0xA1 prefix for verifyQuorum)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      depositAmount,
			},
		},
		SessionData: `{"updated": "data"}`,
	}
	packedAppUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := appWalletSigner.Sign(packedAppUpdate)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2", // Next version
		Allocations: []rpc.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      depositAmount.String(),
			},
		},
		SessionData: `{"updated": "data"}`,
	}

	// Mock expectations
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Once()
	mockStore.On("CheckOpenChannel", participant1, asset).Return("0x03", true, nil).Once()
	mockStore.On("GetLastUserState", participant1, asset, false).Return(currentUserState, nil).Once()
	mockStore.On("EnsureNoOngoingStateTransitions", participant1, asset).Return(nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()

	// Mock allocations check - empty initially
	mockStore.On("GetParticipantAllocations", appSessionID).Return(
		map[string]map[string]decimal.Decimal{},
		nil,
	).Once()

	// Mock ledger entry recording
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, asset, depositAmount).Return(nil).Once()

	// Mock app session update
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == appSessionID &&
			session.Version == 2 &&
			session.SessionData == `{"updated": "data"}`
	})).Return(nil).Once()

	// Mock user state storage
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == participant1 &&
			state.Version == incomingUserState.Version &&
			state.Transition.Type == core.TransitionTypeCommit &&
			state.NodeSig != nil
	})).Return(nil).Once()

	// Mock transaction recording
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeCommit &&
			tx.Amount.Equal(depositAmount) &&
			tx.ToAccount == appSessionID
	})).Return(nil).Once()

	// Create RPC request
	rpcState := toRPCState(*incomingUserState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	// Execute
	handler.SubmitDepositState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Parse response
	var response rpc.AppSessionsV1SubmitDepositStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.StateNodeSig, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, []byte("packed"), response.StateNodeSig)

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

func TestSubmitDepositState_InvalidTransitionType(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: true,
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	// Test data
	participant1 := "0x1111111111111111111111111111111111111111"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	appSessionID := "0xAppSession123"

	// Create user state with WRONG transition type (transfer_send instead of commit)
	userState := core.State{
		ID:         core.GetStateID(participant1, asset, 1, 2),
		Asset:      asset,
		UserWallet: participant1,
		Epoch:      1,
		Version:    2,
		Transition: core.Transition{
			Type:      core.TransitionTypeTransferSend, // Wrong type!
			TxID:      "tx-id",
			AccountID: appSessionID,
			Amount:    decimal.NewFromInt(100),
		},
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(400),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(-100),
		},
	}

	// Sign the user state
	userKey, _ := crypto.GenerateKey()
	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(userState)
	userSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedUserState).Bytes(), userKey)
	userSigHex := hexutil.Encode(userSigBytes)
	userState.UserSig = &userSigHex

	// Create app state update with proper hex signature (though we'll fail before signature check)
	appSigKey, _ := crypto.GenerateKey()
	depositAmt := decimal.NewFromInt(100)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      depositAmt,
			},
		},
		SessionData: "",
	}
	packedAppStateUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := crypto.Sign(crypto.Keccak256Hash(packedAppStateUpdate).Bytes(), appSigKey)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2",
		Allocations: []rpc.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      "100",
			},
		},
	}

	// Mock expectations
	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
		},
		Quorum:  1,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Maybe()

	// Create RPC request
	rpcState := toRPCState(userState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	// Execute
	handler.SubmitDepositState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit")

	// Verify no mocks were called since we fail early
	mockStore.AssertExpectations(t)
}

func TestSubmitDepositState_QuorumNotMet(t *testing.T) {
	// Setup
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: true,
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	// Test data - create one key for both app session and channel state signing
	userRawSigner := NewMockSigner()
	channelWalletSigner, _ := core.NewChannelDefaultSigner(userRawSigner)
	appWalletSigner, _ := app.NewAppSessionWalletSignerV1(userRawSigner)
	participant1 := strings.ToLower(userRawSigner.PublicKey().Address().String())
	participant2 := "0x2222222222222222222222222222222222222222"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	depositAmount := decimal.NewFromInt(100)
	appSessionID := "0xAppSession123"

	// Create existing app session with higher quorum requirement
	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{
				WalletAddress:   participant1,
				SignatureWeight: 1,
			},
			{
				WalletAddress:   participant2,
				SignatureWeight: 1,
			},
		},
		Quorum:  2, // Need both signatures
		Nonce:   12345,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	// Create user state
	currentUserState := core.State{
		ID: core.GetStateID(participant1, asset, 1, 1),
		Transition: core.Transition{
			Type: core.TransitionTypeVoid,
		},
		Asset:         asset,
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
	}

	incomingUserState := currentUserState.NextState()

	_, err := incomingUserState.ApplyCommitTransition(appSessionID, depositAmount)
	require.NoError(t, err)

	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(*incomingUserState)
	userSig, _ := channelWalletSigner.Sign(packedUserState)
	userSigStr := userSig.String()
	incomingUserState.UserSig = &userSigStr

	// Create app state update and sign with app wallet signer (includes 0xA1 prefix for verifyQuorum)
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      depositAmount,
			},
		},
		SessionData: "",
	}
	packedAppUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := appWalletSigner.Sign(packedAppUpdate)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2",
		Allocations: []rpc.AppAllocationV1{
			{
				Participant: participant1,
				Asset:       asset,
				Amount:      depositAmount.String(),
			},
		},
	}

	// Mock expectations
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Once()
	mockStore.On("CheckOpenChannel", participant1, asset).Return("0x03", true, nil).Once()
	mockStore.On("GetLastUserState", participant1, asset, false).Return(currentUserState, nil).Once()
	mockStore.On("EnsureNoOngoingStateTransitions", participant1, asset).Return(nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()

	// Create RPC request
	rpcState := toRPCState(*incomingUserState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	// Execute
	handler.SubmitDepositState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about quorum
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quorum not met")

	// Verify all mocks were called
	mockStore.AssertExpectations(t)
}

// TestSubmitDepositState_AppRegistryDisabled verifies that when appRegistryEnabled=false,
// app lookup and AllowAction are skipped but deposit still succeeds.
func TestSubmitDepositState_AppRegistryDisabled(t *testing.T) {
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := strings.ToLower(mockSigner.PublicKey().Address().String())
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{Err: errors.New("should not be called")},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: false, // disabled
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	userRawSigner := NewMockSigner()
	channelWalletSigner, _ := core.NewChannelDefaultSigner(userRawSigner)
	appWalletSigner, _ := app.NewAppSessionWalletSignerV1(userRawSigner)
	participant1 := strings.ToLower(userRawSigner.PublicKey().Address().String())
	participant2 := "0x2222222222222222222222222222222222222222"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	depositAmount := decimal.NewFromInt(100)
	appSessionID := "0xAppSession123"

	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum:    1,
		Nonce:     12345,
		Status:    app.AppSessionStatusOpen,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	currentUserState := core.State{
		ID: core.GetStateID(participant1, asset, 1, 1),
		Transition: core.Transition{
			Type: core.TransitionTypeVoid,
		},
		Asset:         asset,
		UserWallet:    participant1,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(500),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
	}

	incomingUserState := currentUserState.NextState()
	_, err := incomingUserState.ApplyCommitTransition(appSessionID, depositAmount)
	require.NoError(t, err)

	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(*incomingUserState)
	userSig, _ := channelWalletSigner.Sign(packedUserState)
	userSigStr := userSig.String()
	incomingUserState.UserSig = &userSigStr

	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: depositAmount},
		},
		SessionData: `{"updated": "data"}`,
	}
	packedAppUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := appWalletSigner.Sign(packedAppUpdate)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2",
		Allocations: []rpc.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: depositAmount.String()},
		},
		SessionData: `{"updated": "data"}`,
	}

	// NO GetApp mock — it should not be called
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Once()
	mockStore.On("CheckOpenChannel", participant1, asset).Return("0x03", true, nil).Once()
	mockStore.On("GetLastUserState", participant1, asset, false).Return(currentUserState, nil).Once()
	mockStore.On("EnsureNoOngoingStateTransitions", participant1, asset).Return(nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()
	mockStore.On("GetParticipantAllocations", appSessionID).Return(
		map[string]map[string]decimal.Decimal{}, nil,
	).Once()
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, asset, depositAmount).Return(nil).Once()
	mockStore.On("UpdateAppSession", mock.MatchedBy(func(session app.AppSessionV1) bool {
		return session.SessionID == appSessionID && session.Version == 2
	})).Return(nil).Once()
	mockStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == participant1 && state.NodeSig != nil
	})).Return(nil).Once()
	mockStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeCommit && tx.Amount.Equal(depositAmount)
	})).Return(nil).Once()

	rpcState := toRPCState(*incomingUserState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	handler.SubmitDepositState(ctx)

	assert.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	assert.Equal(t, rpc.MsgTypeResp, ctx.Response.Type)

	// Strict: GetApp must NOT have been called
	mockStore.AssertNotCalled(t, "GetApp", mock.Anything)
	mockStore.AssertExpectations(t)
}

func TestSubmitDepositState_DuplicateAllocation_Rejected(t *testing.T) {
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: true,
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	userRawSigner := NewMockSigner()
	channelWalletSigner, _ := core.NewChannelDefaultSigner(userRawSigner)
	appWalletSigner, _ := app.NewAppSessionWalletSignerV1(userRawSigner)
	participant1 := strings.ToLower(userRawSigner.PublicKey().Address().String())
	participant2 := "0x2222222222222222222222222222222222222222"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	appSessionID := "0xAppSession123"

	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum:  1,
		Nonce:   12345,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	currentUserState := core.State{
		ID:         core.GetStateID(participant1, asset, 1, 1),
		Transition: core.Transition{Type: core.TransitionTypeVoid},
		Asset:      asset, UserWallet: participant1, Epoch: 1, Version: 1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress", BlockchainID: 1,
			UserBalance: decimal.NewFromInt(500), UserNetFlow: decimal.NewFromInt(500),
		},
	}

	depositAmount := decimal.NewFromInt(100)
	incomingUserState := currentUserState.NextState()
	_, err := incomingUserState.ApplyCommitTransition(appSessionID, depositAmount)
	require.NoError(t, err)

	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(*incomingUserState)
	userSig, _ := channelWalletSigner.Sign(packedUserState)
	userSigStr := userSig.String()
	incomingUserState.UserSig = &userSigStr

	// Duplicate (participant1, USDC) allocations — first inflates sum, second overwrites
	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: depositAmount},
			{Participant: participant1, Asset: asset, Amount: decimal.Zero}, // duplicate
		},
	}
	packedAppUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := appWalletSigner.Sign(packedAppUpdate)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2",
		Allocations: []rpc.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: depositAmount.String()},
			{Participant: participant1, Asset: asset, Amount: "0"}, // duplicate
		},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Once()
	mockStore.On("CheckOpenChannel", participant1, asset).Return("0x03", true, nil).Once()
	mockStore.On("GetLastUserState", participant1, asset, false).Return(currentUserState, nil).Once()
	mockStore.On("EnsureNoOngoingStateTransitions", participant1, asset).Return(nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(
		map[string]map[string]decimal.Decimal{}, nil,
	).Once()

	// The first (non-duplicate) allocation may be processed before the duplicate is detected,
	// so RecordLedgerEntry might be called for the first entry. Allow it.
	mockStore.On("RecordLedgerEntry", participant1, appSessionID, asset, depositAmount).Return(nil).Maybe()

	rpcState := toRPCState(*incomingUserState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	handler.SubmitDepositState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "expected error for duplicate allocation")
	assert.Contains(t, respErr.Error(), "duplicate allocation")
}

func TestSubmitDepositState_InvalidDecimalPrecision_Rejected(t *testing.T) {
	mockStore := new(MockStore)
	mockSigner := NewMockChannelSigner()
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockAssetStore := new(MockAssetStore)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		assetStore:    mockAssetStore,
		actionGateway: &MockActionGateway{},
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockStore)
		},
		signer:             mockSigner,
		nodeAddress:        nodeAddress,
		appRegistryEnabled: true,
		metrics:            metrics.NewNoopRuntimeMetricExporter(),
		maxParticipants:    32,
		maxSessionData:     1024,
		maxSessionKeyIDs:   256,
		maxSignedUpdates:   16,
	}

	userRawSigner := NewMockSigner()
	channelWalletSigner, _ := core.NewChannelDefaultSigner(userRawSigner)
	appWalletSigner, _ := app.NewAppSessionWalletSignerV1(userRawSigner)
	participant1 := strings.ToLower(userRawSigner.PublicKey().Address().String())
	participant2 := "0x2222222222222222222222222222222222222222"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	appSessionID := "0xAppSession123"

	existingAppSession := &app.AppSessionV1{
		SessionID:     appSessionID,
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: participant1, SignatureWeight: 1},
			{WalletAddress: participant2, SignatureWeight: 1},
		},
		Quorum:  1,
		Nonce:   12345,
		Status:  app.AppSessionStatusOpen,
		Version: 1,
	}

	// Amount with 7 decimal places for USDC (which has 6)
	invalidAmount, _ := decimal.NewFromString("100.1234567")
	depositAmount := invalidAmount

	currentUserState := core.State{
		ID:         core.GetStateID(participant1, asset, 1, 1),
		Transition: core.Transition{Type: core.TransitionTypeVoid},
		Asset:      asset, UserWallet: participant1, Epoch: 1, Version: 1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress", BlockchainID: 1,
			UserBalance: decimal.NewFromInt(500), UserNetFlow: decimal.NewFromInt(500),
		},
	}

	incomingUserState := currentUserState.NextState()
	_, err := incomingUserState.ApplyCommitTransition(appSessionID, depositAmount)
	require.NoError(t, err)

	mockStatePacker.On("PackState", mock.Anything).Return([]byte("packed"), nil)
	packedUserState, _ := mockStatePacker.PackState(*incomingUserState)
	userSig, _ := channelWalletSigner.Sign(packedUserState)
	userSigStr := userSig.String()
	incomingUserState.UserSig = &userSigStr

	appStateUpdateCore := app.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations: []app.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: invalidAmount},
		},
	}
	packedAppUpdate, _ := app.PackAppStateUpdateV1(appStateUpdateCore)
	appSigBytes, _ := appWalletSigner.Sign(packedAppUpdate)
	appSigHex := hexutil.Encode(appSigBytes)

	appStateUpdate := rpc.AppStateUpdateV1{
		AppSessionID: appSessionID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      "2",
		Allocations: []rpc.AppAllocationV1{
			{Participant: participant1, Asset: asset, Amount: "100.1234567"},
		},
	}

	mockStore.On("GetApp", "test-app").Return(&app.AppInfoV1{
		App: app.AppV1{ID: "test-app", OwnerWallet: "0x0000000000000000000000000000000000000001"},
	}, nil).Maybe()
	mockStore.On("GetAppSession", appSessionID).Return(existingAppSession, nil).Once()
	mockStore.On("LockUserState", participant1, asset).Return(decimal.Zero, nil).Once()
	mockStore.On("CheckOpenChannel", participant1, asset).Return("0x03", true, nil).Once()
	mockStore.On("GetLastUserState", participant1, asset, false).Return(currentUserState, nil).Once()
	mockStore.On("EnsureNoOngoingStateTransitions", participant1, asset).Return(nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockStore.On("GetParticipantAllocations", appSessionID).Return(
		map[string]map[string]decimal.Decimal{}, nil,
	).Once()

	rpcState := toRPCState(*incomingUserState)
	reqPayload := rpc.AppSessionsV1SubmitDepositStateRequest{
		AppStateUpdate: appStateUpdate,
		QuorumSigs:     []string{appSigHex},
		UserState:      rpcState,
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppSessionsV1SubmitDepositStateMethod), payload),
	}

	handler.SubmitDepositState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error for invalid decimal precision")
	assert.Contains(t, respErr.Error(), "amount exceeds maximum decimal precision")
}
