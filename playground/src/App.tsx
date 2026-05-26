import { useCallback, useEffect, useMemo, useState } from 'react';
import { Toaster } from 'sonner';
import { useWallet } from './hooks/useWallet';
import { useNitrolite } from './hooks/useNitrolite';
import { useChannels } from './hooks/useChannels';
import { useChannelOps } from './hooks/useChannelOps';
import { useSessionKey } from './hooks/useSessionKey';
import WalletBar from './components/WalletBar';
import ActionPanel from './components/ActionPanel';
import ChannelList from './components/ChannelList';
import PendingReceipts from './components/PendingReceipts';
import UnsupportedChainModal from './components/UnsupportedChainModal';
import SetHomechainModal from './components/SetHomechainModal';
import SessionKeyBanner from './components/SessionKeyBanner';
import SessionKeySetupModal from './components/SessionKeySetupModal';

export default function App() {
  const wallet = useWallet();
  const [showSkModal, setShowSkModal] = useState(false);

  const sk = useSessionKey(wallet.address);
  const nitro = useNitrolite(wallet.address, wallet.walletClient, sk.sessionKey);
  const channels = useChannels(nitro.client, wallet.address);

  const refreshAll = useCallback(() => {
    nitro.refresh();
    channels.refresh();
  }, [nitro, channels]);

  const [channelStatesKey, setChannelStatesKey] = useState(0);
  const bumpChannelStates = useCallback(() => setChannelStatesKey(k => k + 1), []);

  const ops = useChannelOps(nitro.client, wallet.address, nitro.supportedAssets, refreshAll, bumpChannelStates);

  const [selectedAsset, setSelectedAsset] = useState('');

  // Default-select an asset when assets load.
  useEffect(() => {
    if (!selectedAsset && nitro.supportedAssets.length) {
      setSelectedAsset(nitro.supportedAssets[0].symbol);
    }
  }, [nitro.supportedAssets, selectedAsset]);

  const isChainSupported = useMemo(() => {
    if (!wallet.chainId || !nitro.supportedChains.length) return true;
    return nitro.supportedChains.some(c => c.id === wallet.chainId);
  }, [wallet.chainId, nitro.supportedChains]);

  const showUnsupportedModal = !!wallet.address && !!nitro.isConnected && !isChainSupported;
  const showSkBanner = !!wallet.address && nitro.isConnected && !sk.sessionKey;

  const onConfirmSk = async () => {
    if (!nitro.client) return;
    await sk.register(
      nitro.client,
      nitro.supportedAssets.map(a => a.symbol),
    );
  };

  // Close modal when registration finishes successfully (sessionKey becomes set).
  useEffect(() => {
    if (sk.sessionKey && showSkModal) setShowSkModal(false);
  }, [sk.sessionKey, showSkModal]);

  return (
    <div className="min-h-screen flex flex-col">
      <Toaster
        theme="dark"
        position="bottom-right"
        toastOptions={{
          style: {
            background: 'var(--bg-elevated)',
            border: '1px solid var(--border)',
            color: 'var(--text-primary)',
            fontFamily: 'Inter, system-ui, sans-serif',
          },
        }}
      />

      <WalletBar
        address={wallet.address}
        chainId={wallet.chainId}
        chains={nitro.supportedChains}
        lastCommsAt={nitro.lastCommsAt}
        nodeError={nitro.nodeError}
        isConnecting={wallet.isConnecting}
        sessionKey={sk.sessionKey}
        onConnect={wallet.connect}
        onDisconnect={wallet.disconnect}
        onSwitchChain={wallet.switchChain}
        onClearSessionKey={sk.clear}
      />

      <main className="flex-1 mx-auto w-full max-w-[1100px] px-6 py-8">
        {!wallet.address ? (
          <ConnectPrompt />
        ) : nitro.isConnecting ? (
          <div className="text-text-muted text-sm text-center py-16">Connecting to Nitronode…</div>
        ) : (
          <>
            {showSkBanner && <SessionKeyBanner onSetup={() => setShowSkModal(true)} />}

            <div className="grid grid-cols-1 md:grid-cols-[360px_1fr] gap-5 items-start">
              <ActionPanel
                assets={nitro.supportedAssets}
                channels={channels.channels}
                selectedAsset={selectedAsset}
                onSelectAsset={setSelectedAsset}
                balance={nitro.balances[selectedAsset]}
                onChainBalance={nitro.onChainBalances[selectedAsset]}
                currentChainId={wallet.chainId}
                chains={nitro.supportedChains}
                onDeposit={ops.deposit}
                onWithdraw={ops.withdraw}
                onTransfer={ops.transfer}
                depositPhase={ops.depositPhase}
                withdrawPhase={ops.withdrawPhase}
                transferPhase={ops.transferPhase}
                needsApproval={ops.needsApproval}
                checkDepositAllowance={ops.checkDepositAllowance}
                disabled={!nitro.client || showUnsupportedModal}
                onSwitchChain={wallet.switchChain}
                closingAsset={ops.closingAsset}
              />

              <div>
                <PendingReceipts
                  client={nitro.client}
                  channels={channels.channels}
                  balances={nitro.balances}
                  onAfterAck={refreshAll}
                />
                <ChannelList
                  channels={channels.channels}
                  client={nitro.client}
                  address={wallet.address}
                  chains={nitro.supportedChains}
                  currentChainId={wallet.chainId}
                  balances={nitro.balances}
                  isLoading={channels.isLoading}
                  closingAsset={ops.closingAsset}
                  onRefresh={refreshAll}
                  onClose={ops.closeChannel}
                  onSwitchToHomeChain={wallet.switchChain}
                  onSelectAsset={setSelectedAsset}
                  onAfterOp={refreshAll}
                  channelStatesKey={channelStatesKey}
                />
              </div>
            </div>
          </>
        )}
      </main>

      {showUnsupportedModal && (
        <UnsupportedChainModal chains={nitro.supportedChains} onSwitchChain={wallet.switchChain} />
      )}

      {ops.homechainModalAsset && (
        <SetHomechainModal
          asset={ops.homechainModalAsset}
          assets={nitro.supportedAssets}
          chains={nitro.supportedChains}
          onConfirm={ops.onHomechainSelected}
          onCancel={ops.onHomechainModalDismiss}
        />
      )}

      {showSkModal && (
        <SessionKeySetupModal
          assets={nitro.supportedAssets.map(a => a.symbol)}
          isRegistering={sk.isRegistering}
          mode={sk.sessionKey ? 'renew' : 'setup'}
          onConfirm={onConfirmSk}
          onCancel={() => setShowSkModal(false)}
        />
      )}
    </div>
  );
}

function ConnectPrompt() {
  return (
    <div className="card max-w-md mx-auto mt-12 p-8 text-center">
      <h1 className="text-xl font-semibold mb-2">Nitrolite Playground</h1>
      <p className="text-text-muted text-sm mb-6">
        Connect a wallet to inspect channels, deposit, withdraw, and transfer assets via the Nitronode.
      </p>
      <p className="text-text-muted text-xs">Use the Connect MetaMask button in the top right.</p>
    </div>
  );
}
