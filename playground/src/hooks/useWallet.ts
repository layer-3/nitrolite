import { useCallback, useEffect, useState } from 'react';
import { createWalletClient, custom, type WalletClient, type Address } from 'viem';

declare global {
  interface Window {
    ethereum?: {
      request: (args: { method: string; params?: unknown[] }) => Promise<unknown>;
      on?: (event: string, listener: (...args: unknown[]) => void) => void;
      removeListener?: (event: string, listener: (...args: unknown[]) => void) => void;
    };
  }
}

export interface WalletState {
  address: Address | null;
  chainId: bigint | null;
  walletClient: WalletClient | null;
  error: string | null;
  isConnecting: boolean;
}

export interface UseWalletResult extends WalletState {
  connect: () => Promise<void>;
  disconnect: () => void;
  switchChain: (chainId: bigint) => Promise<void>;
}

export function useWallet(): UseWalletResult {
  const [state, setState] = useState<WalletState>({
    address: null,
    chainId: null,
    walletClient: null,
    error: null,
    isConnecting: false,
  });

  const buildClient = useCallback((address: Address): WalletClient => {
    return createWalletClient({
      account: address,
      transport: custom(window.ethereum!),
    });
  }, []);

  const fetchChainId = useCallback(async (): Promise<bigint | null> => {
    if (!window.ethereum) return null;
    const hex = (await window.ethereum.request({ method: 'eth_chainId' })) as string;
    return BigInt(hex);
  }, []);

  const connect = useCallback(async () => {
    if (!window.ethereum) {
      setState(s => ({ ...s, error: 'MetaMask not detected. Install the extension to continue.' }));
      return;
    }
    setState(s => ({ ...s, isConnecting: true, error: null }));
    try {
      const accounts = (await window.ethereum.request({ method: 'eth_requestAccounts' })) as string[];
      if (!accounts.length) throw new Error('No account returned by wallet');
      const address = accounts[0] as Address;
      const chainId = await fetchChainId();
      setState({
        address,
        chainId,
        walletClient: buildClient(address),
        error: null,
        isConnecting: false,
      });
    } catch (err) {
      setState(s => ({
        ...s,
        isConnecting: false,
        error: err instanceof Error ? err.message : String(err),
      }));
    }
  }, [buildClient, fetchChainId]);

  const disconnect = useCallback(() => {
    setState({
      address: null,
      chainId: null,
      walletClient: null,
      error: null,
      isConnecting: false,
    });
  }, []);

  const switchChain = useCallback(async (chainId: bigint) => {
    if (!window.ethereum) return;
    const hex = `0x${chainId.toString(16)}`;
    try {
      await window.ethereum.request({ method: 'wallet_switchEthereumChain', params: [{ chainId: hex }] });
    } catch (err) {
      setState(s => ({ ...s, error: err instanceof Error ? err.message : String(err) }));
    }
  }, []);

  useEffect(() => {
    if (!window.ethereum) {
      setState(s => ({ ...s, error: 'MetaMask not detected. Install the extension to continue.' }));
      return;
    }

    const onAccountsChanged = (...args: unknown[]) => {
      const accounts = args[0] as string[];
      if (!accounts || accounts.length === 0) {
        disconnect();
      } else {
        const address = accounts[0] as Address;
        setState(s => ({ ...s, address, walletClient: buildClient(address) }));
      }
    };

    const onChainChanged = (...args: unknown[]) => {
      const hex = args[0] as string;
      setState(s => ({ ...s, chainId: BigInt(hex) }));
    };

    window.ethereum.on?.('accountsChanged', onAccountsChanged);
    window.ethereum.on?.('chainChanged', onChainChanged);

    // Probe existing accounts on mount (so a previously connected dapp doesn't require re-click).
    window.ethereum
      .request({ method: 'eth_accounts' })
      .then(async result => {
        const accounts = result as string[];
        if (accounts && accounts.length) {
          const address = accounts[0] as Address;
          const chainId = await fetchChainId();
          setState({
            address,
            chainId,
            walletClient: buildClient(address),
            error: null,
            isConnecting: false,
          });
        }
      })
      .catch(() => {});

    return () => {
      window.ethereum?.removeListener?.('accountsChanged', onAccountsChanged);
      window.ethereum?.removeListener?.('chainChanged', onChainChanged);
    };
  }, [buildClient, disconnect, fetchChainId]);

  return { ...state, connect, disconnect, switchChain };
}
