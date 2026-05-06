package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

const (
	getLastKeyStatesDefaultLimit uint32 = 10
	getLastKeyStatesMaxLimit     uint32 = 10
)

// GetLastKeyStates retrieves the latest channel session key states for a user with optional filtering by session key.
// Mandatory pagination caps response size to prevent unbounded reads.
func (h *Handler) GetLastKeyStates(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var req rpc.ChannelsV1GetLastKeyStatesRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if req.UserAddress == "" {
		c.Fail(rpc.Errorf("user_address is required"), "")
		return
	}

	var limit, offset uint32
	if req.Pagination != nil {
		if req.Pagination.Limit != nil {
			limit = *req.Pagination.Limit
		}
		if req.Pagination.Offset != nil {
			offset = *req.Pagination.Offset
		}
	}
	if limit == 0 {
		limit = getLastKeyStatesDefaultLimit
	}
	if limit > getLastKeyStatesMaxLimit {
		limit = getLastKeyStatesMaxLimit
	}

	logger.Debug("retrieving channel session key states",
		"userAddress", req.UserAddress,
		"sessionKey", req.SessionKey,
		"limit", limit,
		"offset", offset)

	var states []core.ChannelSessionKeyStateV1
	var totalCount uint32

	err := h.useStoreInTx(func(tx Store) error {
		var err error
		states, totalCount, err = tx.GetLastChannelSessionKeyStates(req.UserAddress, req.SessionKey, limit, offset)
		return err
	})

	if err != nil {
		logger.Error("failed to retrieve channel session key states", "error", err)
		c.Fail(err, "failed to retrieve channel session key states")
		return
	}

	rpcStates := make([]rpc.ChannelSessionKeyStateV1, len(states))
	for i, state := range states {
		rpcStates[i] = mapChannelSessionKeyStateV1(&state)
	}

	var pageCount, page uint32
	if totalCount > 0 {
		pageCount = uint32((uint64(totalCount) + uint64(limit) - 1) / uint64(limit))
		page = 1
		if offset > 0 {
			page = (offset / limit) + 1
		}
	}

	resp := rpc.ChannelsV1GetLastKeyStatesResponse{
		States: rpcStates,
		Metadata: rpc.PaginationMetadataV1{
			Page:       page,
			PerPage:    limit,
			TotalCount: totalCount,
			PageCount:  pageCount,
		},
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
