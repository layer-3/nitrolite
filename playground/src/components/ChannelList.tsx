import { RefreshCw } from 'lucide-react';
import type { Client, Channel, Blockchain } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import ChannelRow from './ChannelRow';

interface Props {
  channels: Channel[];
  client: Client | null;
  address: Address | null;
  chains: Blockchain[];
  currentChainId: bigint | null;
  balances: Record<string, Decimal>;
  isLoading: boolean;
  closingAsset: string | null;
  onRefresh: () => void;
  onClose: (asset: string, blockchainId: bigint) => void;
  onSwitchToHomeChain: (chainId: bigint) => void;
  onAfterOp: () => void;
}

export default function ChannelList({
  channels,
  client,
  address,
  chains,
  currentChainId,
  balances,
  isLoading,
  closingAsset,
  onRefresh,
  onClose,
  onSwitchToHomeChain,
  onAfterOp,
}: Props) {
  return (
    <div className="card">
      <div className="card-header">
        <div className="flex items-center gap-2">
          <span className="card-title">Channels</span>
          <span className="text-text-muted text-sm mono">({channels.length})</span>
        </div>
        <button
          className="btn btn-ghost btn-sm"
          onClick={onRefresh}
          disabled={isLoading}
          title="Refresh"
        >
          <RefreshCw size={13} className={isLoading ? 'animate-spin' : ''} />
        </button>
      </div>

      <div className="p-4">
        {channels.length === 0 ? (
          <p className="text-text-muted text-sm text-center py-8">
            {isLoading ? 'Loading…' : 'No channels yet. Deposit to open one.'}
          </p>
        ) : (
          channels.map(c => (
            <ChannelRow
              key={c.channelId}
              channel={c}
              client={client}
              address={address}
              chains={chains}
              currentChainId={currentChainId}
              enforcedBalance={balances[c.asset] ?? null}
              onClose={onClose}
              onSwitchToHomeChain={onSwitchToHomeChain}
              onAfterOp={onAfterOp}
              isClosing={closingAsset?.toLowerCase() === c.asset.toLowerCase()}
            />
          ))
        )}
      </div>
    </div>
  );
}
