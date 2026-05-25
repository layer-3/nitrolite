import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import type { Address } from 'viem';
import type { Client } from '@yellow-org/sdk';
import {
  loadSessionKey,
  saveSessionKey,
  clearSessionKey,
  registerSessionKey,
  type StoredSessionKey,
} from '../sessionKey';

const NODE_URL = import.meta.env.VITE_NITRONODE_URL ?? 'wss://nitronode-sandbox.yellow.org/v1/ws';

export interface UseSessionKeyResult {
  sessionKey: StoredSessionKey | null;
  isRegistering: boolean;
  register: (client: Client, assetSymbols: string[]) => Promise<void>;
  clear: () => void;
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

  // Load on address change. loadSessionKey clears expired entries internally.
  useEffect(() => {
    if (!address) {
      setSessionKey(null);
      return;
    }
    setSessionKey(loadSessionKey(NODE_URL, address));
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
    }, 60_000);
    return () => clearInterval(id);
  }, [address]);

  const register = useCallback(
    async (client: Client, assetSymbols: string[]) => {
      if (!address || !assetSymbols.length) {
        toast.error('Cannot register: missing address or assets');
        return;
      }
      setIsRegistering(true);
      try {
        // Find next version. Include inactive so we don't collide with prior
        // (expired or revoked) registrations recorded on the node.
        let nextVersion = 1n;
        try {
          const existing = await client.getLastChannelKeyStates(address, undefined, {
            includeInactive: true,
          });
          if (existing.length) {
            const highest = existing.reduce((max, s) => {
              const v = BigInt(s.version);
              return v > max ? v : max;
            }, 0n);
            nextVersion = highest + 1n;
          }
        } catch {
          // First-ever registration may legitimately throw; fall back to 1.
        }

        const sk = await registerSessionKey({
          client,
          walletAddress: address,
          assets: assetSymbols,
          nextVersion,
        });
        saveSessionKey(NODE_URL, sk);
        setSessionKey(sk);
        toast.success('Session key active — state ops will no longer prompt MetaMask');
      } catch (err) {
        const e = err as { code?: number; message?: string };
        if (e?.code === 4001) toast('Cancelled');
        else toast.error(`Session key setup failed: ${e?.message ?? String(err)}`);
      } finally {
        setIsRegistering(false);
      }
    },
    [address],
  );

  const clear = useCallback(() => {
    if (address) clearSessionKey(NODE_URL, address);
    setSessionKey(null);
    toast('Session key cleared');
  }, [address]);

  return { sessionKey, isRegistering, register, clear };
}
