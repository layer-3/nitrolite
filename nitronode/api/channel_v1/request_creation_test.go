package channel_v1

import (
	"context"
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
