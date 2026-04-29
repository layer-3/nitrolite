import { useState, useEffect, useCallback } from 'react';
import {
  ArrowDownToLine, ArrowUpFromLine, Send, XCircle,
  Shield, AlertTriangle, CheckCircle2, ArrowRightLeft,
  RefreshCw, Key, Loader2, ChevronDown, ChevronUp,
  Users, Database, Copy, Check, Lock, Unlock, AppWindow, Plus,
} from 'lucide-react';
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import type { WalletClient } from 'viem';
import Decimal from 'decimal.js';
import {
  getChannelSessionKeyAuthMetadataHashV1,
  packChannelKeyStateV1,
} from '@yellow-org/sdk';
import type { Client, ChannelSessionKeyStateV1, ActionAllowance, AppInfoV1 } from '@yellow-org/sdk';
import type { SessionKeyState, StatusMessage } from '../types';
import { formatAddress, formatBalance, timeAgo, formatTxType } from '../utils';
import { Button } from './ui/button';
import ActionModal from './ActionModal';

/**
 * Sign a session key state with the main wallet (EIP-191).
 * Used for both registration and revocation.
 */
async function signSessionKeyStateWithWallet(
  wc: WalletClient,
  state: { session_key: string; version: string; assets: string[]; expires_at: string },
): Promise<`0x${string}`> {
  if (!wc.account) throw new Error('Wallet client does not have an account');
  const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
    BigInt(state.version),
    state.assets,
    BigInt(state.expires_at),
  );
  const packed = packChannelKeyStateV1(state.session_key as `0x${string}`, metadataHash);
  return wc.signMessage({ account: wc.account, message: { raw: packed } });
}

interface WalletDashboardProps {
  client: Client;
  address: string;
  chainId: string;
  asset: string;
  assets: string[];
  sessionKey: SessionKeyState | null;
  walletClient: WalletClient;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
  onSetSessionKey: (sk: SessionKeyState) => void;
  onActivateSessionKey: (sk: SessionKeyState) => Promise<void>;
  onClearSessionKey: () => Promise<void>;
  onAssetChange: (asset: string) => void;
}

function isAllowanceError(error: unknown): boolean {
  const msg = error instanceof Error ? error.message : String(error);
  const lower = msg.toLowerCase();
  // OpenZeppelin ERC20InsufficientAllowance(address,uint256,uint256) selector
  if (lower.includes('0xfb8f41b2')) return true;
  return lower.includes('allowance') && lower.includes('sufficient');
}


const MAX_APPROVE_AMOUNT = new Decimal('1e18');

export default function WalletDashboard({
  client, address, chainId, asset, assets,
  sessionKey, walletClient, showStatus,
  onSetSessionKey, onActivateSessionKey, onClearSessionKey, onAssetChange,
}: WalletDashboardProps) {
  // Auto-fetched data
  const [balances, setBalances] = useState<any[]>([]);
  const [latestState, setLatestState] = useState<any>(null);
  const [signedState, setSignedState] = useState<any>(null);
  const [homeChannel, setHomeChannel] = useState<any>(null);
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // UI state
  const [activeModal, setActiveModal] = useState<'deposit' | 'withdraw' | 'transfer' | 'close' | null>(null);
  const [acknowledging, setAcknowledging] = useState(false);
  const [checkpointing, setCheckpointing] = useState(false);
  const [skLoading, setSkLoading] = useState(false);
  const [selectedTx, setSelectedTx] = useState<any>(null);
  const [copied, setCopied] = useState(false);

  // Security token state
  const [lockedBalance, setLockedBalance] = useState<Decimal | null>(null);
  const [allowances, setAllowances] = useState<ActionAllowance[]>([]);
  const [lockingLoading, setLockingLoading] = useState(false);
  const [lockAmount, setLockAmount] = useState('');

  // Advanced section
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [nodeConfig, setNodeConfig] = useState<any>(null);
  const [, setNodeConfigLoading] = useState(false);
  const [appSessions, setAppSessions] = useState<any[]>([]);
  const [appSessionsMeta, setAppSessionsMeta] = useState<any>(null);
  const [appSessionsLoading, setAppSessionsLoading] = useState(false);
  const [keyStates, setKeyStates] = useState<ChannelSessionKeyStateV1[]>([]);
  const [keyStatesLoading, setKeyStatesLoading] = useState(false);
  const [revokingKey, setRevokingKey] = useState<string | null>(null);

  // App Registry
  const [myApps, setMyApps] = useState<AppInfoV1[]>([]);
  const [myAppsLoading, setMyAppsLoading] = useState(false);
  const [registerAppId, setRegisterAppId] = useState('');
  const [registerAppMetadata, setRegisterAppMetadata] = useState('');
  const [registerAppNoApproval, setRegisterAppNoApproval] = useState(false);
  const [registeringApp, setRegisteringApp] = useState(false);


  // Fetch all wallet data
  const fetchData = useCallback(async (isManual = false) => {
    if (isManual) setRefreshing(true);
    try {
      const results = await Promise.allSettled([
        client.getBalances(address as `0x${string}`),
        client.getLatestState(address as `0x${string}`, asset, false),
        client.getLatestState(address as `0x${string}`, asset, true),
        client.getHomeChannel(address as `0x${string}`, asset),
        client.getTransactions(address as `0x${string}`, { page: 1, pageSize: 10 }),
        client.getLockedBalance(BigInt(chainId), address),
        client.getActionAllowances(address as `0x${string}`),
        !nodeConfig ? client.getConfig() : Promise.resolve(null),
      ]);

      if (results[0].status === 'fulfilled') setBalances(results[0].value as any[]);
      setLatestState(results[1].status === 'fulfilled' ? results[1].value : null);
      setSignedState(results[2].status === 'fulfilled' ? results[2].value : null);
      setHomeChannel(results[3].status === 'fulfilled' ? results[3].value : null);
      if (results[4].status === 'fulfilled') {
        setTransactions((results[4].value as any).transactions || []);
      }
      if (results[5].status === 'fulfilled') setLockedBalance(results[5].value as Decimal);
      if (results[6].status === 'fulfilled') setAllowances(results[6].value as ActionAllowance[]);
      if (results[7].status === 'fulfilled' && results[7].value) setNodeConfig(results[7].value);
    } catch (e) {
      console.error('Fetch error:', e);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [client, address, asset, chainId]);

  // Initial fetch + polling
  useEffect(() => {
    setLoading(true);
    fetchData();
    const interval = setInterval(() => fetchData(), 15000);
    return () => clearInterval(interval);
  }, [fetchData]);

  // Determine if acknowledge is needed
  const needsAcknowledge = (() => {
    if (!latestState) return false;
    if (!signedState) return true;
    try {
      return BigInt(latestState.version) > BigInt(signedState.version);
    } catch {
      return false;
    }
  })();


  // Determine if optional checkpoint is available
  const canCheckpoint = (() => {
    if (!homeChannel || !signedState) return false;
    try {
      const onChain = BigInt(homeChannel.stateVersion || 0);
      const offChain = BigInt(signedState.version);
      return onChain < offChain;
    } catch {
      return false;
    }
  })();

  // Current balance for selected asset
  const currentBalance = balances.find(
    (b: any) => b.asset?.toLowerCase() === asset.toLowerCase()
  );

  // --- Handlers ---

  const handleAcknowledge = async () => {
    try {
      setAcknowledging(true);
      await client.acknowledge(asset);
      showStatus('success', 'State acknowledged');
      await fetchData();
    } catch (error) {
      showStatus('error', 'Acknowledge failed', error instanceof Error ? error.message : String(error));
    } finally {
      setAcknowledging(false);
    }
  };

  const handleCheckpoint = async () => {
    try {
      setCheckpointing(true);
      let txHash: string;
      try {
        txHash = await client.checkpoint(asset);
      } catch (error) {
        if (!isAllowanceError(error)) throw error;
        const state = await client.getLatestState(
          client.getUserAddress() as `0x${string}`, asset, true,
        );
        const cid = state.homeLedger.blockchainId;
        await client.approveToken(cid, asset, MAX_APPROVE_AMOUNT);
        txHash = await client.checkpoint(asset);
      }
      showStatus('success', 'Checkpoint complete', `Tx: ${txHash}`);
      await fetchData();
    } catch (error) {
      showStatus('error', 'Checkpoint failed', error instanceof Error ? error.message : String(error));
    } finally {
      setCheckpointing(false);
    }
  };

  const handleActionComplete = useCallback(() => {
    setActiveModal(null);
    fetchData();
  }, [fetchData]);

  // Session key: enable auto sign
  const handleEnableAutoSign = async () => {
    if (!walletClient.account) return;
    try {
      setSkLoading(true);
      const privateKey = generatePrivateKey();
      const account = privateKeyToAccount(privateKey);
      const newSk: SessionKeyState = {
        privateKey,
        address: account.address,
        metadataHash: '',
        authSig: '',
        active: false,
      };
      onSetSessionKey(newSk);

      const assetList = [...assets];
      let version = 1n;
      try {
        const existing = await client.getLastChannelKeyStates(address, newSk.address);
        if (existing?.length > 0) version = BigInt(existing[0].version) + 1n;
      } catch { /* no existing keys */ }

      const expiresAt = BigInt(Math.floor(Date.now() / 1000) + 24 * 3600);

      // Build state object
      const state = {
        user_address: address,
        session_key: newSk.address,
        version: version.toString(),
        assets: assetList,
        expires_at: expiresAt.toString(),
        user_sig: '',
      };

      // Sign using the SDK method (goes through ChannelDefaultSigner, strips prefix)
      const sig = await client.signChannelSessionKeyState(state);
      state.user_sig = sig;

      // Submit to nitronode
      await client.submitChannelSessionKeyState(state);

      // Compute metadata hash for the session key signer
      const metadataHash = getChannelSessionKeyAuthMetadataHashV1(version, assetList, expiresAt);

      const activeSk: SessionKeyState = {
        ...newSk,
        metadataHash: metadataHash as string,
        authSig: sig,
        active: true,
      };
      await onActivateSessionKey(activeSk);
      showStatus('success', 'Auto Sign enabled', 'Expires in 24 hours');
    } catch (error) {
      showStatus('error', 'Failed to enable Auto Sign', error instanceof Error ? error.message : String(error));
    } finally {
      setSkLoading(false);
    }
  };

  const handleDisableAutoSign = async () => {
    if (!walletClient.account || !sessionKey) return;
    try {
      setSkLoading(true);

      // Revoke on-chain: submit a new version with zero assets
      try {
        const existing = await client.getLastChannelKeyStates(address, sessionKey.address);
        if (existing && existing.length > 0) {
          const latest = existing[0];
          const newVersion = BigInt(latest.version) + 1n;
          const revokeState = {
            user_address: address,
            session_key: sessionKey.address,
            version: newVersion.toString(),
            assets: [] as string[],
            expires_at: latest.expires_at,
            user_sig: '',
          };
          const sig = await signSessionKeyStateWithWallet(walletClient, revokeState);
          revokeState.user_sig = sig;
          await client.submitChannelSessionKeyState(revokeState);
        }
      } catch (revokeErr) {
        // Log but don't block the local clear
        console.error('Revoke on-chain failed:', revokeErr);
      }

      await onClearSessionKey();
      if (showAdvanced) await fetchKeyStates();
      showStatus('success', 'Auto Sign disabled and key revoked');
    } catch (error) {
      showStatus('error', 'Failed to disable Auto Sign', error instanceof Error ? error.message : String(error));
    } finally {
      setSkLoading(false);
    }
  };

  // --- Security token handlers ---

  const GATED_ACTION_LABELS: Record<string, string> = {
    transfer: 'Transfers',
    app_session_creation: 'App Session Creations',
    app_session_operation: 'App Session Updates',
    app_session_deposit: 'App Session Deposits',
    app_session_withdrawal: 'App Session Withdrawals',
  };

  const handleLockTokens = async () => {
    if (!lockAmount || lockingLoading) return;
    try {
      setLockingLoading(true);
      const amount = new Decimal(lockAmount);
      try {
        await client.escrowSecurityTokens(address, BigInt(chainId), amount);
      } catch (error) {
        if (!isAllowanceError(error)) throw error;
        await client.approveSecurityToken(BigInt(chainId), MAX_APPROVE_AMOUNT);
        await client.escrowSecurityTokens(address, BigInt(chainId), amount);
      }
      showStatus('success', 'Tokens locked successfully');
      setLockAmount('');
      await fetchData();
    } catch (error) {
      showStatus('error', 'Lock failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLockingLoading(false);
    }
  };

  const handleInitiateUnlock = async () => {
    try {
      setLockingLoading(true);
      await client.initiateSecurityTokensWithdrawal(BigInt(chainId));
      showStatus('success', 'Unlock initiated', 'Tokens will be available for withdrawal after the unlock period');
      await fetchData();
    } catch (error) {
      showStatus('error', 'Unlock initiation failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLockingLoading(false);
    }
  };

  const handleCancelUnlock = async () => {
    try {
      setLockingLoading(true);
      await client.cancelSecurityTokensWithdrawal(BigInt(chainId));
      showStatus('success', 'Unlock cancelled, tokens re-locked');
      await fetchData();
    } catch (error) {
      showStatus('error', 'Cancel unlock failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLockingLoading(false);
    }
  };

  const handleWithdrawTokens = async () => {
    try {
      setLockingLoading(true);
      await client.withdrawSecurityTokens(BigInt(chainId), address);
      showStatus('success', 'Tokens withdrawn successfully');
      await fetchData();
    } catch (error) {
      showStatus('error', 'Withdraw failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLockingLoading(false);
    }
  };

  // --- Advanced section handlers ---

  const fetchNodeConfig = async () => {
    try {
      setNodeConfigLoading(true);
      const config = await client.getConfig();
      setNodeConfig(config);
    } catch (error) {
      showStatus('error', 'Failed to load node info', error instanceof Error ? error.message : String(error));
    } finally {
      setNodeConfigLoading(false);
    }
  };

  const fetchAppSessions = async () => {
    try {
      setAppSessionsLoading(true);
      const { sessions, metadata } = await client.getAppSessions({
        wallet: address as `0x${string}`,
        page: 1,
        pageSize: 20,
      });
      setAppSessions(sessions);
      setAppSessionsMeta(metadata);
    } catch {
      // Silently handle — user may not have any app sessions
      setAppSessions([]);
      setAppSessionsMeta(null);
    } finally {
      setAppSessionsLoading(false);
    }
  };

  const fetchKeyStates = async () => {
    try {
      setKeyStatesLoading(true);
      const states = await client.getLastChannelKeyStates(address);
      setKeyStates(states);
    } catch (error) {
      // Silently handle — may not have any keys
      setKeyStates([]);
    } finally {
      setKeyStatesLoading(false);
    }
  };

  const fetchMyApps = async () => {
    try {
      setMyAppsLoading(true);
      const { apps } = await client.getApps({ ownerWallet: address });
      setMyApps(apps);
    } catch {
      setMyApps([]);
    } finally {
      setMyAppsLoading(false);
    }
  };

  const handleRegisterApp = async () => {
    if (!registerAppId || registeringApp) return;
    try {
      setRegisteringApp(true);
      await client.registerApp(registerAppId, registerAppMetadata, registerAppNoApproval);
      showStatus('success', 'App registered', `App ID: ${registerAppId}`);
      setRegisterAppId('');
      setRegisterAppMetadata('');
      setRegisterAppNoApproval(false);
      await fetchMyApps();
    } catch (error) {
      showStatus('error', 'Registration failed', error instanceof Error ? error.message : String(error));
    } finally {
      setRegisteringApp(false);
    }
  };

  const handleRevokeKey = async (ks: ChannelSessionKeyStateV1) => {
    if (!walletClient.account) return;
    const revokeId = `${ks.session_key}-${ks.version}`;
    try {
      setRevokingKey(revokeId);
      const newVersion = BigInt(ks.version) + 1n;
      const revokeState = {
        user_address: address,
        session_key: ks.session_key,
        version: newVersion.toString(),
        assets: [] as string[],
        expires_at: ks.expires_at,
        user_sig: '',
      };
      const sig = await signSessionKeyStateWithWallet(walletClient, revokeState);
      revokeState.user_sig = sig;
      await client.submitChannelSessionKeyState(revokeState);

      showStatus('success', 'Session key revoked', `Key ${formatAddress(ks.session_key)}`);

      // If revoked key is the currently active one, clear it
      if (sessionKey?.active && sessionKey.address.toLowerCase() === ks.session_key.toLowerCase()) {
        await onClearSessionKey();
      }

      await fetchKeyStates();
    } catch (error) {
      showStatus('error', 'Revoke failed', error instanceof Error ? error.message : String(error));
    } finally {
      setRevokingKey(null);
    }
  };

  // Auto-fetch advanced data when section opens
  useEffect(() => {
    if (showAdvanced) {
      if (!nodeConfig) fetchNodeConfig();
      fetchAppSessions();
      fetchKeyStates();
      fetchMyApps();
    }
  }, [showAdvanced]); // eslint-disable-line react-hooks/exhaustive-deps

  // --- Transaction display helpers ---

  const getTxDisplay = (tx: any) => {
    const { label, isIncoming } = formatTxType(tx.txType, {
      fromAccount: tx.fromAccount,
      userAddress: address,
    });
    const str = String(tx.txType).toLowerCase();

    let Icon = ArrowRightLeft;
    let color = 'text-muted-foreground';
    if (isIncoming) {
      Icon = ArrowDownToLine;
      color = 'text-green-400';
    } else if (str === '11' || str.includes('withdraw')) {
      Icon = ArrowUpFromLine;
      color = 'text-orange-400';
    } else if (str === '30' || str.includes('send') || str === 'transfer') {
      Icon = Send;
      color = 'text-red-400';
    }

    const counterparty = isIncoming ? tx.fromAccount : tx.toAccount;

    return { label, isIncoming, Icon, color, counterparty };
  };

  // --- Render ---

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-accent" />
          <span className="text-xs text-muted-foreground uppercase tracking-wider">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4 animate-fade-in">
      {/* Acknowledge banner */}
      {needsAcknowledge && (
        <div className="bg-accent/10 border border-accent/30 px-4 py-3 flex items-center justify-between rounded-xl">
          <div className="flex items-center gap-2 text-xs">
            <AlertTriangle className="h-3.5 w-3.5 text-accent flex-shrink-0" />
            <span>New state v{latestState?.version?.toString()} needs acknowledgement</span>
          </div>
          <Button
            onClick={handleAcknowledge}
            disabled={acknowledging}
            size="sm"
            className="h-7 text-xs"
          >
            {acknowledging ? (
              <><Loader2 className="h-3 w-3 animate-spin mr-1" /> Acknowledging...</>
            ) : 'Acknowledge'}
          </Button>
        </div>
      )}

      {/* Hero: Balance + Actions + Security Token */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Channel Balance Card — spans 2 cols */}
        <div className="lg:col-span-2 glass glass-hover rounded-xl p-6 transition-colors duration-200 flex flex-col">
          {/* Top: asset tabs + address */}
          <div className="flex items-center justify-between mb-4">
            <div className="flex gap-1.5">
              {assets.map(a => (
                <button
                  key={a}
                  onClick={() => onAssetChange(a)}
                  className={`px-3.5 py-1.5 text-xs font-semibold uppercase tracking-wider transition-all rounded-lg ${
                    a === asset
                      ? 'bg-accent text-accent-foreground'
                      : 'glass-input text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {a}
                </button>
              ))}
            </div>
            <button
              onClick={() => {
                navigator.clipboard.writeText(address);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
              }}
              title={address}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors font-mono"
            >
              {formatAddress(address)}
              {copied ? <Check className="h-3 w-3 text-green-400" /> : <Copy className="h-3 w-3" />}
            </button>
          </div>

          {/* Balance display */}
          <div className="mb-6">
            <div className="text-4xl lg:text-5xl font-semibold tracking-tight tabular-nums">
              {formatBalance(currentBalance?.balance)}
            </div>
            <div className="text-sm text-muted-foreground uppercase tracking-wider mt-1">
              {asset.toUpperCase()} Balance
            </div>
          </div>

          {/* Action buttons — grid of tiles */}
          <div className="grid grid-cols-4 gap-2 mb-5">
            <button
              onClick={() => setActiveModal('deposit')}
              className="group flex flex-col items-center gap-1.5 py-3 rounded-lg glass-input hover:bg-green-500/10 hover:border-green-500/20 transition-all"
            >
              <ArrowDownToLine className="h-4 w-4 text-green-400 group-hover:scale-110 transition-transform" />
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground group-hover:text-foreground transition-colors">Deposit</span>
            </button>
            <button
              onClick={() => setActiveModal('withdraw')}
              className="group flex flex-col items-center gap-1.5 py-3 rounded-lg glass-input hover:bg-orange-500/10 hover:border-orange-500/20 transition-all"
            >
              <ArrowUpFromLine className="h-4 w-4 text-orange-400 group-hover:scale-110 transition-transform" />
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground group-hover:text-foreground transition-colors">Withdraw</span>
            </button>
            <button
              onClick={() => setActiveModal('transfer')}
              className="group flex flex-col items-center gap-1.5 py-3 rounded-lg glass-input hover:bg-blue-500/10 hover:border-blue-500/20 transition-all"
            >
              <Send className="h-4 w-4 text-blue-400 group-hover:scale-110 transition-transform" />
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground group-hover:text-foreground transition-colors">Send</span>
            </button>
            <button
              onClick={() => setActiveModal('close')}
              className="group flex flex-col items-center gap-1.5 py-3 rounded-lg glass-input hover:bg-red-500/10 hover:border-red-500/20 transition-all"
            >
              <XCircle className="h-4 w-4 text-red-400 group-hover:scale-110 transition-transform" />
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground group-hover:text-foreground transition-colors">Close</span>
            </button>
          </div>

          {/* Channel sync status */}
          <div className="flex items-center gap-3 text-xs mt-auto pt-4 border-t border-white/[0.06]">
            {canCheckpoint ? (
              <>
                <div className="flex items-center gap-1.5 text-yellow-500">
                  <AlertTriangle className="h-3.5 w-3.5" />
                  <span>On-chain state behind</span>
                  <span className="text-muted-foreground ml-1">
                    v{homeChannel?.stateVersion?.toString()} &rarr; v{signedState?.version?.toString()}
                  </span>
                </div>
                <Button
                  onClick={handleCheckpoint}
                  disabled={checkpointing}
                  size="sm"
                  variant="outline"
                  className="h-6 text-xs gap-1 px-2"
                >
                  {checkpointing ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Shield className="h-3 w-3" />
                  )}
                  {checkpointing ? 'Syncing...' : 'Sync'}
                </Button>
              </>
            ) : homeChannel ? (
              <div className="flex items-center gap-1.5 text-green-500">
                <CheckCircle2 className="h-3.5 w-3.5" />
                <span>Channel synced</span>
                {homeChannel.stateVersion != null && (
                  <span className="text-muted-foreground ml-1">v{homeChannel.stateVersion.toString()}</span>
                )}
              </div>
            ) : (
              <span className="text-muted-foreground">No channel yet &mdash; deposit to get started</span>
            )}
          </div>
        </div>

        {/* Security Token Card */}
        <div className="glass glass-hover rounded-xl p-6 transition-colors duration-200 flex flex-col">
          <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-4">
            <Lock className="h-3.5 w-3.5" />
            <span>Security Token</span>
          </div>

          <div className="mb-6">
            <div className="text-3xl lg:text-4xl font-semibold tracking-tight tabular-nums">
              {lockedBalance ? lockedBalance.toString() : '0'}
            </div>
            <div className="text-sm text-muted-foreground uppercase tracking-wider mt-1">
              YELLOW Locked
            </div>
          </div>

          {/* Lock input */}
          <div className="flex items-center gap-2 mb-3">
            <input
              type="text"
              value={lockAmount}
              onChange={(e) => setLockAmount(e.target.value)}
              placeholder="Amount"
              className="h-9 flex-1 px-3 text-xs glass-input rounded-lg font-mono focus:outline-none focus:border-accent"
            />
            <Button
              onClick={handleLockTokens}
              disabled={lockingLoading || !lockAmount}
              size="sm"
              className="h-9 text-xs gap-1.5"
            >
              {lockingLoading ? <Loader2 className="h-3 w-3 animate-spin" /> : <Lock className="h-3 w-3" />}
              Lock
            </Button>
          </div>

          {/* Unlock controls — only when there's a balance */}
          {lockedBalance && !lockedBalance.isZero() && (
            <div className="grid grid-cols-3 gap-2 mt-auto">
              <Button
                onClick={handleInitiateUnlock}
                disabled={lockingLoading}
                variant="outline"
                size="sm"
                className="h-8 text-xs gap-1"
              >
                <Unlock className="h-3 w-3" />
                Unlock
              </Button>
              <Button
                onClick={handleCancelUnlock}
                disabled={lockingLoading}
                variant="outline"
                size="sm"
                className="h-8 text-[11px]"
              >
                Cancel
              </Button>
              <Button
                onClick={handleWithdrawTokens}
                disabled={lockingLoading}
                variant="outline"
                size="sm"
                className="h-8 text-xs gap-1"
              >
                <ArrowUpFromLine className="h-3 w-3" />
                Claim
              </Button>
            </div>
          )}
        </div>
      </div>

      {/* Activity Feed */}
      <div className="glass glass-hover rounded-xl p-6 transition-colors duration-200">
        <div className="flex items-center justify-between mb-4">
          <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Recent Activity
          </div>
          <button
            onClick={() => fetchData(true)}
            disabled={refreshing}
            className="text-muted-foreground hover:text-foreground transition-colors p-1"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {transactions.length === 0 ? (
          <div className="text-sm text-muted-foreground text-center py-8">
            No transactions yet
          </div>
        ) : (
          <div className="space-y-0.5 max-h-[280px] overflow-y-auto scrollbar-thin">
            {transactions.slice(0, 10).map((tx: any, idx: number) => {
              const { label, isIncoming, Icon, color, counterparty } = getTxDisplay(tx);
              return (
                <div
                  key={tx.id || idx}
                  onClick={() => setSelectedTx(tx)}
                  className="flex items-center gap-3 py-2.5 px-2 border-b border-border/40 last:border-0 hover:bg-white/[0.03] transition-colors cursor-pointer rounded-lg"
                >
                  <div className={`w-8 h-8 flex items-center justify-center flex-shrink-0 ${color}`}>
                    <Icon className="h-4 w-4" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{label}</div>
                    {counterparty && (
                      <div className="text-xs text-muted-foreground font-mono truncate">
                        {formatAddress(counterparty)}
                      </div>
                    )}
                  </div>
                  <div className="text-right flex-shrink-0">
                    <div className={`text-sm font-mono font-medium ${isIncoming ? 'text-green-400' : ''}`}>
                      {isIncoming ? '+' : '-'}{tx.amount?.toString()} {tx.asset?.toUpperCase()}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {timeAgo(tx.createdAt)}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Auto Sign (Session Key) */}
      <div className="glass glass-hover rounded-xl px-6 py-4 transition-colors duration-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Key className="h-4 w-4 text-muted-foreground flex-shrink-0" />
            <div>
              <div className="text-xs font-semibold uppercase tracking-wider">Auto Sign</div>
              {sessionKey?.active ? (
                <div className="text-xs text-muted-foreground">
                  <span className="text-green-400">Active</span>
                  <span className="mx-1.5">&middot;</span>
                  <span className="font-mono">{formatAddress(sessionKey.address)}</span>
                  <span className="mx-1.5">&middot;</span>
                  <span>No wallet popups needed</span>
                </div>
              ) : (
                <div className="text-xs text-muted-foreground">
                  Disabled &mdash; each operation requires a wallet popup to sign
                </div>
              )}
            </div>
          </div>
          <div>
            {sessionKey?.active ? (
              <Button
                onClick={handleDisableAutoSign}
                disabled={skLoading}
                variant="outline"
                size="sm"
                className="h-7 text-xs"
              >
                {skLoading ? (
                  <><Loader2 className="h-3 w-3 animate-spin mr-1" /> Disabling...</>
                ) : 'Disable'}
              </Button>
            ) : (
              <Button
                onClick={handleEnableAutoSign}
                disabled={skLoading}
                size="sm"
                className="h-7 text-xs"
              >
                {skLoading ? (
                  <><Loader2 className="h-3 w-3 animate-spin mr-1" /> Enabling...</>
                ) : 'Enable'}
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Action Allowances */}
      {allowances.length > 0 && (
        <div className="glass glass-hover rounded-xl px-6 py-4 transition-colors duration-200">
          <div className="mb-3">
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              <Shield className="h-3.5 w-3.5" />
              <span>Action Allowances</span>
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              Allowances increase as you lock more security tokens
            </div>
          </div>
          <div className="space-y-2">
            {allowances.map((a) => {
              const used = Number(a.used);
              const total = Number(a.allowance);
              const pct = total > 0 ? Math.min((used / total) * 100, 100) : 0;
              return (
                <div key={a.gatedAction} className="space-y-1">
                  <div className="flex items-center justify-between text-xs">
                    <span>{GATED_ACTION_LABELS[a.gatedAction] || a.gatedAction}</span>
                    <span className="font-mono text-muted-foreground">
                      {used}/{total}
                      <span className="ml-1.5 text-[10px]">({a.timeWindow})</span>
                    </span>
                  </div>
                  <div className="h-1.5 bg-muted overflow-hidden rounded-full">
                    <div
                      className={`h-full transition-all rounded-full ${pct >= 90 ? 'bg-red-500 shadow-[0_0_6px_rgba(239,68,68,0.4)]' : pct >= 60 ? 'bg-yellow-500 shadow-[0_0_6px_rgba(234,179,8,0.4)]' : 'bg-accent shadow-[0_0_6px_rgba(234,179,8,0.3)]'}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Advanced Section Toggle */}
      <div className="glass rounded-xl transition-colors duration-200">
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="w-full px-6 py-3 flex items-center justify-between text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors rounded-xl"
        >
          <span>Advanced</span>
          {showAdvanced ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </button>

        {showAdvanced && (
          <div className="px-6 pb-6 space-y-6 animate-fade-in">
            {/* App Sessions */}
            <div>
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
                <Users className="h-3.5 w-3.5" />
                <span>App Sessions</span>
                <button
                  onClick={fetchAppSessions}
                  disabled={appSessionsLoading}
                  className="ml-auto text-muted-foreground hover:text-foreground transition-colors"
                >
                  <RefreshCw className={`h-3 w-3 ${appSessionsLoading ? 'animate-spin' : ''}`} />
                </button>
              </div>
              {appSessionsLoading ? (
                <div className="flex items-center gap-2 text-xs text-muted-foreground py-4">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Loading sessions...</span>
                </div>
              ) : appSessions.length === 0 ? (
                <div className="text-xs text-muted-foreground bg-muted p-4 text-center">
                  No app sessions found
                </div>
              ) : (
                <div className="space-y-2 max-h-[240px] overflow-y-auto scrollbar-thin">
                  {appSessionsMeta && (
                    <div className="text-xs text-muted-foreground mb-2">
                      {appSessions.length} of {appSessionsMeta.totalCount} sessions
                    </div>
                  )}
                  {appSessions.map((session: any, idx: number) => (
                    <div key={idx} className="bg-muted p-3 space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Database className="h-3 w-3 text-muted-foreground" />
                          <span className="text-xs font-mono font-medium">
                            {formatAddress(session.appSessionId || String(idx))}
                          </span>
                        </div>
                        {session.isClosed ? (
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <XCircle className="h-3 w-3" /> Closed
                          </span>
                        ) : (
                          <span className="text-xs text-accent flex items-center gap-1">
                            <CheckCircle2 className="h-3 w-3" /> Open
                          </span>
                        )}
                      </div>
                      <div className="grid grid-cols-4 gap-2 text-xs text-muted-foreground">
                        <div>
                          <span className="uppercase tracking-wider">App: {session.appDefinition.applicationId}</span>
                        </div>
                        <div>
                          <span className="uppercase tracking-wider">v{session.version?.toString()}</span>
                        </div>
                        <div>
                          <span className="uppercase tracking-wider">Quorum: {session.appDefinition.quorum}</span>
                        </div>
                        <div>
                          <span className="uppercase tracking-wider">{session.appDefinition.participants?.length || 0} participants</span>
                        </div>
                      </div>
                      {session.allocations && session.allocations.length > 0 && (
                        <div className="flex flex-wrap gap-2 pt-1">
                          {session.allocations.map((a: any, aidx: number) => (
                            <span key={aidx} className="text-xs font-mono bg-background px-2 py-0.5 border border-border">
                              {formatAddress(a.participant)}: {a.amount?.toString()} {a.asset}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* App Registry */}
            <div>
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
                <AppWindow className="h-3.5 w-3.5" />
                <span>My Apps</span>
                <button
                  onClick={fetchMyApps}
                  disabled={myAppsLoading}
                  className="ml-auto text-muted-foreground hover:text-foreground transition-colors"
                >
                  <RefreshCw className={`h-3 w-3 ${myAppsLoading ? 'animate-spin' : ''}`} />
                </button>
              </div>

              {/* Register new app form */}
              <div className="glass-input rounded-lg p-3 mb-3 space-y-2">
                <div className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                  <Plus className="h-3 w-3" />
                  Register New App
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={registerAppId}
                    onChange={(e) => setRegisterAppId(e.target.value)}
                    placeholder="app-id (lowercase, hyphens)"
                    className="h-8 flex-1 px-2.5 text-xs glass-input rounded-lg font-mono focus:outline-none focus:border-accent"
                  />
                  <input
                    type="text"
                    value={registerAppMetadata}
                    onChange={(e) => setRegisterAppMetadata(e.target.value)}
                    placeholder="Metadata (optional)"
                    className="h-8 flex-1 px-2.5 text-xs glass-input rounded-lg focus:outline-none focus:border-accent"
                  />
                </div>
                <div className="flex items-center justify-between">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={registerAppNoApproval}
                      onChange={(e) => setRegisterAppNoApproval(e.target.checked)}
                      className="accent-accent"
                    />
                    <span className="text-xs text-muted-foreground">No approval required for session creation</span>
                  </label>
                  <Button
                    onClick={handleRegisterApp}
                    disabled={registeringApp || !registerAppId}
                    size="sm"
                    className="h-7 text-xs gap-1.5"
                  >
                    {registeringApp ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3 w-3" />}
                    Register
                  </Button>
                </div>
              </div>

              {/* Apps list */}
              {myAppsLoading ? (
                <div className="flex items-center gap-2 text-xs text-muted-foreground py-4">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Loading apps...</span>
                </div>
              ) : myApps.length === 0 ? (
                <div className="text-xs text-muted-foreground bg-muted p-4 text-center">
                  No registered apps
                </div>
              ) : (
                <div className="space-y-2 max-h-[240px] overflow-y-auto scrollbar-thin">
                  {myApps.map((app) => (
                    <div key={app.id} className="bg-muted p-3 space-y-1.5">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-mono font-medium">{app.id}</span>
                        <span className="text-[10px] uppercase tracking-wider text-muted-foreground">v{app.version}</span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span>
                          Approval: {app.creation_approval_not_required
                            ? <span className="text-green-400">Not required</span>
                            : <span className="text-yellow-500">Required</span>
                          }
                        </span>
                        {app.metadata && <span className="truncate max-w-[200px]" title={app.metadata}>Meta: {app.metadata}</span>}
                      </div>
                      <div className="text-[10px] text-muted-foreground">
                        Created {new Date(Number(app.created_at) * 1000).toLocaleString()}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Session Keys */}
            <div>
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
                <Key className="h-3.5 w-3.5" />
                <span>Session Keys</span>
                <button
                  onClick={fetchKeyStates}
                  disabled={keyStatesLoading}
                  className="ml-auto text-muted-foreground hover:text-foreground transition-colors"
                >
                  <RefreshCw className={`h-3 w-3 ${keyStatesLoading ? 'animate-spin' : ''}`} />
                </button>
              </div>
              {keyStatesLoading ? (
                <div className="flex items-center gap-2 text-xs text-muted-foreground py-4">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Loading keys...</span>
                </div>
              ) : keyStates.filter(ks => ks.assets.length > 0).length === 0 ? (
                <div className="text-xs text-muted-foreground bg-muted p-4 text-center">
                  No active session keys
                </div>
              ) : (
                <div className="space-y-2">
                  {keyStates.filter(ks => ks.assets.length > 0).map((ks) => {
                    const expiresAt = Number(ks.expires_at);
                    const isExpired = expiresAt <= Math.floor(Date.now() / 1000);
                    const expiresDate = new Date(expiresAt * 1000);
                    const revokeId = `${ks.session_key}-${ks.version}`;
                    const isCurrentKey = sessionKey?.active &&
                      sessionKey.address.toLowerCase() === ks.session_key.toLowerCase();
                    return (
                      <div key={revokeId} className="bg-muted p-3 flex items-center gap-3">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-mono font-medium">{formatAddress(ks.session_key)}</span>
                            {isCurrentKey && (
                              <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 bg-accent/20 text-accent font-semibold">
                                Current
                              </span>
                            )}
                          </div>
                          <div className="flex items-center gap-3 text-xs text-muted-foreground mt-1">
                            <span>v{ks.version}</span>
                            <span>{ks.assets.join(', ')}</span>
                            {isExpired ? (
                              <span className="text-destructive">Expired</span>
                            ) : (
                              <span>Expires {expiresDate.toLocaleString()}</span>
                            )}
                          </div>
                        </div>
                        {!isExpired && (
                          <Button
                            variant="destructive"
                            onClick={() => handleRevokeKey(ks)}
                            disabled={revokingKey === revokeId}
                            size="sm"
                            className="h-7 text-xs flex-shrink-0"
                          >
                            {revokingKey === revokeId ? (
                              <Loader2 className="h-3 w-3 animate-spin" />
                            ) : 'Revoke'}
                          </Button>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Close Channel (Danger Zone) */}
            <div className="pt-4 border-t border-border">
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-destructive mb-3">
                <AlertTriangle className="h-3.5 w-3.5" />
                <span>Danger Zone</span>
              </div>
              <div className="flex items-center justify-between bg-destructive/5 border border-destructive/20 p-4">
                <div>
                  <div className="text-sm font-medium">Close Channel</div>
                  <div className="text-xs text-muted-foreground">
                    Finalize and close your payment channel.
                  </div>
                </div>
                <Button
                  onClick={() => setActiveModal('close')}
                  variant="destructive"
                  size="sm"
                  className="flex-shrink-0 ml-4"
                >
                  <XCircle className="h-3.5 w-3.5 mr-1.5" />
                  Close
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Transaction Detail Modal */}
      {selectedTx && (() => {
        const { label, isIncoming } = getTxDisplay(selectedTx);
        return (
          <div
            className="fixed inset-0 bg-black/70 backdrop-blur-md flex items-center justify-center z-50"
            onClick={() => setSelectedTx(null)}
          >
            <div
              className="glass-heavy rounded-xl p-6 max-w-md w-full mx-4 animate-scale-in"
              onClick={e => e.stopPropagation()}
            >
              <div className="flex items-center justify-between mb-5">
                <div className="text-lg font-semibold uppercase tracking-tight">{label}</div>
                <button
                  onClick={() => setSelectedTx(null)}
                  className="text-muted-foreground hover:text-foreground transition-colors p-1"
                >
                  <XCircle className="h-5 w-5" />
                </button>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between py-2 border-b border-border/40">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground">Amount</span>
                  <span className={`text-sm font-mono font-medium ${isIncoming ? 'text-green-400' : ''}`}>
                    {isIncoming ? '+' : '-'}{selectedTx.amount?.toString()} {selectedTx.asset?.toUpperCase()}
                  </span>
                </div>

                <div className="flex items-center justify-between py-2 border-b border-border/40">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground">From</span>
                  <button
                    onClick={() => navigator.clipboard.writeText(selectedTx.fromAccount)}
                    title={selectedTx.fromAccount}
                    className="text-xs font-mono hover:text-accent transition-colors"
                  >
                    {formatAddress(selectedTx.fromAccount)}
                  </button>
                </div>

                <div className="flex items-center justify-between py-2 border-b border-border/40">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground">To</span>
                  <button
                    onClick={() => navigator.clipboard.writeText(selectedTx.toAccount)}
                    title={selectedTx.toAccount}
                    className="text-xs font-mono hover:text-accent transition-colors"
                  >
                    {formatAddress(selectedTx.toAccount)}
                  </button>
                </div>

                <div className="flex items-center justify-between py-2 border-b border-border/40">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground">Time</span>
                  <span className="text-xs">
                    {selectedTx.createdAt ? new Date(selectedTx.createdAt).toLocaleString() : '—'}
                  </span>
                </div>

                <div className="py-2 border-b border-border/40">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground block mb-1">Transaction ID</span>
                  <span className="text-xs font-mono break-all text-muted-foreground">{selectedTx.id}</span>
                </div>

              </div>

              <Button onClick={() => setSelectedTx(null)} variant="outline" className="w-full mt-5">
                Close
              </Button>
            </div>
          </div>
        );
      })()}

      {/* Action Modal */}
      {activeModal && (
        <ActionModal
          type={activeModal}
          client={client}
          chainId={chainId}
          asset={asset}
          showStatus={showStatus}
          onClose={() => setActiveModal(null)}
          onComplete={handleActionComplete}
        />
      )}
    </div>
  );
}
