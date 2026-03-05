package apps_v1

import (
	"context"
	"testing"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetApps_Success(t *testing.T) {
	mockStore := &MockStore{
		getAppsFn: func(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
			return []app.AppInfoV1{
				{App: app.AppV1{ID: "app-1", OwnerWallet: "0x1111", Metadata: "0x00", Version: 1}},
				{App: app.AppV1{ID: "app-2", OwnerWallet: "0x2222", Metadata: "0x00", Version: 1}},
			}, core.PaginationMetadata{TotalCount: 2, Page: 1, PerPage: 50}, nil
		},
	}

	handler := NewHandler(mockStore, nil, nil, 4096)

	reqPayload := rpc.AppsV1GetAppsRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1GetAppsMethod), payload),
	}

	handler.GetApps(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var resp rpc.AppsV1GetAppsResponse
	require.NoError(t, ctx.Response.Payload.Translate(&resp))
	assert.Len(t, resp.Apps, 2)
	assert.Equal(t, "app-1", resp.Apps[0].ID)
	assert.Equal(t, "app-2", resp.Apps[1].ID)
	assert.Equal(t, uint32(2), resp.Metadata.TotalCount)
}

func TestGetApps_EmptyResults(t *testing.T) {
	mockStore := &MockStore{
		getAppsFn: func(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
			return []app.AppInfoV1{}, core.PaginationMetadata{TotalCount: 0}, nil
		},
	}

	handler := NewHandler(mockStore, nil, nil, 4096)

	reqPayload := rpc.AppsV1GetAppsRequest{}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1GetAppsMethod), payload),
	}

	handler.GetApps(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var resp rpc.AppsV1GetAppsResponse
	require.NoError(t, ctx.Response.Payload.Translate(&resp))
	assert.Empty(t, resp.Apps)
	assert.Equal(t, uint32(0), resp.Metadata.TotalCount)
}

func TestGetApps_FilterByAppID(t *testing.T) {
	mockStore := &MockStore{
		getAppsFn: func(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
			require.NotNil(t, appID)
			assert.Equal(t, "app-1", *appID)
			return []app.AppInfoV1{
				{App: app.AppV1{ID: "app-1", OwnerWallet: "0x1111", Version: 1}},
			}, core.PaginationMetadata{TotalCount: 1, Page: 1, PerPage: 50}, nil
		},
	}

	handler := NewHandler(mockStore, nil, nil, 4096)

	aid := "app-1"
	reqPayload := rpc.AppsV1GetAppsRequest{AppID: &aid}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1GetAppsMethod), payload),
	}

	handler.GetApps(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var resp rpc.AppsV1GetAppsResponse
	require.NoError(t, ctx.Response.Payload.Translate(&resp))
	assert.Len(t, resp.Apps, 1)
	assert.Equal(t, "app-1", resp.Apps[0].ID)
}

func TestGetApps_FilterByOwnerWallet(t *testing.T) {
	mockStore := &MockStore{
		getAppsFn: func(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
			require.NotNil(t, ownerWallet)
			assert.Equal(t, "0x1111", *ownerWallet)
			return []app.AppInfoV1{
				{App: app.AppV1{ID: "app-1", OwnerWallet: "0x1111", Version: 1}},
			}, core.PaginationMetadata{TotalCount: 1, Page: 1, PerPage: 50}, nil
		},
	}

	handler := NewHandler(mockStore, nil, nil, 4096)

	owner := "0x1111"
	reqPayload := rpc.AppsV1GetAppsRequest{OwnerWallet: &owner}
	payload, err := rpc.NewPayload(reqPayload)
	require.NoError(t, err)

	ctx := &rpc.Context{
		Context: context.Background(),
		Request: rpc.NewRequest(1, string(rpc.AppsV1GetAppsMethod), payload),
	}

	handler.GetApps(ctx)

	require.NotNil(t, ctx.Response)
	require.NoError(t, ctx.Response.Error())

	var resp rpc.AppsV1GetAppsResponse
	require.NoError(t, ctx.Response.Payload.Translate(&resp))
	assert.Len(t, resp.Apps, 1)
	assert.Equal(t, "0x1111", resp.Apps[0].OwnerWallet)
}
