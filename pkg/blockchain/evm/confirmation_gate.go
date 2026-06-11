package evm

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
)

// recentMultiplier controls how long forwardedSet entries are retained:
// (recentMultiplier × delay). This is the window during which a post-gate
// Removed:true can be matched against a previously forwarded event and emit
// the post-gate reorg WARN.
const recentMultiplier = 3

// queueEntry holds a pending event waiting for the confirmation delay to expire.
type queueEntry struct {
	log       types.Log
	arrivedAt time.Time // derived from eventLog.BlockTimestamp; fallback time.Now() when zero
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

// forwardedExpiry pairs a forwardedKey with the wall-clock time at which the event
// was forwarded, for O(1) FIFO eviction from forwardedSet.
type forwardedExpiry struct {
	key         forwardedKey
	forwardedAt time.Time
}

// ConfirmationGate buffers incoming events for a configurable delay before forwarding
// them to a downstream handler, providing a window to cancel events that are reorged
// out before the delay expires.
//
// The gate is pure in-memory: it reads arrival time from eventLog.BlockTimestamp and
// performs no RPC. The caller (Listener) is responsible for ensuring BlockTimestamp
// is populated before invoking HandleEvent.
type ConfirmationGate struct {
	delay   time.Duration
	chainID uint64
	handler HandleEvent
	logger  log.Logger

	mu             sync.Mutex
	queue          []queueEntry               // append-tail, pop-head
	pending        map[eventKey]common.Hash   // live (txHash, logIndex) -> blockHash; source of truth for live entries
	forwardedSet   map[forwardedKey]time.Time // key -> forwardedAt
	forwardedQueue []forwardedExpiry          // FIFO of (key, forwardedAt) for O(1) eviction

	kick    chan struct{} // buffered 1; non-blocking sends
	timer   *time.Timer   // created in Start(ctx)
	fatalCh chan error    // buffered 1; first handler error wins; non-blocking send
	done    chan struct{} // closed once on fatal; gates run's select so the goroutine exits
}

// NewConfirmationGate creates a ConfirmationGate that holds events for delay before
// forwarding them to handler. delay must be > 0; delay <= 0 returns an error
// (the wiring layer is responsible for skipping gate construction when the operator
// configured delay == 0).
func NewConfirmationGate(
	delay time.Duration,
	chainID uint64,
	handler HandleEvent,
	logger log.Logger,
) (*ConfirmationGate, error) {
	if delay <= 0 {
		return nil, errors.New("confirmation gate requires delay > 0")
	}
	return &ConfirmationGate{
		delay:          delay,
		chainID:        chainID,
		handler:        handler,
		logger:         logger.WithName("confirmation-gate"),
		pending:        make(map[eventKey]common.Hash),
		forwardedSet:   make(map[forwardedKey]time.Time),
		forwardedQueue: nil,
		kick:           make(chan struct{}, 1),
		fatalCh:        make(chan error, 1),
		done:           make(chan struct{}),
	}, nil
}

// FatalErrors returns a read-only channel that receives the first handler error
// encountered after the confirmation delay. The channel is buffered (size 1);
// only the first error is delivered. When the channel fires, the gate's drain
// goroutine has already stopped forwarding. The listener should unsubscribe and
// return the error to trigger process restart and DB-cursor replay.
func (g *ConfirmationGate) FatalErrors() <-chan error {
	return g.fatalCh
}

// Start begins the background goroutine that forwards matured entries to the
// downstream handler. The timer is created here (tied to the goroutine's lifecycle)
// and stopped on shutdown. The goroutine exits when ctx is cancelled.
func (g *ConfirmationGate) Start(ctx context.Context) {
	g.timer = time.NewTimer(time.Hour) // arbitrary long initial; will be reset on first drain
	if !g.timer.Stop() {
		<-g.timer.C
	}
	go g.run(ctx)
}

// HandleEvent is the entry point called by the upstream Listener for each event.
//
// A non-removed event is queued and will be forwarded after the confirmation delay.
// A removed event cancels its pending queue entry (pre-gate reorg) or — if the entry
// was already forwarded — records a post-gate reorg warning.
func (g *ConfirmationGate) HandleEvent(_ context.Context, eventLog types.Log) error {
	ek := eventKey{txHash: eventLog.TxHash, logIndex: uint(eventLog.Index)}

	if !eventLog.Removed {
		// Derive arrival time from the event's block timestamp. The listener
		// guarantees this is non-zero in steady state; the fallback is
		// defense-in-depth for tests/edge cases. No log here — the listener
		// owns the warning when it cannot ensure the timestamp.
		var ts time.Time
		if eventLog.BlockTimestamp != 0 {
			ts = time.Unix(int64(eventLog.BlockTimestamp), 0)
		} else {
			ts = time.Now()
		}

		g.mu.Lock()
		g.pending[ek] = eventLog.BlockHash
		g.queue = append(g.queue, queueEntry{log: eventLog, arrivedAt: ts})
		g.mu.Unlock()

		// Non-blocking kick so the poller wakes up to (re)compute the timer
		// even when it is currently sleeping on a far-future deadline.
		select {
		case g.kick <- struct{}{}:
		default:
		}
		return nil
	}

	// eventLog.Removed == true: attempt pre-gate or post-gate cancellation.
	fk := forwardedKey{txHash: eventLog.TxHash, blockHash: eventLog.BlockHash, logIndex: uint(eventLog.Index)}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Pre-gate cancel: the live pending entry corresponds to this block.
	// Delete from pending; the tombstoned queue entry is skipped on pop.
	if liveBlockHash, ok := g.pending[ek]; ok && liveBlockHash == eventLog.BlockHash {
		delete(g.pending, ek)
		return nil
	}

	// Post-gate: the event has already been forwarded.
	if _, ok := g.forwardedSet[fk]; ok {
		g.logger.Warn("post-gate reorg detected",
			"txHash", eventLog.TxHash.Hex(),
			"blockHash", eventLog.BlockHash.Hex(),
			"logIndex", eventLog.Index,
			"chainID", g.chainID,
		)
		// Delete from the membership map; leave the forwardedQueue entry in
		// place — it expires on its own. The eviction loop's value-check makes
		// the later delete safe even if the same key is forwarded again.
		delete(g.forwardedSet, fk)
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

// run is the background goroutine that wakes on a kick, on the timer firing, or on
// ctx cancellation. It forwards matured entries, evicts stale forwardedSet entries,
// and reschedules the timer for the next head deadline.
func (g *ConfirmationGate) run(ctx context.Context) {
	defer g.timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-g.done:
			return
		case <-g.kick:
		case <-g.timer.C:
		}
		g.drainAndReschedule()
	}
}

// drainAndReschedule forwards all queue entries whose confirmation delay has
// elapsed, evicts forwardedSet entries older than (recentMultiplier × delay),
// and resets the timer to the next head deadline.
func (g *ConfirmationGate) drainAndReschedule() {
	g.mu.Lock()
	now := time.Now()

	// Step 1: drain matured head entries.
	for len(g.queue) > 0 && !g.queue[0].arrivedAt.Add(g.delay).After(now) {
		entry := g.queue[0]
		g.queue = g.queue[1:]

		ek := eventKey{txHash: entry.log.TxHash, logIndex: uint(entry.log.Index)}

		// Tombstone check: if the live pending entry no longer points at this
		// blockHash, a reorg-replacement event has superseded it. Drop silently.
		// Do NOT touch pending[ek] — it refers to the new live event (still in
		// the queue) and deleting it would break the next tombstone check or the
		// next Removed cancel.
		liveBlockHash, ok := g.pending[ek]
		if !ok || liveBlockHash != entry.log.BlockHash {
			continue
		}

		// Forward: clear pending, insert into forwardedSet + forwardedQueue
		// BEFORE releasing mu so that a fast Removed:true arriving immediately
		// after the handler call still sees the entry and emits the post-gate WARN.
		delete(g.pending, ek)
		fk := forwardedKey{
			txHash:    entry.log.TxHash,
			blockHash: entry.log.BlockHash,
			logIndex:  uint(entry.log.Index),
		}
		g.forwardedSet[fk] = now
		g.forwardedQueue = append(g.forwardedQueue, forwardedExpiry{key: fk, forwardedAt: now})

		g.mu.Unlock()

		evCtx := log.SetContextLogger(context.Background(), g.logger)
		if err := g.handler(evCtx, entry.log); err != nil {
			g.logger.Error("handler error after confirmation delay, signalling fatal",
				"error", err,
				"chainID", g.chainID,
			)
			select {
			case g.fatalCh <- err:
			default:
			}
			// Close done to signal the run goroutine to exit immediately.
			// This is safe: only this fatal branch closes done, and it runs at most
			// once — once done is closed the run loop exits and drainAndReschedule
			// is no longer called.
			close(g.done)
			return
		}

		g.mu.Lock()
	}

	// Step 2: FIFO eviction of forwardedSet entries older than recentMultiplier × delay.
	for len(g.forwardedQueue) > 0 && now.Sub(g.forwardedQueue[0].forwardedAt) > recentMultiplier*g.delay {
		popped := g.forwardedQueue[0]
		g.forwardedQueue = g.forwardedQueue[1:]

		// Only delete from forwardedSet if the stored timestamp still equals
		// the popped entry's timestamp. This guards the rare re-forward case
		// (same key forwarded again after a chain un-reorg) so the older FIFO
		// entry does not evict newer set membership. Tolerates the §2.4 Removed
		// path having already deleted the entry (no-op).
		if storedAt, ok := g.forwardedSet[popped.key]; ok && storedAt.Equal(popped.forwardedAt) {
			delete(g.forwardedSet, popped.key)
		}
	}

	// Step 3: reset timer to next head deadline using the standard drain pattern.
	if !g.timer.Stop() {
		select {
		case <-g.timer.C:
		default:
		}
	}
	if len(g.queue) > 0 {
		g.timer.Reset(time.Until(g.queue[0].arrivedAt.Add(g.delay)))
	}
	// else: leave the timer stopped; the next kick recomputes.

	g.mu.Unlock()
}
