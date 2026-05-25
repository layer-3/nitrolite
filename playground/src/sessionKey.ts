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

export function loadSessionKey(nodeUrl: string, walletAddress: Address): StoredSessionKey | null {
  const key = storageKeyFor(nodeUrl, walletAddress);
  const raw = localStorage.getItem(key);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as StoredSessionKey;
    if (isExpired(parsed)) {
      localStorage.removeItem(key);
      return null;
    }
    return parsed;
  } catch {
    localStorage.removeItem(key);
    return null;
  }
}

export function saveSessionKey(nodeUrl: string, sk: StoredSessionKey): void {
  localStorage.setItem(storageKeyFor(nodeUrl, sk.walletAddress), JSON.stringify(sk));
}

export function clearSessionKey(nodeUrl: string, walletAddress: Address): void {
  localStorage.removeItem(storageKeyFor(nodeUrl, walletAddress));
}

/**
 * Register a fresh session key against the node using a wallet-backed Client.
 * Returns the persisted record. Caller must already have a Client whose state
 * signer is the wallet (i.e. SK is not active yet) so `signChannelSessionKeyState`
 * pops MetaMask exactly once.
 */
export async function registerSessionKey(args: {
  client: Client;
  walletAddress: Address;
  assets: string[];
  nextVersion: bigint;
}): Promise<StoredSessionKey> {
  const { client, walletAddress, assets, nextVersion } = args;

  const skPrivateKey = generatePrivateKey();
  const skAccount = privateKeyToAccount(skPrivateKey);
  const skMsgSigner = new EthereumMsgSigner(skPrivateKey);

  const expiresAtSec = BigInt(Math.floor(Date.now() / 1000) + EXPIRY_SECONDS);

  const state: ChannelSessionKeyStateV1 = {
    user_address: walletAddress,
    session_key: skAccount.address,
    version: nextVersion.toString(),
    assets,
    expires_at: expiresAtSec.toString(),
    user_sig: '',
    session_key_sig: '',
  };

  // user_sig — MetaMask popup (single, EIP-191). SDK strips the signer-type prefix.
  state.user_sig = await client.signChannelSessionKeyState(state);
  // session_key_sig — purely local (no popup).
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
