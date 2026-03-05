package channel_v1

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetLatestState_Success(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
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
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"

	state := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 5),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       5,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(500),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      stringPtr("0xUserSig"),
		NodeSig:      stringPtr("0xNodeSig"),
	}

	// Mock expectations
	mockTxStore.On("GetLastUserState", userWallet, asset, false).Return(state, nil)

	// Create RPC request
	reqPayload := rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     userWallet,
		Asset:      asset,
		OnlySigned: false,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.get_latest_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetLatestState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetLatestStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, state.ID, response.State.ID)
	assert.Equal(t, userWallet, response.State.UserWallet)
	assert.Equal(t, asset, response.State.Asset)
	assert.Equal(t, "5", response.State.Version)
	assert.Equal(t, "1", response.State.Epoch)
	assert.NotNil(t, response.State.UserSig)
	assert.NotNil(t, response.State.NodeSig)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestGetLatestState_OnlySigned(t *testing.T) {
	// Setup
	mockTxStore := new(MockStore)
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
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     minChallenge,
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	asset := "USDC"
	homeChannelID := "0xHomeChannel123"

	state := core.State{
		ID:            core.GetStateID(userWallet, asset, 1, 3),
		Transition:    core.Transition{},
		Asset:         asset,
		UserWallet:    userWallet,
		Epoch:         1,
		Version:       3,
		HomeChannelID: &homeChannelID,
		HomeLedger: core.Ledger{
			TokenAddress: "0xTokenAddress",
			BlockchainID: 1,
			UserBalance:  decimal.NewFromInt(500),
			UserNetFlow:  decimal.NewFromInt(100),
			NodeBalance:  decimal.NewFromInt(500),
			NodeNetFlow:  decimal.NewFromInt(0),
		},
		EscrowLedger: nil,
		UserSig:      stringPtr("0xUserSig"),
		NodeSig:      stringPtr("0xNodeSig"),
	}

	// Mock expectations
	mockTxStore.On("GetLastUserState", userWallet, asset, true).Return(state, nil)

	// Create RPC request
	reqPayload := rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     userWallet,
		Asset:      asset,
		OnlySigned: true,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.get_latest_state",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetLatestState(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetLatestStateResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, state.ID, response.State.ID)
	assert.Equal(t, "3", response.State.Version)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}
