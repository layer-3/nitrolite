package evm

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

// LockingContractReactorStoreTxHandler is a function that executes Store operations within a transaction.
// If the handler returns an error, the transaction is rolled back; otherwise it's committed.
type LockingContractReactorStoreTxHandler func(LockingContractReactorStore) error

// LockingContractReactorStoreTxProvider wraps Store operations in a database transaction.
// It accepts a LockingContractReactorStoreTxHandler and manages transaction lifecycle (begin, commit, rollback).
// Returns an error if the handler fails or the transaction cannot be committed.
type LockingContractReactorStoreTxProvider func(LockingContractReactorStoreTxHandler) error

// LockingContractReactorStore defines the persistence layer interface for channel and state data.
// All methods should be implemented to work within database transactions.
// Implementations are typically provided by the database layer and wrapped by LockingContractReactorStoreTxProvider.
type LockingContractReactorStore interface {
	// UpdateUserStaked updates the total staked amount for a user on a specific blockchain.
	UpdateUserStaked(wallet string, blockchainID uint64, amount decimal.Decimal) error

	// StoreContractEvent persists a processed contract event to the database, ensuring it's recorded in the same transaction as state updates.
	StoreContractEvent(ev core.BlockchainEvent) error
}

var lockingContractAbi *abi.ABI
var lockingContractFilterer *AppRegistryFilterer
var eventMapping map[common.Hash]string

func initLockingContract() {
	var err error
	lockingContractAbi, err = AppRegistryMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	// Create a filterer for parsing events (address not needed for parsing)
	contract := bind.NewBoundContract(common.Address{}, *lockingContractAbi, nil, nil, nil)
	lockingContractFilterer = &AppRegistryFilterer{contract: contract}

	eventMapping = make(map[common.Hash]string)
	for name, event := range lockingContractAbi.Events {
		eventMapping[event.ID] = name
	}
}

type LockingContractReactor struct {
	blockchainID     uint64
	eventHandler     core.LockingContractEventHandler
	useStoreInTx     LockingContractReactorStoreTxProvider
	tokenDecimals    int32
	onEventProcessed func(blockchainID uint64, success bool)
}

func NewLockingContractReactor(blockchainID uint64, eventHandler core.LockingContractEventHandler, getTokenDecimals func() (uint8, error), useStoreInTx LockingContractReactorStoreTxProvider) (*LockingContractReactor, error) {
	tokenDecimals, err := getTokenDecimals()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token decimals")
	}
	return &LockingContractReactor{
		blockchainID:  blockchainID,
		eventHandler:  eventHandler,
		tokenDecimals: int32(tokenDecimals),
		useStoreInTx:  useStoreInTx,
	}, nil
}

// SetOnEventProcessed sets an optional callback invoked after each event is processed.
func (r *LockingContractReactor) SetOnEventProcessed(fn func(blockchainID uint64, success bool)) {
	r.onEventProcessed = fn
}

func (r *LockingContractReactor) HandleEvent(ctx context.Context, l types.Log) error {
	logger := log.FromContext(ctx)

	eventID := l.Topics[0]
	eventName, ok := eventMapping[eventID]
	if !ok {
		logger.Warn("unknown event ID", "eventID", eventID.Hex(), "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)
		return nil
	}
	logger.Debug("received event", "name", eventName, "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)

	err := r.useStoreInTx(func(store LockingContractReactorStore) error {
		var err error
		switch eventID {
		case lockingContractAbi.Events["Locked"].ID:
			err = r.handleLocked(ctx, store, l)
		case lockingContractAbi.Events["Relocked"].ID:
			err = r.handleRelocked(ctx, store, l)
		case lockingContractAbi.Events["UnlockInitiated"].ID:
			err = r.handleUnlockInitiated(ctx, store, l)
		case lockingContractAbi.Events["Withdrawn"].ID:
			err = r.handleWithdrawn(ctx, store, l)
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

func (r *LockingContractReactor) handleLocked(ctx context.Context, store LockingContractReactorStore, l types.Log) error {
	event, err := lockingContractFilterer.ParseLocked(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Locked event")
	}

	ev := core.UserLockedBalanceUpdatedEvent{
		UserAddress:  strings.ToLower(event.User.String()),
		BlockchainID: r.blockchainID,
		Balance:      decimal.NewFromBigInt(event.NewBalance, -r.tokenDecimals),
	}
	return r.eventHandler.HandleUserLockedBalanceUpdated(ctx, store, &ev)
}

func (r *LockingContractReactor) handleRelocked(ctx context.Context, store LockingContractReactorStore, l types.Log) error {
	event, err := lockingContractFilterer.ParseRelocked(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Relocked event")
	}
	ev := core.UserLockedBalanceUpdatedEvent{
		UserAddress:  strings.ToLower(event.User.String()),
		BlockchainID: r.blockchainID,
		Balance:      decimal.NewFromBigInt(event.Balance, -r.tokenDecimals),
	}
	return r.eventHandler.HandleUserLockedBalanceUpdated(ctx, store, &ev)
}

func (r *LockingContractReactor) handleUnlockInitiated(ctx context.Context, store LockingContractReactorStore, l types.Log) error {
	event, err := lockingContractFilterer.ParseUnlockInitiated(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Unlockinitiated event")
	}
	ev := core.UserLockedBalanceUpdatedEvent{
		UserAddress:  strings.ToLower(event.User.String()),
		BlockchainID: r.blockchainID,
		Balance:      decimal.Zero,
	}
	return r.eventHandler.HandleUserLockedBalanceUpdated(ctx, store, &ev)
}

func (r *LockingContractReactor) handleWithdrawn(ctx context.Context, store LockingContractReactorStore, l types.Log) error {
	event, err := lockingContractFilterer.ParseWithdrawn(l)
	if err != nil {
		return errors.Wrap(err, "failed to parse Withdrawn event")
	}
	ev := core.UserLockedBalanceUpdatedEvent{
		UserAddress:  strings.ToLower(event.User.String()),
		BlockchainID: r.blockchainID,
		Balance:      decimal.Zero,
	}
	return r.eventHandler.HandleUserLockedBalanceUpdated(ctx, store, &ev)
}
