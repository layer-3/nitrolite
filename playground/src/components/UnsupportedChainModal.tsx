import { AlertTriangle } from 'lucide-react';
import type { Blockchain } from '@yellow-org/sdk';

interface Props {
  chains: Blockchain[];
  onSwitchChain: (chainId: bigint) => void;
}

export default function UnsupportedChainModal({ chains, onSwitchChain }: Props) {
  return (
    <div className="modal-overlay">
      <div className="modal-card p-6">
        <div className="flex items-center justify-center w-12 h-12 rounded-full bg-accent-dim mb-4">
          <AlertTriangle size={20} className="text-accent" />
        </div>
        <h2 className="text-lg font-semibold mb-2">Unsupported network</h2>
        <p className="text-sm text-text-muted mb-5">
          Your wallet is on a chain Nitrolite doesn't support yet. Switch to one of the supported networks below to
          continue.
        </p>
        <div className="space-y-2">
          {chains.length === 0 ? (
            <p className="text-text-muted text-sm">No supported chains discovered yet.</p>
          ) : (
            chains.map(c => (
              <button
                key={c.id.toString()}
                className="btn w-full justify-start"
                onClick={() => onSwitchChain(c.id)}
              >
                <span className="dot" />
                <span>Switch to {c.name}</span>
                <span className="mono text-text-muted text-xs ml-auto">Chain ID: {c.id.toString()}</span>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
