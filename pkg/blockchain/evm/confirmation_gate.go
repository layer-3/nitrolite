package evm

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
)

const pollInterval = 50 * time.Millisecond
const recentMultiplier = 3 // recentlyForwarded entries are kept for (recentMultiplier × delay) to catch post-gate reorgs

// queueEntry holds a pending event waiting for the confirmation delay to expire.
type queueEntry struct {
	log       types.Log
	arrivedAt time.Time // block timestamp from fetcher; fallback time.Now() on error
}

// eventKey identifies an event by tx and log index; blockHash is intentionally excluded
// so that a reorg-replacement event (same tx, same index, different block) can match
// and cancel the original pending entry.
type eventKey struct {
	txHash   common.Hash
	logIndex uint
}

// forwardedKey identifies an event that has already been forwarded to the downstream
// handler; blockHash is included so a Removed notification from a different block fork
// does NOT falsely trigger post-gate reorg logic.
type forwardedKey struct {
	txHash    common.Hash
	blockHash common.Hash
	logIndex  uint
}

// ConfirmationGate buffers incoming events for a configurable delay before forwarding
// them to a downstream handler, providing a window to cancel events that are reorged
// out before the delay expires.
type ConfirmationGate struct {
	delay                 time.Duration
	chainID               uint64
	handler               HandleEvent
	blockTimestampFetcher func(blockHash common.Hash) (time.Time, error)

	mu                  sync.Mutex
	queue               []queueEntry
	recentlyForwarded map[forwardedKey]time.Time // TTL = recentMultiplier × delay; protected by mu
	// blockTimestampCache holds the timestamp for every block that has delivered at
	// least one event to the gate. It avoids a redundant RPC call when the same block
	// produces multiple events (e.g. a batch open with two ChannelDeposited logs).
	// Entries are evicted by the Poller once the block timestamp is older than
	// recentMultiplier × delay — by that point every event from the block has either
	// been forwarded or cancelled, so the entry will never be read again.
	blockTimestampCache map[common.Hash]time.Time // protected by mu
	logger              log.Logger
}

// NewConfirmationGate creates a ConfirmationGate that holds events for delay before
// forwarding them to handler. fetcher is called once per unique blockHash to obtain the
// block's timestamp, which is used as the event's arrivedAt reference. If fetcher fails,
// time.Now() is used as a fallback.
func NewConfirmationGate(
	delay time.Duration,
	chainID uint64,
	handler HandleEvent,
	fetcher func(blockHash common.Hash) (time.Time, error),
	logger log.Logger,
) *ConfirmationGate {
	return &ConfirmationGate{
		delay:                 delay,
		chainID:               chainID,
		handler:               handler,
		blockTimestampFetcher: fetcher,
		recentlyForwarded:     make(map[forwardedKey]time.Time),
		blockTimestampCache:   make(map[common.Hash]time.Time),
		logger:                logger.WithName("confirmation-gate"),
	}
}

// Start begins the polling goroutine that forwards matured entries to the downstream
// handler. If delay is zero the gate is fully transparent and no goroutine is started.
func (g *ConfirmationGate) Start(ctx context.Context) {
	if g.delay == 0 {
		return
	}
	go g.poll(ctx)
}

// HandleEvent is the entry point called by the upstream Listener for each event.
//
// When delay == 0 the gate is fully transparent: every event (including Removed ones)
// is forwarded to the downstream handler immediately.
//
// When delay > 0:
//   - A non-removed event is queued and will be forwarded after the confirmation delay.
//   - A removed event cancels its pending queue entry (pre-gate reorg), or — if the
//     entry was already forwarded — records a post-gate reorg warning.
func (g *ConfirmationGate) HandleEvent(ctx context.Context, eventLog types.Log) error {
	if g.delay == 0 {
		// Removed:true events are never forwarded to the reactor regardless of delay
		// setting — the reactor was never designed to handle them and has no guard on
		// Topics[0]. This preserves the pre-gate listener behavior of dropping reorged
		// logs before they reach any downstream handler.
		if eventLog.Removed {
			return nil
		}
		return g.handler(ctx, eventLog)
	}

	key := eventKey{txHash: eventLog.TxHash, logIndex: uint(eventLog.Index)}

	if !eventLog.Removed {
		// Fetch block timestamp, using cache to avoid redundant RPC calls.
		var ts time.Time

		g.mu.Lock()
		cached, hit := g.blockTimestampCache[eventLog.BlockHash]
		if hit {
			ts = cached
		}
		g.mu.Unlock()

		if !hit {
			fetched, err := g.blockTimestampFetcher(eventLog.BlockHash)
			if err != nil {
				g.logger.Warn("failed to fetch block timestamp, falling back to now",
					"error", err,
					"blockHash", eventLog.BlockHash.Hex(),
					"chainID", g.chainID,
				)
				// Use gate entry arrival time as a fallback to avoid blocking events indefinitely when the fetcher fails.
				ts = time.Now()
			} else {
				ts = fetched

				// Update cache for future events from the same block.
				g.mu.Lock()
				g.blockTimestampCache[eventLog.BlockHash] = ts
				g.mu.Unlock()
			}
		}

		g.mu.Lock()
		// Remove any existing queue entry for the same (txHash, logIndex) so that
		// a re-delivered event (after reorg, with different blockHash) replaces
		// the original and resets the confirmation timer.
		g.removeFromQueueByKey(key)
		g.queue = append(g.queue, queueEntry{log: eventLog, arrivedAt: ts})
		g.mu.Unlock()

		return nil
	}

	// eventLog.Removed == true: attempt pre-gate cancellation.
	g.mu.Lock()
	defer g.mu.Unlock()

	// Build the full key once; it is reused for both the queue scan and the
	// recentlyForwarded lookup. blockHash is included so that a Removed notification for
	// an old block does not accidentally cancel a re-mined entry with the same tx/logIndex
	// in a new block.
	fk := forwardedKey{txHash: eventLog.TxHash, blockHash: eventLog.BlockHash, logIndex: uint(eventLog.Index)}
	if g.removeFromQueueByFullKey(fk) {
		return nil
	}

	// Not in queue — check whether it was already forwarded (post-gate reorg).
	if _, ok := g.recentlyForwarded[fk]; ok {
		g.logger.Warn("post-gate reorg detected",
			"txHash", eventLog.TxHash.Hex(),
			"blockHash", eventLog.BlockHash.Hex(),
			"logIndex", eventLog.Index,
			"chainID", g.chainID,
		)
		delete(g.recentlyForwarded, fk)
		return nil
	}

	g.logger.Debug("removal for unknown/stale event",
		"txHash", eventLog.TxHash.Hex(),
		"blockHash", eventLog.BlockHash.Hex(),
		"logIndex", eventLog.Index,
		"chainID", g.chainID,
	)
	return nil
}

// removeFromQueueByKey removes the first queue entry matching key (ignores blockHash).
// Used when a non-removed re-delivery replaces an earlier entry for the same logical event.
// Caller must hold mu.
func (g *ConfirmationGate) removeFromQueueByKey(key eventKey) {
	for i, e := range g.queue {
		ek := eventKey{txHash: e.log.TxHash, logIndex: uint(e.log.Index)}
		if ek == key {
			g.queue = append(g.queue[:i], g.queue[i+1:]...)
			return
		}
	}
}

// removeFromQueueByFullKey removes the first queue entry matching txHash, blockHash, and
// logIndex. Used in the Removed handler so that a removal notification for an old block
// does not accidentally cancel a re-mined entry with the same tx/logIndex in a new block.
// Caller must hold mu.
func (g *ConfirmationGate) removeFromQueueByFullKey(fk forwardedKey) bool {
	for i, e := range g.queue {
		if e.log.TxHash == fk.txHash && e.log.BlockHash == fk.blockHash && uint(e.log.Index) == fk.logIndex {
			g.queue = append(g.queue[:i], g.queue[i+1:]...)
			return true
		}
	}
	return false
}

// poll is the background goroutine that wakes on each pollInterval tick, forwards
// all matured queue entries to the downstream handler, and evicts stale recentlyForwarded
// entries whose TTL (recentMultiplier × delay) has elapsed.
func (g *ConfirmationGate) poll(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.mu.Lock()
			now := time.Now()

			// Forward all entries whose confirmation delay has elapsed.
			for len(g.queue) > 0 && !g.queue[0].arrivedAt.Add(g.delay).After(now) {
				entry := g.queue[0]
				g.queue = g.queue[1:]

				fk := forwardedKey{
					txHash:    entry.log.TxHash,
					blockHash: entry.log.BlockHash,
					logIndex:  uint(entry.log.Index),
				}
				g.recentlyForwarded[fk] = now

				g.mu.Unlock()

				evCtx := log.SetContextLogger(context.Background(), g.logger)
				if err := g.handler(evCtx, entry.log); err != nil {
					g.logger.Error("handler error after confirmation delay",
						"error", err,
						"chainID", g.chainID,
					)
				}

				g.mu.Lock()
			}

			// Evict recentlyForwarded entries older than (recentMultiplier × delay).
			for k, forwardedAt := range g.recentlyForwarded {
				if now.Sub(forwardedAt) > recentMultiplier*g.delay {
					delete(g.recentlyForwarded, k)
				}
			}

			// Evict blockTimestampCache entries whose block timestamp is older than
			// (recentMultiplier × delay). The listener delivers events in block order,
			// so once a block is old enough, all of its events have been forwarded or
			// cancelled and the cached timestamp will never be read again.
			for bh, ts := range g.blockTimestampCache {
				if now.Sub(ts) > recentMultiplier*g.delay {
					delete(g.blockTimestampCache, bh)
				}
			}

			g.mu.Unlock()
		}
	}
}
