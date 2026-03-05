# Migrating from v0.5.3 to Compat Layer

This guide explains how to migrate your Nitrolite dApp from the v0.5.3 SDK to the **compat layer**, which bridges the old API to the v1.0.0 runtime with minimal code changes.

## 1. Why Use the Compat Layer

The v1.0.0 protocol introduces breaking changes across wire format, authentication, WebSocket lifecycle, unit system, asset resolution, and more. A direct migration touches **20+ files** per app with deep, scattered rewrites.

The compat layer centralises this complexity into **~5 file changes** per app. Instead of rewriting every RPC call, you swap imports and replace the `create-sign-send-parse` pattern with direct client method calls.

## 2. Installation

```bash
npm install @layer-3/nitrolite-compat
# Peer dependencies
npm install @layer-3/nitrolite viem
```

## 3. Import Swap

| Before (v0.5.3) | After (compat) |
|-----------------|----------------|
| `import { createGetChannelsMessage, parseGetChannelsResponse } from '@layer-3/nitrolite'` | `import { NitroliteClient } from '@layer-3/nitrolite-compat'` |
| Types: `AppSession`, `LedgerChannel`, `RPCAppDefinition` | Same types — re-exported from `@layer-3/nitrolite-compat` |

For **types**, just change the package name. For **functions**, switch to client methods instead of `create*Message` / `parse*Response`.

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

- **Type shapes:** `AppSession`, `LedgerChannel`, `RPCAppDefinition`, `RPCBalance`, `RPCAsset`, etc.
- **Response formats:** Balances, ledger entries, app sessions — same structure as v0.5.3.
- **Auth helpers:** `createAuthRequestMessage`, `createAuthVerifyMessage`, `createAuthVerifyMessageWithJWT`, and `createEIP712AuthMessageSigner` remain available for legacy auth flows.

## 6. What Changes

| Concern | v0.5.3 | Compat |
|---------|--------|--------|
| WebSocket | App creates and manages `WebSocket` | Managed internally by the client |
| Signing | App passes `signer.sign` into every message | Internal — client uses `WalletClient` |
| Amounts | Raw `BigInt` everywhere | Compat accepts both; conversion handled internally |
| Contract addresses | Manual config | Fetched from clearnode `get_config` |
| Channel creation | Explicit `createChannel()` | Implicit on first `deposit()` |

## 7. Quick Start Example

```typescript
import { NitroliteClient, WalletStateSigner, blockchainRPCsFromEnv } from '@layer-3/nitrolite-compat';

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
