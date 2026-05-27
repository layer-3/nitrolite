import { useState } from 'react';
import { ChevronDown, Search } from 'lucide-react';
import type { Client, Channel, Blockchain } from '@yellow-org/sdk';
import { ChannelStatus } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import CopyButton from './CopyButton';
import StateViewer from './StateViewer';
import { formatAddress } from '../utils';
import { chainDisplayName } from '../chainMeta';

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
}: Props) {
  const [expanded, setExpanded] = useState(defaultExpanded ?? false);
  const closed = channel.status === ChannelStatus.Closed;

  const homeChainObj = chains.find(c => c.id === channel.blockchainId);
  const homeChainName = chainDisplayName(channel.blockchainId, homeChainObj?.name);
  const wrongChain = !closed && currentChainId != null && currentChainId !== channel.blockchainId;

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
