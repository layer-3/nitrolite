package node_v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetAssets_Success(t *testing.T) {
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
	assets := []core.Asset{
		{
			Name:                  "USD Coin",
			Symbol:                "USDC",
			Decimals:              6,
			SuggestedBlockchainID: 1,
			Tokens: []core.Token{
				{
					Name:         "USDC on Ethereum",
					Symbol:       "USDC",
					Address:      "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
					BlockchainID: 1,
					Decimals:     6,
				},
				{
					Name:         "USDC on Polygon",
					Symbol:       "USDC",
					Address:      "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
					BlockchainID: 137,
					Decimals:     6,
				},
			},
		},
		{
			Name:                  "Tether",
			Symbol:                "USDT",
			SuggestedBlockchainID: 1,
			Tokens: []core.Token{
				{
					Name:         "USDT on Ethereum",
					Symbol:       "USDT",
					Address:      "0xdAC17F958D2ee523a2206206994597C13D831ec7",
					BlockchainID: 1,
					Decimals:     6,
				},
			},
		},
	}

	// Mock expectations
	var nilBlockchainID *uint64
	mockMemoryStore.On("GetAssets", nilBlockchainID).Return(assets, nil)

	// Create RPC request
	reqPayload := rpc.NodeV1GetAssetsRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "node.v1.get_assets",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetAssets(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.NodeV1GetAssetsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Assets, 2)

	// Verify USDC asset
	assert.Equal(t, "USD Coin", response.Assets[0].Name)
	assert.Equal(t, "USDC", response.Assets[0].Symbol)
	assert.Equal(t, uint8(6), response.Assets[0].Decimals)
	assert.Len(t, response.Assets[0].Tokens, 2)
	assert.Equal(t, "USDC on Ethereum", response.Assets[0].Tokens[0].Name)
	assert.Equal(t, "USDC", response.Assets[0].Tokens[0].Symbol)
	assert.Equal(t, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", response.Assets[0].Tokens[0].Address)
	assert.Equal(t, "1", response.Assets[0].Tokens[0].BlockchainID)
	assert.Equal(t, uint8(6), response.Assets[0].Tokens[0].Decimals)
	assert.Equal(t, "USDC on Polygon", response.Assets[0].Tokens[1].Name)
	assert.Equal(t, "137", response.Assets[0].Tokens[1].BlockchainID)

	// Verify USDT asset
	assert.Equal(t, "Tether", response.Assets[1].Name)
	assert.Equal(t, "USDT", response.Assets[1].Symbol)
	assert.Len(t, response.Assets[1].Tokens, 1)
	assert.Equal(t, "USDT on Ethereum", response.Assets[1].Tokens[0].Name)
	assert.Equal(t, "0xdAC17F958D2ee523a2206206994597C13D831ec7", response.Assets[1].Tokens[0].Address)

	// Verify all mock expectations
	mockMemoryStore.AssertExpectations(t)
}
