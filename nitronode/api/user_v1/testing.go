package user_v1

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/layer-3/nitrolite/nitronode/action_gateway"
	"github.com/layer-3/nitrolite/pkg/core"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetUserBalances(wallet string) ([]core.BalanceEntry, error) {
	args := m.Called(wallet)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]core.BalanceEntry), args.Error(1)
}

func (m *MockStore) GetUserTransactions(Wallet string, Asset *string, TxType *core.TransactionType, FromTime *uint64, ToTime *uint64, Paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error) {
	args := m.Called(Wallet, Asset, TxType, FromTime, ToTime, Paginate)
	if args.Get(0) == nil {
		return nil, core.PaginationMetadata{}, args.Error(2)
	}
	var metadata core.PaginationMetadata
	if args.Get(1) != nil {
		metadata = args.Get(1).(core.PaginationMetadata)
	}
	return args.Get(0).([]core.Transaction), metadata, args.Error(2)
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
	Allowances []core.ActionAllowance
	Err        error
}

func (m *MockActionGateway) GetUserAllowances(_ action_gateway.Store, _ string) ([]core.ActionAllowance, error) {
	return m.Allowances, m.Err
}
