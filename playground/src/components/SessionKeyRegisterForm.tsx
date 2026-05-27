import { useState, useEffect, useMemo } from 'react';
import { Key } from 'lucide-react';

interface SessionKeyRegisterFormProps {
  mode: 'register' | 'update';
  initialAssets?: string[];
  initialExpiryUnix?: bigint;
  supportedAssets: string[];
  isSubmitting: boolean;
  onSubmit: (assets: string[], expiresAt: bigint) => Promise<void>;
  onCancel: () => void;
}

type ExpiryMode = 'duration' | 'date' | 'unix';

export default function SessionKeyRegisterForm({
  mode,
  initialAssets,
  initialExpiryUnix,
  supportedAssets,
  isSubmitting,
  onSubmit,
  onCancel,
}: SessionKeyRegisterFormProps) {
  const [selectedAssets, setSelectedAssets] = useState<string[]>(
    initialAssets ?? [...supportedAssets]
  );

  const [expiryMode, setExpiryMode] = useState<ExpiryMode>('duration');

  const [durationValue, setDurationValue] = useState('24');
  const [durationUnit, setDurationUnit] = useState<'minutes' | 'hours' | 'days'>('hours');

  const [dateValue, setDateValue] = useState('');

  const [unixValue, setUnixValue] = useState('');

  useEffect(() => {
    if (initialExpiryUnix) {
      const expDate = new Date(Number(initialExpiryUnix) * 1000);
      setDateValue(expDate.toISOString().slice(0, 16));
      setUnixValue(initialExpiryUnix.toString());
      const secsFromNow = Number(initialExpiryUnix) - Math.floor(Date.now() / 1000);
      if (secsFromNow > 0) {
        const hoursFromNow = Math.ceil(secsFromNow / 3600);
        setDurationValue(String(hoursFromNow));
      }
    }
  }, [initialExpiryUnix]);

  const computedExpiresAt: bigint | null = useMemo(() => {
    const now = Math.floor(Date.now() / 1000);
    try {
      if (expiryMode === 'duration') {
        const val = parseInt(durationValue, 10);
        if (!val || val <= 0) return null;
        const seconds = durationUnit === 'minutes' ? val * 60 : durationUnit === 'hours' ? val * 3600 : val * 86400;
        return BigInt(now + seconds);
      }
      if (expiryMode === 'date') {
        if (!dateValue) return null;
        const ts = Math.floor(new Date(dateValue).getTime() / 1000);
        if (ts <= now + 60) return null;
        return BigInt(ts);
      }
      if (expiryMode === 'unix') {
        const ts = parseInt(unixValue, 10);
        if (!ts || ts <= now + 60) return null;
        return BigInt(ts);
      }
    } catch {
      return null;
    }
    return null;
  }, [expiryMode, durationValue, durationUnit, dateValue, unixValue]);

  const computedExpiryDisplay = computedExpiresAt
    ? new Date(Number(computedExpiresAt) * 1000).toISOString().replace('T', ' ').slice(0, 19) + ' UTC'
    : '—';

  const isValid = selectedAssets.length > 0 && computedExpiresAt !== null;

  const handleSubmit = async () => {
    if (!isValid || !computedExpiresAt) return;
    await onSubmit(selectedAssets, computedExpiresAt);
  };

  const toggleAsset = (asset: string) => {
    setSelectedAssets(prev =>
      prev.includes(asset) ? prev.filter(a => a !== asset) : [...prev, asset]
    );
  };

  return (
    <div className="modal-overlay">
      <div
        className="modal-card"
        style={{
          width: 520,
          maxWidth: 'calc(100vw - 32px)',
          padding: 28,
          display: 'flex',
          flexDirection: 'column',
          gap: 20,
        }}
      >
        {/* Header */}
        <div style={{ textAlign: 'center' }}>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: '50%',
              background: 'var(--accent-dim)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 14px',
            }}
          >
            <Key size={20} color="var(--accent)" />
          </div>
          <h2 style={{ fontSize: 17, fontWeight: 600, margin: '0 0 6px' }}>
            {mode === 'update' ? 'Update session key' : 'Register a new session key'}
          </h2>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', margin: 0, lineHeight: 1.5 }}>
            {mode === 'update'
              ? 'Modify the assets or extend the expiry. A new version will be signed.'
              : 'A session key signs state updates on your behalf — no MetaMask for every operation.'}
          </p>
        </div>

        {/* Divider */}
        <div style={{ height: 1, background: 'var(--border)', margin: '0 -28px' }} />

        {/* Assets + Expiry side by side */}
        <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>

          {/* Assets column */}
          <div style={{ flex: '0 0 180px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
              <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Assets
              </span>
              <span style={{ display: 'flex', gap: 6, fontSize: 11, color: 'var(--text-muted)' }}>
                <button
                  type="button"
                  style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', padding: 0, fontSize: 11 }}
                  onClick={() => setSelectedAssets([...supportedAssets])}
                >
                  All
                </button>
                <span>·</span>
                <button
                  type="button"
                  style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', padding: 0, fontSize: 11 }}
                  onClick={() => setSelectedAssets([])}
                >
                  None
                </button>
              </span>
            </div>
            {/* Scrollable asset list — shows ~3 items before scrolling */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6, maxHeight: 138, overflowY: 'auto' }}>
              {supportedAssets.map(asset => {
                const checked = selectedAssets.includes(asset);
                return (
                  <div
                    key={asset}
                    onClick={() => toggleAsset(asset)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 10,
                      padding: '8px 12px',
                      borderRadius: 8,
                      cursor: 'pointer',
                      background: checked ? 'var(--accent-dim)' : 'var(--bg-elevated)',
                      border: `1px solid ${checked ? 'var(--accent)' : 'var(--border)'}`,
                      transition: 'border-color 0.15s, background 0.15s',
                      flexShrink: 0,
                    }}
                  >
                    <div
                      style={{
                        width: 15,
                        height: 15,
                        borderRadius: 4,
                        flexShrink: 0,
                        border: `2px solid ${checked ? 'var(--accent)' : 'var(--border)'}`,
                        background: checked ? 'var(--accent)' : 'transparent',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        transition: 'background 0.15s, border-color 0.15s',
                      }}
                    >
                      {checked && (
                        <svg width="9" height="7" viewBox="0 0 10 8" fill="none">
                          <path d="M1 4l3 3 5-6" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                      )}
                    </div>
                    <span style={{ fontSize: 13, fontWeight: 500 }}>{asset}</span>
                  </div>
                );
              })}
            </div>
            {selectedAssets.length === 0 && (
              <p style={{ fontSize: 11, color: 'var(--error)', marginTop: 5 }}>Select at least one.</p>
            )}
          </div>

          {/* Vertical separator */}
          <div style={{ width: 1, background: 'var(--border)', alignSelf: 'stretch' }} />

          {/* Expiry column */}
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 12, fontWeight: 500, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 10 }}>
              Expiration
            </div>

            {/* Segmented control */}
            <div style={{ display: 'flex', gap: 2, background: 'var(--bg-elevated)', border: '1px solid var(--border)', borderRadius: 8, padding: 3, marginBottom: 12 }}>
              {(['duration', 'date', 'unix'] as ExpiryMode[]).map(m => (
                <button
                  key={m}
                  type="button"
                  onClick={() => setExpiryMode(m)}
                  style={{
                    flex: 1,
                    padding: '5px 0',
                    borderRadius: 6,
                    fontSize: 11,
                    fontWeight: 500,
                    cursor: 'pointer',
                    border: expiryMode === m ? '1px solid var(--border)' : '1px solid transparent',
                    background: expiryMode === m ? 'var(--bg-surface)' : 'transparent',
                    color: expiryMode === m ? 'var(--text-primary)' : 'var(--text-muted)',
                    transition: 'background 0.15s, color 0.15s',
                    textTransform: 'capitalize',
                    fontFamily: 'inherit',
                  }}
                >
                  {m}
                </button>
              ))}
            </div>

            {expiryMode === 'duration' && (
              <div className="input" style={{ display: 'flex', alignItems: 'center' }}>
                <input
                  type="number"
                  min="1"
                  value={durationValue}
                  onChange={e => setDurationValue(e.target.value)}
                  style={{ flex: 1, background: 'transparent', border: 'none', outline: 'none', padding: '7px 10px', fontFamily: 'JetBrains Mono, monospace', fontSize: 13, color: 'var(--text-primary)', minWidth: 0 }}
                  placeholder="24"
                />
                <select
                  value={durationUnit}
                  onChange={e => setDurationUnit(e.target.value as 'minutes' | 'hours' | 'days')}
                  style={{ background: 'var(--bg-base)', border: 'none', borderLeft: '1px solid var(--border)', padding: '7px 8px', fontSize: 12, color: 'var(--text-primary)', cursor: 'pointer', outline: 'none', fontFamily: 'inherit' }}
                >
                  <option value="minutes">minutes</option>
                  <option value="hours">hours</option>
                  <option value="days">days</option>
                </select>
              </div>
            )}

            {expiryMode === 'date' && (
              <div className="input">
                <input
                  type="datetime-local"
                  value={dateValue}
                  onChange={e => setDateValue(e.target.value)}
                  style={{ flex: 1, background: 'transparent', border: 'none', outline: 'none', padding: '7px 10px', fontSize: 12, color: 'var(--text-primary)', width: '100%', fontFamily: 'inherit' }}
                />
              </div>
            )}

            {expiryMode === 'unix' && (
              <div className="input">
                <input
                  type="number"
                  value={unixValue}
                  onChange={e => setUnixValue(e.target.value)}
                  placeholder="1748424600"
                  style={{ flex: 1, background: 'transparent', border: 'none', outline: 'none', padding: '7px 10px', fontFamily: 'JetBrains Mono, monospace', fontSize: 13, color: 'var(--text-primary)', width: '100%' }}
                />
              </div>
            )}

            <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 8 }}>
              Expires:{' '}
              <span className="mono" style={{ color: computedExpiresAt ? 'var(--text-primary)' : 'var(--error)' }}>
                {computedExpiryDisplay}
              </span>
            </div>
          </div>
        </div>

        {/* Divider */}
        <div style={{ height: 1, background: 'var(--border)', margin: '0 -28px' }} />

        {/* Summary box */}
        <div
          style={{
            background: 'var(--bg-elevated)',
            border: '1px solid var(--border)',
            borderRadius: 8,
            padding: '12px 14px',
            fontSize: 12,
            color: 'var(--text-muted)',
            lineHeight: 1.6,
          }}
        >
          This key will authorize:{' '}
          <strong style={{ color: 'var(--text-primary)', fontWeight: 500 }}>
            {selectedAssets.length > 0 ? selectedAssets.join(', ') : 'no assets selected'}
          </strong>{' '}
          and expire in{' '}
          <strong style={{ color: 'var(--text-primary)', fontWeight: 500 }}>
            {computedExpiresAt
              ? expiryMode === 'duration'
                ? `${durationValue} ${durationUnit === 'minutes' ? 'min' : durationUnit}`
                : computedExpiryDisplay
              : '—'}
          </strong>
          .<br />
          On-chain operations (deposit, checkpoint, approve) will still require MetaMask.
          {mode === 'update' && (
            <>
              <br />
              Version will be incremented from the current one.
            </>
          )}
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
            className="btn btn-primary"
            style={{ flex: 1 }}
            onClick={handleSubmit}
            disabled={!isValid || isSubmitting}
          >
            {isSubmitting ? (
              <>
                <span className="spinner" /> Signing…
              </>
            ) : mode === 'update' ? (
              'Update & Sign'
            ) : (
              'Register & Sign'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
