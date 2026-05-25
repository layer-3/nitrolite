import { isAddress } from 'viem';

export function formatAddress(address: string): string {
  if (!address || address.length < 10) return address;
  return `${address.slice(0, 6)}…${address.slice(-4)}`;
}

export function formatBalance(value: { toString(): string } | null | undefined): string {
  if (!value) return '0.00';
  const str = value.toString();
  const parts = str.split('.');
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  if (parts.length === 1) parts.push('00');
  else if (parts[1].length === 1) parts[1] = parts[1] + '0';
  return parts.join('.');
}

export function timeAgo(date: Date | string | number | null): string {
  if (!date) return 'never';
  const then = typeof date === 'number' ? date : new Date(date).getTime();
  const diff = Math.max(0, Date.now() - then);
  if (diff < 2000) return 'just now';
  if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`;
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  return new Date(then).toLocaleDateString();
}

export function isValidAddress(value: string): boolean {
  try {
    return isAddress(value, { strict: false });
  } catch {
    return false;
  }
}
