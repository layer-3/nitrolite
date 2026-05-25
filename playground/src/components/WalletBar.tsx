import { useEffect, useState } from 'react';
import { Key, X } from 'lucide-react';
import type { Address } from 'viem';
import type { Blockchain } from '@yellow-org/sdk';
import CopyButton from './CopyButton';
import { formatAddress, timeAgo } from '../utils';
import type { StoredSessionKey } from '../sessionKey';
import { secondsUntilExpiry } from '../sessionKey';

interface Props {
  address: Address | null;
  chainId: bigint | null;
  chains: Blockchain[];
  lastCommsAt: Date | null;
  nodeError: string | null;
  isConnecting: boolean;
  walletError: string | null;
  sessionKey: StoredSessionKey | null;
  onConnect: () => void;
  onDisconnect: () => void;
  onSwitchChain: (chainId: bigint) => void;
  onClearSessionKey: () => void;
}

export default function WalletBar({
  address,
  chainId,
  chains,
  lastCommsAt,
  nodeError,
  isConnecting,
  walletError,
  sessionKey,
  onConnect,
  onDisconnect,
  onSwitchChain,
  onClearSessionKey,
}: Props) {
  // Tick every second so "Xs ago" stays fresh
  const [, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 1000);
    return () => clearInterval(id);
  }, []);

  const currentChainName = chains.find(c => c.id === chainId)?.name;

  return (
    <nav className="sticky top-0 z-10 flex items-center justify-between px-6 h-14 bg-bg-surface border-b border-border">
      <div className="flex items-center gap-3">
        <span className="text-accent font-semibold tracking-tight">Nitrolite</span>
        <span className="text-text-muted text-xs uppercase tracking-wider">Playground</span>
      </div>

      <div className="flex items-center gap-3 text-xs">
        {nodeError ? (
          <span className="flex items-center gap-1.5 text-error">
            <span className="dot error" />
            <span className="truncate max-w-[200px]" title={nodeError}>{nodeError}</span>
          </span>
        ) : lastCommsAt ? (
          <span className="flex items-center gap-1.5 text-text-muted">
            <span className="dot" />
            <span className="mono">{timeAgo(lastCommsAt)}</span>
          </span>
        ) : address ? (
          <span className="flex items-center gap-1.5 text-text-muted">
            <span className="dot muted" />
            <span>connecting…</span>
          </span>
        ) : null}
      </div>

      <div className="flex items-center gap-2">
        {address && sessionKey && (
          <span
            className="chip text-xs"
            title={`Session key ${sessionKey.sessionKeyAddress} · expires at ${new Date(Number(sessionKey.expiresAt) * 1000).toLocaleString()}`}
            style={{ borderColor: 'rgba(245,166,35,0.4)', color: 'var(--accent)' }}
          >
            <Key size={11} />
            <span>SK · {formatSkExpiry(sessionKey)}</span>
            <button
              type="button"
              className="text-text-muted hover:text-error transition-colors"
              onClick={onClearSessionKey}
              title="Clear session key"
              aria-label="Clear session key"
            >
              <X size={11} />
            </button>
          </span>
        )}

        {address ? (
          <>
            <select
              value={chainId ? chainId.toString() : ''}
              onChange={e => onSwitchChain(BigInt(e.target.value))}
              className="chip mono text-xs cursor-pointer"
              title={currentChainName ?? `Chain ${chainId}`}
            >
              {!currentChainName && chainId && (
                <option value={chainId.toString()}>Chain {chainId.toString()}</option>
              )}
              {chains.map(c => (
                <option key={c.id.toString()} value={c.id.toString()}>
                  {c.name}
                </option>
              ))}
            </select>

            <span className="chip mono">
              {formatAddress(address)}
              <CopyButton value={address} size={11} />
            </span>

            <button className="btn btn-ghost btn-sm" onClick={onDisconnect}>
              Disconnect
            </button>
          </>
        ) : (
          <>
            {walletError && (
              <span className="text-error text-xs max-w-[260px] truncate" title={walletError}>
                {walletError}
              </span>
            )}
            <button className="btn btn-primary btn-sm" onClick={onConnect} disabled={isConnecting}>
              {isConnecting ? <span className="spinner" /> : 'Connect MetaMask'}
            </button>
          </>
        )}
      </div>
    </nav>
  );
}

function formatSkExpiry(sk: StoredSessionKey): string {
  const sec = secondsUntilExpiry(sk);
  if (sec <= 0) return 'expired';
  if (sec < 60) return `${sec}s`;
  if (sec < 3600) return `${Math.floor(sec / 60)}m`;
  return `${Math.floor(sec / 3600)}h`;
}
