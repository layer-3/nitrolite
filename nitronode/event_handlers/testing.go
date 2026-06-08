package event_handlers

import (
	"github.com/shopspring/decimal"
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

// GetLastUserState mocks retrieving the most recent state for a user's asset across
// all channels and detached chain entries.
func (m *MockStore) GetLastUserState(wallet, asset string, signed bool) (*core.State, error) {
	args := m.Called(wallet, asset, signed)
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

// ScheduleChallenge mocks scheduling a challenge operation
func (m *MockStore) ScheduleChallenge(stateID string, chainID uint64) error {
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

// SetNodeBalance mocks upserting the on-chain liquidity
func (m *MockStore) SetNodeBalance(blockchainID uint64, asset string, value decimal.Decimal) error {
	args := m.Called(blockchainID, asset, value)
	return args.Error(0)
}

// RefreshUserEnforcedBalance mocks recomputing the locked balance from DB
func (m *MockStore) RefreshUserEnforcedBalance(wallet, asset string) error {
	args := m.Called(wallet, asset)
	return args.Error(0)
}

// UpdateStateSigsIfMissing mocks backfilling missing user and/or node signatures
// for a stored state.
func (m *MockStore) UpdateStateSigsIfMissing(channelID string, version uint64, userSig, nodeSig string) error {
	args := m.Called(channelID, version, userSig, nodeSig)
	return args.Error(0)
}

// SumNetTransitionAmountAfterVersion mocks the net-change query used to compute
// challenge-rescue amounts on a closed channel.
func (m *MockStore) SumNetTransitionAmountAfterVersion(channelID string, minVersion uint64) (decimal.Decimal, error) {
	args := m.Called(channelID, minVersion)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// LockUserState mocks acquiring SELECT ... FOR UPDATE on a user's balance row.
func (m *MockStore) LockUserState(wallet, asset string) (decimal.Decimal, error) {
	args := m.Called(wallet, asset)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// HasSignedFinalize mocks the existence check for a node-signed Finalize state on the given home channel.
func (m *MockStore) HasSignedFinalize(channelID string) (bool, error) {
	args := m.Called(channelID)
	return args.Bool(0), args.Error(1)
}

// StoreUserState mocks persisting a user state row.
func (m *MockStore) StoreUserState(state core.State, applicationID string) error {
	args := m.Called(state, applicationID)
	return args.Error(0)
}

// RecordTransaction mocks recording a transaction row.
func (m *MockStore) RecordTransaction(tx core.Transaction, applicationID string) error {
	args := m.Called(tx, applicationID)
	return args.Error(0)
}
