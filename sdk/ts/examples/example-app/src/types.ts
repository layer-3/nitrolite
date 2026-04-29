import { Client } from '@yellow-org/sdk';

export interface SessionKeyState {
  privateKey: string;   // hex private key
  address: string;      // derived session key address
  metadataHash: string; // from registration (empty if not registered)
  authSig: string;      // from registration (empty if not registered)
  active: boolean;      // whether client currently uses this signer
}

export interface NetworkConfig {
  chainId: string;
  name: string;
  rpcUrl: string;
}

export interface NitronodeConfig {
  name: string;
  url: string;
}

export interface AppState {
  client: Client | null;
  address: string | null;
  connected: boolean;
  nodeUrl: string;
  selectedChainId: string;
  selectedAsset: string;
  sessionKey: SessionKeyState | null;
}

export interface StatusMessage {
  type: 'success' | 'error' | 'info';
  message: string;
  details?: string;
}
