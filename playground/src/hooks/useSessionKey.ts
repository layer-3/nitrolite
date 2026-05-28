import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { showErrorToast } from "../toastError";
import type { Address } from 'viem';
import type { Client } from '@yellow-org/sdk';
import {
  loadSessionKey,
  loadAllSessionKeys,
  saveSessionKey,
  clearSessionKey,
  selectKeyInStorage,
  registerSessionKey,
  type StoredSessionKey,
} from '../sessionKey';
import { NODE_URL } from '../config';

export interface UseSessionKeyResult {
  sessionKey: StoredSessionKey | null;
  allKeys: StoredSessionKey[];
  isRegistering: boolean;
  register: (client: Client, assetSymbols: string[], expiresAt?: bigint) => Promise<void>;
  selectKey: (sessionKeyAddress: Address) => void;
  clear: () => void;
  refreshAllKeys: () => void;
}

/**
 * Owns the local session-key state. `register` is supplied the SDK Client at
 * call time because the same Client must be used for signing (it must be the
 * wallet-backed one — registering while a session-key client is active would
 * cause the SK to sign its own authorization, which the node rejects).
 */
export function useSessionKey(address: Address | null): UseSessionKeyResult {
  const [sessionKey, setSessionKey] = useState<StoredSessionKey | null>(null);
  const [isRegistering, setIsRegistering] = useState(false);
  const [allKeys, setAllKeys] = useState<StoredSessionKey[]>([]);

  // Load on address change. loadSessionKey clears expired entries internally.
  useEffect(() => {
    if (!address) {
      setSessionKey(null);
      setAllKeys([]);
      return;
    }
    setSessionKey(loadSessionKey(NODE_URL, address));
    setAllKeys(loadAllSessionKeys(NODE_URL, address));
  }, [address]);

  // Re-check expiry once a minute so the chip can flip to "expired"/banner can
  // re-appear without requiring a page reload.
  useEffect(() => {
    if (!address) return;
    const id = setInterval(() => {
      const fresh = loadSessionKey(NODE_URL, address);
      setSessionKey(prev => {
        if (prev?.privateKey === fresh?.privateKey) return prev;
        return fresh;
      });
      setAllKeys(loadAllSessionKeys(NODE_URL, address));
    }, 60_000);
    return () => clearInterval(id);
  }, [address]);

  const register = useCallback(
    async (client: Client, assetSymbols: string[], expiresAt?: bigint) => {
      if (!address || !assetSymbols.length) {
        showErrorToast('Cannot register: missing address or assets');
        return;
      }
      setIsRegistering(true);
      try {
        const sk = await registerSessionKey({
          client,
          walletAddress: address,
          assets: assetSymbols,
          nextVersion: 1n,
          expiresAt,
        });
        saveSessionKey(NODE_URL, sk);
        setSessionKey(sk);
        setAllKeys(loadAllSessionKeys(NODE_URL, address));
        toast.success('Session key active — state ops will no longer prompt MetaMask');
      } catch (err) {
        const e = err as { code?: number; message?: string };
        if (e?.code === 4001) toast('Cancelled');
        else showErrorToast(`Session key setup failed: ${e?.message ?? String(err)}`);
      } finally {
        setIsRegistering(false);
      }
    },
    [address],
  );

  const selectKey = useCallback((sessionKeyAddress: Address) => {
    if (!address) return;
    const selected = selectKeyInStorage(NODE_URL, address, sessionKeyAddress);
    if (selected) {
      setSessionKey(selected);
      setAllKeys(loadAllSessionKeys(NODE_URL, address));
      toast(`Switched to session key ${sessionKeyAddress.slice(0, 6)}…${sessionKeyAddress.slice(-4)}`);
    }
  }, [address]);

  const clear = useCallback(() => {
    if (address) clearSessionKey(NODE_URL, address);
    setSessionKey(null);
    if (address) setAllKeys(loadAllSessionKeys(NODE_URL, address));
    toast('Session key cleared');
  }, [address]);

  const refreshAllKeys = useCallback(() => {
    if (address) setAllKeys(loadAllSessionKeys(NODE_URL, address));
  }, [address]);

  return { sessionKey, allKeys, isRegistering, register, selectKey, clear, refreshAllKeys };
}
