import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Search, Wallet, Check, Key } from 'lucide-react';
import type { Client, Channel, Blockchain } from '@yellow-org/sdk';
import { ChannelStatus } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import CopyButton from './CopyButton';
import StateViewer from './StateViewer';
import { formatAddress } from '../utils';
import { chainDisplayName } from '../chainMeta';
import type { StoredSessionKey } from '../sessionKey';
import { secondsUntilExpiry } from '../sessionKey';

interface Props {
  channel: Channel;
  client: Client | null;
  address: Address | null;
  chains: Blockchain[];
  currentChainId: bigint | null;
  enforcedBalance: Decimal | null | undefined;
  onClose: (asset: string, blockchainId: bigint) => void;
  onSwitchToHomeChain: (chainId: bigint) => void;
  onSelectAsset: (asset: string) => void;
  onAfterOp?: () => void;
  isClosing: boolean;
  channelStatesKey?: number;
  defaultExpanded?: boolean;
  sessionKey: StoredSessionKey | null;
  allSessionKeys: StoredSessionKey[];
  onSelectSessionKey: (sessionKeyAddress: Address | null) => void;
  onManageSessionKeys: () => void;
}

function formatSkExpiry(sk: StoredSessionKey): string {
  const secs = secondsUntilExpiry(sk);
  if (secs <= 0) return 'expired';
  if (secs < 60) return `${secs}s`;
  if (secs < 3600) return `${Math.floor(secs / 60)}m`;
  return `${Math.floor(secs / 3600)}h`;
}

// Dropdown item component to handle hover via React state cleanly
function DropdownItem({
  onClick,
  selected,
  disabled,
  children,
}: {
  onClick?: () => void;
  selected?: boolean;
  disabled?: boolean;
  children: React.ReactNode;
}) {
  const [hovered, setHovered] = useState(false);
  const bg = selected
    ? 'var(--accent-dim)'
    : hovered && !disabled
      ? 'rgba(255,255,255,0.05)'
      : 'transparent';

  if (disabled) {
    return (
      <div style={{
        display: 'flex', alignItems: 'center', gap: 10,
        padding: '8px 14px', fontSize: 13, opacity: 0.4, cursor: 'not-allowed',
        background: 'transparent', color: 'var(--text-primary)',
      }}>
        {children}
      </div>
    );
  }

  return (
    <button
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        display: 'flex', alignItems: 'center', gap: 10, width: '100%',
        padding: '8px 14px', fontSize: 13, border: 'none', cursor: 'pointer',
        background: bg, color: 'var(--accent)', textAlign: 'left', fontFamily: 'inherit',
        transition: 'background 0.1s',
      }}
    >
      {children}
    </button>
  );
}

export default function ChannelRow({
  channel,
  client,
  address,
  chains,
  currentChainId,
  enforcedBalance,
  onClose,
  onSwitchToHomeChain,
  onSelectAsset,
  onAfterOp,
  isClosing,
  channelStatesKey,
  defaultExpanded,
  sessionKey,
  allSessionKeys,
  onSelectSessionKey,
  onManageSessionKeys,
}: Props) {
  const [expanded, setExpanded] = useState(defaultExpanded ?? false);
  const [skOpen, setSkOpen] = useState(false);
  const skRef = useRef<HTMLDivElement>(null);
  const closed = channel.status === ChannelStatus.Closed;

  useEffect(() => {
    if (!skOpen) return;
    const handler = (e: MouseEvent) => {
      if (skRef.current && !skRef.current.contains(e.target as Node)) setSkOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [skOpen]);

  const validKeys = allSessionKeys.filter(k => secondsUntilExpiry(k) > 0);
  const expiredKeys = allSessionKeys.filter(k => secondsUntilExpiry(k) <= 0);
  const activeSkSecs = sessionKey ? secondsUntilExpiry(sessionKey) : -1;
  const homeChainObj = chains.find(c => c.id === channel.blockchainId);
  const homeChainName = chainDisplayName(channel.blockchainId, homeChainObj?.name);
  const wrongChain = !closed && currentChainId != null && currentChainId !== channel.blockchainId;

  // Signer selector — shown in header for non-closed channels when keys exist
  const showSkSelector = !closed && allSessionKeys.length > 0;

  const skIcon = sessionKey && activeSkSecs > 0
    ? <Key size={11} style={{ color: 'var(--accent)', flexShrink: 0 }} />
    : <span style={{ width: 6, height: 6, borderRadius: '50%', flexShrink: 0, background: 'var(--text-muted)' }} />;

  const skLabel = sessionKey && activeSkSecs > 0
    ? `${sessionKey.sessionKeyAddress.slice(0, 5)}…${sessionKey.sessionKeyAddress.slice(-3)}`
    : 'Wallet';

  return (
    <div className={`border border-border rounded-lg mb-2 ${closed ? 'opacity-60' : ''}`}>
      {/* ── Row header ── */}
      <div
        className={`flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bg-elevated/40 rounded-t-lg ${!expanded ? 'rounded-b-lg' : ''}`}
        onClick={() => setExpanded(e => !e)}
      >
        {/* Asset icon */}
        <span
          className="inline-flex items-center justify-center w-9 h-9 rounded-full text-xs font-semibold flex-shrink-0"
          style={{
            background: closed ? 'rgba(102,102,102,0.12)' : 'var(--accent-dim)',
            color: closed ? 'var(--text-muted)' : 'var(--accent)',
          }}
        >
          {channel.asset.slice(0, 4).toUpperCase()}
        </span>

        {/* Name + ID */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 text-sm font-medium flex-wrap">
            <span>{channel.asset.toUpperCase()}</span>
            <span className="tooltip-wrap">
              <button
                type="button"
                className="flex items-center text-text-muted hover:text-accent transition-colors"
                onClick={e => { e.stopPropagation(); onSelectAsset(channel.asset); }}
                aria-label="Select asset"
              >
                <Search size={12} />
              </button>
              <span className="tip">Select asset</span>
            </span>
            {/* Chain pill */}
            <span className="chip text-[10px] font-normal" style={{ padding: '2px 7px', borderRadius: 6 }}>
              {homeChainName}
            </span>
            {/* Signer selector — inline after chain pill, non-closed only */}
            {showSkSelector && (
              <div
                ref={skRef}
                style={{ position: 'relative' }}
                onClick={e => e.stopPropagation()}
              >
                <button
                  className="chip text-xs"
                  onClick={() => setSkOpen(o => !o)}
                  style={{ cursor: 'pointer', gap: 5 }}
                  title="Select signing key"
                >
                  {skIcon}
                  <span className="mono">{skLabel}</span>
                  <ChevronDown size={10} style={{ transition: 'transform 0.15s', transform: skOpen ? 'rotate(180deg)' : undefined }} />
                </button>

                {skOpen && (
                  <div style={{
                    position: 'absolute', top: 'calc(100% + 6px)', left: 0, zIndex: 50,
                    background: 'var(--bg-elevated)', border: '1px solid var(--border)',
                    borderRadius: 10, boxShadow: '0 8px 24px rgba(0,0,0,0.55)',
                    minWidth: 250, padding: '6px 0',
                  }}>
                    <DropdownItem
                      selected={!sessionKey || activeSkSecs <= 0}
                      onClick={() => { onSelectSessionKey(null); setSkOpen(false); }}
                    >
                      <Wallet size={13} style={{ color: 'var(--text-muted)', flexShrink: 0 }} />
                      <span style={{ flex: 1, color: 'var(--text-primary)' }}>Wallet (no key)</span>
                      {(!sessionKey || activeSkSecs <= 0) && <Check size={13} style={{ color: 'var(--accent)', flexShrink: 0 }} />}
                      <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>MetaMask</span>
                    </DropdownItem>

                    {validKeys.length > 0 && (
                      <>
                        <div style={{ height: 1, background: 'var(--border)', margin: '4px 0' }} />
                        {validKeys.map(sk => {
                          const secs = secondsUntilExpiry(sk);
                          const expiring = secs < 3600;
                          const selected = sessionKey?.sessionKeyAddress.toLowerCase() === sk.sessionKeyAddress.toLowerCase();
                          return (
                            <DropdownItem
                              key={sk.sessionKeyAddress}
                              selected={selected}
                              onClick={() => { onSelectSessionKey(sk.sessionKeyAddress); setSkOpen(false); }}
                            >
                              <Key size={12} style={{ flexShrink: 0, color: 'var(--accent)' }} />
                              <span className="mono" style={{ flex: 1, fontSize: 12, color: 'var(--text-primary)' }}>
                                {sk.sessionKeyAddress.slice(0, 5)}…{sk.sessionKeyAddress.slice(-3)}
                              </span>
                              {selected && <Check size={13} style={{ color: 'var(--accent)', flexShrink: 0 }} />}
                              <span style={{ fontSize: 11, color: expiring ? '#f97316' : 'var(--text-muted)' }}>
                                {formatSkExpiry(sk)}
                              </span>
                            </DropdownItem>
                          );
                        })}
                      </>
                    )}

                    {expiredKeys.length > 0 && (
                      <>
                        <div style={{ height: 1, background: 'var(--border)', margin: '4px 0' }} />
                        <div style={{ padding: '6px 14px', fontSize: 11, color: 'var(--text-muted)', fontStyle: 'italic' }}>
                          And {expiredKeys.length} more expired session {expiredKeys.length === 1 ? 'key' : 'keys'}
                        </div>
                      </>
                    )}

                    <div style={{ height: 1, background: 'var(--border)', margin: '4px 0' }} />
                    <DropdownItem onClick={() => { onManageSessionKeys(); setSkOpen(false); }}>
                      Manage session keys →
                    </DropdownItem>
                  </div>
                )}
              </div>
            )}
            {wrongChain && (
              <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border border-accent/40 text-accent">
                wrong chain
              </span>
            )}
            {closed && (
              <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded border border-border text-text-muted">
                closed
              </span>
            )}
          </div>
          <div className="mono text-xs text-text-muted flex items-center gap-1 mt-0.5">
            {formatAddress(channel.channelId)}
            <CopyButton value={channel.channelId} size={11} />
          </div>
        </div>

        <ChevronDown
          size={16}
          className={`text-text-muted transition-transform flex-shrink-0 ${expanded ? 'rotate-180' : ''}`}
        />
      </div>

      {/* ── Expanded panel ── */}
      {expanded && (
        <div className="border-t border-border px-4 py-3 space-y-3 bg-bg-base/30 rounded-b-lg">
          {wrongChain && (
            <div className="flex items-center justify-between gap-2 px-3 py-2 rounded border border-accent/30 bg-accent-dim">
              <span className="text-accent text-xs">
                Wallet is on a different chain. Switch to "{homeChainName}" to interact.
              </span>
              <button
                className="btn btn-sm"
                style={{ borderColor: 'var(--accent)', color: 'var(--accent)' }}
                onClick={() => onSwitchToHomeChain(channel.blockchainId)}
              >
                Switch
              </button>
            </div>
          )}

          {!closed && (
            <div className="flex gap-2">
              <button
                className="btn btn-danger btn-sm"
                onClick={() => onClose(channel.asset, channel.blockchainId)}
                disabled={isClosing || wrongChain}
              >
                {isClosing ? <span className="spinner" /> : 'Close channel'}
              </button>
            </div>
          )}

          {closed ? (
            <p className="text-text-muted text-xs">Channel closed. Final balance settled on-chain.</p>
          ) : (
            <StateViewer
              client={client}
              address={address}
              asset={channel.asset}
              enforcedBalance={enforcedBalance}
              onAfterOp={onAfterOp}
              isLocked={isClosing}
              refreshKey={channelStatesKey}
            />
          )}
        </div>
      )}
    </div>
  );
}
