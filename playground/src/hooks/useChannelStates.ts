import { useCallback, useEffect, useState } from 'react';
import type { Client, State, Channel } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';

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
      if (homeChannel && enforcedBalance != null) {
        setEnforced({ stateVersion: homeChannel.stateVersion, amount: enforcedBalance });
      } else {
        setEnforced(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsLoading(false);
    }
  }, [client, address, asset, enforcedBalance]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const canAcknowledge = !!issued && (signed === null || issued.version > signed.version);
  const canCheckpoint = !!signed && (enforced === null || signed.version > enforced.stateVersion);

  const acknowledge = useCallback(async () => {
    if (!client) return;
    setIsAcknowledging(true);
    try {
      await client.acknowledge(asset);
      await refresh();
      onAfterOp?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsAcknowledging(false);
    }
  }, [client, asset, refresh, onAfterOp]);

  const checkpoint = useCallback(async () => {
    if (!client) return;
    setIsCheckpointing(true);
    try {
      await client.checkpoint(asset);
      await refresh();
      onAfterOp?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsCheckpointing(false);
    }
  }, [client, asset, refresh, onAfterOp]);

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
