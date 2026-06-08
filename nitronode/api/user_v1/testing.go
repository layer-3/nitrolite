package user_v1

import (
	"github.com/stretchr/testify/mock"

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
