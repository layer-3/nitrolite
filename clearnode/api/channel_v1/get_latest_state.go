package channel_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetLatestState retrieves the current state of the user stored on the Node.
func (h *Handler) GetLatestState(c *rpc.Context) {
	var req rpc.ChannelsV1GetLatestStateRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	normalizedWallet, err := core.NormalizeHexAddress(req.Wallet)
	if err != nil {
		c.Fail(rpc.Errorf("invalid wallet: %v", err), "")
		return
	}
	req.Wallet = normalizedWallet

	var state core.State
	err = h.useStoreInTx(func(tx Store) error {
		lastState, err := tx.GetLastUserState(req.Wallet, req.Asset, req.OnlySigned)
		if err != nil {
			return rpc.Errorf("failed to get last user state: %v", err)
		}

		if lastState == nil {
			return rpc.Errorf("channel not found")
		}

		state = *lastState
		return nil
	})

	if err != nil {
		c.Fail(err, "failed to get latest user state")
		return
	}

	response := rpc.ChannelsV1GetLatestStateResponse{
		State: coreStateToRPC(state),
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
