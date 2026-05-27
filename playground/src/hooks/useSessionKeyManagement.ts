import { useState, useCallback } from 'react';
import { toast } from 'sonner';
import { showErrorToast } from '../toastError';
import type { Address, WalletClient } from 'viem';
import {
  Client,
  ChannelDefaultSigner,
  type StateSigner,
  type TransactionSigner,
  type ChannelSessionKeyStateV1,
} from '@yellow-org/sdk';
import {
  registerSessionKey,
  updateSessionKey,
  loadAllSessionKeys,
  saveSessionKey,
  saveKeyInactive,
  type StoredSessionKey,
} from '../sessionKey';
import { WalletStateSigner, WalletTransactionSigner } from '../walletSigners';

const NODE_URL = import.meta.env.VITE_NITRONODE_URL ?? 'wss://nitronode-sandbox.yellow.org/v1/ws';

interface UseSessionKeyManagementResult {
  serverKeys: ChannelSessionKeyStateV1[];
  isLoading: boolean;
  isSubmitting: boolean;
  fetchKeys: () => Promise<void>;
  register: (walletAddress: Address, assets: string[], expiresAt: bigint) => Promise<StoredSessionKey | null>;
  update: (walletAddress: Address, currentKey: ChannelSessionKeyStateV1, assets: string[], expiresAt: bigint) => Promise<StoredSessionKey | null>;
  revoke: (walletAddress: Address, currentKey: ChannelSessionKeyStateV1) => Promise<void>;
}

export function useSessionKeyManagement(
  client: Client | null,
  address: Address | null,
  walletClient: WalletClient | null,
): UseSessionKeyManagementResult {
  const [serverKeys, setServerKeys] = useState<ChannelSessionKeyStateV1[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fetchKeys = useCallback(async () => {
    if (!client || !address) return;
    setIsLoading(true);
    try {
      const keys = await client.getLastChannelKeyStates(address, undefined, { includeInactive: true });
      setServerKeys(keys);
    } catch (err) {
      showErrorToast(`Failed to load session keys: ${(err as Error).message ?? String(err)}`);
    } finally {
      setIsLoading(false);
    }
  }, [client, address]);

  // Build a temporary wallet-backed Client so signChannelSessionKeyState always
  // uses the 0x00 wallet signer regardless of what the main client's signer is.
  const buildWalletOnlyClient = useCallback(async (): Promise<Client | null> => {
    if (!walletClient) return null;
    try {
      const stateSigner = new ChannelDefaultSigner(
        new WalletStateSigner(walletClient) as unknown as StateSigner,
      );
      const txSigner = new WalletTransactionSigner(walletClient) as unknown as TransactionSigner;
      return await Client.create(NODE_URL, stateSigner, txSigner);
    } catch {
      return null;
    }
  }, [walletClient]);

  const register = useCallback(async (
    walletAddress: Address,
    assets: string[],
    expiresAt: bigint,
  ): Promise<StoredSessionKey | null> => {
    if (!walletClient) { showErrorToast('Wallet not connected'); return null; }
    setIsSubmitting(true);
    const signingClient = await buildWalletOnlyClient();
    if (!signingClient) { showErrorToast('Could not create signing client'); setIsSubmitting(false); return null; }
    try {
      const sk = await registerSessionKey({ client: signingClient, walletAddress, assets, nextVersion: 1n, expiresAt });
      saveSessionKey(NODE_URL, sk);
      toast.success('Session key registered');
      await fetchKeys();
      return sk;
    } catch (err) {
      const e = err as { code?: number; message?: string };
      if (e?.code === 4001) toast('Cancelled');
      else showErrorToast(`Registration failed: ${e?.message ?? String(err)}`);
      return null;
    } finally {
      setIsSubmitting(false);
      signingClient.close().catch(() => {});
    }
  }, [walletClient, buildWalletOnlyClient, fetchKeys]);

  const update = useCallback(async (
    walletAddress: Address,
    currentKey: ChannelSessionKeyStateV1,
    assets: string[],
    expiresAt: bigint,
  ): Promise<StoredSessionKey | null> => {
    if (!walletClient) { showErrorToast('Wallet not connected'); return null; }
    const localKey = loadAllSessionKeys(NODE_URL, walletAddress)
      .find(k => k.sessionKeyAddress.toLowerCase() === currentKey.session_key.toLowerCase());
    if (!localKey) { showErrorToast('Session key not found in local storage'); return null; }
    setIsSubmitting(true);
    const signingClient = await buildWalletOnlyClient();
    if (!signingClient) { showErrorToast('Could not create signing client'); setIsSubmitting(false); return null; }
    try {
      const sk = await updateSessionKey({ client: signingClient, existingKey: localKey, assets, expiresAt });
      saveSessionKey(NODE_URL, sk);
      toast.success('Session key updated');
      await fetchKeys();
      return sk;
    } catch (err) {
      const e = err as { code?: number; message?: string };
      if (e?.code === 4001) toast('Cancelled');
      else showErrorToast(`Update failed: ${e?.message ?? String(err)}`);
      return null;
    } finally {
      setIsSubmitting(false);
      signingClient.close().catch(() => {});
    }
  }, [walletClient, buildWalletOnlyClient, fetchKeys]);

  const revoke = useCallback(async (
    walletAddress: Address,
    currentKey: ChannelSessionKeyStateV1,
  ): Promise<void> => {
    if (!walletClient) { showErrorToast('Wallet not connected'); return; }
    const localKey = loadAllSessionKeys(NODE_URL, walletAddress)
      .find(k => k.sessionKeyAddress.toLowerCase() === currentKey.session_key.toLowerCase());
    if (!localKey) { showErrorToast('Session key not found in local storage'); return; }
    setIsSubmitting(true);
    const signingClient = await buildWalletOnlyClient();
    if (!signingClient) { showErrorToast('Could not create signing client'); setIsSubmitting(false); return; }
    try {
      const expiredAt = BigInt(Math.floor(Date.now() / 1000) - 1);
      const revokedSk = await updateSessionKey({ client: signingClient, existingKey: localKey, assets: localKey.assets, expiresAt: expiredAt });
      // Persist expired expiresAt to localStorage so allSessionKeys reflects the change immediately.
      saveKeyInactive(NODE_URL, revokedSk);
      toast.success('Session key revoked');
      await fetchKeys();
    } catch (err) {
      const e = err as { code?: number; message?: string };
      if (e?.code === 4001) toast('Cancelled');
      else showErrorToast(`Revoke failed: ${e?.message ?? String(err)}`);
    } finally {
      setIsSubmitting(false);
      signingClient.close().catch(() => {});
    }
  }, [walletClient, buildWalletOnlyClient, fetchKeys]);

  return { serverKeys, isLoading, isSubmitting, fetchKeys, register, update, revoke };
}
