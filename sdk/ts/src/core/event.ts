import { Address, Hex } from 'viem';

// ============================================================================
// Base Event Types
// ============================================================================

/**
 * Base channel event structure
 */
export interface ChannelEvent {
  channelId: string;
  stateVersion: bigint; // uint64
}

/**
 * Channel challenged event structure (includes challenge expiry)
 */
export interface ChannelChallengedEvent {
  channelId: string;
  stateVersion: bigint; // uint64
  challengeExpiry: bigint; // uint64
}

/**
 * Blockchain event metadata
 */
export interface BlockchainEvent {
  contractAddress: Address;
  blockchainId: bigint; // uint64
  name: string;
  blockNumber: bigint; // uint64
  transactionHash: Hex;
  logIndex: number; // uint32
}

// ============================================================================
// Home Channel Events
// ============================================================================

/**
 * HomeChannelCreatedEvent represents the ChannelCreated event
 */
export type HomeChannelCreatedEvent = ChannelEvent;

/**
 * HomeChannelMigratedEvent represents the ChannelMigrated event
 */
export type HomeChannelMigratedEvent = ChannelEvent;

/**
 * HomeChannelCheckpointedEvent represents the Checkpointed event
 */
export type HomeChannelCheckpointedEvent = ChannelEvent;

/**
 * HomeChannelChallengedEvent represents the Challenged event
 */
export type HomeChannelChallengedEvent = ChannelChallengedEvent;

/**
 * HomeChannelClosedEvent represents the Closed event
 */
export type HomeChannelClosedEvent = ChannelEvent;

// ============================================================================
// Escrow Deposit Events
// ============================================================================

/**
 * EscrowDepositInitiatedEvent represents the EscrowDepositInitiated event
 */
export type EscrowDepositInitiatedEvent = ChannelEvent;

/**
 * EscrowDepositChallengedEvent represents the EscrowDepositChallenged event
 */
export type EscrowDepositChallengedEvent = ChannelChallengedEvent;

/**
 * EscrowDepositFinalizedEvent represents the EscrowDepositFinalized event
 */
export type EscrowDepositFinalizedEvent = ChannelEvent;

// ============================================================================
// Escrow Withdrawal Events
// ============================================================================

/**
 * EscrowWithdrawalInitiatedEvent represents the EscrowWithdrawalInitiated event
 */
export type EscrowWithdrawalInitiatedEvent = ChannelEvent;

/**
 * EscrowWithdrawalChallengedEvent represents the EscrowWithdrawalChallenged event
 */
export type EscrowWithdrawalChallengedEvent = ChannelChallengedEvent;

/**
 * EscrowWithdrawalFinalizedEvent represents the EscrowWithdrawalFinalized event
 */
export type EscrowWithdrawalFinalizedEvent = ChannelEvent;

// ============================================================================
// Validator Events
// ============================================================================

/**
 * Emitted when the node registers a new signature validator on ChannelHub.
 * Users must react to unexpected registrations by revoking ERC20 approvals
 * granted to ChannelHub — see contracts/SECURITY.md.
 *
 * `validator` is always EIP-55 checksummed. Compare with getAddress(ev.validator)
 * to avoid silent mismatches against lowercase or non-checksummed config values.
 *
 * Pass `blockNumber + 1n` as `fromBlock` on reconnect for gap-free monitoring.
 */
export interface ValidatorRegisteredEvent {
    blockchainId: bigint;
    validatorId: number; // uint8
    validator: Address; // EIP-55 checksummed
    blockNumber: bigint; // use as fromBlock on reconnect
}
