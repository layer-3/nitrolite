import { Address } from 'viem';
import Decimal from 'decimal.js';

// ============================================================================
// Enums
// ============================================================================

/**
 * AppStateUpdateIntent represents the intent of an app state update
 */
export enum AppStateUpdateIntent {
  Operate = 0,
  Deposit = 1,
  Withdraw = 2,
  Close = 3,
  Rebalance = 4,
}

/**
 * AppSessionStatus represents the status of an app session
 */
export enum AppSessionStatus {
  Void = 0,
  Open = 1,
  Closed = 2,
}

// ============================================================================
// Helper Functions for Enums
// ============================================================================

export function appStateUpdateIntentToString(intent: AppStateUpdateIntent): string {
  switch (intent) {
    case AppStateUpdateIntent.Operate:
      return 'operate';
    case AppStateUpdateIntent.Deposit:
      return 'deposit';
    case AppStateUpdateIntent.Withdraw:
      return 'withdraw';
    case AppStateUpdateIntent.Close:
      return 'close';
    case AppStateUpdateIntent.Rebalance:
      return 'rebalance';
    default:
      return 'unknown';
  }
}

export function appSessionStatusToString(status: AppSessionStatus): string {
  switch (status) {
    case AppSessionStatus.Void:
      return '';
    case AppSessionStatus.Open:
      return 'open';
    case AppSessionStatus.Closed:
      return 'closed';
    default:
      return 'unknown';
  }
}

// ============================================================================
// Core Data Structures
// ============================================================================

/**
 * AppSessionV1 represents an application session
 */
export interface AppSessionV1 {
  sessionId: string;
  application: string;
  participants: AppParticipantV1[];
  quorum: number; // uint8
  nonce: bigint; // uint64
  status: AppSessionStatus;
  version: bigint; // uint64
  sessionData: string;
  createdAt: Date;
  updatedAt: Date;
}

/**
 * AppParticipantV1 represents a participant in an app session
 */
export interface AppParticipantV1 {
  walletAddress: Address;
  signatureWeight: number; // uint8
}

/**
 * AppDefinitionV1 represents the definition for an app session
 */
export interface AppDefinitionV1 {
  applicationId: string;
  participants: AppParticipantV1[];
  quorum: number; // uint8
  nonce: bigint; // uint64
}

/**
 * AppSessionVersionV1 represents a session ID and version pair for rebalancing operations
 */
export interface AppSessionVersionV1 {
  sessionId: string;
  version: bigint; // uint64
}

/**
 * AppAllocationV1 represents the allocation of assets to a participant in an app session
 */
export interface AppAllocationV1 {
  participant: Address;
  asset: string;
  amount: Decimal;
}

/**
 * AppStateUpdateV1 represents the current state of an application session
 */
export interface AppStateUpdateV1 {
  appSessionId: string;
  intent: AppStateUpdateIntent;
  version: bigint; // uint64
  allocations: AppAllocationV1[];
  sessionData: string;
}

/**
 * SignedAppStateUpdateV1 represents a signed application session state update
 */
export interface SignedAppStateUpdateV1 {
  appStateUpdate: AppStateUpdateV1;
  quorumSigs: string[];
}

/**
 * AppSessionInfoV1 represents information about an application session
 */
export interface AppSessionInfoV1 {
  appSessionId: string;
  appDefinition: AppDefinitionV1;
  isClosed: boolean;
  sessionData: string;
  version: bigint; // uint64
  allocations: AppAllocationV1[];
}

/**
 * SessionKeyV1 represents a session key with spending allowances
 */
export interface SessionKeyV1 {
  id: bigint; // uint
  sessionKey: string;
  application: string;
  allowances: AssetAllowanceV1[];
  scope?: string;
  expiresAt: Date;
  createdAt: Date;
}

/**
 * AppSessionKeyStateV1 represents the state of a session key delegation.
 */
export interface AppSessionKeyStateV1 {
  /** User wallet address */
  user_address: string;
  /** Session key address for delegation */
  session_key: string;
  /** Version of the session key state */
  version: string;
  /** Application IDs associated with this session key */
  application_ids: string[];
  /** App session IDs associated with this session key */
  app_session_ids: string[];
  /** Unix timestamp in seconds indicating when the session key expires */
  expires_at: string;
  /** User's signature over the session key metadata */
  user_sig: string;
}

/**
 * AssetAllowanceV1 represents an asset allowance with usage tracking
 */
export interface AssetAllowanceV1 {
  asset: string;
  allowance: Decimal;
  used: Decimal;
}
