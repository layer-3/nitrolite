import { useState } from 'react';
import { toast } from 'sonner';
import { showErrorToast } from "../toastError";
import type { Client, Channel } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import { Inbox } from 'lucide-react';
import { formatBalance } from '../utils';

interface Props {
  client: Client | null;
  channels: Channel[];
  balances: Record<string, Decimal>;
  onAfterAck: () => void;
}

export default function PendingReceipts({ client, channels, balances, onAfterAck }: Props) {
  const channelAssets = new Set(channels.map(c => c.asset));
  const pending = Object.entries(balances).filter(
    ([asset, bal]) => bal.gt(0) && !channelAssets.has(asset),
  );

  const [acking, setAcking] = useState<string | null>(null);

  if (pending.length === 0) return null;

  const acknowledge = async (asset: string) => {
    if (!client) return;
    setAcking(asset);
    try {
      await client.acknowledge(asset);
      toast.success(`Acknowledged ${asset.toUpperCase()} receipt`);
      onAfterAck();
    } catch (err) {
      const e = err as { code?: number; message?: string };
      if (e?.code === 4001) {
        toast('Cancelled');
      } else {
        showErrorToast(`Acknowledge failed: ${e?.message ?? String(err)}`);
      }
    } finally {
      setAcking(null);
    }
  };

  return (
    <div className="card mb-4">
      <div className="card-header">
        <div className="flex items-center gap-2">
          <Inbox size={15} className="text-accent" />
          <span className="card-title">Pending receipts</span>
          <span className="text-text-muted text-sm mono">({pending.length})</span>
        </div>
      </div>
      <div className="p-4 space-y-2">
        <p className="text-text-muted text-xs mb-2">
          You received these assets but haven't co-signed the state yet. Acknowledge to claim the balance and open
          a channel.
        </p>
        {pending.map(([asset, bal]) => (
          <div key={asset} className="flex items-center justify-between gap-3 px-3 py-2.5 rounded border border-border bg-bg-elevated/40">
            <div className="flex items-center gap-3">
              <span
                className="inline-flex items-center justify-center w-9 h-9 rounded-full text-xs font-semibold"
                style={{ background: 'var(--accent-dim)', color: 'var(--accent)' }}
              >
                {asset.slice(0, 4).toUpperCase()}
              </span>
              <div>
                <div className="text-sm font-medium">{asset.toUpperCase()}</div>
                <div className="mono text-xs text-text-muted">{formatBalance(bal)} incoming</div>
              </div>
            </div>
            <button
              className="btn btn-primary btn-sm"
              onClick={() => acknowledge(asset)}
              disabled={acking !== null}
            >
              {acking === asset ? (
                <>
                  <span className="spinner" />
                  Acknowledging…
                </>
              ) : (
                'Acknowledge'
              )}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
