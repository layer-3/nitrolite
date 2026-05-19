package evm

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/log"
)

// WatchValidatorRegistered subscribes to ValidatorRegistered events emitted by the
// ChannelHub contract at contractAddress and delivers them on the returned channel.
//
// Historical replay: if fromBlock > 0 the function first fetches all matching logs
// from fromBlock to the current chain head before switching to live events. This
// fills the gap between the last processed block and "now", ensuring no event is
// missed during a reconnect. Pass fromBlock = 0 to skip historical fetch (e.g. on
// the very first call when no prior state exists).
//
// The channel is closed when ctx is cancelled or the underlying subscription is
// lost. Callers should treat a closed channel as a signal to resubscribe, passing
// the BlockNumber of the last received event + 1 as fromBlock to avoid gaps.
//
// Reorg safety: logs marked Removed (chain reorganisation) are silently skipped.
//
// client must support event subscriptions (WebSocket or IPC transport).
func WatchValidatorRegistered(ctx context.Context, contractAddress common.Address, client EVMClient, blockchainID uint64, fromBlock uint64) (<-chan *core.ValidatorRegisteredEvent, error) {
	logger := log.FromContext(ctx).WithName("evm")

	topic := channelHubAbi.Events["ValidatorRegistered"].ID

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{topic}},
	}

	logCh := make(chan types.Log, 16)
	sub, err := client.SubscribeFilterLogs(ctx, query, logCh)
	if err != nil {
		return nil, errors.Wrap(err, "failed to subscribe to ValidatorRegistered events (ensure a WebSocket RPC endpoint is configured)")
	}

	eventCh := make(chan *core.ValidatorRegisteredEvent, 16)

	go func() {
		defer close(eventCh)
		defer sub.Unsubscribe()

		// headBlock is the upper bound of the historical getLogs query. Any live
		// event with BlockNumber ≤ headBlock was already fetched historically and
		// must be skipped to prevent duplicate delivery from the subscription-to-
		// getLogs transition window (the subscription starts before HeaderByNumber
		// returns, so logs in that window land in both streams).
		var headBlock uint64

		// Historical phase: fetch logs from fromBlock to the current head before
		// processing live events, so reconnects don't miss events emitted during
		// any outage window.
		if fromBlock > 0 {
			header, err := client.HeaderByNumber(ctx, nil)
			if err != nil {
				if ctx.Err() == nil {
					logger.Warn("failed to get chain head for historical ValidatorRegistered fetch — skipping gap fill", "error", err)
				}
			} else {
				headBlock = header.Number.Uint64()
				histQuery := ethereum.FilterQuery{
					FromBlock: new(big.Int).SetUint64(fromBlock),
					ToBlock:   header.Number,
					Addresses: []common.Address{contractAddress},
					Topics:    [][]common.Hash{{topic}},
				}
				histLogs, err := client.FilterLogs(ctx, histQuery)
				if err != nil {
					if ctx.Err() == nil {
						logger.Warn("failed to fetch historical ValidatorRegistered logs — gap fill incomplete", "error", err, "fromBlock", fromBlock)
					}
				} else {
					logger.Info("replaying historical ValidatorRegistered logs", "count", len(histLogs), "fromBlock", fromBlock, "toBlock", header.Number)
					for _, l := range histLogs {
						if ev := parseAndSend(ctx, eventCh, l, blockchainID, logger); ev == nil {
							return // ctx cancelled during delivery
						}
					}
				}
			}
		}

		// Live phase.
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				// Suppress the error log on clean ctx cancellation: go-ethereum delivers
				// a cancellation error on sub.Err() before ctx.Done() is scheduled,
				// which would otherwise log a spurious "subscription error" on shutdown.
				if err != nil && ctx.Err() == nil {
					logger.Error("ValidatorRegistered subscription error", "error", err, "contract", contractAddress.Hex(), "blockchainID", blockchainID)
				}
				return
			case l, ok := <-logCh:
				if !ok {
					return
				}
				if l.Removed {
					logger.Warn("skipping removed ValidatorRegistered log (reorg)", "blockchainID", blockchainID, "txHash", l.TxHash.Hex())
					continue
				}
				// Skip events already covered by the historical getLogs query to
				// prevent duplicate delivery from the subscription overlap window.
				if l.BlockNumber <= headBlock {
					continue
				}
				if parseAndSend(ctx, eventCh, l, blockchainID, logger) == nil {
					return
				}
			}
		}
	}()

	return eventCh, nil
}

// parseAndSend parses a ValidatorRegistered log and forwards it to eventCh.
// Returns the event on success, nil if ctx was cancelled before delivery.
func parseAndSend(ctx context.Context, eventCh chan<- *core.ValidatorRegisteredEvent, l types.Log, blockchainID uint64, logger log.Logger) *core.ValidatorRegisteredEvent {
	parsed, err := channelHubFilterer.ParseValidatorRegistered(l)
	if err != nil {
		logger.Error("failed to parse ValidatorRegistered log", "error", err, "txHash", l.TxHash.Hex())
		return &core.ValidatorRegisteredEvent{} // non-nil signals caller to continue
	}
	ev := &core.ValidatorRegisteredEvent{
		BlockchainID: blockchainID,
		ValidatorID:  parsed.ValidatorId,
		Validator:    parsed.Validator.Hex(),
		BlockNumber:  l.BlockNumber,
	}
	logger.Info("ValidatorRegistered event", "blockchainID", blockchainID, "validatorID", ev.ValidatorID, "validator", ev.Validator, "block", ev.BlockNumber)
	select {
	case eventCh <- ev:
		return ev
	case <-ctx.Done():
		return nil
	}
}
