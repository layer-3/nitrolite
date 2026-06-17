import { useState, useCallback } from 'react';
import { RefreshCw } from 'lucide-react';
import type { Client, Channel, Blockchain } from '@yellow-org/sdk';
import { ChannelStatus } from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import ChannelRow from './ChannelRow';
import IncomingChannelRow from './IncomingChannelRow';

interface Props {
  channels: Channel[];
  client: Client | null;
  address: Address | null;
  chains: Blockchain[];
  currentChainId: bigint | null;
  balances: Record<string, Decimal>;
  isLoading: boolean;
  closingAsset: string | null;
  awaitingCloseAsset: string | null;
  onRefresh: () => void;
  onClose: (asset: string, blockchainId: bigint) => void;
  onSwitchToHomeChain: (chainId: bigint) => void;
  onSelectAsset: (asset: string) => void;
  onAfterOp: () => void;
  channelStatesKey?: number;
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
  awaitingCloseAsset,
  onRefresh,
  onClose,
  onSwitchToHomeChain,
  onSelectAsset,
  onAfterOp,
  channelStatesKey,
}: Props) {
  const [expandedIncoming, setExpandedIncoming] = useState<Set<string>>(new Set());

  const handleIncomingExpandChange = useCallback((asset: string, expanded: boolean) => {
    setExpandedIncoming(prev => {
      const next = new Set(prev);
      if (expanded) next.add(asset.toLowerCase());
      else next.delete(asset.toLowerCase());
      return next;
    });
  }, []);

  const channelAssets = new Set(
    channels.filter(c => c.status !== ChannelStatus.Closed).map(c => c.asset.toLowerCase()),
  );
  const incomingAssets = Object.entries(balances).filter(
    ([asset, bal]) => bal.gt(0) && !channelAssets.has(asset.toLowerCase()),
  );

  const isEmpty = channels.length === 0 && incomingAssets.length === 0;

  return (
    <div className="card">
      <div className="card-header">
        <div className="flex items-center gap-2">
          <span className="card-title">Channels</span>
          <span className="text-text-muted text-sm mono">({channels.length + incomingAssets.length})</span>
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
        {isEmpty ? (
          <p className="text-text-muted text-sm text-center py-8">
            {isLoading ? 'Loading…' : 'No channels yet. Deposit to open one.'}
          </p>
        ) : (
          <>
            {incomingAssets.map(([asset, bal]) => (
              <IncomingChannelRow
                key={`incoming-${asset}`}
                asset={asset}
                balance={bal}
                client={client}
                address={address}
                currentChainId={currentChainId}
                onSelectAsset={onSelectAsset}
                onAfterAck={onAfterOp}
                onExpandChange={handleIncomingExpandChange}
              />
            ))}
            {channels.map(c => (
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
                onSelectAsset={onSelectAsset}
                onAfterOp={onAfterOp}
                isClosing={closingAsset?.toLowerCase() === c.asset.toLowerCase()}
                isAwaitingClose={awaitingCloseAsset?.toLowerCase() === c.asset.toLowerCase()}
                channelStatesKey={channelStatesKey}
                defaultExpanded={expandedIncoming.has(c.asset.toLowerCase())}
              />
            ))}
          </>
        )}
      </div>
    </div>
  );
}
