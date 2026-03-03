package apps_v1

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
)

// MockStore implements the Store interface for testing.
type MockStore struct {
	createAppFn func(entry app.AppV1) error
	getAppsFn   func(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error)
}

func (m *MockStore) CreateApp(entry app.AppV1) error {
	if m.createAppFn != nil {
		return m.createAppFn(entry)
	}
	return nil
}

func (m *MockStore) GetApps(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
	if m.getAppsFn != nil {
		return m.getAppsFn(appID, ownerWallet, pagination)
	}
	return nil, core.PaginationMetadata{}, nil
}
