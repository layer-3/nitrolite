package apps_v1

import (
	"strconv"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetApps retrieves registered applications with optional filtering.
func (h *Handler) GetApps(c *rpc.Context) {
	var req rpc.AppsV1GetAppsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if req.OwnerWallet != nil {
		normalizedOwnerWallet, err := core.NormalizeHexAddress(*req.OwnerWallet)
		if err != nil {
			c.Fail(rpc.Errorf("invalid owner_wallet: %v", err), "")
			return
		}
		req.OwnerWallet = &normalizedOwnerWallet
	}

	var paginationParams core.PaginationParams
	if req.Pagination != nil {
		paginationParams.Offset = req.Pagination.Offset
		paginationParams.Limit = req.Pagination.Limit
		paginationParams.Sort = req.Pagination.Sort
	}

	apps, metadata, err := h.store.GetApps(req.AppID, req.OwnerWallet, &paginationParams)
	if err != nil {
		c.Fail(err, "failed to retrieve apps")
		return
	}

	response := rpc.AppsV1GetAppsResponse{
		Apps:     make([]rpc.AppInfoV1, len(apps)),
		Metadata: mapPaginationMetadataV1(metadata),
	}

	for i, a := range apps {
		response.Apps[i] = mapAppInfoV1(a)
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}

func mapAppInfoV1(info app.AppInfoV1) rpc.AppInfoV1 {
	return rpc.AppInfoV1{
		AppV1: rpc.AppV1{
			ID:                          info.App.ID,
			OwnerWallet:                 info.App.OwnerWallet,
			Metadata:                    info.App.Metadata,
			Version:                     strconv.FormatUint(info.App.Version, 10),
			CreationApprovalNotRequired: info.App.CreationApprovalNotRequired,
		},
		CreatedAt: strconv.FormatInt(info.CreatedAt.Unix(), 10),
		UpdatedAt: strconv.FormatInt(info.UpdatedAt.Unix(), 10),
	}
}

func mapPaginationMetadataV1(meta core.PaginationMetadata) rpc.PaginationMetadataV1 {
	return rpc.PaginationMetadataV1{
		Page:       meta.Page,
		PerPage:    meta.PerPage,
		TotalCount: meta.TotalCount,
		PageCount:  meta.PageCount,
	}
}
