import { useState } from 'react';
import { Copy, Check } from 'lucide-react';

interface Props {
  value: string;
  className?: string;
  size?: number;
}

export default function CopyButton({ value, className = '', size = 13 }: Props) {
  const [copied, setCopied] = useState(false);

  const onClick = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard denied */
    }
  };

  return (
    <button
      type="button"
      onClick={onClick}
      className={`inline-flex items-center justify-center p-1 text-text-muted hover:text-text-primary transition-colors ${className}`}
      title={copied ? 'Copied' : 'Copy'}
      aria-label="Copy"
    >
      {copied ? <Check size={size} className="text-success" /> : <Copy size={size} />}
    </button>
  );
}
