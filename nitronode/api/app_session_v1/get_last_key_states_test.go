package app_session_v1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func newGetLastKeyStatesHandler(store Store) *Handler {
	return &Handler{
		useStoreInTx: func(fn StoreTxHandler) error {
			return fn(store)
		},
	}
}

func callGetLastKeyStates(t *testing.T, h *Handler, req rpc.AppSessionsV1GetLastKeyStatesRequest) *rpc.Context {
	t.Helper()
	payload, err := rpc.NewPayload(req)
	require.NoError(t, err)
	c := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, rpc.AppSessionsV1GetLastKeyStatesMethod.String(), payload),
	}
	h.GetLastKeyStates(c)
	return c
}

func extractGetLastKeyStatesResponse(t *testing.T, c *rpc.Context) rpc.AppSessionsV1GetLastKeyStatesResponse {
	t.Helper()
	require.NotNil(t, c.Response)
	require.Nil(t, c.Response.Error())
	var resp rpc.AppSessionsV1GetLastKeyStatesResponse
	require.NoError(t, c.Response.Payload.Translate(&resp))
	return resp
}

func TestGetLastKeyStates_DefaultsToPageOneOnEmptyResult(t *testing.T) {
	mockStore := new(MockStore)
	h := newGetLastKeyStatesHandler(mockStore)

	mockStore.On("GetLastAppSessionKeyStates", "0xuser", (*string)(nil), uint32(10), uint32(0)).
		Return([]app.AppSessionKeyStateV1{}, 0, nil)

	c := callGetLastKeyStates(t, h, rpc.AppSessionsV1GetLastKeyStatesRequest{UserAddress: "0xuser"})
	resp := extractGetLastKeyStatesResponse(t, c)

	assert.Empty(t, resp.States)
	// Empty results must not produce {page: 0, page_count: 0} — page is always 1-based.
	assert.Equal(t, uint32(1), resp.Metadata.Page)
	assert.Equal(t, uint32(10), resp.Metadata.PerPage)
	assert.Equal(t, uint32(0), resp.Metadata.TotalCount)
	assert.Equal(t, uint32(0), resp.Metadata.PageCount)
}

func TestGetLastKeyStates_PaginationMetadata_AlignedOffset(t *testing.T) {
	mockStore := new(MockStore)
	h := newGetLastKeyStatesHandler(mockStore)

	limit := uint32(10)
	offset := uint32(10)
	pagination := &rpc.PaginationParamsV1{Limit: &limit, Offset: &offset}

	mockStore.On("GetLastAppSessionKeyStates", "0xuser", (*string)(nil), uint32(10), uint32(10)).
		Return([]app.AppSessionKeyStateV1{
			{UserAddress: "0xuser", SessionKey: "0xkey", Version: 1, ExpiresAt: time.Now().Add(time.Hour)},
		}, 25, nil)

	c := callGetLastKeyStates(t, h, rpc.AppSessionsV1GetLastKeyStatesRequest{
		UserAddress: "0xuser",
		Pagination:  pagination,
	})
	resp := extractGetLastKeyStatesResponse(t, c)

	assert.Equal(t, uint32(2), resp.Metadata.Page)
	assert.Equal(t, uint32(3), resp.Metadata.PageCount) // ceil(25/10)
	assert.Equal(t, uint32(25), resp.Metadata.TotalCount)
}

func TestGetLastKeyStates_ClampsLimitToMax(t *testing.T) {
	mockStore := new(MockStore)
	h := newGetLastKeyStatesHandler(mockStore)

	excessive := uint32(1000)
	pagination := &rpc.PaginationParamsV1{Limit: &excessive}

	mockStore.On("GetLastAppSessionKeyStates", "0xuser", (*string)(nil), rpc.GetLastKeyStatesPageLimit, uint32(0)).
		Return([]app.AppSessionKeyStateV1{}, 0, nil)

	c := callGetLastKeyStates(t, h, rpc.AppSessionsV1GetLastKeyStatesRequest{
		UserAddress: "0xuser",
		Pagination:  pagination,
	})
	resp := extractGetLastKeyStatesResponse(t, c)

	assert.Equal(t, rpc.GetLastKeyStatesPageLimit, resp.Metadata.PerPage)
	mockStore.AssertExpectations(t)
}

func TestGetLastKeyStates_RejectsSortField(t *testing.T) {
	mockStore := new(MockStore)
	h := newGetLastKeyStatesHandler(mockStore)

	sort := "asc"
	pagination := &rpc.PaginationParamsV1{Sort: &sort}

	c := callGetLastKeyStates(t, h, rpc.AppSessionsV1GetLastKeyStatesRequest{
		UserAddress: "0xuser",
		Pagination:  pagination,
	})

	require.NotNil(t, c.Response)
	require.NotNil(t, c.Response.Error())
	assert.Contains(t, c.Response.Error().Error(), "sort is not supported")
	mockStore.AssertNotCalled(t, "GetLastAppSessionKeyStates", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetLastKeyStates_RequiresUserAddress(t *testing.T) {
	mockStore := new(MockStore)
	h := newGetLastKeyStatesHandler(mockStore)

	c := callGetLastKeyStates(t, h, rpc.AppSessionsV1GetLastKeyStatesRequest{})

	require.NotNil(t, c.Response)
	require.NotNil(t, c.Response.Error())
	assert.Contains(t, c.Response.Error().Error(), "user_address is required")
}
