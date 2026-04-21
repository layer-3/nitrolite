# Nitrolite Compat SDK

[![License](https://img.shields.io/npm/l/@yellow-org/sdk-compat.svg)](https://github.com/layer-3/nitrolite/blob/main/LICENSE)

`@yellow-org/sdk-compat` is a migration layer that preserves selected Nitrolite SDK **v0.5.3 app-facing APIs** over the **v1 runtime**.

- Keep supported v0.5.3-style app-facing calls in your code.
- Run them through `@yellow-org/sdk-compat`, backed by `@yellow-org/sdk`.
- Treat it as a migration aid, not a drop-in replacement for the full published v0.5.3 package.

## Compatibility Scope

`@yellow-org/sdk-compat` is intentionally narrower than the published v0.5.3 package surface.

- **Preserved app-facing APIs**: the `NitroliteClient` facade, selected auth helpers, app-session signing helpers, and many app-facing types remain available for supported migration paths.
- **Transitional helper surfaces**: many legacy `create*Message` / `parse*Response` exports remain so imports can keep compiling during migration, but several are transitional shims rather than one-to-one v1 RPC mappings.
- **Unsupported full-package parity**: low-level internals, broad root-export parity, and every legacy helper being runtime-faithful are not promised by this package.

## Why

The v1 protocol changes wire format, authentication, WebSocket lifecycle, unit handling, asset resolution, and more. For apps built around the old surface, a direct migration can require scattered rewrites across transport, signing, and amount-handling paths.

The compat layer centralizes the supported migration paths into one package so app code can move to client-level methods incrementally instead of rewriting every call site at once.

## Build Size

Measured on **February 24, 2026** from the package directory using:

```bash
npm run build:prod
npm pack --dry-run --json
```

| Metric | Size |
|---|---:|
| npm tarball (`size`) | 16,503 bytes (16.1 KB) |
| unpacked package (`unpackedSize`) | 73,292 bytes (71.6 KB) |
| compiled JS in `dist/*.js` | 38,146 bytes (37.3 KB) |
| type declarations in `dist/*.d.ts` | 20,293 bytes (19.8 KB) |
| total emitted runtime + types (`.js` + `.d.ts`) | 58,439 bytes (57.1 KB) |

## Migration Guide

Step-by-step guides for migrating supported app-facing paths from v0.5.3:

- [Overview & Quick Start](./docs/migration-overview.md) — pattern changes, import swaps
- [On-Chain Changes](./docs/migration-onchain.md) — deposits, withdrawals, channels
- [Off-Chain Changes](./docs/migration-offchain.md) — auth, app sessions, transfers, ledger queries

## Installation

```bash
npm install @yellow-org/sdk-compat
# peer dependencies
npm install @yellow-org/sdk viem
```

## Quick Start

### 1. Initialize the client

Replace `new Client(ws, signer)` with `NitroliteClient.create()`:

```typescript
import { NitroliteClient, blockchainRPCsFromEnv } from '@yellow-org/sdk-compat';

const client = await NitroliteClient.create({
  wsURL: 'wss://clearnode.example.com/ws',
  walletClient,          // viem WalletClient with account
  chainId: 11155111,     // Sepolia
  blockchainRPCs: blockchainRPCsFromEnv(),
});
```

### 2. Deposit & create a channel

In v1, channel creation is implicit on deposit — no separate `createChannel()` call needed:

```typescript
const tokenAddress = '0x6E2C4707DA119425DF2C722E2695300154652F56'; // USDC on Sepolia
const amount = 11_000_000n; // 11 USDC in raw units (6 decimals)

await client.deposit(tokenAddress as Address, amount);
```

### 3. Query channels, balances, ledger entries

```typescript
const channels = await client.getChannels();
const balances = await client.getBalances();
const entries  = await client.getLedgerEntries();
const sessions = await client.getAppSessionsList();
const assets   = await client.getAssetsList();
```

### 4. Transfer off-chain

The compat `transfer(destination, allocations)` preserves the v0.5.3-style array-of-allocations signature. Each `TransferAllocation.amount` is a **raw-unit string** (smallest denomination). The compat layer divides by token decimals before delegating to the v1 SDK's `transfer(wallet, asset, Decimal)`:

```typescript
// 5 USDC = 5_000_000 raw units (6 decimals)
await client.transfer(recipientAddress, [
  { asset: 'usdc', amount: '5000000' },
]);

// For direct SDK access with human-readable Decimal:
// await client.innerClient.transfer('0xRecip...', 'usdc', new Decimal(5));
```

### 5. Close & clean up

```typescript
await client.closeChannel();
await client.close();
```

## Method Cheat Sheet

### Channel Operations

| Method | Description |
|---|---|
| `deposit(token, amount)` | Deposit to channel (creates if needed) |
| `depositAndCreateChannel(token, amount)` | Alias for `deposit()` |
| `withdrawal(token, amount)` | Withdraw from channel |
| `closeChannel()` | Close all open channels |
| `resizeChannel({ allocate_amount, token })` | Resize an existing channel |
| `challengeChannel({ state })` | Challenge a channel on-chain |
| `acknowledge(tokenAddress)` | Acknowledge a pending state or create a channel |
| `checkTokenAllowance(chainId, tokenAddress)` | Check ERC-20 allowance for the ChannelHub |

### Queries

| Method | Description |
|---|---|
| `getChannels()` | List all ledger channels (open, closed, etc.) |
| `getBalances(wallet?)` | Get ledger balances |
| `getLedgerEntries(wallet?)` | Get transaction history |
| `getAppSessionsList(wallet?, status?)` | List app sessions (filter by `'open'`/`'closed'`) |
| `getAssetsList()` | List supported assets |
| `getAccountInfo()` | Aggregate balance + channel count |
| `getConfig()` | Node configuration |
| `getBlockchains()` | List supported blockchains |
| `getActionAllowances(wallet?)` | Get gated action allowances for a wallet |
| `getEscrowChannel(escrowChannelId)` | Query an escrow channel by ID |
| `getChannelData(channelId)` | Full channel + state for a specific channel |
| `getLastAppSessionsListError()` | Last `getAppSessionsList()` error message (if any) |

### App Sessions

| Method | Description |
|---|---|
| `createAppSession(definition, allocations, quorumSigs, opts?)` | Create an app session with quorum signatures (optional `ownerSig` via opts) |
| `closeAppSession(appSessionId, allocations, quorumSigs)` | Close an app session with quorum signatures |
| `submitAppState(params)` | Submit state update (operate/deposit/withdraw/close) |
| `getAppDefinition(appSessionId)` | Get the definition for a session |

### App Registry

| Method | Description |
|---|---|
| `getApps(options?)` | List registered applications (filter by appId, owner, pagination) |
| `registerApp(appID, metadata, creationApprovalNotRequired)` | Register a new application |

### App Session Signing Helpers

| Helper | Description |
|---|---|
| `packCreateAppSessionHash(params)` | Deterministic hash for `createAppSession` quorum signing |
| `packSubmitAppStateHash(params)` | Deterministic hash for `submitAppState` quorum signing |
| `toWalletQuorumSignature(signature)` | Prefixes wallet signature to compat app-session quorum format |
| `toSessionKeyQuorumSignature(signature)` | Prefixes app-session key signature (`0xa2`) to compat quorum format |

### Session Key Operations

| Method | Description |
|---|---|
| `signChannelSessionKeyState(state)` | Sign a channel session-key state payload |
| `submitChannelSessionKeyState(state)` | Register/submit a channel session-key state |
| `getLastChannelKeyStates(userAddress, sessionKey?)` | Fetch channel session-key states for wallet/key |
| `signSessionKeyState(state)` | Sign an app-session key state payload |
| `submitSessionKeyState(state)` | Register/submit an app-session key state |
| `getLastKeyStates(userAddress, sessionKey?)` | Fetch app-session key states for wallet/key |

### Transfers

| Method | Description |
|---|---|
| `transfer(destination, allocations)` | Off-chain transfer to another participant |

### Asset Resolution

| Method | Description |
|---|---|
| `resolveToken(tokenAddress)` | Look up asset info by token address |
| `resolveAsset(symbol)` | Look up asset info by symbol name |
| `resolveAssetDisplay(tokenAddress, chainId?)` | Get display-friendly symbol + decimals |
| `getTokenDecimals(tokenAddress)` | Get decimals for a token |
| `formatAmount(tokenAddress, rawAmount)` | Convert raw bigint → human-readable string |
| `parseAmount(tokenAddress, humanAmount)` | Convert human-readable string → raw bigint |
| `findOpenChannel(tokenAddress, chainId?)` | Find an open channel for a given token |

### Security Token Locking

| Method | Description |
|---|---|
| `lockSecurityTokens(targetWallet, chainId, amount)` | Lock tokens into the Locking contract for a target address |
| `initiateSecurityTokensWithdrawal(chainId)` | Start the unlock process for locked tokens |
| `cancelSecurityTokensWithdrawal(chainId)` | Re-lock tokens, cancelling a pending unlock |
| `withdrawSecurityTokens(chainId, destination)` | Withdraw unlocked tokens to a destination address |
| `approveSecurityToken(chainId, amount)` | Approve the Locking contract to spend tokens |
| `getLockedBalance(chainId, wallet?)` | Query locked balance (returns raw bigint) |

### Lifecycle

| Method | Description |
|---|---|
| `ping()` | Health check |
| `close()` | Close the WebSocket connection |
| `waitForClose()` | Returns a promise that resolves when the connection is closed |
| `refreshAssets()` | Re-fetch the asset map from the clearnode |

### Accessing the v1 SDK Directly

The underlying v1 `Client` is exposed for advanced use cases not covered by the compat surface:

```typescript
const v1Client = client.innerClient;
await v1Client.getHomeChannel(wallet, 'usdc');
```

## Configuration

### `NitroliteClientConfig`

```typescript
interface NitroliteClientConfig {
  wsURL: string;                           // Clearnode WebSocket URL
  walletClient: WalletClient;              // viem WalletClient with account
  chainId: number;                         // Chain ID (e.g. 11155111 for Sepolia)
  blockchainRPCs?: Record<number, string>; // Optional chain ID → RPC URL map
  channelSessionKeySigner?: {
    sessionKeyPrivateKey: Hex;
    walletAddress: Address;
    metadataHash: Hex;
    authSig: Hex;
  };
  addresses?: ContractAddresses;           // Deprecated — ignored, addresses come from get_config
  challengeDuration?: bigint;              // Deprecated — ignored
}
```

### Environment Variables

`blockchainRPCsFromEnv()` reads `NEXT_PUBLIC_BLOCKCHAIN_RPCS`:

```text
NEXT_PUBLIC_BLOCKCHAIN_RPCS=11155111:https://rpc.sepolia.io,1:https://mainnet.infura.io/v3/KEY
```

## Signers

### `WalletStateSigner`

A v0.5.3-compatible signer class that wraps a `WalletClient`. Actual state signing in v1 is handled internally by `ChannelDefaultSigner`; this class exists so existing store types compile:

```typescript
import { WalletStateSigner } from '@yellow-org/sdk-compat';

const signer = new WalletStateSigner(walletClient);
```

### `createECDSAMessageSigner`

Creates a `MessageSigner` function from a private key, compatible with the v0.5.3 signing pattern:

```typescript
import { createECDSAMessageSigner } from '@yellow-org/sdk-compat';

const sign = createECDSAMessageSigner(privateKey);
const signature = await sign(payload);
```

## Error Handling

The compat layer provides typed error classes for common failure modes:

| Error Class | Code | Description |
|---|---|---|
| `CompatError` | *(varies)* | Base class for all compat errors |
| `AllowanceError` | `ALLOWANCE_INSUFFICIENT` | Token approval needed |
| `UserRejectedError` | `USER_REJECTED` | User cancelled in wallet |
| `InsufficientFundsError` | `INSUFFICIENT_FUNDS` | Not enough balance |
| `NotInitializedError` | `NOT_INITIALIZED` | Client not connected |
| `OngoingStateTransitionError` | `ONGOING_STATE_TRANSITION` | Previous action still finalizing |

### `getUserFacingMessage(error)`

Returns a human-friendly string suitable for UI display:

```typescript
import { getUserFacingMessage, AllowanceError } from '@yellow-org/sdk-compat';

try {
  await client.deposit(token, amount);
} catch (err) {
  showToast(getUserFacingMessage(err));
  // → "Transaction was rejected. Please approve the transaction in your wallet to continue."
}
```

### `NitroliteClient.classifyError(error)`

Converts raw SDK/wallet errors into the appropriate typed error:

```typescript
try {
  await client.deposit(token, amount);
} catch (err) {
  const typed = NitroliteClient.classifyError(err);
  if (typed instanceof AllowanceError) {
    // prompt user to approve
  }
}
```

## Event Polling

v0.5.3 used server-push WebSocket events. v1 uses a polling model. The `EventPoller` bridges this gap:

```typescript
import { EventPoller } from '@yellow-org/sdk-compat';

const poller = new EventPoller(client, {
  onChannelUpdate: (channels) => updateUI(channels),
  onBalanceUpdate: (balances) => updateBalances(balances),
  onAssetsUpdate:  (assets)   => updateAssets(assets),
  onError:         (err)      => console.error(err),
}, 5000); // poll every 5 seconds

poller.start();

// Later:
poller.stop();
poller.setInterval(10000); // change interval
```

## Security Token Locking

Lock tokens into the on-chain Locking contract to provide security deposits. The locking token and its decimals are resolved from the contract at runtime — `blockchainRPCs` must be configured for the target chain or these methods will throw.

```typescript
const chainId = 11155111; // Sepolia
// Yellow token on Sepolia has 18 decimals; 100 YELLOW = 100 * 10^18
const amount = 100_000_000_000_000_000_000n;

// Approve the Locking contract to spend tokens
await client.approveSecurityToken(chainId, amount);

// Lock tokens for a target address
await client.lockSecurityTokens(targetWallet, chainId, amount);

// Query locked balance (returns raw bigint in token's smallest unit)
const locked = await client.getLockedBalance(chainId);

// Initiate unlock (starts the unlock period)
await client.initiateSecurityTokensWithdrawal(chainId);

// Cancel unlock (re-lock tokens)
await client.cancelSecurityTokensWithdrawal(chainId);

// After unlock period elapses, withdraw to a destination
await client.withdrawSecurityTokens(chainId, destinationWallet);
```

### Amount conventions

The compat layer accepts raw amounts (smallest token unit) and converts to human-readable `Decimal` before delegating to the v1 SDK.

| Method group | Input type | Example: 100 tokens (18 decimals) |
|---|---|---|
| `deposit`, `withdrawal`, `lockSecurityTokens`, `approveSecurityToken`, `getLockedBalance` | Raw `bigint` | `100_000_000_000_000_000_000n` |
| `transfer` | Raw string via `TransferAllocation.amount` | `'100000000000000000000'` |

> For direct access to the v1 SDK's human-readable `Decimal` API, use `client.innerClient`.

## RPC Stubs

The following functions remain exported primarily so legacy `create*Message` / `parse*Response` imports can keep compiling while an app migrates.
Many `create*` helpers are transitional shims rather than protocol-backed one-to-one v1 RPC mappings, and `parse*` helpers only do lightweight normalization of known response shapes.
Prefer `NitroliteClient` methods directly for new integrations:

```typescript
// Transitional compat exports:
createGetChannelsMessage, parseGetChannelsResponse,
createGetLedgerBalancesMessage, parseGetLedgerBalancesResponse,
parseGetLedgerEntriesResponse, parseGetAppSessionsResponse,
createTransferMessage, createAppSessionMessage, parseCreateAppSessionResponse,
createCloseAppSessionMessage, parseCloseAppSessionResponse,
createCreateChannelMessage, parseCreateChannelResponse,
createCloseChannelMessage, parseCloseChannelResponse,
createResizeChannelMessage, parseResizeChannelResponse,
createPingMessage,
convertRPCToClientChannel, convertRPCToClientState,
parseAnyRPCResponse, NitroliteRPC
```

## Auth Helpers

Compat exports auth helpers for apps still using the v0.5.3 auth request/verify flow:

```typescript
createAuthRequestMessage(params)            // builds auth_request RPC message
createAuthVerifyMessage(signer, response)   // signs and builds auth_verify RPC message
createAuthVerifyMessageWithJWT(jwt)         // builds JWT-based auth_verify RPC message
createEIP712AuthMessageSigner(wallet, ...)  // creates EIP-712 signer for auth_verify challenge
```

## Types Reference

All legacy compat types are re-exported from `@yellow-org/sdk-compat`:

### Enums

- `RPCMethod` — RPC method names (`Ping`, `GetConfig`, `GetChannels`, etc.)
- `RPCChannelStatus` — Channel status values (`Open`, `Closed`, `Resizing`, `Challenged`)

### Wire Types

- `MessageSigner` — `(payload: Uint8Array) => Promise<string>`
- `NitroliteRPCMessage` — `{ req: [number, string, any, number]; sig: string }`
- `RPCResponse` — `{ requestId, method, params }`
- `RPCBalance` — `{ asset, amount }`
- `RPCAsset` — `{ token, chainId, symbol, decimals }`
- `RPCChannelUpdate` — Full channel update payload
- `RPCLedgerEntry` — Ledger transaction entry
- `AccountID` — String alias for account identifiers

### Channel Operation Types

- `ContractAddresses` — `{ custody, adjudicator }`
- `Allocation` — `{ destination, token, amount }`
- `FinalState` — Final channel state with signatures
- `ChannelData` — `{ lastValidState, stateData }`
- `CreateChannelResponseParams`, `CloseChannelResponseParams`
- `ResizeChannelRequestParams`
- `TransferAllocation` — `{ asset, amount }`

### App Session Types

- `RPCAppDefinition` — `{ application, protocol, participants, weights, quorum, challenge, nonce }`
- `RPCAppSessionAllocation` — `{ participant, asset, amount }`
- `CloseAppSessionRequestParams`

### State Channel Primitives

- `Channel` — Channel metadata (id, participants, adjudicator, challenge, nonce, version)
- `State` — Channel state (channelId, version, data, allocations)
- `AppLogic<T>` — Interface for custom app logic implementations

### Clearnode Response Types

- `AccountInfo` — `{ balances: LedgerBalance[], channelCount: bigint }`
- `LedgerChannel` — Full ledger channel record (id, participant, status, token, amount, chain_id, etc.)
- `LedgerBalance` — `{ asset, amount }`
- `LedgerEntry` — Ledger entry with credit/debit
- `AppSession` — App session record
- `ClearNodeAsset` — `{ token, chainId, symbol, decimals }`

## Advanced Configuration

### `buildClientOptions`

Converts a `CompatClientConfig` into v1 `Option[]` values passed to `Client.create()`. Useful if you need to customise the underlying SDK client beyond what `NitroliteClient.create()` exposes:

```typescript
import { buildClientOptions, type CompatClientConfig } from '@yellow-org/sdk-compat';

const opts = buildClientOptions({
  wsURL: 'wss://clearnode.example.com/ws',
  blockchainRPCs: { 11155111: 'https://rpc.sepolia.io' },
});
```

## Next.js Integration Notes

When using the compat package in a Next.js app with Turbopack:

1. **Add to `transpilePackages`** in `next.config.ts`:

```typescript
const nextConfig = {
  transpilePackages: ['@yellow-org/sdk', '@yellow-org/sdk-compat'],
};
```

2. The package declares `"sideEffects": false` in its `package.json`, enabling tree-shaking of unused exports.

## Peer Dependencies

| Package | Version |
|---|---|
| `@yellow-org/sdk` | `>=1.2.0` |
| `viem` | `^2.0.0` |

## License

MIT
