package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientAssetStore(t *testing.T) {
	t.Parallel()
	// Create mock dialer
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	// Prepare mock response
	mockResp := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{
			{
				Name:                  "USDC",
				Symbol:                "USDC",
				Decimals:              6,
				SuggestedBlockchainID: "137",
				Tokens: []rpc.TokenV1{
					{
						BlockchainID: "137",
						Address:      "0xToken137",
						Name:         "USDC (Polygon)",
						Symbol:       "USDC",
						Decimals:     6,
					},
					{
						BlockchainID: "1",
						Address:      "0xToken1",
						Name:         "USDC (Mainnet)",
						Symbol:       "USDC",
						Decimals:     6,
					},
				},
			},
		},
	}
	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), mockResp)

	// Create client with mock dialer
	rpcClient := rpc.NewClient(mockDialer)
	client := &Client{
		rpcClient: rpcClient,
	}

	// Create asset store
	store := newClientAssetStore(client)

	// Test GetAssetDecimals
	decimals, err := store.GetAssetDecimals("usdc")
	require.NoError(t, err)
	assert.Equal(t, uint8(6), decimals)

	// Test GetTokenDecimals
	decimals, err = store.GetTokenDecimals(137, "0xToken137")
	require.NoError(t, err)
	assert.Equal(t, uint8(6), decimals)

	// Test GetTokenAddress
	addr, err := store.GetTokenAddress("USDC", 1)
	require.NoError(t, err)
	assert.Equal(t, "0xToken1", addr)

	// Test GetSuggestedBlockchainID
	chainID, err := store.GetSuggestedBlockchainID("USDC")
	require.NoError(t, err)
	assert.Equal(t, uint64(137), chainID)

	// Test AssetExistsOnBlockchain
	exists, err := store.AssetExistsOnBlockchain(137, "USDC")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = store.AssetExistsOnBlockchain(999, "USDC")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test not found cases
	// Invalid asset
	_, err = store.GetAssetDecimals("INVALID")
	assert.Error(t, err)

	// Invalid token address
	_, err = store.GetTokenDecimals(137, "0xInvalid")
	assert.Error(t, err)

	// Invalid chain for asset
	_, err = store.GetTokenAddress("USDC", 999)
	assert.Error(t, err)
}

func TestClientAssetStore_Caching(t *testing.T) {
	t.Parallel()
	// Create mock dialer
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	// Prepare mock response
	mockResp := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{
			{
				Name:                  "USDC",
				Symbol:                "USDC",
				Decimals:              6,
				SuggestedBlockchainID: "137",
				Tokens:                []rpc.TokenV1{},
			},
		},
	}
	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), mockResp)

	// Create client
	rpcClient := rpc.NewClient(mockDialer)
	client := &Client{
		rpcClient: rpcClient,
	}
	store := newClientAssetStore(client)

	// First call should fetch from API
	_, err := store.GetAssetDecimals("USDC")
	require.NoError(t, err)

	// Clear responses to ensure second call hits cache (would error if it called API)
	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), nil)

	_, err = store.GetAssetDecimals("USDC")
	require.NoError(t, err)
}

func TestClientAssetStore_DecimalsValidation(t *testing.T) {
	t.Parallel()
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.NodeV1GetAssetsResponse{
		Assets: []rpc.AssetV1{
			{
				Name:                  "USDC",
				Symbol:                "USDC",
				Decimals:              18,
				SuggestedBlockchainID: "137",
				Tokens: []rpc.TokenV1{
					{
						BlockchainID: "137",
						Address:      "0xToken137",
						Name:         "USDC (Polygon)",
						Symbol:       "USDC",
						Decimals:     6,
					},
				},
			},
		},
	}
	mockDialer.RegisterResponse(rpc.NodeV1GetAssetsMethod.String(), mockResp)

	rpcClient := rpc.NewClient(mockDialer)
	client := &Client{
		rpcClient: rpcClient,
	}
	store := newClientAssetStore(client)

	_, err := store.GetAssetDecimals("USDC")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "asset USDC decimals (18) must not exceed token USDC decimals (6)")
}

func TestDefaultWebsocketDialerConfig(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 5*time.Second, rpc.DefaultWebsocketDialerConfig.HandshakeTimeout)
	assert.Equal(t, 15*time.Second, rpc.DefaultWebsocketDialerConfig.PingTimeout)
}
