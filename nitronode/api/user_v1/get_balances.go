package user_v1

import (
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// GetBalances retrieves the balances of the user.
func (h *Handler) GetBalances(c *rpc.Context) {
	var req rpc.UserV1GetBalancesRequest
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

	balances, err := h.store.GetUserBalances(req.Wallet)
	if err != nil {
		c.Fail(err, "failed to retrieve balances")
		return
	}

	response := rpc.UserV1GetBalancesResponse{
		Balances: []rpc.BalanceEntryV1{},
	}

	for _, entry := range balances {
		response.Balances = append(response.Balances, mapBalanceEntryV1(entry))
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
