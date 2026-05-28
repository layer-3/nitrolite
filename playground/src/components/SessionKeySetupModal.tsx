import { Key } from 'lucide-react';

interface Props {
  assets: string[];
  isRegistering: boolean;
  mode: 'setup' | 'renew';
  onConfirm: () => void;
  onCancel: () => void;
}

export default function SessionKeySetupModal({ assets, isRegistering, mode, onConfirm, onCancel }: Props) {
  return (
    <div className="modal-overlay">
      <div className="modal-card p-6">
        <div className="flex items-center justify-center w-12 h-12 rounded-full bg-accent-dim mb-4 mx-auto">
          <Key size={20} className="text-accent" />
        </div>
        <h2 className="text-lg font-semibold text-center mb-2">
          {mode === 'renew' ? 'Renew session key' : 'Set up a session key'}
        </h2>
        <p className="text-sm text-text-muted text-center mb-5">
          A session key is a temporary key that signs state updates on your behalf, so you stop seeing a MetaMask
          popup for every deposit, withdraw, or transfer.
        </p>

        <ul className="text-xs text-text-muted space-y-1.5 mb-5 list-disc pl-5">
          <li>
            <span className="text-text-primary">Expires in 24 hours.</span> You'll be asked to renew.
          </li>
          <li>
            Authorizes all currently supported assets
            {assets.length > 0 && (
              <span className="mono text-text-primary"> ({assets.join(', ').toUpperCase()})</span>
            )}
            .
          </li>
          <li>
            <span className="text-text-primary">On-chain transactions still require MetaMask</span> (deposit funds,
            checkpoint, approve token).
          </li>
          <li>Stored in this browser's localStorage. Clear anytime from the wallet bar.</li>
        </ul>

        <p className="text-xs text-text-muted text-center mb-5">
          Confirm to sign one MetaMask request and authorize the session key.
        </p>

        <div className="flex gap-2 justify-end">
          <button className="btn btn-ghost" onClick={onCancel} disabled={isRegistering}>
            Cancel
          </button>
          <button className="btn btn-primary" onClick={onConfirm} disabled={isRegistering}>
            {isRegistering ? (
              <>
                <span className="spinner" /> Signing…
              </>
            ) : mode === 'renew' ? (
              'Renew & sign'
            ) : (
              'Set up & sign'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
