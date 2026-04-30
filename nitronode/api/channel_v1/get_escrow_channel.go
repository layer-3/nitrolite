package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetEscrowChannel retrieves current on-chain escrow channel information.
func (h *Handler) GetEscrowChannel(c *rpc.Context) {
	var req rpc.ChannelsV1GetEscrowChannelRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var channel *core.Channel
	err := h.useStoreInTx(func(tx Store) error {
		var err error
		channel, err = tx.GetChannelByID(req.EscrowChannelID)
		if err != nil {
			return rpc.Errorf("failed to get channel: %v", err)
		}

		if channel == nil {
			return rpc.Errorf("channel_not_found")
		}

		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get escrow channel")
		return
	}

	response := rpc.ChannelsV1GetEscrowChannelResponse{
		Channel: coreChannelToRPC(*channel),
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
