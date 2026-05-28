// ============================================================================
// Main SDK Exports
// ============================================================================

// Export SDK Client (main entry point)
export { Client, DEFAULT_CHALLENGE_PERIOD, type StateSigner, type TransactionSigner } from './client.js';

// Export signers
export {
  EthereumMsgSigner,
  EthereumRawSigner,
  ChannelDefaultSigner,
  ChannelSessionKeyStateSigner,
  AppSessionWalletSignerV1,
  AppSessionKeySignerV1,
  createSigners,
} from './signers.js';

// Export configuration
export {
  type Config,
  DefaultConfig,
  type Option,
  withHandshakeTimeout,
  withErrorHandler,
  withBlockchainRPC,
  withApplicationID,
  APPLICATION_ID_QUERY_PARAM,
  appendApplicationIDQueryParam
} from './config.js';

// Export asset store
export { ClientAssetStore } from './asset_store.js';

// Export utility functions
export * from './utils.js';

// ============================================================================
// Core Modules (types, state management, utilities)
// ============================================================================

export * from './core/index.js';
export * from './app/index.js';

// ============================================================================
// Blockchain Modules
// ============================================================================

export * from './blockchain/index.js';

// ============================================================================
// RPC Modules
// ============================================================================

export * from './rpc/index.js';
