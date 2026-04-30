package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetHomeChannel retrieves current on-chain home channel information.
func (h *Handler) GetHomeChannel(c *rpc.Context) {
	var req rpc.ChannelsV1GetHomeChannelRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse request")
		return
	}

	normalizedWallet, err := core.NormalizeHexAddress(req.Wallet)
	if err != nil {
		c.Fail(rpc.Errorf("invalid wallet: %v", err), "")
		return
	}
	req.Wallet = normalizedWallet

	var channel *core.Channel
	err = h.useStoreInTx(func(tx Store) error {
		var err error
		channel, err = tx.GetActiveHomeChannel(req.Wallet, req.Asset)
		if err != nil {
			return rpc.Errorf("failed to get home channel: %v", err)
		}

		if channel == nil {
			return rpc.Errorf("channel_not_found")
		}

		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get home channel")
		return
	}

	response := rpc.ChannelsV1GetHomeChannelResponse{
		Channel: coreChannelToRPC(*channel),
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
