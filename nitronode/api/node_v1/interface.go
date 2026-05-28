package node_v1

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
}
