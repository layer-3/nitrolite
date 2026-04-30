package sdk

import (
	"context"
	"fmt"
	"strconv"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// ============================================================================
// Node Information Methods
// ============================================================================

// Ping checks connectivity to the nitronode server.
// This is useful for health checks and verifying the connection is active.
//
// Example:
//
//	if err := client.Ping(ctx); err != nil {
//	    log.Printf("Server is unreachable: %v", err)
//	}
func (c *Client) Ping(ctx context.Context) error {
	if err := c.rpcClient.NodeV1Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// GetConfig retrieves the nitronode configuration including node identity and supported blockchains.
//
// Returns:
//   - NodeConfig containing the node address, version, and list of supported blockchain networks
//   - Error if the request fails
//
// Example:
//
//	config, err := client.GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Node: %s (v%s)\n", config.NodeAddress, config.NodeVersion)
func (c *Client) GetConfig(ctx context.Context) (*core.NodeConfig, error) {
	resp, err := c.rpcClient.NodeV1GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return transformNodeConfig(resp)
}

// GetBlockchains retrieves the list of supported blockchain networks.
// This is a convenience method that calls GetConfig and extracts the blockchains list.
//
// Returns:
//   - Slice of Blockchain containing name, chain ID, and contract address for each network
//   - Error if the request fails
//
// Example:
//
//	blockchains, err := client.GetBlockchains(ctx)
//	for _, bc := range blockchains {
//	    fmt.Printf("%s: %s\n", bc.Name, bc.ChannelHubAddress)
//	}
func (c *Client) GetBlockchains(ctx context.Context) ([]core.Blockchain, error) {
	config, err := c.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchains: %w", err)
	}
	return config.Blockchains, nil
}

// GetAssets retrieves all supported assets with optional blockchain filter.
//
// Parameters:
//   - blockchainID: Optional blockchain ID to filter assets (pass nil for all assets)
//
// Returns:
//   - Slice of Asset containing asset information and token implementations
//   - Error if the request fails
//
// Example:
//
//	assets, err := client.GetAssets(ctx, nil)
//	for _, asset := range assets {
//	    fmt.Printf("%s (%s): %d tokens\n", asset.Name, asset.Symbol, len(asset.Tokens))
//	}
func (c *Client) GetAssets(ctx context.Context, blockchainID *uint64) ([]core.Asset, error) {
	req := rpc.NodeV1GetAssetsRequest{}
	if blockchainID != nil {
		blockchainIDStr := strconv.FormatUint(*blockchainID, 10)
		req.BlockchainID = &blockchainIDStr
	}
	resp, err := c.rpcClient.NodeV1GetAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	return transformAssets(resp.Assets)
}
