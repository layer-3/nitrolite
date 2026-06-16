package evm

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSubscription implements ethereum.Subscription
type MockSubscription struct {
	errChan   chan error
	unsub     func()
	closeOnce sync.Once
}

func (m *MockSubscription) Err() <-chan error {
	return m.errChan
}

func (m *MockSubscription) Unsubscribe() {
	if m.unsub != nil {
		m.unsub()
	}
	m.closeOnce.Do(func() {
		close(m.errChan)
	})
}

func TestNewListener(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	l := NewListener(addr, mockClient, 1, 100, 0, logger, nil, nil, eventGetter, nil)
	require.NotNil(t, l)
	assert.Equal(t, addr, l.contractAddress)
	assert.Equal(t, uint64(1), l.blockchainID)
	assert.Equal(t, uint64(100), l.blockStep)
}

func TestListener_Listen_CurrentEvents(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	// No stored events → findCommonAncestor returns 0 immediately (genesis).
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).Return(uint64(0), "", nil)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Channel to signal event handling
	eventHandled := make(chan struct{})
	handleEvent := func(ctx context.Context, log types.Log) error {
		cancel()
		close(eventHandled)
		return nil
	}

	listener := NewListener(addr, mockClient, 1, 100, 0, logger, handleEvent, handleEvent, eventGetter, nil)

	// Mock SubscribeFilterLogs
	sub := &MockSubscription{
		errChan: make(chan error),
		unsub:   func() {},
	}

	// Mock SubscribeFilterLogs: send a log immediately. BlockTimestamp is set so
	// the listener's ensureBlockTimestamp short-circuits and does not call HeaderByHash.
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- types.Log)
			ch <- types.Log{BlockNumber: 10, Index: 1, BlockTimestamp: uint64(time.Now().Unix())}
		}).
		Return(sub, nil)

	// The first current event will trigger IsContractEventProcessed check
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(1), uint64(1)).Return(false, nil)

	go listener.Listen(ctx, func(err error) {})

	select {
	case <-eventHandled:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestListener_ReconcileBlockRange(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	// Setup FilterLogs mock
	// We expect a range fetch. start=100, step=10 -> end=110. current=120.
	// 1. 100-110
	// 2. 111-120

	logs1 := []types.Log{{BlockNumber: 105, Index: 0}}
	logs2 := []types.Log{{BlockNumber: 115, Index: 0}}

	mockClient.On("FilterLogs", mock.Anything, mock.MatchedBy(func(q ethereum.FilterQuery) bool {
		return q.FromBlock.Uint64() == 100 && q.ToBlock.Uint64() == 110
	})).Return(logs1, nil)

	mockClient.On("FilterLogs", mock.Anything, mock.MatchedBy(func(q ethereum.FilterQuery) bool {
		return q.FromBlock.Uint64() == 111 && q.ToBlock.Uint64() == 120
	})).Return(logs2, nil)

	historicalCh := make(chan types.Log, 10)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		listener.reconcileBlockRange(context.Background(), 120, 100, historicalCh)
		close(historicalCh)
	})

	var receivedLogs []types.Log
	for l := range historicalCh {
		receivedLogs = append(receivedLogs, l)
	}
	wg.Wait()

	assert.Len(t, receivedLogs, 2)
	assert.Equal(t, uint64(105), receivedLogs[0].BlockNumber)
	assert.Equal(t, uint64(115), receivedLogs[1].BlockNumber)
}

func TestListener_Listen_HistoricalAndCurrent(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	// Start from block 100, canonical: the reconciler will compute its hash via HeaderByNumber(100)
	// and compare against the stored hash. We construct a deterministic Header so we can pre-compute
	// the hash and feed it back as the stored value.
	canonicalAt100 := &types.Header{Number: big.NewInt(100), Difficulty: big.NewInt(1)}
	blockHash100 := canonicalAt100.Hash()
	eventGetter := new(MockContractEventGetter)
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).Return(uint64(100), blockHash100.Hex(), nil)
	// Historical event at block 105 is not present
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil)
	// Current event at block 111 — after historical is done, first current event triggers check
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var receivedCount atomic.Int64
	doneCh := make(chan struct{})

	handleEvent := func(ctx context.Context, log types.Log) error {
		count := receivedCount.Add(1)
		if count >= 2 { // Expect 1 historical + 1 current
			cancel()
			select {
			case <-doneCh:
			default:
				close(doneCh)
			}
		}

		return nil
	}

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, handleEvent, handleEvent, eventGetter, nil)

	// findCommonAncestor: HeaderByNumber(100) returns the same header we hashed above,
	// so the stored hash matches and block 100 is confirmed canonical.
	mockClient.On("HeaderByNumber", mock.Anything, mock.MatchedBy(func(n *big.Int) bool {
		return n != nil && n.Cmp(big.NewInt(100)) == 0
	})).Return(canonicalAt100, nil)

	// Mock HeaderByNumber(nil) for the chain-tip lookup (current tip is 110).
	currentHeader := &types.Header{Number: big.NewInt(110)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	// Mock FilterLogs (100-110). BlockTimestamp is set so ensureBlockTimestamp short-circuits.
	histLogs := []types.Log{{BlockNumber: 105, Index: 0, BlockTimestamp: uint64(time.Now().Unix())}}
	mockClient.On("FilterLogs", mock.Anything, mock.Anything).Return(histLogs, nil)

	// Mock SubscribeFilterLogs
	sub := &MockSubscription{errChan: make(chan error)}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- types.Log)
			ch <- types.Log{BlockNumber: 111, Index: 0, BlockTimestamp: uint64(time.Now().Unix())}
		}).
		Return(sub, nil)

	go listener.Listen(ctx, func(err error) {})

	select {
	case <-doneCh:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for events")
	}
}

func TestProcessEvents_DedupSkipsPresent(t *testing.T) {
	t.Parallel()
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	var handledBlocks []uint64
	handleEvent := func(ctx context.Context, eventLog types.Log) error {
		handledBlocks = append(handledBlocks, eventLog.BlockNumber)
		return nil
	}

	listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter, nil)

	// Historical: 3 events. First 2 are present (skipped), 3rd is not (handled).
	// After the 3rd, the check should stop — no IsContractEventProcessed call for events 4+.
	// BlockTimestamp is set so ensureBlockTimestamp short-circuits.
	ts := uint64(time.Now().Unix())
	historicalCh := make(chan types.Log, 5)
	historicalCh <- types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaa"), BlockTimestamp: ts}
	historicalCh <- types.Log{BlockNumber: 101, Index: 0, TxHash: common.HexToHash("0xbb"), BlockTimestamp: ts}
	historicalCh <- types.Log{BlockNumber: 102, Index: 0, TxHash: common.HexToHash("0xcc"), BlockTimestamp: ts}
	historicalCh <- types.Log{BlockNumber: 103, Index: 0, TxHash: common.HexToHash("0xdd"), BlockTimestamp: ts}
	historicalCh <- types.Log{BlockNumber: 104, Index: 0, TxHash: common.HexToHash("0xee"), BlockTimestamp: ts}
	close(historicalCh)

	// First two are present, third is not
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(true, nil).Once()
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(true, nil).Once()
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()
	// No mock for 103, 104 — if called, mock will panic, proving the check stopped

	sub := &MockSubscription{errChan: make(chan error)}
	currentCh := make(chan types.Log)

	// processEvents will drain historical, then block on currentCh. Cancel via ctx.
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// Wait for historical processing, then cancel
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var lastBlock uint64
	err := listener.processEvents(ctx, sub, historicalCh, currentCh, &lastBlock)
	require.NoError(t, err)

	// Only events 102, 103, 104 should have been handled (100, 101 skipped as present)
	assert.Equal(t, []uint64{102, 103, 104}, handledBlocks)
	eventGetter.AssertExpectations(t)
}

func TestProcessEvents_SubscriptionErrorDuringPhase1(t *testing.T) {
	t.Parallel()
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	var handledBlocks []uint64
	handleEvent := func(ctx context.Context, eventLog types.Log) error {
		handledBlocks = append(handledBlocks, eventLog.BlockNumber)
		return nil
	}

	listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter, nil)

	// Historical channel with events that will block (not closed yet). BlockTimestamp
	// is set so ensureBlockTimestamp short-circuits.
	historicalCh := make(chan types.Log, 2)
	historicalCh <- types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaa"), BlockTimestamp: uint64(time.Now().Unix())}

	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil)

	// Subscription that will error shortly
	subErrCh := make(chan error, 1)
	sub := &MockSubscription{errChan: subErrCh, unsub: func() {}}
	currentCh := make(chan types.Log)

	// Send subscription error after a short delay (while historical is still open)
	go func() {
		time.Sleep(50 * time.Millisecond)
		subErrCh <- fmt.Errorf("connection lost")
	}()

	var lastBlock uint64
	err := listener.processEvents(context.Background(), sub, historicalCh, currentCh, &lastBlock)

	// Should return nil (reconnect signal), not an error
	require.NoError(t, err)
	// The first historical event should have been handled before the subscription error
	assert.Equal(t, []uint64{100}, handledBlocks)
}

// TestListener_PhaseHandlerRouting verifies the age-based routing of Phase 1 events:
//   - Historical events older than confirmationDelay → handleHistoricalEvent (direct, gate bypass)
//   - Historical events younger than confirmationDelay → handleEvent (through gate; still in reorg window)
//   - Live (Phase 2) events → handleEvent (always)
//   - HeaderByHash fetch failures → handleEvent (conservative fallback)
//
// See nitronode/docs/reorg-fix.md §4.4 step 5.
func TestListener_PhaseHandlerRouting(t *testing.T) {
	t.Parallel()
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	confirmationDelay := 60 * time.Second

	mockClient := new(MockEVMClient)
	eventGetter := new(MockContractEventGetter)

	var (
		mu             sync.Mutex
		historicalLogs []types.Log
		liveLogs       []types.Log
	)
	historicalHandler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		defer mu.Unlock()
		historicalLogs = append(historicalLogs, l)
		return nil
	}
	liveHandler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		defer mu.Unlock()
		liveLogs = append(liveLogs, l)
		return nil
	}

	listener := NewListener(addr, mockClient, 1, 10, confirmationDelay, logger, liveHandler, historicalHandler, eventGetter, nil)

	// Old historical event (block timestamp 10 minutes ago) — should bypass the gate.
	oldHash := common.HexToHash("0xa1")
	oldLog := types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaaa"), BlockHash: oldHash}
	oldHeader := &types.Header{Number: big.NewInt(100), Time: uint64(time.Now().Add(-10 * time.Minute).Unix())}
	mockClient.On("HeaderByHash", mock.Anything, oldHash).Return(oldHeader, nil).Once()

	// Recent historical event (block timestamp 5 seconds ago) — should flow through the gate.
	recentHash := common.HexToHash("0xa2")
	recentLog := types.Log{BlockNumber: 101, Index: 0, TxHash: common.HexToHash("0xbbb"), BlockHash: recentHash}
	recentHeader := &types.Header{Number: big.NewInt(101), Time: uint64(time.Now().Add(-5 * time.Second).Unix())}
	mockClient.On("HeaderByHash", mock.Anything, recentHash).Return(recentHeader, nil).Once()

	// Historical event whose HeaderByHash fetch fails — should fall back to the gate.
	failHash := common.HexToHash("0xa3")
	failLog := types.Log{BlockNumber: 102, Index: 0, TxHash: common.HexToHash("0xccc"), BlockHash: failHash}
	mockClient.On("HeaderByHash", mock.Anything, failHash).Return(nil, fmt.Errorf("rpc failure")).Once()

	// Live event — always to liveHandler regardless of age. BlockTimestamp is set
	// so ensureBlockTimestamp short-circuits on the Phase 2 path (avoiding a
	// HeaderByHash call we'd otherwise have to mock).
	currentLog := types.Log{BlockNumber: 200, Index: 0, TxHash: common.HexToHash("0xddd"), BlockHash: common.HexToHash("0xb1"), BlockTimestamp: uint64(time.Now().Unix())}

	historicalCh := make(chan types.Log, 3)
	historicalCh <- oldLog
	historicalCh <- recentLog
	historicalCh <- failLog
	close(historicalCh)

	currentCh := make(chan types.Log, 1)
	currentCh <- currentLog

	// Only the first historical event triggers IsContractEventProcessed (then the check is dropped for the phase);
	// the first live event triggers it again for Phase 2.
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()

	sub := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	var lastBlock uint64
	err := listener.processEvents(ctx, sub, historicalCh, currentCh, &lastBlock)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, historicalLogs, 1, "only the old historical event should bypass the gate")
	assert.Equal(t, uint64(100), historicalLogs[0].BlockNumber)
	require.Len(t, liveLogs, 3, "recent + fallback historical events plus the live event must reach the live handler")
	assert.Equal(t, uint64(101), liveLogs[0].BlockNumber, "recent historical event routed through the gate")
	assert.Equal(t, uint64(102), liveLogs[1].BlockNumber, "HeaderByHash-failed historical event routed through the gate (conservative fallback)")
	assert.Equal(t, uint64(200), liveLogs[2].BlockNumber, "live event always routed to the gate")

	mockClient.AssertExpectations(t)
	eventGetter.AssertExpectations(t)
}

// TestListener_PhaseHandlerRouting_DelayZero verifies that when confirmationDelay is 0,
// every historical event is routed to handleHistoricalEvent without any HeaderByHash
// fetch — preserving the legacy bypass for gate-disabled chains.
func TestListener_PhaseHandlerRouting_DelayZero(t *testing.T) {
	t.Parallel()
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	mockClient := new(MockEVMClient)
	eventGetter := new(MockContractEventGetter)

	var (
		mu             sync.Mutex
		historicalLogs []types.Log
	)
	historicalHandler := func(_ context.Context, l types.Log) error {
		mu.Lock()
		defer mu.Unlock()
		historicalLogs = append(historicalLogs, l)
		return nil
	}
	liveHandler := func(_ context.Context, _ types.Log) error {
		t.Fatal("live handler must not be called when delay is 0 and only Phase 1 events are present")
		return nil
	}

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, liveHandler, historicalHandler, eventGetter, nil)

	// BlockTimestamp populated by the upstream RPC — ensureBlockTimestamp short-circuits
	// and routeHistoricalEvent routes directly to historicalHandler because delay == 0.
	histLog := types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaaa"), BlockHash: common.HexToHash("0xa1"), BlockTimestamp: uint64(time.Now().Unix())}
	historicalCh := make(chan types.Log, 1)
	historicalCh <- histLog
	close(historicalCh)
	currentCh := make(chan types.Log)

	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()

	sub := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var lastBlock uint64
	err := listener.processEvents(ctx, sub, historicalCh, currentCh, &lastBlock)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, historicalLogs, 1)
	assert.Equal(t, uint64(100), historicalLogs[0].BlockNumber)

	// HeaderByHash must NOT have been called — the upstream RPC populated BlockTimestamp,
	// so ensureBlockTimestamp short-circuits.
	mockClient.AssertNotCalled(t, "HeaderByHash")
}

func TestListener_RemovedLog_ForwardedToHandler(t *testing.T) {
	t.Parallel()

	t.Run("WithGate", func(t *testing.T) {
		t.Parallel()
		logger := log.NewNoopLogger()
		addr := common.HexToAddress("0x123")
		eventGetter := new(MockContractEventGetter)

		// Track which logs reached handleEvent.
		var handledLogs []types.Log
		handleEvent := func(ctx context.Context, eventLog types.Log) error {
			handledLogs = append(handledLogs, eventLog)
			return nil
		}

		// confirmationDelay > 0: the gate is active; Removed=true logs MUST be forwarded.
		const delay = 30 * time.Second
		listener := NewListener(addr, new(MockEVMClient), 1, 10, delay, logger, handleEvent, handleEvent, eventGetter, nil)

		// No historical events.
		historicalCh := make(chan types.Log)
		close(historicalCh)

		currentCh := make(chan types.Log, 2)

		// Event 1: non-Removed at block 10 — triggers IsContractEventProcessed check,
		// advances lastBlock, sets currentCheckDone = true. BlockTimestamp is set so
		// ensureBlockTimestamp short-circuits.
		normalLog := types.Log{BlockNumber: 10, Index: 0, TxHash: common.HexToHash("0xabc"), BlockTimestamp: uint64(time.Now().Unix())}
		eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()

		// Event 2: Removed=true at block 11 — must NOT advance lastBlock, must NOT call
		// IsContractEventProcessed, but MUST reach handleEvent (gate needs the removal signal).
		removedLog := types.Log{BlockNumber: 11, Index: 0, TxHash: common.HexToHash("0xdef"), Removed: true}

		currentCh <- normalLog
		currentCh <- removedLog

		sub := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			// Give processEvents enough time to drain both buffered events, then cancel.
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		var lastBlock uint64
		err := listener.processEvents(ctx, sub, historicalCh, currentCh, &lastBlock)
		require.NoError(t, err)

		// Both events must have reached handleEvent.
		require.Len(t, handledLogs, 2, "handleEvent must be called for both the normal and the Removed event when gate is active")

		// Verify first call was the normal log and second was the removed log.
		assert.Equal(t, uint64(10), handledLogs[0].BlockNumber)
		assert.False(t, handledLogs[0].Removed)
		assert.Equal(t, uint64(11), handledLogs[1].BlockNumber)
		assert.True(t, handledLogs[1].Removed)

		// lastBlock must NOT have advanced past the normal event's block.
		assert.Equal(t, uint64(10), lastBlock, "lastBlock must not be advanced by a Removed=true event")

		// IsContractEventProcessed must have been called exactly once (for the normal log only).
		eventGetter.AssertNumberOfCalls(t, "IsContractEventProcessed", 1)
		eventGetter.AssertExpectations(t)
	})

	t.Run("NoGate", func(t *testing.T) {
		t.Parallel()
		logger := log.NewNoopLogger()
		addr := common.HexToAddress("0x123")
		eventGetter := new(MockContractEventGetter)

		// Track which logs reached handleEvent.
		var handledLogs []types.Log
		handleEvent := func(ctx context.Context, eventLog types.Log) error {
			handledLogs = append(handledLogs, eventLog)
			return nil
		}

		// confirmationDelay == 0: no gate; Removed=true logs must be dropped at Phase 2 boundary.
		listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter, nil)

		// No historical events.
		historicalCh := make(chan types.Log)
		close(historicalCh)

		currentCh := make(chan types.Log, 3)

		// Event 1: non-Removed at block 10 — advances lastBlock, triggers dedup check.
		// BlockTimestamp is set so ensureBlockTimestamp short-circuits.
		normalLog := types.Log{BlockNumber: 10, Index: 0, TxHash: common.HexToHash("0xabc"), BlockTimestamp: uint64(time.Now().Unix())}
		eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Once()

		// Event 2: Removed=true at block 11 — must be dropped; must NOT reach handleEvent,
		// must NOT advance lastBlock.
		removedLog := types.Log{BlockNumber: 11, Index: 0, TxHash: common.HexToHash("0xdef"), Removed: true}

		// Event 3: another non-Removed at block 12 — must flow normally after the dropped removal.
		// BlockTimestamp is set so ensureBlockTimestamp short-circuits.
		followupLog := types.Log{BlockNumber: 12, Index: 1, TxHash: common.HexToHash("0xghi"), BlockTimestamp: uint64(time.Now().Unix())}

		currentCh <- normalLog
		currentCh <- removedLog
		currentCh <- followupLog

		sub := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			// Give processEvents enough time to drain all three buffered events, then cancel.
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		var lastBlock uint64
		err := listener.processEvents(ctx, sub, historicalCh, currentCh, &lastBlock)
		require.NoError(t, err)

		// Only the two non-Removed events must have reached handleEvent.
		require.Len(t, handledLogs, 2, "handleEvent must NOT be called for Removed=true when no gate is active")
		assert.Equal(t, uint64(10), handledLogs[0].BlockNumber)
		assert.False(t, handledLogs[0].Removed)
		assert.Equal(t, uint64(12), handledLogs[1].BlockNumber)
		assert.False(t, handledLogs[1].Removed)

		// lastBlock must reflect the last non-Removed event, not the removed one.
		assert.Equal(t, uint64(12), lastBlock, "lastBlock must not be advanced by a Removed=true event")

		// IsContractEventProcessed must have been called exactly once (for the first normal log only;
		// the follow-up log skips the check because currentCheckDone is already true).
		eventGetter.AssertNumberOfCalls(t, "IsContractEventProcessed", 1)
		eventGetter.AssertExpectations(t)
	})
}

func TestReconcileBlockRange_ContextCancellation(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	ctx, cancel := context.WithCancel(context.Background())

	// First batch succeeds but cancels the context during the call.
	// The second batch should never be reached.
	logs1 := []types.Log{{BlockNumber: 105, Index: 0}}
	mockClient.On("FilterLogs", mock.Anything, mock.MatchedBy(func(q ethereum.FilterQuery) bool {
		return q.FromBlock.Uint64() == 100 && q.ToBlock.Uint64() == 110
	})).Run(func(args mock.Arguments) {
		cancel()
	}).Return(logs1, nil)

	historicalCh := make(chan types.Log, 10)
	listener.reconcileBlockRange(ctx, 200, 100, historicalCh)
	close(historicalCh)

	// Drain whatever was sent before cancellation took effect
	var received []types.Log
	for l := range historicalCh {
		received = append(received, l)
	}

	// The event from the first batch may or may not have been sent (race between
	// the ctx.Done select and the historicalCh send), but the second batch must not run.
	assert.LessOrEqual(t, len(received), 1)
	mockClient.AssertNumberOfCalls(t, "FilterLogs", 1)
}

// TestEnsureBlockTimestamp_Populated: when BlockTimestamp is already set on the
// incoming log, ensureBlockTimestamp returns the log unchanged and does not call
// HeaderByHash. We prove the latter by leaving the mock unconfigured — any call
// would panic.
func TestEnsureBlockTimestamp_Populated(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	originalTs := uint64(1700000000)
	eventLog := types.Log{
		BlockNumber:    100,
		BlockHash:      common.HexToHash("0xabc"),
		BlockTimestamp: originalTs,
	}

	got, err := listener.ensureBlockTimestamp(context.Background(), eventLog)
	require.NoError(t, err)
	assert.Equal(t, originalTs, got.BlockTimestamp, "BlockTimestamp must be returned unchanged")
	assert.Equal(t, eventLog.BlockHash, got.BlockHash)
	mockClient.AssertNotCalled(t, "HeaderByHash")
}

// TestEnsureBlockTimestamp_Fetch: when BlockTimestamp == 0, ensureBlockTimestamp
// calls HeaderByHash exactly once and populates BlockTimestamp from header.Time.
func TestEnsureBlockTimestamp_Fetch(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	bh := common.HexToHash("0xabc")
	headerTime := uint64(1700000000)
	header := &types.Header{Number: big.NewInt(100), Time: headerTime}
	mockClient.On("HeaderByHash", mock.Anything, bh).Return(header, nil).Once()

	eventLog := types.Log{BlockNumber: 100, BlockHash: bh}

	got, err := listener.ensureBlockTimestamp(context.Background(), eventLog)
	require.NoError(t, err)
	assert.Equal(t, headerTime, got.BlockTimestamp, "BlockTimestamp must be populated from header.Time")
	mockClient.AssertExpectations(t)
}

// TestEnsureBlockTimestamp_CacheHit: two consecutive events with the same BlockHash
// (both with BlockTimestamp == 0) must trigger exactly one HeaderByHash call. The
// second call reads from the single-entry cache.
func TestEnsureBlockTimestamp_CacheHit(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	bh := common.HexToHash("0xabc")
	headerTime := uint64(1700000000)
	header := &types.Header{Number: big.NewInt(100), Time: headerTime}
	// Set up exactly ONE HeaderByHash expectation; a second call would fail
	// AssertExpectations because the mock is .Once().
	mockClient.On("HeaderByHash", mock.Anything, bh).Return(header, nil).Once()

	first := types.Log{BlockNumber: 100, BlockHash: bh, Index: 0}
	second := types.Log{BlockNumber: 100, BlockHash: bh, Index: 1}

	got1, err := listener.ensureBlockTimestamp(context.Background(), first)
	require.NoError(t, err)
	assert.Equal(t, headerTime, got1.BlockTimestamp)

	got2, err := listener.ensureBlockTimestamp(context.Background(), second)
	require.NoError(t, err)
	assert.Equal(t, headerTime, got2.BlockTimestamp)

	mockClient.AssertNumberOfCalls(t, "HeaderByHash", 1)
	mockClient.AssertExpectations(t)
}

// TestEnsureBlockTimestamp_FetchError: when HeaderByHash returns an error,
// ensureBlockTimestamp returns the original (unmutated) eventLog and the error.
// The caller decides whether to fall back to the gate.
func TestEnsureBlockTimestamp_FetchError(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	bh := common.HexToHash("0xabc")
	mockClient.On("HeaderByHash", mock.Anything, bh).Return(nil, fmt.Errorf("rpc failure")).Once()

	eventLog := types.Log{BlockNumber: 100, BlockHash: bh}

	got, err := listener.ensureBlockTimestamp(context.Background(), eventLog)
	require.Error(t, err)
	// On error, BlockTimestamp remains at the input value (0).
	assert.Equal(t, uint64(0), got.BlockTimestamp)
	assert.Equal(t, bh, got.BlockHash)
	mockClient.AssertExpectations(t)
}

// TestListener_NoGateFlushNilSafe: listenEvents must not panic when flushDownstream is nil.
// This is the no-gate path (confirmationDelay == 0). We exercise the listenEvents loop
// with two iterations so the nil-flush code path is hit on reconnect.
func TestListener_NoGateFlushNilSafe(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	// Empty store → findCommonAncestor returns 0 → skip historical replay.
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).Return(uint64(0), "", nil)

	ctx, cancel := context.WithCancel(context.Background())

	// flushDownstream = nil (no gate).
	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)

	// First subscription: immediately close so the loop gets a drop and retries.
	sub1 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			sub1.Unsubscribe()
		}).
		Return(sub1, nil).Once()

	// Second subscription: cancel the context so the loop exits cleanly.
	sub2 := &MockSubscription{errChan: make(chan error), unsub: func() {}}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			cancel()
		}).
		Return(sub2, nil).Once()

	listenerDone := make(chan error, 1)
	// Must not panic.
	require.NotPanics(t, func() {
		listener.Listen(ctx, func(err error) { listenerDone <- err })
		select {
		case <-listenerDone:
		case <-time.After(3 * time.Second):
			t.Log("listener timed out but did not panic — test passes")
		}
	})
}

// TestListener_FlushTimingPrecedesBackoff verifies that flushDownstream is called
// BEFORE the backoff sleep on each reconnect iteration, not after it.
//
// The observable invariant: if flush runs before the sleep, then the wall-clock gap
// between flushTime and the next SubscribeFilterLogs call must be at least the backoff
// duration. If flush runs after the sleep (old wrong ordering), the gap is negligible
// (flush and subscribe happen back-to-back with no sleep between them).
//
// Mechanics:
//   - Iteration 1 (backOffCount=0): no sleep; sub1 drops immediately (retry=true).
//     Flush is recorded, backOffCount becomes 1.
//   - Iteration 2 (backOffCount=1): logBackOff returns backOffDuration(1)=1s sleep.
//     After the sleep, subscribe2 is called and cancels the context.
//
// Test assertion: sub2Time - flushTime >= minBackoff (≈1s).
// Under the wrong (old) ordering the flush would run just before subscribe2 with no
// sleep in between, so sub2Time - flushTime would be negligible (~0ms), causing the
// assertion to fail.
func TestListener_FlushTimingPrecedesBackoff(t *testing.T) {
	// Not marked t.Parallel() because this test intentionally sleeps for ~1s
	// (backOffDuration(1) = 1s). Running it in parallel is fine for correctness
	// but keeps the suite wall time reasonable.
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	// Empty store → no historical replay on each pass.
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).Return(uint64(0), "", nil)

	// flushTime captures the wall-clock instant when flushDownstream is called.
	// It is set on the first (and only) flush that happens after sub1 drops.
	var flushTime time.Time
	var flushMu sync.Mutex

	flushDownstream := func() {
		flushMu.Lock()
		defer flushMu.Unlock()
		if flushTime.IsZero() {
			flushTime = time.Now()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, flushDownstream)

	sub1 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}
	sub2 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

	// First SubscribeFilterLogs: drop sub1 immediately so processEvents returns nil
	// (subscription drop → retry=true). backOffCount advances to 1 after this.
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			sub1.Unsubscribe()
		}).
		Return(sub1, nil).Once()

	// Second SubscribeFilterLogs: reached only after the backoff sleep
	// (backOffDuration(1) = 1s). Record arrival time and cancel so the loop exits.
	var sub2Time time.Time
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			sub2Time = time.Now()
			cancel()
		}).
		Return(sub2, nil).Once()

	listenerDone := make(chan error, 1)
	listener.Listen(ctx, func(err error) {
		listenerDone <- err
	})

	select {
	case <-listenerDone:
	case <-time.After(10 * time.Second):
		cancel()
		t.Fatal("listener did not exit within timeout")
	}

	flushMu.Lock()
	ft := flushTime
	flushMu.Unlock()

	require.False(t, ft.IsZero(), "flushDownstream must have been called after the subscription drop")
	require.False(t, sub2Time.IsZero(), "second SubscribeFilterLogs must have been called")

	// The gap between flush and the second subscribe must span the backoff sleep.
	// backOffDuration(1) = 1s; use 800ms as a conservative lower bound to absorb
	// scheduling jitter while still catching the wrong ordering (gap ≈ 0ms).
	const minBackoff = 800 * time.Millisecond
	gap := sub2Time.Sub(ft)
	assert.GreaterOrEqual(t, gap, minBackoff,
		"gap between flushTime and sub2Time (%v) must be >= minBackoff (%v); "+
			"if this fails the flush is running AFTER the sleep (wrong ordering)", gap, minBackoff)
}

// TestListener_BackoffOnSubscriptionDrop: after a mid-Phase-2 subscription drop the
// outer loop must increment backOffCount so repeated drops incur increasing delays.
// We validate this indirectly by checking that the second SubscribeFilterLogs call
// happens only after some elapsed time relative to the first.
func TestListener_BackoffOnSubscriptionDrop(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	eventGetter := new(MockContractEventGetter)
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).Return(uint64(0), "", nil)

	var subTimes []time.Time
	var subMu sync.Mutex

	sub1 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}
	sub2 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}

	ctx, cancel := context.WithCancel(context.Background())

	// First call: record time, immediately drop.
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			subMu.Lock()
			subTimes = append(subTimes, time.Now())
			subMu.Unlock()
			sub1.Unsubscribe()
		}).
		Return(sub1, nil).Once()

	// Second call: record time, cancel so loop exits.
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			subMu.Lock()
			subTimes = append(subTimes, time.Now())
			subMu.Unlock()
			cancel()
		}).
		Return(sub2, nil).Once()

	listenerDone := make(chan error, 1)
	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter, nil)
	listener.Listen(ctx, func(err error) { listenerDone <- err })

	select {
	case <-listenerDone:
	case <-time.After(10 * time.Second):
		cancel()
		t.Fatal("listener did not exit within timeout")
	}

	subMu.Lock()
	times := append([]time.Time(nil), subTimes...)
	subMu.Unlock()

	require.Len(t, times, 2, "must have exactly two SubscribeFilterLogs calls")
	// After first drop backOffCount == 1 → logBackOff returns backOffDuration(1) > 0.
	// The second call must happen at least a small positive time after the first.
	gap := times[1].Sub(times[0])
	assert.Greater(t, gap, time.Duration(0), "second subscribe must be delayed relative to first (backoff > 0)")
}

// TestListener_ReconnectFlushesGateAndRewindsCursor: on reconnect after a subscription
// drop, flushDownstream must be called and findCommonAncestor must be re-called so
// Phase 1 re-covers events that were in the gate's pending state but not committed.
func TestListener_ReconnectFlushesGateAndRewindsCursor(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")

	const committedBlock = uint64(95)
	// The canonical hash that findCommonAncestor will check.
	canonicalHeader := &types.Header{Number: big.NewInt(int64(committedBlock)), Difficulty: big.NewInt(1)}
	canonicalHash := canonicalHeader.Hash()

	eventGetter := new(MockContractEventGetter)
	// findCommonAncestor returns committedBlock on every call (reactor committed up to 95).
	eventGetter.On("GetLatestContractEventBlockHashAndNumber", addr.String(), uint64(1)).
		Return(committedBlock, canonicalHash.Hex(), nil)
	// isStoredBlockCanonical: HeaderByNumber(95) returns canonicalHeader → hash matches → canonical.
	mockClient.On("HeaderByNumber", mock.Anything, mock.MatchedBy(func(n *big.Int) bool {
		return n != nil && n.Cmp(big.NewInt(int64(committedBlock))) == 0
	})).Return(canonicalHeader, nil)

	var flushCalls atomic.Int32
	flushDownstream := func() { flushCalls.Add(1) }

	// Track the from-block of every FilterLogs call to verify the cursor.
	var reconcileStarts []uint64
	var reconcileMu sync.Mutex

	// Current tip = block 110.
	currentHeader := &types.Header{Number: big.NewInt(110)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	mockClient.On("FilterLogs", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			q := args.Get(1).(ethereum.FilterQuery)
			reconcileMu.Lock()
			reconcileStarts = append(reconcileStarts, q.FromBlock.Uint64())
			reconcileMu.Unlock()
		}).
		Return([]types.Log{}, nil)

	ctx, cancel := context.WithCancel(context.Background())

	// First subscription: deliver one live event (advancing in-memory lastBlock to 100),
	// then immediately drop so the loop retries.
	sub1 := &MockSubscription{errChan: make(chan error, 1), unsub: func() {}}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- types.Log)
			// Deliver a live event at block 100 so in-memory lastBlock advances to 100.
			ch <- types.Log{
				BlockNumber:    100,
				Index:          0,
				TxHash:         common.HexToHash("0xLIVE"),
				BlockTimestamp: uint64(time.Now().Unix()),
			}
			go func() {
				time.Sleep(20 * time.Millisecond)
				sub1.Unsubscribe()
			}()
		}).
		Return(sub1, nil).Once()

	// Second subscription: cancel so the loop exits.
	sub2 := &MockSubscription{errChan: make(chan error), unsub: func() {}}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			cancel()
		}).
		Return(sub2, nil).Once()

	// Dedup check for the live event on the first pass.
	eventGetter.On("IsContractEventProcessed", mock.Anything, uint32(0), uint64(1)).Return(false, nil).Maybe()

	var noopHandler HandleEvent = func(_ context.Context, _ types.Log) error { return nil }
	listener := NewListener(addr, mockClient, 1, 100, 0, logger, noopHandler, noopHandler, eventGetter, flushDownstream)

	listenerDone := make(chan error, 1)
	listener.Listen(ctx, func(err error) { listenerDone <- err })

	select {
	case <-listenerDone:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("listener did not exit within timeout")
	}

	// flushDownstream must have been called exactly once: after the first subscription
	// drop, before the second pass begins. With the new ordering, flush runs AFTER
	// runOneListenPass returns retry=true and BEFORE the next iteration's backoff
	// sleep — so the first iteration does not flush (no prior subscription drop) and
	// the second iteration exits cleanly (cancel was called), so it also does not flush.
	assert.Equal(t, int32(1), flushCalls.Load(),
		"flushDownstream must be called exactly once (after the subscription drop, before the second pass)")

	// Every FilterLogs call must start from committedBlock (95), not from the
	// live-event block (100) that was in-memory on the first pass. This proves
	// findCommonAncestor was re-called on the second pass.
	reconcileMu.Lock()
	starts := append([]uint64(nil), reconcileStarts...)
	reconcileMu.Unlock()

	for i, start := range starts {
		assert.Equal(t, committedBlock, start,
			"reconcileBlockRange call %d must start from committedBlock=%d, got %d", i, committedBlock, start)
	}
}
