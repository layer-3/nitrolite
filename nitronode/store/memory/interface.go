package memory

import (
	"github.com/layer-3/nitrolite/pkg/core"
)

// MemoryStore defines an in-memory data store interface for retrieving
// supported blockchains and assets.
type MemoryStore interface {
	// GetBlockchains retrieves the list of supported blockchains.
	GetBlockchains() ([]core.Blockchain, error)

	// GetAssets retrieves the list of supported assets.
	// If blockchainID is provided, filters assets to only include tokens on that blockchain.
	GetAssets(blockchainID *uint64) ([]core.Asset, error)

	// GetChannelSigValidators retrieves the channel signature validators for a specific blockchain.
	GetChannelSigValidators(blockchainID uint64) (map[uint8]string, error)

	// GetTokenAddress retrieves the token address for a given asset on a specific blockchain.
	GetTokenAddress(asset string, blockchainID uint64) (string, error)

	// IsAssetSupported checks if a given asset (token) is supported on the specified blockchain.
	IsAssetSupported(asset, tokenAddress string, blockchainID uint64) (bool, error)

	// GetAssetDecimals checks if an asset exists and returns its decimals in YN
	GetAssetDecimals(asset string) (uint8, error)

	// GetTokenDecimals returns the decimals for a token on a specific blockchain
	GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error)

	// GetTokenAsset returns the asset for a token on a specific blockchain
	GetTokenAsset(blockchainID uint64, tokenAddress string) (string, error)
}
