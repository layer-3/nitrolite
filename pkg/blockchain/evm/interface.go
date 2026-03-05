package evm

import (
	"context"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/core"
)

type HandleEvent func(ctx context.Context, eventLog types.Log)
type StoreContractEvent func(ev core.BlockchainEvent) error
type LatestEventGetter func(contractAddress string, blockchainID uint64) (ev core.BlockchainEvent, err error)

type AssetStore interface {
	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)

	// GetTokenAddress returns the token address for a given asset on a specific blockchain
	GetTokenAddress(asset string, blockchainID uint64) (string, error)
}

type EVMClient interface {
	ethereum.ChainStateReader
	bind.ContractBackend
}
