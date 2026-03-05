package event_handlers

import (
	"github.com/stretchr/testify/mock"

	"github.com/layer-3/nitrolite/pkg/core"
)

// MockStore is a mock implementation of the Store interface for testing
type MockStore struct {
	mock.Mock
}

// GetLastStateByChannelID mocks retrieving the last state for a channel
func (m *MockStore) GetLastStateByChannelID(channelID string, signed bool) (*core.State, error) {
	args := m.Called(channelID, signed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.State), args.Error(1)
}

// GetStateByChannelIDAndVersion mocks retrieving a specific state version for a channel
func (m *MockStore) GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error) {
	args := m.Called(channelID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.State), args.Error(1)
}

// UpdateChannel mocks updating a channel in the database
func (m *MockStore) UpdateChannel(channel core.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

// GetChannelByID mocks retrieving a channel by its ID
func (m *MockStore) GetChannelByID(channelID string) (*core.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

// ScheduleCheckpoint mocks scheduling a checkpoint operation
func (m *MockStore) ScheduleCheckpoint(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

// ScheduleInitiateEscrowDeposit mocks scheduling an escrow deposit checkpoint
func (m *MockStore) ScheduleInitiateEscrowDeposit(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

// ScheduleFinalizeEscrowDeposit mocks scheduling an escrow deposit checkpoint
func (m *MockStore) ScheduleFinalizeEscrowDeposit(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

// ScheduleFinalizeEscrowWithdrawal mocks scheduling an escrow withdrawal checkpoint
func (m *MockStore) ScheduleFinalizeEscrowWithdrawal(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}
