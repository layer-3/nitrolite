package evm

import (
	"context"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type HandleEvent func(ctx context.Context, eventLog types.Log) error

// ContractEventGetter is used by Listener for resumption and deduplication.
type ContractEventGetter interface {
	// GetLatestContractEventBlockNumber returns the block to resume from (0 = start fresh).
	GetLatestContractEventBlockNumber(contractAddress string, blockchainID uint64) (lastBlock uint64, err error)
	// IsContractEventPresent checks whether a specific event was already processed.
	IsContractEventPresent(blockchainID, blockNumber uint64, txHash string, logIndex uint32) (isPresent bool, err error)
}

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)

	// GetTokenAddress returns the token address for a given asset on a specific blockchain
	GetTokenAddress(asset string, blockchainID uint64) (string, error)

	// GetTokenAsset returns the asset for a token on a specific blockchain
	GetTokenAsset(blockchainID uint64, tokenAddress string) (string, error)
}

type EVMClient interface {
	ethereum.ChainStateReader
	bind.ContractBackend
}
