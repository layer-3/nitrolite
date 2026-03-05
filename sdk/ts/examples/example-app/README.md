# Nitrolite SDK Integration Guide

This example app demonstrates how to integrate the `@layer-3/nitrolite` TypeScript SDK into a React application. Use it as a reference for building your own Yellow-Network-powered app.

## Quick Start

```bash
npm install
npm run dev
```

Opens at `http://localhost:3000`. Requires MetaMask.

## Install the SDK

```bash
npm install @layer-3/nitrolite viem decimal.js
```

- **`@layer-3/nitrolite`** — the state channel SDK
- **`viem`** — Ethereum wallet and signing primitives
- **`decimal.js`** — precise decimal arithmetic for token amounts

## Core Concepts

The SDK connects to a **clearnode** (a state channel server) over WebSocket. All operations (deposits, withdrawals, transfers) happen off-chain as signed state updates. You can **checkpoint** at any time to sync state on-chain.

The client needs two signers:

| Signer | Purpose | Signing Method |
|---|---|---|
| **StateSigner** | Signs off-chain state updates | EIP-191 (prefixed message) |
| **TransactionSigner** | Signs on-chain transactions | EIP-712 typed data |

## Integration Steps

### 1. Create Wallet Signers

Adapt your wallet (MetaMask, WalletConnect, etc.) to the SDK's signer interfaces:

```ts
// walletSigners.ts
import type { WalletClient, Address, Hex } from 'viem';

// Signs off-chain state updates (EIP-191)
export class WalletStateSigner {
  constructor(private walletClient: WalletClient) {}

  getAddress(): Address {
    return this.walletClient.account!.address;
  }

  async signMessage(hash: Hex): Promise<Hex> {
    return this.walletClient.signMessage({
      account: this.walletClient.account!,
      message: { raw: hash },
    });
  }
}

// Signs on-chain transactions (EIP-712)
export class WalletTransactionSigner {
  constructor(private walletClient: WalletClient) {}

  getAddress(): Address {
    return this.walletClient.account!.address;
  }

  async signMessage(message: { raw: Hex }): Promise<Hex> {
    return this.walletClient.signTypedData({
      account: this.walletClient.account!,
      domain: { name: 'Nitrolite', version: '1', chainId: 1 },
      types: { Message: [{ name: 'data', type: 'bytes32' }] },
      primaryType: 'Message',
      message: { data: message.raw },
    });
  }
}
```

### 2. Create the Client

```ts
import {
  Client,
  withBlockchainRPC,
  ChannelDefaultSigner,
} from '@layer-3/nitrolite';
import { createWalletClient, custom } from 'viem';
import { mainnet } from 'viem/chains';

// Get wallet client from MetaMask (or any EIP-1193 provider)
const walletClient = createWalletClient({
  account: address,
  chain: mainnet,
  transport: custom(window.ethereum),
});

// Wrap your signers
const stateSigner = new ChannelDefaultSigner(new WalletStateSigner(walletClient));
const txSigner = new WalletTransactionSigner(walletClient);

// Connect
const client = await Client.create(
  'wss://clearnode-v1-rc.yellow.org/ws',
  stateSigner,
  txSigner,
  withBlockchainRPC(11155111n, 'https://ethereum-sepolia-rpc.publicnode.com'),
);
```

`ChannelDefaultSigner` wraps your state signer and prepends a protocol type byte (`0x00`) to signatures. This is required.

### 3. Operations

All amounts use `Decimal` from `decimal.js`:

```ts
import Decimal from 'decimal.js';
```

**Deposit** — fund your channel from on-chain:

```ts
await client.deposit(11155111n, 'usdc', new Decimal('100'));
await client.checkpoint('usdc'); // sync to chain
```

**Withdraw** — move funds back on-chain:

```ts
await client.withdraw(11155111n, 'usdc', new Decimal('50'));
await client.checkpoint('usdc');
```

**Transfer** — send to another address (off-chain, instant):

```ts
await client.transfer('0xRecipient...' as `0x${string}`, 'usdc', new Decimal('25'));
// no checkpoint needed — transfers are off-chain only
```

**Acknowledge** — accept an incoming state update:

```ts
await client.acknowledge('usdc');
```

**Close channel** — finalize and close:

```ts
await client.closeHomeChannel('usdc');
await client.checkpoint('usdc');
```

**Checkpoint** — submit the latest signed state on-chain at any time:

```ts
const txHash = await client.checkpoint('usdc');
```

### 4. Query Data

```ts
// Account balances
const balances = await client.getBalances(address);

// Latest channel state (onlySigned=false includes pending states)
const latestState = await client.getLatestState(address, 'usdc', false);

// Latest fully-signed state (both parties signed)
const signedState = await client.getLatestState(address, 'usdc', true);

// On-chain channel info
const channel = await client.getHomeChannel(address, 'usdc');

// Transaction history
const { transactions } = await client.getTransactions(address, { page: 1, pageSize: 10 });

// Node configuration
const config = await client.getConfig();
```

### 5. Handle Token Allowance

On-chain operations (deposit, checkpoint) may fail if the token allowance is insufficient. Catch the error and approve:

```ts
try {
  await client.deposit(chainId, 'usdc', amount);
} catch (e) {
  if (e.message.toLowerCase().includes('allowance')) {
    await client.approveToken(chainId, 'usdc', new Decimal('1e18'));
    await client.deposit(chainId, 'usdc', amount);
  } else {
    throw e;
  }
}
```

### 6. Detect State Changes

Poll for new states and prompt the user to acknowledge:

```ts
const latestState = await client.getLatestState(address, asset, false);
const signedState = await client.getLatestState(address, asset, true);

const needsAcknowledge =
  !signedState || BigInt(latestState.version) > BigInt(signedState.version);

if (needsAcknowledge) {
  await client.acknowledge(asset);
}
```

### 7. Cleanup

Close the WebSocket connection when the user disconnects:

```ts
await client.close();
```

## Session Keys (Auto-Sign)

Session keys let your app sign state updates automatically without wallet popups on every operation. A temporary key is generated in the browser, registered with the clearnode, and used for a limited time.

### Enable

```ts
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import {
  ChannelSessionKeyStateSigner,
  getChannelSessionKeyAuthMetadataHashV1,
} from '@layer-3/nitrolite';

// 1. Generate a temporary key pair
const privateKey = generatePrivateKey();
const account = privateKeyToAccount(privateKey);

// 2. Determine version (increment if prior keys exist)
let version = 1n;
const existing = await client.getLastChannelKeyStates(address, account.address);
if (existing?.length > 0) version = BigInt(existing[0].version) + 1n;

// 3. Build the session key state
const expiresAt = BigInt(Math.floor(Date.now() / 1000) + 24 * 3600); // 24h
const state = {
  user_address: address,
  session_key: account.address,
  version: version.toString(),
  assets: ['usdc', 'weth'],
  expires_at: expiresAt.toString(),
  user_sig: '',
};

// 4. Sign with the main wallet and submit
state.user_sig = await client.signChannelSessionKeyState(state);
await client.submitChannelSessionKeyState(state);

// 5. Compute the metadata hash
const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
  version,
  ['usdc', 'weth'],
  expiresAt,
);

// 6. Recreate the client with the session key signer
const sessionSigner = new ChannelSessionKeyStateSigner(
  privateKey,
  address,
  metadataHash,
  state.user_sig,
);
const txSigner = new WalletTransactionSigner(walletClient);

const sessionClient = await Client.create(
  'wss://clearnode-v1-rc.yellow.org/ws',
  sessionSigner,
  txSigner,
  withBlockchainRPC(11155111n, 'https://ethereum-sepolia-rpc.publicnode.com'),
);
```

Now `sessionClient` signs off-chain state updates with the session key — no wallet popups.

### Revoke

Submit a new version with empty assets to revoke a session key:

```ts
import { packChannelKeyStateV1 } from '@layer-3/nitrolite';

const existing = await client.getLastChannelKeyStates(address, sessionKeyAddress);
const latest = existing[0];

const revokeState = {
  user_address: address,
  session_key: sessionKeyAddress,
  version: (BigInt(latest.version) + 1n).toString(),
  assets: [],
  expires_at: latest.expires_at,
  user_sig: '',
};

// Sign the revocation with the main wallet (EIP-191)
const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
  BigInt(revokeState.version),
  [],
  BigInt(revokeState.expires_at),
);
revokeState.user_sig = await walletClient.signMessage({
  account: walletClient.account,
  message: { raw: packChannelKeyStateV1(sessionKeyAddress, metadataHash) },
});

await client.submitChannelSessionKeyState(revokeState);
```

Then recreate the client with `ChannelDefaultSigner` to go back to wallet signing.

## Client Options

```ts
import {
  withBlockchainRPC,
  withHandshakeTimeout,
  withPingInterval,
  withErrorHandler,
} from '@layer-3/nitrolite';

const client = await Client.create(
  wsUrl,
  stateSigner,
  txSigner,
  // Map chain IDs to RPC endpoints (required for on-chain operations)
  withBlockchainRPC(11155111n, 'https://ethereum-sepolia-rpc.publicnode.com'),
  withBlockchainRPC(84532n, 'https://base-sepolia-rpc.publicnode.com'),
  // Tune WebSocket behavior
  withHandshakeTimeout(5000),  // default: 5000ms
  withPingInterval(5000),      // default: 5000ms
  // Global error handler
  withErrorHandler((error) => console.error('SDK error:', error)),
);
```

## Project Structure

```
src/
  App.tsx              — wallet connection, client creation, session key lifecycle
  walletSigners.ts     — adapts viem WalletClient to SDK signer interfaces
  types.ts             — app-level TypeScript types
  utils.ts             — formatting helpers (addresses, balances, time)
  components/
    WalletDashboard.tsx — balance display, actions, acknowledge, session keys
    ActionModal.tsx     — deposit/withdraw/transfer/close with progress steps
    StatusBar.tsx       — toast notifications
    ErrorBoundary.tsx   — React error boundary
```

## Resources

- [GitHub Repository](https://github.com/layer-3/nitrolite)
- [Viem Documentation](https://viem.sh)
