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
	l := NewListener(addr, mockClient, 1, 100, logger, nil, eventGetter)
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
	eventGetter.On("GetLatestContractEventBlockNumber", addr.String(), uint64(1)).Return(uint64(0), nil)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Channel to signal event handling
	eventHandled := make(chan struct{})
	handleEvent := func(ctx context.Context, log types.Log) error {
		cancel()
		close(eventHandled)
		return nil
	}

	listener := NewListener(addr, mockClient, 1, 100, logger, handleEvent, eventGetter)

	// Mock SubscribeFilterLogs
	sub := &MockSubscription{
		errChan: make(chan error),
		unsub:   func() {},
	}

	// Mock SubscribeFilterLogs: send a log immediately
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- types.Log)
			// Send a log immediately
			ch <- types.Log{BlockNumber: 10, Index: 1}
		}).
		Return(sub, nil)

	// The first current event will trigger IsContractEventPresent check
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(10), mock.Anything, uint32(1)).Return(false, nil)

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
	listener := NewListener(addr, mockClient, 1, 10, logger, nil, eventGetter)

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

	// Start from block 100
	eventGetter := new(MockContractEventGetter)
	eventGetter.On("GetLatestContractEventBlockNumber", addr.String(), uint64(1)).Return(uint64(100), nil)
	// Historical event at block 105 is not present
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(105), mock.Anything, uint32(0)).Return(false, nil)
	// Current event at block 111 — after historical is done, first current event triggers check
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(111), mock.Anything, uint32(0)).Return(false, nil)

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

	listener := NewListener(addr, mockClient, 1, 10, logger, handleEvent, eventGetter)

	// Mock HeaderByNumber (current tip is 110)
	currentHeader := &types.Header{Number: big.NewInt(110)}
	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(currentHeader, nil)

	// Mock FilterLogs (100-110)
	histLogs := []types.Log{{BlockNumber: 105, Index: 0}}
	mockClient.On("FilterLogs", mock.Anything, mock.Anything).Return(histLogs, nil)

	// Mock SubscribeFilterLogs
	sub := &MockSubscription{errChan: make(chan error)}
	mockClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- types.Log)
			// Send a current log
			ch <- types.Log{BlockNumber: 111, Index: 0}
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

	listener := NewListener(addr, new(MockEVMClient), 1, 10, logger, handleEvent, eventGetter)

	// Historical: 3 events. First 2 are present (skipped), 3rd is not (handled).
	// After the 3rd, the check should stop — no IsContractEventPresent call for events 4+.
	historicalCh := make(chan types.Log, 5)
	historicalCh <- types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaa")}
	historicalCh <- types.Log{BlockNumber: 101, Index: 0, TxHash: common.HexToHash("0xbb")}
	historicalCh <- types.Log{BlockNumber: 102, Index: 0, TxHash: common.HexToHash("0xcc")}
	historicalCh <- types.Log{BlockNumber: 103, Index: 0, TxHash: common.HexToHash("0xdd")}
	historicalCh <- types.Log{BlockNumber: 104, Index: 0, TxHash: common.HexToHash("0xee")}
	close(historicalCh)

	// First two are present, third is not
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(100), mock.Anything, uint32(0)).Return(true, nil).Once()
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(101), mock.Anything, uint32(0)).Return(true, nil).Once()
	eventGetter.On("IsContractEventPresent", uint64(1), uint64(102), mock.Anything, uint32(0)).Return(false, nil).Once()
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

	listener := NewListener(addr, new(MockEVMClient), 1, 10, logger, handleEvent, eventGetter)

	// Historical channel with events that will block (not closed yet)
	historicalCh := make(chan types.Log, 2)
	historicalCh <- types.Log{BlockNumber: 100, Index: 0, TxHash: common.HexToHash("0xaa")}

	eventGetter.On("IsContractEventPresent", uint64(1), uint64(100), mock.Anything, uint32(0)).Return(false, nil)

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

func TestReconcileBlockRange_ContextCancellation(t *testing.T) {
	t.Parallel()
	mockClient := new(MockEVMClient)
	logger := log.NewNoopLogger()
	addr := common.HexToAddress("0x123")
	eventGetter := new(MockContractEventGetter)

	listener := NewListener(addr, mockClient, 1, 10, logger, nil, eventGetter)

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
