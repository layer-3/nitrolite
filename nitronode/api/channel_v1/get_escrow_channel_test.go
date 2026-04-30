package channel_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/nitronode/metrics"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetEscrowChannel_Success(t *testing.T) {
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
	escrowChannelID := "0xEscrowChannel456"

	escrowChannel := core.Channel{
		ChannelID:         escrowChannelID,
		UserWallet:        userWallet,
		Asset:             "usdc",
		Type:              core.ChannelTypeEscrow,
		BlockchainID:      2,
		TokenAddress:      "0xTokenAddress",
		ChallengeDuration: 86400,
		Nonce:             12345,
		Status:            core.ChannelStatusOpen,
		StateVersion:      2,
	}

	// Mock expectations
	mockTxStore.On("GetChannelByID", escrowChannelID).Return(&escrowChannel, nil)

	// Create RPC request
	reqPayload := rpc.ChannelsV1GetEscrowChannelRequest{
		EscrowChannelID: escrowChannelID,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "channels.v1.get_escrow_channel",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetEscrowChannel(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetEscrowChannelResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, escrowChannelID, response.Channel.ChannelID)
	assert.Equal(t, userWallet, response.Channel.UserWallet)
	assert.Equal(t, "escrow", response.Channel.Type)
	assert.Equal(t, "2", response.Channel.BlockchainID)
	assert.Equal(t, "open", response.Channel.Status)

	// Verify all mock expectations
	mockTxStore.AssertExpectations(t)
}
