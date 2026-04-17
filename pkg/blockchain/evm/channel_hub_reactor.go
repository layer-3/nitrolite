package evm

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

// ChannelHubReactorStoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type ChannelHubReactorStoreTxHandler func(ChannelHubReactorStore) error

// ChannelHubReactorStoreTxProvider wraps Store operations in a database transaction.
// It accepts a ChannelHubReactorStoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type ChannelHubReactorStoreTxProvider func(ChannelHubReactorStoreTxHandler) error

// ChannelHubReactorStore defines the persistence layer interface for channel and state data.
// All methods should be implemented to work within database transactions.
// Implementations are typically provided by the database layer and wrapped by ChannelHubReactorStoreTxProvider.
type ChannelHubReactorStore interface {
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

	// SetNodeBalance upserts the on-chain liquidity for a given blockchain and asset.
	SetNodeBalance(blockchainID uint64, asset string, value decimal.Decimal) error

	// RefreshUserEnforcedBalance recomputes the locked balance from the user's open home channel on-chain state.
	RefreshUserEnforcedBalance(wallet, asset string) error

	// StoreContractEvent persists a blockchain event to the database.
	StoreContractEvent(ev core.BlockchainEvent) error
}

var channelHubAbi *abi.ABI
var channelHubFilterer *ChannelHubFilterer
var channelHubEventMapping map[common.Hash]string

func initChannelHub() {
	var err error
	channelHubAbi, err = ChannelHubMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	// Create a filterer for parsing events (address not needed for parsing)
	contract := bind.NewBoundContract(common.Address{}, *channelHubAbi, nil, nil, nil)
	channelHubFilterer = &ChannelHubFilterer{contract: contract}

	channelHubEventMapping = make(map[common.Hash]string)
	for name, event := range channelHubAbi.Events {
		channelHubEventMapping[event.ID] = name
	}
}

type ChannelHubReactor struct {
	blockchainID     uint64
	nodeAddress      string
	eventHandler     core.ChannelHubEventHandler
	assetStore       AssetStore
	useStoreInTx     ChannelHubReactorStoreTxProvider
	onEventProcessed func(blockchainID uint64, success bool)
}

func NewChannelHubReactor(blockchainID uint64, nodeAddress string, eventHandler core.ChannelHubEventHandler, assetStore AssetStore, useStoreInTx ChannelHubReactorStoreTxProvider) *ChannelHubReactor {
	return &ChannelHubReactor{
		blockchainID: blockchainID,
		nodeAddress:  nodeAddress,
		eventHandler: eventHandler,
		assetStore:   assetStore,
		useStoreInTx: useStoreInTx,
	}
}

// SetOnEventProcessed sets an optional callback invoked after each event is processed.
func (r *ChannelHubReactor) SetOnEventProcessed(fn func(blockchainID uint64, success bool)) {
	r.onEventProcessed = fn
}

func (r *ChannelHubReactor) HandleEvent(ctx context.Context, l types.Log) error {
	logger := log.FromContext(ctx)

	eventID := l.Topics[0]
	eventName, ok := channelHubEventMapping[eventID]
	if !ok {
		logger.Warn("unknown event ID", "eventID", eventID.Hex(), "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
		return nil
	}
	logger.Debug("received event", "name", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)

	err := r.useStoreInTx(func(store ChannelHubReactorStore) error {
		var err error
		switch eventID {
		case channelHubAbi.Events["NodeBalanceUpdated"].ID:
			err = r.handleNodeBalanceUpdated(ctx, store, l)
		case channelHubAbi.Events["ChannelCreated"].ID:
			err = r.handleHomeChannelCreated(ctx, store, l)
		case channelHubAbi.Events["ChannelCheckpointed"].ID:
			err = r.handleHomeChannelCheckpointed(ctx, store, l)
		case channelHubAbi.Events["ChannelDeposited"].ID:
			err = r.handleChannelDeposited(ctx, store, l)
		case channelHubAbi.Events["ChannelWithdrawn"].ID:
			err = r.handleChannelWithdrawn(ctx, store, l)
		case channelHubAbi.Events["ChannelChallenged"].ID:
			err = r.handleHomeChannelChallenged(ctx, store, l)
		case channelHubAbi.Events["ChannelClosed"].ID:
			err = r.handleHomeChannelClosed(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositInitiated"].ID:
			err = r.handleEscrowDepositInitiated(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositChallenged"].ID:
			err = r.handleEscrowDepositChallenged(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositFinalized"].ID:
			err = r.handleEscrowDepositFinalized(ctx, store, l)
		case channelHubAbi.Events["EscrowWithdrawalInitiated"].ID:
			err = r.handleEscrowWithdrawalInitiated(ctx, store, l)
		case channelHubAbi.Events["EscrowWithdrawalChallenged"].ID:
			err = r.handleEscrowWithdrawalChallenged(ctx, store, l)
		case channelHubAbi.Events["EscrowWithdrawalFinalized"].ID:
			err = r.handleEscrowWithdrawalFinalized(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositInitiatedOnHome"].ID:
			err = r.handleEscrowDepositInitiatedOnHome(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositFinalizedOnHome"].ID:
			err = r.handleEscrowDepositFinalizedOnHome(ctx, store, l)
		case channelHubAbi.Events["EscrowWithdrawalInitiatedOnHome"].ID:
			err = r.handleEscrowWithdrawalInitiatedOnHome(ctx, store, l)
		case channelHubAbi.Events["EscrowWithdrawalFinalizedOnHome"].ID:
			err = r.handleEscrowWithdrawalFinalizedOnHome(ctx, store, l)
		// NOTE: Unimplemented handlers:
		case channelHubAbi.Events["MigrationInInitiated"].ID:
			err = r.handleHomeChannelMigrated(ctx, store, l)
		case channelHubAbi.Events["MigrationInFinalized"].ID:
			err = r.handleMigrationInFinalized(ctx, store, l)
		case channelHubAbi.Events["MigrationOutInitiated"].ID:
			err = r.handleMigrationOutInitiated(ctx, store, l)
		case channelHubAbi.Events["MigrationOutFinalized"].ID:
			err = r.handleMigrationOutFinalized(ctx, store, l)
		case channelHubAbi.Events["Deposited"].ID:
			err = r.handleDeposited(ctx, store, l)
		case channelHubAbi.Events["Withdrawn"].ID:
			err = r.handleWithdrawn(ctx, store, l)
		case channelHubAbi.Events["EscrowDepositsPurged"].ID:
			err = r.handleEscrowDepositsPurged(ctx, store, l)
		default:
			logger.Warn("unknown event: " + eventID.Hex())
		}
		if err != nil {
			logger.Warn("error processing event", "error", err)
			return errors.Wrap(err, "error processing event")
		}

		if err := store.StoreContractEvent(core.BlockchainEvent{
			BlockNumber:     l.BlockNumber,
			BlockchainID:    r.blockchainID,
			Name:            eventName,
			ContractAddress: l.Address.Hex(),
			TransactionHash: l.TxHash.String(),
			LogIndex:        uint32(l.Index),
		}); err != nil {
			logger.Warn("error storing contract event", "error", err, "event", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
			return errors.Wrap(err, "error storing contract event")
		}

		logger.Info("processed event", "event", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
		return nil
	})
	if r.onEventProcessed != nil {
		r.onEventProcessed(r.blockchainID, err == nil)
	}
	return err
}

func (r *ChannelHubReactor) handleNodeBalanceUpdated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseNodeBalanceUpdated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse NodeBalanceUpdated event")
	}

	asset, err := r.assetStore.GetTokenAsset(r.blockchainID, event.Token.String())
	if err != nil {
		return errors.Wrap(err, "failed to get token asset")
	}

	decimals, err := r.assetStore.GetTokenDecimals(r.blockchainID, event.Token.String())
	if err != nil {
		return errors.Wrap(err, "failed to get token decimals")
	}

	balance := decimal.NewFromBigInt(event.Amount, -int32(decimals))

	ev := core.NodeBalanceUpdatedEvent{
		BlockchainID: r.blockchainID,
		Asset:        asset,
		Balance:      balance,
	}
	return r.eventHandler.HandleNodeBalanceUpdated(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelCreated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	logger := log.FromContext(ctx)

	event, err := channelHubFilterer.ParseChannelCreated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCreated event")
	}

	if common.HexToAddress(r.nodeAddress) != event.Definition.Node {
		logger.Warn("ignoring ChannelCreated for different node", "eventNode", event.Definition.Node.Hex(), "ourNode", r.nodeAddress)
		return nil
	}

	ev := core.HomeChannelCreatedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.InitialState.Version,
	}
	return r.eventHandler.HandleHomeChannelCreated(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelMigrated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseMigrationInInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationInInitiated event")
	}

	ev := core.HomeChannelMigratedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelMigrated(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelCheckpointed(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelCheckpointed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCheckpointed event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleChannelDeposited(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelDeposited(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelDeposited event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleChannelWithdrawn(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelWithdrawn event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelChallenged(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelChallenged event")
	}

	ev := core.HomeChannelChallengedEvent{
		ChannelID:       hexutil.Encode(event.ChannelId[:]),
		StateVersion:    event.Candidate.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleHomeChannelChallenged(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelClosed(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelClosed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelClosed event")
	}

	ev := core.HomeChannelClosedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.FinalState.Version,
	}
	return r.eventHandler.HandleHomeChannelClosed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositInitiated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiated event")
	}

	ev := core.EscrowDepositInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositInitiated(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositChallenged(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositChallenged event")
	}

	ev := core.EscrowDepositChallengedEvent{
		ChannelID:       hexutil.Encode(event.EscrowId[:]),
		StateVersion:    event.State.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleEscrowDepositChallenged(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositFinalized(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalized event")
	}

	ev := core.EscrowDepositFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositFinalized(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalInitiated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiated event")
	}

	ev := core.EscrowWithdrawalInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalInitiated(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalChallenged(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalChallenged event")
	}

	ev := core.EscrowWithdrawalChallengedEvent{
		ChannelID:       hexutil.Encode(event.EscrowId[:]),
		StateVersion:    event.State.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleEscrowWithdrawalChallenged(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalFinalized(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalFinalized event")
	}

	ev := core.EscrowWithdrawalFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalFinalized(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositInitiatedOnHome(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositFinalizedOnHome(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositFinalizedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalizedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalInitiatedOnHome(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalFinalizedOnHome(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalFinalizedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalFinalizedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, store, &ev)
}

// Additional event handlers for events not yet defined in core.BlockchainEventHandler

func (r *ChannelHubReactor) handleMigrationInFinalized(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseMigrationInFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationInFinalized event")
	}
	logger := log.FromContext(ctx)
	logger.Info("MigrationInFinalized event",
		"channelId", hexutil.Encode(event.ChannelId[:]),
		"stateVersion", event.State.Version)
	// TODO: Add handler method to core.BlockchainEventHandler interface and implement
	return nil
}

func (r *ChannelHubReactor) handleMigrationOutInitiated(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseMigrationOutInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationOutInitiated event")
	}
	logger := log.FromContext(ctx)
	logger.Info("MigrationOutInitiated event",
		"channelId", hexutil.Encode(event.ChannelId[:]),
		"stateVersion", event.State.Version)
	// TODO: Add handler method to core.BlockchainEventHandler interface and implement
	return nil
}

func (r *ChannelHubReactor) handleMigrationOutFinalized(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseMigrationOutFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationOutFinalized event")
	}
	logger := log.FromContext(ctx)
	logger.Info("MigrationOutFinalized event",
		"channelId", hexutil.Encode(event.ChannelId[:]),
		"stateVersion", event.State.Version)
	// TODO: Add handler method to core.BlockchainEventHandler interface and implement
	return nil
}

func (r *ChannelHubReactor) handleDeposited(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseDeposited(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Deposited event")
	}
	logger := log.FromContext(ctx)
	logger.Info("Deposited event",
		"token", event.Token.Hex(),
		"amount", event.Amount.String())

	return nil
}

func (r *ChannelHubReactor) handleWithdrawn(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Withdrawn event")
	}
	logger := log.FromContext(ctx)
	logger.Info("Withdrawn event",
		"token", event.Token.Hex(),
		"amount", event.Amount.String())

	return nil
}

func (r *ChannelHubReactor) handleEscrowDepositsPurged(ctx context.Context, store ChannelHubReactorStore, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositsPurged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositsPurged event")
	}
	logger := log.FromContext(ctx)
	logger.Info("EscrowDepositsPurged event", "purgedCount", event.PurgedCount.String())

	return nil
}
