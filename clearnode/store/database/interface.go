package database

import (
	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// StoreTxHandler is a function that executes Store operations within a transaction.
type StoreTxHandler func(DatabaseStore) error

// DatabaseStore defines the unified persistence layer.
type DatabaseStore interface {
	// ExecuteInTransaction runs the provided handler within a database transaction.
	// If the handler returns an error, the transaction is rolled back.
	// If the handler completes successfully, the transaction is committed.
	ExecuteInTransaction(handler StoreTxHandler) error

	// --- User & Balance Operations ---

	// GetUserBalances retrieves the balances for a user's wallet.
	GetUserBalances(wallet string) ([]core.BalanceEntry, error)

	// LockUserState locks a user's balance row for update (postgres only, must be used within a transaction).
	// Uses INSERT ... ON CONFLICT DO NOTHING to ensure the row exists, then SELECT ... FOR UPDATE to lock it.
	// Returns the current balance or zero if the row was just inserted.
	LockUserState(wallet, asset string) (decimal.Decimal, error)

	// GetUserTransactions retrieves transaction history for a user with optional filters.
	GetUserTransactions(wallet string, asset *string, txType *core.TransactionType, fromTime *uint64, toTime *uint64, paginate *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error)

	// RecordTransaction creates a transaction record linking state transitions.
	RecordTransaction(tx core.Transaction) error

	// --- Channel Operations ---

	// CreateChannel creates a new channel entity in the database.
	CreateChannel(channel core.Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	GetChannelByID(channelID string) (*core.Channel, error)

	// GetActiveHomeChannel retrieves the active home channel for a user's wallet and asset.
	GetActiveHomeChannel(wallet, asset string) (*core.Channel, error)

	// CheckOpenChannel verifies if a user has an active channel for the given asset
	// and returns the approved signature validators if such a channel exists.
	CheckOpenChannel(wallet, asset string) (string, bool, error)

	// UpdateChannel persists changes to a channel's metadata (status, version, etc).
	UpdateChannel(channel core.Channel) error

	// GetUserChannels retrieves all channels for a user with optional status, asset, and type filters.
	GetUserChannels(wallet string, status *core.ChannelStatus, asset *string, channelType *core.ChannelType, limit, offset uint32) ([]core.Channel, uint32, error)

	// --- State Management ---

	// GetLastStateByChannelID retrieves the most recent state for a given channel.
	// If signed is true, only returns states with both user and node signatures.
	GetLastStateByChannelID(channelID string, signed bool) (*core.State, error)

	// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
	// Returns nil if the state with the specified version does not exist.
	GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error)

	// GetLastUserState retrieves the most recent state for a user's asset.
	GetLastUserState(wallet, asset string, signed bool) (*core.State, error)

	// StoreUserState persists a new user state to the database.
	StoreUserState(state core.State) error

	// EnsureNoOngoingStateTransitions validates that no conflicting blockchain operations are pending.
	EnsureNoOngoingStateTransitions(wallet, asset string) error

	// --- Blockchain Action Operations ---

	// ScheduleInitiateEscrowWithdrawal queues a blockchain action to initiate withdrawal.
	// This queues the state to be submitted on-chain to initiate an escrow withdrawal.
	ScheduleInitiateEscrowWithdrawal(stateID string, chainID uint64) error

	// ScheduleCheckpoint schedules a checkpoint operation for a home channel state.
	// This queues the state to be submitted on-chain to update the channel's on-chain state.
	ScheduleCheckpoint(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowDeposit schedules a checkpoint for an escrow deposit operation.
	// This queues the state to be submitted on-chain to finalize an escrow deposit.
	ScheduleFinalizeEscrowDeposit(stateID string, chainID uint64) error

	// ScheduleFinalizeEscrowWithdrawal schedules a checkpoint for an escrow withdrawal operation.
	// This queues the state to be submitted on-chain to finalize an escrow withdrawal.
	ScheduleFinalizeEscrowWithdrawal(stateID string, chainID uint64) error

	// ScheduleInitiateEscrowDeposit schedules a checkpoint for an escrow deposit operation.
	// This queues the state to be submitted on-chain for an escrow deposit on home chain.
	ScheduleInitiateEscrowDeposit(stateID string, chainID uint64) error

	// Fail marks a blockchain action as failed and increments the retry counter.
	Fail(actionID int64, err string) error

	// FailNoRetry marks a blockchain action as failed without incrementing the retry counter.
	FailNoRetry(actionID int64, err string) error

	// RecordAttempt records a failed attempt for a blockchain action and increments the retry counter.
	// The action remains in pending status.
	RecordAttempt(actionID int64, err string) error

	// Complete marks a blockchain action as completed with the given transaction hash.
	Complete(actionID int64, txHash string) error

	// GetActions retrieves pending blockchain actions, optionally limited by count.
	GetActions(limit uint8, chainID uint64) ([]BlockchainAction, error)

	// GetStateByID retrieves a state by its deterministic ID.
	GetStateByID(stateID string) (*core.State, error)

	// --- App Registry Operations ---

	// CreateApp registers a new application. Returns an error if the app ID already exists.
	CreateApp(entry app.AppV1) error

	// GetApp retrieves a single application by ID. Returns nil if not found.
	GetApp(appID string) (*app.AppInfoV1, error)

	// GetApps retrieves applications with optional filtering by app ID, owner wallet, and pagination.
	GetApps(appID *string, ownerWallet *string, pagination *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error)

	// --- App Session Operations ---

	// CreateAppSession initializes a new application session.
	CreateAppSession(session app.AppSessionV1) error

	// GetAppSession retrieves a specific session by ID.
	GetAppSession(sessionID string) (*app.AppSessionV1, error)

	// GetAppSessions retrieves filtered sessions with pagination.
	GetAppSessions(appSessionID *string, participant *string, status app.AppSessionStatus, pagination *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error)

	// UpdateAppSession updates existing session data.
	UpdateAppSession(session app.AppSessionV1) error

	// --- App Ledger Operations ---

	// GetAppSessionBalances retrieves the total balances associated with a session.
	GetAppSessionBalances(sessionID string) (map[string]decimal.Decimal, error)

	// GetParticipantAllocations retrieves specific asset allocations per participant.
	GetParticipantAllocations(sessionID string) (map[string]map[string]decimal.Decimal, error)

	// RecordLedgerEntry logs a movement of funds within the internal ledger.
	RecordLedgerEntry(userWallet, accountID, asset string, amount decimal.Decimal) error

	// --- App Session Key State Operations ---

	// StoreAppSessionKeyState stores or updates a session key state.
	StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error

	GetAppSessionKeyOwner(sessionKey, appSessionId string) (string, error)

	// SessionKeyStateExists returns the latest version of a non-expired session key state for a user.
	// Returns 0 if no state exists.
	GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error)

	// GetLatestSessionKeyState retrieves the latest version of a specific session key for a user.
	// Returns nil if no state exists.
	GetLastAppSessionKeyState(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error)

	// GetLastKeyStates retrieves the latest session key states for a user with optional filtering.
	GetLastAppSessionKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error)

	// --- Channel Session Key State Operations ---

	// StoreChannelSessionKeyState stores or updates a channel session key state.
	StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error

	// GetLastChannelSessionKeyVersion returns the latest version for a (wallet, sessionKey) pair.
	// Returns 0 if no state exists.
	GetLastChannelSessionKeyVersion(wallet, sessionKey string) (uint64, error)

	// GetLastChannelSessionKeyStates retrieves the latest channel session key states for a user,
	// optionally filtered by session key.
	GetLastChannelSessionKeyStates(wallet string, sessionKey *string) ([]core.ChannelSessionKeyStateV1, error)

	// ValidateChannelSessionKeyForAsset checks that a valid, non-expired session key state
	// exists at its latest version for the (wallet, sessionKey) pair, includes the given asset,
	// and matches the metadata hash.
	ValidateChannelSessionKeyForAsset(wallet, sessionKey, asset, metadataHash string) (bool, error)

	// --- Metric Aggregation ---

	// CountAppSessionsByStatus returns app session counts grouped by (application, status).
	CountAppSessionsByStatus() ([]AppSessionCount, error)

	// CountChannelsByStatus returns channel counts grouped by (asset, status).
	CountChannelsByStatus() ([]ChannelCount, error)

	// --- Contract Event Operations ---

	// StoreContractEvent stores a blockchain event to prevent duplicate processing.
	StoreContractEvent(ev core.BlockchainEvent) error

	// GetLatestEvent returns the latest block number and log index for a given contract.
	GetLatestEvent(contractAddress string, blockchainID uint64) (core.BlockchainEvent, error)
}
