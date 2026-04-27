package channel_v1

import (
	"fmt"
	"strings"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

func channelStatusFromString(s string) (core.ChannelStatus, error) {
	switch strings.ToLower(s) {
	case "void":
		return core.ChannelStatusVoid, nil
	case "open":
		return core.ChannelStatusOpen, nil
	case "challenged":
		return core.ChannelStatusChallenged, nil
	case "closed":
		return core.ChannelStatusClosed, nil
	default:
		return 0, fmt.Errorf("unknown channel status: %q", s)
	}
}

func channelTypeFromString(s string) (core.ChannelType, error) {
	switch strings.ToLower(s) {
	case "home":
		return core.ChannelTypeHome, nil
	case "escrow":
		return core.ChannelTypeEscrow, nil
	default:
		return 0, fmt.Errorf("unknown channel type: %q", s)
	}
}

// GetChannels retrieves all channels for a user with optional status/asset/type filtering and pagination.
func (h *Handler) GetChannels(c *rpc.Context) {
	var req rpc.ChannelsV1GetChannelsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse request")
		return
	}

	if req.Wallet == "" {
		c.Fail(rpc.Errorf("wallet is required"), "missing wallet")
		return
	}
	normalizedWallet, err := core.NormalizeHexAddress(req.Wallet)
	if err != nil {
		c.Fail(rpc.Errorf("invalid wallet: %v", err), "")
		return
	}
	req.Wallet = normalizedWallet

	var statusFilter *core.ChannelStatus
	if req.Status != nil && *req.Status != "" {
		s, err := channelStatusFromString(*req.Status)
		if err != nil {
			c.Fail(rpc.Errorf("invalid status: %v", err), "invalid status filter")
			return
		}
		statusFilter = &s
	}

	var typeFilter *core.ChannelType
	if req.ChannelType != nil && *req.ChannelType != "" {
		t, err := channelTypeFromString(*req.ChannelType)
		if err != nil {
			c.Fail(rpc.Errorf("invalid channel_type: %v", err), "invalid channel type filter")
			return
		}
		typeFilter = &t
	}

	const defaultLimit uint32 = 100
	const maxLimit uint32 = 1000

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
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	var channels []core.Channel
	var totalCount uint32

	err = h.useStoreInTx(func(tx Store) error {
		var err error
		channels, totalCount, err = tx.GetUserChannels(req.Wallet, statusFilter, req.Asset, typeFilter, limit, offset)
		if err != nil {
			return rpc.Errorf("failed to get channels: %v", err)
		}
		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get channels")
		return
	}

	rpcChannels := make([]rpc.ChannelV1, len(channels))
	for i, ch := range channels {
		rpcChannels[i] = coreChannelToRPC(ch)
	}

	var pageCount, page uint32
	if totalCount > 0 {
		pageCount = uint32((uint64(totalCount) + uint64(limit) - 1) / uint64(limit))
		page = 1
		if offset > 0 {
			page = (offset / limit) + 1
		}
	}

	response := rpc.ChannelsV1GetChannelsResponse{
		Channels: rpcChannels,
		Metadata: rpc.PaginationMetadataV1{
			Page:       page,
			PerPage:    limit,
			TotalCount: totalCount,
			PageCount:  pageCount,
		},
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
