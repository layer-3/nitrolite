package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetLastKeyStates retrieves the latest channel session key states for a user with optional filtering by session key.
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

	logger.Debug("retrieving channel session key states",
		"userAddress", req.UserAddress,
		"sessionKey", req.SessionKey)

	var states []core.ChannelSessionKeyStateV1

	err := h.useStoreInTx(func(tx Store) error {
		var err error
		states, err = tx.GetLastChannelSessionKeyStates(req.UserAddress, req.SessionKey)
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

	resp := rpc.ChannelsV1GetLastKeyStatesResponse{
		States: rpcStates,
	}

	payload, err := rpc.NewPayload(resp)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
