package evm

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
)

const (
	maxBackOffCount   = 5
	rpcRequestTimeout = 1 * time.Minute
)

// Listener watches a single contract for on-chain events, combining historical
// log reconciliation with a live WebSocket subscription to guarantee gap-free,
// deduplicated delivery even across restarts. Cancel the context passed to Listen
// for graceful shutdown.
type Listener struct {
	contractAddress common.Address
	client          bind.ContractBackend
	blockchainID    uint64
	blockStep       uint64 // max blocks per FilterLogs call during reconciliation
	logger          log.Logger
	handleEvent     HandleEvent
	eventGetter     ContractEventGetter
}

// NewListener creates a Listener. blockStep controls how many blocks are fetched
// per RPC call during historical reconciliation.
func NewListener(contractAddress common.Address, client bind.ContractBackend, blockchainID uint64, blockStep uint64, logger log.Logger, eventHandler HandleEvent, eventGetter ContractEventGetter) *Listener {
	return &Listener{
		contractAddress: contractAddress,
		client:          client,
		blockchainID:    blockchainID,
		blockStep:       blockStep,
		logger:          logger.WithName("evm"),
		handleEvent:     eventHandler,
		eventGetter:     eventGetter,
	}
}

// Listen starts the listener in a background goroutine. handleClosure is called
// exactly once after the listener stops; err is non-nil only if the handler failed.
func (l *Listener) Listen(ctx context.Context, handleClosure func(err error)) {
	childCtx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)

	var closureErr error
	var closureErrMu sync.Mutex
	childHandleClosure := func(err error) {
		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		if err != nil && closureErr == nil {
			closureErr = err
		}

		cancel()
		wg.Done()
	}

	go func() {
		childHandleClosure(l.listenEvents(childCtx))
	}()

	go func() {
		wg.Wait()

		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		handleClosure(closureErr)
	}()
}

// logBackOff computes the backoff duration and logs accordingly.
// Returns the duration and true if the caller should proceed, or false if the limit was exceeded (fatal logged).
func (l *Listener) logBackOff(count uint64, originator string) (time.Duration, bool) {
	d := backOffDuration(int(count))
	if d < 0 {
		l.logger.Fatal("back off limit reached, exiting", "originator", originator, "backOffCollisionCount", count)
		return 0, false
	}
	if d > 0 {
		l.logger.Info("backing off", "originator", originator, "backOffCollisionCount", count)
	}
	return d, true
}

// listenEvents is the main loop. Each iteration:
//  1. Subscribes to live events (buffered in currentCh).
//  2. Fetches the chain tip — done after subscribing so no events fall through the gap.
//  3. Launches reconcileBlockRange in a goroutine (lastBlock → chain tip → historicalCh).
//  4. Calls processEvents: drains historicalCh first, then switches to currentCh.
//
// On subscription failure it retries with exponential backoff. Returns non-nil only
// when the handler or the event-presence check fails.
func (l *Listener) listenEvents(ctx context.Context) error {
	lastBlock, err := l.eventGetter.GetLatestContractEventBlockNumber(l.contractAddress.String(), l.blockchainID)
	if err != nil {
		return fmt.Errorf("failed to get latest processed block: %w", err)
	}

	var backOffCount atomic.Uint64

	l.logger.Info("starting listening events", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
	for {
		d, ok := l.logBackOff(backOffCount.Load(), "event subscription")
		if !ok {
			return nil
		}
		select {
		case <-ctx.Done():
			l.logger.Info("stopping event listener", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			return nil
		case <-time.After(d):
		}
		if ctx.Err() != nil {
			l.logger.Info("stopping event listener", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			return nil
		}

		historicalCh := make(chan types.Log, 1)
		currentCh := make(chan types.Log, 100)

		// Subscribe to live events first so nothing is missed while reconciling.
		watchFQ := ethereum.FilterQuery{
			Addresses: []common.Address{l.contractAddress},
		}
		eventSubscription, err := l.client.SubscribeFilterLogs(context.Background(), watchFQ, currentCh)
		if err != nil {
			l.logger.Error("failed to subscribe on events", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			backOffCount.Add(1)
			continue
		}

		// Fetch current block height after subscribing to avoid a gap.
		var cancelReconcile context.CancelFunc
		if lastBlock == 0 {
			l.logger.Info("skipping historical logs fetching",
				"blockchainID", l.blockchainID,
				"contractAddress", l.contractAddress.String())
			close(historicalCh)
		} else {
			headerCtx, headerCancel := context.WithTimeout(context.Background(), rpcRequestTimeout)
			header, err := l.client.HeaderByNumber(headerCtx, nil)
			headerCancel()
			if err != nil {
				l.logger.Error("failed to get latest block", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
				eventSubscription.Unsubscribe()
				backOffCount.Add(1)
				continue
			}

			var reconcileCtx context.Context
			reconcileCtx, cancelReconcile = context.WithCancel(ctx)
			currentBlock := header.Number.Uint64()
			go func() {
				l.reconcileBlockRange(reconcileCtx, currentBlock, lastBlock, historicalCh)
				close(historicalCh)
			}()
		}

		l.logger.Info("watching events", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
		backOffCount.Store(0)

		err = l.processEvents(ctx, eventSubscription, historicalCh, currentCh, &lastBlock)
		if cancelReconcile != nil {
			cancelReconcile()
		}
		if err != nil {
			return err
		}
	}
}

// processEvents runs two sequential phases: historical (historicalCh until closed),
// then live (currentCh until ctx or subscription death). In each phase the first
// events are checked via IsContractEventPresent; once a non-present event is found
// the check is skipped for the rest of that phase (events are strictly ordered).
// Returns nil on subscription loss (reconnect), non-nil on handler/check failure.
func (l *Listener) processEvents(
	ctx context.Context,
	eventSubscription interface {
		Unsubscribe()
		Err() <-chan error
	},
	historicalCh <-chan types.Log,
	currentCh <-chan types.Log,
	lastBlock *uint64,
) error {
	// Phase 1: drain all historical events before processing live ones.
	historicalCheckDone := false
	for historicalCh != nil {
		select {
		case <-ctx.Done():
			l.logger.Info("stopping event listener", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			eventSubscription.Unsubscribe()
			return nil
		case eventLog, ok := <-historicalCh:
			if !ok {
				historicalCh = nil
				break
			}
			if !historicalCheckDone {
				present, err := l.eventGetter.IsContractEventPresent(l.blockchainID, eventLog.BlockNumber, eventLog.TxHash.Hex(), uint32(eventLog.Index))
				if err != nil {
					eventSubscription.Unsubscribe()
					return fmt.Errorf("failed to check historical event presence: %w", err)
				}
				if present {
					l.logger.Debug("skipping already present historical event", "blockchainID", l.blockchainID, "blockNumber", eventLog.BlockNumber, "logIndex", eventLog.Index)
					continue
				}
				historicalCheckDone = true
			}
			l.logger.Debug("received historical event", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String(), "blockNumber", eventLog.BlockNumber, "logIndex", eventLog.Index)
			evCtx := log.SetContextLogger(context.Background(), l.logger)
			if err := l.handleEvent(evCtx, eventLog); err != nil {
				eventSubscription.Unsubscribe()
				return err
			}
		case err := <-eventSubscription.Err():
			if err != nil {
				l.logger.Error("event subscription error", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			} else {
				l.logger.Debug("subscription closed, resubscribing", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			}
			eventSubscription.Unsubscribe()
			return nil
		}
	}

	// Phase 2: process live events from subscription.
	currentCheckDone := false
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("stopping event listener", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			eventSubscription.Unsubscribe()
			return nil
		case eventLog := <-currentCh:
			// During a chain reorganization geth re-delivers orphaned logs with
			// Removed: true. Skip them to avoid applying phantom state changes.
			if eventLog.Removed {
				l.logger.Warn("skipping removed log from reorg", "blockchainID", l.blockchainID, "blockNumber", eventLog.BlockNumber, "logIndex", eventLog.Index, "txHash", eventLog.TxHash.Hex())
				continue
			}
			*lastBlock = eventLog.BlockNumber
			if !currentCheckDone {
				present, err := l.eventGetter.IsContractEventPresent(l.blockchainID, eventLog.BlockNumber, eventLog.TxHash.Hex(), uint32(eventLog.Index))
				if err != nil {
					eventSubscription.Unsubscribe()
					return fmt.Errorf("failed to check current event presence: %w", err)
				}
				if present {
					l.logger.Debug("skipping already present current event", "blockchainID", l.blockchainID, "blockNumber", eventLog.BlockNumber, "logIndex", eventLog.Index)
					continue
				}
				currentCheckDone = true
			}
			l.logger.Debug("received current event", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String(), "blockNumber", eventLog.BlockNumber, "logIndex", eventLog.Index)
			evCtx := log.SetContextLogger(context.Background(), l.logger)
			if err := l.handleEvent(evCtx, eventLog); err != nil {
				eventSubscription.Unsubscribe()
				return err
			}
		case err := <-eventSubscription.Err():
			if err != nil {
				l.logger.Error("event subscription error", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			} else {
				l.logger.Debug("subscription closed, resubscribing", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			}
			eventSubscription.Unsubscribe()
			return nil
		}
	}
}

// reconcileBlockRange fetches logs from lastBlock to currentBlock in blockStep-sized
// windows, sending each log to historicalCh. Caller closes historicalCh after return.
// Uses a dedicated context so it can be cancelled when the subscription drops.
func (l *Listener) reconcileBlockRange(
	ctx context.Context,
	currentBlock uint64,
	lastBlock uint64,
	historicalCh chan types.Log,
) {
	var backOffCount atomic.Uint64
	startBlock := lastBlock
	endBlock := startBlock + l.blockStep

	for currentBlock > startBlock {
		d, ok := l.logBackOff(backOffCount.Load(), "reconcile block range")
		if !ok {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
		}
		if ctx.Err() != nil {
			return
		}

		if endBlock > currentBlock {
			endBlock = currentBlock
		}

		fetchFQ := ethereum.FilterQuery{
			Addresses: []common.Address{l.contractAddress},
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
		}

		logsCtx, cancel := context.WithTimeout(ctx, rpcRequestTimeout)
		logs, err := l.client.FilterLogs(logsCtx, fetchFQ)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			backOffCount.Add(1)
			l.logger.Error("failed to filter logs",
				"error", err,
				"blockchainID", l.blockchainID,
				"contractAddress", l.contractAddress.String(),
				"startBlock", startBlock,
				"endBlock", endBlock)
			continue
		}
		l.logger.Info("fetched historical logs",
			"blockchainID", l.blockchainID,
			"contractAddress", l.contractAddress.String(),
			"count", len(logs),
			"startBlock", startBlock,
			"endBlock", endBlock)

		for _, ethLog := range logs {
			select {
			case <-ctx.Done():
				return
			case historicalCh <- ethLog:
			}
		}

		startBlock = endBlock + 1
		endBlock += l.blockStep
	}
}

// TODO: the current reorg handling (skipping Removed logs) prevents new damage but
// does not undo side effects from the original delivery if it was already processed.
// A more robust approach is a confirmation buffer: hold live logs in memory keyed by
// block number, only apply them after N confirmations (new blocks on top), and discard
// any log that arrives with Removed: true while still in the buffer. This adds N blocks
// of latency (~12s × N on mainnet) but guarantees that only finalized events reach the
// handler. On L2s where reorgs are near-zero, the latency trade-off may not be worth it,
// so this should be configurable per chain.
