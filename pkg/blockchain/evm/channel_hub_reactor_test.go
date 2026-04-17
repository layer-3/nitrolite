package evm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/core"
)

// mockChannelHubStore implements ChannelHubReactorStore for testing.
type mockChannelHubStore struct {
	mock.Mock
}

func (m *mockChannelHubStore) GetLastStateByChannelID(channelID string, signed bool) (*core.State, error) {
	args := m.Called(channelID, signed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.State), args.Error(1)
}

func (m *mockChannelHubStore) GetStateByChannelIDAndVersion(channelID string, version uint64) (*core.State, error) {
	args := m.Called(channelID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.State), args.Error(1)
}

func (m *mockChannelHubStore) UpdateChannel(channel core.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

func (m *mockChannelHubStore) GetChannelByID(channelID string) (*core.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Channel), args.Error(1)
}

func (m *mockChannelHubStore) ScheduleCheckpoint(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

func (m *mockChannelHubStore) ScheduleInitiateEscrowDeposit(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

func (m *mockChannelHubStore) ScheduleFinalizeEscrowDeposit(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

func (m *mockChannelHubStore) ScheduleFinalizeEscrowWithdrawal(stateID string, chainID uint64) error {
	args := m.Called(stateID, chainID)
	return args.Error(0)
}

func (m *mockChannelHubStore) SetNodeBalance(blockchainID uint64, asset string, value decimal.Decimal) error {
	args := m.Called(blockchainID, asset, value)
	return args.Error(0)
}

func (m *mockChannelHubStore) RefreshUserEnforcedBalance(wallet, asset string) error {
	args := m.Called(wallet, asset)
	return args.Error(0)
}

func (m *mockChannelHubStore) StoreContractEvent(ev core.BlockchainEvent) error {
	args := m.Called(ev)
	return args.Error(0)
}

// mockChannelHubEventHandler captures events dispatched by the reactor.
type mockChannelHubEventHandler struct {
	mock.Mock
}

func (m *mockChannelHubEventHandler) HandleNodeBalanceUpdated(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.NodeBalanceUpdatedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleHomeChannelCreated(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.HomeChannelCreatedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleHomeChannelMigrated(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.HomeChannelMigratedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleHomeChannelCheckpointed(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.HomeChannelCheckpointedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleHomeChannelChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.HomeChannelChallengedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleHomeChannelClosed(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.HomeChannelClosedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowDepositInitiated(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowDepositInitiatedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowDepositChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowDepositChallengedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowDepositFinalized(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowDepositFinalizedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowWithdrawalInitiated(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowWithdrawalInitiatedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowWithdrawalChallenged(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowWithdrawalChallengedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

func (m *mockChannelHubEventHandler) HandleEscrowWithdrawalFinalized(ctx context.Context, tx core.ChannelHubEventHandlerStore, ev *core.EscrowWithdrawalFinalizedEvent) error {
	args := m.Called(ctx, tx, ev)
	return args.Error(0)
}

// makeState returns a minimal valid ABI State struct with the given version.
func makeState(version uint64) State {
	return State{
		Version:  version,
		Intent:   0,
		Metadata: [32]byte{},
		HomeLedger: Ledger{
			ChainId:        1,
			Token:          common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			Decimals:       6,
			UserAllocation: big.NewInt(1000),
			UserNetFlow:    big.NewInt(0),
			NodeAllocation: big.NewInt(2000),
			NodeNetFlow:    big.NewInt(0),
		},
		NonHomeLedger: Ledger{
			ChainId:        0,
			Token:          common.Address{},
			Decimals:       0,
			UserAllocation: big.NewInt(0),
			UserNetFlow:    big.NewInt(0),
			NodeAllocation: big.NewInt(0),
			NodeNetFlow:    big.NewInt(0),
		},
		UserSig: []byte{},
		NodeSig: []byte{},
	}
}

// packNonIndexed ABI-encodes the non-indexed parameters for the given event.
func packNonIndexed(t *testing.T, eventName string, args ...interface{}) []byte {
	t.Helper()
	ev, ok := channelHubAbi.Events[eventName]
	require.True(t, ok, "event %s not found in ABI", eventName)

	nonIndexed := ev.Inputs.NonIndexed()
	data, err := nonIndexed.Pack(args...)
	require.NoError(t, err, "failed to pack non-indexed args for %s", eventName)
	return data
}

// newReactor creates a ChannelHubReactor wired to the provided mocks.
func newReactor(blockchainID uint64, nodeAddress string, handler *mockChannelHubEventHandler, assetStore *MockAssetStore, store *mockChannelHubStore) *ChannelHubReactor {
	useStoreInTx := func(fn ChannelHubReactorStoreTxHandler) error {
		return fn(store)
	}
	return NewChannelHubReactor(blockchainID, nodeAddress, handler, assetStore, useStoreInTx)
}

// expectStoreContractEvent sets up the mock expectation for StoreContractEvent.
func expectStoreContractEvent(store *mockChannelHubStore, eventName string, blockNumber uint64, blockchainID uint64) {
	store.On("StoreContractEvent", mock.MatchedBy(func(ev core.BlockchainEvent) bool {
		return ev.Name == eventName && ev.BlockNumber == blockNumber && ev.BlockchainID == blockchainID
	})).Return(nil)
}

func TestChannelHubReactor_HandleNodeBalanceUpdated(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tokenAddr := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	amount := big.NewInt(1_000_000) // 1 USDC (6 decimals)

	data := packNonIndexed(t, "NodeBalanceUpdated", amount)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["NodeBalanceUpdated"].ID,
			common.BytesToHash(tokenAddr.Bytes()),
		},
		Data:        data,
		BlockNumber: 200,
		TxHash:      common.HexToHash("0xabc123"),
		Index:       0,
	}

	t.Run("success", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		assetStore.On("GetTokenAsset", blockchainID, tokenAddr.String()).Return("usdc", nil)
		assetStore.On("GetTokenDecimals", blockchainID, tokenAddr.String()).Return(uint8(6), nil)

		handler.On("HandleNodeBalanceUpdated", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.NodeBalanceUpdatedEvent) bool {
			return ev.BlockchainID == blockchainID &&
				ev.Asset == "usdc" &&
				ev.Balance.Equal(decimal.NewFromBigInt(amount, -6))
		})).Return(nil)

		expectStoreContractEvent(store, "NodeBalanceUpdated", 200, blockchainID)

		reactor := newReactor(blockchainID, nodeAddr.String(), handler, assetStore, store)
		err := reactor.HandleEvent(context.Background(), logEntry)
		require.NoError(t, err)
		handler.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("handler error", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		assetStore.On("GetTokenAsset", blockchainID, tokenAddr.String()).Return("usdc", nil)
		assetStore.On("GetTokenDecimals", blockchainID, tokenAddr.String()).Return(uint8(6), nil)

		handler.On("HandleNodeBalanceUpdated", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		reactor := newReactor(blockchainID, nodeAddr.String(), handler, assetStore, store)

		var processedSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			processedSuccess = success
		})

		err := reactor.HandleEvent(context.Background(), logEntry)
		require.Error(t, err)
		assert.False(t, processedSuccess)
	})
}

func TestChannelHubReactor_HandleHomeChannelCreated(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	userAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	channelID := common.HexToHash("0xcc01")

	state := makeState(1)
	def := ChannelDefinition{
		ChallengeDuration:           3600,
		User:                        userAddr,
		Node:                        nodeAddr,
		Nonce:                       1,
		ApprovedSignatureValidators: big.NewInt(0),
		Metadata:                    [32]byte{},
	}

	data := packNonIndexed(t, "ChannelCreated", def, state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelCreated"].ID,
			channelID,
			common.BytesToHash(userAddr.Bytes()),
		},
		Data:        data,
		BlockNumber: 300,
		TxHash:      common.HexToHash("0xdef456"),
		Index:       1,
	}

	t.Run("success", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		handler.On("HandleHomeChannelCreated", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCreatedEvent) bool {
			return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 1
		})).Return(nil)

		expectStoreContractEvent(store, "ChannelCreated", 300, blockchainID)

		reactor := newReactor(blockchainID, nodeAddr.String(), handler, assetStore, store)
		err := reactor.HandleEvent(context.Background(), logEntry)
		require.NoError(t, err)
		handler.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("ignores other nodes", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		expectStoreContractEvent(store, "ChannelCreated", 300, blockchainID)

		otherNode := "0x3333333333333333333333333333333333333333"
		reactor := newReactor(blockchainID, otherNode, handler, assetStore, store)
		err := reactor.HandleEvent(context.Background(), logEntry)
		require.NoError(t, err)
		handler.AssertNotCalled(t, "HandleHomeChannelCreated")
		store.AssertExpectations(t)
	})
}

func TestChannelHubReactor_HandleHomeChannelCheckpointed(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc02")

	state := makeState(5)
	data := packNonIndexed(t, "ChannelCheckpointed", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelCheckpointed"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 400,
		TxHash:      common.HexToHash("0x111"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 5
	})).Return(nil)

	expectStoreContractEvent(store, "ChannelCheckpointed", 400, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleHomeChannelChallenged(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc03")

	state := makeState(4)
	challengeExpiry := uint64(9999999)
	data := packNonIndexed(t, "ChannelChallenged", state, challengeExpiry)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelChallenged"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 500,
		TxHash:      common.HexToHash("0x222"),
		Index:       2,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelChallenged", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelChallengedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) &&
			ev.StateVersion == 4 &&
			ev.ChallengeExpiry == challengeExpiry
	})).Return(nil)

	expectStoreContractEvent(store, "ChannelChallenged", 500, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleHomeChannelClosed(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc04")

	state := makeState(10)
	data := packNonIndexed(t, "ChannelClosed", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelClosed"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 600,
		TxHash:      common.HexToHash("0x333"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelClosed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelClosedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 10
	})).Return(nil)

	expectStoreContractEvent(store, "ChannelClosed", 600, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleChannelDeposited(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc05")

	state := makeState(7)
	data := packNonIndexed(t, "ChannelDeposited", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelDeposited"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 700,
		TxHash:      common.HexToHash("0x444"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	// ChannelDeposited dispatches HandleHomeChannelCheckpointed
	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 7
	})).Return(nil)

	expectStoreContractEvent(store, "ChannelDeposited", 700, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleChannelWithdrawn(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc06")

	state := makeState(8)
	data := packNonIndexed(t, "ChannelWithdrawn", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelWithdrawn"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 800,
		TxHash:      common.HexToHash("0x555"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	// ChannelWithdrawn dispatches HandleHomeChannelCheckpointed
	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 8
	})).Return(nil)

	expectStoreContractEvent(store, "ChannelWithdrawn", 800, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowDepositInitiated(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee01")
	channelID := common.HexToHash("0xcc01")

	state := makeState(1)
	data := packNonIndexed(t, "EscrowDepositInitiated", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowDepositInitiated"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 900,
		TxHash:      common.HexToHash("0x666"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowDepositInitiated", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowDepositInitiatedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) && ev.StateVersion == 1
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowDepositInitiated", 900, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowDepositChallenged(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee02")

	state := makeState(3)
	challengeExpiry := uint64(8888888)
	data := packNonIndexed(t, "EscrowDepositChallenged", state, challengeExpiry)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowDepositChallenged"].ID,
			escrowID,
		},
		Data:        data,
		BlockNumber: 1000,
		TxHash:      common.HexToHash("0x777"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowDepositChallenged", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowDepositChallengedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) &&
			ev.StateVersion == 3 &&
			ev.ChallengeExpiry == challengeExpiry
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowDepositChallenged", 1000, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowDepositFinalized(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee03")
	channelID := common.HexToHash("0xcc01")

	state := makeState(5)
	data := packNonIndexed(t, "EscrowDepositFinalized", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowDepositFinalized"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1100,
		TxHash:      common.HexToHash("0x888"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowDepositFinalized", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowDepositFinalizedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) && ev.StateVersion == 5
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowDepositFinalized", 1100, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowWithdrawalInitiated(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee04")
	channelID := common.HexToHash("0xcc01")

	state := makeState(1)
	data := packNonIndexed(t, "EscrowWithdrawalInitiated", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowWithdrawalInitiated"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1200,
		TxHash:      common.HexToHash("0x999"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowWithdrawalInitiated", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowWithdrawalInitiatedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) && ev.StateVersion == 1
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowWithdrawalInitiated", 1200, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowWithdrawalChallenged(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee05")

	state := makeState(3)
	challengeExpiry := uint64(7777777)
	data := packNonIndexed(t, "EscrowWithdrawalChallenged", state, challengeExpiry)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowWithdrawalChallenged"].ID,
			escrowID,
		},
		Data:        data,
		BlockNumber: 1300,
		TxHash:      common.HexToHash("0xaaa"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowWithdrawalChallenged", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowWithdrawalChallengedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) &&
			ev.StateVersion == 3 &&
			ev.ChallengeExpiry == challengeExpiry
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowWithdrawalChallenged", 1300, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowWithdrawalFinalized(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee06")
	channelID := common.HexToHash("0xcc01")

	state := makeState(5)
	data := packNonIndexed(t, "EscrowWithdrawalFinalized", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowWithdrawalFinalized"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1400,
		TxHash:      common.HexToHash("0xbbb"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleEscrowWithdrawalFinalized", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.EscrowWithdrawalFinalizedEvent) bool {
		return ev.ChannelID == hexutil.Encode(escrowID[:]) && ev.StateVersion == 5
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowWithdrawalFinalized", 1400, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowDepositInitiatedOnHome(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee07")
	channelID := common.HexToHash("0xcc01")

	state := makeState(3)
	data := packNonIndexed(t, "EscrowDepositInitiatedOnHome", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowDepositInitiatedOnHome"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1500,
		TxHash:      common.HexToHash("0xccc"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	// Dispatches HandleHomeChannelCheckpointed with the channelId topic
	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 3
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowDepositInitiatedOnHome", 1500, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowDepositFinalizedOnHome(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee08")
	channelID := common.HexToHash("0xcc01")

	state := makeState(6)
	data := packNonIndexed(t, "EscrowDepositFinalizedOnHome", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowDepositFinalizedOnHome"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1600,
		TxHash:      common.HexToHash("0xddd"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 6
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowDepositFinalizedOnHome", 1600, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowWithdrawalInitiatedOnHome(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee09")
	channelID := common.HexToHash("0xcc01")

	state := makeState(4)
	data := packNonIndexed(t, "EscrowWithdrawalInitiatedOnHome", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowWithdrawalInitiatedOnHome"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1700,
		TxHash:      common.HexToHash("0xeee"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 4
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowWithdrawalInitiatedOnHome", 1700, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_HandleEscrowWithdrawalFinalizedOnHome(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	escrowID := common.HexToHash("0xee10")
	channelID := common.HexToHash("0xcc01")

	state := makeState(9)
	data := packNonIndexed(t, "EscrowWithdrawalFinalizedOnHome", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["EscrowWithdrawalFinalizedOnHome"].ID,
			escrowID,
			channelID,
		},
		Data:        data,
		BlockNumber: 1800,
		TxHash:      common.HexToHash("0xfff"),
		Index:       0,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.MatchedBy(func(ev *core.HomeChannelCheckpointedEvent) bool {
		return ev.ChannelID == hexutil.Encode(channelID[:]) && ev.StateVersion == 9
	})).Return(nil)

	expectStoreContractEvent(store, "EscrowWithdrawalFinalizedOnHome", 1800, blockchainID)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
	handler.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestChannelHubReactor_UnknownEvent(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"

	logEntry := types.Log{
		Topics: []common.Hash{
			common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"),
		},
		BlockNumber: 999,
	}

	store := new(mockChannelHubStore)
	handler := new(mockChannelHubEventHandler)
	assetStore := new(MockAssetStore)

	reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)
	err := reactor.HandleEvent(context.Background(), logEntry)
	require.NoError(t, err)
}

func TestChannelHubReactor_OnEventProcessedCallback(t *testing.T) {
	blockchainID := uint64(1)
	nodeAddr := "0x1111111111111111111111111111111111111111"
	channelID := common.HexToHash("0xcc99")

	state := makeState(1)
	data := packNonIndexed(t, "ChannelCheckpointed", state)

	logEntry := types.Log{
		Topics: []common.Hash{
			channelHubAbi.Events["ChannelCheckpointed"].ID,
			channelID,
		},
		Data:        data,
		BlockNumber: 50,
		TxHash:      common.HexToHash("0xcb01"),
		Index:       0,
	}

	t.Run("callback receives true on success", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		expectStoreContractEvent(store, "ChannelCheckpointed", 50, blockchainID)

		reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)

		var cbBlockchainID uint64
		var cbSuccess bool
		reactor.SetOnEventProcessed(func(bid uint64, success bool) {
			cbBlockchainID = bid
			cbSuccess = success
		})

		err := reactor.HandleEvent(context.Background(), logEntry)
		require.NoError(t, err)
		assert.Equal(t, blockchainID, cbBlockchainID)
		assert.True(t, cbSuccess)
	})

	t.Run("callback receives false on handler error", func(t *testing.T) {
		store := new(mockChannelHubStore)
		handler := new(mockChannelHubEventHandler)
		assetStore := new(MockAssetStore)

		handler.On("HandleHomeChannelCheckpointed", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		reactor := newReactor(blockchainID, nodeAddr, handler, assetStore, store)

		var cbSuccess bool
		reactor.SetOnEventProcessed(func(_ uint64, success bool) {
			cbSuccess = success
		})

		err := reactor.HandleEvent(context.Background(), logEntry)
		require.Error(t, err)
		assert.False(t, cbSuccess)
	})
}
