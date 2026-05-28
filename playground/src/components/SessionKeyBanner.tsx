import { Key } from 'lucide-react';

interface Props {
  onSetup: () => void;
}

export default function SessionKeyBanner({ onSetup }: Props) {
  return (
    <div
      className="card mb-4 flex items-center justify-between gap-4 px-4 py-3"
      style={{ borderColor: 'rgba(245,166,35,0.3)', background: 'var(--accent-dim)' }}
    >
      <div className="flex items-center gap-3">
        <div
          className="flex items-center justify-center w-9 h-9 rounded-full"
          style={{ background: 'rgba(245,166,35,0.2)' }}
        >
          <Key size={16} className="text-accent" />
        </div>
        <div>
          <div className="text-sm font-medium text-text-primary">Skip the MetaMask popups</div>
          <div className="text-xs text-text-muted">
            Authorize a 24h session key to sign state updates without prompts.
          </div>
        </div>
      </div>
      <button className="btn btn-primary btn-sm whitespace-nowrap" onClick={onSetup}>
        Set up
      </button>
    </div>
  );
}
