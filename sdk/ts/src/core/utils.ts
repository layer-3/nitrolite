import { Address, Hex, encodeAbiParameters, keccak256, toHex, pad, slice } from 'viem';
import Decimal from 'decimal.js';
import {
  Transition,
  TransitionType,
  INTENT_OPERATE,
  INTENT_CLOSE,
  INTENT_DEPOSIT,
  INTENT_WITHDRAW,
  INTENT_INITIATE_ESCROW_DEPOSIT,
  INTENT_FINALIZE_ESCROW_DEPOSIT,
  INTENT_INITIATE_ESCROW_WITHDRAWAL,
  INTENT_FINALIZE_ESCROW_WITHDRAWAL,
  INTENT_INITIATE_MIGRATION,
} from './types';

// Configure Decimal.js for high precision arithmetic
Decimal.set({ precision: 50 });

/**
 * ChannelHubVersion is the version of the ChannelHub contract.
 * Encoded as the first byte of channelId to prevent replay attacks
 * across different ChannelHub deployments on the same chain.
 */
export const CHANNEL_HUB_VERSION = 1;

// ============================================================================
// Intent Conversion
// ============================================================================

/**
 * TransitionToIntent maps a transition type to its on-chain intent value
 * @param transition - The transition to convert
 * @returns Intent value (uint8)
 * @throws Error if transition is null or has unexpected type
 */
export function transitionToIntent(transition: Transition): number {
  switch (transition.type) {
    case TransitionType.Void:
    case TransitionType.TransferSend:
    case TransitionType.TransferReceive:
    case TransitionType.Commit:
    case TransitionType.Release:
      return INTENT_OPERATE;
    case TransitionType.Finalize:
      return INTENT_CLOSE;
    case TransitionType.HomeDeposit:
      return INTENT_DEPOSIT;
    case TransitionType.HomeWithdrawal:
      return INTENT_WITHDRAW;
    case TransitionType.MutualLock:
      return INTENT_INITIATE_ESCROW_DEPOSIT;
    case TransitionType.EscrowDeposit:
      return INTENT_FINALIZE_ESCROW_DEPOSIT;
    case TransitionType.EscrowLock:
      return INTENT_INITIATE_ESCROW_WITHDRAWAL;
    case TransitionType.EscrowWithdraw:
      return INTENT_FINALIZE_ESCROW_WITHDRAWAL;
    case TransitionType.Migrate:
      return INTENT_INITIATE_MIGRATION;
    default:
      return INTENT_OPERATE;
  }
}

// ============================================================================
// Decimal Validation & Conversion
// ============================================================================

/**
 * ValidateDecimalPrecision validates that an amount doesn't exceed the maximum allowed decimal places
 * @param amount - The decimal amount to validate
 * @param maxDecimals - Maximum allowed decimal places (uint8)
 * @throws Error if amount exceeds precision
 */
export function validateDecimalPrecision(amount: Decimal, maxDecimals: number): void {
  const exponent = amount.decimalPlaces();
  if (exponent > maxDecimals) {
    throw new Error(`amount exceeds maximum decimal precision: max ${maxDecimals} decimals allowed, got ${exponent}`);
  }
}

/**
 * DecimalToBigInt converts a decimal.Decimal amount to bigint scaled to the token's smallest unit.
 * For example, 1.23 USDC (6 decimals) becomes 1230000n.
 * This is used when preparing amounts for smart contract calls.
 * @param amount - The decimal amount
 * @param decimals - Number of decimals for the token (uint8)
 * @returns Scaled bigint amount
 * @throws Error if amount has too many decimal places
 */
export function decimalToBigInt(amount: Decimal, decimals: number): bigint {
  // Calculate the multiplier (e.g., 10^6)
  const multiplier = new Decimal(10).pow(decimals);

  // Scale the amount
  const scaled = amount.mul(multiplier);

  // Check if it's an integer
  if (!scaled.isInteger()) {
    throw new Error(`amount ${amount.toString()} exceeds maximum decimal precision: max ${decimals} decimals allowed`);
  }

  // Convert to bigint
  return BigInt(scaled.toFixed(0));
}

// ============================================================================
// Channel ID Generation
// ============================================================================

/**
 * GetHomeChannelID generates a unique identifier for a primary channel based on its definition.
 * This matches the Solidity getChannelId function which computes keccak256(abi.encode(ChannelDefinition)).
 * The metadata is derived from the asset: first 8 bytes of keccak256(asset) padded to 32 bytes.
 * @param node - Node address
 * @param user - User wallet address
 * @param asset - Asset symbol
 * @param nonce - Channel nonce (uint64)
 * @param challengeDuration - Challenge period in seconds (uint32)
 * @returns Channel ID as hex string
 */
export function getHomeChannelId(node: Address, user: Address, asset: string, nonce: bigint, challengeDuration: number, approvedSigValidators: string = '0x00'): string {
  // Generate metadata from asset
  const metadata = generateChannelMetadata(asset);

  // Convert hex bitmap to bigint for ABI encoding
  const validatorsBigInt = BigInt(approvedSigValidators || '0x00');

  // Define the channel definition struct matching Solidity
  // struct ChannelDefinition {
  //   uint32 challengeDuration;
  //   address user;
  //   address node;
  //   uint64 nonce;
  //   uint256 approvedSignatureValidators;
  //   bytes32 metadata;
  // }
  const packed = encodeAbiParameters(
    [
      {
        name: 'channelDefinition',
        type: 'tuple',
        components: [
          { name: 'challengeDuration', type: 'uint32' },
          { name: 'user', type: 'address' },
          { name: 'node', type: 'address' },
          { name: 'nonce', type: 'uint64' },
          { name: 'approvedSignatureValidators', type: 'uint256' },
          { name: 'metadata', type: 'bytes32' },
        ],
      },
    ],
    [
      {
        challengeDuration: challengeDuration,
        user: user,
        node: node,
        nonce: nonce,
        approvedSignatureValidators: validatorsBigInt,
        metadata: metadata,
      },
    ],
  );

  // Calculate base channelId
  const baseId = keccak256(packed);

  // Set the first byte (most significant byte) to the ChannelHub version
  // This matches Go's getHomeChannelID: versionedId[0] = channelHubVersion
  const versionHex = CHANNEL_HUB_VERSION.toString(16).padStart(2, '0');
  return `0x${versionHex}${baseId.slice(4)}` as Hex;
}

/**
 * GetEscrowChannelID derives an escrow-specific channel ID based on a home channel and state version.
 * This matches the Solidity getEscrowId function which computes keccak256(abi.encode(channelId, version)).
 * @param homeChannelId - Home channel ID (bytes32)
 * @param stateVersion - State version (uint64)
 * @returns Escrow channel ID as hex string
 */
export function getEscrowChannelId(homeChannelId: string, stateVersion: bigint): string {
  const packed = encodeAbiParameters([{ type: 'bytes32' }, { type: 'uint64' }], [homeChannelId as `0x${string}`, stateVersion]);

  return keccak256(packed);
}

// ============================================================================
// State ID Generation
// ============================================================================

/**
 * GetStateID creates a unique hash representing a specific snapshot of a user's wallet and asset state.
 * @param userWallet - User wallet address
 * @param asset - Asset symbol
 * @param epoch - User epoch index (uint64)
 * @param version - State version (uint64)
 * @returns State ID as hex string
 */
export function getStateId(userWallet: Address, asset: string, epoch: bigint, version: bigint): string {
  const packed = encodeAbiParameters([{ type: 'address' }, { type: 'string' }, { type: 'uint256' }, { type: 'uint256' }], [userWallet, asset, epoch, version]);

  return keccak256(packed);
}

/**
 * GetStateTransitionHash hashes a single transition into metadata
 * @param transition - The transition to hash
 * @returns Hash as bytes32 (hex string)
 */
export function getStateTransitionHash(transition: Transition): string {
  const contractTransition = {
    type: transition.type,
    txId: hexToBytes32(transition.txId),
    accountId: parseAccountIdToBytes32(transition.accountId),
    amount: transition.amount.toString(),
  };

  const packed = encodeAbiParameters(
    [
      {
        type: 'tuple',
        components: [
          { name: 'type', type: 'uint8' },
          { name: 'txId', type: 'bytes32' },
          { name: 'accountId', type: 'bytes32' },
          { name: 'amount', type: 'string' },
        ],
      },
    ],
    [contractTransition],
  );

  return keccak256(packed);
}
// ============================================================================
// Transaction ID Generation
// ============================================================================

/**
 * GetSenderTransactionID calculates and returns a unique transaction ID reference for actions initiated by user.
 * @param toAccount - Recipient account
 * @param senderNewStateId - Sender's new state ID
 * @returns Transaction ID as hex string
 */
export function getSenderTransactionId(toAccount: string, senderNewStateId: string): string {
  return getTransactionId(toAccount, senderNewStateId);
}

/**
 * GetReceiverTransactionID calculates and returns a unique transaction ID reference for actions initiated by node.
 * @param fromAccount - Sender account
 * @param receiverNewStateId - Receiver's new state ID
 * @returns Transaction ID as hex string
 */
export function getReceiverTransactionId(fromAccount: string, receiverNewStateId: string): string {
  return getTransactionId(fromAccount, receiverNewStateId);
}

function getTransactionId(account: string, newStateId: string): string {
  const packed = encodeAbiParameters([{ type: 'string' }, { type: 'bytes32' }], [account, newStateId as `0x${string}`]);

  return keccak256(packed);
}

// ============================================================================
// Metadata Generation
// ============================================================================

/**
 * GenerateChannelMetadata creates metadata from an asset by taking the first 8 bytes of keccak256(asset)
 * and padding the rest with zeros to make a 32-byte array.
 * @param asset - Asset symbol
 * @returns 32-byte metadata as hex string
 */
export function generateChannelMetadata(asset: string): `0x${string}` {
  // Hash the asset
  const assetHash = keccak256(toHex(asset));

  // Take first 8 bytes and pad with zeros to 32 bytes (pad on the right)
  const first8Bytes = slice(assetHash, 0, 8);
  const metadata = pad(first8Bytes, { dir: 'right', size: 32 });

  return metadata;
}

// ============================================================================
// Helper Functions for Bytes32 Conversion
// ============================================================================

/**
 * hexToBytes32 converts a hex string (with or without 0x prefix) to bytes32
 * Matches Go's common.HexToHash behavior: attempts to decode hex, returns zeros on failure
 * @param hexStr - Hex string representing a 32-byte hash
 * @returns Normalized bytes32 hex string
 */
function hexToBytes32(hexStr: string): `0x${string}` {
  // Remove 0x prefix if present
  let cleaned = hexStr.startsWith('0x') ? hexStr.slice(2) : hexStr;

  // If odd length, prepend '0' (matches Go's FromHex behavior)
  if (cleaned.length % 2 === 1) {
    cleaned = '0' + cleaned;
  }

  // Try to decode hex - if any character is invalid, this will parse what it can
  // Match Go's behavior: decode valid hex pairs, stop at first invalid character
  const bytes: number[] = [];
  for (let i = 0; i < cleaned.length; i += 2) {
    const hexPair = cleaned.slice(i, i + 2);
    // Strictly validate that both characters are valid hex
    if (!/^[0-9a-fA-F]{2}$/.test(hexPair)) {
      // Invalid hex pair - stop here (matches Go's behavior)
      break;
    }
    bytes.push(parseInt(hexPair, 16));
  }

  // Convert to hex string and right-pad to 32 bytes (matches Go's BytesToHash)
  const hexResult = bytes.map((b) => b.toString(16).padStart(2, '0')).join('');
  return pad(`0x${hexResult}` as `0x${string}`, { dir: 'left', size: 32 });
}

/**
 * parseAccountIdToBytes32 converts an account ID (address or hash) to bytes32
 * - If the input is a 20-byte address (40 hex chars), it's left-padded with zeros
 * - If the input is a 32-byte hash (64 hex chars), it's used as-is
 * In Ethereum, when an address is stored in bytes32, it occupies the rightmost 20 bytes,
 * with the leftmost 12 bytes being zeros.
 * Matches Go's behavior for invalid hex strings.
 * @param accountId - Account ID (address or hash)
 * @returns Normalized bytes32 hex string
 */
function parseAccountIdToBytes32(accountId: string | undefined): `0x${string}` {
  // Handle undefined, null, or empty string - return zero bytes32
  if (!accountId || accountId === '') {
    return '0x0000000000000000000000000000000000000000000000000000000000000000';
  }

  // Remove 0x prefix if present
  let cleaned = accountId.startsWith('0x') ? accountId.slice(2) : accountId;

  // Check length to determine if it's a valid address (40 hex chars) or hash (64 hex chars)
  const hexLength = cleaned.length;

  if (hexLength === 40 || hexLength === 64) {
    // Try to validate it's proper hex
    if (/^[0-9a-fA-F]+$/.test(cleaned)) {
      // Valid hex - pad accordingly
      if (hexLength === 40) {
        // Address - left-pad with zeros
        return pad(`0x${cleaned}` as Address, { size: 32 });
      } else {
        // Already 32-byte hash
        return `0x${cleaned}` as `0x${string}`;
      }
    }
  }

  // Not a standard address or hash, or contains invalid hex
  // Match Go's behavior: try to parse as hex, decode what we can
  // If odd length, prepend '0'
  if (cleaned.length % 2 === 1) {
    cleaned = '0' + cleaned;
  }

  // Decode valid hex pairs, stop at first invalid character
  const bytes: number[] = [];
  for (let i = 0; i < cleaned.length; i += 2) {
    const hexPair = cleaned.slice(i, i + 2);
    // Strictly validate that both characters are valid hex
    if (!/^[0-9a-fA-F]{2}$/.test(hexPair)) {
      // Invalid hex pair - stop here (matches Go's behavior)
      break;
    }
    bytes.push(parseInt(hexPair, 16));
  }

  // Convert to hex string and left-pad to 32 bytes (matches Go's behavior)
  const hexResult = bytes.map((b) => b.toString(16).padStart(2, '0')).join('');
  return pad(`0x${hexResult}` as `0x${string}`, { dir: 'left', size: 32 });
}

/**
 * Computes the metadata hash for a channel session key authorization.
 * Matches Go SDK's GetChannelSessionKeyAuthMetadataHashV1.
 *
 * @param version - Session key state version
 * @param assets - Asset symbols associated with the session key
 * @param expiresAt - Unix timestamp in seconds when the session key expires
 * @returns Keccak256 hash of the ABI-encoded metadata
 */
export function getChannelSessionKeyAuthMetadataHashV1(version: bigint, assets: string[], expiresAt: bigint): `0x${string}` {
  const packed = encodeAbiParameters(
    [
      { type: 'uint64' }, // version
      { type: 'string[]' }, // assets
      { type: 'uint64' }, // expires_at
    ],
    [version, assets, expiresAt],
  );
  return keccak256(packed);
}

/**
 * Packs a channel session key state for signing.
 * Matches Go SDK's PackChannelKeyStateV1.
 *
 * @param sessionKey - The session key address
 * @param metadataHash - The metadata hash from getChannelSessionKeyAuthMetadataHashV1
 * @returns ABI-encoded (sessionKey, metadataHash) ready for EIP-191 signing
 */
export function packChannelKeyStateV1(sessionKey: Address, metadataHash: `0x${string}`): `0x${string}` {
  return encodeAbiParameters(
    [
      { type: 'address' }, // session_key
      { type: 'bytes32' }, // hashed metadata
    ],
    [sessionKey, metadataHash],
  );
}
