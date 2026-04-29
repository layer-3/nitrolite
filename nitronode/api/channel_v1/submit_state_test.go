package channel_v1

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestSubmitState_InvalidUserWallet_Rejected(t *testing.T) {
	mockTxStore := new(MockStore)

	handler := &Handler{
		useStoreInTx: func(h StoreTxHandler) error { return h(mockTxStore) },
		metrics:      metrics.NewNoopRuntimeMetricExporter(),
	}

	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpc.StateV1{
			UserWallet: "0xnot-a-valid-address",
			Asset:      "USDC",
			Epoch:      "1",
			Version:    "1",
			Transition: rpc.TransitionV1{Amount: "0"},
		},
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.submit_state", Payload: payload},
	}

	handler.SubmitState(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr)
	assert.Contains(t, respErr.Error(), "invalid user_wallet")
	mockTxStore.AssertNotCalled(t, "LockUserState", mock.Anything, mock.Anything)
}

func TestSubmitState_TransferSend_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive senderWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	senderWallet := userSigner.PublicKey().Address().String()
	receiverWallet := "0x0987654321098765432109876543210987654321"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	transferAmount := decimal.NewFromInt(100)

	// Create sender's current state (before transfer)
	currentSenderState := core.State{
		ID:            core.GetStateID(senderWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    senderWallet,
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

	// Create incoming sender state (with transfer send transition)
	incomingSenderState := currentSenderState.NextState()

	// Apply the transfer send transition to update balances
	transferSendTransition, err := incomingSenderState.ApplyTransferSendTransition(receiverWallet, transferAmount)
	require.NoError(t, err)

	// Sign the incoming sender state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Once()
	packedSenderState, _ := core.PackState(*incomingSenderState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedSenderState)
	userSigStr := userSig.String()
	incomingSenderState.UserSig = &userSigStr

	// Create receiver's current state
	currentReceiverState := core.State{
		ID:            core.GetStateID(receiverWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    receiverWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(200),
			UserNetFlow:  decimal.NewFromInt(200),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Expected receiver state after transfer receive
	expectedReceiverState := currentReceiverState.NextState()
	_, err = expectedReceiverState.ApplyTransferReceiveTransition(senderWallet, transferAmount, transferSendTransition.TxID)
	require.NoError(t, err)

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("LockUserState", senderWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", senderWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", senderWallet, asset).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedSenderState, nil).Maybe()

	// For issueTransferReceiverState
	mockTxStore.On("LockUserState", receiverWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("GetLastUserState", receiverWallet, asset, false).Return(currentReceiverState, nil)
	mockTxStore.On("GetLastUserState", receiverWallet, asset, true).Return(nil, nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		// Verify receiver state
		return state.UserWallet == receiverWallet &&
			state.Version == expectedReceiverState.Version &&
			state.Transition.Type == core.TransitionTypeTransferReceive &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// For recordTransaction
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeTransfer &&
			tx.Amount.Equal(transferAmount) &&
			tx.FromAccount == senderWallet &&
			tx.ToAccount == receiverWallet
	}), mock.Anything).Return(nil)

	// For storing sender state
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		// Verify sender state
		return state.UserWallet == senderWallet &&
			state.Version == incomingSenderState.Version &&
			state.Transition.Type == core.TransitionTypeTransferSend &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingSenderState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedSenderState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_TransferSend_ReceiverWithEscrowLock_Rejected(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive senderWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	senderWallet := userSigner.PublicKey().Address().String()
	receiverWallet := "0x0987654321098765432109876543210987654321"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	transferAmount := decimal.NewFromInt(100)

	// Create sender's current state (before transfer)
	currentSenderState := core.State{
		ID:            core.GetStateID(senderWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    senderWallet,
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

	// Create incoming sender state (with transfer send transition)
	incomingSenderState := currentSenderState.NextState()

	// Apply the transfer send transition to update balances
	_, err := incomingSenderState.ApplyTransferSendTransition(receiverWallet, transferAmount)
	require.NoError(t, err)

	// Sign the incoming sender state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Once()
	packedSenderState, _ := core.PackState(*incomingSenderState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedSenderState)
	userSigStr := userSig.String()
	incomingSenderState.UserSig = &userSigStr

	// Create receiver's current state
	currentReceiverState := core.State{
		ID:            core.GetStateID(receiverWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    receiverWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(200),
			UserNetFlow:  decimal.NewFromInt(200),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Receiver's last signed state has an active escrow channel
	escrowChannelID := "0xEscrowChannel456"
	lastSignedReceiverState := core.State{
		Asset:           asset,
		UserWallet:      receiverWallet,
		Epoch:           1,
		Version:         1,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
	}

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("LockUserState", senderWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", senderWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", senderWallet, asset).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedSenderState, nil).Maybe()

	// Sender state is stored before the transition-specific logic
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == senderWallet && state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// For issueTransferReceiverState - receiver has an active escrow lock
	mockTxStore.On("LockUserState", receiverWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("GetLastUserState", receiverWallet, asset, false).Return(currentReceiverState, nil)
	mockTxStore.On("GetLastUserState", receiverWallet, asset, true).Return(lastSignedReceiverState, nil)

	// Create RPC request
	rpcState := toRPCState(*incomingSenderState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert - should fail because receiver has an active escrow lock
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error when receiver has active escrow lock")
	assert.Contains(t, respErr.Error(), "last signed state is a lock with escrow channel")

	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_TransferSend_SameWalletCaseInsensitive_Rejected(t *testing.T) {
	// Verify that the sender==receiver check is case-insensitive.
	// Ethereum addresses are hex and may differ only in case (checksummed vs lowercased).
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockTxStore)
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Derive senderWallet from a real key — Address().String() returns checksummed (mixed case)
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	senderWallet := userSigner.PublicKey().Address().String() // checksummed, e.g. "0xAbCdEf..."

	// Use uppercased hex digits for the same address so the test guarantees a case mismatch.
	// Without EqualFold, "0xAbCdEf..." != "0xABCDEF..." would bypass the self-transfer check.
	receiverWallet := "0x" + strings.ToUpper(senderWallet[2:])

	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	transferAmount := decimal.NewFromInt(100)

	currentSenderState := core.State{
		ID:            core.GetStateID(senderWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    senderWallet,
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

	incomingSenderState := currentSenderState.NextState()
	_, err := incomingSenderState.ApplyTransferSendTransition(receiverWallet, transferAmount)
	require.NoError(t, err)

	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	packedSenderState, _ := core.PackState(*incomingSenderState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedSenderState)
	userSigStr := userSig.String()
	incomingSenderState.UserSig = &userSigStr

	// Mock expectations — should reach the issueTransferReceiverState check
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockTxStore.On("LockUserState", senderWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", senderWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", senderWallet, asset).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedSenderState, nil).Maybe()

	// Sender state is stored before the transition-specific logic
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == senderWallet && state.NodeSig != nil
	}), mock.Anything).Return(nil)

	rpcState := toRPCState(*incomingSenderState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	handler.SubmitState(ctx)

	// Should fail because sender and receiver are the same address (different case)
	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.NotNil(t, respErr, "Expected error when sender and receiver are the same wallet")
	assert.Contains(t, respErr.Error(), "sender and receiver wallets are the same")

	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_EscrowLock_Success(t *testing.T) {
	t.Skip("transition is not supported yet")
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	lockAmount := decimal.NewFromInt(100)
	nonce := uint64(12345)
	challenge := uint32(86400)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
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

	// Create incoming state with escrow lock transition
	incomingState := currentState.NextState()

	// Apply the escrow lock transition to update balances
	_, err := incomingState.ApplyEscrowLockTransition(2, "0xTokenAddress", lockAmount)
	require.NoError(t, err)
	escrowChannelID := *incomingState.EscrowChannelID

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Create home channel for mocking
	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        userWallet,
		Asset:             "usdc",
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: challenge,
		Nonce:             nonce,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("GetChannelByID", homeChannelID).Return(&homeChannel, nil)
	mockMemoryStore.On("IsAssetSupported", asset, "0xTokenAddress", uint64(2)).Return(true, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.ChannelID == escrowChannelID &&
			channel.Type == core.ChannelTypeEscrow &&
			channel.UserWallet == userWallet
	})).Return(nil)
	mockTxStore.On("ScheduleInitiateEscrowWithdrawal", incomingState.ID, uint64(2)).Return(nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowLock &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(lockAmount)
	}), mock.Anything).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			state.Transition.Type == core.TransitionTypeEscrowLock &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_EscrowWithdraw_Success(t *testing.T) {
	t.Skip("transition is not supported yet")
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	escrowChannelID := "0xEscrowChannel456"
	withdrawAmount := decimal.NewFromInt(100)

	// Create user's current state (signed, with escrow ledger)
	// The last transition must be an EscrowLock for the EscrowWithdraw to be valid
	currentSignedState := core.State{
		ID: core.GetStateID(userWallet, asset, 1, 2),
		Transition: core.Transition{
			Type:      core.TransitionTypeEscrowLock,
			TxID:      "0xPreviousEscrowLockTx",
			AccountID: "",
			Amount:    withdrawAmount,
		},
		Asset:           asset,
		UserWallet:      userWallet,
		Epoch:           1,
		Version:         2,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(400),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: &core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 2,
			UserBalance:  decimal.NewFromInt(0),
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(100),
			NodeNetFlow:  decimal.NewFromInt(100),
		},
		UserSig: stringPtr("0xPreviousUserSig"),
		NodeSig: stringPtr("0xPreviousNodeSig"),
	}

	// Create incoming state with escrow withdraw transition
	incomingState := currentSignedState.NextState()

	// Apply the escrow withdraw transition to update balances
	_, err := incomingState.ApplyEscrowWithdrawTransition(withdrawAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Create a copy for the unsigned state mock
	currentUnsignedState := currentSignedState

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentUnsignedState, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, true).Return(currentSignedState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)

	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)

	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowWithdraw &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(withdrawAmount)
	}), mock.Anything).Return(nil)

	// Store incoming state with node signature
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			state.Transition.Type == core.TransitionTypeEscrowWithdraw &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_HomeDeposit_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	depositAmount := decimal.NewFromInt(100)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(0),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(500),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with home deposit transition
	incomingState := currentState.NextState()

	// Apply the home deposit transition to update balances
	_, err := incomingState.ApplyHomeDepositTransition(depositAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)

	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeHomeDeposit &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == userWallet &&
			tx.Amount.Equal(depositAmount)
	}), mock.Anything).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			state.Transition.Type == core.TransitionTypeHomeDeposit &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_HomeWithdrawal_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	withdrawalAmount := decimal.NewFromInt(100)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(300),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(-100),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with home withdrawal transition
	incomingState := currentState.NextState()

	// Apply the home withdrawal transition to update balances
	_, err := incomingState.ApplyHomeWithdrawalTransition(withdrawalAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)

	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeHomeWithdrawal &&
			tx.FromAccount == userWallet &&
			tx.ToAccount == homeChannelID &&
			tx.Amount.Equal(withdrawalAmount)
	}), mock.Anything).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&

			state.Transition.Type == core.TransitionTypeHomeWithdrawal &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_MutualLock_Success(t *testing.T) {
	t.Skip("transition is not supported yet")
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	lockAmount := decimal.NewFromInt(100)
	nonce := uint64(12345)
	challenge := uint32(86400)

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
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

	// Create incoming state with mutual lock transition
	incomingState := currentState.NextState()

	// Apply the mutual lock transition to update balances
	_, err := incomingState.ApplyMutualLockTransition(2, "0xTokenAddress", lockAmount)
	require.NoError(t, err)
	escrowChannelID := *incomingState.EscrowChannelID

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Create home channel for mocking
	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        userWallet,
		Asset:             "usdc",
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: challenge,
		Nonce:             nonce,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("GetChannelByID", homeChannelID).Return(&homeChannel, nil)
	mockMemoryStore.On("IsAssetSupported", asset, "0xTokenAddress", uint64(2)).Return(true, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.ChannelID == escrowChannelID &&
			channel.Type == core.ChannelTypeEscrow &&
			channel.UserWallet == userWallet
	})).Return(nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeMutualLock &&
			tx.FromAccount == homeChannelID &&
			tx.ToAccount == escrowChannelID &&
			tx.Amount.Equal(lockAmount)
	}), mock.Anything).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&

			state.Transition.Type == core.TransitionTypeMutualLock &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_EscrowDeposit_Success(t *testing.T) {
	t.Skip("transition is not supported yet")
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	escrowChannelID := "0xEscrowChannel456"
	depositAmount := decimal.NewFromInt(100)

	// Create user's current state (signed, with escrow ledger)
	// The last transition must be a MutualLock for the EscrowDeposit to be valid
	currentSignedState := core.State{
		ID: core.GetStateID(userWallet, asset, 1, 2),
		Transition: core.Transition{

			Type:      core.TransitionTypeMutualLock,
			TxID:      "0xPreviousMutualLockTx",
			AccountID: "",
			Amount:    depositAmount,
		},
		Asset:           asset,
		UserWallet:      userWallet,
		Epoch:           1,
		Version:         2,
		HomeChannelID:   &homeChannelID,
		EscrowChannelID: &escrowChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(400),
			UserNetFlow:  decimal.NewFromInt(400),
			NodeBalance:  decimal.NewFromInt(100),
			NodeNetFlow:  decimal.NewFromInt(100),
		},
		EscrowLedger: &core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 2,
			UserBalance:  decimal.NewFromInt(100),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		UserSig: stringPtr("0xPreviousUserSig"),
		NodeSig: stringPtr("0xPreviousNodeSig"),
	}

	// Create incoming state with escrow deposit transition
	incomingState := currentSignedState.NextState()

	// Apply the escrow deposit transition to update balances
	_, err := incomingState.ApplyEscrowDepositTransition(depositAmount)
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Create a copy for the unsigned state mock
	currentUnsignedState := currentSignedState

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentUnsignedState, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, true).Return(currentSignedState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)

	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)

	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeEscrowDeposit &&
			tx.FromAccount == escrowChannelID &&
			tx.ToAccount == userWallet &&
			tx.Amount.Equal(depositAmount)
	}), mock.Anything).Return(nil)

	// Store incoming state with node signature
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&

			state.Transition.Type == core.TransitionTypeEscrowDeposit &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestSubmitState_Finalize_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"
	userBalance := decimal.NewFromInt(300)

	// Create user's current state (with existing home channel and balance)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       1,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  userBalance,
			UserNetFlow:  userBalance,
			NodeBalance:  decimal.NewFromInt(0),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      nil,
		NodeSig:      nil,
	}

	// Create incoming state with finalize transition
	incomingState := currentState.NextState()

	// Apply the finalize transition to update balances
	finalizeTransition, err := incomingState.ApplyFinalizeTransition()
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	mockAssetStore.On("GetTokenDecimals", uint64(2), "0xEscrowToken").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil)
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)

	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeFinalize &&
			tx.FromAccount == userWallet &&
			tx.ToAccount == homeChannelID &&
			tx.Amount.Equal(userBalance)
	}), mock.Anything).Return(nil)
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&

			state.Transition.Type == core.TransitionTypeFinalize &&
			state.Transition.Amount.Equal(userBalance) &&
			state.HomeLedger.UserBalance.IsZero() &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)

	// Additional assertions specific to finalize transition
	assert.True(t, incomingState.IsFinal(), "State should be marked as final")
	assert.True(t, incomingState.HomeLedger.UserBalance.IsZero(), "User balance should be zero after finalize")
	assert.Equal(t, userBalance, finalizeTransition.Amount, "Finalize amount should equal the original user balance")
}

func TestSubmitState_Acknowledgement_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)
	mockStatePacker := new(MockStatePacker)

	handler := &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			err := handler(mockTxStore)
			if err != nil {
				return err
			}
			return nil
		},
		memoryStore:      mockMemoryStore,
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"

	// Create user's current state (with existing home channel)
	currentState := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 1),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
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

	// Create incoming state with acknowledgement transition
	incomingState := currentState.NextState()
	_, err := incomingState.ApplyAcknowledgementTransition()
	require.NoError(t, err)

	// Sign the incoming state with user's wallet signer (adds 0x01 prefix)
	mockAssetStore.On("GetTokenDecimals", uint64(1), "0xTokenAddress").Return(uint8(6), nil).Maybe()
	packedState, _ := core.PackState(*incomingState, mockAssetStore)
	userSig, _ := userWalletSigner.Sign(packedState)
	userSigStr := userSig.String()
	incomingState.UserSig = &userSigStr

	// Mock expectations
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil)
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil)
	mockTxStore.On("CheckOpenChannel", userWallet, asset).Return("0x03", true, nil)
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(currentState, nil)
	mockTxStore.On("EnsureNoOngoingStateTransitions", userWallet, asset).Return(nil)
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil).Maybe()

	// For acknowledgement: no RecordTransaction call expected, only StoreUserState
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Version == incomingState.Version &&
			state.Transition.Type == core.TransitionTypeAcknowledgement &&
			state.NodeSig != nil
	}), mock.Anything).Return(nil)

	// Create RPC request
	rpcState := toRPCState(*incomingState)
	reqPayload := rpc.ChannelsV1SubmitStateRequest{
		State: rpcState,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.submit_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.SubmitState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)

	var response rpc.ChannelsV1SubmitStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.Nil(t, ctx.Response.Error())
	assert.NotEmpty(t, response.Signature, "Node signature should be present")

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mock expectations - notably RecordTransaction should NOT have been called
	mockTxStore.AssertExpectations(t)
	mockTxStore.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything)
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to convert core.State to rpc.StateV1
func toRPCState(state core.State) rpc.StateV1 {
	transition := rpc.TransitionV1{
		Type:      state.Transition.Type,
		TxID:      state.Transition.TxID,
		AccountID: state.Transition.AccountID,
		Amount:    state.Transition.Amount.String(),
	}

	rpcState := rpc.StateV1{
		ID:              state.ID,
		Transition:      transition,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           strconv.FormatUint(state.Epoch, 10),
		Version:         strconv.FormatUint(state.Version, 10),
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeLedger: rpc.LedgerV1{
			TokenAddress: state.HomeLedger.TokenAddress,
			BlockchainID: strconv.FormatUint(state.HomeLedger.BlockchainID, 10),
			UserBalance:  state.HomeLedger.UserBalance.String(),
			UserNetFlow:  state.HomeLedger.UserNetFlow.String(),
			NodeBalance:  state.HomeLedger.NodeBalance.String(),
			NodeNetFlow:  state.HomeLedger.NodeNetFlow.String(),
		},
		UserSig: state.UserSig,
		NodeSig: state.NodeSig,
	}

	if state.EscrowLedger != nil {
		rpcState.EscrowLedger = &rpc.LedgerV1{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: strconv.FormatUint(state.EscrowLedger.BlockchainID, 10),
			UserBalance:  state.EscrowLedger.UserBalance.String(),
			UserNetFlow:  state.EscrowLedger.UserNetFlow.String(),
			NodeBalance:  state.EscrowLedger.NodeBalance.String(),
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow.String(),
		}
	}

	return rpcState
}
