package app_session_v1

import (
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetAppSessions retrieves application sessions with optional filtering.
// Requires either app_session_id or participant to be provided.
func (h *Handler) GetAppSessions(c *rpc.Context) {
	var req rpc.AppSessionsV1GetAppSessionsRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	// Validate that either app_session_id or participant is provided
	if req.AppSessionID == nil && req.Participant == nil {
		c.Fail(nil, "either app_session_id or participant must be provided")
		return
	}

	var paginationParams core.PaginationParams
	if req.Pagination != nil {
		paginationParams.Offset = req.Pagination.Offset
		paginationParams.Limit = req.Pagination.Limit
		paginationParams.Sort = req.Pagination.Sort
	}

	var sessions []app.AppSessionV1
	var metadata core.PaginationMetadata
	sessionAllocations := make(map[string]map[string]map[string]decimal.Decimal)

	err := h.useStoreInTx(func(store Store) error {
		var err error
		status := app.AppSessionStatusVoid
		if req.Status != nil {
			switch *req.Status {
			case "open":
				status = app.AppSessionStatusOpen
			case "closed":
				status = app.AppSessionStatusClosed
			default:
				return rpc.Errorf("invalid status: %s", *req.Status)
			}
		}
		sessions, metadata, err = store.GetAppSessions(req.AppSessionID, req.Participant, status, &paginationParams)
		if err != nil {
			return err
		}

		// Fetch allocations for each session
		for _, session := range sessions {
			allocations, err := store.GetParticipantAllocations(session.SessionID)
			if err != nil {
				return err
			}
			sessionAllocations[session.SessionID] = allocations
		}

		return nil
	})

	if err != nil {
		c.Fail(err, "failed to retrieve app sessions")
		return
	}

	response := rpc.AppSessionsV1GetAppSessionsResponse{
		AppSessions: []rpc.AppSessionInfoV1{},
		Metadata:    mapPaginationMetadataV1(metadata),
	}

	for _, session := range sessions {
		allocations := sessionAllocations[session.SessionID]
		if allocations == nil {
			allocations = make(map[string]map[string]decimal.Decimal)
		}
		response.AppSessions = append(response.AppSessions, mapAppSessionInfoV1(session, allocations))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
