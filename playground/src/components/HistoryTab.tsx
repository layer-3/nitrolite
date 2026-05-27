import { useState, useEffect, useCallback } from 'react';
import { RefreshCw, ChevronRight, ChevronDown, Filter, X, Copy, Check } from 'lucide-react';
import { TransactionType } from '@yellow-org/sdk';
import type { Client, Transaction, Blockchain, State } from '@yellow-org/sdk';
import type { Address } from 'viem';
import { tokenIconUrl } from '../icons';
import { formatAddress, timeAgo } from '../utils';

function formatStateId(id: string): string {
  if (id.length <= 14) return id;
  return `${id.slice(0, 6)}…${id.slice(-6)}`;
}

const PAGE_SIZE = 25;
const FETCH_LIMIT = 200;

type TxVariant = 'deposit' | 'withdraw' | 'transfer' | 'finalize' | 'muted';

interface HistoryFilters {
  types: TransactionType[];
  assets: string[];
  fromAddress: string;
  toAddress: string;
}

interface Props {
  client: Client | null;
  address: Address | null;
  chains: Blockchain[];
}

const EMPTY_FILTERS: HistoryFilters = { types: [], assets: [], fromAddress: '', toAddress: '' };

const TX_LABELS: Partial<Record<TransactionType, string>> = {
  [TransactionType.HomeDeposit]:    'Home Deposit',
  [TransactionType.HomeWithdrawal]: 'Home Withdrawal',
  [TransactionType.EscrowDeposit]:  'Escrow Deposit',
  [TransactionType.EscrowWithdraw]: 'Escrow Withdraw',
  [TransactionType.Transfer]:       'Transfer',
  [TransactionType.Commit]:         'Commit',
  [TransactionType.Release]:        'Release',
  [TransactionType.Rebalance]:      'Rebalance',
  [TransactionType.Migrate]:        'Migrate',
  [TransactionType.EscrowLock]:     'Escrow Lock',
  [TransactionType.MutualLock]:     'Mutual Lock',
  [TransactionType.Finalize]:       'Finalize',
};

const ALL_TX_TYPES = [
  TransactionType.HomeDeposit,
  TransactionType.HomeWithdrawal,
  TransactionType.EscrowDeposit,
  TransactionType.EscrowWithdraw,
  TransactionType.Transfer,
  TransactionType.Commit,
  TransactionType.Release,
  TransactionType.Rebalance,
  TransactionType.Migrate,
  TransactionType.EscrowLock,
  TransactionType.MutualLock,
  TransactionType.Finalize,
];

function txVariant(type: TransactionType): TxVariant {
  if (type === TransactionType.HomeDeposit || type === TransactionType.EscrowDeposit) return 'deposit';
  if (type === TransactionType.HomeWithdrawal || type === TransactionType.EscrowWithdraw) return 'withdraw';
  if (type === TransactionType.Transfer) return 'transfer';
  if (type === TransactionType.Finalize) return 'finalize';
  return 'muted';
}

const VARIANT_STYLE: Record<TxVariant, { color: string; bg: string }> = {
  deposit:  { color: 'var(--success)',    bg: 'rgba(34,197,94,0.12)'   },
  withdraw: { color: 'var(--error)',      bg: 'rgba(239,68,68,0.12)'   },
  transfer: { color: '#60a5fa',           bg: 'rgba(96,165,250,0.12)'  },
  finalize: { color: '#a78bfa',           bg: 'rgba(167,139,250,0.12)' },
  muted:    { color: 'var(--text-muted)', bg: 'rgba(255,255,255,0.06)' },
};

function amountPrefix(type: TransactionType, toAccount?: string, address?: Address | null): string {
  const v = txVariant(type);
  if (v === 'deposit') return '+';
  if (v === 'withdraw') return '−';
  if (v === 'transfer') {
    if (address && toAccount && toAccount.toLowerCase() === address.toLowerCase()) return '+';
    return '−';
  }
  return '';
}

function isOnChainTx(type: TransactionType): boolean {
  return type === TransactionType.HomeDeposit ||
    type === TransactionType.HomeWithdrawal ||
    type === TransactionType.EscrowDeposit ||
    type === TransactionType.EscrowWithdraw ||
    type === TransactionType.Finalize;
}

type ConfirmStatus = 'cosigned' | 'pending';

function getConfirmStatus(
  tx: Transaction,
  address: Address | null,
  latestStates: Record<string, State | null>,
): ConfirmStatus {
  // On-chain txs have their own confirmation model; off-chain senders are co-signed synchronously.
  if (isOnChainTx(tx.txType)) return 'cosigned';
  if (!address || tx.toAccount.toLowerCase() !== address.toLowerCase()) return 'cosigned';

  const state = latestStates[tx.asset];
  if (!state || !tx.receiverNewStateId) return 'cosigned';

  // Pending = latest state IS this receiver state AND the user hasn't signed it yet.
  if (
    state.id.toLowerCase() === tx.receiverNewStateId.toLowerCase() &&
    state.nodeSig &&
    !state.userSig
  ) return 'pending';

  return 'cosigned';
}

// ── Sub-components ────────────────────────────────────────────────────────────

function TypeBadge({ type, isActive, onClick }: {
  type: TransactionType;
  isActive: boolean;
  onClick?: (e: React.MouseEvent) => void;
}) {
  const label = TX_LABELS[type] ?? `Type ${type}`;
  const { color, bg } = VARIANT_STYLE[txVariant(type)];
  return (
    <span
      className={`qf${isActive ? ' qf-on' : ''}`}
      data-tip={isActive ? `Remove filter: ${label}` : `Quick filter: ${label}`}
      onClick={onClick}
    >
      <span style={{
        display: 'inline-flex',
        alignItems: 'center',
        fontSize: '11px',
        fontWeight: 500,
        padding: '2px 7px',
        borderRadius: '4px',
        background: bg,
        color,
        whiteSpace: 'nowrap',
      }}>
        {label}
      </span>
    </span>
  );
}

function AddrCell({ value, isActive, onClick }: {
  value: string;
  isActive: boolean;
  onClick: (e: React.MouseEvent) => void;
}) {
  const [copied, setCopied] = useState(false);
  const display = formatAddress(value);

  const onCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch { /* clipboard denied */ }
  };

  return (
    <div className="addr-cell">
      <span
        className={`qf${isActive ? ' qf-on' : ''}`}
        data-tip={isActive ? `Remove filter: ${display}` : `Quick filter: ${display}`}
        onClick={onClick}
      >
        <span className="mono" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{display}</span>
      </span>
      <button className="copy-btn-addr" onClick={onCopy} title={copied ? 'Copied' : 'Copy address'}>
        {copied
          ? <Check size={11} style={{ color: 'var(--success)' }} />
          : <Copy size={11} />}
      </button>
    </div>
  );
}

function AssetCell({ asset, isActive, onClick }: {
  asset: string;
  isActive: boolean;
  onClick: (e: React.MouseEvent) => void;
}) {
  const iconUrl = tokenIconUrl(asset);
  const [imgErr, setImgErr] = useState(false);
  return (
    <span
      className={`qf${isActive ? ' qf-on' : ''}`}
      data-tip={isActive ? `Remove filter: ${asset}` : `Quick filter: ${asset}`}
      onClick={onClick}
      style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
    >
      {iconUrl && !imgErr ? (
        <img
          src={iconUrl}
          alt={asset}
          width={16}
          height={16}
          style={{ borderRadius: '50%', flexShrink: 0 }}
          onError={() => setImgErr(true)}
        />
      ) : (
        <span style={{
          width: 16, height: 16, borderRadius: '50%', background: 'var(--border)',
          display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
          fontSize: '9px', fontWeight: 700, color: 'var(--text-muted)', flexShrink: 0,
        }}>
          {asset.slice(0, 1).toUpperCase()}
        </span>
      )}
      <span style={{ fontWeight: 500, fontSize: '13px' }}>{asset.toUpperCase()}</span>
    </span>
  );
}

// ── Filter popover ────────────────────────────────────────────────────────────

function FilterIcon({ active }: { active: boolean }) {
  return (
    <Filter
      size={11}
      style={{ color: active ? 'var(--accent)' : 'var(--text-muted)', flexShrink: 0 }}
    />
  );
}

// ── Detail panel ─────────────────────────────────────────────────────────────

type TsFormat = 'date' | 'unix';

function DetailPanel({ tx, tsFormat, onToggleTsFormat, confirmStatus }: {
  tx: Transaction;
  tsFormat: TsFormat;
  onToggleTsFormat: () => void;
  confirmStatus: ConfirmStatus;
}) {
  const onChain = isOnChainTx(tx.txType);
  const tsDate = tx.createdAt.toISOString().replace('T', ' ').slice(0, 19) + ' UTC';
  const tsUnix = Math.floor(tx.createdAt.getTime() / 1000).toString();
  const tsValue = tsFormat === 'date' ? tsDate : tsUnix;
  const tsTip = tsFormat === 'date' ? 'Display as Unix seconds' : 'Display as UTC date';

  const copyInline = (value: string) => {
    navigator.clipboard.writeText(value).catch(() => {});
  };

  return (
    <div style={{
      width: '100%',
      padding: '16px 20px',
      borderTop: '1px solid var(--border)',
      background: 'rgba(10,10,10,0.3)',
      display: 'grid',
      gridTemplateColumns: 'minmax(0,1fr) minmax(0,1fr) minmax(0,1fr)',
      gap: '16px 12px',
      boxSizing: 'border-box',
    }}>
      <DetailField
        label="Sender new state ID"
        value={tx.senderNewStateId}
        displayValue={tx.senderNewStateId ? formatStateId(tx.senderNewStateId) : undefined}
        onCopy={copyInline}
      />
      <DetailField
        label="Receiver new state ID"
        value={tx.receiverNewStateId}
        displayValue={tx.receiverNewStateId ? formatStateId(tx.receiverNewStateId) : undefined}
        onCopy={copyInline}
      />
      {/* Timestamp — rendered inline so the value itself is the interactive target */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', minWidth: 0 }}>
        <span style={{ fontSize: '11px', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Timestamp
        </span>
        <span
          className="ts-toggle"
          data-tip={tsTip}
          onClick={e => { e.stopPropagation(); onToggleTsFormat(); }}
        >
          <span
            className={tsFormat === 'unix' ? 'mono' : undefined}
            style={{ fontSize: '12px', color: 'var(--text-primary)' }}
          >
            {tsValue}
          </span>
        </span>
      </div>

      {/* Confirmation timeline — spans all 3 columns */}
      <div style={{ gridColumn: '1 / -1' }}>
        <span style={{ fontSize: '11px', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '8px' }}>
          Confirmation
        </span>
        <ConfirmTimeline onChain={onChain} status={confirmStatus} />
      </div>
    </div>
  );
}

function DetailField({ label, value, displayValue, mono = true, onCopy }: {
  label: string;
  value?: string;
  displayValue?: string;
  mono?: boolean;
  onCopy?: (v: string) => void;
}) {
  const shown = displayValue ?? value;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', minWidth: 0 }}>
      <span style={{ fontSize: '11px', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</span>
      {shown ? (
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', minWidth: 0 }}>
          <span
            className={mono ? 'mono' : undefined}
            style={{ fontSize: '12px', color: 'var(--text-primary)', wordBreak: 'break-all', minWidth: 0 }}
          >
            {shown}
          </span>
          {onCopy && value && (
            <button
              onClick={e => { e.stopPropagation(); onCopy(value); }}
              style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '1px', color: 'var(--text-muted)', display: 'inline-flex', flexShrink: 0 }}
              title="Copy"
            >
              <Copy size={11} />
            </button>
          )}
        </span>
      ) : (
        <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>—</span>
      )}
    </div>
  );
}

function ConfirmTimeline({ onChain, status }: { onChain: boolean; status: ConfirmStatus }) {
  type Step = { label: string; done: boolean };
  const steps: Step[] = onChain
    ? [
        { label: 'Signed', done: true },
        { label: 'Broadcasted', done: true },
        { label: 'Confirmed', done: true },
      ]
    : [
        { label: 'Signed', done: true },
        { label: 'Co-signed', done: status === 'cosigned' },
      ];

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 0 }}>
      {steps.map(({ label, done }, i) => (
        <span key={label} style={{ display: 'inline-flex', alignItems: 'center' }}>
          <span style={{
            fontSize: '12px',
            borderRadius: '4px',
            padding: '2px 8px',
            ...(done
              ? { color: 'var(--text-primary)', background: 'rgba(34,197,94,0.12)', border: '1px solid rgba(34,197,94,0.3)' }
              : { color: 'var(--text-muted)', background: 'rgba(255,255,255,0.04)', border: '1px solid var(--border)' }
            ),
          }}>
            {label}{!done ? '…' : ''}
          </span>
          {i < steps.length - 1 && (
            <span style={{ width: '28px', height: '1px', background: 'var(--border)', display: 'inline-block', margin: '0 2px' }} />
          )}
        </span>
      ))}
    </div>
  );
}

// ── Table header with filter popover ─────────────────────────────────────────

function TxIdCell({ id }: { id: string }) {
  const [copied, setCopied] = useState(false);
  const display = formatAddress(id);

  const onCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(id);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch { /* clipboard denied */ }
  };

  return (
    <div className="addr-cell">
      <span className="mono" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>{display}</span>
      <button className="copy-btn-addr" onClick={onCopy} title={copied ? 'Copied' : 'Copy ID'}>
        {copied
          ? <Check size={11} style={{ color: 'var(--success)' }} />
          : <Copy size={11} />}
      </button>
    </div>
  );
}

function ThFilter({ label, isActive, isOpen, onToggle, width, children }: {
  label: string;
  isActive: boolean;
  isOpen: boolean;
  onToggle: (e: React.MouseEvent) => void;
  width?: string;
  children?: React.ReactNode;
}) {
  return (
    <th style={{ padding: '10px 12px', textAlign: 'left', fontWeight: 500, fontSize: '12px', color: 'var(--text-muted)', whiteSpace: 'nowrap', ...(width ? { width } : {}) }}>
      <div className="th-popover-wrap">
        <div
          className={`th-inner filterable${isActive ? ' filter-active' : ''}`}
          onClick={onToggle}
        >
          {label}
          <FilterIcon active={isActive} />
        </div>
        {isOpen && (
          <div className="th-popover" onClick={e => e.stopPropagation()}>
            {children}
          </div>
        )}
      </div>
    </th>
  );
}

// ── Main component ────────────────────────────────────────────────────────────

export default function HistoryTab({ client, address }: Props) {
  const [allTxs, setAllTxs] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [filters, setFilters] = useState<HistoryFilters>(EMPTY_FILTERS);
  // pending filter edits inside popovers before Apply
  const [pendingTypes, setPendingTypes] = useState<TransactionType[]>([]);
  const [pendingAssets, setPendingAssets] = useState<string[]>([]);
  const [pendingFrom, setPendingFrom] = useState('');
  const [pendingTo, setPendingTo] = useState('');

  const [latestStates, setLatestStates] = useState<Record<string, State | null>>({});

  const [openPopover, setOpenPopover] = useState<'type' | 'asset' | 'from' | 'to' | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [tsFormat, setTsFormat] = useState<TsFormat>('date');
  const [, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 30000);
    return () => clearInterval(id);
  }, []);

  const fetchTxs = useCallback(async () => {
    if (!client || !address) return;
    setIsLoading(true);
    setError(null);
    try {
      const { transactions } = await client.getTransactions(address, {
        pageSize: FETCH_LIMIT,
        page: 1,
      });
      setAllTxs(transactions.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load transactions');
    } finally {
      setIsLoading(false);
    }
  }, [client, address]);

  useEffect(() => {
    setPage(1);
    fetchTxs();
  }, [fetchTxs]);

  // After transactions load, fetch the latest state for each asset where the user is receiver.
  useEffect(() => {
    if (!client || !address || allTxs.length === 0) return;
    const assets = [...new Set(
      allTxs
        .filter(tx => !isOnChainTx(tx.txType) && tx.toAccount.toLowerCase() === address.toLowerCase())
        .map(tx => tx.asset)
    )];
    if (assets.length === 0) return;
    Promise.all(
      assets.map(asset =>
        client.getLatestState(address, asset, false)
          .then(state => [asset, state] as const)
          .catch(() => [asset, null] as const)
      )
    ).then(entries => setLatestStates(Object.fromEntries(entries)));
  }, [client, address, allTxs]);

  // Close popover on outside click
  useEffect(() => {
    if (!openPopover) return;
    const handler = (e: MouseEvent) => {
      if (!(e.target as Element).closest('.th-popover-wrap')) {
        setOpenPopover(null);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [openPopover]);

  // Client-side filter
  const filteredTxs = allTxs.filter(tx => {
    if (filters.types.length && !filters.types.includes(tx.txType)) return false;
    if (filters.assets.length && !filters.assets.includes(tx.asset)) return false;
    if (filters.fromAddress && !tx.fromAccount.toLowerCase().includes(filters.fromAddress.toLowerCase())) return false;
    if (filters.toAddress && !tx.toAccount.toLowerCase().includes(filters.toAddress.toLowerCase())) return false;
    return true;
  });

  const totalCount = filteredTxs.length;
  const pageCount = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const pageTxs = filteredTxs.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const availableAssets = [...new Set(allTxs.map(t => t.asset))].sort();

  // Active filter count for the badge
  const activeFilterCount =
    (filters.types.length ? 1 : 0) +
    (filters.assets.length ? 1 : 0) +
    (filters.fromAddress ? 1 : 0) +
    (filters.toAddress ? 1 : 0);

  // ── Quick filter handlers ──────────────────────────────────────────────────

  const toggleType = (type: TransactionType, e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters(f => ({
      ...f,
      types: f.types.includes(type) ? f.types.filter(t => t !== type) : [...f.types, type],
    }));
    setPage(1);
  };

  const toggleAsset = (asset: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters(f => ({
      ...f,
      assets: f.assets.includes(asset) ? f.assets.filter(a => a !== asset) : [...f.assets, asset],
    }));
    setPage(1);
  };

  const toggleFrom = (addr: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters(f => ({ ...f, fromAddress: f.fromAddress === addr ? '' : addr }));
    setPage(1);
  };

  const toggleTo = (addr: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setFilters(f => ({ ...f, toAddress: f.toAddress === addr ? '' : addr }));
    setPage(1);
  };

  // ── Popover open/close helpers ─────────────────────────────────────────────

  const openFilter = (col: typeof openPopover, e: React.MouseEvent) => {
    e.stopPropagation();
    if (openPopover === col) {
      setOpenPopover(null);
      return;
    }
    // Sync pending state with current filters
    setPendingTypes(filters.types);
    setPendingAssets(filters.assets);
    setPendingFrom(filters.fromAddress);
    setPendingTo(filters.toAddress);
    setOpenPopover(col);
  };

  const applyTypeFilter = () => {
    setFilters(f => ({ ...f, types: pendingTypes }));
    setOpenPopover(null);
    setPage(1);
  };

  const applyAssetFilter = () => {
    setFilters(f => ({ ...f, assets: pendingAssets }));
    setOpenPopover(null);
    setPage(1);
  };

  const applyFromFilter = () => {
    setFilters(f => ({ ...f, fromAddress: pendingFrom }));
    setOpenPopover(null);
    setPage(1);
  };

  const applyToFilter = () => {
    setFilters(f => ({ ...f, toAddress: pendingTo }));
    setOpenPopover(null);
    setPage(1);
  };

  const clearFilter = (col: keyof HistoryFilters) => {
    setFilters(f => ({
      ...f,
      ...(col === 'types' ? { types: [] } : {}),
      ...(col === 'assets' ? { assets: [] } : {}),
      ...(col === 'fromAddress' ? { fromAddress: '' } : {}),
      ...(col === 'toAddress' ? { toAddress: '' } : {}),
    }));
    setOpenPopover(null);
    setPage(1);
  };

  // ── Render ─────────────────────────────────────────────────────────────────

  const thBase: React.CSSProperties = {
    padding: '10px 12px',
    textAlign: 'left',
    fontWeight: 500,
    fontSize: '12px',
    color: 'var(--text-muted)',
    whiteSpace: 'nowrap',
    borderBottom: '1px solid var(--border)',
    background: 'var(--bg-elevated)',
  };

  const tdBase: React.CSSProperties = {
    padding: '10px 12px',
    verticalAlign: 'middle',
    borderBottom: '1px solid var(--border)',
  };

  return (
    <div className="card">
      {/* Card header */}
      <div className="card-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <span className="card-title">Transaction History</span>
          {activeFilterCount > 0 && (
            <span style={{
              fontSize: '11px',
              color: 'var(--accent)',
              background: 'var(--accent-dim)',
              borderRadius: '999px',
              padding: '2px 8px',
            }}>
              {activeFilterCount} filter{activeFilterCount !== 1 ? 's' : ''}
            </span>
          )}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {activeFilterCount > 0 && (
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => { setFilters(EMPTY_FILTERS); setPage(1); }}
            >
              <X size={11} />
              Clear filters
            </button>
          )}
          <button
            className="btn btn-ghost btn-sm"
            onClick={fetchTxs}
            disabled={isLoading}
          >
            <RefreshCw size={12} className={isLoading ? 'animate-spin' : ''} />
            {isLoading ? 'Loading…' : 'Refresh'}
          </button>
        </div>
      </div>

      {/* Error state */}
      {error && (
        <div style={{ padding: '16px 20px', color: 'var(--error)', fontSize: '13px' }}>
          {error}
        </div>
      )}

      {/* Empty / loading state */}
      {!error && !isLoading && allTxs.length === 0 && (
        <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '13px' }}>
          {client && address ? 'No transactions found.' : 'Connect wallet to view history.'}
        </div>
      )}

      {/* Table */}
      {(allTxs.length > 0 || isLoading) && (
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', tableLayout: 'fixed' }}>
            <thead>
              <tr>
                {/* expand chevron */}
                <th style={{ ...thBase, width: '36px', padding: '10px 8px 10px 16px' }} />

                {/* Time */}
                <th style={{ ...thBase, width: '80px' }}>Time</th>

                {/* Type with filter popover */}
                <ThFilter
                  label="Type"
                  isActive={filters.types.length > 0}
                  isOpen={openPopover === 'type'}
                  onToggle={e => openFilter('type', e)}
                >
                  <div className="th-popover-title">Filter by type</div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', maxHeight: '220px', overflowY: 'auto' }}>
                    {ALL_TX_TYPES.map(t => (
                      <label key={t} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '12px' }}>
                        <input
                          type="checkbox"
                          checked={pendingTypes.includes(t)}
                          onChange={() => setPendingTypes(prev =>
                            prev.includes(t) ? prev.filter(x => x !== t) : [...prev, t]
                          )}
                          style={{ accentColor: 'var(--accent)' }}
                        />
                        <TypeBadge type={t} isActive={false} />
                      </label>
                    ))}
                  </div>
                  <div className="th-popover-actions">
                    <button className="btn btn-ghost btn-sm" onClick={() => clearFilter('types')}>Clear</button>
                    <button className="btn btn-primary btn-sm" onClick={applyTypeFilter}>Apply</button>
                  </div>
                </ThFilter>

                {/* Asset with filter popover */}
                <ThFilter
                  label="Asset"
                  isActive={filters.assets.length > 0}
                  isOpen={openPopover === 'asset'}
                  onToggle={e => openFilter('asset', e)}
                  width="15%"
                >
                  <div className="th-popover-title">Filter by asset</div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    {availableAssets.map(a => (
                      <label key={a} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '12px' }}>
                        <input
                          type="checkbox"
                          checked={pendingAssets.includes(a)}
                          onChange={() => setPendingAssets(prev =>
                            prev.includes(a) ? prev.filter(x => x !== a) : [...prev, a]
                          )}
                          style={{ accentColor: 'var(--accent)' }}
                        />
                        <AssetCell asset={a} isActive={false} onClick={e => e.stopPropagation()} />
                      </label>
                    ))}
                  </div>
                  <div className="th-popover-actions">
                    <button className="btn btn-ghost btn-sm" onClick={() => clearFilter('assets')}>Clear</button>
                    <button className="btn btn-primary btn-sm" onClick={applyAssetFilter}>Apply</button>
                  </div>
                </ThFilter>

                {/* Amount */}
                <th style={{ ...thBase, width: '110px', textAlign: 'right' }}>Amount</th>

                {/* From with filter popover */}
                <ThFilter
                  label="From"
                  isActive={!!filters.fromAddress}
                  isOpen={openPopover === 'from'}
                  onToggle={e => openFilter('from', e)}
                  width="15%"
                >
                  <div className="th-popover-title">Filter by from address</div>
                  <input
                    className="th-text-input"
                    type="text"
                    placeholder="0x…"
                    value={pendingFrom}
                    onChange={e => setPendingFrom(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && applyFromFilter()}
                    autoFocus
                  />
                  <div className="th-popover-actions">
                    <button className="btn btn-ghost btn-sm" onClick={() => clearFilter('fromAddress')}>Clear</button>
                    <button className="btn btn-primary btn-sm" onClick={applyFromFilter}>Apply</button>
                  </div>
                </ThFilter>

                {/* To with filter popover */}
                <ThFilter
                  label="To"
                  isActive={!!filters.toAddress}
                  isOpen={openPopover === 'to'}
                  onToggle={e => openFilter('to', e)}
                  width="15%"
                >
                  <div className="th-popover-title">Filter by to address</div>
                  <input
                    className="th-text-input"
                    type="text"
                    placeholder="0x…"
                    value={pendingTo}
                    onChange={e => setPendingTo(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && applyToFilter()}
                    autoFocus
                  />
                  <div className="th-popover-actions">
                    <button className="btn btn-ghost btn-sm" onClick={() => clearFilter('toAddress')}>Clear</button>
                    <button className="btn btn-primary btn-sm" onClick={applyToFilter}>Apply</button>
                  </div>
                </ThFilter>

                {/* Tx ID */}
                <th style={{ ...thBase, width: '15%' }}>Tx ID</th>
              </tr>
            </thead>
            <tbody>
              {isLoading && allTxs.length === 0 ? (
                <tr>
                  <td colSpan={8} style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)' }}>
                    <span className="spinner" style={{ marginRight: '8px' }} />
                    Loading transactions…
                  </td>
                </tr>
              ) : pageTxs.length === 0 ? (
                <tr>
                  <td colSpan={8} style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '13px' }}>
                    No transactions match the current filters.
                  </td>
                </tr>
              ) : (
                pageTxs.map(tx => {
                  const isExpanded = expandedId === tx.id;
                  const variant = txVariant(tx.txType);
                  const { color: amtColor } = VARIANT_STYLE[variant];
                  const prefix = amountPrefix(tx.txType, tx.toAccount, address);
                  const confirmStatus = getConfirmStatus(tx, address, latestStates);

                  return (
                    <>
                      <tr
                        key={tx.id}
                        onClick={() => setExpandedId(isExpanded ? null : tx.id)}
                        style={{
                          cursor: 'pointer',
                          background: isExpanded ? 'rgba(255,255,255,0.02)' : undefined,
                          transition: 'background 0.1s',
                        }}
                        onMouseEnter={e => { if (!isExpanded) (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.02)'; }}
                        onMouseLeave={e => { if (!isExpanded) (e.currentTarget as HTMLElement).style.background = ''; }}
                      >
                        {/* Expand chevron */}
                        <td style={{ ...tdBase, padding: '10px 8px 10px 16px', width: '32px' }}>
                          {isExpanded
                            ? <ChevronDown size={14} style={{ color: 'var(--text-muted)' }} />
                            : <ChevronRight size={14} style={{ color: 'var(--text-muted)' }} />}
                        </td>

                        {/* Time */}
                        <td style={{ ...tdBase, color: 'var(--text-muted)', fontSize: '12px', whiteSpace: 'nowrap' }}>
                          <span
                            className="tooltip-wrap"
                            title={tx.createdAt.toISOString().replace('T', ' ').slice(0, 19) + ' UTC'}
                          >
                            {timeAgo(tx.createdAt)}
                          </span>
                        </td>

                        {/* Type */}
                        <td style={tdBase}>
                          <TypeBadge
                            type={tx.txType}
                            isActive={filters.types.includes(tx.txType)}
                            onClick={e => toggleType(tx.txType, e)}
                          />
                        </td>

                        {/* Asset */}
                        <td style={tdBase}>
                          <AssetCell
                            asset={tx.asset}
                            isActive={filters.assets.includes(tx.asset)}
                            onClick={e => toggleAsset(tx.asset, e)}
                          />
                        </td>

                        {/* Amount */}
                        <td style={{ ...tdBase, textAlign: 'right' }}>
                          <span className="mono" style={{ color: amtColor, fontSize: '13px', whiteSpace: 'nowrap' }}>
                            {prefix}{tx.amount.toString()}
                          </span>
                        </td>

                        {/* From */}
                        <td style={tdBase}>
                          <AddrCell
                            value={tx.fromAccount}
                            isActive={filters.fromAddress === tx.fromAccount}
                            onClick={e => toggleFrom(tx.fromAccount, e)}
                          />
                        </td>

                        {/* To */}
                        <td style={tdBase}>
                          <AddrCell
                            value={tx.toAccount}
                            isActive={filters.toAddress === tx.toAccount}
                            onClick={e => toggleTo(tx.toAccount, e)}
                          />
                        </td>

                        {/* Tx ID */}
                        <td style={{ ...tdBase, color: 'var(--text-muted)', fontSize: '12px' }}>
                          {tx.id ? <TxIdCell id={tx.id} /> : <span>—</span>}
                        </td>
                      </tr>

                      {/* Expanded detail */}
                      {isExpanded && (
                        <tr key={`${tx.id}-detail`}>
                          <td colSpan={8} style={{ padding: 0 }}>
                            <DetailPanel
                              tx={tx}
                              tsFormat={tsFormat}
                              onToggleTsFormat={() => setTsFormat(f => f === 'date' ? 'unix' : 'date')}
                              confirmStatus={confirmStatus}
                            />
                          </td>
                        </tr>
                      )}
                    </>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination footer */}
      {totalCount > 0 && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '12px 20px',
          borderTop: '1px solid var(--border)',
          fontSize: '12px',
          color: 'var(--text-muted)',
        }}>
          <span>
            Showing {Math.min((page - 1) * PAGE_SIZE + 1, totalCount)}–{Math.min(page * PAGE_SIZE, totalCount)} of {totalCount} transaction{totalCount !== 1 ? 's' : ''}
          </span>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              ← Prev
            </button>
            <span style={{ minWidth: '60px', textAlign: 'center' }}>
              {page} / {pageCount}
            </span>
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => setPage(p => Math.min(pageCount, p + 1))}
              disabled={page === pageCount}
            >
              Next →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
