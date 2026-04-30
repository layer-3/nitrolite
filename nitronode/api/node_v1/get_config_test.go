package node_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetConfig_Success(t *testing.T) {
	// Setup
	mockMemoryStore := new(MockMemoryStore)
	nodeAddress := "0x1234567890123456789012345678901234567890"
	nodeVersion := "v1.0.0"

	handler := &Handler{
		memoryStore: mockMemoryStore,
		nodeAddress: nodeAddress,
		nodeVersion: nodeVersion,
	}

	// Test data
	blockchains := []core.Blockchain{
		{
			Name:              "Ethereum",
			ID:                1,
			ChannelHubAddress: "0xContract1",
		},
		{
			Name:              "Polygon",
			ID:                137,
			ChannelHubAddress: "0xContract137",
		},
	}

	// Mock expectations
	mockMemoryStore.On("GetBlockchains").Return(blockchains, nil)

	// Create RPC request
	reqPayload := rpc.NodeV1GetConfigRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "node.v1.get_config",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetConfig(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.NodeV1GetConfigResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Equal(t, nodeAddress, response.NodeAddress)
	assert.Equal(t, nodeVersion, response.NodeVersion)
	assert.Len(t, response.Blockchains, 2)
	assert.Equal(t, "Ethereum", response.Blockchains[0].Name)
	assert.Equal(t, "1", response.Blockchains[0].BlockchainID)
	assert.Equal(t, "0xContract1", response.Blockchains[0].ChannelHubAddress)
	assert.Equal(t, "Polygon", response.Blockchains[1].Name)
	assert.Equal(t, "137", response.Blockchains[1].BlockchainID)
	assert.Equal(t, "0xContract137", response.Blockchains[1].ChannelHubAddress)

	// Verify all mock expectations
	mockMemoryStore.AssertExpectations(t)
}
