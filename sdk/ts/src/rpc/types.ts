/**
 * RPC API types for the Nitrolite Node service
 * This file contains common types and structs shared across V1 API groups
 *
 * NOTE: All field names use snake_case to match the actual API response format
 */

import { Address } from 'viem';
import { TransitionType, TransactionType } from '../core/types.js';

// ============================================================================
// Channel Types
// ============================================================================

/**
 * ChannelV1 represents an on-chain channel
 */
export interface ChannelV1 {
  /** Unique identifier for the channel */
  channel_id: string;
  /** User wallet address */
  user_wallet: Address;
  /** Asset symbol (e.g. usdc, eth) */
  asset: string;
  /** Type of the channel (home, escrow) */
  type: string;
  /** Unique identifier for the blockchain */
  blockchain_id: string; // uint64 as string
  /** Address of the token used in the channel */
  token_address: Address;
  /** Challenge period for the channel in seconds */
  challenge_duration: number; // uint32
  /** Timestamp when challenge expires */
  challenge_expires_at?: string;
  /** Nonce for the channel */
  nonce: string; // uint64 as string
  /** Hex string bitmap of approved signature validators */
  approved_sig_validators: string;
  /** Current status of the channel (void, open, challenged, closed) */
  status: string;
  /** On-chain state version of the channel */
  state_version: string;
}

/**
 * ChannelDefinitionV1 represents the configuration for creating a channel
 */
export interface ChannelDefinitionV1 {
  /** Unique number to prevent replay attacks */
  nonce: string; // uint64 as string to preserve precision
  /** Challenge period for the channel in seconds */
  challenge: number; // uint32
  /** Hex string bitmap representing the approved signature validators for the channel */
  approved_sig_validators: string;
}

// ============================================================================
// State Types
// ============================================================================

/**
 * TransitionV1 represents a state transition
 */
export interface TransitionV1 {
  /** Type of state transition */
  type: TransitionType;
  /** Transaction ID associated with the transition */
  tx_id: string;
  /** Account identifier (varies based on transition type, may be empty string) */
  account_id: string;
  /** Amount involved in the transition */
  amount: string;
}

/**
 * LedgerV1 represents ledger balances for a channel
 */
export interface LedgerV1 {
  /** Address of the token used in this channel */
  token_address: Address;
  /** Unique identifier for the blockchain */
  blockchain_id: string; // uint64 as string
  /** User balance in the channel */
  user_balance: string;
  /** User net flow in the channel */
  user_net_flow: string;
  /** Node balance in the channel */
  node_balance: string;
  /** Node net flow in the channel */
  node_net_flow: string;
}

/**
 * StateV1 represents the current state of the user stored on Node
 */
export interface StateV1 {
  /** Deterministic ID (hash) of the state */
  id: string;
  /** Transition included in the state */
  transition: TransitionV1;
  /** Asset type of the state */
  asset: string;
  /** User wallet address */
  user_wallet: Address;
  /** User Epoch Index */
  epoch: string;
  /** Version of the state */
  version: string;
  /** Identifier for the home Channel ID */
  home_channel_id?: string;
  /** Identifier for the escrow Channel ID */
  escrow_channel_id?: string;
  /** User and node balances for the home channel */
  home_ledger: LedgerV1;
  /** User and node balances for the escrow channel */
  escrow_ledger?: LedgerV1;
  /** User signature for the state */
  user_sig?: string;
  /** Node signature for the state */
  node_sig?: string;
}

// ============================================================================
// App Session Types
// ============================================================================
// Note: App session types are exported from the app module

// ============================================================================
// Channel Session Key Types
// ============================================================================

/**
 * ChannelSessionKeyStateV1 represents the state of a channel session key delegation
 */
export interface ChannelSessionKeyStateV1 {
  /** User wallet address */
  user_address: string;
  /** Session key address for delegation */
  session_key: string;
  /** Version of the session key state */
  version: string;
  /** Assets associated with this session key */
  assets: string[];
  /** Unix timestamp in seconds indicating when the session key expires */
  expires_at: string;
  /** User's signature over the session key metadata */
  user_sig: string;
}

// ============================================================================
// App Registry Types
// ============================================================================

/**
 * AppV1 represents a registered application definition (without timestamps)
 */
export interface AppV1 {
  /** Application identifier */
  id: string;
  /** Owner's wallet address */
  owner_wallet: string;
  /** Application metadata */
  metadata: string;
  /** Current version */
  version: string;
  /** Whether sessions can be created without owner approval */
  creation_approval_not_required: boolean;
}

/**
 * AppInfoV1 represents full application info including timestamps
 */
export interface AppInfoV1 extends AppV1 {
  /** Creation timestamp (unix seconds) */
  created_at: string;
  /** Last update timestamp (unix seconds) */
  updated_at: string;
}

// ============================================================================
// Asset and Blockchain Types
// ============================================================================

/**
 * AssetV1 represents information about a supported asset
 */
export interface AssetV1 {
  /** Asset name */
  name: string;
  /** Asset symbol */
  symbol: string;
  /** Number of decimal places for the asset */
  decimals: number; // uint8
  /** Suggested blockchain network ID for this asset */
  suggested_blockchain_id: string; // uint64 as string
  /** List of supported tokens for the asset */
  tokens: TokenV1[];
}

/**
 * TokenV1 represents information about a supported token
 */
export interface TokenV1 {
  /** Token name */
  name: string;
  /** Token symbol */
  symbol: string;
  /** Token contract address */
  address: Address;
  /** Blockchain network ID */
  blockchain_id: string; // uint64 as string
  /** Number of decimal places */
  decimals: number; // uint8
}

/**
 * BlockchainInfoV1 represents information about a supported network
 */
export interface BlockchainInfoV1 {
  /** Blockchain name */
  name: string;
  /** Blockchain network ID */
  blockchain_id: string; // uint64 as string
  /** Channel hub contract address on this network */
  channel_hub_address: Address;
  /** Locking contract address on this network */
  locking_contract_address?: Address;
}

// ============================================================================
// Balance and Transaction Types
// ============================================================================

/**
 * BalanceEntryV1 represents a balance for a specific asset
 */
export interface BalanceEntryV1 {
  /** Asset symbol */
  asset: string;
  /** Balance amount */
  amount: string;
  /** On-chain enforced balance */
  enforced: string;
}

/**
 * TransactionV1 represents a transaction record
 */
export interface TransactionV1 {
  /** Unique transaction reference */
  id: string;
  /** Asset symbol */
  asset: string;
  /** Transaction type */
  tx_type: TransactionType;
  /** Account that sent the funds */
  from_account: Address;
  /** Account that received the funds */
  to_account: Address;
  /** ID of the new sender's channel state */
  sender_new_state_id?: string;
  /** ID of the new receiver's channel state */
  receiver_new_state_id?: string;
  /** Transaction amount */
  amount: string;
  /** When the transaction was created */
  created_at: string;
}

// ============================================================================
// Pagination Types
// ============================================================================

/**
 * PaginationParamsV1 represents pagination request parameters
 */
export interface PaginationParamsV1 {
  /** Pagination offset (number of items to skip) */
  offset?: number; // uint32
  /** Number of items to return */
  limit?: number; // uint32
}

// ============================================================================
// Action Allowance Types
// ============================================================================

/**
 * ActionAllowanceV1 represents the allowance information for a specific gated action
 */
export interface ActionAllowanceV1 {
  /** The specific action being gated (transfer, app_session_deposit, app_session_operation, app_session_withdrawal) */
  gated_action: string;
  /** Time window for which the allowance is valid (e.g. "24h0m0s") */
  time_window: string;
  /** Total allowance for the action within the time window */
  allowance: string;
  /** Amount already used within the time window */
  used: string;
}

/**
 * PaginationMetadataV1 represents pagination information
 */
export interface PaginationMetadataV1 {
  /** Current page number */
  page: number; // uint32
  /** Number of items per page */
  per_page: number; // uint32
  /** Total number of items */
  total_count: number; // uint32
  /** Total number of pages */
  page_count: number; // uint32
}
