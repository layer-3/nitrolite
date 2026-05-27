import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { ChevronDown, Droplets } from 'lucide-react';
import type { Asset, Blockchain, Channel } from '@yellow-org/sdk';
import { ChannelType, ChannelStatus } from '@yellow-org/sdk';
import { tokenIconUrl, chainIconUrl } from '../icons';
import { chainDisplayName } from '../chainMeta';
import { FAUCET_ASSETS } from '../utils';

interface Props {
  assets: Asset[];
  selectedAsset: string;
  onSelectAsset: (symbol: string) => void;
  channels: Channel[];
  chains: Blockchain[];
  disabled?: boolean;
}

export default function TokenSelector({
  assets,
  selectedAsset,
  onSelectAsset,
  channels,
  chains,
  disabled,
}: Props) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const selectedObj = assets.find(a => a.symbol === selectedAsset);

  return (
    <div ref={ref} className="relative mb-3">
      <button
        type="button"
        className="w-full flex items-center gap-2.5 px-3 py-2.5 rounded-lg border border-border bg-bg-elevated hover:border-[#333] transition-colors text-sm disabled:opacity-40 disabled:cursor-not-allowed"
        onClick={() => { if (!disabled && assets.length) setOpen(o => !o); }}
        disabled={disabled || !assets.length}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        {selectedObj ? (
          <AssetRow asset={selectedObj} channels={channels} chains={chains} />
        ) : (
          <span className="text-text-muted flex-1 text-left">— no assets —</span>
        )}
        <ChevronDown
          size={14}
          className={`flex-shrink-0 text-text-muted transition-transform duration-150 ${open ? 'rotate-180' : ''}`}
        />
      </button>

      {open && (
        <div
          role="listbox"
          className="absolute top-full left-0 right-0 mt-1 rounded-lg border border-border bg-bg-elevated shadow-2xl z-20 overflow-hidden"
        >
          {assets.map(asset => (
            <button
              key={asset.symbol}
              type="button"
              role="option"
              aria-selected={asset.symbol === selectedAsset}
              className={`w-full flex items-center gap-2.5 px-3 py-2.5 text-sm transition-colors ${
                asset.symbol === selectedAsset
                  ? 'bg-bg-surface'
                  : 'hover:bg-bg-surface'
              }`}
              onClick={() => { onSelectAsset(asset.symbol); setOpen(false); }}
            >
              <AssetRow asset={asset} channels={channels} chains={chains} />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function AssetRow({
  asset,
  channels,
  chains,
}: {
  asset: Asset;
  channels: Channel[];
  chains: Blockchain[];
}) {
  const homeChannel = channels.find(
    c =>
      c.asset.toLowerCase() === asset.symbol.toLowerCase() &&
      c.type === ChannelType.Home &&
      c.status !== ChannelStatus.Closed,
  );

  const assetChains = asset.tokens
    .map(t => chains.find(c => c.id === t.blockchainId))
    .filter((c): c is Blockchain => c !== undefined)
    .sort((a, b) => (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));

  return (
    <>
      <TokenIcon symbol={asset.symbol} size={22} />
      <span className="font-medium text-text-primary flex-1 text-left flex items-center gap-1.5">
        {asset.symbol.toUpperCase()}
        {FAUCET_ASSETS.has(asset.symbol.toLowerCase()) && (
          <span className="tooltip-wrap">
            <Droplets size={13} className="text-accent flex-shrink-0" />
            <span className="tip">Faucet available</span>
          </span>
        )}
      </span>
      <div className="flex items-center gap-1.5">
        {assetChains.map(chain => (
          <ChainIcon
            key={chain.id.toString()}
            chain={chain}
            dimmed={homeChannel !== undefined && chain.id !== homeChannel.blockchainId}
          />
        ))}
      </div>
    </>
  );
}

function TokenIcon({ symbol, size }: { symbol: string; size: number }) {
  const [errored, setErrored] = useState(false);
  const url = tokenIconUrl(symbol);

  if (!url || errored) {
    return (
      <span
        className="rounded-full flex items-center justify-center bg-bg-surface border border-border text-text-muted font-bold flex-shrink-0 select-none"
        style={{ width: size, height: size, fontSize: Math.round(size * 0.45) }}
      >
        {symbol[0]?.toUpperCase()}
      </span>
    );
  }

  return (
    <img
      src={url}
      alt={symbol}
      width={size}
      height={size}
      className="rounded-full flex-shrink-0 object-cover"
      onError={() => setErrored(true)}
    />
  );
}

const CHAIN_ICON_SIZE = 20;

function ChainIcon({ chain, dimmed }: { chain: Blockchain; dimmed: boolean }) {
  const [errored, setErrored] = useState(false);
  const [tooltip, setTooltip] = useState<{ x: number; y: number } | null>(null);
  const url = chainIconUrl(chain.id);
  const dimStyle = dimmed ? { opacity: 0.2, filter: 'grayscale(1)' } : undefined;
  const displayName = chainDisplayName(chain.id, chain.name);

  const handleMouseEnter = (e: React.MouseEvent) => {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setTooltip({ x: rect.left + rect.width / 2, y: rect.top });
  };

  const icon = !url || errored ? (
    <span
      className="rounded-full bg-bg-surface border border-[#3a3a3a] flex items-center justify-center text-[9px] text-text-muted flex-shrink-0 select-none"
      style={{ width: CHAIN_ICON_SIZE, height: CHAIN_ICON_SIZE, ...dimStyle }}
    >
      {displayName[0]?.toUpperCase()}
    </span>
  ) : (
    <img
      src={url}
      alt={displayName}
      width={CHAIN_ICON_SIZE}
      height={CHAIN_ICON_SIZE}
      className="rounded-full flex-shrink-0 object-cover ring-1 ring-[#3a3a3a] block"
      style={dimStyle}
      onError={() => setErrored(true)}
    />
  );

  return (
    <span
      className="inline-flex flex-shrink-0"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={() => setTooltip(null)}
    >
      {icon}
      {tooltip && createPortal(
        <div
          style={{
            position: 'fixed',
            left: tooltip.x,
            top: tooltip.y - 8,
            transform: 'translate(-50%, -100%)',
            background: 'var(--bg-elevated)',
            border: '1px solid var(--border)',
            borderRadius: 6,
            color: 'var(--text-primary)',
            fontSize: 11,
            padding: '6px 10px',
            whiteSpace: 'nowrap',
            zIndex: 9999,
            pointerEvents: 'none',
            fontFamily: 'Inter, sans-serif',
          }}
        >
          {displayName}
        </div>,
        document.body,
      )}
    </span>
  );
}
