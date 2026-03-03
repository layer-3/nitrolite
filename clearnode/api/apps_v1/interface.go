package apps_v1

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
)

// Store defines the persistence layer interface for app registry operations.
type Store interface {
	// CreateApp registers a new application. Returns an error if the app ID already exists.
	CreateApp(entry app.AppV1) error

	// GetApps retrieves applications with optional filtering.
	GetApps(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error)
}
