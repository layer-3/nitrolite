package channel_v1

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/clearnode/metrics"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func newGetChannelsHandler(mockTxStore *MockStore) *Handler {
	mockAssetStore := new(MockAssetStore)
	mockSigner := NewMockSigner()
	nodeSigner, _ := core.NewChannelDefaultSigner(mockSigner)
	nodeAddress := mockSigner.PublicKey().Address().String()
	mockStatePacker := new(MockStatePacker)

	return &Handler{
		stateAdvancer: core.NewStateAdvancerV1(mockAssetStore),
		statePacker:   mockStatePacker,
		useStoreInTx: func(handler StoreTxHandler) error {
			return handler(mockTxStore)
		},
		nodeSigner:       nodeSigner,
		nodeAddress:      nodeAddress,
		minChallenge:     uint32(3600),
		metrics:          metrics.NewNoopRuntimeMetricExporter(),
		maxSessionKeyIDs: 256,
		actionGateway:    &MockActionGateway{},
	}
}

func TestGetChannels_Success(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	userWallet := "0x1234567890123456789012345678901234567890"

	channels := []core.Channel{
		{
			ChannelID:    "0xChannel1",
			UserWallet:   userWallet,
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xToken1",
			Nonce:        1,
			Status:       core.ChannelStatusOpen,
			StateVersion: 3,
		},
		{
			ChannelID:    "0xChannel2",
			UserWallet:   userWallet,
			Asset:        "usdc",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xToken1",
			Nonce:        2,
			Status:       core.ChannelStatusClosed,
			StateVersion: 5,
		},
	}

	mockTxStore.On("GetUserChannels", userWallet, (*core.ChannelStatus)(nil), (*string)(nil), (*core.ChannelType)(nil), uint32(100), uint32(0)).
		Return(channels, uint32(2), nil)

	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: userWallet,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.NotNil(t, ctx.Response.Payload)
	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetChannelsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Channels, 2)
	assert.Equal(t, "0xChannel1", response.Channels[0].ChannelID)
	assert.Equal(t, "open", response.Channels[0].Status)
	assert.Equal(t, "0xChannel2", response.Channels[1].ChannelID)
	assert.Equal(t, "closed", response.Channels[1].Status)

	assert.Equal(t, uint32(1), response.Metadata.Page)
	assert.Equal(t, uint32(100), response.Metadata.PerPage)
	assert.Equal(t, uint32(2), response.Metadata.TotalCount)
	assert.Equal(t, uint32(1), response.Metadata.PageCount)

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_WithStatusFilter(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	userWallet := "0x1234567890123456789012345678901234567890"
	statusClosed := core.ChannelStatusClosed

	closedChannel := core.Channel{
		ChannelID:    "0xChannel2",
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		BlockchainID: 1,
		Status:       core.ChannelStatusClosed,
		StateVersion: 5,
	}

	mockTxStore.On("GetUserChannels", userWallet, &statusClosed, (*string)(nil), (*core.ChannelType)(nil), uint32(100), uint32(0)).
		Return([]core.Channel{closedChannel}, uint32(1), nil)

	statusFilterStr := "closed"
	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: userWallet,
		Status: &statusFilterStr,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetChannelsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Channels, 1)
	assert.Equal(t, "closed", response.Channels[0].Status)
	assert.Equal(t, uint32(1), response.Metadata.TotalCount)

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_WithPagination(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	userWallet := "0x1234567890123456789012345678901234567890"
	limit := uint32(10)
	offset := uint32(20)

	channel := core.Channel{
		ChannelID:    "0xChannel21",
		UserWallet:   userWallet,
		Asset:        "usdc",
		Type:         core.ChannelTypeHome,
		BlockchainID: 1,
		Status:       core.ChannelStatusOpen,
	}

	mockTxStore.On("GetUserChannels", userWallet, (*core.ChannelStatus)(nil), (*string)(nil), (*core.ChannelType)(nil), limit, offset).
		Return([]core.Channel{channel}, uint32(25), nil)

	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: userWallet,
		Pagination: &rpc.PaginationParamsV1{
			Limit:  &limit,
			Offset: &offset,
		},
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetChannelsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Channels, 1)
	assert.Equal(t, uint32(3), response.Metadata.Page)
	assert.Equal(t, uint32(10), response.Metadata.PerPage)
	assert.Equal(t, uint32(25), response.Metadata.TotalCount)
	assert.Equal(t, uint32(3), response.Metadata.PageCount)

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_EmptyResult(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	userWallet := "0xNoChannelsUser"

	mockTxStore.On("GetUserChannels", userWallet, (*core.ChannelStatus)(nil), (*string)(nil), (*core.ChannelType)(nil), uint32(100), uint32(0)).
		Return([]core.Channel{}, uint32(0), nil)

	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: userWallet,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.Nil(t, ctx.Response.Error())

	var response rpc.ChannelsV1GetChannelsResponse
	err = ctx.Response.Payload.Translate(&response)
	require.NoError(t, err)

	assert.Len(t, response.Channels, 0)
	assert.Equal(t, uint32(0), response.Metadata.TotalCount)

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_MissingWallet(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: "",
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "wallet")

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_InvalidStatus(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	garbage := "garbage"
	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: "0x1234567890123456789012345678901234567890",
		Status: &garbage,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.Equal(t, rpc.MsgTypeRespErr, ctx.Response.Type, "invalid status should produce error response")

	mockTxStore.AssertExpectations(t)
}

func TestGetChannels_StoreError(t *testing.T) {
	mockTxStore := new(MockStore)
	handler := newGetChannelsHandler(mockTxStore)

	userWallet := "0x1234567890123456789012345678901234567890"

	mockTxStore.On("GetUserChannels", userWallet, (*core.ChannelStatus)(nil), (*string)(nil), (*core.ChannelType)(nil), uint32(100), uint32(0)).
		Return(nil, uint32(0), fmt.Errorf("database connection lost"))

	reqPayload := rpc.ChannelsV1GetChannelsRequest{
		Wallet: userWallet,
	}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.Message{Method: "channels.v1.get_channels", Payload: payload},
	}

	handler.GetChannels(ctx)

	assert.NotNil(t, ctx.Response.Error())
	assert.Contains(t, ctx.Response.Error().Error(), "failed to get channels")

	mockTxStore.AssertExpectations(t)
}
