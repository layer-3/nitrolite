package node_v1

import (
	"github.com/stretchr/testify/mock"

	"github.com/layer-3/nitrolite/pkg/core"
)

// MockMemoryStore is a mock implementation of the MemoryStore interface
type MockMemoryStore struct {
	mock.Mock
}

func (m *MockMemoryStore) GetBlockchains() ([]core.Blockchain, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]core.Blockchain), args.Error(1)
}

func (m *MockMemoryStore) GetAssets(blockchainID *uint64) ([]core.Asset, error) {
	args := m.Called(blockchainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]core.Asset), args.Error(1)
}
