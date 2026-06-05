package app_session_v1

import (
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetLastKeyStates retrieves the latest session key states for a user with optional filtering by session key.
// Mandatory pagination caps response size to prevent unbounded reads.
func (h *Handler) GetLastKeyStates(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var req rpc.AppSessionsV1GetLastKeyStatesRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if req.UserAddress == "" {
		c.Fail(rpc.Errorf("user_address is required"), "")
		return
	}

	var paginationParams core.PaginationParams
	if req.Pagination != nil {
		// The endpoint orders rows by (created_at DESC, id ASC) for stable pagination;
		// callers cannot override this, so any sort value is rejected rather than silently
		// ignored. PaginationParamsV1.Sort is shared across the v1 API and other endpoints
		// honor it, which is why we validate here instead of dropping the field.
		if req.Pagination.Sort != nil && *req.Pagination.Sort != "" {
			c.Fail(rpc.Errorf("invalid_pagination: sort is not supported by get_last_key_states"), "")
			return
		}
		paginationParams.Offset = req.Pagination.Offset
		paginationParams.Limit = req.Pagination.Limit
	}
	// GetOffsetAndLimit caps the limit and clamps the offset so the later
	// int(offset) conversion in the store never wraps negative. An explicit
	// limit of 0 still falls back to the page limit.
	offset, limit := paginationParams.GetOffsetAndLimit(rpc.GetLastKeyStatesPageLimit, rpc.GetLastKeyStatesPageLimit)
	if limit == 0 {
		limit = rpc.GetLastKeyStatesPageLimit
	}

	includeInactive := req.IncludeInactive != nil && *req.IncludeInactive

	logger.Debug("retrieving session key states",
		"userAddress", req.UserAddress,
		"sessionKey", req.SessionKey,
		"includeInactive", includeInactive,
		"limit", limit,
		"offset", offset)

	var states []app.AppSessionKeyStateV1
	var totalCount uint32

	err := h.useStoreInTx(func(tx Store) error {
		var err error
		states, totalCount, err = tx.GetLastAppSessionKeyStates(req.UserAddress, req.SessionKey, includeInactive, limit, offset)
		return err
	})

	if err != nil {
		logger.Error("failed to retrieve session key states", "error", err)
		c.Fail(err, "failed to retrieve session key states")
		return
	}

	rpcStates := make([]rpc.AppSessionKeyStateV1, len(states))
	for i, state := range states {
		rpcStates[i] = mapSessionKeyStateV1(&state)
	}

	resp := rpc.AppSessionsV1GetLastKeyStatesResponse{
		States:   rpcStates,
		Metadata: buildPageMetadata(totalCount, limit, offset),
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}

// buildPageMetadata returns the standard pagination metadata for get_last_key_states.
// Page is 1-based and defaults to 1 (including the empty-result case, so the metadata is
// never `{page: 0, page_count: 0}`). For non-aligned offsets the page formula treats the
// offset as a row-skip count and reports the page that contains row `offset+1` — callers
// that need exact page semantics should pass offset as a multiple of limit.
func buildPageMetadata(totalCount, limit, offset uint32) rpc.PaginationMetadataV1 {
	page := uint32(1)
	if limit > 0 && offset >= limit {
		page = (offset / limit) + 1
	}

	var pageCount uint32
	if totalCount > 0 && limit > 0 {
		pageCount = (totalCount + limit - 1) / limit
	}

	return rpc.PaginationMetadataV1{
		Page:       page,
		PerPage:    limit,
		TotalCount: totalCount,
		PageCount:  pageCount,
	}
}
