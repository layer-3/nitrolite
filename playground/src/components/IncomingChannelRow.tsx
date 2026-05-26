import { useState, useEffect } from 'react';
import { ChevronDown, Search } from 'lucide-react';
import { toast } from 'sonner';
import type { Client, State } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import { showErrorToast } from '../toastError';
import { formatBalance } from '../utils';

interface Props {
  asset: string;
  balance: Decimal;
  client: Client | null;
  address: Address | null;
  currentChainId: bigint | null;
  onSelectAsset: (asset: string) => void;
  onAfterAck: () => void;
  onExpandChange?: (asset: string, expanded: boolean) => void;
}

export default function IncomingChannelRow({ asset, balance, client, address, currentChainId, onSelectAsset, onAfterAck, onExpandChange }: Props) {
  const [expanded, setExpanded] = useState(false);

  const toggleExpanded = () => {
    const next = !expanded;
    setExpanded(next);
    onExpandChange?.(asset, next);
  };
  const [issuedState, setIssuedState] = useState<State | null>(null);
  const [isAcknowledging, setIsAcknowledging] = useState(false);

  useEffect(() => {
    if (!client || !address) return;
    client.getLatestState(address, asset, false)
      .then(s => setIssuedState(s))
      .catch(() => {});
  }, [client, address, asset]);

  const handleAcknowledge = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!client || !currentChainId) return;
    setIsAcknowledging(true);
    try {
      await client.setHomeBlockchain(asset, currentChainId);
      await client.acknowledge(asset);
      toast.success(`Acknowledged ${asset.toUpperCase()} receipt`);
      onAfterAck();
    } catch (err) {
      const error = err as { code?: number; message?: string };
      if (error?.code === 4001) {
        toast('Cancelled');
      } else {
        showErrorToast(`Acknowledge failed: ${error?.message ?? String(err)}`);
      }
    } finally {
      setIsAcknowledging(false);
    }
  };

  const displayAmount = (issuedState as { homeLedger?: { userBalance?: Decimal } } | null)?.homeLedger?.userBalance ?? balance;

  return (
    <div className="border border-border rounded-lg mb-2">
      <div
        className={`flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bg-elevated/40 rounded-t-lg ${!expanded ? 'rounded-b-lg' : ''}`}
        onClick={toggleExpanded}
      >
        <span
          className="inline-flex items-center justify-center w-9 h-9 rounded-full text-xs font-semibold"
          style={{ background: 'var(--accent-dim)', color: 'var(--accent)' }}
        >
          {asset.slice(0, 4).toUpperCase()}
        </span>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 text-sm font-medium">
            <span>{asset.toUpperCase()}</span>
            <span className="tooltip-wrap">
              <button
                type="button"
                className="flex items-center text-text-muted hover:text-accent transition-colors"
                onClick={e => { e.stopPropagation(); onSelectAsset(asset); }}
                aria-label="Select asset"
              >
                <Search size={12} />
              </button>
              <span className="tip">Select asset</span>
            </span>
            <span className="tooltip-wrap">
              <span className="text-[10px] px-1 rounded-xl border border-rose-800/50 bg-rose-950/40 text-rose-400 font-normal cursor-default select-none">
                NO HOME CHAIN
              </span>
              <span className="tip">Chain the acknowledge is invoked on will become the Home chain</span>
            </span>
          </div>
          <div className="mono text-xs text-text-muted mt-0.5">
            {formatBalance(displayAmount)} incoming
          </div>
        </div>

        <ChevronDown
          size={16}
          className={`text-text-muted transition-transform ${expanded ? 'rotate-180' : ''}`}
        />
      </div>

      {expanded && (
        <div className="border-t border-border px-4 py-3 bg-bg-base/30 rounded-b-lg">
          <div className="border border-border rounded-lg">
            <div className="grid grid-cols-[120px_70px_1fr_auto] gap-3 px-4 py-2 text-text-muted text-[10px] uppercase tracking-wider border-b border-border bg-bg-base/40 rounded-t-lg">
              <span>State</span>
              <span>Version</span>
              <span>Amount</span>
              <span></span>
            </div>
            <div className="grid grid-cols-[120px_70px_1fr_auto] gap-3 items-center px-4 py-3 rounded-b-lg hover:bg-bg-elevated/50 transition-colors">
              <div className="flex items-center gap-2 text-xs font-semibold" style={{ color: '#60a5fa' }}>
                <span className="inline-block w-2 h-2 rounded-full flex-shrink-0" style={{ background: '#60a5fa' }} />
                Issued
                <span className="tooltip-wrap">
                  <span className="help-icon">?</span>
                  <span className="tip">Node proposed this state. Acknowledge to co-sign and open the channel.</span>
                </span>
              </div>
              <span className="mono text-xs text-text-muted">
                {issuedState ? `v${issuedState.version.toString()}` : '—'}
              </span>
              <span className="mono text-sm text-text-primary">
                {formatBalance(displayAmount)} {asset.toUpperCase()}
              </span>
              <span>
                <button
                  className="btn btn-sm"
                  style={{ borderColor: 'rgba(96,165,250,0.4)', color: '#60a5fa' }}
                  onClick={handleAcknowledge}
                  disabled={isAcknowledging || !currentChainId}
                >
                  {isAcknowledging ? <span className="spinner" /> : 'Acknowledge'}
                </button>
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
