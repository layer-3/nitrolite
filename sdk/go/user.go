package sdk

import (
	"context"
	"fmt"

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
