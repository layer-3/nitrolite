package sdk

import (
	"context"
	"testing"

	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetConfirmationDelay(t *testing.T) {
	t.Parallel()

	t.Run("returns confirmation delay for matching chain", func(t *testing.T) {
		t.Parallel()
		mockDialer := NewMockDialer()
		mockDialer.Dial(context.Background(), "", nil)

		mockDialer.RegisterResponse(rpc.NodeV1GetConfigMethod.String(), rpc.NodeV1GetConfigResponse{
			NodeAddress: "0xNodeAddress",
			Blockchains: []rpc.BlockchainInfoV1{
				{Name: "Ethereum", BlockchainID: "1", ConfirmationDelaySecs: 36},
				{Name: "Polygon", BlockchainID: "137", ConfirmationDelaySecs: 5},
			},
		})

		client := &Client{
			rpcClient: rpc.NewClient(mockDialer),
		}

		delay, err := client.GetConfirmationDelay(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, uint32(36), delay)
	})

	t.Run("returns 0 when gate is disabled", func(t *testing.T) {
		t.Parallel()
		mockDialer := NewMockDialer()
		mockDialer.Dial(context.Background(), "", nil)

		mockDialer.RegisterResponse(rpc.NodeV1GetConfigMethod.String(), rpc.NodeV1GetConfigResponse{
			NodeAddress: "0xNodeAddress",
			Blockchains: []rpc.BlockchainInfoV1{
				{Name: "Polygon", BlockchainID: "137", ConfirmationDelaySecs: 0},
			},
		})

		client := &Client{
			rpcClient: rpc.NewClient(mockDialer),
		}

		delay, err := client.GetConfirmationDelay(context.Background(), 137)
		require.NoError(t, err)
		assert.Equal(t, uint32(0), delay)
	})

	t.Run("returns error when chain not found", func(t *testing.T) {
		t.Parallel()
		mockDialer := NewMockDialer()
		mockDialer.Dial(context.Background(), "", nil)

		mockDialer.RegisterResponse(rpc.NodeV1GetConfigMethod.String(), rpc.NodeV1GetConfigResponse{
			NodeAddress: "0xNodeAddress",
			Blockchains: []rpc.BlockchainInfoV1{
				{Name: "Polygon", BlockchainID: "137", ConfirmationDelaySecs: 5},
			},
		})

		client := &Client{
			rpcClient: rpc.NewClient(mockDialer),
		}

		_, err := client.GetConfirmationDelay(context.Background(), 999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "999")
		assert.Contains(t, err.Error(), "not found in node config")
	})
}
