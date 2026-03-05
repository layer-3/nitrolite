# Yellow TypeScript SDK

[![npm version](https://img.shields.io/npm/v/@yellow-org/sdk.svg)](https://www.npmjs.com/package/@yellow-org/sdk)
[![License](https://img.shields.io/npm/l/@yellow-org/sdk.svg)](https://github.com/layer-3/nitrolite/blob/main/LICENSE)

TypeScript SDK for Clearnode payment channels providing both high-level and low-level operations in a unified client:
- **State Operations**: `deposit()`, `withdraw()`, `transfer()`, `closeHomeChannel()`, `acknowledge()` - build and co-sign states off-chain
- **Blockchain Settlement**: `checkpoint()` - the single entry point for all on-chain transactions
- **Low-Level Operations**: Direct RPC access for custom flows and advanced use cases
- **Full Feature Parity**: 100% compatibility with Go SDK functionality

> If you are migrating from `@layer-3/nitrolite@v0.5.3`, please consider using the [@yellow-org/sdk-compat](https://www.npmjs.com/package/@yellow-org/sdk-compat) package. It is a translation layer that uses this SDK underneath and maps the familiar v0.5.3 API surfaces to this SDK.

## Method Cheat Sheet

### State Operations (Off-Chain)
```typescript
client.deposit(blockchainId, asset, amount)       // Prepare deposit state
client.withdraw(blockchainId, asset, amount)      // Prepare withdrawal state
client.transfer(recipientWallet, asset, amount)   // Prepare transfer state
client.closeHomeChannel(asset)                    // Prepare finalize state
client.acknowledge(asset)                         // Acknowledge received state
```

### Blockchain Settlement
```typescript
client.checkpoint(asset)                          // Settle latest state on-chain
client.challenge(state)                           // Submit on-chain challenge
client.approveToken(chainId, asset, amount)       // Approve token spending
client.checkTokenAllowance(chainId, token, owner) // Check token allowance
```

### Node Information
```typescript
client.ping()                        // Health check
client.getConfig()                   // Node configuration
client.getBlockchains()              // Supported blockchains
client.getAssets(blockchainId?)      // Supported assets
```

### User Queries
```typescript
client.getBalances(wallet)              // User balances
client.getTransactions(wallet, opts)    // Transaction history
```

### Channel Queries
```typescript
client.getChannels(wallet, options?)               // List all channels
client.getHomeChannel(wallet, asset)               // Home channel info
client.getEscrowChannel(escrowChannelId)           // Escrow channel info
client.getLatestState(wallet, asset, onlySigned)   // Latest state
```

### App Registry
```typescript
client.getApps(opts)                                            // List registered apps
client.registerApp(appID, metadata, approvalNotRequired)         // Register new app
```

### App Sessions
```typescript
client.getAppSessions(opts)                                     // List sessions
client.getAppDefinition(appSessionId)                           // Session definition
client.createAppSession(definition, sessionData, sigs)          // Create session
client.submitAppSessionDeposit(update, sigs, asset, amount)     // Deposit to session
client.submitAppState(update, sigs)                             // Update session
client.rebalanceAppSessions(signedUpdates)                      // Atomic rebalance
```

### App Session Keys
```typescript
client.signSessionKeyState(state)                               // Sign app session key state
client.submitSessionKeyState(state)                             // Register/update app session key
client.getLastKeyStates(userAddress, sessionKey?)               // Get active app session key states
```

### Channel Session Keys
```typescript
client.signChannelSessionKeyState(state)                        // Sign channel session key state
client.submitChannelSessionKeyState(state)                      // Register/update channel session key
client.getLastChannelKeyStates(userAddress, sessionKey?)        // Get active channel session key states
```

### Shared Utilities
```typescript
client.close()                              // Close connection
client.waitForClose()                       // Connection monitor promise
client.signState(state)                     // Sign a state (advanced)
client.getUserAddress()                     // Get signer's address
client.setHomeBlockchain(asset, chainId)    // Set default blockchain for asset
```

## Installation

```bash
npm install @yellow-org/sdk
# or
yarn add @yellow-org/sdk
# or
pnpm add @yellow-org/sdk
```

## Quick Start

### Unified Client (High-Level + Low-Level)

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

async function main() {
  // Create signers from private key
  const { stateSigner, txSigner } = createSigners(
    process.env.PRIVATE_KEY as `0x${string}`
  );

  // Create unified client
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, 'https://polygon-amoy.alchemy.com/v2/KEY')
  );

  try {
    // Step 1: Build and co-sign states off-chain
    const state = await client.deposit(80002n, 'usdc', new Decimal(100));
    console.log('Deposit state version:', state.version);

    // Step 2: Settle on-chain via checkpoint
    const txHash = await client.checkpoint('usdc');
    console.log('On-chain tx:', txHash);

    // Transfer (off-chain only, no checkpoint needed for existing channels)
    const transferState = await client.transfer('0xRecipient...', 'usdc', new Decimal(50));

    // Low-level operations - same client
    const config = await client.getConfig();
    const balances = await client.getBalances(client.getUserAddress());
  } finally {
    await client.close();
  }
}

main().catch(console.error);
```

## Architecture

```
sdk/ts/src/
├── client.ts         # Core client, constructors, high-level operations
├── signers.ts        # EthereumMsgSigner and EthereumRawSigner
├── config.ts         # Configuration options
├── asset_store.ts    # Asset metadata caching
├── utils.ts          # Type transformations
├── core/             # State management, transitions, types
├── rpc/              # WebSocket RPC client
├── blockchain/       # EVM blockchain interactions
└── app/              # App session types and logic
```

## Client API

### Creating a Client

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';

// Step 1: Create signers from private key
const { stateSigner, txSigner } = createSigners('0x1234...');

// Step 2: Create unified client
const client = await Client.create(
  wsURL,
  stateSigner,  // For signing channel states
  txSigner,     // For signing blockchain transactions
  withBlockchainRPC(chainId, rpcURL), // Required for Checkpoint
  withHandshakeTimeout(10000),         // Optional: connection timeout
  withPingInterval(5000)               // Optional: keepalive interval
);

// Step 3: (Optional) Set home blockchain for assets
// Required for Transfer operations that may trigger channel creation
await client.setHomeBlockchain('usdc', 80002n);
```

### Signer Implementations

The SDK provides two signer types matching the Go SDK patterns:

#### EthereumMsgSigner (for channel states)

Signs channel state updates with EIP-191 "Ethereum Signed Message" prefix.

```typescript
import { EthereumMsgSigner } from '@yellow-org/sdk';
import { privateKeyToAccount } from 'viem/accounts';

// From private key
const signer1 = new EthereumMsgSigner('0x...');

// From viem account
const account = privateKeyToAccount('0x...');
const signer2 = new EthereumMsgSigner(account);
```

**When to use**: All off-chain operations (state signatures, transfers)

#### EthereumRawSigner (for blockchain transactions)

Signs raw hashes directly without prefix for on-chain operations.

```typescript
import { EthereumRawSigner } from '@yellow-org/sdk';

const signer = new EthereumRawSigner('0x...');
```

**When to use**: On-chain operations (deposits, withdrawals, channel creation)

#### Helper: createSigners()

```typescript
import { createSigners } from '@yellow-org/sdk';

// Creates both signers at once
const { stateSigner, txSigner } = createSigners('0x...');
const client = await Client.create(wsURL, stateSigner, txSigner);
```

### Configuring Home Blockchain

#### `setHomeBlockchain(asset, blockchainId)`

Sets the default blockchain network for a specific asset. Required for `transfer()` operations that may trigger channel creation.

```typescript
await client.setHomeBlockchain('usdc', 80002n);
```

**Important Notes:**
- This mapping is immutable once set for the client instance
- The asset must be supported on the specified blockchain
- Required before calling `transfer()` on a new channel

### State Operations

All state operations build and co-sign a state off-chain. They return `Promise<core.State>`. Use `checkpoint()` to settle the state on-chain.

#### `deposit(blockchainId, asset, amount): Promise<core.State>`

Prepares a deposit state. Creates a new channel if none exists, otherwise advances the existing state.

```typescript
const state = await client.deposit(80002n, 'usdc', new Decimal(100));
const txHash = await client.checkpoint('usdc'); // settle on-chain
```

**Requirements:**
- Sufficient token balance (checked on-chain during checkpoint)

**Scenarios:**
1. **No channel exists**: Creates new channel with initial deposit
2. **Channel exists**: Advances the existing state with a deposit transition

#### `withdraw(blockchainId, asset, amount): Promise<core.State>`

Prepares a withdrawal state to remove funds from the channel.

```typescript
const state = await client.withdraw(80002n, 'usdc', new Decimal(25));
const txHash = await client.checkpoint('usdc'); // settle on-chain
```

**Requirements:**
- Existing channel with sufficient balance

#### `transfer(recipientWallet, asset, amount): Promise<core.State>`

Prepares an off-chain transfer to another wallet. For existing channels, no checkpoint is needed.

```typescript
const state = await client.transfer(
  '0xRecipient...',   // Recipient address
  'usdc',             // Asset symbol
  new Decimal(50)     // Amount
);
```

**Requirements:**
- Existing channel with sufficient balance OR
- Home blockchain configured via `setHomeBlockchain()` (for new channels)

#### `closeHomeChannel(asset): Promise<core.State>`

Prepares a finalize state to close the user's channel for a specific asset.

```typescript
const state = await client.closeHomeChannel('usdc');
const txHash = await client.checkpoint('usdc'); // close on-chain
```

**Requirements:**
- Existing channel (user must have deposited first)

#### `acknowledge(asset): Promise<core.State>`

Acknowledges a received state (e.g., after receiving a transfer).

```typescript
const state = await client.acknowledge('usdc');
```

**Requirements:**
- Home blockchain configured via `setHomeBlockchain()` when no channel exists

### Blockchain Settlement

#### `checkpoint(asset): Promise<string>`

Settles the latest co-signed state on-chain. This is the single entry point for all blockchain transactions. Based on the transition type and on-chain channel status, it calls the appropriate blockchain method:

- **Channel not on-chain** (status Void): Creates the channel
- **Deposit/Withdrawal on existing channel**: Checkpoints the state
- **Finalize**: Closes the channel

```typescript
const txHash = await client.checkpoint('usdc');
```

**Requirements:**
- Blockchain RPC configured via `withBlockchainRPC()`
- A co-signed state must exist (call `deposit()`, `withdraw()`, etc. first)
- Sufficient gas for the blockchain transaction

#### `challenge(state): Promise<string>`

Submits an on-chain challenge for a channel using a co-signed state. Initiates a dispute period on-chain.

```typescript
const state = await client.getLatestState(wallet, 'usdc', true);
const txHash = await client.challenge(state);
```

**Requirements:**
- State must have both user and node signatures

#### `approveToken(chainId, asset, amount): Promise<string>`

Approves the ChannelHub contract to spend tokens on behalf of the user. Required before depositing ERC-20 tokens.

```typescript
const txHash = await client.approveToken(80002n, 'usdc', new Decimal(1000));
```

#### `checkTokenAllowance(chainId, tokenAddress, owner): Promise<bigint>`

Checks the current token allowance for the ChannelHub contract.

```typescript
const allowance = await client.checkTokenAllowance(80002n, '0xToken...', '0xOwner...');
```

## Low-Level API

All low-level RPC methods are available on the same Client instance.

### Node Information

```typescript
await client.ping();
const config = await client.getConfig();
const blockchains = await client.getBlockchains();
const assets = await client.getAssets(); // or client.getAssets(blockchainId)
```

### User Data

```typescript
const balances = await client.getBalances(wallet);
const { transactions, metadata } = await client.getTransactions(wallet, {
  page: 1,
  pageSize: 50,
});
```

### Channel Queries

```typescript
const { channels, metadata } = await client.getChannels(wallet);
const channel = await client.getHomeChannel(wallet, asset);
const escrow = await client.getEscrowChannel(escrowChannelId);
const state = await client.getLatestState(wallet, asset, onlySigned);
```

**Note:** State submission and channel creation are handled internally by state operations (`deposit()`, `withdraw()`, `transfer()`). On-chain settlement is handled by `checkpoint()`.

### App Registry

```typescript
// List registered applications with optional filtering
const { apps, metadata } = await client.getApps({
  appId: 'my-app',
  ownerWallet: '0x1234...',
  page: 1,
  pageSize: 10,
});

// Register a new application
await client.registerApp('my-app', '{"name": "My App"}', false);
```

### App Sessions (Low-Level)

```typescript
// Query sessions
const { sessions, metadata } = await client.getAppSessions(opts);
const definition = await client.getAppDefinition(appSessionId);

// Create and manage sessions
const { appSessionId, version, status } = await client.createAppSession(
  definition,
  sessionData,
  signatures
);

const nodeSig = await client.submitAppSessionDeposit(
  appUpdate,
  quorumSigs,
  asset,
  depositAmount
);

await client.submitAppState(appUpdate, quorumSigs);

const batchId = await client.rebalanceAppSessions(signedUpdates);
```

### App Session Keys

```typescript
// Sign and submit an app session key state
const sig = await client.signSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  application_ids: ['app1'],
  app_session_ids: [],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: '0x',
});

await client.submitSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  application_ids: ['app1'],
  app_session_ids: [],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: sig,
});

// Query active app session key states
const states = await client.getLastKeyStates('0x1234...');
const filtered = await client.getLastKeyStates('0x1234...', '0xSessionKey...');
```

### Channel Session Keys

```typescript
// Sign and submit a channel session key state
const sig = await client.signChannelSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  assets: ['usdc'],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: '0x',
});

await client.submitChannelSessionKeyState({
  user_address: '0x1234...',
  session_key: '0xabcd...',
  version: '1',
  assets: ['usdc'],
  expires_at: String(Math.floor(Date.now() / 1000) + 86400),
  user_sig: sig,
});

// Query active channel session key states
const states = await client.getLastChannelKeyStates('0x1234...');
const filtered = await client.getLastChannelKeyStates('0x1234...', '0xSessionKey...');
```

## Key Concepts

### State Management

Payment channels use versioned states signed by both user and node. The SDK uses a two-step pattern:

```typescript
// Step 1: Build and co-sign state off-chain
const state = await client.deposit(...);   // Returns core.State
const state = await client.withdraw(...);  // Returns core.State
const state = await client.transfer(...);  // Returns core.State

// Step 2: Settle on-chain (when needed)
const txHash = await client.checkpoint('usdc');
```

**State Flow (Internal):**
1. Get latest state with `getLatestState()`
2. Create next state with `nextState()`
3. Apply transition (deposit, withdraw, transfer, etc.)
4. Sign state with `signState()`
5. Submit to node for co-signing
6. Return co-signed state

On-chain settlement is handled separately by `checkpoint()`.

### Signing

States are signed using ECDSA with EIP-191/EIP-155:

```typescript
// Create signers from private key
const { stateSigner, txSigner } = createSigners('0x...');

// Get address
const address = stateSigner.getAddress();
```

**Signing Process:**
1. State -> ABI Encode (via `packState`)
2. Packed State -> Keccak256 Hash
3. Hash -> ECDSA Sign (via signer)
4. Result: 65-byte signature (R || S || V)

**Two Signer Types:**
- `EthereumMsgSigner`: Signs channel state updates (off-chain signatures) with EIP-191 prefix
- `EthereumRawSigner`: Signs blockchain transactions (on-chain operations) without prefix

### Channel Lifecycle

1. **Void**: No channel exists
2. **Create**: Deposit creates channel on-chain
3. **Open**: Channel active, can deposit/withdraw/transfer
4. **Challenged**: Dispute initiated (advanced)
5. **Closed**: Channel finalized (advanced)

## When to Use State Operations vs Low-Level Operations

### Use State Operations When:
- Building user-facing applications
- Need simple deposit/withdraw/transfer
- Want automatic state management with two-step pattern
- Don't need custom flows

### Use Low-Level Operations When:
- Building infrastructure/tooling
- Implementing custom state transitions
- Need fine-grained control
- Working with app sessions directly

## Error Handling

All errors include context:

```typescript
try {
  const state = await client.deposit(80002n, 'usdc', amount);
  const txHash = await client.checkpoint('usdc');
} catch (error) {
  // State error: "channel not created, deposit first"
  // Checkpoint error: "failed to create channel on blockchain: insufficient balance"
  console.error('Operation failed:', error);
}
```

### Common Errors

| Error Message | Cause | Solution |
|--------------|-------|----------|
| `"channel not created, deposit first"` | Transfer before deposit | Deposit funds first |
| `"home blockchain not set for asset"` | Missing `setHomeBlockchain()` | Call `setHomeBlockchain()` before transfer |
| `"blockchain client not configured"` | Missing `withBlockchainRPC()` | Add `withBlockchainRPC()` configuration |
| `"insufficient balance"` | Not enough funds | Deposit more funds |
| `"failed to sign state"` | Invalid private key or state | Check signer configuration |
| `"no channel exists for asset"` | Checkpoint called without a co-signed state | Call `deposit()`, `withdraw()`, etc. first |
| `"transition type ... does not require a blockchain operation"` | Checkpoint called on unsupported transition | Only checkpoint after deposit, withdraw, close, or acknowledge |

### Custom Error Handler

```typescript
const client = await Client.create(
  wsURL,
  stateSigner,
  txSigner,
  withErrorHandler((error) => {
    console.error('[Connection Error]', error);
    // Custom error handling logic
  })
);
```

## Configuration Options

```typescript
import {
  withBlockchainRPC,
  withHandshakeTimeout,
  withPingInterval,
  withErrorHandler
} from '@yellow-org/sdk';

const client = await Client.create(
  wsURL,
  stateSigner,
  txSigner,
  withBlockchainRPC(chainId, rpcURL),  // Configure blockchain RPC (required for Checkpoint)
  withHandshakeTimeout(10000),          // Connection timeout (ms, default: 5000)
  withPingInterval(5000),               // Keepalive interval (ms, default: 5000)
  withErrorHandler(func)                // Connection error handler
);
```

## Complete Examples

### Example 1: Basic Deposit and Transfer

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

async function basicExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.RPC_URL!)
  );

  try {
    console.log('User:', client.getUserAddress());

    // Set home blockchain
    await client.setHomeBlockchain('usdc', 80002n);

    // Step 1: Build and co-sign deposit state
    const depositState = await client.deposit(80002n, 'usdc', new Decimal(100));
    console.log('Deposit state version:', depositState.version);

    // Step 2: Settle on-chain
    const txHash = await client.checkpoint('usdc');
    console.log('On-chain tx:', txHash);

    // Check balance
    const balances = await client.getBalances(client.getUserAddress());
    console.log('Balances:', balances);

    // Transfer 50 USDC (off-chain, no checkpoint needed)
    const transferState = await client.transfer(
      '0xRecipient...',
      'usdc',
      new Decimal(50)
    );
    console.log('Transfer state version:', transferState.version);
  } finally {
    await client.close();
  }
}

basicExample().catch(console.error);
```

### Example 2: Multi-Chain Operations

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

async function multiChainExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.POLYGON_RPC!), // Polygon Amoy
    withBlockchainRPC(11155111n, process.env.SEPOLIA_RPC!) // Sepolia
  );

  try {
    // Set home blockchains
    await client.setHomeBlockchain('usdc', 80002n);
    await client.setHomeBlockchain('eth', 11155111n);

    // Deposit on different chains (two-step pattern)
    await client.deposit(80002n, 'usdc', new Decimal(100));
    await client.checkpoint('usdc');

    await client.deposit(11155111n, 'eth', new Decimal(0.1));
    await client.checkpoint('eth');

    // Check balances across all chains
    const balances = await client.getBalances(client.getUserAddress());
    balances.forEach(b => console.log(`${b.asset}: ${b.balance}`));
  } finally {
    await client.close();
  }
}

multiChainExample().catch(console.error);
```

### Example 3: Transaction History with Pagination

```typescript
import { Client, createSigners } from '@yellow-org/sdk';

async function queryTransactions() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner
  );

  try {
    const wallet = client.getUserAddress();

    // Get paginated transactions
    const result = await client.getTransactions(wallet, {
      page: 1,
      pageSize: 10,
    });

    console.log(`Total: ${result.metadata.totalCount}`);
    console.log(`Page ${result.metadata.page} of ${result.metadata.pageCount}`);

    result.transactions.forEach((tx, i) => {
      console.log(`${i + 1}. ${tx.txType}: ${tx.amount} ${tx.asset}`);
    });
  } finally {
    await client.close();
  }
}

queryTransactions().catch(console.error);
```

### Example 4: App Session Workflow

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

async function appSessionExample() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);
  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, process.env.RPC_URL!)
  );

  try {
    // Create app session
    const definition = {
      application: 'chess-v1',
      participants: [
        { walletAddress: client.getUserAddress(), signatureWeight: 1 },
        { walletAddress: '0xOpponent...', signatureWeight: 1 },
      ],
      quorum: 2,
      nonce: 1n,
    };

    const { appSessionId } = await client.createAppSession(
      definition,
      '{}',
      ['sig1', 'sig2']
    );
    console.log('Session created:', appSessionId);

    // Deposit to app session
    const appUpdate = {
      appSessionId,
      intent: 1, // Deposit
      version: 1n,
      allocations: [{
        participant: client.getUserAddress(),
        asset: 'usdc',
        amount: new Decimal(50),
      }],
      sessionData: '{}',
    };

    const nodeSig = await client.submitAppSessionDeposit(
      appUpdate,
      ['sig1'],
      'usdc',
      new Decimal(50)
    );
    console.log('Deposit signature:', nodeSig);

    // Query sessions
    const { sessions } = await client.getAppSessions({
      wallet: client.getUserAddress(),
    });
    console.log(`Found ${sessions.length} sessions`);
  } finally {
    await client.close();
  }
}

appSessionExample().catch(console.error);
```

### Example 5: Connection Monitoring

```typescript
import { Client, createSigners, withErrorHandler, withPingInterval } from '@yellow-org/sdk';

async function monitorConnection() {
  const { stateSigner, txSigner } = createSigners(process.env.PRIVATE_KEY!);

  const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withPingInterval(3000),
    withErrorHandler((error) => {
      console.error('Connection error:', error);
    })
  );

  // Monitor connection
  client.waitForClose().then(() => {
    console.log('Connection closed, reconnecting...');
    // Reconnection logic here
  });

  // Perform operations
  const config = await client.getConfig();
  console.log('Connected to:', config.nodeAddress);

  // Keep alive...
  await new Promise(resolve => setTimeout(resolve, 30000));
  await client.close();
}

monitorConnection().catch(console.error);
```

## TypeScript-Specific Notes

### Type Imports

```typescript
import type {
  State,
  Channel,
  Transaction,
  BalanceEntry,
  Asset,
  Blockchain,
  AppSessionInfoV1,
  AppDefinitionV1,
  AppSessionKeyStateV1,
  ChannelSessionKeyStateV1,
  PaginationMetadata,
} from '@yellow-org/sdk';

// App Registry types (from rpc/types)
import type { AppV1, AppInfoV1 } from '@yellow-org/sdk';
```

### BigInt for Chain IDs

```typescript
// Use 'n' suffix for bigint literals
const polygonAmoy = 80002n;
const ethereum = 1n;

await client.deposit(polygonAmoy, 'usdc', amount);
```

### Decimal.js for Amounts

```typescript
import Decimal from 'decimal.js';

const amount1 = new Decimal(100);
const amount2 = new Decimal('123.456');
const amount3 = Decimal.div(1000, 3);

await client.deposit(chainId, 'usdc', amount1);
```

### Viem Integration

```typescript
import { privateKeyToAccount } from 'viem/accounts';
import type { Address } from 'viem';

const account = privateKeyToAccount('0x...');
const stateSigner = new EthereumMsgSigner(account);

const wallet: Address = '0x1234...';
const balances = await client.getBalances(wallet);
```

### Async/Await

```typescript
// All SDK methods are async
// State operations return core.State
const state = await client.deposit(chainId, asset, amount);
// Checkpoint returns a transaction hash
const txHash = await client.checkpoint(asset);

// Or with .then()
client.deposit(chainId, asset, amount)
  .then(state => console.log('Deposit state version:', state.version))
  .catch(error => console.error('Error:', error));
```

## Operation Internals

For understanding how operations work under the hood:

### Deposit Flow (New Channel)
1. Create channel definition
2. Create void state
3. Set home ledger (token, chain)
4. Apply deposit transition
5. Sign state
6. Request channel creation from node (co-sign)
7. Return co-signed state

### Deposit Flow (Existing Channel)
1. Get latest state
2. Create next state
3. Apply deposit transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### Withdraw Flow
1. Get latest state
2. Create next state
3. Apply withdrawal transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### Transfer Flow
1. Get latest state
2. Create next state
3. Apply transfer transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### CloseHomeChannel Flow
1. Get latest state
2. Verify channel exists
3. Create next state
4. Apply finalize transition
5. Sign state
6. Submit to node (co-sign)
7. Return co-signed state

### Checkpoint Flow
1. Get latest signed state (both signatures)
2. Determine blockchain ID from state's home ledger
3. Get on-chain channel status
4. Route based on transition type + status:
   - Void channel -> `blockchainClient.create()`
   - Existing channel -> `blockchainClient.checkpoint()`
   - Finalize -> `blockchainClient.close()`
5. Return transaction hash

## Requirements

- **Node.js**: 20.0.0 or later
- **TypeScript**: 5.3.0 or later (for development)
- **Running Clearnode instance** or access to public node
- **Blockchain RPC endpoint** (for on-chain operations via `checkpoint()`)

## License

Part of the Nitrolite project. See [LICENSE](../../LICENSE) for details.

## Related Projects

- [Nitrolite TS Compat](https://www.npmjs.com/package/@yellow-org/sdk-compat) - Compatibility layer for older TypeScript versions
- [Nitrolite Go SDK](https://github.com/layer-3/nitrolite/tree/stable/sdk/go) - Go implementation with same API
- [Nitrolite Smart Contracts](https://github.com/layer-3/nitrolite/tree/stable/contracts) - On-chain contracts

---

**Built with Nitrolite** - Powering the next generation of scalable blockchain applications.
