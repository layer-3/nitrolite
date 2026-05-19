package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetEscrowChannel retrieves current on-chain escrow channel information.
//
// Note: when the escrow channel has been closed by the on-chain purge queue
// (no signed FINALIZE_ESCROW_DEPOSIT was received before expiry), StateVersion
// on the returned channel reflects the initiate version (N) and does not
// advance to the finalize version (N+1).
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
		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get escrow channel")
		return
	}

	response := rpc.ChannelsV1GetEscrowChannelResponse{}
	if channel != nil {
		rpcChannel := coreChannelToRPC(*channel)
		response.Channel = &rpcChannel
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
