import { useState } from 'react';
import { Decimal } from 'decimal.js';
import type { Address } from 'viem';
import type { Asset, Blockchain } from '@yellow-org/sdk';
import { formatBalance, isValidAddress } from '../utils';

type Tab = 'deposit' | 'withdraw' | 'transfer';

interface Props {
  assets: Asset[];
  selectedAsset: string;
  onSelectAsset: (asset: string) => void;
  balance: Decimal | undefined;
  onChainBalance: Decimal | null | undefined;
  currentChainId: bigint | null;
  chains: Blockchain[];
  onDeposit: (chainId: bigint, asset: string, amount: Decimal) => void;
  onWithdraw: (chainId: bigint, asset: string, amount: Decimal) => void;
  onTransfer: (to: Address, asset: string, amount: Decimal) => void;
  isDepositing: boolean;
  isWithdrawing: boolean;
  isTransferring: boolean;
  disabled: boolean;
}

export default function ActionPanel({
  assets,
  selectedAsset,
  onSelectAsset,
  balance,
  onChainBalance,
  currentChainId,
  chains,
  onDeposit,
  onWithdraw,
  onTransfer,
  isDepositing,
  isWithdrawing,
  isTransferring,
  disabled,
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
  const operatingChainName = chains.find(c => c.id === operatingChainId)?.name;

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

  const transferRecipientValid = tab !== 'transfer' || (isValidAddress(recipient) && !recipientError);
  const canSubmit =
    !disabled &&
    !isBusy &&
    !amountInvalid &&
    !amountExceedsChannel &&
    !amountExceedsOnChain &&
    !!asset &&
    !!operatingChainId &&
    transferRecipientValid;

  return (
    <div className="card">
      <div className="card-header">
        <span className="card-title">Actions</span>
      </div>

      {/* Balance display — persistent */}
      <div className="px-5 py-4 border-b border-border">
        <select
          value={selectedAsset}
          onChange={e => onSelectAsset(e.target.value)}
          className="chip mono cursor-pointer mb-3"
          disabled={!assets.length}
        >
          {!assets.length && <option value="">— no assets —</option>}
          {assets.map(a => (
            <option key={a.symbol} value={a.symbol}>
              {a.symbol.toUpperCase()}
            </option>
          ))}
        </select>

        <div className="flex justify-between items-baseline mb-1">
          <span className="text-text-muted text-xs">Channel</span>
          <span className="mono text-2xl font-medium text-accent">{formatBalance(channelBalance)}</span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-text-muted text-xs">On-chain</span>
          <span className="mono text-text-muted">
            {onChainBalance == null ? '—' : formatBalance(onChainBalance)}
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 px-5 pt-4">
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
      <div className="px-5 pt-4 pb-5">
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
        <div className="input">
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
          <p className="text-error text-xs mt-1.5">Amount exceeds channel balance</p>
        )}
        {amountExceedsOnChain && (
          <p className="text-error text-xs mt-1.5">Amount exceeds on-chain balance</p>
        )}

        {operatingChainName && tab === 'deposit' && (
          <p className="text-text-muted text-xs mt-2">
            Will deposit on <span className="text-text-primary">{operatingChainName}</span>
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
              {tab === 'deposit' && 'Depositing…'}
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
    </div>
  );
}
