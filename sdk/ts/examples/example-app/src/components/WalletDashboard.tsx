import { useState, useEffect, useCallback } from 'react';
import {
  ArrowDownToLine, ArrowUpFromLine, Send, XCircle,
  Shield, AlertTriangle, CheckCircle2, ArrowRightLeft,
  RefreshCw, Key, Loader2, ChevronDown, ChevronUp,
  Users, Database, Copy, Check,
} from 'lucide-react';
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import type { WalletClient } from 'viem';
import Decimal from 'decimal.js';
import {
  getChannelSessionKeyAuthMetadataHashV1,
  packChannelKeyStateV1,
} from '@yellow-org/sdk';
import type { Client, ChannelSessionKeyStateV1 } from '@yellow-org/sdk';
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
  return msg.toLowerCase().includes('allowance') && msg.toLowerCase().includes('sufficient');
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
      ]);

      if (results[0].status === 'fulfilled') setBalances(results[0].value as any[]);
      setLatestState(results[1].status === 'fulfilled' ? results[1].value : null);
      setSignedState(results[2].status === 'fulfilled' ? results[2].value : null);
      setHomeChannel(results[3].status === 'fulfilled' ? results[3].value : null);
      if (results[4].status === 'fulfilled') {
        setTransactions((results[4].value as any).transactions || []);
      }
    } catch (e) {
      console.error('Fetch error:', e);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [client, address, asset]);

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

      // Submit to clearnode
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
        <div className="bg-accent/10 border border-accent/30 px-4 py-3 flex items-center justify-between">
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

      {/* Top row: Balance + Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        {/* Balance Card */}
        <div className="lg:col-span-3 bg-card border border-border p-6 transition-colors duration-200">
          {/* Asset tabs */}
          <div className="flex gap-1.5 mb-5">
            {assets.map(a => (
              <button
                key={a}
                onClick={() => onAssetChange(a)}
                className={`px-3.5 py-1.5 text-xs font-semibold uppercase tracking-wider transition-all ${
                  a === asset
                    ? 'bg-accent text-accent-foreground'
                    : 'bg-muted text-muted-foreground hover:text-foreground'
                }`}
              >
                {a}
              </button>
            ))}
          </div>

          {/* Balance display */}
          <div className="mb-5">
            <div className="text-3xl lg:text-4xl font-semibold tracking-tight tabular-nums">
              {formatBalance(currentBalance?.balance)}
            </div>
            <div className="text-sm text-muted-foreground uppercase tracking-wider mt-1">
              {asset.toUpperCase()} Balance
            </div>
            <button
              onClick={() => {
                navigator.clipboard.writeText(address);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
              }}
              title={address}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors mt-2 font-mono"
            >
              {formatAddress(address)}
              {copied ? <Check className="h-3 w-3 text-green-400" /> : <Copy className="h-3 w-3" />}
            </button>
          </div>

          {/* Channel sync status */}
          <div className="flex items-center gap-3 text-xs">
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

        {/* Quick Actions */}
        <div className="lg:col-span-2 bg-card border border-border p-6 transition-colors duration-200">
          <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-4">
            Actions
          </div>
          <div className="grid grid-cols-3 gap-3">
            <button
              onClick={() => setActiveModal('deposit')}
              className="flex flex-col items-center gap-2.5 p-4 bg-muted hover:bg-muted/70 transition-colors group"
            >
              <div className="w-10 h-10 flex items-center justify-center bg-green-500/10 text-green-400 group-hover:bg-green-500/20 transition-colors">
                <ArrowDownToLine className="h-5 w-5" />
              </div>
              <span className="text-xs font-semibold uppercase tracking-wider">Deposit</span>
            </button>

            <button
              onClick={() => setActiveModal('withdraw')}
              className="flex flex-col items-center gap-2.5 p-4 bg-muted hover:bg-muted/70 transition-colors group"
            >
              <div className="w-10 h-10 flex items-center justify-center bg-orange-500/10 text-orange-400 group-hover:bg-orange-500/20 transition-colors">
                <ArrowUpFromLine className="h-5 w-5" />
              </div>
              <span className="text-xs font-semibold uppercase tracking-wider">Withdraw</span>
            </button>

            <button
              onClick={() => setActiveModal('transfer')}
              className="flex flex-col items-center gap-2.5 p-4 bg-muted hover:bg-muted/70 transition-colors group"
            >
              <div className="w-10 h-10 flex items-center justify-center bg-blue-500/10 text-blue-400 group-hover:bg-blue-500/20 transition-colors">
                <Send className="h-5 w-5" />
              </div>
              <span className="text-xs font-semibold uppercase tracking-wider">Send</span>
            </button>
          </div>
        </div>
      </div>

      {/* Activity Feed */}
      <div className="bg-card border border-border p-6 transition-colors duration-200">
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
                  className="flex items-center gap-3 py-2.5 px-2 border-b border-border/40 last:border-0 hover:bg-muted/30 transition-colors cursor-pointer"
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
      <div className="bg-card border border-border px-6 py-4 transition-colors duration-200">
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

      {/* Advanced Section Toggle */}
      <div className="bg-card border border-border transition-colors duration-200">
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="w-full px-6 py-3 flex items-center justify-between text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors"
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
            className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50"
            onClick={() => setSelectedTx(null)}
          >
            <div
              className="bg-card border border-border p-6 max-w-md w-full mx-4 animate-scale-in"
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
