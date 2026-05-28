package user_v1

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestGetActionAllowances_Success(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(
		mockStore,
		storeTxProvider,
		&MockActionGateway{
			Allowances: []core.ActionAllowance{
				{GatedAction: core.GatedActionTransfer, TimeWindow: "24h", Allowance: 100, Used: 5},
				{GatedAction: core.GatedActionAppSessionCreation, TimeWindow: "24h", Allowance: 50, Used: 0},
			},
		},
	)

	reqPayload := rpc.UserV1GetActionAllowancesRequest{
		Wallet: "0x1234567890123456789012345678901234567890",
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "user.v1.get_action_allowances", payload),
	}

	handler.GetActionAllowances(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var response rpc.UserV1GetActionAllowancesResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Allowances, 2)
	assert.Equal(t, core.GatedActionTransfer, response.Allowances[0].GatedAction)
	assert.Equal(t, "24h", response.Allowances[0].TimeWindow)
	assert.Equal(t, "100", response.Allowances[0].Allowance)
	assert.Equal(t, "5", response.Allowances[0].Used)
	assert.Equal(t, core.GatedActionAppSessionCreation, response.Allowances[1].GatedAction)
	assert.Equal(t, "50", response.Allowances[1].Allowance)
	assert.Equal(t, "0", response.Allowances[1].Used)
}

func TestGetActionAllowances_EmptyResult(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(mockStore, storeTxProvider, &MockActionGateway{})

	reqPayload := rpc.UserV1GetActionAllowancesRequest{
		Wallet: "0x1234567890123456789012345678901234567890",
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "user.v1.get_action_allowances", payload),
	}

	handler.GetActionAllowances(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var response rpc.UserV1GetActionAllowancesResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Empty(t, response.Allowances)
}

func TestGetActionAllowances_MissingWallet(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(mockStore, storeTxProvider, &MockActionGateway{})

	reqPayload := rpc.UserV1GetActionAllowancesRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "user.v1.get_action_allowances", payload),
	}

	handler.GetActionAllowances(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wallet is required")
}

func TestGetActionAllowances_GatewayError(t *testing.T) {
	mockStore := new(MockStore)
	storeTxProvider := func(fn StoreTxHandler) error {
		return fn(mockStore)
	}

	handler := NewHandler(mockStore, storeTxProvider, &MockActionGateway{
		Err: fmt.Errorf("gateway failure"),
	})

	reqPayload := rpc.UserV1GetActionAllowancesRequest{
		Wallet: "0x1234567890123456789012345678901234567890",
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, "user.v1.get_action_allowances", payload),
	}

	handler.GetActionAllowances(ctx)

	require.NotNil(t, ctx.Response)
	err = ctx.Response.Error()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve action allowances")
}
