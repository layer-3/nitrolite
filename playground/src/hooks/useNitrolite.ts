import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Client,
  ChannelDefaultSigner,
  withBlockchainRPC,
  type StateSigner,
  type TransactionSigner,
} from '@yellow-org/sdk';
import type { Asset, Blockchain } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address, WalletClient } from 'viem';
import { WalletStateSigner, WalletTransactionSigner } from '../walletSigners';
import { rpcUrlFor } from '../networks';
import { buildSessionKeyStateSigner, type StoredSessionKey } from '../sessionKey';

const NODE_URL = import.meta.env.VITE_NITRONODE_URL ?? 'wss://nitronode-sandbox.yellow.org/v1/ws';

export interface UseNitroliteResult {
  client: Client | null;
  isConnecting: boolean;
  isConnected: boolean;
  lastCommsAt: Date | null;
  nodeError: string | null;
  supportedAssets: Asset[];
  supportedChains: Blockchain[];
  balances: Record<string, Decimal>;
  onChainBalances: Record<string, Decimal | null>;
  isSessionKeyActive: boolean;
  refresh: () => Promise<void>;
  touch: () => void;
}

export function useNitrolite(
  address: Address | null,
  walletClient: WalletClient | null,
  sessionKey: StoredSessionKey | null,
): UseNitroliteResult {
  const [client, setClient] = useState<Client | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [lastCommsAt, setLastCommsAt] = useState<Date | null>(null);
  const [nodeError, setNodeError] = useState<string | null>(null);
  const [supportedAssets, setSupportedAssets] = useState<Asset[]>([]);
  const [supportedChains, setSupportedChains] = useState<Blockchain[]>([]);
  const [balances, setBalances] = useState<Record<string, Decimal>>({});
  const [onChainBalances, setOnChainBalances] = useState<Record<string, Decimal | null>>({});

  const clientRef = useRef<Client | null>(null);

  const touch = useCallback(() => setLastCommsAt(new Date()), []);

  const refresh = useCallback(async () => {
    const c = clientRef.current;
    if (!c || !address) return;
    try {
      const entries = await c.getBalances(address);
      const next: Record<string, Decimal> = {};
      for (const e of entries) next[e.asset] = e.balance;
      setBalances(next);
      touch();
    } catch (err) {
      setNodeError(err instanceof Error ? err.message : String(err));
    }
  }, [address, touch]);

  // Lifecycle: rebuild client when address or walletClient changes
  useEffect(() => {
    if (!address || !walletClient) {
      // teardown
      const prev = clientRef.current;
      clientRef.current = null;
      setClient(null);
      setIsConnected(false);
      if (prev) prev.close().catch(() => {});
      return;
    }

    let cancelled = false;
    setIsConnecting(true);
    setNodeError(null);

    const timer = setTimeout(async () => {
      // Tear down any existing client before building a new one
      const prev = clientRef.current;
      clientRef.current = null;
      if (prev) {
        try {
          await prev.close();
        } catch {
          /* ignore */
        }
      }

      try {
        // State signer choice: session key when active (no MetaMask popup), else
        // wallet-backed default. txSigner is always wallet-backed because on-chain
        // txs (deposit/withdraw/checkpoint/approve) must come from the user.
        let stateSigner: StateSigner;
        if (sessionKey && sessionKey.walletAddress.toLowerCase() === address.toLowerCase()) {
          stateSigner = buildSessionKeyStateSigner(sessionKey);
        } else {
          // Wrap in ChannelDefaultSigner so the SDK prepends the 0x00 type byte that
          // the nitronode expects (raw EIP-191 sigs are rejected as "signature type 28").
          const walletSigner = new WalletStateSigner(walletClient) as unknown as StateSigner;
          stateSigner = new ChannelDefaultSigner(walletSigner);
        }
        const txSigner = new WalletTransactionSigner(walletClient) as unknown as TransactionSigner;

        // Build a temporary client to discover supported chains, then rebuild with their RPCs.
        // The first connect just needs *some* RPC; we'll add them per chain after getConfig().
        const probe = await Client.create(NODE_URL, stateSigner, txSigner);
        if (cancelled) {
          await probe.close().catch(() => {});
          return;
        }

        const cfg = await probe.getConfig();
        const chains = cfg.blockchains;
        const assets = await probe.getAssets();
        if (cancelled) {
          await probe.close().catch(() => {});
          return;
        }
        await probe.close().catch(() => {});

        // Build final client with all RPC options applied
        const opts = chains
          .map(c => rpcUrlFor(c.id))
          .map((url, i) => (url ? withBlockchainRPC(chains[i].id, url) : null))
          .filter((o): o is NonNullable<typeof o> => o !== null);

        const finalClient = await Client.create(NODE_URL, stateSigner, txSigner, ...opts);
        if (cancelled) {
          await finalClient.close().catch(() => {});
          return;
        }

        clientRef.current = finalClient;
        setClient(finalClient);
        setSupportedAssets(assets);
        setSupportedChains(chains);
        setIsConnected(true);
        setIsConnecting(false);
        touch();

        // Initial balances
        const entries = await finalClient.getBalances(address);
        const nextBal: Record<string, Decimal> = {};
        for (const e of entries) nextBal[e.asset] = e.balance;
        if (!cancelled) {
          setBalances(nextBal);
          touch();

          // On-chain balances per asset on its suggestedBlockchainId. Failures → null.
          const ocb: Record<string, Decimal | null> = {};
          await Promise.all(
            assets.map(async a => {
              try {
                const bal = await finalClient.getOnChainBalance(a.suggestedBlockchainId, a.symbol, address);
                ocb[a.symbol] = bal;
              } catch {
                ocb[a.symbol] = null;
              }
            }),
          );
          if (!cancelled) setOnChainBalances(ocb);
        }
      } catch (err) {
        if (!cancelled) {
          setNodeError(err instanceof Error ? err.message : String(err));
          setIsConnecting(false);
          setIsConnected(false);
        }
      }
    }, 200); // debounce per plan H-2

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [address, walletClient, sessionKey, touch]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      const prev = clientRef.current;
      if (prev) prev.close().catch(() => {});
    };
  }, []);

  const stableBalances = useMemo(() => balances, [JSON.stringify(serializable(balances))]);
  const stableOnChain = useMemo(() => onChainBalances, [JSON.stringify(serializable(onChainBalances))]);

  const isSessionKeyActive =
    !!sessionKey && !!address && sessionKey.walletAddress.toLowerCase() === address.toLowerCase();

  return {
    client,
    isConnecting,
    isConnected,
    lastCommsAt,
    nodeError,
    supportedAssets,
    supportedChains,
    balances: stableBalances,
    onChainBalances: stableOnChain,
    isSessionKeyActive,
    refresh,
    touch,
  };
}

function serializable(rec: Record<string, Decimal | null>): Record<string, string | null> {
  const out: Record<string, string | null> = {};
  for (const [k, v] of Object.entries(rec)) out[k] = v ? v.toString() : null;
  return out;
}
