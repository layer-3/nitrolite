import { useState } from 'react';
import type { Asset, Blockchain } from '@yellow-org/sdk';

interface Props {
  asset: string;
  assets: Asset[];
  chains: Blockchain[];
  onConfirm: (asset: string, chainId: bigint) => void;
  onCancel: () => void;
}

export default function SetHomechainModal({ asset, assets, chains, onConfirm, onCancel }: Props) {
  const tokens = assets.find(a => a.symbol === asset)?.tokens ?? [];
  const supportedChainIds = new Set(tokens.map(t => t.blockchainId.toString()));
  const eligible = chains.filter(c => supportedChainIds.has(c.id.toString()));
  const [selected, setSelected] = useState<bigint | null>(eligible[0]?.id ?? null);

  return (
    <div className="modal-overlay">
      <div className="modal-card p-6">
        <div className="flex items-center gap-2 mb-4">
          <h2 className="text-lg font-semibold">Select home blockchain</h2>
          <span
            className="text-xs font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full border border-accent/40 text-accent"
            style={{ background: 'var(--accent-dim)' }}
          >
            {asset.toUpperCase()}
          </span>
        </div>
        <p className="text-sm text-text-muted mb-4">
          Choose the chain where this asset will settle on-chain. You can only set this once per asset.
        </p>

        <div className="space-y-2 mb-5">
          {eligible.length === 0 ? (
            <p className="text-text-muted text-sm">No chains support this asset.</p>
          ) : (
            eligible.map(c => (
              <button
                key={c.id.toString()}
                className={`btn w-full justify-start ${selected === c.id ? 'border-accent' : ''}`}
                style={selected === c.id ? { borderColor: 'var(--accent)', background: 'var(--accent-dim)' } : undefined}
                onClick={() => setSelected(c.id)}
              >
                <span
                  className="inline-flex items-center justify-center w-4 h-4 rounded-full border"
                  style={{ borderColor: selected === c.id ? 'var(--accent)' : 'var(--border)' }}
                >
                  {selected === c.id && (
                    <span className="w-2 h-2 rounded-full" style={{ background: 'var(--accent)' }} />
                  )}
                </span>
                <span>{c.name}</span>
                <span className="mono text-text-muted text-xs ml-auto">Chain ID: {c.id.toString()}</span>
              </button>
            ))
          )}
        </div>

        <div className="flex gap-2 justify-end">
          <button className="btn btn-ghost" onClick={onCancel}>
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={() => selected != null && onConfirm(asset, selected)}
            disabled={selected == null}
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
}
