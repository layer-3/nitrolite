package user_v1

import (
	"strconv"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// GetActionAllowances retrieves the action allowances for a user.
func (h *Handler) GetActionAllowances(c *rpc.Context) {
	var req rpc.UserV1GetActionAllowancesRequest
	if err := c.Request.Payload.Translate(&req); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if req.Wallet == "" {
		c.Fail(nil, "wallet is required")
		return
	}

	var allowances []core.ActionAllowance
	err := h.useStoreInTx(func(tx Store) error {
		var err error
		allowances, err = h.actionGateway.GetUserAllowances(h.store, req.Wallet)
		if err != nil {
			return rpc.Errorf("failed to retrieve action allowances: %w", err)
		}

		return nil
	})
	if err != nil {
		c.Fail(err, "failed to retrieve action allowances")
		return
	}

	rpcAllowances := make([]rpc.ActionAllowanceV1, len(allowances))
	for i, a := range allowances {
		rpcAllowances[i] = rpc.ActionAllowanceV1{
			GatedAction: a.GatedAction,
			TimeWindow:  a.TimeWindow,
			Allowance:   strconv.FormatUint(a.Allowance, 10),
			Used:        strconv.FormatUint(a.Used, 10),
		}
	}

	response := rpc.UserV1GetActionAllowancesResponse{
		Allowances: rpcAllowances,
	}

	payload, err := rpc.NewPayload(response)
	if err != nil {
		c.Fail(err, "failed to create response")
		return
	}

	c.Succeed(c.Request.Method, payload)
}
