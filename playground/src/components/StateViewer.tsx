import type { Client } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import { useChannelStates } from '../hooks/useChannelStates';
import { formatBalance } from '../utils';

interface Props {
  client: Client | null;
  address: Address | null;
  asset: string;
  enforcedBalance: Decimal | null | undefined;
  onAfterOp?: () => void;
  isLocked?: boolean;
  refreshKey?: number;
}

export default function StateViewer({ client, address, asset, enforcedBalance, onAfterOp, isLocked, refreshKey }: Props) {
  const {
    enforced,
    signed,
    issued,
    isLoading,
    error,
    canAcknowledge,
    canCheckpoint,
    acknowledge,
    checkpoint,
    isAcknowledging,
    isCheckpointing,
  } = useChannelStates(client, address, asset, enforcedBalance, onAfterOp, refreshKey);

  if (isLoading && !enforced && !signed && !issued) {
    return <div className="text-text-muted text-xs px-4 py-3">Loading states…</div>;
  }
  if (error) {
    return <div className="text-error text-xs px-4 py-3">Failed to fetch states: {error}</div>;
  }
  if (!enforced && !signed && !issued) {
    return <div className="text-text-muted text-xs px-4 py-3">No state yet</div>;
  }

  return (
    <div className="border border-border rounded-lg">
      <div className="grid grid-cols-[120px_70px_1fr_auto] gap-3 px-4 py-2 text-text-muted text-[10px] uppercase tracking-wider border-b border-border bg-bg-base/40 rounded-t-lg">
        <span>State</span>
        <span>Version</span>
        <span>Amount</span>
        <span></span>
      </div>

      <StateRow
        label="Enforced"
        dotColor="success"
        tooltip="Last on-chain checkpoint. Guaranteed by the blockchain."
        version={enforced?.stateVersion}
        amount={enforced?.amount}
        asset={asset}
      />
      <StateRow
        label="Signed"
        dotColor="accent"
        tooltip="Both you and the node signed this state. Safe bilateral proof."
        version={signed?.version}
        amount={signed?.homeLedger.userBalance}
        asset={asset}
        action={
          canCheckpoint ? (
            <button className="btn btn-sm" onClick={checkpoint} disabled={isCheckpointing || isLocked}>
              {isCheckpointing || isLocked ? <span className="spinner" /> : 'Checkpoint'}
            </button>
          ) : null
        }
      />
      <StateRow
        label="Issued"
        dotColor="blue"
        tooltip="Node proposed this state. Click Acknowledge to co-sign it."
        version={issued?.version}
        amount={issued?.homeLedger.userBalance}
        asset={asset}
        action={
          canAcknowledge ? (
            <button
              className="btn btn-sm"
              style={{ borderColor: 'rgba(96,165,250,0.4)', color: '#60a5fa' }}
              onClick={acknowledge}
              disabled={isAcknowledging || isLocked}
            >
              {isAcknowledging || isLocked ? <span className="spinner" /> : 'Acknowledge'}
            </button>
          ) : null
        }
        isLast
      />
    </div>
  );
}

interface RowProps {
  label: string;
  dotColor: 'success' | 'accent' | 'blue';
  tooltip: string;
  version?: bigint;
  amount?: Decimal;
  asset: string;
  action?: React.ReactNode;
  isLast?: boolean;
}

function StateRow({ label, dotColor, tooltip, version, amount, asset, action, isLast }: RowProps) {
  const colorMap = {
    success: 'var(--success)',
    accent: 'var(--accent)',
    blue: '#60a5fa',
  };
  return (
    <div
      className={`grid grid-cols-[120px_70px_1fr_auto] gap-3 items-center px-4 py-3 ${
        isLast ? 'rounded-b-lg' : 'border-b border-border'
      } hover:bg-bg-elevated/50 transition-colors`}
    >
      <div className="flex items-center gap-2 text-xs font-semibold" style={{ color: colorMap[dotColor] }}>
        <span
          className="inline-block w-2 h-2 rounded-full"
          style={{ background: colorMap[dotColor] }}
        />
        {label}
        <span className="tooltip-wrap">
          <span className="help-icon">?</span>
          <span className="tip">{tooltip}</span>
        </span>
      </div>
      <span className="mono text-xs text-text-muted">
        {version != null ? `v${version.toString()}` : '—'}
      </span>
      <span className="mono text-sm text-text-primary">
        {amount != null ? `${formatBalance(amount)} ${asset.toUpperCase()}` : '—'}
      </span>
      <span>{action ?? <span className="text-text-muted text-xs">—</span>}</span>
    </div>
  );
}
