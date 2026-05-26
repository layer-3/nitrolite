import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import type { Client, Asset } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';

interface PendingTransfer {
  to: Address;
  asset: string;
  amount: Decimal;
}

export interface UseChannelOpsResult {
  deposit: (blockchainId: bigint, asset: string, amount: Decimal) => Promise<void>;
  withdraw: (blockchainId: bigint, asset: string, amount: Decimal) => Promise<void>;
  transfer: (to: Address, asset: string, amount: Decimal) => Promise<void>;
  closeChannel: (asset: string) => Promise<void>;
  isApproving: boolean;
  isDepositing: boolean;
  isWithdrawing: boolean;
  isTransferring: boolean;
  isClosing: boolean;
  homechainModalAsset: string | null;
  onHomechainSelected: (asset: string, chainId: bigint) => Promise<void>;
  onHomechainModalDismiss: () => void;
}

export function useChannelOps(
  client: Client | null,
  address: Address | null,
  supportedAssets: Asset[],
  onAfterOp?: () => void,
): UseChannelOpsResult {
  const [isApproving, setIsApproving] = useState(false);
  const [isDepositing, setIsDepositing] = useState(false);
  const [isWithdrawing, setIsWithdrawing] = useState(false);
  const [isTransferring, setIsTransferring] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [homechainModalAsset, setHomechainModalAsset] = useState<string | null>(null);
  const pendingTransferRef = useRef<PendingTransfer | null>(null);

  const generationRef = useRef(0);
  useEffect(() => {
    generationRef.current += 1;
  }, [address]);

  const tokenInfoFor = useCallback(
    (asset: string, chainId: bigint): { address: string; decimals: number } | undefined => {
      const a = supportedAssets.find(x => x.symbol === asset);
      const token = a?.tokens.find(t => t.blockchainId === chainId);
      if (!token) return undefined;
      return { address: token.address, decimals: token.decimals };
    },
    [supportedAssets],
  );

  const handleError = (err: unknown, label: string) => {
    // EIP-1193 user rejection: code 4001
    const e = err as { code?: number; message?: string };
    if (e?.code === 4001) {
      toast('Transaction cancelled');
      return;
    }
    const msg = e?.message ?? String(err);
    toast.error(`${label} failed: ${msg}`);
  };

  const deposit = useCallback(
    async (blockchainId: bigint, asset: string, amount: Decimal) => {
      if (!client || !address) return;
      const gen = generationRef.current;
      setIsDepositing(true);
      try {
        const tokenInfo = tokenInfoFor(asset, blockchainId);
        if (!tokenInfo) throw new Error(`token not found for ${asset} on chain ${blockchainId}`);

        // Native tokens (address 0x0…0) don't need allowance/approve.
        const isNative = /^0x0+$/i.test(tokenInfo.address);
        if (!isNative) {
          // checkTokenAllowance returns smallest-unit bigint; convert deposit amount to the same scale.
          const allowance = await client.checkTokenAllowance(blockchainId, tokenInfo.address, address);
          if (generationRef.current !== gen) {
            toast('Wallet changed — operation cancelled');
            return;
          }
          const amountUnits = BigInt(
            amount.mul(new Decimal(10).pow(tokenInfo.decimals)).floor().toFixed(0),
          );
          if (allowance < amountUnits) {
            // Approve effectively-unlimited once so subsequent deposits skip the
            // approval popup. SDK's approveToken accepts Decimal in human units
            // and multiplies by 10^decimals internally; pass floor(MaxUint256 /
            // 10^decimals) so the scaled bigint never exceeds uint256.
            const maxUnits = (1n << 256n) - 1n;
            const divisor = 10n ** BigInt(tokenInfo.decimals);
            const humanMax = new Decimal((maxUnits / divisor).toString());
            toast('Approving token (one-time)…');
            setIsApproving(true);
            await client.approveToken(blockchainId, asset, humanMax);
            setIsApproving(false);
            if (generationRef.current !== gen) {
              toast('Wallet changed — operation cancelled');
              return;
            }
          }
        }

        toast('Depositing…');
        await client.deposit(blockchainId, asset, amount);
        if (generationRef.current !== gen) return;
        await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        toast.success(`Deposited ${amount.toString()} ${asset}`);
        onAfterOp?.();
      } catch (err) {
        handleError(err, 'Deposit');
      } finally {
        setIsApproving(false);
        setIsDepositing(false);
      }
    },
    [client, address, tokenInfoFor, onAfterOp],
  );

  const withdraw = useCallback(
    async (blockchainId: bigint, asset: string, amount: Decimal) => {
      if (!client || !address) return;
      const gen = generationRef.current;
      setIsWithdrawing(true);
      try {
        toast('Withdrawing…');
        await client.withdraw(blockchainId, asset, amount);
        if (generationRef.current !== gen) return;
        await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        toast.success(`Withdrew ${amount.toString()} ${asset}`);
        onAfterOp?.();
      } catch (err) {
        handleError(err, 'Withdraw');
      } finally {
        setIsWithdrawing(false);
      }
    },
    [client, address, onAfterOp],
  );

  const performTransfer = useCallback(
    async (to: Address, asset: string, amount: Decimal): Promise<boolean> => {
      if (!client) return false;
      const gen = generationRef.current;
      setIsTransferring(true);
      try {
        await client.transfer(to, asset, amount);
        if (generationRef.current !== gen) return false;
        toast.success(`Transferred ${amount.toString()} ${asset}`);
        onAfterOp?.();
        return true;
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        // Heuristic: no home blockchain configured for this asset.
        if (msg.toLowerCase().includes('home blockchain') && msg.toLowerCase().includes('not set')) {
          pendingTransferRef.current = { to, asset, amount };
          setHomechainModalAsset(asset);
          return false;
        }
        handleError(err, 'Transfer');
        return false;
      } finally {
        setIsTransferring(false);
      }
    },
    [client, onAfterOp],
  );

  const transfer = useCallback(
    async (to: Address, asset: string, amount: Decimal) => {
      await performTransfer(to, asset, amount);
    },
    [performTransfer],
  );

  const onHomechainSelected = useCallback(
    async (asset: string, chainId: bigint) => {
      if (!client) return;
      try {
        await client.setHomeBlockchain(asset, chainId);
        toast.success(`Home chain set for ${asset}`);
        const pending = pendingTransferRef.current;
        pendingTransferRef.current = null;
        setHomechainModalAsset(null);
        if (pending) {
          await performTransfer(pending.to, pending.asset, pending.amount);
        }
      } catch (err) {
        handleError(err, 'Set home chain');
      }
    },
    [client, performTransfer],
  );

  const onHomechainModalDismiss = useCallback(() => {
    pendingTransferRef.current = null;
    setHomechainModalAsset(null);
  }, []);

  const closeChannel = useCallback(
    async (asset: string) => {
      if (!client) return;
      const gen = generationRef.current;
      setIsClosing(true);
      try {
        toast('Closing channel…');
        await client.closeHomeChannel(asset);
        if (generationRef.current !== gen) return;
        await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        toast.success(`Closed channel for ${asset}`);
        onAfterOp?.();
      } catch (err) {
        handleError(err, 'Close');
      } finally {
        setIsClosing(false);
      }
    },
    [client, onAfterOp],
  );

  return {
    deposit,
    withdraw,
    transfer,
    closeChannel,
    isApproving,
    isDepositing,
    isWithdrawing,
    isTransferring,
    isClosing,
    homechainModalAsset,
    onHomechainSelected,
    onHomechainModalDismiss,
  };
}
