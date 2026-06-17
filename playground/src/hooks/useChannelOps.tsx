import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { showErrorToast } from "../toastError";
import type { Client, Asset, Blockchain } from '@yellow-org/sdk';
import { TransitionType } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address, Hash } from 'viem';
import { createPublicClient, http } from 'viem';
import { rpcUrlFor } from '../networks';

interface PendingTransfer {
  to: Address;
  asset: string;
  amount: Decimal;
}

export type DepositPhase = 'idle' | 'approving' | 'signing_state' | 'signing_tx' | 'confirming' | 'awaiting_node';
export type WithdrawPhase = 'idle' | 'signing_state' | 'signing_tx' | 'confirming' | 'awaiting_node';
export type TransferPhase = 'idle' | 'signing_state';

export interface UseChannelOpsResult {
  deposit: (blockchainId: bigint, asset: string, amount: Decimal) => Promise<void>;
  withdraw: (blockchainId: bigint, asset: string, amount: Decimal) => Promise<void>;
  transfer: (to: Address, asset: string, amount: Decimal) => Promise<void>;
  closeChannel: (asset: string, blockchainId: bigint) => Promise<void>;
  depositPhase: DepositPhase;
  withdrawPhase: WithdrawPhase;
  transferPhase: TransferPhase;
  needsApproval: boolean | null;
  checkDepositAllowance: (blockchainId: bigint, asset: string, amount: Decimal) => Promise<void>;
  closingAsset: string | null;
  /** Asset currently in the post-receipt confirmation-window wait after a close tx. */
  awaitingCloseAsset: string | null;
  homechainModalAsset: string | null;
  onHomechainSelected: (asset: string, chainId: bigint) => Promise<void>;
  onHomechainModalDismiss: () => void;
}

export function useChannelOps(
  client: Client | null,
  address: Address | null,
  supportedAssets: Asset[],
  supportedChains: Blockchain[],
  onAfterOp?: () => void,
  onAfterTxMined?: () => void,
): UseChannelOpsResult {
  const [depositPhase, setDepositPhase] = useState<DepositPhase>('idle');
  const [withdrawPhase, setWithdrawPhase] = useState<WithdrawPhase>('idle');
  const [transferPhase, setTransferPhase] = useState<TransferPhase>('idle');
  const [needsApproval, setNeedsApproval] = useState<boolean | null>(null);
  const [closingAsset, setClosingAsset] = useState<string | null>(null);
  const [awaitingCloseAsset, setAwaitingCloseAsset] = useState<string | null>(null);
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

  const checkDepositAllowance = useCallback(
    async (blockchainId: bigint, asset: string, amount: Decimal) => {
      if (!client || !address || amount.lte(0)) { setNeedsApproval(null); return; }
      try {
        const tokenInfo = tokenInfoFor(asset, blockchainId);
        if (!tokenInfo || /^0x0+$/i.test(tokenInfo.address)) { setNeedsApproval(false); return; }
        const allowance = await client.checkTokenAllowance(blockchainId, tokenInfo.address, address);
        const amountUnits = BigInt(
          amount.mul(new Decimal(10).pow(tokenInfo.decimals)).floor().toFixed(0),
        );
        setNeedsApproval(allowance < amountUnits);
      } catch {
        setNeedsApproval(null);
      }
    },
    [client, address, tokenInfoFor],
  );

  // Cap active polling so a large hard-finality delay (e.g. 780s) never holds the spinner that
  // long. Covers the production "quick" values (Eth 36s, Polygon 5s, BNB 2s) in full; for larger
  // delays we fall back to the soft toast and let the normal refresh surface the credit later.
  const MAX_ACTIVE_WAIT_SECS = 60;

  const waitForCredit = useCallback(
    async (asset: string, delaySecs: number, baselineEnforced: Decimal | null, direction: 'up' | 'down', gen: number): Promise<boolean> => {
      if (!client || !address || delaySecs <= 0) return true; // gate disabled → nothing to wait for
      // No reliable baseline (the pre-op read failed): we cannot tell a pre-existing balance
      // from a freshly-credited one, so don't guess. Fall back to the soft toast.
      if (baselineEnforced === null) return false;
      const activeWaitSecs = Math.min(delaySecs, MAX_ACTIVE_WAIT_SECS);
      const deadline = Date.now() + (activeWaitSecs + 15) * 1000; // capped delay + buffer for block time / RPC lag
      const intervalMs = 2000;
      while (Date.now() < deadline) {
        if (generationRef.current !== gen) return false; // wallet changed mid-wait
        await new Promise(r => setTimeout(r, intervalMs));
        if (generationRef.current !== gen) return false;
        try {
          const entries = await client.getBalances(address);
          const e = entries.find(x => x.asset === asset);
          if (e) {
            const moved = direction === 'up' ? e.enforced.gt(baselineEnforced) : e.enforced.lt(baselineEnforced);
            if (moved) return true;
          }
        } catch { /* transient RPC; keep polling until deadline */ }
      }
      return false; // timed out — caller shows a softer message
    },
    [client, address],
  );

  const handleError = (err: unknown, label: string) => {
    // EIP-1193 user rejection: code 4001
    const e = err as { code?: number; message?: string };
    if (e?.code === 4001) {
      toast('Transaction cancelled');
      return;
    }
    showErrorToast(`${label} failed: ${e?.message ?? String(err)}`);
  };

  const deposit = useCallback(
    async (blockchainId: bigint, asset: string, amount: Decimal) => {
      if (!client || !address) return;
      const gen = generationRef.current;
      setDepositPhase('approving'); // start — shows "Approve" or transitions to approval immediately
      try {
        const tokenInfo = tokenInfoFor(asset, blockchainId);
        if (!tokenInfo) throw new Error(`token not found for ${asset} on chain ${blockchainId}`);

        const chain = supportedChains.find(c => c.id === blockchainId);
        const delaySecs = chain?.confirmationDelaySecs ?? 0;

        const isNative = /^0x0+$/i.test(tokenInfo.address);
        if (!isNative) {
          const allowance = await client.checkTokenAllowance(blockchainId, tokenInfo.address, address);
          if (generationRef.current !== gen) {
            toast('Wallet changed — operation cancelled');
            return;
          }
          const amountUnits = BigInt(
            amount.mul(new Decimal(10).pow(tokenInfo.decimals)).floor().toFixed(0),
          );
          if (allowance < amountUnits) {
            // Approve effectively-unlimited once so future deposits skip the popup.
            const maxUnits = (1n << 256n) - 1n;
            const divisor = 10n ** BigInt(tokenInfo.decimals);
            const humanMax = new Decimal((maxUnits / divisor).toString());
            await client.approveToken(blockchainId, asset, humanMax);
            setNeedsApproval(false);
            if (generationRef.current !== gen) {
              toast('Wallet changed — operation cancelled');
              return;
            }
          }
        }

        let baselineEnforced: Decimal | null = new Decimal(0);
        if (delaySecs > 0) {
          try {
            const entries = await client.getBalances(address);
            baselineEnforced = entries.find(e => e.asset === asset)?.enforced ?? new Decimal(0);
          } catch { baselineEnforced = null; /* unknown baseline → waitForCredit falls back to soft toast */ }
        }

        setDepositPhase('signing_state');
        await client.deposit(blockchainId, asset, amount);
        if (generationRef.current !== gen) return;
        onAfterTxMined?.(); // signed/issued states update; unified balance stays until tx mines
        setDepositPhase('signing_tx');
        const depositTxHash = await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        setDepositPhase('confirming');
        const depositRpcUrl = rpcUrlFor(blockchainId);
        if (depositRpcUrl && depositTxHash) {
          const depositClient = createPublicClient({ transport: http(depositRpcUrl) });
          await depositClient.waitForTransactionReceipt({ hash: depositTxHash as Hash });
          if (generationRef.current !== gen) return;
        }
        if (delaySecs > 0) {
          toast(`Tx mined — awaiting node (~${delaySecs}s)…`);
          setDepositPhase('awaiting_node');
          onAfterTxMined?.(); // refresh signed/issued states + checkpoint button now; credit not yet enforced
          const credited = await waitForCredit(asset, delaySecs, baselineEnforced, 'up', gen);
          if (generationRef.current !== gen) return;
          if (credited) {
            toast.success(`Deposited ${amount.toString()} ${asset}`);
          } else {
            toast(`Deposit submitted — credit will appear after node confirmation (~${delaySecs}s)`);
          }
          onAfterOp?.();
        } else {
          toast.success(`Deposited ${amount.toString()} ${asset}`);
          onAfterOp?.();       // updates unified balance and on-chain balance
          onAfterTxMined?.();  // forces channel-states refresh (enforced version, checkpoint button)
        }
      } catch (err) {
        handleError(err, 'Deposit');
      } finally {
        setDepositPhase('idle');
      }
    },
    [client, address, tokenInfoFor, supportedChains, waitForCredit, onAfterOp, onAfterTxMined],
  );

  const withdraw = useCallback(
    async (blockchainId: bigint, asset: string, amount: Decimal) => {
      if (!client || !address) return;
      const gen = generationRef.current;
      setWithdrawPhase('signing_state');
      try {
        const chain = supportedChains.find(c => c.id === blockchainId);
        const delaySecs = chain?.confirmationDelaySecs ?? 0;

        let baselineEnforced: Decimal | null = new Decimal(0);
        if (delaySecs > 0) {
          try {
            const entries = await client.getBalances(address);
            baselineEnforced = entries.find(e => e.asset === asset)?.enforced ?? new Decimal(0);
          } catch { baselineEnforced = null; /* unknown baseline → waitForCredit falls back to soft toast */ }
        }

        await client.withdraw(blockchainId, asset, amount);
        if (generationRef.current !== gen) return;
        onAfterOp?.();      // unified balance updates immediately after state is signed
        onAfterTxMined?.(); // signed/issued states update
        setWithdrawPhase('signing_tx');
        const withdrawTxHash = await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        setWithdrawPhase('confirming');
        const withdrawRpcUrl = rpcUrlFor(blockchainId);
        if (withdrawRpcUrl && withdrawTxHash) {
          const withdrawClient = createPublicClient({ transport: http(withdrawRpcUrl) });
          await withdrawClient.waitForTransactionReceipt({ hash: withdrawTxHash as Hash });
          if (generationRef.current !== gen) return;
        }
        if (delaySecs > 0) {
          toast(`Tx mined — awaiting node (~${delaySecs}s)…`);
          setWithdrawPhase('awaiting_node');
          onAfterTxMined?.(); // refresh signed/issued states + checkpoint button now; credit not yet enforced
          const credited = await waitForCredit(asset, delaySecs, baselineEnforced, 'down', gen);
          if (generationRef.current !== gen) return;
          if (credited) {
            toast.success(`Withdrew ${amount.toString()} ${asset}`);
          } else {
            toast(`Withdrawal submitted — on-chain settlement after node confirmation (~${delaySecs}s)`);
          }
          onAfterOp?.();
        } else {
          toast.success(`Withdrew ${amount.toString()} ${asset}`);
          onAfterOp?.();       // updates unified balance and on-chain balance
          onAfterTxMined?.();  // forces channel-states refresh (enforced version, checkpoint button)
        }
      } catch (err) {
        handleError(err, 'Withdraw');
      } finally {
        setWithdrawPhase('idle');
      }
    },
    [client, address, supportedChains, waitForCredit, onAfterOp, onAfterTxMined],
  );

  const performTransfer = useCallback(
    async (to: Address, asset: string, amount: Decimal): Promise<boolean> => {
      if (!client) return false;
      const gen = generationRef.current;
      setTransferPhase('signing_state');
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
        setTransferPhase('idle');
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
    async (asset: string, blockchainId: bigint) => {
      if (!client || !address) return;
      const gen = generationRef.current;
      setClosingAsset(asset);
      try {
        toast('Closing channel…');

        const chain = supportedChains.find(c => c.id === blockchainId);
        const delaySecs = chain?.confirmationDelaySecs ?? 0;

        let baselineEnforced: Decimal | null = new Decimal(0);
        if (delaySecs > 0) {
          try {
            const entries = await client.getBalances(address);
            baselineEnforced = entries.find(e => e.asset === asset)?.enforced ?? new Decimal(0);
          } catch { baselineEnforced = null; /* unknown baseline → waitForCredit falls back to soft toast */ }
        }

        // If a Finalize state is already signed (e.g. a previous checkpoint tx failed or
        // the channel is in Closing status on-chain), skip re-signing and go straight to
        // the on-chain transaction.
        const signedState = await client.getLatestState(address, asset, true);
        if (generationRef.current !== gen) return;
        if (!signedState || signedState.transition.type !== TransitionType.Finalize) {
          await client.closeHomeChannel(asset);
          if (generationRef.current !== gen) return;
        }
        const txHash = await client.checkpoint(asset);
        if (generationRef.current !== gen) return;
        // Wait for the transaction to be mined before reporting success.
        const rpcUrl = rpcUrlFor(blockchainId);
        if (rpcUrl && txHash) {
          const publicClient = createPublicClient({ transport: http(rpcUrl) });
          await publicClient.waitForTransactionReceipt({ hash: txHash as Hash });
          if (generationRef.current !== gen) return;
        }
        if (delaySecs > 0) {
          toast(`Channel close mined — finalizing after node confirmation (~${delaySecs}s)…`);
          // Transition: tx mined, waiting for node confirmation window.
          setClosingAsset(null);
          setAwaitingCloseAsset(asset);
          const credited = await waitForCredit(asset, delaySecs, baselineEnforced, 'down', gen);
          if (generationRef.current !== gen) return;
          if (credited) {
            toast.success(`Closed channel for ${asset}`);
          } else {
            toast(`Close submitted — will finalize after node confirmation`);
          }
        } else {
          toast.success(`Closed channel for ${asset}`);
        }
        onAfterOp?.();
        onAfterTxMined?.();
      } catch (err) {
        handleError(err, 'Close');
      } finally {
        setClosingAsset(null);
        setAwaitingCloseAsset(null);
      }
    },
    [client, address, supportedChains, waitForCredit, onAfterOp, onAfterTxMined],
  );

  return {
    deposit,
    withdraw,
    transfer,
    closeChannel,
    depositPhase,
    withdrawPhase,
    transferPhase,
    needsApproval,
    checkDepositAllowance,
    closingAsset,
    awaitingCloseAsset,
    homechainModalAsset,
    onHomechainSelected,
    onHomechainModalDismiss,
  };
}
