import { Address, Hex } from 'viem';
import Decimal from 'decimal.js';

// ============================================================================
// Enums
// ============================================================================

export enum ChannelType {
  Home = 1,
  Escrow = 2,
}

export enum ChannelParticipant {
  User = 0,
  Node = 1,
}

export enum ChannelSignerType {
  Default = 0x00,
  SessionKey = 0x01,
}

export enum ChannelStatus {
  Void = 0,
  Open = 1,
  Challenged = 2,
  Closed = 3,
}

export enum TransitionType {
  Void = 0,
  Acknowledgement = 1,
  HomeDeposit = 10,
  HomeWithdrawal = 11,
  EscrowDeposit = 20,
  EscrowWithdraw = 21,
  TransferSend = 30,
  TransferReceive = 31,
  Commit = 40,
  Release = 41,
  Migrate = 100,
  EscrowLock = 110,
  MutualLock = 120,
  Finalize = 200,
}

export enum TransactionType {
  HomeDeposit = 10,
  HomeWithdrawal = 11,
  EscrowDeposit = 20,
  EscrowWithdraw = 21,
  Transfer = 30,
  Commit = 40,
  Release = 41,
  Rebalance = 42,
  Migrate = 100,
  EscrowLock = 110,
  MutualLock = 120,
  Finalize = 200,
}

// Intent constants for on-chain representation
export const INTENT_OPERATE = 0;
export const INTENT_CLOSE = 1;
export const INTENT_DEPOSIT = 2;
export const INTENT_WITHDRAW = 3;
export const INTENT_INITIATE_ESCROW_DEPOSIT = 4;
export const INTENT_FINALIZE_ESCROW_DEPOSIT = 5;
export const INTENT_INITIATE_ESCROW_WITHDRAWAL = 6;
export const INTENT_FINALIZE_ESCROW_WITHDRAWAL = 7;
export const INTENT_INITIATE_MIGRATION = 8;
export const INTENT_FINALIZE_MIGRATION = 9;

// ============================================================================
// Core Data Structures
// ============================================================================

/**
 * ChannelDefinition represents configuration for creating a channel
 */
export interface ChannelDefinition {
  nonce: bigint; // uint64 - A unique number to prevent replay attacks
  challenge: number; // uint32 - Challenge period for the channel in seconds
  approvedSigValidators: string; // Hex string bitmap of approved signature validators
}

/**
 * Channel represents an on-chain channel
 */
export interface Channel {
  channelId: string; // Unique identifier for the channel
  userWallet: Address; // User wallet address
  asset: string; // Asset symbol (e.g. usdc, eth)
  type: ChannelType; // Type of the channel (home, escrow)
  blockchainId: bigint; // uint64 - Unique identifier for the blockchain
  tokenAddress: Address; // Address of the token used in the channel
  challengeDuration: number; // uint32 - Challenge period for the channel in seconds
  challengeExpiresAt?: Date; // Timestamp when the challenge period elapses
  nonce: bigint; // uint64 - Nonce for the channel
  approvedSigValidators: string; // Hex string bitmap of approved signature validators
  status: ChannelStatus; // Current status of the channel (void, open, challenged, closed)
  stateVersion: bigint; // uint64 - On-chain state version of the channel
  subId?: number; // uint48 - Optional sub-account ID for parametric tokens
  isParametric?: boolean; // Whether this channel uses a parametric token
}

/**
 * Ledger represents user and node balances for a channel
 */
export interface Ledger {
  tokenAddress: Address;
  blockchainId: bigint; // uint64
  userBalance: Decimal;
  userNetFlow: Decimal;
  nodeBalance: Decimal;
  nodeNetFlow: Decimal;
}

/**
 * Transition represents a state transition
 */
export interface Transition {
  type: TransitionType;
  txId: string;
  accountId?: string; // Optional - may be undefined for certain transition types
  amount: Decimal;
}

/**
 * State represents the current state of the user stored on Node
 */
export interface State {
  id: string; // Deterministic ID (hash) of the state
  transition: Transition; // Transition included in the state
  asset: string; // Asset type of the state
  userWallet: Address; // User wallet address
  epoch: bigint; // uint64 - User Epoch Index
  version: bigint; // uint64 - Version of the state
  homeChannelId?: string; // Identifier for the home Channel ID
  escrowChannelId?: string; // Identifier for the escrow Channel ID
  homeLedger: Ledger; // User and node balances for the home channel
  escrowLedger?: Ledger; // User and node balances for the escrow channel
  userSig?: Hex; // User signature for the state
  nodeSig?: Hex; // Node signature for the state
}

/**
 * Transaction represents a transaction record
 */
export interface Transaction {
  id: string;
  asset: string;
  txType: TransactionType;
  fromAccount: Address;
  toAccount: Address;
  senderNewStateId?: string;
  receiverNewStateId?: string;
  amount: Decimal;
  createdAt: Date;
}

// ============================================================================
// Blockchain & Asset Types
// ============================================================================

export interface Blockchain {
  name: string;
  id: bigint; // uint64
  channelHubAddress: Address;
  lockingContractAddress?: Address;
  blockStep: bigint; // uint64
}

export interface Token {
  name: string;
  symbol: string;
  address: Address;
  blockchainId: bigint; // uint64
  decimals: number; // uint8
}

export interface Asset {
  name: string;
  decimals: number; // uint8
  symbol: string;
  suggestedBlockchainId: bigint; // uint64
  tokens: Token[];
}

// ============================================================================
// Session & Balance Types
// ============================================================================

export interface AssetAllowance {
  asset: string;
  allowance: Decimal;
  used: Decimal;
}

export interface SessionKey {
  id: bigint; // uint64
  sessionKey: string;
  application: string;
  allowances: AssetAllowance[];
  scope?: string;
  expiresAt: string;
  createdAt: string;
}

export interface BalanceEntry {
  asset: string;
  balance: Decimal;
}

export interface ActionAllowance {
  gatedAction: string;
  timeWindow: string;
  allowance: bigint;
  used: bigint;
}

// ============================================================================
// Pagination Types
// ============================================================================

export interface PaginationParams {
  offset?: number; // uint32
  limit?: number; // uint32
}

export interface PaginationMetadata {
  page: number; // uint32
  perPage: number; // uint32
  totalCount: number; // uint32
  pageCount: number; // uint32
}

// ============================================================================
// Node Configuration
// ============================================================================

export interface NodeConfig {
  nodeAddress: Address;
  nodeVersion: string;
  supportedSigValidators: number[];
  blockchains: Blockchain[];
}

// ============================================================================
// Blockchain Response Types
// ============================================================================

export interface HomeChannelDataResponse {
  definition: ChannelDefinition;
  node: Address;
  lastState: State;
  challengeExpiry: bigint; // uint64
  subId: number; // uint48
}

export interface EscrowDepositDataResponse {
  escrowChannelId: string;
  node: Address;
  lastState: State;
  unlockExpiry: bigint; // uint64
  challengeExpiry: bigint; // uint64
}

export interface EscrowWithdrawalDataResponse {
  escrowChannelId: string;
  node: Address;
  lastState: State;
}

// ============================================================================
// Constructor Functions
// ============================================================================

/**
 * NewChannel creates a new Channel instance
 * Matches: func NewChannel(channelID, userWallet, asset string, ChType ChannelType, blockchainID uint64, tokenAddress string, nonce uint64, challenge uint32) *Channel
 */
export function newChannel(
  channelId: string,
  userWallet: Address,
  asset: string,
  type: ChannelType,
  blockchainId: bigint,
  tokenAddress: Address,
  nonce: bigint,
  challenge: number,
  approvedSigValidators: string = '0x00',
  subId?: number,
): Channel {
  return {
    channelId,
    userWallet,
    asset,
    type,
    blockchainId,
    tokenAddress,
    nonce,
    challengeDuration: challenge,
    approvedSigValidators,
    status: ChannelStatus.Void,
    stateVersion: 0n,
    subId,
  };
}

/**
 * NewVoidState creates a new void state with zero balances
 * Matches: func NewVoidState(asset, userWallet string) *State
 */
export function newVoidState(asset: string, userWallet: Address): State {
  return {
    id: '', // Will be set by getStateId
    transition: { type: TransitionType.Void, txId: '', amount: new Decimal(0) },
    asset,
    userWallet,
    epoch: 0n,
    version: 0n,
    homeLedger: {
      tokenAddress: '0x0' as Address,
      blockchainId: 0n,
      userBalance: new Decimal(0),
      userNetFlow: new Decimal(0),
      nodeBalance: new Decimal(0),
      nodeNetFlow: new Decimal(0),
    },
  };
}

/**
 * NewTransition creates a new Transition instance
 */
export function newTransition(
  type: TransitionType,
  txId: string,
  accountId: string,
  amount: Decimal,
): Transition {
  return {
    type,
    txId,
    accountId,
    amount,
  };
}

/**
 * NewTransaction creates a new Transaction instance
 */
export function newTransaction(
  id: string,
  asset: string,
  txType: TransactionType,
  fromAccount: Address,
  toAccount: Address,
  amount: Decimal,
): Transaction {
  return {
    id,
    asset,
    txType,
    fromAccount,
    toAccount,
    amount,
    createdAt: new Date(),
  };
}

// ============================================================================
// Helper Functions
// ============================================================================

export function transitionToString(type: TransitionType): string {
  switch (type) {
    case TransitionType.Void:
      return 'Void';
    case TransitionType.Acknowledgement:
      return 'Acknowledgement';
    case TransitionType.HomeDeposit:
      return 'HomeDeposit';
    case TransitionType.HomeWithdrawal:
      return 'HomeWithdrawal';
    case TransitionType.EscrowDeposit:
      return 'EscrowDeposit';
    case TransitionType.EscrowWithdraw:
      return 'EscrowWithdraw';
    case TransitionType.TransferSend:
      return 'TransferSend';
    case TransitionType.TransferReceive:
      return 'TransferReceive';
    case TransitionType.Commit:
      return 'Commit';
    case TransitionType.Release:
      return 'Release';
    case TransitionType.Migrate:
      return 'Migrate';
    case TransitionType.EscrowLock:
      return 'EscrowLock';
    case TransitionType.MutualLock:
      return 'MutualLock';
    case TransitionType.Finalize:
      return 'Finalize';
    default:
      return 'Unknown';
  }
}

export function transitionRequiresOpenChannel(type: TransitionType): boolean {
  return type !== TransitionType.TransferReceive && type !== TransitionType.Release;
}

export function transitionsEqual(a: Transition, b: Transition): string | null {
  if (a.type !== b.type) {
    return `type mismatch: expected=${a.type}, proposed=${b.type}`;
  }
  if (a.txId !== b.txId) {
    return `tx ID mismatch: expected=${a.txId}, proposed=${b.txId}`;
  }
  if (a.accountId !== b.accountId) {
    return `account ID mismatch: expected=${a.accountId}, proposed=${b.accountId}`;
  }
  if (!a.amount.equals(b.amount)) {
    return `amount mismatch: expected=${a.amount.toString()}, proposed=${b.amount.toString()}`;
  }
  return null;
}

// ============================================================================
// Ledger Methods
// ============================================================================

export function ledgerEqual(a: Ledger, b: Ledger): string | null {
  if (a.tokenAddress.toLowerCase() !== b.tokenAddress.toLowerCase()) {
    return `token address mismatch: expected=${a.tokenAddress}, proposed=${b.tokenAddress}`;
  }
  if (a.blockchainId !== b.blockchainId) {
    return `blockchain ID mismatch: expected=${a.blockchainId}, proposed=${b.blockchainId}`;
  }
  if (!a.userBalance.equals(b.userBalance)) {
    return `user balance mismatch: expected=${a.userBalance.toString()}, proposed=${b.userBalance.toString()}`;
  }
  if (!a.userNetFlow.equals(b.userNetFlow)) {
    return `user net flow mismatch: expected=${a.userNetFlow.toString()}, proposed=${b.userNetFlow.toString()}`;
  }
  if (!a.nodeBalance.equals(b.nodeBalance)) {
    return `node balance mismatch: expected=${a.nodeBalance.toString()}, proposed=${b.nodeBalance.toString()}`;
  }
  if (!a.nodeNetFlow.equals(b.nodeNetFlow)) {
    return `node net flow mismatch: expected=${a.nodeNetFlow.toString()}, proposed=${b.nodeNetFlow.toString()}`;
  }
  return null;
}

export function validateLedger(ledger: Ledger): void {
  if (!ledger.tokenAddress || ledger.tokenAddress === ('0x0' as Address)) {
    throw new Error('invalid token address');
  }
  if (ledger.blockchainId === 0n) {
    throw new Error('invalid blockchain ID');
  }
  if (ledger.userBalance.isNegative()) {
    throw new Error('user balance cannot be negative');
  }
  if (ledger.nodeBalance.isNegative()) {
    throw new Error('node balance cannot be negative');
  }

  const sumBalances = ledger.userBalance.add(ledger.nodeBalance);
  const sumNetFlows = ledger.userNetFlow.add(ledger.nodeNetFlow);
  if (!sumBalances.equals(sumNetFlows)) {
    throw new Error(
      `ledger balances do not match net flows: balances=${sumBalances.toString()}, net_flows=${sumNetFlows.toString()}`,
    );
  }
}

// ============================================================================
// Pagination Utilities
// ============================================================================

export function getOffsetAndLimit(
  params: PaginationParams | undefined,
  defaultLimit: number,
  maxLimit: number,
): { offset: number; limit: number } {
  if (!params) {
    return { offset: 0, limit: defaultLimit };
  }

  const offset = params.offset ?? 0;
  let limit = params.limit ?? defaultLimit;

  // Enforce max limit
  if (limit > maxLimit) {
    limit = maxLimit;
  }

  return { offset, limit };
}
