package evm

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

// MockEVMClient implements EVMClient interface
type MockEVMClient struct {
	mock.Mock
}

// ChainStateReader methods
func (m *MockEVMClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVMClient) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, key, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVMClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVMClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(uint64), args.Error(1)
}

// ContractBackend methods
func (m *MockEVMClient) CodeAt2(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	// Redundant with CodeAt? Interface definition might vary slightly.
	// bind.ContractBackend includes CodeAt.
	return m.CodeAt(ctx, contract, blockNumber)
}

func (m *MockEVMClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVMClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEVMClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := m.Called(ctx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEVMClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEVMClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVMClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEVMClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, call)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEVMClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEVMClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *MockEVMClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := m.Called(ctx, query, ch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

// MockContractEventGetter implements ContractEventGetter interface
type MockContractEventGetter struct {
	mock.Mock
}

func (m *MockContractEventGetter) GetLatestContractEventBlockNumber(contractAddress string, blockchainID uint64) (uint64, error) {
	args := m.Called(contractAddress, blockchainID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockContractEventGetter) IsContractEventPresent(blockchainID, blockNumber uint64, txHash string, logIndex uint32) (bool, error) {
	args := m.Called(blockchainID, blockNumber, txHash, logIndex)
	return args.Bool(0), args.Error(1)
}

// MockAssetStore implements AssetStore interface
type MockAssetStore struct {
	mock.Mock
}

func (m *MockAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	args := m.Called(asset)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	args := m.Called(blockchainID, tokenAddress)
	return args.Get(0).(uint8), args.Error(1)
}

func (m *MockAssetStore) GetTokenAddress(asset string, blockchainID uint64) (string, error) {
	args := m.Called(asset, blockchainID)
	return args.String(0), args.Error(1)
}

func (m *MockAssetStore) GetTokenAsset(blockchainID uint64, tokenAddress string) (string, error) {
	args := m.Called(blockchainID, tokenAddress)
	return args.String(0), args.Error(1)
}
