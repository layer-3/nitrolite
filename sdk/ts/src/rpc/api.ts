/**
 * RPC API request and response definitions
 * This file implements the API request and response definitions
 * with versioned types organized by functional groups
 */

import { Address } from 'viem';
import {
  ChannelV1,
  ChannelDefinitionV1,
  ChannelSessionKeyStateV1,
  StateV1,
  BalanceEntryV1,
  TransactionV1,
  PaginationParamsV1,
  PaginationMetadataV1,
  AssetV1,
  BlockchainInfoV1,
  AppV1,
  AppInfoV1,
} from './types';
import {
  AppDefinitionV1,
  AppStateUpdateV1,
  AppSessionInfoV1,
  AppAllocationV1,
  AppSessionKeyStateV1,
  SignedAppStateUpdateV1,
} from '../app/types';
import { TransactionType } from '../core/types';

// ============================================================================
// Channels Group - V1 API
// ============================================================================

export interface ChannelsV1GetHomeChannelRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
}

export interface ChannelsV1GetHomeChannelResponse {
  /** On-chain channel information */
  channel: ChannelV1;
}

export interface ChannelsV1GetEscrowChannelRequest {
  /** Escrow channel ID */
  escrow_channel_id: string;
}

export interface ChannelsV1GetEscrowChannelResponse {
  /** On-chain channel information */
  channel: ChannelV1;
}

export interface ChannelsV1GetChannelsRequest {
  /** User's wallet address */
  wallet: Address;
  /** Status filter */
  status?: string;
  /** Asset filter */
  asset?: string;
  /** Channel type filter ("home" or "escrow") */
  channel_type?: string;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface ChannelsV1GetChannelsResponse {
  /** List of channels */
  channels: ChannelV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface ChannelsV1GetLatestStateRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
  /** Enable to get the latest signed state */
  only_signed: boolean;
}

export interface ChannelsV1GetLatestStateResponse {
  /** Current state of the user */
  state: StateV1;
}

export interface ChannelsV1GetStatesRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
  /** User epoch index filter */
  epoch?: bigint; // uint64
  /** Home/Escrow Channel ID filter */
  channel_id?: string;
  /** Return only signed states */
  only_signed: boolean;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface ChannelsV1GetStatesResponse {
  /** List of states */
  states: StateV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface ChannelsV1RequestCreationRequest {
  /** State to be submitted */
  state: StateV1;
  /** Definition of the channel to be created */
  channel_definition: ChannelDefinitionV1;
}

export interface ChannelsV1RequestCreationResponse {
  /** Node's signature for the state */
  signature: string;
}

export interface ChannelsV1SubmitStateRequest {
  /** State to be submitted */
  state: StateV1;
}

export interface ChannelsV1SubmitStateResponse {
  /** Node's signature for the state */
  signature: string;
}

export interface ChannelsV1HomeChannelCreatedEvent {
  /** Created home channel information */
  channel: ChannelV1;
  /** Initial state of the home channel */
  initial_state: StateV1;
}

// ============================================================================
// Channel Session Key State Group - V1 API
// ============================================================================

export interface ChannelsV1SubmitSessionKeyStateRequest {
  /** Session key state containing delegation information */
  state: ChannelSessionKeyStateV1;
}

export interface ChannelsV1SubmitSessionKeyStateResponse {}

export interface ChannelsV1GetLastKeyStatesRequest {
  /** User's wallet address */
  user_address: string;
  /** Optionally filter by session key address */
  session_key?: string;
}

export interface ChannelsV1GetLastKeyStatesResponse {
  /** List of active channel session key states for the user */
  states: ChannelSessionKeyStateV1[];
}

// ============================================================================
// App Sessions Group - V1 API
// ============================================================================

export interface AppSessionsV1SubmitDepositStateRequest {
  /** Application session state update to be submitted */
  app_state_update: AppStateUpdateV1;
  /** List of participant signatures for the app state update */
  quorum_sigs: string[];
  /** User state */
  user_state: StateV1;
}

export interface AppSessionsV1SubmitDepositStateResponse {
  /** Node's signature for the deposit state */
  signature: string;
}

export interface AppSessionsV1SubmitAppStateRequest {
  /** Application session state update to be submitted */
  app_state_update: AppStateUpdateV1;
  /** Signature quorum for the application session */
  quorum_sigs: string[];
}

export interface AppSessionsV1SubmitAppStateResponse {}

export interface AppSessionsV1RebalanceAppSessionsRequest {
  /** List of signed application session state updates */
  signed_updates: SignedAppStateUpdateV1[];
}

export interface AppSessionsV1RebalanceAppSessionsResponse {
  /** Unique identifier for this rebalancing operation */
  batch_id: string;
}

export interface AppSessionsV1GetAppDefinitionRequest {
  /** Application session ID */
  app_session_id: string;
}

export interface AppSessionsV1GetAppDefinitionResponse {
  /** Application definition */
  definition: AppDefinitionV1;
}

export interface AppSessionsV1GetAppSessionsRequest {
  /** Application session ID filter */
  app_session_id?: string;
  /** Participant wallet address filter */
  participant?: Address;
  /** Status filter (open/closed) */
  status?: string;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface AppSessionsV1GetAppSessionsResponse {
  /** List of application sessions */
  app_sessions: AppSessionInfoV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface AppSessionsV1CreateAppSessionRequest {
  /** Application definition including participants and quorum */
  definition: AppDefinitionV1;
  /** Optional JSON stringified session data */
  session_data: string;
  /** Participant signatures for the app session creation */
  quorum_sigs?: string[];
}

export interface AppSessionsV1CreateAppSessionResponse {
  /** Created application session ID */
  app_session_id: string;
  /** Initial version of the session */
  version: string;
  /** Status of the session */
  status: string;
}

export interface AppSessionsV1CloseAppSessionRequest {
  /** Application session ID to close */
  app_session_id: string;
  /** Final asset allocations when closing the session */
  allocations: AppAllocationV1[];
  /** Optional final JSON stringified session data */
  session_data?: string;
}

export interface AppSessionsV1CloseAppSessionResponse {
  /** Closed application session ID */
  app_session_id: string;
  /** Final version of the session */
  version: string;
  /** Status of the session (closed) */
  status: string;
}

// ============================================================================
// App Session Key State Group - V1 API
// ============================================================================

export interface AppSessionsV1SubmitSessionKeyStateRequest {
  /** Session key state containing delegation information */
  state: AppSessionKeyStateV1;
}

export interface AppSessionsV1SubmitSessionKeyStateResponse {}

export interface AppSessionsV1GetLastKeyStatesRequest {
  /** User's wallet address */
  user_address: string;
  /** Optionally filter by session key address */
  session_key?: string;
}

export interface AppSessionsV1GetLastKeyStatesResponse {
  /** List of active session key states for the user */
  states: AppSessionKeyStateV1[];
}

// ============================================================================
// Apps Group - V1 API
// ============================================================================

export interface AppsV1GetAppsRequest {
  /** Application ID filter */
  app_id?: string;
  /** Owner wallet address filter */
  owner_wallet?: string;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface AppsV1GetAppsResponse {
  /** List of registered applications */
  apps: AppInfoV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface AppsV1SubmitAppVersionRequest {
  /** Application definition */
  app: AppV1;
  /** Owner's signature over the packed app data */
  owner_sig: string;
}

export interface AppsV1SubmitAppVersionResponse {}

// ============================================================================
// User Group - V1 API
// ============================================================================

export interface UserV1GetBalancesRequest {
  /** User's wallet address */
  wallet: Address;
}

export interface UserV1GetBalancesResponse {
  /** List of asset balances */
  balances: BalanceEntryV1[];
}

export interface UserV1GetTransactionsRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol filter */
  asset?: string;
  /** Transaction type filter */
  tx_type?: TransactionType;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
  /** Start time filter (Unix timestamp) */
  from_time?: bigint; // uint64
  /** End time filter (Unix timestamp) */
  to_time?: bigint; // uint64
}

export interface UserV1GetTransactionsResponse {
  /** List of transactions */
  transactions: TransactionV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

// ============================================================================
// Node Group - V1 API
// ============================================================================

export interface NodeV1PingRequest {}

export interface NodeV1PingResponse {}

export interface NodeV1GetConfigRequest {}

export interface NodeV1GetConfigResponse {
  /** Node wallet address */
  node_address: Address;
  /** Node software version */
  node_version: string;
  /** List of supported signature validators identifiers for state sig verification */
  supported_sig_validators: number[];
  /** List of supported networks */
  blockchains: BlockchainInfoV1[];
}

export interface NodeV1GetAssetsRequest {
  /** Blockchain network ID filter */
  blockchain_id?: bigint; // uint64
}

export interface NodeV1GetAssetsResponse {
  /** List of supported assets */
  assets: AssetV1[];
}
