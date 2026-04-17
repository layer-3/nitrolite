package evm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
)

// mockLockingStore implements LockingContractReactorStore for testing
type mockLockingStore struct {
	mock.Mock
}

func (m *mockLockingStore) UpdateUserStaked(wallet string, blockchainID uint64, amount decimal.Decimal) error {
	args := m.Called(wallet, blockchainID, amount)
	return args.Error(0)
}

func (m *mockLockingStore) StoreContractEvent(ev core.BlockchainEvent) error {
	args := m.Called(ev)
	return args.Error(0)
}

func TestAppRegistryReactor_HandleLocked(t *testing.T) {
	userAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	deposited := big.NewInt(500_000_000)    // 500 USDC (6 decimals)
	newBalance := big.NewInt(1_000_000_000) // 1000 USDC
	var tokenDecimals int32 = 6
	blockchainID := uint64(1)

	// ABI-encode the non-indexed parameters: deposited, newBalance
	depositedPadded := common.LeftPadBytes(deposited.Bytes(), 32)
	newBalancePadded := common.LeftPadBytes(newBalance.Bytes(), 32)
	data := append(depositedPadded, newBalancePadded...)

	lockedEventID := lockingContractAbi.Events["Locked"].ID
	logEntry := types.Log{
		Topics: []common.Hash{
			lockedEventID,
			common.BytesToHash(userAddr.Bytes()), // indexed user
		},
		Data:        data,
		BlockNumber: 100,
		TxHash:      common.HexToHash("0xdeadbeef"),
		Index:       0,
	}

	t.Run("success", func(t *testing.T) {
		store := new(mockLockingStore)
		var capturedEvent *core.UserLockedBalanceUpdatedEvent
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, _ core.LockingContractEventHandlerStore, ev *core.UserLockedBalanceUpdatedEvent) error {
				capturedEvent = ev
				return nil
			},
		}

		useStoreInTx := func(handler LockingContractReactorStoreTxHandler) error {
			return handler(store)
		}

		// Expect StoreContractEvent to be called
		store.On("StoreContractEvent", mock.MatchedBy(func(ev core.BlockchainEvent) bool {
			return ev.Name == "Locked" && ev.BlockNumber == 100 && ev.BlockchainID == blockchainID
		})).Return(nil)

		reactor, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return uint8(tokenDecimals), nil
		}, useStoreInTx)
		require.NoError(t, err)

		var processedSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			processedSuccess = success
		})

		reactor.HandleEvent(context.Background(), logEntry)

		require.NotNil(t, capturedEvent)
		assert.Equal(t, userAddr.String(), common.HexToAddress(capturedEvent.UserAddress).String())
		assert.Equal(t, blockchainID, capturedEvent.BlockchainID)
		expectedBalance := decimal.NewFromBigInt(newBalance, -tokenDecimals)
		assert.True(t, expectedBalance.Equal(capturedEvent.Balance), "expected %s, got %s", expectedBalance, capturedEvent.Balance)

		assert.True(t, processedSuccess)
		store.AssertExpectations(t)
	})

	t.Run("getTokenDecimals error", func(t *testing.T) {
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, _ core.LockingContractEventHandlerStore, _ *core.UserLockedBalanceUpdatedEvent) error {
				t.Fatal("handler should not be called")
				return nil
			},
		}

		useStoreInTx := func(handler LockingContractReactorStoreTxHandler) error {
			return handler(nil)
		}

		_, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return 0, assert.AnError
		}, useStoreInTx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get token decimals")
	})

	t.Run("handler error", func(t *testing.T) {
		store := new(mockLockingStore)
		handler := &mockAppRegistryEventHandler{
			handleFn: func(_ context.Context, _ core.LockingContractEventHandlerStore, _ *core.UserLockedBalanceUpdatedEvent) error {
				return assert.AnError
			},
		}

		useStoreInTx := func(handler LockingContractReactorStoreTxHandler) error {
			return handler(store)
		}

		reactor, err := NewLockingContractReactor(blockchainID, handler, func() (uint8, error) {
			return uint8(tokenDecimals), nil
		}, useStoreInTx)
		require.NoError(t, err)

		var processedSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			processedSuccess = success
		})

		reactor.HandleEvent(context.Background(), logEntry)
		assert.False(t, processedSuccess)
	})
}

type mockAppRegistryEventHandler struct {
	handleFn func(context.Context, core.LockingContractEventHandlerStore, *core.UserLockedBalanceUpdatedEvent) error
}

func (m *mockAppRegistryEventHandler) HandleUserLockedBalanceUpdated(ctx context.Context, tx core.LockingContractEventHandlerStore, ev *core.UserLockedBalanceUpdatedEvent) error {
	return m.handleFn(ctx, tx, ev)
}
