package event_handlers

import (
	"github.com/layer-3/nitrolite/pkg/core"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type StoreTxHandler func(Store) error

// StoreTxProvider wraps Store operations in a database transaction.
// It accepts a StoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type StoreTxProvider func(StoreTxHandler) error

// Store defines the persistence layer interface for channel and state data.
// All methods should be implemented to work within database transactions.
// Implementations are typically provided by the database layer and wrapped by StoreTxProvider.
type Store interface {
	// GetLastStateByChannelID retrieves the most recent state for a given channel.
	// If signed is true, only returns states with both user and node signatures.
	// Returns nil if no matching state exists.
	GetLastStateByChannelID(channelID string, signed bool) (*core.State, error)

	// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
	// Returns nil if the state with the specified version does not exist.
	GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error)

	// UpdateChannel persists changes to a channel's metadata (status, version, etc).
	// The channel must already exist in the database.
	UpdateChannel(channel core.Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	// Returns nil if the channel does not exist.
	GetChannelByID(channelID string) (*core.Channel, error)

	// ScheduleCheckpoint schedules a checkpoint operation for a home channel state.
	// This queues the state to be submitted on-chain to update the channel's on-chain state.
	ScheduleCheckpoint(stateID string, chainID uint64) error

	// ScheduleInitiateEscrowDeposit schedules an initiate for an escrow deposit operation.
	// This queues the state to be submitted on-chain to finalize an escrow deposit.
	ScheduleInitiateEscrowDeposit(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowDeposit schedules a finalize for an escrow deposit operation.
	// This queues the state to be submitted on-chain to finalize an escrow deposit.
	ScheduleFinalizeEscrowDeposit(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowWithdrawal schedules a checkpoint for an escrow withdrawal operation.
	// This queues the state to be submitted on-chain to finalize an escrow withdrawal.
	ScheduleFinalizeEscrowWithdrawal(stateID string, chainID uint64) error
}
