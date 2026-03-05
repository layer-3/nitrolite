import { useState, useCallback, useEffect } from 'react';
import { Sun, Moon } from 'lucide-react';
import { Client, withBlockchainRPC, ChannelDefaultSigner, ChannelSessionKeyStateSigner } from '@yellow-org/sdk';
import { createWalletClient, custom, type WalletClient } from 'viem';
import { mainnet } from 'viem/chains';
import { WalletStateSigner, WalletTransactionSigner } from './walletSigners';
import WalletDashboard from './components/WalletDashboard';
import StatusBar from './components/StatusBar';
import type { AppState, ClearnodeConfig, NetworkConfig, SessionKeyState, StatusMessage } from './types';
import { formatAddress } from './utils';

const CLEARNODES: ClearnodeConfig[] = [
  { name: 'YN Testnet', url: 'wss://clearnode-v1-rc.yellow.org/ws' },
];

const NETWORKS: NetworkConfig[] = [
  { chainId: '11155111', name: 'Sepolia', rpcUrl: 'https://ethereum-sepolia-rpc.publicnode.com' },
];

function App() {
  const [appState, setAppState] = useState<AppState>(() => {
    let sessionKey: SessionKeyState | null = null;
    try {
      const stored = localStorage.getItem('nitrolite_session_key');
      if (stored) sessionKey = JSON.parse(stored);
    } catch { /* ignore */ }

    return {
      client: null,
      address: null,
      connected: false,
      nodeUrl: CLEARNODES[0].url,
      selectedChainId: NETWORKS[0].chainId,
      selectedAsset: '',
      sessionKey,
    };
  });
  const [assets, setAssets] = useState<string[]>([]);
  const [status, setStatus] = useState<StatusMessage | null>(null);
  const [walletClient, setWalletClient] = useState<WalletClient | null>(null);
  const [autoConnecting, setAutoConnecting] = useState(true);

  // Theme state
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    return (localStorage.getItem('theme') as 'dark' | 'light') || 'dark';
  });

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggleTheme = useCallback(() => {
    setTheme(prev => prev === 'dark' ? 'light' : 'dark');
  }, []);

  const showStatus = useCallback((type: StatusMessage['type'], message: string, details?: string) => {
    setStatus({ type, message, details });
    setTimeout(() => setStatus(null), 5000);
  }, []);

  const buildClient = useCallback(async (
    wc: WalletClient,
    nodeUrl: string,
    sessionKey: SessionKeyState | null,
    address: string,
  ): Promise<Client> => {
    const options = NETWORKS.map(n =>
      withBlockchainRPC(BigInt(n.chainId), n.rpcUrl)
    );

    let stateSigner;
    if (sessionKey && sessionKey.active && sessionKey.metadataHash && sessionKey.authSig) {
      stateSigner = new ChannelSessionKeyStateSigner(
        sessionKey.privateKey as `0x${string}`,
        address as `0x${string}`,
        sessionKey.metadataHash as `0x${string}`,
        sessionKey.authSig as `0x${string}`,
      );
    } else {
      stateSigner = new ChannelDefaultSigner(new WalletStateSigner(wc));
    }

    const txSigner = new WalletTransactionSigner(wc);
    return await Client.create(nodeUrl, stateSigner, txSigner, ...options);
  }, []);

  const fetchAssets = useCallback(async (client: Client) => {
    try {
      const nodeAssets = await client.getAssets();
      const symbols = nodeAssets.map(a => a.symbol.toLowerCase());
      setAssets(symbols);
      setAppState(prev => ({
        ...prev,
        selectedAsset: prev.selectedAsset || symbols[0] || '',
      }));
    } catch (error) {
      console.error('Failed to fetch assets:', error);
    }
  }, []);

  // Auto-reconnect on page load
  useEffect(() => {
    const autoConnect = async () => {
      try {
        const wasConnected = localStorage.getItem('metamask_connected');
        if (!wasConnected || typeof window.ethereum === 'undefined') {
          setAutoConnecting(false);
          return;
        }

        const accounts = await window.ethereum.request({
          method: 'eth_accounts'
        }) as string[];

        if (accounts && accounts.length > 0) {
          const address = accounts[0];
          const wc = createWalletClient({
            account: address as `0x${string}`,
            chain: mainnet,
            transport: custom(window.ethereum),
          });

          setWalletClient(wc);

          try {
            let storedSk: SessionKeyState | null = null;
            try {
              const raw = localStorage.getItem('nitrolite_session_key');
              if (raw) storedSk = JSON.parse(raw);
            } catch { /* ignore */ }

            const sdkClient = await buildClient(wc, CLEARNODES[0].url, storedSk, address);
            if (storedSk && !(storedSk.active && storedSk.metadataHash && storedSk.authSig)) {
              storedSk = { ...storedSk, active: false };
            }

            setAppState(prev => ({ ...prev, address, client: sdkClient, connected: true, sessionKey: storedSk }));
            await fetchAssets(sdkClient);
          } catch (nodeError) {
            console.error('Auto node connection failed:', nodeError);
            setAppState(prev => ({ ...prev, address, connected: false }));
            showStatus('info', 'Wallet reconnected', 'Node connection failed — retrying...');
          }
        }
      } catch (error) {
        console.error('Auto-connect failed:', error);
      } finally {
        setAutoConnecting(false);
      }
    };

    autoConnect();
  }, [showStatus, buildClient, fetchAssets]);

  // Listen for account changes
  useEffect(() => {
    if (typeof window.ethereum === 'undefined') return;

    const handleAccountsChanged = (accounts: string[]) => {
      if (accounts.length === 0) {
        localStorage.removeItem('metamask_connected');
        setWalletClient(null);
        setAppState(prev => ({ ...prev, address: null, connected: false, client: null }));
        showStatus('info', 'Wallet disconnected');
      } else if (accounts[0] !== appState.address) {
        const newAddress = accounts[0];
        const wc = createWalletClient({
          account: newAddress as `0x${string}`,
          chain: mainnet,
          transport: custom(window.ethereum),
        });
        setWalletClient(wc);
        setAppState(prev => ({ ...prev, address: newAddress, connected: false, client: null }));
        showStatus('info', 'Account switched', `New address: ${newAddress}`);
      }
    };

    window.ethereum.on('accountsChanged', handleAccountsChanged);
    return () => {
      window.ethereum.removeListener('accountsChanged', handleAccountsChanged);
    };
  }, [appState.address, showStatus]);

  const connectWallet = useCallback(async () => {
    try {
      if (typeof window.ethereum === 'undefined') {
        showStatus('error', 'MetaMask not detected', 'Please install MetaMask extension');
        return;
      }

      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' }) as string[];
      if (!accounts || accounts.length === 0) {
        showStatus('error', 'No accounts found', 'Please unlock MetaMask');
        return;
      }

      const address = accounts[0];
      const wc = createWalletClient({
        account: address as `0x${string}`,
        chain: mainnet,
        transport: custom(window.ethereum),
      });

      localStorage.setItem('metamask_connected', 'true');
      setWalletClient(wc);

      try {
        const sdkClient = await buildClient(wc, appState.nodeUrl, appState.sessionKey, address);
        setAppState(prev => ({ ...prev, address, client: sdkClient, connected: true }));
        await fetchAssets(sdkClient);
        const node = CLEARNODES.find(c => c.url === appState.nodeUrl);
        showStatus('success', `Connected to ${node?.name || 'Clearnode'}`);
      } catch (nodeError) {
        console.error('Node connection failed:', nodeError);
        setAppState(prev => ({ ...prev, address, connected: false }));
        showStatus('info', 'Wallet connected', 'Node connection failed');
      }
    } catch (error) {
      showStatus('error', 'Failed to connect wallet', error instanceof Error ? error.message : String(error));
    }
  }, [showStatus, appState.nodeUrl, appState.sessionKey, buildClient, fetchAssets]);

  const disconnectWallet = useCallback(() => {
    localStorage.removeItem('metamask_connected');
    setWalletClient(null);
    if (appState.client) {
      appState.client.close();
    }
    setAppState(prev => ({
      ...prev,
      address: null,
      connected: false,
      client: null
    }));
    showStatus('info', 'Wallet disconnected');
  }, [appState.client, showStatus]);

  const setSessionKey = useCallback((sk: SessionKeyState) => {
    localStorage.setItem('nitrolite_session_key', JSON.stringify(sk));
    setAppState(prev => ({ ...prev, sessionKey: sk }));
  }, []);

  const activateSessionKey = useCallback(async (sk: SessionKeyState) => {
    if (!walletClient || !appState.address) return;

    localStorage.setItem('nitrolite_session_key', JSON.stringify(sk));

    if (appState.client) {
      await appState.client.close();
    }

    const newClient = await buildClient(walletClient, appState.nodeUrl, sk, appState.address);
    setAppState(prev => ({ ...prev, client: newClient, connected: true, sessionKey: sk }));
    await fetchAssets(newClient);
  }, [walletClient, appState.address, appState.client, appState.nodeUrl, buildClient, fetchAssets]);

  const clearSessionKey = useCallback(async () => {
    if (!walletClient || !appState.address) return;

    localStorage.removeItem('nitrolite_session_key');

    if (appState.client) {
      await appState.client.close();
    }

    const newClient = await buildClient(walletClient, appState.nodeUrl, null, appState.address);
    setAppState(prev => ({ ...prev, client: newClient, connected: true, sessionKey: null }));
    await fetchAssets(newClient);
  }, [walletClient, appState.address, appState.client, appState.nodeUrl, buildClient, fetchAssets]);

  const currentNetwork = NETWORKS.find(n => n.chainId === appState.selectedChainId);

  return (
    <div className="min-h-screen bg-background flex flex-col transition-colors duration-200">
      {/* Header */}
      <header className="border-b border-border bg-card flex-shrink-0 transition-colors duration-200">
        <div className="max-w-5xl mx-auto px-6 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h1 className="text-lg font-semibold tracking-tight uppercase">
                Yellow SDK
              </h1>
              <div className="h-4 w-px bg-border" />
              <span className="text-xs text-muted-foreground uppercase tracking-wider">Demo</span>
            </div>

            <div className="flex items-center gap-2">
              {/* Theme toggle */}
              <button
                onClick={toggleTheme}
                className="p-2 text-muted-foreground hover:text-foreground transition-colors"
                title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {theme === 'dark' ? (
                  <Sun className="h-4 w-4" />
                ) : (
                  <Moon className="h-4 w-4" />
                )}
              </button>

              {/* Network indicator */}
              {currentNetwork && (
                <span className="text-xs text-muted-foreground uppercase tracking-wider px-2 py-1 bg-muted border border-border">
                  {currentNetwork.name}
                </span>
              )}

              {/* Wallet */}
              {autoConnecting ? (
                <div className="flex items-center gap-2 px-4 py-2 bg-muted border border-border">
                  <div className="animate-spin rounded-full h-3 w-3 border-2 border-accent border-t-transparent" />
                  <span className="text-xs uppercase tracking-wider font-medium">Connecting...</span>
                </div>
              ) : !appState.address ? (
                <button
                  onClick={connectWallet}
                  className="flex items-center gap-2 px-4 py-2 bg-accent text-accent-foreground text-xs font-semibold uppercase tracking-wider hover:bg-accent/90 transition-colors"
                >
                  Connect Wallet
                </button>
              ) : (
                <div className="flex items-center gap-2">
                  <div className="flex items-center gap-2 px-3 py-2 border border-border">
                    <div className={`h-1.5 w-1.5 rounded-full ${appState.connected ? 'bg-green-400 animate-pulse' : 'bg-muted-foreground'}`} />
                    <code className="text-xs font-mono">{formatAddress(appState.address)}</code>
                  </div>
                  <button
                    onClick={disconnectWallet}
                    className="px-2 py-2 text-muted-foreground hover:text-foreground transition-colors"
                    title="Disconnect wallet"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Status Bar */}
      {status && <StatusBar status={status} onClose={() => setStatus(null)} />}

      {/* Main Content */}
      <main className="max-w-5xl mx-auto px-6 py-6 w-full flex-1">
        {appState.connected && appState.client && appState.address && walletClient ? (
          <WalletDashboard
            client={appState.client}
            address={appState.address}
            chainId={appState.selectedChainId}
            asset={appState.selectedAsset}
            assets={assets}
            sessionKey={appState.sessionKey}
            walletClient={walletClient}
            showStatus={showStatus}
            onSetSessionKey={setSessionKey}
            onActivateSessionKey={activateSessionKey}
            onClearSessionKey={clearSessionKey}
            onAssetChange={(a) => setAppState(prev => ({ ...prev, selectedAsset: a }))}
          />
        ) : !autoConnecting && !appState.address ? (
          <div className="flex items-center justify-center py-32 animate-fade-in">
            <div className="text-center max-w-md">
              <h2 className="text-3xl font-semibold tracking-tight uppercase mb-2">
                Yellow Network
              </h2>
              <p className="text-sm text-muted-foreground mb-1">SDK Demo</p>
              <p className="text-xs text-muted-foreground mb-8">
                State channel payments &mdash; fast, secure, off-chain
              </p>
              <button
                onClick={connectWallet}
                className="px-8 py-3 bg-accent text-accent-foreground text-sm font-semibold uppercase tracking-wider hover:bg-accent/90 transition-colors"
              >
                Connect Wallet
              </button>
            </div>
          </div>
        ) : null}
      </main>

      {/* Footer */}
      <footer className="border-t border-border py-4 flex-shrink-0 transition-colors duration-200">
        <div className="max-w-5xl mx-auto px-6 flex items-center justify-between text-xs text-muted-foreground">
          <span className="uppercase tracking-wider">Yellow SDK v1.0.0</span>
          <a
            href="https://github.com/layer-3/nitrolite"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-accent transition-colors"
          >
            GitHub Repository
          </a>
        </div>
      </footer>
    </div>
  );
}

declare global {
  interface Window {
    ethereum?: any;
  }
}

export default App;
