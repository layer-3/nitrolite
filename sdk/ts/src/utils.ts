/**
 * Utility functions for type transformations and helper methods
 */

import * as core from './core/types.js';
import * as API from './rpc/api.js';
import { AssetV1, BalanceEntryV1, ChannelV1, LedgerV1, TransitionV1, StateV1, TransactionV1, PaginationMetadataV1, ActionAllowanceV1 } from './rpc/types.js';
import { Decimal } from 'decimal.js';
import { Address } from 'viem';

/**
 * Generate a nonce based on current timestamp with microsecond precision
 */
export function generateNonce(): bigint {
  return BigInt(Date.now()) * 1000000n + BigInt(Math.floor(Math.random() * 1000000));
}

// ============================================================================
// NodeConfig and Blockchain Transformations
// ============================================================================

/**
 * Decode supported_sig_validators from the server response.
 * Go's encoding/json serializes []uint8-based types as base64 strings,
 * so we handle both base64 strings and number arrays.
 */
function decodeSigValidators(raw: unknown): number[] {
  if (Array.isArray(raw)) return raw;
  if (typeof raw === 'string' && raw.length > 0) {
    const binary = atob(raw);
    return Array.from(binary, (c) => c.charCodeAt(0));
  }
  return [];
}

/**
 * Transform RPC NodeV1GetConfigResponse to core NodeConfig type
 */
export function transformNodeConfig(resp: API.NodeV1GetConfigResponse): core.NodeConfig {
  const blockchains: core.Blockchain[] = resp.blockchains.map((info) => ({
    name: info.name,
    id: BigInt(info.blockchain_id),
    channelHubAddress: info.channel_hub_address as Address,
    lockingContractAddress: info.locking_contract_address as Address | undefined,
    blockStep: 0n, // Not provided in RPC response
  }));

  return {
    nodeAddress: resp.node_address as Address,
    nodeVersion: resp.node_version,
    supportedSigValidators: decodeSigValidators(resp.supported_sig_validators),
    blockchains,
  };
}

// ============================================================================
// Asset and Token Transformations
// ============================================================================

/**
 * Transform RPC AssetV1 array to core Asset array
 */
export function transformAssets(assets: AssetV1[]): core.Asset[] {
  return assets.map((asset) => ({
    name: asset.name,
    symbol: asset.symbol,
    decimals: asset.decimals,
    suggestedBlockchainId: BigInt(asset.suggested_blockchain_id),
    tokens: asset.tokens.map((token) => ({
      name: token.name,
      symbol: token.symbol,
      address: token.address as Address,
      blockchainId: BigInt(token.blockchain_id),
      decimals: token.decimals,
    })),
  }));
}

// ============================================================================
// Balance Transformations
// ============================================================================

/**
 * Transform RPC BalanceEntryV1 array to core BalanceEntry array
 */
export function transformBalances(balances: BalanceEntryV1[]): core.BalanceEntry[] {
  return balances.map((balance) => ({
    asset: balance.asset,
    balance: new Decimal(balance.amount),
  }));
}

// ============================================================================
// Channel Transformations
// ============================================================================

/**
 * Parse channel type from string
 */
function parseChannelType(type: string): core.ChannelType {
  switch (type.toLowerCase()) {
    case 'home':
      return core.ChannelType.Home;
    case 'escrow':
      return core.ChannelType.Escrow;
    default:
      return core.ChannelType.Home;
  }
}

/**
 * Parse channel status from string
 */
function parseChannelStatus(status: string): core.ChannelStatus {
  switch (status.toLowerCase()) {
    case 'void':
      return core.ChannelStatus.Void;
    case 'open':
      return core.ChannelStatus.Open;
    case 'challenged':
      return core.ChannelStatus.Challenged;
    case 'closed':
      return core.ChannelStatus.Closed;
    default:
      return core.ChannelStatus.Void;
  }
}

/**
 * Transform a single RPC ChannelV1 to core Channel
 */
export function transformChannel(channel: ChannelV1): core.Channel {
  const result: core.Channel = {
    channelId: channel.channel_id,
    userWallet: channel.user_wallet as Address,
    asset: channel.asset,
    type: parseChannelType(channel.type),
    blockchainId: BigInt(channel.blockchain_id),
    tokenAddress: channel.token_address as Address,
    challengeDuration: channel.challenge_duration,
    nonce: BigInt(channel.nonce),
    approvedSigValidators: channel.approved_sig_validators,
    status: parseChannelStatus(channel.status),
    stateVersion: BigInt(channel.state_version),
  };

  if (channel.challenge_expires_at) {
    result.challengeExpiresAt = new Date(channel.challenge_expires_at);
  }

  return result;
}

// ============================================================================
// Ledger Transformations
// ============================================================================

/**
 * Transform RPC LedgerV1 to core Ledger
 */
export function transformLedger(ledger: LedgerV1): core.Ledger {
  return {
    tokenAddress: ledger.token_address as Address,
    blockchainId: BigInt(ledger.blockchain_id),
    userBalance: new Decimal(ledger.user_balance),
    userNetFlow: new Decimal(ledger.user_net_flow),
    nodeBalance: new Decimal(ledger.node_balance),
    nodeNetFlow: new Decimal(ledger.node_net_flow),
  };
}

// ============================================================================
// Transition Transformations
// ============================================================================

/**
 * Transform RPC TransitionV1 to core Transition
 */
export function transformTransition(transition: TransitionV1): core.Transition {
  const result: core.Transition = {
    type: transition.type, // Already TransitionType enum in RPC
    txId: transition.tx_id,
    amount: new Decimal(transition.amount),
  };

  if (transition.account_id) {
    result.accountId = transition.account_id;
  }

  return result;
}

// ============================================================================
// State Transformations
// ============================================================================

/**
 * Transform RPC StateV1 to core State
 */
export function transformState(state: StateV1): core.State {
  const result: core.State = {
    id: state.id,
    transition: transformTransition(state.transition),
    asset: state.asset,
    userWallet: state.user_wallet as Address,
    epoch: BigInt(state.epoch),
    version: BigInt(state.version),
    homeLedger: transformLedger(state.home_ledger),
  };

  if (state.home_channel_id) {
    result.homeChannelId = state.home_channel_id;
  }

  if (state.escrow_channel_id) {
    result.escrowChannelId = state.escrow_channel_id;
  }

  if (state.escrow_ledger) {
    result.escrowLedger = transformLedger(state.escrow_ledger);
  }

  if (state.user_sig) {
    result.userSig = state.user_sig as `0x${string}`;
  }

  if (state.node_sig) {
    result.nodeSig = state.node_sig as `0x${string}`;
  }

  return result;
}

// ============================================================================
// Transaction Transformations
// ============================================================================

/**
 * Transform RPC TransactionV1 to core Transaction
 */
export function transformTransaction(tx: TransactionV1): core.Transaction {
  const result: core.Transaction = {
    id: tx.id,
    asset: tx.asset,
    txType: tx.tx_type, // Already TransactionType enum in RPC
    fromAccount: tx.from_account as Address,
    toAccount: tx.to_account as Address,
    amount: new Decimal(tx.amount),
    createdAt: new Date(tx.created_at),
  };

  if (tx.sender_new_state_id) {
    result.senderNewStateId = tx.sender_new_state_id;
  }

  if (tx.receiver_new_state_id) {
    result.receiverNewStateId = tx.receiver_new_state_id;
  }

  return result;
}

// ============================================================================
// Pagination Transformations
// ============================================================================

/**
 * Transform RPC PaginationMetadataV1 to core PaginationMetadata
 */
export function transformPaginationMetadata(
  metadata: PaginationMetadataV1
): core.PaginationMetadata {
  return {
    page: metadata.page,
    perPage: metadata.per_page,
    totalCount: metadata.total_count,
    pageCount: metadata.page_count,
  };
}

// ============================================================================
// Action Allowance Transformations
// ============================================================================

/**
 * Transform RPC ActionAllowanceV1 to core ActionAllowance
 */
export function transformActionAllowance(a: ActionAllowanceV1): core.ActionAllowance {
  return {
    gatedAction: a.gated_action,
    timeWindow: a.time_window,
    allowance: BigInt(a.allowance),
    used: BigInt(a.used),
  };
}

// ============================================================================
// App Session Transformations
// ============================================================================

import { AppDefinitionV1, AppStateUpdateV1, SignedAppStateUpdateV1, AppParticipantV1, AppAllocationV1, AppSessionInfoV1 } from './app/types.js';
import * as RPCApp from './rpc/api.js';

/**
 * Transform SDK AppDefinitionV1 to RPC AppDefinitionV1 for requests
 * Converts camelCase SDK fields to snake_case RPC fields
 */
export function transformAppDefinitionToRPC(def: AppDefinitionV1): any {
  return {
    application_id: def.applicationId,
    participants: def.participants.map(p => ({
      wallet_address: p.walletAddress,
      signature_weight: p.signatureWeight,
    })),
    quorum: def.quorum,
    nonce: def.nonce.toString(),
  };
}

/**
 * Transform SDK AppStateUpdateV1 to RPC AppStateUpdateV1 for requests
 * Converts camelCase SDK fields to snake_case RPC fields
 */
export function transformAppStateUpdateToRPC(update: AppStateUpdateV1) {
  return {
    app_session_id: update.appSessionId,
    intent: update.intent,
    version: update.version.toString(),
    allocations: update.allocations.map(a => ({
      participant: a.participant,
      asset: a.asset,
      amount: a.amount.toString(),
    })),
    session_data: update.sessionData,
  };
}

/**
 * Transform SDK SignedAppStateUpdateV1 to RPC SignedAppStateUpdateV1 for requests
 * Converts camelCase SDK fields to snake_case RPC fields
 */
export function transformSignedAppStateUpdateToRPC(signed: SignedAppStateUpdateV1) {
  return {
    app_state_update: transformAppStateUpdateToRPC(signed.appStateUpdate),
    quorum_sigs: signed.quorumSigs,
  };
}

// ============================================================================
// App Session Incoming Transformations (RPC snake_case → SDK camelCase)
// ============================================================================

/**
 * Transform RPC AppSessionInfoV1 (snake_case) to SDK AppSessionInfoV1 (camelCase).
 * The server returns snake_case JSON that needs conversion to SDK types.
 */
export function transformAppSessionInfo(raw: any): AppSessionInfoV1 {
  return {
    appSessionId: raw.app_session_id,
    appDefinition: transformAppDefinitionFromRPC(raw.app_definition),
    isClosed: raw.status === 'closed',
    sessionData: raw.session_data || '',
    version: BigInt(raw.version),
    allocations: (raw.allocations || []).map((a: any) => ({
      participant: a.participant as Address,
      asset: a.asset,
      amount: new Decimal(a.amount),
    })),
  };
}

/**
 * Transform RPC AppDefinitionV1 (snake_case) to SDK AppDefinitionV1 (camelCase).
 * The server returns snake_case JSON that needs conversion to SDK types.
 */
export function transformAppDefinitionFromRPC(raw: any): AppDefinitionV1 {
  if (!raw.application_id || raw.nonce === undefined || raw.nonce === null) {
    throw new Error('Invalid app definition: missing required fields (application_id, nonce)');
  }
  return {
    applicationId: raw.application_id,
    participants: (raw.participants || []).map((p: any) => ({
      walletAddress: p.wallet_address as Address,
      signatureWeight: p.signature_weight,
    })),
    quorum: raw.quorum,
    nonce: BigInt(raw.nonce),
  };
}
