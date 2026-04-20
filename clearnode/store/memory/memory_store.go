package memory

import (
	"fmt"
	"slices"
	"strings"

	"github.com/layer-3/nitrolite/pkg/core"
)

type MemoryStoreV1 struct {
	blockchains          []core.Blockchain
	assets               []core.Asset
	channelSigValidators map[uint64]map[uint8]string      // map[blockchain_id]map[validator_id]validator_address
	supportedAssets      map[string]map[uint64]string     // map[asset_symbol]map[blockchain_id]string
	tokenAssets          map[uint64]map[string]core.Asset // map[blockchain_id]map[token_address]asset
	tokenDecimals        map[uint64]map[string]uint8      // map[blockchain_id]map[token_address]decimals
	assetDecimals        map[string]uint8                 // map[asset_symbol]decimals
}

func NewMemoryStoreV1(assetsConfig AssetsConfig, blockchainsConfig map[uint64]BlockchainConfig) (MemoryStore, error) {
	supportedBlockchainIDs := make(map[uint64]struct{})
	blockchains := make([]core.Blockchain, 0, len(blockchainsConfig))
	channelSigValidators := make(map[uint64]map[uint8]string)
	for _, bc := range blockchainsConfig {
		if bc.Disabled {
			continue
		}

		if bc.ChannelHubAddress != "" {
			supportedBlockchainIDs[bc.ID] = struct{}{}
			channelSigValidators[bc.ID] = bc.ChannelHubSigValidators
		}

		blockchains = append(blockchains, core.Blockchain{
			ID:                     bc.ID,
			Name:                   bc.Name,
			ChannelHubAddress:      bc.ChannelHubAddress,
			LockingContractAddress: bc.LockingContractAddress,
			BlockStep:              bc.BlockStep,
		})
	}
	slices.SortFunc(blockchains, func(a, b core.Blockchain) int {
		if a.ID < b.ID {
			return -1
		} else if a.ID > b.ID {
			return 1
		}
		return 0
	})

	supportedAssets := make(map[string]map[uint64]string)
	tokenAssets := make(map[uint64]map[string]core.Asset)
	tokenDecimals := make(map[uint64]map[string]uint8)
	assetDecimals := make(map[string]uint8)
	assets := make([]core.Asset, 0, len(assetsConfig.Assets))
	for _, asset := range assetsConfig.Assets {
		if asset.Disabled {
			continue
		}

		var suggestedBlockchainID uint64
		tokens := make([]core.Token, 0, len(asset.Tokens))
		for _, token := range asset.Tokens {
			if _, ok := supportedBlockchainIDs[token.BlockchainID]; !ok {
				continue
			}
			if token.Disabled {
				continue
			}

			tokenAddress := strings.ToLower(token.Address)
			tokens = append(tokens, core.Token{
				Name:         token.Name,
				Symbol:       token.Symbol,
				Address:      tokenAddress,
				BlockchainID: token.BlockchainID,
				Decimals:     token.Decimals,
			})

			if _, ok := supportedAssets[asset.Symbol]; !ok {
				supportedAssets[asset.Symbol] = make(map[uint64]string)
			}
			supportedAssets[asset.Symbol][token.BlockchainID] = tokenAddress

			if _, ok := tokenDecimals[token.BlockchainID]; !ok {
				tokenDecimals[token.BlockchainID] = make(map[string]uint8)
			}
			tokenDecimals[token.BlockchainID][tokenAddress] = token.Decimals

			if _, ok := tokenAssets[token.BlockchainID]; !ok {
				tokenAssets[token.BlockchainID] = make(map[string]core.Asset)
			}
			tokenAssets[token.BlockchainID][tokenAddress] = core.Asset{
				Symbol:   asset.Symbol,
				Name:     asset.Name,
				Decimals: asset.Decimals,
			}

			if asset.SuggestedBlockchainID == token.BlockchainID {
				suggestedBlockchainID = token.BlockchainID
			}
		}
		if len(tokens) == 0 {
			continue
		}
		if suggestedBlockchainID == 0 {
			return nil, fmt.Errorf("asset '%s' does not have a valid suggested blockchain ID", asset.Symbol)
		}

		slices.SortFunc(tokens, func(a, b core.Token) int {
			if a.BlockchainID < b.BlockchainID {
				return -1
			} else if a.BlockchainID > b.BlockchainID {
				return 1
			}
			return 0
		})

		assets = append(assets, core.Asset{
			Symbol:                asset.Symbol,
			Name:                  asset.Name,
			Decimals:              asset.Decimals,
			SuggestedBlockchainID: suggestedBlockchainID,
			Tokens:                tokens,
		})

		assetDecimals[asset.Symbol] = asset.Decimals
	}

	slices.SortFunc(assets, func(a, b core.Asset) int {
		if a.Symbol < b.Symbol {
			return -1
		} else if a.Symbol > b.Symbol {
			return 1
		}
		return 0
	})

	return &MemoryStoreV1{
		blockchains:          blockchains,
		assets:               assets,
		channelSigValidators: channelSigValidators,
		supportedAssets:      supportedAssets,
		tokenAssets:          tokenAssets,
		tokenDecimals:        tokenDecimals,
		assetDecimals:        assetDecimals,
	}, nil
}

func NewMemoryStoreV1FromConfig(configDirPath string) (MemoryStore, error) {
	blockchainConfig, err := LoadEnabledBlockchains(configDirPath)
	if err != nil {
		return nil, err
	}
	assetsConfig, err := LoadAssets(configDirPath)
	if err != nil {
		return nil, err
	}
	return NewMemoryStoreV1(assetsConfig, blockchainConfig)
}

// GetBlockchains retrieves the list of supported blockchains.
func (ms *MemoryStoreV1) GetBlockchains() ([]core.Blockchain, error) {
	return ms.blockchains, nil
}

// GetAssets retrieves the list of supported assets.
// If blockchainID is provided, filters assets to only include tokens on that blockchain.
func (ms *MemoryStoreV1) GetAssets(blockchainID *uint64) ([]core.Asset, error) {
	if blockchainID == nil {
		return ms.assets, nil
	}

	filteredAssets := make([]core.Asset, 0)
	for _, asset := range ms.assets {
		filteredTokens := make([]core.Token, 0)
		for _, token := range asset.Tokens {
			if token.BlockchainID == *blockchainID {
				filteredTokens = append(filteredTokens, token)
			}
		}

		if len(filteredTokens) > 0 {
			filteredAsset := asset
			filteredAsset.Tokens = filteredTokens
			filteredAssets = append(filteredAssets, filteredAsset)
		}
	}

	return filteredAssets, nil
}

func (ms *MemoryStoreV1) GetChannelSigValidators(blockchainID uint64) (map[uint8]string, error) {
	channelSigValidators, ok := ms.channelSigValidators[blockchainID]
	if !ok {
		return nil, fmt.Errorf("blockchain with ID '%d' is not supported", blockchainID)
	}
	return channelSigValidators, nil
}

func (ms *MemoryStoreV1) GetTokenAddress(asset string, blockchainID uint64) (string, error) {
	tokensOnchain, ok := ms.supportedAssets[asset]
	if !ok {
		return "", fmt.Errorf("asset '%s' is not supported", asset)
	}
	tokenAddress, ok := tokensOnchain[blockchainID]
	if !ok {
		return "", fmt.Errorf("asset '%s' is not supported on blockchain with ID '%d'", asset, blockchainID)
	}
	return tokenAddress, nil
}

// IsAssetSupported checks if a given asset (token) is supported on the specified blockchain.
func (ms *MemoryStoreV1) IsAssetSupported(asset, tokenAddress string, blockchainID uint64) (bool, error) {
	tokensOnchain, ok := ms.supportedAssets[asset]
	if !ok {
		return false, nil
	}
	_tokenAddress, ok := tokensOnchain[blockchainID]
	if !ok {
		return false, nil
	}
	return strings.EqualFold(tokenAddress, _tokenAddress), nil
}

// GetAssetDecimals checks if an asset exists and returns its decimals in YN
func (ms *MemoryStoreV1) GetAssetDecimals(asset string) (uint8, error) {
	decimals, ok := ms.assetDecimals[asset]
	if !ok {
		return 0, fmt.Errorf("asset '%s' is not supported", asset)
	}
	return decimals, nil
}

// GetTokenDecimals returns the decimals for a token on a specific blockchain
func (ms *MemoryStoreV1) GetTokenDecimals(blockchainID uint64, tokenAddress string) (uint8, error) {
	tokenAddress = strings.ToLower(tokenAddress)

	decimalsOnChain, ok := ms.tokenDecimals[blockchainID]
	if !ok {
		return 0, fmt.Errorf("blockchain with ID '%d' is not supported", blockchainID)
	}
	decimals, ok := decimalsOnChain[tokenAddress]
	if !ok {
		return 0, fmt.Errorf("token %s is not supported on blockchain with ID '%d'", tokenAddress, blockchainID)
	}
	return decimals, nil
}

// GetTokenAsset returns the asset for a token on a specific blockchain
func (ms *MemoryStoreV1) GetTokenAsset(blockchainID uint64, tokenAddress string) (string, error) {
	tokenAddress = strings.ToLower(tokenAddress)

	assetsOnChain, ok := ms.tokenAssets[blockchainID]
	if !ok {
		return "", fmt.Errorf("blockchain with ID '%d' is not supported", blockchainID)
	}
	asset, ok := assetsOnChain[tokenAddress]
	if !ok {
		return "", fmt.Errorf("token %s is not supported on blockchain with ID '%d'", tokenAddress, blockchainID)
	}
	return asset.Symbol, nil
}
