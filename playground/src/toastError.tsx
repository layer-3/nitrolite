import { useState } from 'react';
import { toast } from 'sonner';
import { X, Copy, Check } from 'lucide-react';

const MAX_CHARS = 200;

// Not exported — callers only see showErrorToast().
function ErrorToastContent({ id, message }: { id: string | number; message: string }) {
  const isLong = message.length > MAX_CHARS;
  const display = isLong ? `${message.slice(0, MAX_CHARS)}…` : message;
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(message).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'flex-start',
      gap: 10,
      padding: '12px 36px 12px 14px',
      background: 'var(--bg-elevated)',
      border: '1px solid var(--border)',
      borderLeft: '3px solid var(--error)',
      borderRadius: 8,
      boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
      fontFamily: 'Inter, system-ui, sans-serif',
      fontSize: 13,
      color: 'var(--text-primary)',
      width: '100%',
      boxSizing: 'border-box',
      position: 'relative',
    }}>
      <span style={{ flex: 1, lineHeight: 1.5, wordBreak: 'break-word' }}>
        {display}
      </span>
      <div style={{
        position: 'absolute',
        top: 8,
        right: 8,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
      }}>
        <button
          onClick={() => toast.dismiss(id)}
          style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, color: 'var(--text-muted)', display: 'flex', lineHeight: 1 }}
          aria-label="Dismiss"
        >
          <X size={13} />
        </button>
        {isLong && (
          <button
            onClick={handleCopy}
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2, display: 'flex', lineHeight: 1, color: copied ? 'var(--success)' : 'var(--text-muted)' }}
            aria-label="Copy full error"
            title="Copy full error"
          >
            {copied ? <Check size={12} /> : <Copy size={12} />}
          </button>
        )}
      </div>
    </div>
  );
}

export function showErrorToast(message: string) {
  toast.custom(id => <ErrorToastContent id={id} message={message} />, {
    duration: Infinity,
    // toast.custom() sets data-styled="false" so Sonner skips its own width/padding CSS.
    // Strip the li's toastOptions styles and restore the correct width.
    style: { padding: 0, background: 'transparent', border: 'none', boxShadow: 'none', width: 'var(--width)' },
  });
}
