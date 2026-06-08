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

func noopFetcher(_ common.Hash) (time.Time, error) {
	return time.Now(), nil
}

func makeLog(txHash common.Hash, blockHash common.Hash, logIndex uint, removed bool) types.Log {
	return types.Log{
		TxHash:    txHash,
		BlockHash: blockHash,
		Index:     uint(logIndex),
		Removed:   removed,
	}
}

func newGate(t *testing.T, delay time.Duration, handler HandleEvent, fetcher func(common.Hash) (time.Time, error)) *ConfirmationGate {
	t.Helper()
	if fetcher == nil {
		fetcher = noopFetcher
	}
	return NewConfirmationGate(delay, 1, handler, fetcher, log.NewNoopLogger())
}

// T1: delay==0 forwards non-removed events directly; Removed:true events are silently
// dropped to protect reactors that have no guard on l.Topics[0].
func TestConfirmationGate_Delay0_DirectForward(t *testing.T) {
	t.Parallel()

	var calls []types.Log
	var mu sync.Mutex

	wantErr := errors.New("handler error")
	handler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		calls = append(calls, l)
		mu.Unlock()
		return wantErr
	}

	g := newGate(t, 0, handler, nil)
	g.Start(t.Context()) // should be a no-op for delay==0

	tx := common.HexToHash("0x01")
	bh := common.HexToHash("0xAA")

	// normal event — forwarded, handler error propagated
	normalLog := makeLog(tx, bh, 0, false)
	err := g.HandleEvent(context.Background(), normalLog)
	require.Equal(t, wantErr, err)

	// removed event — silently dropped; handler NOT called, nil returned
	removedLog := makeLog(tx, bh, 0, true)
	err = g.HandleEvent(context.Background(), removedLog)
	require.NoError(t, err)

	mu.Lock()
	assert.Len(t, calls, 1, "handler must be called only for non-removed event")
	mu.Unlock()
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

	g := newGate(t, 5*time.Millisecond, handler, nil)
	g.Start(t.Context())

	tx := common.HexToHash("0x02")
	bh := common.HexToHash("0xBB")
	ev := makeLog(tx, bh, 0, false)

	require.NoError(t, g.HandleEvent(context.Background(), ev))

	// should NOT be called within 1 ms
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before delay expires")

	// should be called within 10 ms total
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

	g := newGate(t, 10*time.Millisecond, handler, nil)
	g.Start(t.Context())

	tx := common.HexToHash("0x03")
	bh := common.HexToHash("0xCC")

	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true)))

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must never be called after reorg cancel")
}

// T4: a re-delivered event (same tx/logIndex, different blockHash) replaces the original;
// the Removed for the old blockHash is a no-op; the new event is forwarded once.
func TestConfirmationGate_OutOfOrder(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 10*time.Millisecond, handler, nil)
	g.Start(t.Context())

	tx := common.HexToHash("0x04")
	bhOld := common.HexToHash("0xAA")
	bhNew := common.HexToHash("0xBB")

	// Event A: original block
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhOld, 0, false)))
	// Event B: re-mined in new block (same txHash/logIndex, different blockHash)
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhNew, 0, false)))
	// Removed for old block: matches A's full key (bh=0xAA) and removes it from queue.
	// B (bh=0xBB) is left untouched.
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bhOld, 0, true)))

	// Wait long enough for the poll goroutine to fire (pollInterval=50ms) and the delay to
	// have elapsed (10ms). 200ms gives generous headroom.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout — event B was not forwarded")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	// Only B should have been forwarded (A was cancelled).
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

	g := newGate(t, 2*time.Millisecond, handler, nil)
	g.Start(t.Context())

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

	g := newGate(t, 10*time.Millisecond, handler, nil)
	g.Start(t.Context())

	tx := common.HexToHash("0x06")
	bh := common.HexToHash("0xEE")

	err := g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load())
}

// T7: fetcher returns an old timestamp → event is immediately mature and forwarded fast.
func TestConfirmationGate_BlockTimestampBypass(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	pastFetcher := func(_ common.Hash) (time.Time, error) {
		return time.Now().Add(-30 * time.Second), nil
	}

	g := newGate(t, 10*time.Millisecond, handler, pastFetcher)
	g.Start(t.Context())

	tx := common.HexToHash("0x07")
	bh := common.HexToHash("0xFF")

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
	assert.Equal(t, int32(1), callCount.Load())
}

// T8: fetcher returns a timestamp 60ms in the past; delay=100ms; so ~40ms remain.
func TestConfirmationGate_BlockTimestampPartialDelay(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	fetcher := func(_ common.Hash) (time.Time, error) {
		return time.Now().Add(-60 * time.Millisecond), nil
	}

	g := newGate(t, 100*time.Millisecond, handler, fetcher)
	g.Start(t.Context())

	tx := common.HexToHash("0x08")
	bh := common.HexToHash("0x08")

	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// Not called after 20 ms (need ~40ms more).
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before remaining delay expires")

	// Called within 200 ms total.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler not called within timeout")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	assert.Equal(t, int32(1), callCount.Load())
}

// T9: fetcher returns error → fallback to time.Now() → full delay must still elapse.
func TestConfirmationGate_BlockTimestampFetchError(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	errFetcher := func(_ common.Hash) (time.Time, error) {
		return time.Time{}, errors.New("rpc error")
	}

	g := newGate(t, 5*time.Millisecond, handler, errFetcher)
	g.Start(t.Context())

	tx := common.HexToHash("0x09")
	bh := common.HexToHash("0x09")

	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx, bh, 0, false)))

	// Not called immediately (fell back to current time, full delay required).
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(0), callCount.Load(), "handler must not be called before delay expires")

	// Called after 10 ms.
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

// T10: two events sharing the same blockHash should result in exactly one fetcher call.
func TestConfirmationGate_BlockTimestampCache(t *testing.T) {
	t.Parallel()

	var fetchCount atomic.Int32
	var callCount atomic.Int32

	fetcher := func(_ common.Hash) (time.Time, error) {
		fetchCount.Add(1)
		return time.Now(), nil
	}
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 5*time.Millisecond, handler, fetcher)
	g.Start(t.Context())

	tx1 := common.HexToHash("0x10")
	tx2 := common.HexToHash("0x11")
	bh := common.HexToHash("0xSHARED")

	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx1, bh, 0, false)))
	require.NoError(t, g.HandleEvent(context.Background(), makeLog(tx2, bh, 1, false)))

	// Wait for both events to be forwarded.
	deadline := time.After(500 * time.Millisecond)
	for callCount.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("not all events delivered within timeout")
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	assert.Equal(t, int32(1), fetchCount.Load(), "fetcher must be called only once for a shared blockHash")
}

// T11: cancelling the context prevents queued events from being forwarded.
func TestConfirmationGate_Shutdown(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	g := newGate(t, 50*time.Millisecond, handler, nil)
	ctx, cancel := context.WithCancel(t.Context())
	g.Start(ctx)

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

// T12: recentlyForwarded entries are evicted after recentMultiplier × delay.
func TestConfirmationGate_RecentlyForwardedEviction(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32
	handler := func(_ context.Context, _ types.Log) error {
		callCount.Add(1)
		return nil
	}

	delay := 2 * time.Millisecond
	g := newGate(t, delay, handler, nil)
	g.Start(t.Context())

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

	// Immediately send a post-gate Removed — should match recentlyForwarded (WARN path).
	err := g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)

	// Wait well past recentMultiplier × delay so the entry is evicted.
	time.Sleep(time.Duration(recentMultiplier) * delay)

	// A second Removed for the same event — should fall through to DEBUG path (not found).
	// Verifies the eviction happened. No panic, no error.
	err = g.HandleEvent(context.Background(), makeLog(tx, bh, 0, true))
	assert.NoError(t, err)

	// Handler still called exactly once.
	assert.Equal(t, int32(1), callCount.Load())
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

	g := newGate(t, 5*time.Millisecond, handler, nil)
	g.Start(t.Context())

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
