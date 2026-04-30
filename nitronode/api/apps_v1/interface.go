package apps_v1

import (
	"github.com/layer-3/nitrolite/nitronode/action_gateway"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
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
	// CreateApp registers a new application. Returns an error if the app ID already exists.
	CreateApp(entry app.AppV1) error

	// GetApps retrieves applications with optional filtering.
	GetApps(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error)

	action_gateway.Store
}

type ActionGateway interface {
	// AllowAppRegistration checks if a user is allowed to register a new application based on their past activity and allowances.
	AllowAppRegistration(tx action_gateway.Store, userAddress string) error
}
