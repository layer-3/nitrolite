package sdk

import (
	"context"
	"fmt"
	"strconv"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// ============================================================================
// User Query Methods
// ============================================================================

// GetBalances retrieves the balance information for a user.
//
// Parameters:
//   - wallet: The user's wallet address
//
// Returns:
//   - Slice of Balance containing asset balances
//   - Error if the request fails
//
// Example:
//
//	balances, err := client.GetBalances(ctx, "0x1234567890abcdef...")
//	for _, balance := range balances {
//	    fmt.Printf("%s: %s\n", balance.Asset, balance.Balance)
//	}
func (c *Client) GetBalances(ctx context.Context, wallet string) ([]core.BalanceEntry, error) {
	req := rpc.UserV1GetBalancesRequest{
		Wallet: wallet,
	}
	resp, err := c.rpcClient.UserV1GetBalances(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}
	bals, err := transformBalances(resp.Balances)
	if err != nil {
		return nil, err
	}
	return bals, nil
}

// GetTransactionsOptions contains optional filters for GetTransactions.
type GetTransactionsOptions struct {
	// Asset filters by asset symbol
	Asset *string

	// Pagination parameters
	Pagination *core.PaginationParams
}

// GetTransactions retrieves transaction history for a user with optional filtering.
//
// Parameters:
//   - wallet: The user's wallet address
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of Transaction
//   - core.PaginationMetadata with pagination information
//   - Error if the request fails
//
// Example:
//
//	txs, meta, err := client.GetTransactions(ctx, "0x1234...", nil)
//	for _, tx := range txs {
//	    fmt.Printf("%s: %s → %s (%s %s)\n", tx.TxType, tx.FromAccount, tx.ToAccount, tx.Amount, tx.Asset)
//	}
func (c *Client) GetTransactions(ctx context.Context, wallet string, opts *GetTransactionsOptions) ([]core.Transaction, core.PaginationMetadata, error) {
	req := rpc.UserV1GetTransactionsRequest{
		Wallet: wallet,
	}
	if opts != nil {
		req.Asset = opts.Asset
		req.Pagination = transformPaginationParams(opts.Pagination)
	}
	resp, err := c.rpcClient.UserV1GetTransactions(ctx, req)
	if err != nil {
		return nil, core.PaginationMetadata{}, fmt.Errorf("failed to get transactions: %w", err)
	}
	txs, err := transformTransactions(resp.Transactions)
	if err != nil {
		return nil, core.PaginationMetadata{}, err
	}
	return txs, transformPaginationMetadata(resp.Metadata), nil
}

// GetActionAllowances retrieves the action allowances for a user based on their staking level.
//
// Parameters:
//   - wallet: The user's wallet address
//
// Returns:
//   - Slice of ActionAllowance containing allowance information per gated action
//   - Error if the request fails
func (c *Client) GetActionAllowances(ctx context.Context, wallet string) ([]core.ActionAllowance, error) {
	req := rpc.UserV1GetActionAllowancesRequest{Wallet: wallet}
	resp, err := c.rpcClient.UserV1GetActionAllowances(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get action allowances: %w", err)
	}
	allowances, err := transformActionAllowances(resp.Allowances)
	if err != nil {
		return nil, fmt.Errorf("failed to transform action allowances: %w", err)
	}
	return allowances, nil
}

// transformActionAllowances converts RPC ActionAllowanceV1 slice to core.ActionAllowance slice.
func transformActionAllowances(allowances []rpc.ActionAllowanceV1) ([]core.ActionAllowance, error) {
	result := make([]core.ActionAllowance, 0, len(allowances))
	for _, a := range allowances {
		allowance, err := strconv.ParseUint(a.Allowance, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid allowance value %q for action %q: %w", a.Allowance, a.GatedAction, err)
		}
		used, err := strconv.ParseUint(a.Used, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid used value %q for action %q: %w", a.Used, a.GatedAction, err)
		}
		result = append(result, core.ActionAllowance{
			GatedAction: a.GatedAction,
			TimeWindow:  a.TimeWindow,
			Allowance:   allowance,
			Used:        used,
		})
	}
	return result, nil
}
