package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/layer-3/nitrolite/pkg/core"
)

// TODO: refactor by not using client as a dependency as this creates a circular usage.
// clientAssetStore implements core.AssetStore by fetching data from the Clearnode API.
type clientAssetStore struct {
	client *Client
	cache  map[string]core.Asset // lowercase asset symbol -> Asset
}

func newClientAssetStore(client *Client) *clientAssetStore {
	return &clientAssetStore{
		client: client,
		cache:  make(map[string]core.Asset),
	}
}

// populateCache fetches all assets from the node and populates the cache.
func (s *clientAssetStore) populateCache() error {
	assets, err := s.client.GetAssets(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to fetch assets: %w", err)
	}
	for _, a := range assets {
		for _, token := range a.Tokens {
			if a.Decimals > token.Decimals {
				return fmt.Errorf("asset %s decimals (%d) must not exceed token %s decimals (%d) on blockchain %d",
					a.Symbol, a.Decimals, token.Symbol, token.Decimals, token.BlockchainID)
			}
		}
		s.cache[strings.ToLower(a.Symbol)] = a
	}
	return nil
}

// GetAssetDecimals returns the decimals for an asset as stored in Clearnode.
func (s *clientAssetStore) GetAssetDecimals(asset string) (uint8, error) {
	key := strings.ToLower(asset)

	// Check cache first
	if cached, ok := s.cache[key]; ok {
		return cached.Decimals, nil
	}

	// Fetch from node
	if err := s.populateCache(); err != nil {
		return 0, err
	}

	if cached, ok := s.cache[key]; ok {
		return cached.Decimals, nil
	}

	return 0, fmt.Errorf("asset %s not found", asset)
}

// GetTokenDecimals returns the decimals for a specific token on a blockchain.
func (s *clientAssetStore) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		if err := s.populateCache(); err != nil {
			return 0, err
		}
	}

	// Search through all assets for matching token
	tokenAddressLower := strings.ToLower(tokenAddress)
	for _, asset := range s.cache {
		for _, token := range asset.Tokens {
			if token.BlockchainID == blockchainID &&
				strings.EqualFold(token.Address, tokenAddressLower) {
				return token.Decimals, nil
			}
		}
	}

	return 0, fmt.Errorf("token %s on blockchain %d not found", tokenAddress, blockchainID)
}

// GetTokenAsset returns the asset for a specific token on a blockchain.
func (s *clientAssetStore) GetTokenAsset(blockchainID uint64, tokenAddress string) (string, error) {
	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		if err := s.populateCache(); err != nil {
			return "", err
		}
	}

	// Search through all assets for matching token
	tokenAddressLower := strings.ToLower(tokenAddress)
	for _, asset := range s.cache {
		for _, token := range asset.Tokens {
			if token.BlockchainID == blockchainID &&
				strings.EqualFold(token.Address, tokenAddressLower) {
				return asset.Symbol, nil
			}
		}
	}

	return "", fmt.Errorf("token %s on blockchain %d not found", tokenAddress, blockchainID)
}

// GetTokenAddress returns the token address for a given asset on a specific blockchain.
func (s *clientAssetStore) GetTokenAddress(asset string, blockchainID uint64) (string, error) {
	key := strings.ToLower(asset)

	// Fetch all assets if cache is empty
	if len(s.cache) == 0 {
		if err := s.populateCache(); err != nil {
			return "", err
		}
	}

	// Check cache by key
	if a, ok := s.cache[key]; ok {
		for _, token := range a.Tokens {
			if token.BlockchainID == blockchainID {
				return token.Address, nil
			}
		}
		return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
	}

	// Asset not found in cache, try fetching again
	if err := s.populateCache(); err != nil {
		return "", err
	}

	if a, ok := s.cache[key]; ok {
		for _, token := range a.Tokens {
			if token.BlockchainID == blockchainID {
				return token.Address, nil
			}
		}
		return "", fmt.Errorf("asset %s not available on blockchain %d", asset, blockchainID)
	}

	return "", fmt.Errorf("asset %s not found", asset)
}

// GetSuggestedBlockchainID returns the suggested blockchain ID for a given asset.
func (s *clientAssetStore) GetSuggestedBlockchainID(asset string) (uint64, error) {
	key := strings.ToLower(asset)

	// Check cache first
	if a, ok := s.cache[key]; ok {
		if a.SuggestedBlockchainID == 0 {
			return 0, fmt.Errorf("no suggested blockchain ID for asset %s", asset)
		}
		return a.SuggestedBlockchainID, nil
	}

	// Not in cache, fetch from API
	if err := s.populateCache(); err != nil {
		return 0, err
	}

	if a, ok := s.cache[key]; ok {
		if a.SuggestedBlockchainID == 0 {
			return 0, fmt.Errorf("no suggested blockchain ID for asset %s", asset)
		}
		return a.SuggestedBlockchainID, nil
	}

	return 0, fmt.Errorf("asset %s not found", asset)
}

// AssetExistsOnBlockchain checks if a specific asset is supported on a specific blockchain.
func (s *clientAssetStore) AssetExistsOnBlockchain(blockchainID uint64, asset string) (bool, error) {
	key := strings.ToLower(asset)

	if a, ok := s.cache[key]; ok {
		for _, token := range a.Tokens {
			if token.BlockchainID == blockchainID {
				return true, nil
			}
		}
		return false, nil
	}

	if err := s.populateCache(); err != nil {
		return false, err
	}

	if a, ok := s.cache[key]; ok {
		for _, token := range a.Tokens {
			if token.BlockchainID == blockchainID {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}
