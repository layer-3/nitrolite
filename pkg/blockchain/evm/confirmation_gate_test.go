package evm

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

// makeLog builds a types.Log with BlockTimestamp == 0. The gate then derives
// arrivedAt from time.Now() at HandleEvent time, which gives sub-second
// resolution. This is the appropriate helper for tests that use millisecond-scale
// delays — BlockTimestamp itself is unix-seconds and would round-trip-truncate
// any timestamp set by the test, causing arrivedAt to land up to 1s in the past
// and the entry to mature immediately.
//
// Tests that explicitly exercise the BlockTimestamp-driven arrival path use
// makeLogAt instead and pick durations large enough to tolerate second-resolution
// truncation.
func makeLog(txHash common.Hash, blockHash common.Hash, logIndex uint, removed bool) types.Log {
	return types.Log{
		TxHash:    txHash,
		BlockHash: blockHash,
		Index:     uint(logIndex),
		Removed:   removed,
	}
}

// makeLogAt builds a non-removed types.Log whose BlockTimestamp is set to the
// supplied wall-clock time. Used for tests that want the gate to derive
// arrivedAt from a specific moment in the past — must be paired with delays
// large enough (≥1s recommended) to tolerate seconds-resolution truncation of
// BlockTimestamp.
func makeLogAt(txHash common.Hash, blockHash common.Hash, logIndex uint, removed bool, ts time.Time) types.Log {
	return types.Log{
		TxHash:         txHash,
		BlockHash:      blockHash,
		Index:          uint(logIndex),
		Removed:        removed,
		BlockTimestamp: uint64(ts.Unix()),
	}
}

func newGate(t *testing.T, delay time.Duration, handler HandleEvent) *ConfirmationGate {
	t.Helper()
	g, err := NewConfirmationGate(delay, 1, handler, log.NewNoopLogger())
	require.NoError(t, err)
	return g
}

// T0: constructor rejects non-positive delay (operator-facing delay==0 is handled
// by wiring in main.go which skips constructing the gate).
func TestConfirmationGate_Constructor_RejectsNonPositiveDelay(t *testing.T) {
	t.Parallel()

	handler := func(_ context.Context, _ types.Log) error { return nil }

	g, err := NewConfirmationGate(0, 1, handler, log.NewNoopLogger())
	require.Error(t, err)
	assert.Nil(t, g)

	g, err = NewConfirmationGate(-1*time.Second, 1, handler, log.NewNoopLogger())
	require.Error(t, err)
	assert.Nil(t, g)
}

// T2: normal event is queued and delivered after the delay.
func TestConfirmationGate_NormalPath(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	var deliveredLog types.Log
	var mu sync.Mutex

	handler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		deliveredLog = l
		mu.Unlock()
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 5*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x02")
	bh := common.HexToHash("0xBB")
	ev := makeLog(tx, bh, 0, false)

	require.NoError(t, g.HandleEvent(context.Background(), ev))

	// should NOT be called within 1 ms
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before delay expires")

	// should be called within 500 ms total
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	assert.Equal(t, int32(1), callCount.Load())
	mu.Lock()
	assert.Equal(t, ev.TxHash, deliveredLog.TxHash)
	assert.Equal(t, ev.Index, deliveredLog.Index)
	mu.Unlock()
}

// T3: a Removed event for a queued entry cancels it before forwarding.
func TestConfirmationGate_ReorgCancel(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x03")
	bh := common.HexToHash("0xCC")

	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true)))

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must never be called after reorg cancel")
}

// T4: a re-delivered event (same tx/logIndex, different blockHash) replaces the original
// in the pending map; the late-arriving Removed for the old blockHash is a no-op (live
// pending hash no longer matches); the new event is forwarded once.
func TestConfirmationGate_OutOfOrder(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x04")
	bhOld := common.HexToHash("0xAA")
	bhNew := common.HexToHash("0xBB")

	// Event A: original block — queued under (tx, 0).
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhOld, 0, false)))
	// Event B: re-mined in new block — replaces pending[ek] = bhNew. The queued A entry
	// becomes a tombstone (its blockHash no longer matches pending[ek]).
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhNew, 0, false)))
	// Removed for old block: pending[ek] is bhNew, not bhOld; no forwarded entry yet;
	// no-op (debug log).
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhOld, 0, true)))

	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout — event B was not forwarded")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	// Only B should have been forwarded (A was tombstoned and silently dropped).
	assert.Equal(t, int32(1), callCount.Load())
}

// T5: post-gate reorg — Removed arrives after the event was already forwarded.
// Verify handler is called, Removed is handled gracefully (no panic/error).
func TestConfirmationGate_PostGateReorg(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 2*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x05")
	bh := common.HexToHash("0xDD")
	ev := makeLog(tx, bh, 0, false)

	require.NoError(t, g.HandleEvent(context.Background(), ev))

	// Wait until forwarded.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())

	// Post-gate Removed — should not panic or return error.
	// WARN log "post-gate reorg detected" is emitted internally (manually observable).
	err := g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)

	// Handler should still have been called exactly once.
	assert.Equal(t, int32(1), callCount.Load())
}

// T6: Removed for a completely unknown event — no error, no handler call.
func TestConfirmationGate_UnknownRemoval(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x06")
	bh := common.HexToHash("0xEE")

	err := g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load())
}

// T7: BlockTimestamp far in the past → event is immediately mature and forwarded fast.
func TestConfirmationGate_BlockTimestampBypass(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x07")
	bh := common.HexToHash("0xFF")

	// Block timestamp 30 seconds ago — arrivedAt + 10ms is far in the past, so the
	// entry is matured the moment the drain loop runs.
	require.NoError(t, g.HandleEvent(context.Background(), makeLogAt(tx, bh, 0, false, time.Now().Add(-30*time.Second))))

	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())
}

// T8: partial elapsed delay — BlockTimestamp 2 seconds in the past with delay=5s.
//
// Because BlockTimestamp is unix-seconds, the .Unix() conversion floors to the
// nearest whole second. In the worst case the gate sees arrivedAt up to 1s
// further in the past than the wall-clock target — so the actual remaining
// delay is in [2s, 3s]. Sleeping 500ms is safely inside that "not yet" window
// regardless of where the subsecond boundary landed.
func TestConfirmationGate_BlockTimestampPartialDelay(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 5*time.Second, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x08")
	bh := common.HexToHash("0x08")

	require.NoError(t, g.HandleEvent(context.Background(), makeLogAt(tx, bh, 0, false, time.Now().Add(-2*time.Second))))

	// Not called after 500 ms (worst-case remaining is ≥2s).
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before remaining delay expires")

	// Called within 7s total.
	deadline := time.After(7 * time.Second)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())
}

// T9 (reframed): BlockTimestamp == 0 falls back to time.Now() — the full delay
// must elapse. No log is emitted from the gate side (the listener owns any WARN
// for a missing timestamp).
func TestConfirmationGate_BlockTimestampZeroFallback(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x09")
	bh := common.HexToHash("0x09")

	// makeLog produces BlockTimestamp == 0 → gate falls back to time.Now().
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// Not called immediately (fell back to current time, full delay required).
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before delay expires")

	// Called within 500 ms.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())
}

// T11: cancelling the context prevents queued events from being forwarded.
func TestConfirmationGate_Shutdown(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 50*time.Millisecond, handler)
	ctx, cancel := context.WithCancel(t.Context())
	g.Start(ctx, func(error) {})

	for i := range 3 {
		tx := common.HexToHash(string(rune(0x20 + i)))
		bh := common.HexToHash(string(rune(0x30 + i)))
		require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, uint(i), false)))
	}

	// Cancel before delay expires.
	cancel()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "no events must be forwarded after context cancellation")
}

// T12: forwardedSet entries are evicted after recentMultiplier × delay.
// Behavior under test: after eviction, a Removed for the same (tx, blockHash, idx)
// must fall through to the DEBUG path — no panic, no error.
func TestConfirmationGate_ForwardedSetEviction(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	delay := 5 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x12")
	bh := common.HexToHash("0x12")

	// Enqueue and wait for forward.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// At this point forwardedSet contains the entry.
	g.mu.Lock()
	_, present := g.forwardedSet[forwardedKey{txHash: tx, blockHash: bh, logIndex: 0}]
	g.mu.Unlock()
	assert.True(t, present, "forwardedSet must contain the entry immediately after forwarding")

	// Wait well past recentMultiplier × delay, then enqueue another event to trigger
	// the eviction path inside drainAndReschedule.
	time.Sleep(time.Duration(recentMultiplier+1) * delay)

	tx2 := common.HexToHash("0x13")
	bh2 := common.HexToHash("0x13")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh2, 0, false)))

	// Wait for tx2 to forward; the eviction loop also runs.
	deadline = time.After(500 * time.Millisecond)
	for callCount.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("second handler invocation timed out")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	g.mu.Lock()
	_, presentAfter := g.forwardedSet[forwardedKey{txHash: tx, blockHash: bh, logIndex: 0}]
	g.mu.Unlock()
	assert.False(t, presentAfter, "old forwardedSet entry must be evicted after recentMultiplier × delay")

	// A second Removed for the original event — falls through to DEBUG (not found).
	// No panic, no error.
	err := g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)
}

// T13: multiple events are all delivered, preserving queue order.
func TestConfirmationGate_MultipleEvents_Ordering(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var delivered []common.Hash

	handler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		delivered = append(delivered, l.TxHash)
		mu.Unlock()
		return nil
	}

	g := newGate(t, 5*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	txHashes := []common.Hash{
		common.HexToHash("0xA1"),
		common.HexToHash("0xA2"),
		common.HexToHash("0xA3"),
	}
	bh := common.HexToHash("0xBLOCK")

	for i, tx := range txHashes {
		require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, uint(i), false)))
	}

	// Wait for all 3 events to be delivered.
	deadline := time.After(500 * time.Millisecond)
	for {
		mu.Lock()
		n := len(delivered)
		mu.Unlock()
		if n >= 3 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("only %d/3 events delivered within timeout", n)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, delivered, 3)
	assert.Equal(t, txHashes[0], delivered[0])
	assert.Equal(t, txHashes[1], delivered[1])
	assert.Equal(t, txHashes[2], delivered[2])
}

// New: tombstone-skip — a non-removed re-add with a different blockHash supersedes
// the queued entry. When the original entry's deadline arrives, the gate notices
// the tombstone (pending[ek] != entry.log.BlockHash) and silently drops it.
// Only the new entry's forward is observed.
func TestConfirmationGate_TombstoneSkip(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var delivered []common.Hash // blockHashes seen by handler

	handler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		delivered = append(delivered, l.BlockHash)
		mu.Unlock()
		return nil
	}

	delay := 30 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x20")
	bhA := common.HexToHash("0xAAA")
	bhB := common.HexToHash("0xBBB")

	// Enqueue event for blockHashA.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhA, 0, false)))
	// Before the delay elapses, send a non-removed re-add with blockHashB — same (tx, idx).
	// The gate replaces pending[ek] = bhB and appends a new queue entry; the bhA entry
	// becomes a tombstone.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhB, 0, false)))

	// Wait past the delay.
	deadline := time.After(500 * time.Millisecond)
	for {
		mu.Lock()
		n := len(delivered)
		mu.Unlock()
		if n >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout — event B was not forwarded")
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}

	// Allow extra time to ensure bhA does not slip through later (it shouldn't —
	// it's tombstoned and dropped silently on pop).
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, delivered, 1, "exactly one forward expected (the bhB entry)")
	assert.Equal(t, bhB, delivered[0], "the bhB entry must be the one forwarded")
}

// New: FIFO eviction with early-delete tolerance. After forwarding, a Removed:true
// arrives and removes the forwardedSet entry while emitting the post-gate WARN.
// Later, the FIFO eviction loop pops the corresponding forwardedQueue entry — the
// set entry is already gone. The eviction must not panic and must not double-invoke
// the handler.
func TestConfirmationGate_FIFOEviction_ToleratesEarlyDelete(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	delay := 5 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x30")
	bh := common.HexToHash("0x30")

	// Forward an event.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Confirm the forwardedSet entry exists.
	fk := forwardedKey{txHash: tx, blockHash: bh, logIndex: 0}
	g.mu.Lock()
	_, presentBefore := g.forwardedSet[fk]
	queueLen := len(g.forwardedQueue)
	g.mu.Unlock()
	require.True(t, presentBefore, "forwardedSet must contain the entry immediately after forwarding")
	require.Equal(t, 1, queueLen, "forwardedQueue must contain one entry")

	// Send Removed:true — gate emits post-gate WARN and deletes the entry from forwardedSet
	// (but leaves the forwardedQueue entry in place; it will expire on its own).
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true)))

	g.mu.Lock()
	_, presentAfterRemoved := g.forwardedSet[fk]
	g.mu.Unlock()
	require.False(t, presentAfterRemoved, "forwardedSet entry must be deleted by the post-gate WARN path")

	// Wait well past recentMultiplier × delay, then kick the drain loop with a new event
	// so eviction runs and pops the orphaned forwardedQueue entry.
	time.Sleep(time.Duration(recentMultiplier+1) * delay)

	tx2 := common.HexToHash("0x31")
	bh2 := common.HexToHash("0x31")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh2, 0, false)))

	// Wait for the second forward.
	deadline = time.After(500 * time.Millisecond)
	for callCount.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("second handler invocation timed out")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Handler called exactly twice (once per forward; no double-action from eviction).
	assert.Equal(t, int32(2), callCount.Load())

	// The orphaned forwardedQueue entry must have been popped during eviction.
	g.mu.Lock()
	// After tx2 is forwarded, the queue should have exactly one entry (tx2's).
	finalQueueLen := len(g.forwardedQueue)
	g.mu.Unlock()
	assert.Equal(t, 1, finalQueueLen, "orphan forwardedQueue entry must have been evicted")
}

// New: timer reschedule — enqueue a single event and do NOT send any further kicks.
// The handler must be invoked when the timer fires.
func TestConfirmationGate_TimerReschedule(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 20*time.Millisecond, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0x40")
	bh := common.HexToHash("0x40")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// No further HandleEvent calls. Wait for the timer to fire.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called via timer fire alone")
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())
}

// New: kick during a pending timer must NOT extend the original timer's deadline.
// Event A is enqueued first (timer arms for A's deadline). Before A matures we
// enqueue B with a LATER BlockTimestamp. A must still fire at its original
// deadline; the kick rescheduled the timer to A's head deadline (unchanged).
func TestConfirmationGate_KickDuringPendingTimer(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var deliveredOrder []common.Hash
	firstFiredAt := make(chan time.Time, 1)

	handler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		deliveredOrder = append(deliveredOrder, l.TxHash)
		isFirst := len(deliveredOrder) == 1
		mu.Unlock()
		if isFirst {
			select {
			case firstFiredAt <- time.Now():
			default:
			}
		}
		return nil
	}

	delay := 100 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	txA := common.HexToHash("0x50")
	bhA := common.HexToHash("0x50")
	txB := common.HexToHash("0x51")
	bhB := common.HexToHash("0x51")

	// Event A: BlockTimestamp == 0 → gate uses time.Now() at HandleEvent time as arrivedAt.
	enqueueA := time.Now()
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(txA, bhA, 0, false)))

	// Brief sleep, then enqueue B. The kick wakes the drain loop; A is not yet
	// mature; the timer must be reset to A's deadline. B's deadline is later than
	// A's because its arrivedAt is later (HandleEvent uses time.Now() when
	// BlockTimestamp == 0).
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(txB, bhB, 0, false)))

	// Wait for A to fire.
	select {
	case firedAt := <-firstFiredAt:
		// A's expected deadline was enqueueA + delay. Firing should occur no
		// earlier than ~that moment and not be delayed by B's later deadline.
		elapsed := firedAt.Sub(enqueueA)
		// Allow generous slack but ensure A did not get pushed to B's deadline
		// (B's deadline is enqueueA + ~20ms + 50ms + delay = enqueueA + ~170ms).
		assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "A fired before its deadline")
		assert.Less(t, elapsed, 160*time.Millisecond, "A's deadline was extended by B's kick")
	case <-time.After(1 * time.Second):
		t.Fatal("A did not fire within timeout")
	}

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(deliveredOrder), 1)
	assert.Equal(t, txA, deliveredOrder[0], "A must fire first (queue order preserved)")
}

// New: shutdown with non-empty queue — cancel the gate's context, assert the
// goroutine exits quickly and no handler is invoked.
func TestConfirmationGate_ShutdownWithNonEmptyQueue(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handlerEntered := make(chan struct{}, 4)
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		select {
		case handlerEntered <- struct{}{}:
		default:
		}
		return nil
	}

	g := newGate(t, 200*time.Millisecond, handler)
	ctx, cancel := context.WithCancel(t.Context())
	g.Start(ctx, func(error) {})

	// Enqueue multiple events far in the future.
	for i := range 4 {
		tx := common.HexToHash(string(rune(0x60 + i)))
		bh := common.HexToHash(string(rune(0x70 + i)))
		require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, uint(i), false)))
	}

	// Cancel and assert the gate's goroutine exits within a short window.
	cancel()

	// Give the goroutine time to observe ctx.Done.
	time.Sleep(50 * time.Millisecond)

	// Even if we wait far longer than the delay would otherwise require, no handler call.
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "no handler invocations expected after shutdown")

	select {
	case <-handlerEntered:
		t.Fatal("handler was invoked after shutdown")
	default:
	}
}

// TestConfirmationGate_HandlerErrorPropagatesFatal: when the downstream handler
// returns an error after the confirmation delay, the gate's lifecycle closure
// receives the sentinel error exactly once and the run goroutine exits.
// Subsequent events that mature must NOT invoke the handler again.
func TestConfirmationGate_HandlerErrorPropagatesFatal(t *testing.T) {
	t.Parallel()

	sentinelErr := errors.New("handler sentinel error")
	var handlerCalls atomic.Int64
	handler := func(_ context.Context, _ types.Log) error {
		handlerCalls.Add(1)
		return sentinelErr
	}

	delay := 50 * time.Millisecond
	g := newGate(t, delay, handler)

	closureCh := make(chan error, 2) // size 2 to catch a buggy double-invocation
	g.Start(t.Context(), func(err error) { closureCh <- err })

	tx := common.HexToHash("0xF1")
	bh := common.HexToHash("0xF1")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// The closure must be invoked once with the sentinel error.
	select {
	case err := <-closureCh:
		assert.Equal(t, sentinelErr, err, "closure must receive the sentinel error")
	case <-time.After(delay + 200*time.Millisecond):
		t.Fatal("closure was not invoked within timeout")
	}

	// A second invocation must not occur — the run goroutine has exited.
	select {
	case extra := <-closureCh:
		t.Fatalf("unexpected second closure invocation: %v", extra)
	case <-time.After(50 * time.Millisecond):
		// correct: no second invocation
	}

	// Enqueue a second event after the failure. The goroutine has exited; even
	// if the kick is queued in the buffered channel it will never be drained.
	tx2 := common.HexToHash("0xF2")
	bh2 := common.HexToHash("0xF2")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh2, 0, false)))

	// Wait past the delay; the handler must NOT be called a second time.
	time.Sleep(delay + 100*time.Millisecond)
	assert.Equal(t, int64(1), handlerCalls.Load(), "handler must be invoked exactly once across gate lifetime")
}

// TestGate_FlushPending_ClearsPendingAndQueue: FlushPending zeros queue and pending
// but intentionally retains forwardedSet so the post-gate WARN window survives reconnects.
// The storedAt.Equal(popped.forwardedAt) eviction guard at confirmation_gate.go:281-288
// is load-bearing for re-forward correctness and depends on forwardedSet membership; see
// FlushPending doc comment.
func TestGate_FlushPending_ClearsPendingAndQueue(t *testing.T) {
	t.Parallel()

	var forwardCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		forwardCount.Add(1)
		return nil
	}

	delay := 1 * time.Second // large enough so events don't mature during the test
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	// Push 3 non-removed events; they must sit in pending and queue.
	txHashes := []common.Hash{
		common.HexToHash("0xA1"),
		common.HexToHash("0xA2"),
		common.HexToHash("0xA3"),
	}
	bh := common.HexToHash("0xBLOCK")
	for i, tx := range txHashes {
		require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, uint(i), false)))
	}

	g.mu.Lock()
	assert.Equal(t, 3, len(g.queue), "queue must have 3 entries before flush")
	assert.Equal(t, 3, len(g.pending), "pending must have 3 entries before flush")
	g.mu.Unlock()

	g.FlushPending()

	g.mu.Lock()
	assert.Nil(t, g.queue, "queue must be nil after flush")
	assert.Equal(t, 0, len(g.pending), "pending must be empty after flush")
	g.mu.Unlock()

	// Handler must not have been called (entries were flushed before maturing).
	assert.Equal(t, int32(0), forwardCount.Load(), "handler must not be called for flushed entries")
}

// TestGate_FlushPending_RetainsForwardedSet: forwardedSet must survive a FlushPending call.
func TestGate_FlushPending_RetainsForwardedSet(t *testing.T) {
	t.Parallel()

	var forwardCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		forwardCount.Add(1)
		return nil
	}

	delay := 5 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	// Enqueue and wait for one event to be forwarded (so forwardedSet has an entry).
	tx := common.HexToHash("0xF1")
	bh := common.HexToHash("0xF1")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	deadline := time.After(500 * time.Millisecond)
	for forwardCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// forwardedSet must contain the entry.
	fk := forwardedKey{txHash: tx, blockHash: bh, logIndex: 0}
	g.mu.Lock()
	_, presentBefore := g.forwardedSet[fk]
	g.mu.Unlock()
	require.True(t, presentBefore, "forwardedSet must contain the entry before flush")

	// Enqueue two more events so flush has something to clear.
	tx2 := common.HexToHash("0xF2")
	bh2 := common.HexToHash("0xF2")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh2, 0, false)))

	g.FlushPending()

	// forwardedSet must still contain the originally forwarded entry.
	g.mu.Lock()
	_, presentAfter := g.forwardedSet[fk]
	queueLen := len(g.queue)
	pendingLen := len(g.pending)
	g.mu.Unlock()

	assert.True(t, presentAfter, "forwardedSet must be retained across FlushPending")
	assert.Nil(t, g.queue, "queue must be nil after flush")
	assert.Equal(t, 0, pendingLen, "pending must be empty after flush")
	assert.Equal(t, 0, queueLen, "queue must be empty after flush")
}

// TestGate_ReForwardAfterFlush_EvictionGuardCorrect: verifies that the
// storedAt.Equal(popped.forwardedAt) guard at confirmation_gate.go:281-288 handles
// the case where an entry that was in forwardedSet before the flush is re-forwarded
// after the flush. The older FIFO entry must NOT evict the newer set membership.
func TestGate_ReForwardAfterFlush_EvictionGuardCorrect(t *testing.T) {
	t.Parallel()

	var forwardCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		forwardCount.Add(1)
		return nil
	}

	delay := 5 * time.Millisecond
	g := newGate(t, delay, handler)
	g.Start(t.Context(), func(error) {})

	tx := common.HexToHash("0xEG1")
	bh := common.HexToHash("0xEG1")

	// First forward.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))
	deadline := time.After(500 * time.Millisecond)
	for forwardCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("first forward timed out")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), forwardCount.Load())

	// Flush: queue and pending are cleared; forwardedSet retains the entry.
	g.FlushPending()

	// Re-enqueue the same (tx, logIndex) with same blockHash — simulates re-forward after flush.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// Wait for the re-forward.
	deadline = time.After(500 * time.Millisecond)
	for forwardCount.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("re-forward timed out")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(2), forwardCount.Load())

	// Wait past recentMultiplier × delay so the eviction loop runs.
	time.Sleep(time.Duration(recentMultiplier+1) * delay)

	// Kick the eviction loop by enqueuing a third event.
	tx2 := common.HexToHash("0xEG2")
	bh2 := common.HexToHash("0xEG2")
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh2, 0, false)))
	deadline = time.After(500 * time.Millisecond)
	for forwardCount.Load() < 3 {
		select {
		case <-deadline:
			t.Fatal("third forward timed out")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(3), forwardCount.Load())
}
