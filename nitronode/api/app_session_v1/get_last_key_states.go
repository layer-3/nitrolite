package app_session_v1

import (
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetLastKeyStates retrieves the latest session key states for a user with optional filtering by session key.
func (h *Handler) GetLastKeyStates(c *rpc.Context) {
	ctx := c.Context
	logger := log.FromContext(ctx)

	var req rpc.AppSessionsV1GetLastKeyStatesRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if req.UserAddress == "" {
		c.Fail(rpc.Errorf("wallet is required"), "")
		return
	}

	logger.Debug("retrieving session key states",
		"wallet", req.UserAddress,
		"sessionKey", req.SessionKey)

	var states []app.AppSessionKeyStateV1

	err := h.useStoreInTx(func(tx Store) error {
		var err error
		states, err = tx.GetLastAppSessionKeyStates(req.UserAddress, req.SessionKey)
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
		States: rpcStates,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
