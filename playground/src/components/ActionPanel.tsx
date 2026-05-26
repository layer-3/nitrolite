import { useState } from 'react';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import type { Asset, Blockchain, Channel } from '@yellow-org/sdk';
import { ChannelType, ChannelStatus } from '@yellow-org/sdk';
import { formatBalance, isValidAddress } from '../utils';
import { chainDisplayName } from '../chainMeta';
import TokenSelector from './TokenSelector';

type Tab = 'deposit' | 'withdraw' | 'transfer';

interface Props {
  assets: Asset[];
  channels: Channel[];
  selectedAsset: string;
  onSelectAsset: (asset: string) => void;
  balance: Decimal | undefined;
  onChainBalance: Decimal | null | undefined;
  currentChainId: bigint | null;
  chains: Blockchain[];
  onDeposit: (chainId: bigint, asset: string, amount: Decimal) => void;
  onWithdraw: (chainId: bigint, asset: string, amount: Decimal) => void;
  onTransfer: (to: Address, asset: string, amount: Decimal) => void;
  isApproving: boolean;
  isDepositing: boolean;
  isWithdrawing: boolean;
  isTransferring: boolean;
  disabled: boolean;
  onSwitchChain: (chainId: bigint) => void;
  closingAsset: string | null;
}

export default function ActionPanel({
  assets,
  channels,
  selectedAsset,
  onSelectAsset,
  balance,
  onChainBalance,
  currentChainId,
  chains,
  onDeposit,
  onWithdraw,
  onTransfer,
  isApproving,
  isDepositing,
  isWithdrawing,
  isTransferring,
  disabled,
  onSwitchChain,
  closingAsset,
}: Props) {
  const [tab, setTab] = useState<Tab>('deposit');
  const [amount, setAmount] = useState('');
  const [recipient, setRecipient] = useState('');
  const [recipientError, setRecipientError] = useState<string | null>(null);

  // Pick the chain for deposit/withdraw: current wallet chain if it's supported, else asset's suggested chain.
  const asset = assets.find(a => a.symbol === selectedAsset);
  const operatingChainId =
    asset && currentChainId && asset.tokens.some(t => t.blockchainId === currentChainId)
      ? currentChainId
      : asset?.suggestedBlockchainId ?? null;
  const operatingChain = chains.find(c => c.id === operatingChainId);
  const operatingChainName = operatingChain
    ? chainDisplayName(operatingChain.id, operatingChain.name)
    : undefined;

  const homeChannel = channels.find(
    c =>
      c.asset.toLowerCase() === selectedAsset.toLowerCase() &&
      c.type === ChannelType.Home &&
      c.status !== ChannelStatus.Closed,
  );
  const homeChainId = homeChannel?.blockchainId ?? null;
  const homeChain = homeChainId ? chains.find(c => c.id === homeChainId) : undefined;
  const homeChainName = homeChain ? chainDisplayName(homeChain.id, homeChain.name) : null;
  const isCrossChain = homeChainId !== null && currentChainId !== homeChainId;
  const isChannelClosing =
    !!closingAsset && closingAsset.toLowerCase() === selectedAsset.toLowerCase();

  const channelBalance = balance ?? new Decimal(0);
  const amountDecimal = (() => {
    try {
      return amount ? new Decimal(amount) : new Decimal(0);
    } catch {
      return new Decimal(0);
    }
  })();
  const amountInvalid = amountDecimal.lte(0);
  const amountExceedsChannel = (tab === 'withdraw' || tab === 'transfer') && amountDecimal.gt(channelBalance);
  const amountExceedsOnChain = tab === 'deposit' && onChainBalance != null && amountDecimal.gt(onChainBalance);

  const setMax = () => {
    if (tab === 'deposit' && onChainBalance) setAmount(onChainBalance.toString());
    else if (tab === 'withdraw' || tab === 'transfer') setAmount(channelBalance.toString());
  };

  const validateRecipient = () => {
    if (!recipient) {
      setRecipientError(null);
      return false;
    }
    if (!isValidAddress(recipient)) {
      setRecipientError('Invalid Ethereum address');
      return false;
    }
    setRecipientError(null);
    return true;
  };

  const fire = () => {
    if (!asset || !operatingChainId) return;
    if (tab === 'deposit') onDeposit(operatingChainId, selectedAsset, amountDecimal);
    else if (tab === 'withdraw') onWithdraw(operatingChainId, selectedAsset, amountDecimal);
    else if (tab === 'transfer') {
      if (!validateRecipient()) return;
      onTransfer(recipient as Address, selectedAsset, amountDecimal);
    }
    setAmount('');
  };

  const isBusy =
    (tab === 'deposit' && isDepositing) ||
    (tab === 'withdraw' && isWithdrawing) ||
    (tab === 'transfer' && isTransferring);

  // For deposit, also require on-chain balance to be loaded before allowing submit.
  const depositBalanceReady = tab !== 'deposit' || onChainBalance != null;

  const transferRecipientValid = tab !== 'transfer' || (isValidAddress(recipient) && !recipientError);
  const canSubmit =
    !disabled &&
    !isBusy &&
    !amountInvalid &&
    !amountExceedsChannel &&
    !amountExceedsOnChain &&
    depositBalanceReady &&
    !!asset &&
    !!operatingChainId &&
    transferRecipientValid;

  const amountInputError = amountExceedsChannel || amountExceedsOnChain;

  return (
    <div className="card">
      <div className="card-header">
        <span className="card-title">Actions</span>
      </div>

      {/* Balance display — persistent */}
      <div className="px-5 py-4 border-b border-border">
        <TokenSelector
          assets={assets}
          selectedAsset={selectedAsset}
          onSelectAsset={onSelectAsset}
          channels={channels}
          chains={chains}
          disabled={disabled}
        />

        <div className="flex justify-between items-baseline mb-1">
          <span className="text-text-muted text-xs">Unified balance</span>
          <span className="mono text-2xl font-medium text-accent">{formatBalance(channelBalance)}</span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-text-muted text-xs">On-chain</span>
          <span className="mono text-text-muted">
            {onChainBalance == null ? '—' : formatBalance(onChainBalance)}
          </span>
        </div>
      </div>

      {/* Tabs + Form — blurred when cross-chain or channel is closing */}
      <div className="relative">
        {/* Tabs */}
        <div className={`flex gap-1 px-5 pt-4 ${isCrossChain || isChannelClosing ? 'blur-sm pointer-events-none select-none' : ''}`}>
          <button className={`tab ${tab === 'deposit' ? 'active' : ''}`} onClick={() => setTab('deposit')}>
            Deposit
          </button>
          <button className={`tab ${tab === 'withdraw' ? 'active' : ''}`} onClick={() => setTab('withdraw')}>
            Withdraw
          </button>
          <button className={`tab ${tab === 'transfer' ? 'active' : ''}`} onClick={() => setTab('transfer')}>
            Transfer
          </button>
        </div>

        {/* Form */}
        <div className={`px-5 pt-4 pb-5 ${isCrossChain || isChannelClosing ? 'blur-sm pointer-events-none select-none' : ''}`}>
          {tab === 'transfer' && (
            <div className="mb-3">
              <label className="block text-xs text-text-muted mb-1.5">Recipient address</label>
              <div className={`input ${recipientError ? 'error' : ''}`}>
                <input
                  type="text"
                  placeholder="0x…"
                  value={recipient}
                  onChange={e => {
                    setRecipient(e.target.value);
                    if (recipientError) setRecipientError(null);
                  }}
                  onBlur={validateRecipient}
                />
              </div>
              {recipientError && <p className="text-error text-xs mt-1.5">{recipientError}</p>}
            </div>
          )}

          <label className="block text-xs text-text-muted mb-1.5">Amount</label>
          <div className={`input ${amountInputError ? 'error' : ''}`}>
            <input
              type="text"
              inputMode="decimal"
              placeholder="0.00"
              value={amount}
              onChange={e => setAmount(e.target.value)}
            />
            <button type="button" className="input-max" onClick={setMax} title="Use max available">
              MAX
            </button>
            <span className="input-suffix">{selectedAsset.toUpperCase() || '—'}</span>
          </div>

          {amountExceedsChannel && (
            <p className="text-error text-xs mt-1.5">Amount exceeds Unified balance</p>
          )}
          {amountExceedsOnChain && (
            <p className="text-error text-xs mt-1.5">Amount exceeds on-chain balance</p>
          )}

          {operatingChainName && tab === 'deposit' && (
            <p className="text-text-muted text-xs mt-2">
              Will deposit on <span className="text-text-primary">"{operatingChainName}"</span>
            </p>
          )}

          {operatingChainName && tab === 'withdraw' && (
            <p className="text-text-muted text-xs mt-2">
              Will withdraw to <span className="text-text-primary">"{operatingChainName}"</span>
            </p>
          )}

          <button
            className="btn btn-primary w-full mt-4"
            onClick={fire}
            disabled={!canSubmit}
          >
            {isBusy ? (
              <>
                <span className="spinner" />
                {tab === 'deposit' && (isApproving ? 'Approving…' : 'Depositing…')}
                {tab === 'withdraw' && 'Withdrawing…'}
                {tab === 'transfer' && 'Transferring…'}
              </>
            ) : tab === 'deposit' ? (
              'Deposit via MetaMask'
            ) : tab === 'withdraw' ? (
              'Withdraw to wallet'
            ) : (
              'Transfer'
            )}
          </button>
        </div>

        {/* Cross-chain overlay */}
        {isCrossChain && !isChannelClosing && (
          <div className="absolute inset-0 flex flex-col items-center justify-center px-6 gap-3">
            <p className="text-text-primary text-sm text-center leading-relaxed">
              Cross-chain operations are not yet supported. Please select this asset home chain to perform operations.
            </p>
            <button
              className="btn btn-primary text-xs px-4 py-2"
              onClick={() => homeChainId && onSwitchChain(homeChainId)}
            >
              Select "{homeChainName ?? 'home chain'}"
            </button>
          </div>
        )}

        {/* Channel-closing overlay */}
        {isChannelClosing && (
          <div className="absolute inset-0 flex flex-col items-center justify-center px-6 gap-3">
            <span className="spinner" />
            <p className="text-text-primary text-sm text-center leading-relaxed">
              Channel is being closed
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
