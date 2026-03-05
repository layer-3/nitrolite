package evm

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

var contractAbi *abi.ABI
var contractFilterer *ChannelHubFilterer
var eventMapping map[common.Hash]string

func init() {
	var err error
	contractAbi, err = ChannelHubMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	// Create a filterer for parsing events (address not needed for parsing)
	contract := bind.NewBoundContract(common.Address{}, *contractAbi, nil, nil, nil)
	contractFilterer = &ChannelHubFilterer{contract: contract}

	eventMapping = make(map[common.Hash]string)
	for name, event := range contractAbi.Events {
		eventMapping[event.ID] = name
	}
}

type Reactor struct {
	blockchainID       uint64
	eventHandler       core.BlockchainEventHandler
	storeContractEvent StoreContractEvent
	onEventProcessed   func(blockchainID uint64, success bool)
}

func NewReactor(blockchainID uint64, eventHandler core.BlockchainEventHandler, storeContractEvent StoreContractEvent) *Reactor {
	return &Reactor{
		blockchainID:       blockchainID,
		eventHandler:       eventHandler,
		storeContractEvent: storeContractEvent,
	}
}

// SetOnEventProcessed sets an optional callback invoked after each event is processed.
func (r *Reactor) SetOnEventProcessed(fn func(blockchainID uint64, success bool)) {
	r.onEventProcessed = fn
}

func (r *Reactor) HandleEvent(ctx context.Context, l types.Log) {
	logger := log.FromContext(ctx)

	eventID := l.Topics[0]
	eventName, ok := eventMapping[eventID]
	if !ok {
		logger.Warn("unknown event ID", "eventID", eventID.Hex(), "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
		return
	}
	logger.Debug("received event", "name", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)

	var err error
	switch eventID {
	case contractAbi.Events["ChannelCreated"].ID:
		err = r.handleHomeChannelCreated(ctx, l)
	case contractAbi.Events["ChannelCheckpointed"].ID:
		err = r.handleHomeChannelCheckpointed(ctx, l)
	case contractAbi.Events["ChannelDeposited"].ID:
		err = r.handleChannelDeposited(ctx, l)
	case contractAbi.Events["ChannelWithdrawn"].ID:
		err = r.handleChannelWithdrawn(ctx, l)
	case contractAbi.Events["ChannelChallenged"].ID:
		err = r.handleHomeChannelChallenged(ctx, l)
	case contractAbi.Events["ChannelClosed"].ID:
		err = r.handleHomeChannelClosed(ctx, l)
	case contractAbi.Events["EscrowDepositInitiated"].ID:
		err = r.handleEscrowDepositInitiated(ctx, l)
	case contractAbi.Events["EscrowDepositChallenged"].ID:
		err = r.handleEscrowDepositChallenged(ctx, l)
	case contractAbi.Events["EscrowDepositFinalized"].ID:
		err = r.handleEscrowDepositFinalized(ctx, l)
	case contractAbi.Events["EscrowWithdrawalInitiated"].ID:
		err = r.handleEscrowWithdrawalInitiated(ctx, l)
	case contractAbi.Events["EscrowWithdrawalChallenged"].ID:
		err = r.handleEscrowWithdrawalChallenged(ctx, l)
	case contractAbi.Events["EscrowWithdrawalFinalized"].ID:
		err = r.handleEscrowWithdrawalFinalized(ctx, l)
	case contractAbi.Events["EscrowDepositInitiatedOnHome"].ID:
		err = r.handleEscrowDepositInitiatedOnHome(ctx, l)
	case contractAbi.Events["EscrowDepositFinalizedOnHome"].ID:
		err = r.handleEscrowDepositFinalizedOnHome(ctx, l)
	case contractAbi.Events["EscrowWithdrawalInitiatedOnHome"].ID:
		err = r.handleEscrowWithdrawalInitiatedOnHome(ctx, l)
	case contractAbi.Events["EscrowWithdrawalFinalizedOnHome"].ID:
		err = r.handleEscrowWithdrawalFinalizedOnHome(ctx, l)
	// NOTE: Unimplemented handlers:
	case contractAbi.Events["MigrationInInitiated"].ID:
		err = r.handleHomeChannelMigrated(ctx, l)
	case contractAbi.Events["MigrationInFinalized"].ID:
		err = r.handleMigrationInFinalized(ctx, l)
	case contractAbi.Events["MigrationOutInitiated"].ID:
		err = r.handleMigrationOutInitiated(ctx, l)
	case contractAbi.Events["MigrationOutFinalized"].ID:
		err = r.handleMigrationOutFinalized(ctx, l)
	case contractAbi.Events["Deposited"].ID:
		err = r.handleDeposited(ctx, l)
	case contractAbi.Events["Withdrawn"].ID:
		err = r.handleWithdrawn(ctx, l)
	case contractAbi.Events["EscrowDepositsPurged"].ID:
		err = r.handleEscrowDepositsPurged(ctx, l)
	default:
		err = errors.New("unknown event: " + eventID.Hex())
	}
	if r.onEventProcessed != nil {
		r.onEventProcessed(r.blockchainID, err == nil)
	}
	if err != nil {
		logger.Warn("error processing event", "error", err)
	}

	if err := r.storeContractEvent(core.BlockchainEvent{
		BlockNumber:     l.BlockNumber,
		BlockchainID:    r.blockchainID,
		Name:            eventName,
		ContractAddress: l.Address.Hex(),
		TransactionHash: l.TxHash.String(),
		LogIndex:        uint32(l.Index),
	}); err != nil {
		logger.Warn("error storing contract event", "error", err, "event", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
	}

	logger.Info("processed event", "event", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
}

func (r *Reactor) handleHomeChannelCreated(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelCreated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCreated event")
	}

	ev := core.HomeChannelCreatedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.InitialState.Version,
	}
	return r.eventHandler.HandleHomeChannelCreated(ctx, &ev)
}

func (r *Reactor) handleHomeChannelMigrated(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseMigrationInInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationInInitiated event")
	}

	ev := core.HomeChannelMigratedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelMigrated(ctx, &ev)
}

func (r *Reactor) handleHomeChannelCheckpointed(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelCheckpointed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCheckpointed event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleChannelDeposited(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelDeposited(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelDeposited event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleChannelWithdrawn(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelWithdrawn event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleHomeChannelChallenged(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelChallenged event")
	}

	ev := core.HomeChannelChallengedEvent{
		ChannelID:       hexutil.Encode(event.ChannelId[:]),
		StateVersion:    event.Candidate.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleHomeChannelChallenged(ctx, &ev)
}

func (r *Reactor) handleHomeChannelClosed(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseChannelClosed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelClosed event")
	}

	ev := core.HomeChannelClosedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.FinalState.Version,
	}
	return r.eventHandler.HandleHomeChannelClosed(ctx, &ev)
}

func (r *Reactor) handleEscrowDepositInitiated(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiated event")
	}

	ev := core.EscrowDepositInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositInitiated(ctx, &ev)
}

func (r *Reactor) handleEscrowDepositChallenged(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositChallenged event")
	}

	ev := core.EscrowDepositChallengedEvent{
		ChannelID:       hexutil.Encode(event.EscrowId[:]),
		StateVersion:    event.State.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleEscrowDepositChallenged(ctx, &ev)
}

func (r *Reactor) handleEscrowDepositFinalized(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalized event")
	}

	ev := core.EscrowDepositFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositFinalized(ctx, &ev)
}

func (r *Reactor) handleEscrowWithdrawalInitiated(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowWithdrawalInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiated event")
	}

	ev := core.EscrowWithdrawalInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalInitiated(ctx, &ev)
}

func (r *Reactor) handleEscrowWithdrawalChallenged(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowWithdrawalChallenged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalChallenged event")
	}

	ev := core.EscrowWithdrawalChallengedEvent{
		ChannelID:       hexutil.Encode(event.EscrowId[:]),
		StateVersion:    event.State.Version,
		ChallengeExpiry: event.ChallengeExpireAt,
	}
	return r.eventHandler.HandleEscrowWithdrawalChallenged(ctx, &ev)
}

func (r *Reactor) handleEscrowWithdrawalFinalized(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowWithdrawalFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalFinalized event")
	}

	ev := core.EscrowWithdrawalFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalFinalized(ctx, &ev)
}

func (r *Reactor) handleEscrowDepositInitiatedOnHome(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleEscrowDepositFinalizedOnHome(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositFinalizedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalizedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleEscrowWithdrawalInitiatedOnHome(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowWithdrawalInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *Reactor) handleEscrowWithdrawalFinalizedOnHome(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowWithdrawalFinalizedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalFinalizedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

// Additional event handlers for events not yet defined in core.BlockchainEventHandler

func (r *Reactor) handleMigrationInFinalized(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseMigrationInFinalized(l)
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

func (r *Reactor) handleMigrationOutInitiated(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseMigrationOutInitiated(l)
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

func (r *Reactor) handleMigrationOutFinalized(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseMigrationOutFinalized(l)
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

func (r *Reactor) handleDeposited(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseDeposited(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Deposited event")
	}
	logger := log.FromContext(ctx)
	logger.Info("Deposited event",
		"wallet", event.Wallet.Hex(),
		"token", event.Token.Hex(),
		"amount", event.Amount.String())

	return nil
}

func (r *Reactor) handleWithdrawn(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Withdrawn event")
	}
	logger := log.FromContext(ctx)
	logger.Info("Withdrawn event",
		"wallet", event.Wallet.Hex(),
		"token", event.Token.Hex(),
		"amount", event.Amount.String())

	return nil
}

func (r *Reactor) handleEscrowDepositsPurged(ctx context.Context, l types.Log) error {
	event, err := contractFilterer.ParseEscrowDepositsPurged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositsPurged event")
	}
	logger := log.FromContext(ctx)
	logger.Info("EscrowDepositsPurged event", "purgedCount", event.PurgedCount.String())

	return nil
}
