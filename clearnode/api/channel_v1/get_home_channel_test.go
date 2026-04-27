package channel_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetHomeChannel_Success(t *testing.T) {
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

	homeChannel := core.Channel{
		ChannelID:         homeChannelID,
		UserWallet:        userWallet,
		Asset:             "usdc",
		Type:              core.ChannelTypeHome,
		BlockchainID:      1,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: 86400,
		Nonce:             12345,
		Status:            core.ChannelStatusOpen,
		StateVersion:      1,
	}

	// Mock expectations
	mockTxStore.On("GetActiveHomeChannel", userWallet, asset).Return(&homeChannel, nil)

	// Create RPC request
	reqPayload := rpc.ChannelsV1GetHomeChannelRequest{
		Wallet: userWallet,
		Asset:  asset,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.get_home_channel",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetHomeChannel(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetHomeChannelResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, homeChannelID, response.Channel.ChannelID)
	assert.Equal(t, userWallet, response.Channel.UserWallet)
	assert.Equal(t, "home", response.Channel.Type)
	assert.Equal(t, "1", response.Channel.BlockchainID)
	assert.Equal(t, "open", response.Channel.Status)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

func TestGetHomeChannel_NotFound(t *testing.T) {
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

	// Mock expectations
	mockTxStore.On("GetActiveHomeChannel", userWallet, asset).Return(nil, nil)

	// Create RPC request
	reqPayload := rpc.ChannelsV1GetHomeChannelRequest{
		Wallet: userWallet,
		Asset:  asset,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.get_home_channel",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetHomeChannel(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "channel_not_found")

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}

// TestGetHomeChannel_NormalizesWallet verifies the wallet is normalized before the store call.
func TestGetHomeChannel_NormalizesWallet(t *testing.T) {
	mockTxStore := new(MockStore)

	handler := &Handler{
		useStoreInTx: func(h StoreTxHandler) error { return h(mockTxStore) },
		metrics:      metrics.NewNoopRuntimeMetricExporter(),
	}

	canonicalWallet := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	mixedCaseWallet := "0xABCDEFabcdefABCDEFabcdefABCDEFabcdefABCD"
	asset := "USDC"

	homeChannel := core.Channel{ChannelID: "0xch", UserWallet: canonicalWallet, Asset: asset, Type: core.ChannelTypeHome}
	mockTxStore.On("GetActiveHomeChannel", canonicalWallet, asset).Return(&homeChannel, nil)

	reqPayload := rpc.ChannelsV1GetHomeChannelRequest{Wallet: mixedCaseWallet, Asset: asset}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_home_channel", Payload: payload},
	}

	handler.GetHomeChannel(ctx)

	require.Nil(t, ctx.Response.Error())
	mockTxStore.AssertExpectations(t)
}
