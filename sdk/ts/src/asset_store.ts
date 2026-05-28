/**
 * Asset Store for caching asset information from the Nitrolite API
 */

import { Address } from 'viem';
import * as core from './core/types.js';

/**
 * ClientAssetStore implements asset caching by fetching data from the Nitrolite API.
 */
export class ClientAssetStore {
  private cache: Map<string, core.Asset>;
  private getAssetsFn: () => Promise<core.Asset[]>;

  constructor(getAssetsFn: () => Promise<core.Asset[]>) {
    this.cache = new Map();
    this.getAssetsFn = getAssetsFn;
  }

  /**
   * GetAssetDecimals returns the decimals for an asset as stored in Nitrolite.
   */
  async getAssetDecimals(asset: string): Promise<number> {
    // Check cache first
    const cached = this.cache.get(asset.toLowerCase());
    if (cached) {
      return cached.decimals;
    }

    // Fetch from node
    const assets = await this.getAssetsFn();

    // Update cache and find asset
    for (const a of assets) {
      this.cache.set(a.symbol.toLowerCase(), a);
      if (a.symbol.toLowerCase() === asset.toLowerCase()) {
        return a.decimals;
      }
    }

    throw new Error(`asset ${asset} not found`);
  }

  /**
   * GetTokenDecimals returns the decimals for a specific token on a blockchain.
   */
  async getTokenDecimals(blockchainId: bigint, tokenAddress: string): Promise<number> {
    // Fetch all assets if cache is empty
    if (this.cache.size === 0) {
      const assets = await this.getAssetsFn();
      for (const a of assets) {
        this.cache.set(a.symbol.toLowerCase(), a);
      }
    }

    // Search through all assets for matching token
    const tokenAddrLower = tokenAddress.toLowerCase();
    for (const asset of this.cache.values()) {
      for (const token of asset.tokens) {
        if (
          token.blockchainId === blockchainId &&
          token.address.toLowerCase() === tokenAddrLower
        ) {
          return token.decimals;
        }
      }
    }

    throw new Error(`token ${tokenAddress} on blockchain ${blockchainId} not found`);
  }

  /**
   * GetTokenAddress returns the token address for a given asset on a specific blockchain.
   */
  async getTokenAddress(asset: string, blockchainId: bigint): Promise<Address> {
    // Fetch all assets if cache is empty
    if (this.cache.size === 0) {
      const assets = await this.getAssetsFn();
      for (const a of assets) {
        this.cache.set(a.symbol.toLowerCase(), a);
      }
    }

    // Search for the asset and its token on the specified blockchain
    const assetLower = asset.toLowerCase();
    for (const a of this.cache.values()) {
      if (a.symbol.toLowerCase() === assetLower) {
        for (const token of a.tokens) {
          if (token.blockchainId === blockchainId) {
            return token.address;
          }
        }
        throw new Error(`asset ${asset} not available on blockchain ${blockchainId}`);
      }
    }

    // Asset not found in cache, try fetching again
    const assets = await this.getAssetsFn();
    for (const a of assets) {
      this.cache.set(a.symbol.toLowerCase(), a);
      if (a.symbol.toLowerCase() === assetLower) {
        for (const token of a.tokens) {
          if (token.blockchainId === blockchainId) {
            return token.address;
          }
        }
        throw new Error(`asset ${asset} not available on blockchain ${blockchainId}`);
      }
    }

    throw new Error(`asset ${asset} not found`);
  }

  /**
   * AssetExistsOnBlockchain checks if a specific asset is supported on a specific blockchain.
   */
  async assetExistsOnBlockchain(blockchainId: bigint, asset: string): Promise<boolean> {
    const assetLower = asset.toLowerCase();

    for (const a of this.cache.values()) {
      if (a.symbol.toLowerCase() === assetLower) {
        for (const token of a.tokens) {
          if (token.blockchainId === blockchainId) {
            return true;
          }
        }
        // Asset found in cache, but not on this chain
        return false;
      }
    }

    // Not in cache, fetch from API
    const assets = await this.getAssetsFn();
    for (const a of assets) {
      this.cache.set(a.symbol.toLowerCase(), a);
      if (a.symbol.toLowerCase() === assetLower) {
        for (const token of a.tokens) {
          if (token.blockchainId === blockchainId) {
            return true;
          }
        }
        // Asset found after fetch, but not on this chain
        return false;
      }
    }

    // Asset symbol not found at all
    return false;
  }

  /**
   * GetSuggestedBlockchainId returns the suggested blockchain ID for a given asset.
   */
  async getSuggestedBlockchainId(asset: string): Promise<bigint> {
    const key = asset.toLowerCase();

    // Check cache first
    const cached = this.cache.get(key);
    if (cached) {
      if (cached.suggestedBlockchainId === 0n) {
        throw new Error(`no suggested blockchain ID for asset ${asset}`);
      }
      return cached.suggestedBlockchainId;
    }

    // Not in cache, fetch from API
    const assets = await this.getAssetsFn();
    for (const a of assets) {
      this.cache.set(a.symbol.toLowerCase(), a);
    }

    const fetched = this.cache.get(key);
    if (fetched) {
      if (fetched.suggestedBlockchainId === 0n) {
        throw new Error(`no suggested blockchain ID for asset ${asset}`);
      }
      return fetched.suggestedBlockchainId;
    }

    throw new Error(`asset ${asset} not found`);
  }

  /**
   * Clear the cache
   */
  clearCache(): void {
    this.cache.clear();
  }
}
