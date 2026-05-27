import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import type { Hex, Address } from 'viem';
import {
  Client,
  ChannelSessionKeyStateSigner,
  EthereumMsgSigner,
  getChannelSessionKeyAuthMetadataHashV1,
  type ChannelSessionKeyStateV1,
  type StateSigner,
} from '@yellow-org/sdk';

const EXPIRY_SECONDS = 24 * 60 * 60; // 24h delegation
const RENEW_BUFFER_SECONDS = 5 * 60; // treat as expired if within this margin

export interface StoredSessionKey {
  privateKey: Hex;
  sessionKeyAddress: Address;
  walletAddress: Address;
  version: string;       // bigint stringified
  assets: string[];
  expiresAt: string;     // unix seconds, bigint stringified
  userSig: Hex;          // raw EIP-191 (no signer type prefix)
  isActive?: boolean;    // true = this key is the current signing key
}

export function storageKeyFor(nodeUrl: string, walletAddress: Address): string {
  return `nitrolite_playground_sk_${nodeUrl}::${walletAddress.toLowerCase()}`;
}

export function isExpired(sk: StoredSessionKey): boolean {
  return Number(sk.expiresAt) - RENEW_BUFFER_SECONDS <= Math.floor(Date.now() / 1000);
}

export function secondsUntilExpiry(sk: StoredSessionKey): number {
  return Number(sk.expiresAt) - Math.floor(Date.now() / 1000);
}

// Internal helper: read the array from localStorage, migrating old single-key format if needed
function loadStoredArray(nodeUrl: string, walletAddress: Address): StoredSessionKey[] {
  const key = storageKeyFor(nodeUrl, walletAddress);
  const raw = localStorage.getItem(key);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    // Migration: if old format is a single object (not array), wrap it
    if (!Array.isArray(parsed)) {
      const single = parsed as StoredSessionKey;
      return [{ ...single, isActive: true }];
    }
    return parsed as StoredSessionKey[];
  } catch {
    return [];
  }
}

// Internal helper: write array back to localStorage
function saveStoredArray(nodeUrl: string, walletAddress: Address, keys: StoredSessionKey[]): void {
  localStorage.setItem(storageKeyFor(nodeUrl, walletAddress), JSON.stringify(keys));
}

// Load the active (signing) key. Returns null if none or if expired.
export function loadSessionKey(nodeUrl: string, walletAddress: Address): StoredSessionKey | null {
  const keys = loadStoredArray(nodeUrl, walletAddress);
  const active = keys.find(k => k.isActive);
  if (!active) return null;
  if (isExpired(active)) {
    // Clear the isActive flag but keep the entry for history display
    saveStoredArray(nodeUrl, walletAddress, keys.map(k =>
      k.sessionKeyAddress === active.sessionKeyAddress ? { ...k, isActive: false } : k
    ));
    return null;
  }
  return active;
}

// Load ALL stored keys (including expired/revoked, for display in Session Keys tab)
export function loadAllSessionKeys(nodeUrl: string, walletAddress: Address): StoredSessionKey[] {
  return loadStoredArray(nodeUrl, walletAddress);
}

// Save/update a key in the array (upserts by sessionKeyAddress). Marks it as active.
export function saveSessionKey(nodeUrl: string, sk: StoredSessionKey): void {
  const keys = loadStoredArray(nodeUrl, sk.walletAddress);
  // Mark all others inactive
  const updated = keys
    .filter(k => k.sessionKeyAddress !== sk.sessionKeyAddress)
    .map(k => ({ ...k, isActive: false }));
  updated.push({ ...sk, isActive: true });
  saveStoredArray(nodeUrl, sk.walletAddress, updated);
}

// Clear the active key flag (keeps the entry in history) — called when user clears
export function clearSessionKey(nodeUrl: string, walletAddress: Address): void {
  const keys = loadStoredArray(nodeUrl, walletAddress);
  saveStoredArray(nodeUrl, walletAddress, keys.map(k => ({ ...k, isActive: false })));
}

// Update a key's stored data (e.g. after revoke, to persist the new expiresAt) without
// marking it as active. Upserts by sessionKeyAddress.
export function saveKeyInactive(nodeUrl: string, sk: StoredSessionKey): void {
  const keys = loadStoredArray(nodeUrl, sk.walletAddress);
  const rest = keys.filter(k => k.sessionKeyAddress !== sk.sessionKeyAddress);
  rest.push({ ...sk, isActive: false });
  saveStoredArray(nodeUrl, sk.walletAddress, rest);
}

// Select a different key as the active signing key
export function selectKeyInStorage(nodeUrl: string, walletAddress: Address, sessionKeyAddress: Address): StoredSessionKey | null {
  const keys = loadStoredArray(nodeUrl, walletAddress);
  const target = keys.find(k => k.sessionKeyAddress.toLowerCase() === sessionKeyAddress.toLowerCase());
  if (!target || isExpired(target)) return null;
  saveStoredArray(nodeUrl, walletAddress, keys.map(k => ({
    ...k,
    isActive: k.sessionKeyAddress.toLowerCase() === sessionKeyAddress.toLowerCase(),
  })));
  return { ...target, isActive: true };
}

/**
 * Register a fresh session key. `client` must be wallet-backed (no SK signer active)
 * so that `signChannelSessionKeyState` produces a 0x00-prefixed wallet signature.
 * Pass a temporary wallet-only Client from the caller when an SK is active.
 */
export async function registerSessionKey(args: {
  client: Client;
  walletAddress: Address;
  assets: string[];
  nextVersion: bigint;
  expiresAt?: bigint;
}): Promise<StoredSessionKey> {
  const { client, walletAddress, assets, nextVersion } = args;

  const skPrivateKey = generatePrivateKey();
  const skAccount = privateKeyToAccount(skPrivateKey);
  const skMsgSigner = new EthereumMsgSigner(skPrivateKey);
  const expiresAtSec = args.expiresAt ?? BigInt(Math.floor(Date.now() / 1000) + EXPIRY_SECONDS);

  const state: ChannelSessionKeyStateV1 = {
    user_address: walletAddress,
    session_key: skAccount.address,
    version: nextVersion.toString(),
    assets,
    expires_at: expiresAtSec.toString(),
    user_sig: '',
    session_key_sig: '',
  };

  state.user_sig = await client.signChannelSessionKeyState(state);
  state.session_key_sig = await client.signChannelSessionKeyOwnership(state, skMsgSigner);
  await client.submitChannelSessionKeyState(state);

  return {
    privateKey: skPrivateKey,
    sessionKeyAddress: skAccount.address,
    walletAddress,
    version: state.version,
    assets,
    expiresAt: state.expires_at,
    userSig: state.user_sig as Hex,
  };
}

/**
 * Update an existing session key in-place (same address + private key, incremented version).
 * Used for renew and revoke. `client` must be wallet-backed — pass a temporary wallet-only
 * Client from the caller when an SK is active.
 */
export async function updateSessionKey(args: {
  client: Client;
  existingKey: StoredSessionKey;
  assets: string[];
  expiresAt: bigint;
}): Promise<StoredSessionKey> {
  const { client, existingKey, assets, expiresAt } = args;

  const nextVersion = BigInt(existingKey.version) + 1n;

  const state: ChannelSessionKeyStateV1 = {
    user_address: existingKey.walletAddress,
    session_key: existingKey.sessionKeyAddress,
    version: nextVersion.toString(),
    assets,
    expires_at: expiresAt.toString(),
    user_sig: '',
    session_key_sig: '',
  };

  const skMsgSigner = new EthereumMsgSigner(existingKey.privateKey);
  state.user_sig = await client.signChannelSessionKeyState(state);
  state.session_key_sig = await client.signChannelSessionKeyOwnership(state, skMsgSigner);
  await client.submitChannelSessionKeyState(state);

  return {
    ...existingKey,
    version: nextVersion.toString(),
    assets,
    expiresAt: expiresAt.toString(),
    userSig: state.user_sig as Hex,
  };
}

/**
 * Build the StateSigner the SDK Client should use when SK is active. The
 * session-key signer prepends the 0x01 type byte and embeds the user
 * authorization signature for server-side validation.
 */
export function buildSessionKeyStateSigner(sk: StoredSessionKey): StateSigner {
  const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
    sk.walletAddress,
    BigInt(sk.version),
    sk.assets,
    BigInt(sk.expiresAt),
  );
  return new ChannelSessionKeyStateSigner(sk.privateKey, sk.walletAddress, metadataHash, sk.userSig);
}
