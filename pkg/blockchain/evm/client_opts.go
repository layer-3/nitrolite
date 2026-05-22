package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type ClientAllowanceCheck struct {
	RequireAllowanceCheck bool
}

func (ch ClientAllowanceCheck) apply(c *BlockchainClient) {
	c.requireCheckAllowance = ch.RequireAllowanceCheck
}

type ClientBalanceCheck struct {
	RequireBalanceCheck bool
}

func (ch ClientBalanceCheck) apply(c *BlockchainClient) {
	c.requireCheckBalance = ch.RequireBalanceCheck
}

// ClientGasLimit forces a fixed GasLimit on every transaction sent through the
// client, bypassing eth_estimateGas. Set to 0 to keep the default behavior
// (estimate per tx). Useful for chains whose RPC rejects estimateGas — e.g.
// XRPL EVM testnet returns "gas cap cannot be lower than 21000".
//
// TODO: temporary workaround. A single client-wide gas cap overshoots cheap
// calls and may undershoot expensive ones. Replace with per-action gas
// estimation that picks a healthy limit per tx type (deposit, withdraw,
// challenge, etc.), falling back to a sane floor only when the RPC refuses
// estimateGas.
type ClientGasLimit struct {
	GasLimit uint64
}

func (g ClientGasLimit) apply(c *BlockchainClient) {
	if g.GasLimit > 0 {
		c.transactOpts.GasLimit = g.GasLimit
	}
}

type ClientFeeCheck struct {
	RequirePositiveNativeBalance bool
}

func (ch ClientFeeCheck) apply(c *BlockchainClient) {
	if !ch.RequirePositiveNativeBalance {
		c.checkFeeFn = func(ctx context.Context, account common.Address) error {
			return nil
		}
	} else {
		c.checkFeeFn = getDefaultCheckFeeFn(c.evmClient)
	}
}

func getDefaultCheckFeeFn(evmClient EVMClient) func(ctx context.Context, account common.Address) error {
	return func(ctx context.Context, account common.Address) error {
		balance, err := evmClient.BalanceAt(ctx, account, nil)
		if err != nil {
			return err
		}

		if balance.Sign() <= 0 {
			return fmt.Errorf("insufficient balance for fee")
		}

		return nil
	}
}
