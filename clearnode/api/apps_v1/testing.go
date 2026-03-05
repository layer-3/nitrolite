package apps_v1

import (
	"time"

	"github.com/layer-3/nitrolite/clearnode/action_gateway"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
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

func (m *MockStore) GetAppCount(_ string) (uint64, error) {
	return 0, nil
}

func (m *MockStore) GetTotalUserStaked(_ string) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (m *MockStore) RecordAction(_ string, _ core.GatedAction) error {
	return nil
}

func (m *MockStore) GetUserActionCount(_ string, _ core.GatedAction, _ time.Duration) (uint64, error) {
	return 0, nil
}

func (m *MockStore) GetUserActionCounts(_ string, _ time.Duration) (map[core.GatedAction]uint64, error) {
	return nil, nil
}

type MockActionGateway struct {
	Err error
}

func (m *MockActionGateway) AllowAppRegistration(_ action_gateway.Store, _ string) error {
	return m.Err
}
