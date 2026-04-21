# Migrating from v0.5.3 to Compat Layer

This guide explains how to migrate your Nitrolite dApp from the v0.5.3 SDK to the **compat layer**, a curated migration layer that preserves selected app-facing APIs over the v1.0.0 runtime.

## 1. Why Use the Compat Layer

The v1.0.0 protocol changes wire format, authentication, WebSocket lifecycle, unit handling, and asset resolution. Instead of rewriting every RPC call at once, the compat layer lets supported app-facing paths move over incrementally by swapping imports and replacing the `create-sign-send-parse` pattern with direct client method calls.

## 2. Installation

```bash
npm install @yellow-org/sdk-compat
# Peer dependencies
npm install @yellow-org/sdk viem
```

## 3. Import Swap

| Before (v0.5.3) | After (compat) |
|-----------------|----------------|
| `import { createGetChannelsMessage, parseGetChannelsResponse } from '@layer-3/nitrolite'` | `import { NitroliteClient } from '@yellow-org/sdk-compat'` |
| Types: `AppSession`, `LedgerChannel`, `RPCAppDefinition` | Many app-facing types remain re-exported from `@yellow-org/sdk-compat` |

For **types**, many app-facing imports only need a package-name swap. For **functions**, prefer client methods instead of `create*Message` / `parse*Response`. Some legacy helper imports remain exported only as transitional migration shims.

## 4. The Key Pattern Change

**Before (v0.5.3):** create-sign-send-parse

```typescript
const msg = await createGetChannelsMessage(signer.sign, addr);
const raw = await sendRequest(msg);
const parsed = parseGetChannelsResponse(raw);
const channels = parsed.params.channels;
```

**After (compat):** direct client method

```typescript
const client = await NitroliteClient.create(config);
const channels = await client.getChannels();
```

## 5. What Stays the Same

- **Many app-facing type shapes:** `AppSession`, `LedgerChannel`, `RPCAppDefinition`, `RPCBalance`, `RPCAsset`, etc.
- **Familiar response shapes:** balances, ledger entries, and app sessions remain close to v0.5.3 for the supported app-facing paths.
- **Auth helpers:** `createAuthRequestMessage`, `createAuthVerifyMessage`, `createAuthVerifyMessageWithJWT`, and `createEIP712AuthMessageSigner` remain available for legacy auth flows.

## 6. What Changes

| Concern | v0.5.3 | Compat |
|---------|--------|--------|
| WebSocket | App creates and manages `WebSocket` | Managed internally by the client |
| Signing | App passes `signer.sign` into every message | Internal — client uses `WalletClient` |
| Amounts | Raw `BigInt` everywhere | Compat keeps raw-unit app-facing inputs and handles the v1 bridge internally |
| Contract addresses | Manual config | Fetched from clearnode `get_config` |
| Channel creation | Explicit `createChannel()` | Implicit on first `deposit()` |

## 7. Quick Start Example

```typescript
import { NitroliteClient, WalletStateSigner, blockchainRPCsFromEnv } from '@yellow-org/sdk-compat';

// Create client (replaces new Client(ws, signer))
const client = await NitroliteClient.create({
  wsURL: 'wss://clearnode.example.com/ws',
  walletClient,          // viem WalletClient with account
  chainId: 11155111,    // Sepolia
  blockchainRPCs: blockchainRPCsFromEnv(),
});

// Deposit (creates channel if needed)
await client.deposit(tokenAddress, 11_000_000n);

// Query
const channels = await client.getChannels();
const balances = await client.getBalances();
const sessions = await client.getAppSessionsList();

// Transfer
await client.transfer(recipientAddress, [{ asset: 'usdc', amount: '5.0' }]);

// Cleanup
await client.closeChannel();
await client.close();
```

## 8. Next Steps

- **[On-Chain Changes](./migration-onchain.md)** — Deposits, withdrawals, channel operations, amount handling
- **[Off-Chain Changes](./migration-offchain.md)** — App sessions, transfers, ledger queries, event polling
