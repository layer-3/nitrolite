package evm

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

const (
	maxBackOffCount = 5
)

type Listener struct {
	contractAddress common.Address
	client          bind.ContractBackend
	blockchainID    uint64
	blockStep       uint64
	logger          log.Logger
	handleEvent     HandleEvent
	getLatestEvent  LatestEventGetter
}

func NewListener(contractAddress common.Address, client bind.ContractBackend, blockchainID uint64, blockStep uint64, logger log.Logger, eventHandler HandleEvent, getLatestEvent LatestEventGetter) *Listener {
	if getLatestEvent == nil {
		getLatestEvent = func(contractAddress string, networkID uint64) (core.BlockchainEvent, error) {
			return core.BlockchainEvent{
				BlockNumber: 0,
				LogIndex:    0,
			}, nil
		}
	}
	return &Listener{
		contractAddress: contractAddress,
		client:          client,
		blockchainID:    blockchainID,
		blockStep:       blockStep,
		logger:          logger.WithName("evm"),
		handleEvent:     eventHandler,
		getLatestEvent:  getLatestEvent,
	}
}

// Listen starts the event listener in a background goroutine.
// The handleClosure callback is invoked when the listener exits, with an error if any.
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
		defer childHandleClosure(nil)
		l.listenEvents(childCtx)
	}()

	go func() {
		wg.Wait()

		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		handleClosure(closureErr)
	}()
}

// listenEvents listens for blockchain events and processes them with the provided handler
func (l *Listener) listenEvents(ctx context.Context) {
	ev, err := l.getLatestEvent(l.contractAddress.String(), l.blockchainID)
	if err != nil {
		l.logger.Error("failed to get latest processed event", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
		return
	}
	lastBlock := ev.BlockNumber
	lastIndex := ev.LogIndex

	var backOffCount atomic.Uint64
	var historicalCh, currentCh chan types.Log
	var eventSubscription event.Subscription

	l.logger.Info("starting listening events", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
	for {
		if eventSubscription == nil {
			waitForBackOffTimeout(l.logger, int(backOffCount.Load()), "event subscription")

			historicalCh = make(chan types.Log, 1)
			currentCh = make(chan types.Log, 100)

			if lastBlock == 0 {
				l.logger.Info("skipping historical logs fetching",
					"blockchainID", l.blockchainID,
					"contractAddress", l.contractAddress.String())
			} else {
				var header *types.Header
				var err error
				headerCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				header, err = l.client.HeaderByNumber(headerCtx, nil)
				cancel()
				if err != nil {
					l.logger.Error("failed to get latest block", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
					backOffCount.Add(1)
					continue
				}

				// TODO: ensure that new events start to be processed only after all historical ones are processed
				go l.reconcileBlockRange(
					header.Number.Uint64(),
					lastBlock,
					lastIndex,
					historicalCh,
				)
			}

			watchFQ := ethereum.FilterQuery{
				Addresses: []common.Address{l.contractAddress},
			}
			eventSub, err := l.client.SubscribeFilterLogs(context.Background(), watchFQ, currentCh)
			if err != nil {
				l.logger.Error("failed to subscribe on events", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
				backOffCount.Add(1)
				continue
			}

			eventSubscription = eventSub
			l.logger.Info("watching events", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			backOffCount.Store(0)
		}

		select {
		case <-ctx.Done():
			l.logger.Info("stopping event listener", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			return
		case eventLog := <-historicalCh:
			l.logger.Debug("received new event", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)

			ctx := log.SetContextLogger(context.Background(), l.logger)
			l.handleEvent(ctx, eventLog)
		case eventLog := <-currentCh:
			lastBlock = eventLog.BlockNumber
			l.logger.Debug("received new event", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)

			ctx := log.SetContextLogger(context.Background(), l.logger)
			l.handleEvent(ctx, eventLog)
		case err := <-eventSubscription.Err():
			if err != nil {
				l.logger.Error("event subscription error", "error", err, "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
				eventSubscription.Unsubscribe()
				// NOTE: do not increment backOffCount here, as connection errors on continuous subscriptions are normal
			} else {
				l.logger.Debug("subscription closed, resubscribing", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String())
			}

			eventSubscription = nil
		}
	}
}

func (l *Listener) reconcileBlockRange(
	currentBlock uint64,
	lastBlock uint64,
	lastIndex uint32,
	historicalCh chan types.Log,
) {
	var backOffCount atomic.Uint64
	startBlock := lastBlock
	endBlock := startBlock + l.blockStep

	for currentBlock > startBlock {
		waitForBackOffTimeout(l.logger, int(backOffCount.Load()), "reconcile block range")

		// We need to refetch events starting from last known block without adding 1 to it
		// because it's possible that block includes more than 1 event, and some may be still unprocessed.
		//
		// This will cause duplicate key error in logs, but it's completely fine.
		if endBlock > currentBlock {
			endBlock = currentBlock
		}

		fetchFQ := ethereum.FilterQuery{
			Addresses: []common.Address{l.contractAddress},
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
			// Topics:    topics,
		}

		logsCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		logs, err := l.client.FilterLogs(logsCtx, fetchFQ)
		cancel()
		if err != nil {
			// TODO: divide previous block range by 2
			backOffCount.Add(1)
			l.logger.Error("failed to filter logs",
				"error", err,
				"blockchainID", l.blockchainID,
				"contractAddress", l.contractAddress.String(),
				"startBlock", startBlock,
				"endBlock", endBlock)
			continue // retry with the advised block range
		}
		l.logger.Info("fetched historical logs",
			"blockchainID", l.blockchainID,
			"contractAddress", l.contractAddress.String(),
			"count", len(logs),
			"startBlock", startBlock,
			"endBlock", endBlock)

		for _, ethLog := range logs {
			// Filter out previously known events
			if ethLog.BlockNumber == lastBlock && ethLog.Index <= uint(lastIndex) {
				l.logger.Info("skipping previously known event", "blockchainID", l.blockchainID, "contractAddress", l.contractAddress.String(), "blockNumber", ethLog.BlockNumber, "logIndex", ethLog.Index)
				continue
			}

			historicalCh <- ethLog
		}

		startBlock = endBlock + 1
		endBlock += l.blockStep
	}
}
