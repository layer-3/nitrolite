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
	blockchainID       uint64
	eventHandler       core.ChannelHubEventHandler
	storeContractEvent StoreContractEvent
	onEventProcessed   func(blockchainID uint64, success bool)
}

func NewChannelHubReactor(blockchainID uint64, eventHandler core.ChannelHubEventHandler, storeContractEvent StoreContractEvent) *ChannelHubReactor {
	return &ChannelHubReactor{
		blockchainID:       blockchainID,
		eventHandler:       eventHandler,
		storeContractEvent: storeContractEvent,
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

	var err error
	switch eventID {
	case channelHubAbi.Events["ChannelCreated"].ID:
		err = r.handleHomeChannelCreated(ctx, l)
	case channelHubAbi.Events["ChannelCheckpointed"].ID:
		err = r.handleHomeChannelCheckpointed(ctx, l)
	case channelHubAbi.Events["ChannelDeposited"].ID:
		err = r.handleChannelDeposited(ctx, l)
	case channelHubAbi.Events["ChannelWithdrawn"].ID:
		err = r.handleChannelWithdrawn(ctx, l)
	case channelHubAbi.Events["ChannelChallenged"].ID:
		err = r.handleHomeChannelChallenged(ctx, l)
	case channelHubAbi.Events["ChannelClosed"].ID:
		err = r.handleHomeChannelClosed(ctx, l)
	case channelHubAbi.Events["EscrowDepositInitiated"].ID:
		err = r.handleEscrowDepositInitiated(ctx, l)
	case channelHubAbi.Events["EscrowDepositChallenged"].ID:
		err = r.handleEscrowDepositChallenged(ctx, l)
	case channelHubAbi.Events["EscrowDepositFinalized"].ID:
		err = r.handleEscrowDepositFinalized(ctx, l)
	case channelHubAbi.Events["EscrowWithdrawalInitiated"].ID:
		err = r.handleEscrowWithdrawalInitiated(ctx, l)
	case channelHubAbi.Events["EscrowWithdrawalChallenged"].ID:
		err = r.handleEscrowWithdrawalChallenged(ctx, l)
	case channelHubAbi.Events["EscrowWithdrawalFinalized"].ID:
		err = r.handleEscrowWithdrawalFinalized(ctx, l)
	case channelHubAbi.Events["EscrowDepositInitiatedOnHome"].ID:
		err = r.handleEscrowDepositInitiatedOnHome(ctx, l)
	case channelHubAbi.Events["EscrowDepositFinalizedOnHome"].ID:
		err = r.handleEscrowDepositFinalizedOnHome(ctx, l)
	case channelHubAbi.Events["EscrowWithdrawalInitiatedOnHome"].ID:
		err = r.handleEscrowWithdrawalInitiatedOnHome(ctx, l)
	case channelHubAbi.Events["EscrowWithdrawalFinalizedOnHome"].ID:
		err = r.handleEscrowWithdrawalFinalizedOnHome(ctx, l)
	// NOTE: Unimplemented handlers:
	case channelHubAbi.Events["MigrationInInitiated"].ID:
		err = r.handleHomeChannelMigrated(ctx, l)
	case channelHubAbi.Events["MigrationInFinalized"].ID:
		err = r.handleMigrationInFinalized(ctx, l)
	case channelHubAbi.Events["MigrationOutInitiated"].ID:
		err = r.handleMigrationOutInitiated(ctx, l)
	case channelHubAbi.Events["MigrationOutFinalized"].ID:
		err = r.handleMigrationOutFinalized(ctx, l)
	case channelHubAbi.Events["Deposited"].ID:
		err = r.handleDeposited(ctx, l)
	case channelHubAbi.Events["Withdrawn"].ID:
		err = r.handleWithdrawn(ctx, l)
	case channelHubAbi.Events["EscrowDepositsPurged"].ID:
		err = r.handleEscrowDepositsPurged(ctx, l)
	default:
		logger.Warn("unknown event: " + eventID.Hex())
	}
	if r.onEventProcessed != nil {
		r.onEventProcessed(r.blockchainID, err == nil)
	}
	if err != nil {
		logger.Warn("error processing event", "error", err)
		return errors.Wrap(err, "error processing event")
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
		return errors.Wrap(err, "error storing contract event")
	}

	logger.Info("processed event", "event", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
	return nil
}

func (r *ChannelHubReactor) handleHomeChannelCreated(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelCreated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCreated event")
	}

	ev := core.HomeChannelCreatedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.InitialState.Version,
	}
	return r.eventHandler.HandleHomeChannelCreated(ctx, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelMigrated(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseMigrationInInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse MigrationInInitiated event")
	}

	ev := core.HomeChannelMigratedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelMigrated(ctx, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelCheckpointed(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelCheckpointed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelCheckpointed event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleChannelDeposited(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelDeposited(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelDeposited event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleChannelWithdrawn(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelWithdrawn event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.Candidate.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleHomeChannelChallenged(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelChallenged(l)
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

func (r *ChannelHubReactor) handleHomeChannelClosed(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseChannelClosed(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse ChannelClosed event")
	}

	ev := core.HomeChannelClosedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.FinalState.Version,
	}
	return r.eventHandler.HandleHomeChannelClosed(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositInitiated(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiated event")
	}

	ev := core.EscrowDepositInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositInitiated(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositChallenged(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositChallenged(l)
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

func (r *ChannelHubReactor) handleEscrowDepositFinalized(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalized event")
	}

	ev := core.EscrowDepositFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowDepositFinalized(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalInitiated(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiated event")
	}

	ev := core.EscrowWithdrawalInitiatedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalInitiated(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalChallenged(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalChallenged(l)
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

func (r *ChannelHubReactor) handleEscrowWithdrawalFinalized(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalFinalized(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalFinalized event")
	}

	ev := core.EscrowWithdrawalFinalizedEvent{
		ChannelID:    hexutil.Encode(event.EscrowId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleEscrowWithdrawalFinalized(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositInitiatedOnHome(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowDepositFinalizedOnHome(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositFinalizedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositFinalizedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalInitiatedOnHome(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalInitiatedOnHome(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowWithdrawalInitiatedOnHome event")
	}

	ev := core.HomeChannelCheckpointedEvent{
		ChannelID:    hexutil.Encode(event.ChannelId[:]),
		StateVersion: event.State.Version,
	}
	return r.eventHandler.HandleHomeChannelCheckpointed(ctx, &ev)
}

func (r *ChannelHubReactor) handleEscrowWithdrawalFinalizedOnHome(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowWithdrawalFinalizedOnHome(l)
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

func (r *ChannelHubReactor) handleMigrationInFinalized(ctx context.Context, l types.Log) error {
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

func (r *ChannelHubReactor) handleMigrationOutInitiated(ctx context.Context, l types.Log) error {
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

func (r *ChannelHubReactor) handleMigrationOutFinalized(ctx context.Context, l types.Log) error {
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

func (r *ChannelHubReactor) handleDeposited(ctx context.Context, l types.Log) error {
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

func (r *ChannelHubReactor) handleWithdrawn(ctx context.Context, l types.Log) error {
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

func (r *ChannelHubReactor) handleEscrowDepositsPurged(ctx context.Context, l types.Log) error {
	event, err := channelHubFilterer.ParseEscrowDepositsPurged(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse EscrowDepositsPurged event")
	}
	logger := log.FromContext(ctx)
	logger.Info("EscrowDepositsPurged event", "purgedCount", event.PurgedCount.String())

	return nil
}
