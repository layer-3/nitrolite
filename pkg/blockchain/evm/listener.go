package evm

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
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
	contractAddress       common.Address
	client                EVMClient
	blockchainID          uint64
	blockStep             uint64        // max blocks per FilterLogs call during reconciliation
	confirmationDelay     time.Duration // routing threshold for Phase 1 events; 0 disables age-based routing
	logger                log.Logger
	handleEvent           HandleEvent // live events and recent historical events; typically the ConfirmationGate
	handleHistoricalEvent HandleEvent // historical events older than confirmationDelay; typically the reactor directly
	eventGetter           ContractEventGetter
	flushDownstream       func() // optional: called between reconnect iterations to flush gate pending state; nil = no-op

	// Single-entry block-timestamp cache for ensureBlockTimestamp. The listener's
	// processEvents loop is strictly serial (Phase 1 drains before Phase 2, each
	// phase processes one event at a time), so these fields require no mutex.
	lastBlockHash      common.Hash
	lastBlockTimestamp time.Time
}

// NewListener creates a Listener. blockStep controls how many blocks are fetched
// per RPC call during historical reconciliation.
//
// confirmationDelay controls per-event routing for Phase 1 (historical) events:
//   - When 0: every historical event is routed to historicalEventHandler.
//   - When > 0: each event's block timestamp is fetched via HeaderByHash. Events older
//     than confirmationDelay are routed to historicalEventHandler (their block is past
//     the reorg window, so they are safe to forward directly). Events younger than
//     confirmationDelay are routed to eventHandler so they pass through the gate —
//     historical replay reaching very recent blocks is no safer than live delivery
//     and the gate must still protect against reorgs of those blocks.
//
// Live (Phase 2) events always flow to eventHandler.
//
// eventHandler is typically the ConfirmationGate; historicalEventHandler is typically
// the reactor directly. The two handlers may be the same function when no gate is in use.
//
// flushDownstream is an optional callback called between reconnect iterations to flush
// in-flight gate pending state before re-reading the committed cursor via
// findCommonAncestor. When nil (the no-gate path, confirmationDelay == 0), the call
// is skipped. See ConfirmationGate.FlushPending and nitronode/docs/reorg-fix.md §6.9.
func NewListener(contractAddress common.Address, client EVMClient, blockchainID uint64, blockStep uint64, confirmationDelay time.Duration, logger log.Logger, eventHandler HandleEvent, historicalEventHandler HandleEvent, eventGetter ContractEventGetter, flushDownstream func()) *Listener {
	return &Listener{
		contractAddress:       contractAddress,
		client:                client,
		blockchainID:          blockchainID,
		blockStep:             blockStep,
		confirmationDelay:     confirmationDelay,
		logger:                logger.WithName("evm"),
		handleEvent:           eventHandler,
		handleHistoricalEvent: historicalEventHandler,
		eventGetter:           eventGetter,
		flushDownstream:       flushDownstream,
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

// listenEvents is the main reconnect loop. Each iteration:
//  1. Delegates to runOneListenPass, which resolves the committed cursor, subscribes,
//     reconciles historical, and processes live events.
//  2. On subscription drop (retry=true, err=nil): flushes downstream gate pending
//     state immediately (see §6.9 of reorg-fix.md), increments backoff, then loops.
//  3. On fatal handler/check failure (err != nil): returns the error immediately.
//
// The flush runs immediately after runOneListenPass returns retry=true, BEFORE the
// next iteration's backoff sleep. This ordering is deliberate: the gate's drain
// goroutine runs on an independent timer and can mature a pending orphan during the
// backoff sleep. Flushing first ensures no orphaned entry escapes before
// findCommonAncestor re-reads the committed-only cursor on the next pass.
// The first iteration naturally skips the flush: no retry path has executed yet
// and the gate is empty on initial entry.
//
// findCommonAncestor is called inside each runOneListenPass so the committed-only
// DB cursor is always re-read after a reconnect. The in-memory lastBlock from the
// previous pass is discarded — the per-pass local variable never escapes.
// Any future code that adds cross-iteration state to listenEvents must be aware
// that runOneListenPass is the iteration unit and per-pass state must be re-derived.
//
// Returns non-nil only when the handler or the event-presence check fails.
func (l *Listener) listenEvents(ctx context.Context) error {
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

		retry, err := l.runOneListenPass(ctx, &backOffCount)
		if err != nil {
			return err
		}
		if !retry {
			return nil // graceful shutdown inside the pass
		}

		// Subscription drop — flush gate pending state BEFORE the next iteration's
		// backoff sleep so the gate's drain goroutine cannot mature an orphan during
		// the sleep and forward it to the reactor before findCommonAncestor re-reads
		// the committed cursor. Flushing here ensures Phase 1 on the next pass
		// re-covers the entire [committedCursor, tip] range, which by construction
		// includes every block the gate was holding (uncommitted events sit above
		// the committed cursor). See nitronode/docs/reorg-fix.md §6.9.
		if l.flushDownstream != nil {
			l.flushDownstream()
		}
		backOffCount.Add(1)
	}
}

// runOneListenPass executes one full startup-style pass:
//  1. findCommonAncestor — reads the committed-only DB cursor (fresh each pass).
//  2. SubscribeFilterLogs — opens the live subscription.
//  3. HeaderByNumber(nil) — fetches the chain tip.
//  4. reconcileBlockRange goroutine — Phase 1 historical replay.
//  5. processEvents — drains Phase 1 then processes Phase 2 live events.
//
// Returns (true, nil) on subscription drop (caller retries after flushing the gate
// and applying backoff). Returns (false, nil) on graceful context shutdown.
// Returns (_, err) on fatal handler/check failure.
//
// findCommonAncestor MUST live inside this function for cursor-lifecycle correctness:
// if it were called once in the outer loop, the in-memory lastBlock advanced by
// Phase 2 (listener.go ← processEvents) would persist across reconnects and Phase 1
// on the retry would scan only (lastLiveBlock, tip] — missing every uncommitted event
// that was flushed from the gate. See "Cursor-lifecycle defect" in pr-832-open-comments.md.
func (l *Listener) runOneListenPass(ctx context.Context, backOffCount *atomic.Uint64) (retry bool, err error) {
	lastBlock, err := findCommonAncestor(ctx, l.client, l.eventGetter, l.contractAddress.String(), l.blockchainID, l.logger)
	if err != nil {
		return false, fmt.Errorf("failed to find common ancestor: %w", err)
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
		return true, nil
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
			return true, nil
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
		return false, err
	}

	// processEvents returned nil — subscription drop or graceful ctx shutdown.
	// Distinguish via ctx to allow the caller to skip backoff on clean shutdown.
	if ctx.Err() != nil {
		return false, nil
	}
	return true, nil
}

// processEvents runs two sequential phases: historical (historicalCh until closed),
// then live (currentCh until ctx or subscription death). In each phase the first
// events are checked via IsContractEventProcessed; once a non-present event is found
// the check is skipped for the rest of that phase (events are strictly ordered).
// Returns nil on subscription loss (reconnect), non-nil on handler/check failure.
//
// Both the listener (here) and the reactor (channel_hub_reactor.go) call
// IsContractEventProcessed, so both share a dependency on DB availability. A
// transient Postgres hiccup at either call site surfaces the error, unsubscribes,
// and restarts the process — consistent behavior across the pipeline.
//
// Listener ordering & idempotency invariant
// -----------------------------------------
// Downstream handlers (and any code reasoning about the relative arrival order
// of on-chain events) may rely on the following guarantees provided by this
// loop. Changes that weaken any of them must update every consumer that cites
// this invariant by name.
//
//  1. Strict per-contract ordering. Within a single Listener, events are
//     delivered to handleEvent in ascending (block_number, log_index) order
//     across the historical → live transition. Phase 1 drains historicalCh to
//     completion before phase 2 reads from currentCh, and the upstream
//     reconcileBlockRange + live subscription preserve chain order within each
//     phase.
//
//  2. Idempotent resume. On restart, IsContractEventProcessed gates the first
//     event of each phase: events already persisted in a prior run are skipped
//     rather than reprocessed. Once a non-present event is seen the check is
//     dropped for the remainder of the phase (safe because of guarantee 1).
//     The dedup check identifies events by (txHash, logIndex, blockchainID);
//     reorged events with a re-shuffled block-level log index are not detected
//     here and rely on reactor business-logic idempotency.
//
//  3. Cursor advances only on handler success. lastBlock is updated on each
//     live event, but a non-nil return from handleEvent unsubscribes and
//     surfaces the error to the caller without persisting any state past the
//     failed event; the next Listen invocation re-fetches from the same
//     cursor. Transient handler failures retry instead of silently dropping.
//
//  4. Reorged-out logs are routed by delay configuration.
//     When confirmationDelay > 0, live deliveries with Removed=true are
//     forwarded to the handler (ConfirmationGate) so the gate can cancel
//     any pending confirmation timer for that event; the gate filters them
//     before forwarding confirmed events to the reactor. When
//     confirmationDelay == 0, there is no gate to consume the removal
//     signal, so the listener drops Removed=true logs at the Phase 2
//     boundary — matching pre-PR behavior. In both modes the reactor
//     never sees Removed=true logs directly. The lastBlock cursor and
//     IsContractEventProcessed dedup check are skipped for Removed=true
//     events so neither the resume cursor nor the idempotency guard is
//     corrupted by a reorg signal.
//
// A consequence used by the nitronode event handlers: for any channel that
// closes via Path-1 (challenge-timeout, ChannelHub Closed-from-DISPUTED),
// HandleHomeChannelChallenged is guaranteed to run before HandleHomeChannelClosed
// for that channel. See nitronode/event_handlers/service.go (audit finding
// MF3-I01) for the wedge case this rules out.
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
				present, err := l.eventGetter.IsContractEventProcessed(eventLog.TxHash.Hex(), uint32(eventLog.Index), l.blockchainID)
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
			eventLog, err := l.ensureBlockTimestamp(ctx, eventLog)
			if err != nil {
				l.logger.Warn("failed to ensure block timestamp for historical event, routing through gate",
					"error", err,
					"blockchainID", l.blockchainID,
					"blockNumber", eventLog.BlockNumber,
					"blockHash", eventLog.BlockHash.Hex(),
				)
				if err := l.handleEvent(evCtx, eventLog); err != nil {
					eventSubscription.Unsubscribe()
					return err
				}
				continue
			}
			handler := l.routeHistoricalEvent(eventLog)
			if err := handler(evCtx, eventLog); err != nil {
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
			if eventLog.Removed && l.confirmationDelay == 0 {
				l.logger.Warn("dropping Removed=true live event on no-gate path",
					"blockchainID", l.blockchainID,
					"contractAddress", l.contractAddress.String(),
					"blockNumber", eventLog.BlockNumber,
					"blockHash", eventLog.BlockHash.Hex(),
					"txHash", eventLog.TxHash.Hex(),
					"logIndex", eventLog.Index,
				)
				continue
			}
			if !eventLog.Removed {
				*lastBlock = eventLog.BlockNumber
				if !currentCheckDone {
					present, err := l.eventGetter.IsContractEventProcessed(eventLog.TxHash.Hex(), uint32(eventLog.Index), l.blockchainID)
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
			}
			evCtx := log.SetContextLogger(context.Background(), l.logger)
			if !eventLog.Removed {
				ensured, err := l.ensureBlockTimestamp(ctx, eventLog)
				if err != nil {
					l.logger.Warn("failed to ensure block timestamp for current event, routing through gate",
						"error", err,
						"blockchainID", l.blockchainID,
						"blockNumber", eventLog.BlockNumber,
						"blockHash", eventLog.BlockHash.Hex(),
					)
				} else {
					eventLog = ensured
				}
			}
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

// reconcileBlockRange fetches logs in [lastBlock, currentBlock] (inclusive both bounds)
// in blockStep-sized windows, sending each log to historicalCh. Caller closes historicalCh
// after return. Uses a dedicated context so it can be cancelled when the subscription drops.
func (l *Listener) reconcileBlockRange(
	ctx context.Context,
	currentBlock uint64,
	lastBlock uint64,
	historicalCh chan types.Log,
) {
	var backOffCount atomic.Uint64
	startBlock := lastBlock
	endBlock := startBlock + l.blockStep

	// Inclusive at the lower bound so the all-reorged orphan-height case
	// (lastBlock == currentBlock) still fetches the canonical replacement;
	// downstream dedup absorbs the inevitable re-fetch on restart-at-exact-tip.
	for currentBlock >= startBlock {
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

// ensureBlockTimestamp returns eventLog with BlockTimestamp guaranteed non-zero.
//
// Most EVM chains and providers populate BlockTimestamp in the JSON-RPC response,
// in which case eventLog is returned unchanged. For chains/providers that do NOT
// populate it (notably Avalanche C-Chain via ava-labs/libevm, and older BSC
// dataseed nodes), this method fetches the block header via HeaderByHash and
// populates the field on the local-stack copy of types.Log.
//
// Single-entry cache (lastBlockHash) elides repeat fetches for consecutive events
// from the same block — the only relevant case because the listener delivers events
// in block order.
//
// On HeaderByHash failure, returns the original eventLog and the error. Callers
// decide whether to fall back to the gate (which is the conservative behavior;
// see live-path and routeHistoricalEvent below).
func (l *Listener) ensureBlockTimestamp(ctx context.Context, eventLog types.Log) (types.Log, error) {
	if eventLog.BlockTimestamp != 0 {
		return eventLog, nil
	}

	if eventLog.BlockHash == l.lastBlockHash && !l.lastBlockTimestamp.IsZero() {
		eventLog.BlockTimestamp = uint64(l.lastBlockTimestamp.Unix())
		return eventLog, nil
	}

	headerCtx, cancel := context.WithTimeout(ctx, rpcRequestTimeout)
	defer cancel()
	header, err := l.client.HeaderByHash(headerCtx, eventLog.BlockHash)
	if err != nil {
		return eventLog, err
	}

	blockTime := time.Unix(int64(header.Time), 0)
	l.lastBlockHash = eventLog.BlockHash
	l.lastBlockTimestamp = blockTime
	eventLog.BlockTimestamp = header.Time
	return eventLog, nil
}

// routeHistoricalEvent chooses the handler for a Phase 1 event based on the age of
// its block. Events whose block timestamp is older than confirmationDelay are routed
// to handleHistoricalEvent (they are past the reorg window and safe to forward
// directly). Recent events — whose blocks may still be reorged — are routed to
// handleEvent so they pass through the gate. When confirmationDelay is zero, every
// event is routed to handleHistoricalEvent.
//
// Reads eventLog.BlockTimestamp directly — callers are expected to have invoked
// ensureBlockTimestamp first. Defense-in-depth: if BlockTimestamp is zero (caller
// failed to ensure it), route through handleEvent (the gate) as the conservative
// choice.
func (l *Listener) routeHistoricalEvent(eventLog types.Log) HandleEvent {
	if l.confirmationDelay == 0 {
		return l.handleHistoricalEvent
	}

	if eventLog.BlockTimestamp == 0 {
		return l.handleEvent
	}

	blockTime := time.Unix(int64(eventLog.BlockTimestamp), 0)
	if time.Since(blockTime) < l.confirmationDelay {
		return l.handleEvent
	}
	return l.handleHistoricalEvent
}
