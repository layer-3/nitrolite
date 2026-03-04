package user_v1

import (
	"github.com/erc7824/nitrolite/clearnode/action_gateway"
	"github.com/erc7824/nitrolite/pkg/core"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(Store) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

// Store defines the persistence layer interface for user data management.
// All methods should be implemented to work within database transactions.
type Store interface {
	// GetUserBalances retrieves the balances for a user's wallet.
	GetUserBalances(wallet string) ([]core.BalanceEntry, error)

	// GetUserTransactions retrieves transaction history for a user with optional filters.
	GetUserTransactions(Wallet string,
		Asset *string,
		TxType *core.TransactionType,
		FromTime *uint64,
		ToTime *uint64,
		Paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error)

	action_gateway.Store
}

type ActionGateway interface {
	// GetUserAllowances retrieves the action allowances for a user, which define what actions the user is permitted to perform.
	GetUserAllowances(tx action_gateway.Store, userAddress string) ([]core.ActionAllowance, error)
}
