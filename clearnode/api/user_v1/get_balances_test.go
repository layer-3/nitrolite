package user_v1

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetBalances_Success(t *testing.T) {
	// Setup
	mockStore := new(MockStore)

	handler := &Handler{
		store: mockStore,
	}

	// Test data
	userWallet := "0x1234567890123456789012345678901234567890"
	balances := []core.BalanceEntry{
		{
			Asset:   "usdc",
			Balance: decimal.NewFromFloat(1000.50),
		},
		{
			Asset:   "usdt",
			Balance: decimal.NewFromFloat(500.25),
		},
		{
			Asset:   "eth",
			Balance: decimal.NewFromFloat(2.5),
		},
	}

	// Mock expectations
	mockStore.On("GetUserBalances", userWallet).Return(balances, nil)

	// Create RPC request
	reqPayload := rpc.UserV1GetBalancesRequest{
		Wallet: userWallet,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	rpcRequest := rpc.Message{
		Method:  "user.v1.get_balances",
		Payload: payload,
	}

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpcRequest,
	}

	// Execute
	handler.GetBalances(ctx)

	// Assert
	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.UserV1GetBalancesResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Balances, 3)

	// Verify balances (order may vary in map)
	balanceMap := make(map[string]string)
	for _, b := range response.Balances {
		balanceMap[b.Asset] = b.Amount
	}
	assert.Equal(t, "1000.5", balanceMap["usdc"])
	assert.Equal(t, "500.25", balanceMap["usdt"])
	assert.Equal(t, "2.5", balanceMap["eth"])

	// Verify all mock expectations
	mockStore.AssertExpectations(t)
}

// TestGetBalances_NormalizesWallet verifies the wallet is normalized before the store call.
func TestGetBalances_NormalizesWallet(t *testing.T) {
	mockStore := new(MockStore)

	handler := &Handler{
		store: mockStore,
	}

	canonicalWallet := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	mixedCaseWallet := "0xABCDEFabcdefABCDEFabcdefABCDEFabcdefABCD"

	mockStore.On("GetUserBalances", canonicalWallet).Return([]core.BalanceEntry{}, nil)

	reqPayload := rpc.UserV1GetBalancesRequest{Wallet: mixedCaseWallet}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "user.v1.get_balances", Payload: payload},
	}

	handler.GetBalances(ctx)

	require.Nil(t, ctx.Response.Error())
	mockStore.AssertExpectations(t)
}
