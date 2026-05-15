package channel_v1

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestRequestCreation_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)   // 1 hour
	maxChallenge := uint32(604800) // 7 days
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
		maxChallenge:     maxChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(12345)
	challenge := uint32(86400)
	depositAmount := decimal.NewFromInt(1000)

	// Create void state (starting point)
	voidState := core.NewVoidState(asset, userWallet)

	// Create next state from void
	initialState := voidState.NextState()

	channelDef := core.ChannelDefinition{
		Nonce:                 nonce,
		Challenge:             challenge,
		ApprovedSigValidators: "0x03",
	}
	_, err := initialState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
	require.NoError(t, err)

	// Apply the home deposit transition to update balances
	_, err = initialState.ApplyHomeDepositTransition(depositAmount)
	require.NoError(t, err)

	// Set up mock for PackState (called during signing)
	mockAssetStore.On("GetTokenDecimals", blockchainID, tokenAddress).Return(uint8(6), nil)

	// Sign the initial state with user's wallet signer (adds 0x01 prefix)
	packedState, err := core.PackState(*initialState, mockAssetStore)
	require.NoError(t, err)
	userSig, err := userWalletSigner.Sign(packedState)
	require.NoError(t, err)
	userSigStr := userSig.String()
	initialState.UserSig = &userSigStr

	// Mock expectations for handler
	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil).Once()
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil).Once()
	mockTxStore.On("HasNonClosedHomeChannel", userWallet, asset).Return(false, nil).Once()
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.UserWallet == userWallet &&
			channel.Type == core.ChannelTypeHome &&
			channel.BlockchainID == blockchainID &&
			channel.TokenAddress == tokenAddress &&
			channel.Nonce == nonce &&
			channel.ChallengeDuration == challenge &&
			channel.Status == core.ChannelStatusVoid &&
			channel.StateVersion == 0
	})).Return(nil).Once()
	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		// For home_deposit: fromAccount is homeChannelID, toAccount is userWallet
		return tx.TxType == core.TransactionTypeHomeDeposit &&
			tx.ToAccount == userWallet &&
			tx.FromAccount != "" // homeChannelID will be set by handler
	}), mock.Anything).Return(nil).Once()
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Asset == asset &&
			state.Version == 1 &&
			state.Epoch == 0 &&
			state.NodeSig != nil &&
			state.HomeChannelID != nil
	}), mock.Anything).Return(nil).Once()

	// Create RPC request
	rpcState := toRPCState(*initialState)
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpcState,
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             challenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.ChannelsV1RequestCreationMethod.String(), ctx.Response.Method)
	assert.NotNil(t, ctx.Response.Payload)

	// Verify response contains signature
	var response rpc.ChannelsV1RequestCreationResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Signature)

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mocks were called
	mockMemoryStore.AssertExpectations(t)
	mockAssetStore.AssertExpectations(t)
	mockTxStore.AssertExpectations(t)
}

func TestRequestCreation_Acknowledgement_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)   // 1 hour
	maxChallenge := uint32(604800) // 7 days
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
		maxChallenge:     maxChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data - derive userWallet from a user signer key
	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(12345)
	challenge := uint32(86400)

	// Create void state (starting point)
	voidState := core.NewVoidState(asset, userWallet)

	// Create next state from void
	initialState := voidState.NextState()

	channelDef := core.ChannelDefinition{
		Nonce:                 nonce,
		Challenge:             challenge,
		ApprovedSigValidators: "0x03",
	}
	_, err := initialState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
	require.NoError(t, err)

	// Apply acknowledgement transition (channel creation with no deposit)
	_, err = initialState.ApplyAcknowledgementTransition()
	require.NoError(t, err)

	// Set up mock for PackState (called during signing)
	mockAssetStore.On("GetTokenDecimals", blockchainID, tokenAddress).Return(uint8(6), nil)

	// Sign the initial state with user's wallet signer (adds 0x01 prefix)
	packedState, err := core.PackState(*initialState, mockAssetStore)
	require.NoError(t, err)
	userSig, err := userWalletSigner.Sign(packedState)
	require.NoError(t, err)
	userSigStr := userSig.String()
	initialState.UserSig = &userSigStr

	// Mock expectations for handler
	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil).Once()
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil).Once()
	mockTxStore.On("HasNonClosedHomeChannel", userWallet, asset).Return(false, nil).Once()
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(nil, nil).Once()
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.UserWallet == userWallet &&
			channel.Type == core.ChannelTypeHome &&
			channel.BlockchainID == blockchainID &&
			channel.TokenAddress == tokenAddress &&
			channel.Nonce == nonce &&
			channel.ChallengeDuration == challenge &&
			channel.Status == core.ChannelStatusVoid &&
			channel.StateVersion == 0
	})).Return(nil).Once()
	// For acknowledgement: no RecordTransaction call expected, only StoreUserState
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == userWallet &&
			state.Asset == asset &&
			state.Version == 1 &&
			state.Epoch == 0 &&
			state.NodeSig != nil &&
			state.HomeChannelID != nil &&
			state.Transition.Type == core.TransitionTypeAcknowledgement
	}), mock.Anything).Return(nil).Once()

	// Create RPC request
	rpcState := toRPCState(*initialState)
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpcState,
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             challenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Check for errors first
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}

	assert.Equal(t, rpc.ChannelsV1RequestCreationMethod.String(), ctx.Response.Method)
	assert.NotNil(t, ctx.Response.Payload)

	// Verify response contains signature
	var response rpc.ChannelsV1RequestCreationResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Signature)

	// Verify the node signature is valid and recoverable to the node address
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	// Verify all mocks were called - notably RecordTransaction should NOT have been called
	mockMemoryStore.AssertExpectations(t)
	mockAssetStore.AssertExpectations(t)
	mockTxStore.AssertExpectations(t)
	mockTxStore.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything)
}

func TestRequestCreation_InvalidChallenge(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)   // 1 hour
	maxChallenge := uint32(604800) // 7 days
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
		maxChallenge:     maxChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	tokenAddress := "0xToken"
	nonce := uint64(12345)
	lowChallenge := uint32(1800) // 30 minutes - below minimum

	// Calculate home channel ID
	homeChannelID, err := core.GetHomeChannelID(
		nodeAddress,
		userWallet,
		asset,
		nonce,
		lowChallenge,
		"0x03",
	)
	require.NoError(t, err)

	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, uint64(1)).Return(true, nil).Once()

	// Create RPC request with challenge below minimum
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:            core.GetStateID(userWallet, asset, 1, 1),
			UserWallet:    userWallet,
			Asset:         asset,
			Epoch:         "1",
			Version:       "1",
			HomeChannelID: &homeChannelID,
			Transition: rpc.TransitionV1{
				Amount: "0",
			},
			HomeLedger: rpc.LedgerV1{
				TokenAddress: tokenAddress,
				BlockchainID: "1",
				UserBalance:  "0",
				UserNetFlow:  "0",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             lowChallenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "challenge")

	// Verify all mocks were called
	mockTxStore.AssertExpectations(t)
}

func TestRequestCreation_ChallengeTooHigh(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	minChallenge := uint32(3600)   // 1 hour
	maxChallenge := uint32(604800) // 7 days
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
		maxChallenge:     maxChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	tokenAddress := "0xToken"
	nonce := uint64(12345)
	highChallenge := uint32(1209600) // 14 days - above maximum

	// Calculate home channel ID
	homeChannelID, err := core.GetHomeChannelID(
		nodeAddress,
		userWallet,
		asset,
		nonce,
		highChallenge,
		"0x03",
	)
	require.NoError(t, err)

	// Create RPC request with challenge above maximum
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:            core.GetStateID(userWallet, asset, 1, 1),
			UserWallet:    userWallet,
			Asset:         asset,
			Epoch:         "1",
			Version:       "1",
			HomeChannelID: &homeChannelID,
			Transition: rpc.TransitionV1{
				Amount: "0",
			},
			HomeLedger: rpc.LedgerV1{
				TokenAddress: tokenAddress,
				BlockchainID: "1",
				UserBalance:  "0",
				UserNetFlow:  "0",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             highChallenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	// Execute
	handler.RequestCreation(ctx)

	// Assert
	assert.NotNil(t, ctx.Response)

	// Verify response contains error about challenge being too high
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "challenge")
	assert.Contains(t, err.Error(), "at most")
}

// TestRequestCreation_NonClosedChannelRejection verifies that opening a new channel is
// rejected while a prior channel is still in progress (Closing, Open, or Challenged),
// preventing epoch rebinding by ensuring only one channel lifecycle runs at a time.
func TestRequestCreation_NonClosedChannelRejection(t *testing.T) {
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
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
		minChallenge:     uint32(3600),
		maxChallenge:     uint32(604800),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	userSigner := NewMockSigner()
	userWallet := userSigner.PublicKey().Address().String()
	asset := "USDC"
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(99)
	challenge := uint32(86400)

	homeChannelID, err := core.GetHomeChannelID(nodeAddress, userWallet, asset, nonce, challenge, "0x03")
	require.NoError(t, err)

	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()

	// Gate fires: a non-closed channel exists (e.g., the channel is Closing after off-chain Finalize).
	mockTxStore.On("LockUserState", userWallet, asset).Return(decimal.Zero, nil).Once()
	mockTxStore.On("HasNonClosedHomeChannel", userWallet, asset).Return(true, nil).Once()

	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:            core.GetStateID(userWallet, asset, 1, 1),
			UserWallet:    userWallet,
			Asset:         asset,
			Epoch:         "1",
			Version:       "1",
			HomeChannelID: &homeChannelID,
			Transition:    rpc.TransitionV1{Amount: "100"},
			HomeLedger: rpc.LedgerV1{
				TokenAddress: tokenAddress,
				BlockchainID: "1",
				UserBalance:  "100",
				UserNetFlow:  "100",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             challenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	handler.RequestCreation(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.Error(t, respErr)
	assert.Contains(t, respErr.Error(), "not yet closed")

	mockTxStore.AssertExpectations(t)
	// GetLastUserState, CreateChannel, StoreUserState must NOT be called.
	mockTxStore.AssertNotCalled(t, "GetLastUserState", mock.Anything, mock.Anything, mock.Anything)
	mockTxStore.AssertNotCalled(t, "CreateChannel", mock.Anything)
	mockTxStore.AssertNotCalled(t, "StoreUserState", mock.Anything, mock.Anything)
}

// TestRequestCreation_TransferSend_Success verifies that a TransferSend transition
// on initial channel creation passes the action gateway and produces both
// sender-side and receiver-side state effects. Covers the happy path opposite
// TestRequestCreation_ActionGatewayRejection.
func TestRequestCreation_TransferSend_Success(t *testing.T) {
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
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
		minChallenge:     uint32(3600),
		maxChallenge:     uint32(604800),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	userSigner := NewMockSigner()
	userWalletSigner, _ := core.NewChannelDefaultSigner(userSigner)
	senderWallet := userSigner.PublicKey().Address().String()
	receiverWallet := "0x0987654321098765432109876543210987654321"
	asset := "USDC"
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(12345)
	challenge := uint32(86400)
	transferAmount := decimal.NewFromInt(100)

	// Sender's prior state: positive offchain balance from a prior transfer_receive,
	// no active home channel. This is the bypass scenario MF2-L01 closes.
	currentSenderState := core.State{
		ID:            core.GetStateID(senderWallet, asset, 0, 1),
		Asset:         asset,
		UserWallet:    senderWallet,
		Epoch:         0,
		Version:       1,
		HomeChannelID: nil,
		HomeLedger: core.Ledger{
			TokenAddress: tokenAddress,
			BlockchainID: blockchainID,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.Zero,
			NodeBalance:  decimal.Zero,
			NodeNetFlow:  decimal.NewFromInt(500),
		},
		Transition: core.Transition{Type: core.TransitionTypeTransferReceive},
	}

	// Build proposed sender state: NextState + ApplyChannelCreation + ApplyTransferSend.
	incomingSenderState := currentSenderState.NextState()
	channelDef := core.ChannelDefinition{
		Nonce:                 nonce,
		Challenge:             challenge,
		ApprovedSigValidators: "0x03",
	}
	_, err := incomingSenderState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
	require.NoError(t, err)
	_, err = incomingSenderState.ApplyTransferSendTransition(receiverWallet, transferAmount)
	require.NoError(t, err)

	mockAssetStore.On("GetTokenDecimals", blockchainID, tokenAddress).Return(uint8(6), nil)
	packedState, err := core.PackState(*incomingSenderState, mockAssetStore)
	require.NoError(t, err)
	userSig, err := userWalletSigner.Sign(packedState)
	require.NoError(t, err)
	userSigStr := userSig.String()
	incomingSenderState.UserSig = &userSigStr

	// Handler-side mocks.
	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()
	mockAssetStore.On("GetAssetDecimals", asset).Return(uint8(6), nil).Once()
	mockTxStore.On("LockUserState", senderWallet, asset).Return(decimal.Zero, nil).Once()
	mockTxStore.On("HasNonClosedHomeChannel", senderWallet, asset).Return(false, nil).Once()
	mockTxStore.On("GetLastUserState", senderWallet, asset, false).Return(currentSenderState, nil).Once()
	mockStatePacker.On("PackState", mock.Anything).Return(packedState, nil)
	mockTxStore.On("CreateChannel", mock.MatchedBy(func(channel core.Channel) bool {
		return channel.UserWallet == senderWallet &&
			channel.Type == core.ChannelTypeHome &&
			channel.BlockchainID == blockchainID &&
			channel.TokenAddress == tokenAddress &&
			channel.Nonce == nonce &&
			channel.ChallengeDuration == challenge
	})).Return(nil).Once()
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == senderWallet &&
			state.Asset == asset &&
			state.Transition.Type == core.TransitionTypeTransferSend &&
			state.NodeSig != nil &&
			state.HomeChannelID != nil
	}), mock.Anything).Return(nil).Once()

	// Receiver-side: fresh user, no prior state, no home channel.
	mockTxStore.On("LockUserState", receiverWallet, asset).Return(decimal.Zero, nil).Once()
	mockTxStore.On("GetLastUserState", receiverWallet, asset, false).Return(nil, nil).Once()
	mockTxStore.On("EnsureNoOngoingEscrowOperation", receiverWallet, asset).Return(nil).Once()
	mockTxStore.On("StoreUserState", mock.MatchedBy(func(state core.State) bool {
		return state.UserWallet == receiverWallet &&
			state.Asset == asset &&
			state.Transition.Type == core.TransitionTypeTransferReceive
	}), mock.Anything).Return(nil).Once()

	mockTxStore.On("RecordTransaction", mock.MatchedBy(func(tx core.Transaction) bool {
		return tx.TxType == core.TransactionTypeTransfer &&
			tx.Amount.Equal(transferAmount) &&
			tx.FromAccount == senderWallet &&
			tx.ToAccount == receiverWallet
	}), mock.Anything).Return(nil).Once()

	rpcState := toRPCState(*incomingSenderState)
	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpcState,
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             challenge,
			ApprovedSigValidators: "0x03",
		},
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	handler.RequestCreation(ctx)

	require.NotNil(t, ctx.Response)
	if respErr := ctx.Response.Error(); respErr != nil {
		t.Fatalf("Unexpected error response: %v", respErr)
	}
	var response rpc.ChannelsV1RequestCreationResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Signature)
	VerifyNodeSignature(t, nodeAddress, packedState, response.Signature)

	mockMemoryStore.AssertExpectations(t)
	mockAssetStore.AssertExpectations(t)
	mockTxStore.AssertExpectations(t)
}

// TestRequestCreation_ActionGatewayRejection verifies that an exhausted gated
// action (e.g. transfer_send) is rejected at the gateway before any state is
// signed, stored, or receiver-side state is issued. Mirrors the SubmitState
// gate so users cannot bypass the 24h allowance via initial channel creation.
func TestRequestCreation_ActionGatewayRejection(t *testing.T) {
	mockTxStore := new(MockStore)
	mockMemoryStore := new(MockMemoryStore)
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
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
		minChallenge:     uint32(3600),
		maxChallenge:     uint32(604800),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{Err: errors.New("transfer allowance exhausted")},
	}

	userSigner := NewMockSigner()
	userWallet := userSigner.PublicKey().Address().String()
	receiverWallet := "0x0987654321098765432109876543210987654321"
	asset := "USDC"
	tokenAddress := "0xTokenAddress"
	blockchainID := uint64(1)
	nonce := uint64(77)
	challenge := uint32(86400)

	homeChannelID, err := core.GetHomeChannelID(nodeAddress, userWallet, asset, nonce, challenge, "0x03")
	require.NoError(t, err)

	mockMemoryStore.On("IsAssetSupported", asset, tokenAddress, blockchainID).Return(true, nil).Once()

	reqPayload := rpc.ChannelsV1RequestCreationRequest{
		State: rpc.StateV1{
			ID:            core.GetStateID(userWallet, asset, 1, 1),
			UserWallet:    userWallet,
			Asset:         asset,
			Epoch:         "1",
			Version:       "1",
			HomeChannelID: &homeChannelID,
			Transition: rpc.TransitionV1{
				Type:      core.TransitionTypeTransferSend,
				AccountID: receiverWallet,
				Amount:    "100",
			},
			HomeLedger: rpc.LedgerV1{
				TokenAddress: tokenAddress,
				BlockchainID: "1",
				UserBalance:  "0",
				UserNetFlow:  "0",
				NodeBalance:  "0",
				NodeNetFlow:  "0",
			},
		},
		ChannelDefinition: rpc.ChannelDefinitionV1{
			Nonce:                 strconv.FormatUint(nonce, 10),
			Challenge:             challenge,
			ApprovedSigValidators: "0x03",
		},
	}

	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{
			RequestID: 1,
			Method:    rpc.ChannelsV1RequestCreationMethod.String(),
			Payload:   payload,
		},
	}

	handler.RequestCreation(ctx)

	require.NotNil(t, ctx.Response)
	respErr := ctx.Response.Error()
	require.Error(t, respErr)
	assert.Contains(t, respErr.Error(), "transfer allowance exhausted")

	// Gate fires before any state effect: nothing should be locked, stored, or recorded.
	mockTxStore.AssertNotCalled(t, "LockUserState", mock.Anything, mock.Anything)
	mockTxStore.AssertNotCalled(t, "HasNonClosedHomeChannel", mock.Anything, mock.Anything)
	mockTxStore.AssertNotCalled(t, "GetLastUserState", mock.Anything, mock.Anything, mock.Anything)
	mockTxStore.AssertNotCalled(t, "CreateChannel", mock.Anything)
	mockTxStore.AssertNotCalled(t, "StoreUserState", mock.Anything, mock.Anything)
	mockTxStore.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything)
	mockMemoryStore.AssertExpectations(t)
}
