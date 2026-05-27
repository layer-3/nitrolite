import { useState, useEffect } from 'react';
import { Key, RefreshCw, Plus, Copy, Check, Eye, EyeOff } from 'lucide-react';
import type { Address, WalletClient } from 'viem';
import type { Client, ChannelSessionKeyStateV1 } from '@yellow-org/sdk';
import type { StoredSessionKey } from '../sessionKey';
import { useSessionKeyManagement } from '../hooks/useSessionKeyManagement';
import SessionKeyRegisterForm from './SessionKeyRegisterForm';
import SessionKeyRevokeModal from './SessionKeyRevokeModal';

// ── Expiry format cycling ─────────────────────────────────────────────────────

type ExpFormat = 'relative' | 'date' | 'unix';

const NEXT_FMT: Record<ExpFormat, ExpFormat> = { relative: 'date', date: 'unix', unix: 'relative' };
const FMT_HINT: Record<ExpFormat, string> = {
  relative: 'Show as UTC date',
  date: 'Show as Unix timestamp',
  unix: 'Show as relative time',
};

function formatExpiry(expiresAtStr: string, fmt: ExpFormat): string {
  const expiresAt = Number(expiresAtStr);
  const now = Math.floor(Date.now() / 1000);
  if (fmt === 'unix') return expiresAtStr;
  if (fmt === 'date') return new Date(expiresAt * 1000).toISOString().replace('T', ' ').slice(0, 19) + ' UTC';
  const diff = expiresAt - now;
  if (diff <= 0) return 'expired';
  const h = Math.floor(diff / 3600);
  const m = Math.floor((diff % 3600) / 60);
  return h > 0 ? `${h}h ${m}m` : `${m}m`;
}

// ── Status derivation ─────────────────────────────────────────────────────────

type KeyStatus = 'active' | 'expiring' | 'expired' | 'revoked';

function getKeyStatus(key: ChannelSessionKeyStateV1): KeyStatus {
  const now = Math.floor(Date.now() / 1000);
  const expiresAt = Number(key.expires_at);
  // Revoked = registered with expiresAt well in the past (via revoke flow).
  // Check before 'expired' so the more specific branch wins.
  if (expiresAt < now - 60) return 'revoked';
  if (expiresAt <= now) return 'expired';
  if (expiresAt - now < 3600) return 'expiring'; // < 1 hour
  return 'active';
}

// ── Status badge ──────────────────────────────────────────────────────────────

const STATUS_STYLE: Record<KeyStatus, { bg: string; color: string; label: string }> = {
  active:   { bg: 'rgba(34,197,94,0.12)',   color: '#22c55e',            label: 'Active' },
  expiring: { bg: 'rgba(249,115,22,0.14)',  color: '#f97316',            label: 'Expiring Soon' },
  expired:  { bg: 'rgba(102,102,102,0.14)', color: 'var(--text-muted)',  label: 'Expired' },
  revoked:  { bg: 'rgba(239,68,68,0.12)',   color: 'var(--error)',       label: 'Revoked' },
};

function StatusBadge({ status }: { status: KeyStatus }) {
  const s = STATUS_STYLE[status];
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      borderRadius: 999, padding: '3px 9px', fontSize: 11, fontWeight: 500,
      background: s.bg, color: s.color,
    }}>
      <span style={{ width: 6, height: 6, borderRadius: '50%', background: s.color, flexShrink: 0 }} />
      {s.label}
    </span>
  );
}

// ── Props ─────────────────────────────────────────────────────────────────────

interface Props {
  client: Client | null;
  walletClient: WalletClient | null;
  address: Address | null;
  sessionKey: StoredSessionKey | null;
  allSessionKeys: StoredSessionKey[];
  supportedAssets: string[];
  onKeyActivated: (sk: StoredSessionKey) => void;
  onKeyCleared: () => void;
  onSelectKey: (sessionKeyAddress: Address) => void;
  onRefreshAllKeys: () => void;
}

// ── Main component ────────────────────────────────────────────────────────────

export default function SessionKeysTab({
  client,
  walletClient,
  address,
  sessionKey,
  allSessionKeys,
  supportedAssets,
  onKeyActivated,
  onKeyCleared,
  onSelectKey,
  onRefreshAllKeys,
}: Props) {
  const [expFmt, setExpFmt] = useState<ExpFormat>('relative');
  const [showExpired, setShowExpired] = useState(true);
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [keyForUpdate, setKeyForUpdate] = useState<ChannelSessionKeyStateV1 | null>(null);
  const [keyForRevoke, setKeyForRevoke] = useState<ChannelSessionKeyStateV1 | null>(null);
  const [copiedAddress, setCopiedAddress] = useState<string | null>(null);

  const mgmt = useSessionKeyManagement(client, address, walletClient);

  // Fetch on mount
  useEffect(() => { mgmt.fetchKeys(); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Copy address helper
  const copyAddr = (addr: string) => {
    navigator.clipboard.writeText(addr).catch(() => {});
    setCopiedAddress(addr);
    setTimeout(() => setCopiedAddress(null), 1500);
  };

  // Check if any key is expiring soon (for warning banner)
  const expiringKey = mgmt.serverKeys.find(k => getKeyStatus(k) === 'expiring');

  const displayKeys = showExpired
    ? mgmt.serverKeys
    : mgmt.serverKeys.filter(k => { const s = getKeyStatus(k); return s !== 'expired' && s !== 'revoked'; });

  return (
    <div>
      {/* Expiring Soon Banner */}
      {expiringKey && (
        <div style={{
          display: 'flex', alignItems: 'center', gap: 12,
          background: 'rgba(249,115,22,0.12)',
          border: '1px solid rgba(249,115,22,0.25)',
          borderRadius: 10, padding: '12px 16px', marginBottom: 16,
        }}>
          {/* key icon circle */}
          <div style={{
            width: 32, height: 32, borderRadius: '50%',
            background: 'rgba(249,115,22,0.18)',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
          }}>
            <Key size={15} color="#f97316" />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: '#f97316' }}>Session key expiring soon</div>
            <div className="text-text-muted" style={{ fontSize: 12, marginTop: 2 }}>
              <span className="mono">{expiringKey.session_key.slice(0, 6)}…{expiringKey.session_key.slice(-4)}</span>
              {' '}· {formatExpiry(expiringKey.expires_at, 'relative')} remaining
            </div>
          </div>
          <button
            className="btn btn-ghost btn-sm"
            style={{ borderColor: 'rgba(249,115,22,0.4)', color: '#f97316', flexShrink: 0 }}
            onClick={() => setKeyForUpdate(expiringKey)}
          >
            Renew Key
          </button>
        </div>
      )}

      {/* Main card */}
      <div className="card">
        {/* Card header */}
        <div className="card-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Key size={16} style={{ color: 'var(--text-muted)' }} />
            <span className="card-title">Session Keys</span>
            <span style={{
              background: 'var(--bg-elevated)', border: '1px solid var(--border)',
              borderRadius: 999, padding: '1px 8px', fontSize: 12,
              color: 'var(--text-muted)',
            }} className="mono">
              {mgmt.serverKeys.length}
            </span>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              className="btn btn-ghost btn-sm"
              onClick={mgmt.fetchKeys}
              disabled={mgmt.isLoading}
              title="Refresh"
            >
              <RefreshCw size={13} className={mgmt.isLoading ? 'animate-spin' : ''} />
              Refresh
            </button>
            <button
              className="btn btn-primary btn-sm"
              onClick={() => setShowRegisterModal(true)}
            >
              <Plus size={13} />
              Register New
            </button>
          </div>
        </div>

        {/* Table or empty state */}
        {mgmt.isLoading && mgmt.serverKeys.length === 0 ? (
          <div style={{ padding: '40px 0', textAlign: 'center' }}>
            <span className="spinner" />
          </div>
        ) : mgmt.serverKeys.length === 0 ? (
          <div style={{ padding: '48px 24px', textAlign: 'center' }}>
            <div style={{ fontSize: 14, marginBottom: 16, color: 'var(--text-muted)' }}>
              No session keys yet. Register one to sign state updates without MetaMask prompts.
            </div>
            <button className="btn btn-primary" onClick={() => setShowRegisterModal(true)}>
              <Plus size={14} />
              Register New Key
            </button>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  <th style={{ padding: '10px 16px', textAlign: 'left', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12 }}>Address</th>
                  <th style={{ padding: '10px 16px', textAlign: 'left', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12 }}>Assets</th>
                  <th
                    style={{ padding: '10px 16px', textAlign: 'left', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12, cursor: 'pointer', userSelect: 'none' }}
                    onClick={() => setExpFmt(f => NEXT_FMT[f])}
                    title={FMT_HINT[expFmt]}
                  >
                    Expiration ↕
                  </th>
                  <th style={{ padding: '10px 16px', textAlign: 'left', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                      <button
                        onClick={() => setShowExpired(v => !v)}
                        title={showExpired ? 'Hide expired' : 'Show expired'}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, display: 'flex', alignItems: 'center', color: 'var(--text-muted)' }}
                      >
                        {showExpired ? <Eye size={12} /> : <EyeOff size={12} />}
                      </button>
                      Status
                    </div>
                  </th>
                  <th style={{ padding: '10px 16px', textAlign: 'left', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12 }}>Version</th>
                  <th style={{ padding: '10px 16px', textAlign: 'right', color: 'var(--text-muted)', fontWeight: 500, fontSize: 12 }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {displayKeys.map(key => {
                  const status = getKeyStatus(key);
                  const isTerminal = status === 'expired' || status === 'revoked';
                  const hasLocalKey = allSessionKeys.some(
                    lk => lk.sessionKeyAddress.toLowerCase() === key.session_key.toLowerCase(),
                  );
                  const isCurrentKey = sessionKey?.sessionKeyAddress.toLowerCase() === key.session_key.toLowerCase();
                  const rowOpacity = status === 'revoked' ? 0.55 : status === 'expired' ? 0.7 : 1;

                  return (
                    <tr
                      key={`${key.session_key}-${key.version}`}
                      style={{ borderBottom: '1px solid var(--border)', opacity: rowOpacity }}
                    >
                      {/* Address */}
                      <td style={{ padding: '12px 16px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <Key
                            size={12}
                            color={isTerminal ? 'var(--text-muted)' : status === 'expiring' ? '#f97316' : '#22c55e'}
                          />
                          <span className="mono" style={{ fontSize: 12 }}>
                            {key.session_key.slice(0, 6)}…{key.session_key.slice(-4)}
                          </span>
                          {isCurrentKey && (
                            <span style={{
                              background: 'var(--accent-dim)', color: 'var(--accent)',
                              borderRadius: 4, padding: '1px 6px', fontSize: 10, fontWeight: 700,
                            }}>
                              IN USE
                            </span>
                          )}
                          <button
                            className="btn btn-ghost btn-sm"
                            style={{ padding: '2px 6px', opacity: 0.6 }}
                            onClick={() => copyAddr(key.session_key)}
                            title="Copy address"
                          >
                            {copiedAddress === key.session_key ? <Check size={11} /> : <Copy size={11} />}
                          </button>
                        </div>
                      </td>

                      {/* Assets */}
                      <td style={{ padding: '12px 16px' }}>
                        <span className="mono" style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                          {key.assets.join(', ')}
                        </span>
                      </td>

                      {/* Expiration */}
                      <td style={{ padding: '12px 16px' }}>
                        <span
                          className="ts-toggle"
                          onClick={() => setExpFmt(f => NEXT_FMT[f])}
                          data-tip={FMT_HINT[expFmt]}
                          style={{ color: status === 'expiring' ? '#f97316' : undefined }}
                        >
                          {formatExpiry(key.expires_at, expFmt)}
                        </span>
                      </td>

                      {/* Status */}
                      <td style={{ padding: '12px 16px' }}>
                        <StatusBadge status={status} />
                      </td>

                      {/* Version */}
                      <td style={{ padding: '12px 16px' }}>
                        <span className="mono" style={{ fontSize: 12, color: 'var(--text-muted)' }}>v{key.version}</span>
                      </td>

                      {/* Actions */}
                      <td style={{ padding: '12px 16px', textAlign: 'right' }}>
                        <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
                          {status === 'expiring' && (
                            <button
                              className="btn btn-ghost btn-sm"
                              style={{ borderColor: 'rgba(249,115,22,0.4)', color: '#f97316' }}
                              onClick={() => setKeyForUpdate(key)}
                            >
                              Renew
                            </button>
                          )}
                          {status === 'active' && (
                            <button
                              className="btn btn-ghost btn-sm"
                              onClick={() => setKeyForUpdate(key)}
                            >
                              Update
                            </button>
                          )}
                          {isTerminal && hasLocalKey && (
                            <button
                              className="btn btn-ghost btn-sm"
                              onClick={() => setKeyForUpdate(key)}
                              title="Re-register with future expiry to reactivate"
                            >
                              Reactivate
                            </button>
                          )}
                          {!isTerminal && (
                            <button
                              className="btn btn-danger btn-sm"
                              onClick={() => setKeyForRevoke(key)}
                            >
                              Revoke
                            </button>
                          )}
                          {!isTerminal && !isCurrentKey && (
                            <button
                              className="btn btn-ghost btn-sm"
                              onClick={() => onSelectKey(key.session_key as Address)}
                            >
                              Use
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Register / Update modal */}
      {(showRegisterModal || keyForUpdate) && (
        <SessionKeyRegisterForm
          mode={keyForUpdate ? 'update' : 'register'}
          initialAssets={keyForUpdate?.assets}
          initialExpiryUnix={keyForUpdate ? BigInt(keyForUpdate.expires_at) : undefined}
          supportedAssets={supportedAssets}
          isSubmitting={mgmt.isSubmitting}
          onSubmit={async (assets: string[], expiresAt: bigint) => {
            if (!address) return;
            if (keyForUpdate) {
              const sk = await mgmt.update(address, keyForUpdate, assets, expiresAt);
              if (sk) onKeyActivated(sk);
            } else {
              const sk = await mgmt.register(address, assets, expiresAt);
              if (sk) onKeyActivated(sk);
            }
            setShowRegisterModal(false);
            setKeyForUpdate(null);
          }}
          onCancel={() => { setShowRegisterModal(false); setKeyForUpdate(null); }}
        />
      )}

      {/* Revoke confirmation modal */}
      {keyForRevoke && (
        <SessionKeyRevokeModal
          sessionKeyAddress={keyForRevoke.session_key as Address}
          isSubmitting={mgmt.isSubmitting}
          onConfirm={async () => {
            if (!address) return;
            await mgmt.revoke(address, keyForRevoke);
            if (sessionKey?.sessionKeyAddress.toLowerCase() === keyForRevoke.session_key.toLowerCase()) {
              onKeyCleared();
            }
            onRefreshAllKeys();
            setKeyForRevoke(null);
          }}
          onCancel={() => setKeyForRevoke(null)}
        />
      )}
    </div>
  );
}
