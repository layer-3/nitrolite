package evm

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/core"
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

	l := NewListener(addr, mockClient, 1, 100, logger, nil, nil)
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

	// Setup latest event getter (start from 0)
	getLatestEvent := func(contractAddress string, networkID uint64) (core.BlockchainEvent, error) {
		return core.BlockchainEvent{BlockNumber: 0, LogIndex: 0}, nil
	}

	// Channel to signal event handling
	eventHandled := make(chan struct{})
	handleEvent := func(ctx context.Context, log types.Log) {
		close(eventHandled)
	}

	listener := NewListener(addr, mockClient, 1, 100, logger, handleEvent, getLatestEvent)

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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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

	listener := NewListener(addr, mockClient, 1, 10, logger, nil, nil)

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
		listener.reconcileBlockRange(120, 100, 0, historicalCh)
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
	getLatestEvent := func(contractAddress string, networkID uint64) (core.BlockchainEvent, error) {
		return core.BlockchainEvent{BlockNumber: 100, LogIndex: 0}, nil
	}

	var receivedCount int64
	doneCh := make(chan struct{})

	handleEvent := func(ctx context.Context, log types.Log) {
		count := atomic.AddInt64(&receivedCount, 1)
		if count >= 2 { // Expect 1 historical + 1 current
			select {
			case <-doneCh:
			default:
				close(doneCh)
			}
		}
	}

	listener := NewListener(addr, mockClient, 1, 10, logger, handleEvent, getLatestEvent)

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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go listener.Listen(ctx, func(err error) {})

	select {
	case <-doneCh:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for events")
	}
}
