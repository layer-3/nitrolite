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
	l := NewListener(addr, mockClient, 1, 100, 0, logger, nil, nil, eventGetter)
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

	listener := NewListener(addr, mockClient, 1, 100, 0, logger, handleEvent, handleEvent, eventGetter)

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
	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.reconcileBlockRange(context.Background(), 120, 100, historicalCh)
		close(historicalCh)
	}()

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

	var receivedCount int64
	doneCh := make(chan struct{})

	handleEvent := func(ctx context.Context, log types.Log) error {
		count := atomic.AddInt64(&receivedCount, 1)
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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, handleEvent, handleEvent, eventGetter)

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

	listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter)

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

	listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, confirmationDelay, logger, liveHandler, historicalHandler, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, liveHandler, historicalHandler, eventGetter)

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
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	// Track which logs reached handleEvent.
	var handledLogs []types.Log
	handleEvent := func(ctx context.Context, eventLog types.Log) error {
		handledLogs = append(handledLogs, eventLog)
		return nil
	}

	listener := NewListener(addr, new(MockEVMClient), 1, 10, 0, logger, handleEvent, handleEvent, eventGetter)

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
	// IsContractEventProcessed, but MUST reach handleEvent.
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
	require.Len(t, handledLogs, 2, "handleEvent must be called for both the normal and the Removed event")

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
}

func TestReconcileBlockRange_ContextCancellation(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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

	listener := NewListener(addr, mockClient, 1, 10, 0, logger, nil, nil, eventGetter)

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
