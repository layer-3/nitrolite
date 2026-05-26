import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import type { Client, State, Channel } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address, Hash } from 'viem';
import { createPublicClient, http } from 'viem';
import { showErrorToast } from '../toastError';
import { rpcUrlFor } from '../networks';

export interface EnforcedState {
  stateVersion: bigint;
  amount: Decimal;
}

export interface UseChannelStatesResult {
  enforced: EnforcedState | null;
  signed: State | null;
  issued: State | null;
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  canAcknowledge: boolean;
  canCheckpoint: boolean;
  acknowledge: () => Promise<void>;
  checkpoint: () => Promise<void>;
  isAcknowledging: boolean;
  isCheckpointing: boolean;
}

export function useChannelStates(
  client: Client | null,
  address: Address | null,
  asset: string,
  enforcedBalance: Decimal | null | undefined,
  onAfterOp?: () => void,
): UseChannelStatesResult {
  const [enforced, setEnforced] = useState<EnforcedState | null>(null);
  const [signed, setSigned] = useState<State | null>(null);
  const [issued, setIssued] = useState<State | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isAcknowledging, setIsAcknowledging] = useState(false);
  const [isCheckpointing, setIsCheckpointing] = useState(false);

  // Track the last on-chain version we recorded so the enforced balance only
  // updates when a new checkpoint lands, not on every off-chain state change.
  const lastEnforcedVersionRef = useRef<bigint | null>(null);
  const lastEnforcedAmountRef = useRef<Decimal | null>(null);
  // Keep a ref to the prop so refresh() can read it without being recreated on
  // every enforcedBalance change.
  const enforcedBalancePropRef = useRef(enforcedBalance);
  useEffect(() => { enforcedBalancePropRef.current = enforcedBalance; }, [enforcedBalance]);

  const refresh = useCallback(async () => {
    if (!client || !address) {
      setEnforced(null);
      setSigned(null);
      setIssued(null);
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      const [homeChannel, signedState, issuedState] = await Promise.all([
        client.getHomeChannel(address, asset) as Promise<Channel | null>,
        client.getLatestState(address, asset, true),
        client.getLatestState(address, asset, false),
      ]);
      setSigned(signedState);
      setIssued(issuedState);
      if (homeChannel) {
        const v = homeChannel.stateVersion;
        if (lastEnforcedVersionRef.current !== v) {
          // On-chain version changed — a checkpoint was written. Record the new
          // enforced balance from the matching signed state if available, otherwise
          // fall back to the prop (covers the edge case of first load with
          // unconfirmed states already present).
          lastEnforcedVersionRef.current = v;
          lastEnforcedAmountRef.current =
            signedState?.version === v
              ? signedState.homeLedger.userBalance
              : (enforcedBalancePropRef.current ?? null);
        }
        // Use the cached amount so off-chain state changes (transfers, etc.)
        // never overwrite the enforced row until the next checkpoint.
        setEnforced({ stateVersion: v, amount: lastEnforcedAmountRef.current ?? new Decimal(0) });
      } else {
        setEnforced(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsLoading(false);
    }
  }, [client, address, asset, enforcedBalance]); // enforcedBalance triggers re-fetch; actual enforced amount is guarded by refs

  useEffect(() => {
    refresh();
  }, [refresh]);

  const canAcknowledge = !!issued && (signed === null || issued.version > signed.version);
  const canCheckpoint = !!signed && (enforced === null || signed.version > enforced.stateVersion);

  const handleOpError = (err: unknown, label: string) => {
    const e = err as { code?: number; message?: string };
    if (e?.code === 4001) {
      toast('Transaction cancelled');
      return;
    }
    showErrorToast(`${label} failed: ${e?.message ?? String(err)}`);
  };

  const acknowledge = useCallback(async () => {
    if (!client) return;
    setIsAcknowledging(true);
    try {
      await client.acknowledge(asset);
      await refresh();
      onAfterOp?.();
    } catch (err) {
      handleOpError(err, 'Acknowledge');
    } finally {
      setIsAcknowledging(false);
    }
  }, [client, asset, refresh, onAfterOp]);

  const checkpoint = useCallback(async () => {
    if (!client) return;
    setIsCheckpointing(true);
    try {
      const txHash = await client.checkpoint(asset);
      // Wait for the tx to be mined so the enforced state and on-chain balance
      // reflect the checkpoint when we refresh.
      const blockchainId = signed?.homeLedger.blockchainId;
      if (blockchainId && txHash) {
        const rpcUrl = rpcUrlFor(blockchainId);
        if (rpcUrl) {
          const publicClient = createPublicClient({ transport: http(rpcUrl) });
          await publicClient.waitForTransactionReceipt({ hash: txHash as Hash });
        }
      }
      await refresh();
      onAfterOp?.();
    } catch (err) {
      handleOpError(err, 'Checkpoint');
    } finally {
      setIsCheckpointing(false);
    }
  }, [client, asset, signed, refresh, onAfterOp]);

  return {
    enforced,
    signed,
    issued,
    isLoading,
    error,
    refresh,
    canAcknowledge,
    canCheckpoint,
    acknowledge,
    checkpoint,
    isAcknowledging,
    isCheckpointing,
  };
}
