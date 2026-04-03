import { Address, Hex } from 'viem';
import Decimal from 'decimal.js';
import {
  State,
  ChannelDefinition,
  HomeChannelDataResponse,
  EscrowDepositDataResponse,
  EscrowWithdrawalDataResponse,
} from './types';
import {
  HomeChannelCreatedEvent,
  HomeChannelMigratedEvent,
  HomeChannelCheckpointedEvent,
  HomeChannelChallengedEvent,
  HomeChannelClosedEvent,
  EscrowDepositInitiatedEvent,
  EscrowDepositChallengedEvent,
  EscrowDepositFinalizedEvent,
  EscrowWithdrawalInitiatedEvent,
  EscrowWithdrawalChallengedEvent,
  EscrowWithdrawalFinalizedEvent,
} from './event';

// ============================================================================
// Listener Interface
// ============================================================================

/**
 * Listener defines the interface for listening to channel events
 */
export interface Listener {
  /**
   * Listen starts listening for events
   */
  listen(): Promise<void>;
}

// ============================================================================
// Client Interface
// ============================================================================

/**
 * Client defines the interface for interacting with the ChannelsHub smart contract
 */
export interface Client {
  // ========= Getters - IVault =========

  /**
   * Get balances for multiple accounts and tokens
   * @param accounts - Array of account addresses
   * @param tokens - Array of token addresses
   * @returns 2D array of balances [account][token]
   */
  getAccountsBalances(accounts: Address[], tokens: Address[]): Promise<Decimal[][]>;

  /**
   * Get parametric token sub balances for multiple accounts and subIds
   * @param accounts - Array of account addresses
   * @param token - Parametric token address
   * @param subIds - Array of subaccount IDs
   * @returns 2D array of balances [account][token]
   */
  getAccountsSubBalances(
    accounts: Address[],
    token: Address,
    subIds: number[],
  ): Promise<Decimal[][]>;

  // ========= Getters - Token Balance & Approval =========

  /**
   * Get on-chain token balance (ERC-20 or native) for a wallet
   * @param asset - Asset symbol (e.g., "usdc")
   * @param walletAddress - Wallet address to check
   */
  getTokenBalance(asset: string, walletAddress: Address): Promise<Decimal>;

  /**
   * Approve the contract to spend tokens for an asset
   * @param asset - Asset symbol (e.g., "usdc")
   * @param amount - Amount to approve
   * @returns Transaction hash
   */
  approve(asset: string, amount: Decimal): Promise<Hex>;

  // ========= Getters - ChannelsHub =========

  /**
   * Get the node's balance for a specific token
   * @param token - Token address
   * @param subId - Subaccount ID
   */
  getNodeBalance(token: Address, subId: number): Promise<Decimal>;

  /**
   * Get all open channel IDs for a user
   * @param user - User wallet address
   */
  getOpenChannels(user: Address): Promise<string[]>;

  /**
   * Get home channel data
   * @param homeChannelId - Home channel ID
   */
  getHomeChannelData(homeChannelId: string): Promise<HomeChannelDataResponse>;

  /**
   * Get escrow deposit data
   * @param escrowChannelId - Escrow channel ID
   */
  getEscrowDepositData(escrowChannelId: string): Promise<EscrowDepositDataResponse>;

  /**
   * Get escrow withdrawal data
   * @param escrowChannelId - Escrow channel ID
   */
  getEscrowWithdrawalData(escrowChannelId: string): Promise<EscrowWithdrawalDataResponse>;

  // ========= IVault functions =========

  /**
   * Deposit tokens into custody
   * @param node - Node address
   * @param token - Token address
   * @param amount - Amount to deposit
   * @returns Transaction hash
   */
  deposit(node: Address, token: Address, amount: Decimal): Promise<Hex>;

  /**
   * Withdraw tokens from custody
   * @param node - Node address
   * @param token - Token address
   * @param amount - Amount to withdraw
   * @returns Transaction hash
   */
  withdraw(node: Address, token: Address, amount: Decimal): Promise<Hex>;

  // ========= Channel lifecycle =========

  /**
   * Create a new channel
   * @param def - Channel definition
   * @param initCCS - Initial channel state
   * @returns Transaction hash
   */
  create(def: ChannelDefinition, initCCS: State): Promise<Hex>;

  /**
   * Migrate channel to this blockchain
   * @param def - Channel definition
   * @param candidate - Candidate state
   * @returns Transaction hash
   */
  migrateChannelHere(def: ChannelDefinition, candidate: State): Promise<Hex>;

  /**
   * Checkpoint a channel state on-chain
   * @param candidate - Candidate state
   * @returns Transaction hash
   */
  checkpoint(candidate: State): Promise<Hex>;

  /**
   * Challenge a channel state
   * @param candidate - Candidate state
   * @param challengerSig - Challenger signature
   * @param challengerIdx - Challenger participant index
   * @returns Transaction hash
   */
  challenge(candidate: State, challengerSig: Uint8Array, challengerIdx: number): Promise<Hex>;

  /**
   * Close a channel
   * @param candidate - Candidate state
   * @returns Transaction hash
   */
  close(candidate: State): Promise<Hex>;

  // ========= Escrow deposit =========

  /**
   * Initiate escrow deposit
   * @param def - Channel definition
   * @param initCCS - Initial channel state
   * @returns Transaction hash
   */
  initiateEscrowDeposit(def: ChannelDefinition, initCCS: State): Promise<Hex>;

  /**
   * Challenge escrow deposit
   * @param candidate - Candidate state
   * @param challengerSig - Challenger signature
   * @param challengerIdx - Challenger participant index
   * @returns Transaction hash
   */
  challengeEscrowDeposit(
    candidate: State,
    challengerSig: Uint8Array,
    challengerIdx: number,
  ): Promise<Hex>;

  /**
   * Finalize escrow deposit
   * @param candidate - Candidate state
   * @returns Transaction hash
   */
  finalizeEscrowDeposit(candidate: State): Promise<Hex>;

  // ========= Escrow withdrawal =========

  /**
   * Initiate escrow withdrawal
   * @param def - Channel definition
   * @param initCCS - Initial channel state
   * @returns Transaction hash
   */
  initiateEscrowWithdrawal(def: ChannelDefinition, initCCS: State): Promise<Hex>;

  /**
   * Challenge escrow withdrawal
   * @param candidate - Candidate state
   * @param challengerSig - Challenger signature
   * @param challengerIdx - Challenger participant index
   * @returns Transaction hash
   */
  challengeEscrowWithdrawal(
    candidate: State,
    challengerSig: Uint8Array,
    challengerIdx: number,
  ): Promise<Hex>;

  /**
   * Finalize escrow withdrawal
   * @param candidate - Candidate state
   * @returns Transaction hash
   */
  finalizeEscrowWithdrawal(candidate: State): Promise<Hex>;
}

// ============================================================================
// StateAdvancer Interface
// ============================================================================

/**
 * StateAdvancer validates state transitions
 */
export interface StateAdvancer {
  /**
   * Validate that proposedState is a valid advancement from currentState
   * @param currentState - Current state
   * @param proposedState - Proposed next state
   * @throws Error if advancement is invalid
   */
  validateAdvancement(currentState: State, proposedState: State): Promise<void>;
}

// ============================================================================
// StatePacker Interface
// ============================================================================

/**
 * StatePacker serializes channel states for on-chain submission
 */
export interface StatePacker {
  /**
   * Pack a state into bytes for on-chain submission
   * @param state - State to pack
   * @returns Packed bytes
   */
  packState(state: State): Promise<`0x${string}`>;
}

// ============================================================================
// AssetStore Interface
// ============================================================================

/**
 * AssetStore provides asset and token metadata
 */
export interface AssetStore {
  /**
   * Get asset decimals
   * @param asset - Asset symbol
   * @returns Number of decimals (uint8)
   */
  getAssetDecimals(asset: string): Promise<number>;

  /**
   * Get token decimals for a specific blockchain
   * @param blockchainId - Blockchain ID (uint64)
   * @param tokenAddress - Token contract address
   * @returns Number of decimals (uint8)
   */
  getTokenDecimals(blockchainId: bigint, tokenAddress: Address): Promise<number>;
}

// ============================================================================
// BlockchainEventHandler Interface
// ============================================================================

/**
 * BlockchainEventHandler handles channel lifecycle events from the blockchain
 */
export interface BlockchainEventHandler {
  /**
   * Handle home channel created event
   */
  handleHomeChannelCreated(event: HomeChannelCreatedEvent): Promise<void>;

  /**
   * Handle home channel migrated event
   */
  handleHomeChannelMigrated(event: HomeChannelMigratedEvent): Promise<void>;

  /**
   * Handle home channel checkpointed event
   */
  handleHomeChannelCheckpointed(event: HomeChannelCheckpointedEvent): Promise<void>;

  /**
   * Handle home channel challenged event
   */
  handleHomeChannelChallenged(event: HomeChannelChallengedEvent): Promise<void>;

  /**
   * Handle home channel closed event
   */
  handleHomeChannelClosed(event: HomeChannelClosedEvent): Promise<void>;

  /**
   * Handle escrow deposit initiated event
   */
  handleEscrowDepositInitiated(event: EscrowDepositInitiatedEvent): Promise<void>;

  /**
   * Handle escrow deposit challenged event
   */
  handleEscrowDepositChallenged(event: EscrowDepositChallengedEvent): Promise<void>;

  /**
   * Handle escrow deposit finalized event
   */
  handleEscrowDepositFinalized(event: EscrowDepositFinalizedEvent): Promise<void>;

  /**
   * Handle escrow withdrawal initiated event
   */
  handleEscrowWithdrawalInitiated(event: EscrowWithdrawalInitiatedEvent): Promise<void>;

  /**
   * Handle escrow withdrawal challenged event
   */
  handleEscrowWithdrawalChallenged(event: EscrowWithdrawalChallengedEvent): Promise<void>;

  /**
   * Handle escrow withdrawal finalized event
   */
  handleEscrowWithdrawalFinalized(event: EscrowWithdrawalFinalizedEvent): Promise<void>;
}
