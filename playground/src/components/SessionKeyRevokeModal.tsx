import type { Address } from 'viem';
import { AlertTriangle } from 'lucide-react';

interface SessionKeyRevokeModalProps {
  sessionKeyAddress: Address;
  isSubmitting: boolean;
  onConfirm: () => Promise<void>;
  onCancel: () => void;
}

export default function SessionKeyRevokeModal({
  sessionKeyAddress,
  isSubmitting,
  onConfirm,
  onCancel,
}: SessionKeyRevokeModalProps) {
  return (
    <div className="modal-overlay">
      <div
        className="modal-card"
        style={{ width: 400, maxWidth: 'calc(100vw - 32px)', padding: 28 }}
      >
        {/* Header */}
        <div style={{ textAlign: 'center', marginBottom: 20 }}>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: '50%',
              background: 'rgba(239,68,68,0.12)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 14px',
            }}
          >
            <AlertTriangle size={20} color="var(--error)" />
          </div>
          <h2 style={{ fontSize: 17, fontWeight: 600, margin: '0 0 8px' }}>Revoke session key</h2>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', margin: 0, lineHeight: 1.5 }}>
            This key will no longer authorize any operations. This action requires a MetaMask
            signature.
          </p>
        </div>

        {/* Key address box */}
        <div
          style={{
            background: 'var(--bg-elevated)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            padding: '10px 14px',
            marginBottom: 20,
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>Key:</span>
          <span className="mono" style={{ fontSize: 12 }}>
            {sessionKeyAddress.slice(0, 10)}…{sessionKeyAddress.slice(-8)}
          </span>
        </div>

        {/* Footer */}
        <div style={{ display: 'flex', gap: 10 }}>
          <button
            className="btn btn-ghost"
            style={{ flex: 1 }}
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </button>
          <button
            className="btn btn-danger"
            style={{ flex: 1 }}
            onClick={onConfirm}
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <>
                <span className="spinner" /> Revoking…
              </>
            ) : (
              'Revoke & Sign'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
