package core

import (
	"context"

	"github.com/shopspring/decimal"
)

// ========= Client Interface =========

// Client defines the interface for interacting with the ChannelsHub smart contract
// TODO: add context to all methods
type BlockchainClient interface {
	// Getters - Token Balance & Approval
	GetTokenBalance(asset string, walletAddress string) (decimal.Decimal, error)
	Approve(asset string, amount decimal.Decimal) (string, error)

	// Getters - ChannelsHub
	GetNodeBalance(token string) (decimal.Decimal, error)
	GetOpenChannels(user string) ([]string, error)
	GetHomeChannelData(homeChannelID string) (HomeChannelDataResponse, error)
	GetEscrowDepositData(escrowChannelID string) (EscrowDepositDataResponse, error)
	GetEscrowWithdrawalData(escrowChannelID string) (EscrowWithdrawalDataResponse, error)

	// Node vault functions
	Deposit(token string, amount decimal.Decimal) (string, error)
	Withdraw(to, token string, amount decimal.Decimal) (string, error)

	// Node lifecycle
	EnsureSigValidatorRegistered(validatorID uint8, validatorAddress string, checkOnly bool) error

	// Channel lifecycle
	Create(def ChannelDefinition, initCCS State) (string, error)
	MigrateChannelHere(def ChannelDefinition, candidate State) (string, error)
	Checkpoint(candidate State) (string, error)
	Challenge(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	Close(candidate State) (string, error)

	// Escrow deposit
	InitiateEscrowDeposit(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowDeposit(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	FinalizeEscrowDeposit(candidate State) (string, error)

	// Escrow withdrawal
	InitiateEscrowWithdrawal(def ChannelDefinition, initCCS State) (string, error)
	ChallengeEscrowWithdrawal(candidate State, challengerSig []byte, challengerIdx ChannelParticipant) (string, error)
	FinalizeEscrowWithdrawal(candidate State) (string, error)
}

// ========= AppRegistryClient Interface =========

type AppRegistryClient interface {
	ApproveToken(amount decimal.Decimal) (string, error)
	GetBalance(user string) (decimal.Decimal, error)
	GetTokenDecimals() (uint8, error)

	Lock(targetWallet string, amount decimal.Decimal) (string, error)
	Relock() (string, error)
	Unlock() (string, error)
	Withdraw(destinationWallet string) (string, error)
}

// ========= TransitionValidator Interface =========

// StateAdvancer applies state transitions
type StateAdvancer interface {
	ValidateAdvancement(currentState, proposedState State) error
}

// ========= StatePacker Interface =========

// StatePacker serializes channel states
type StatePacker interface {
	PackState(state State) ([]byte, error)
}

// ========= AssetStore Interface =========

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)
}

// Channel lifecycle event handlers
type ChannelHubEventHandler interface {
	HandleNodeBalanceUpdated(context.Context, ChannelHubEventHandlerStore, *NodeBalanceUpdatedEvent) error
	HandleHomeChannelCreated(context.Context, ChannelHubEventHandlerStore, *HomeChannelCreatedEvent) error
	HandleHomeChannelMigrated(context.Context, ChannelHubEventHandlerStore, *HomeChannelMigratedEvent) error
	HandleHomeChannelCheckpointed(context.Context, ChannelHubEventHandlerStore, *HomeChannelCheckpointedEvent) error
	HandleHomeChannelChallenged(context.Context, ChannelHubEventHandlerStore, *HomeChannelChallengedEvent) error
	HandleHomeChannelClosed(context.Context, ChannelHubEventHandlerStore, *HomeChannelClosedEvent) error
	HandleEscrowDepositInitiated(context.Context, ChannelHubEventHandlerStore, *EscrowDepositInitiatedEvent) error
	HandleEscrowDepositChallenged(context.Context, ChannelHubEventHandlerStore, *EscrowDepositChallengedEvent) error
	HandleEscrowDepositFinalized(context.Context, ChannelHubEventHandlerStore, *EscrowDepositFinalizedEvent) error
	HandleEscrowWithdrawalInitiated(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalInitiatedEvent) error
	HandleEscrowWithdrawalChallenged(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalChallengedEvent) error
	HandleEscrowWithdrawalFinalized(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalFinalizedEvent) error
}

type ChannelHubEventHandlerStore interface {
	// GetLastStateByChannelID retrieves the most recent state for a given channel.
	// If signed is true, only returns states with both user and node signatures.
	// Returns nil if no matching state exists.
	GetLastStateByChannelID(channelID string, signed bool) (*State, error)

	// GetStateByChannelIDAndVersion retrieves a specific state version for a channel.
	// Returns nil if the state with the specified version does not exist.
	GetStateByChannelIDAndVersion(channelID string, version uint64) (*State, error)

	// UpdateChannel persists changes to a channel's metadata (status, version, etc).
	// The channel must already exist in the database.
	UpdateChannel(channel Channel) error

	// GetChannelByID retrieves a channel by its unique identifier.
	// Returns nil if the channel does not exist.
	GetChannelByID(channelID string) (*Channel, error)

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

	// SetNodeBalance upserts the on-chain liquidity for a given blockchain and asset.
	SetNodeBalance(blockchainID uint64, asset string, value decimal.Decimal) error

	// RefreshUserEnforcedBalance recomputes the locked balance from the user's open home channel on-chain state.
	RefreshUserEnforcedBalance(wallet, asset string) error
}

type LockingContractEventHandler interface {
	HandleUserLockedBalanceUpdated(context.Context, LockingContractEventHandlerStore, *UserLockedBalanceUpdatedEvent) error
}

type LockingContractEventHandlerStore interface {
	// UpdateUserStaked updates the total staked amount for a user on a specific blockchain.
	UpdateUserStaked(wallet string, blockchainID uint64, amount decimal.Decimal) error
}
