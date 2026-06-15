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

// ReadOnlyChannelHub is a read-only view of the on-chain ChannelHub contract,
// used by event handlers to converge the Node row with chain after a guard
// drops an event. Each reactor binds a ReadOnlyChannelHub for its own chain
// and threads it into the handler methods that need an authoritative on-chain
// snapshot; no global multi-chain dispatcher is required.
type ReadOnlyChannelHub interface {
	// FetchChannel reads the authoritative on-chain channel snapshot for channelID
	// and returns a RefreshedChannel ready to overwrite the Node's local row. The
	// snapshot reflects on-chain state at RPC-read time, not event-emit time.
	FetchChannel(ctx context.Context, channelID string) (*RefreshedChannel, error)
}

// ChannelHubEventHandler defines the off-chain reactions to ChannelHub
// blockchain events. Only the three home-channel guard-drop handlers
// (HandleHomeChannelChallenged, HandleHomeChannelCheckpointed,
// HandleHomeChannelClosed) take a ReadOnlyChannelHub: they are the entrypoints
// where a version-regression guard may drop an event whose outer transaction
// has nonetheless committed state on chain, and the on-chain refresh is
// required to converge the Node row with chain. Other handlers do not need
// the hub and so do not accept the parameter, keeping the interface narrow.
type ChannelHubEventHandler interface {
	HandleNodeBalanceUpdated(context.Context, ChannelHubEventHandlerStore, *NodeBalanceUpdatedEvent) error
	HandleHomeChannelCreated(context.Context, ChannelHubEventHandlerStore, *HomeChannelCreatedEvent) error
	HandleHomeChannelMigrated(context.Context, ChannelHubEventHandlerStore, *HomeChannelMigratedEvent) error
	HandleHomeChannelCheckpointed(ctx context.Context, tx ChannelHubEventHandlerStore, hub ReadOnlyChannelHub, event *HomeChannelCheckpointedEvent) error
	HandleHomeChannelChallenged(ctx context.Context, tx ChannelHubEventHandlerStore, hub ReadOnlyChannelHub, event *HomeChannelChallengedEvent) error
	HandleHomeChannelClosed(ctx context.Context, tx ChannelHubEventHandlerStore, hub ReadOnlyChannelHub, event *HomeChannelClosedEvent) error
	HandleEscrowDepositInitiated(context.Context, ChannelHubEventHandlerStore, *EscrowDepositInitiatedEvent) error
	HandleEscrowDepositChallenged(context.Context, ChannelHubEventHandlerStore, *EscrowDepositChallengedEvent) error
	HandleEscrowDepositFinalized(context.Context, ChannelHubEventHandlerStore, *EscrowDepositFinalizedEvent) error
	HandleEscrowDepositsPurged(context.Context, ChannelHubEventHandlerStore, *EscrowDepositsPurgedEvent) error
	HandleEscrowWithdrawalInitiated(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalInitiatedEvent) error
	HandleEscrowWithdrawalChallenged(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalChallengedEvent) error
	HandleEscrowWithdrawalFinalized(context.Context, ChannelHubEventHandlerStore, *EscrowWithdrawalFinalizedEvent) error
}

type ChannelHubEventHandlerStore interface {
	// GetLastStateByChannelID retrieves the most recent state for a given channel.
	// If signed is true, only returns states with both user and node signatures.
	// Returns nil if no matching state exists.
	GetLastStateByChannelID(channelID string, signed bool) (*State, error)

	// GetLastUserState retrieves the most recent state for a user's asset across all
	// channels and detached chain entries (HomeChannelID nil). Returns nil if no
	// matching state exists. If signed is true, only fully co-signed states are returned.
	GetLastUserState(wallet, asset string, signed bool) (*State, error)

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

	// ScheduleChallenge schedules a challengeChannel(...) submission on the channel's home
	// blockchain using the provided state and a node-produced challenger signature.
	ScheduleChallenge(stateID string, chainID uint64) error

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

	// LockUserState acquires SELECT ... FOR UPDATE on the user's balance row so the
	// caller's transaction serializes against concurrent RPC paths that already lock
	// the same row before issuing receiver states. Postgres-only; SQLite is a no-op
	// in tests.
	LockUserState(wallet, asset string) (decimal.Decimal, error)

	// LockUserStateForHomeChannel locks the balance row of the user owning channelID. On
	// postgres it derives the lock key from the channel in SQL and returns the channel read
	// under that lock; on non-postgres (sqlite in tests) the snapshot is taken before the lock
	// for test compatibility. Event handlers must use this instead of a GetChannelByID +
	// LockUserState pair, which reads channel status before the lock and races a concurrent
	// submit_state finalization. Returns nil if the channel is absent.
	LockUserStateForHomeChannel(channelID string) (*Channel, error)

	// UpdateStateSigsIfMissing backfills the user and/or node signatures for a stored state
	// when the corresponding column is currently NULL. Used to repair the local record after
	// an on-chain event proves the state was enforced. Either signature may be empty to skip
	// that side; existing values are never overwritten and the call is idempotent on event replay.
	UpdateStateSigsIfMissing(channelID string, version uint64, userSig, nodeSig string) error

	// HasSignedFinalize reports whether a node-signed Finalize state exists for the given
	// home channel. Used to detect the post-Finalize lifecycle when the channel status
	// has been temporarily overwritten by an on-chain challenge.
	HasSignedFinalize(channelID string) (bool, error)

	// SumNetTransitionAmountAfterVersion returns the net effect on the user's
	// home-channel balance of transitions stored against channelID strictly above
	// minVersion. Receiver credits (TransferReceive, Release) contribute positively;
	// sender debits (TransferSend, Commit) contribute negatively. Other transition
	// kinds are excluded. Used to compute the ChallengeRescue amount when a
	// challenged channel is closed.
	SumNetTransitionAmountAfterVersion(channelID string, minVersion uint64) (decimal.Decimal, error)

	// StoreUserState persists a user state row. Used by the event handler to record a
	// ChallengeRescue squash state derived from a closed challenged channel.
	StoreUserState(state State, applicationID string) error

	// RecordTransaction creates a transaction row linking state transitions. Used by the
	// event handler to record the ChallengeRescue transaction associated with the squash.
	RecordTransaction(tx Transaction, applicationID string) error
}
