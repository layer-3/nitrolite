---
name: yellow-sdk-v1
description: |
  The `@yellow-org/sdk` v1.x TypeScript SDK — the active upstream SDK for Yellow Network / Clearnode payment channels. Replaces the frozen `@erc7824/nitrolite@0.5.3`. Built around a single unified `Client` class (created via `Client.create(wsURL, stateSigner, txSigner, ...opts)`) that owns both the WebSocket and the blockchain RPC, plus two signer classes (`EthereumMsgSigner` for state, `EthereumRawSigner` for tx) and a `createSigners(privateKey)` convenience helper. Use when: starting a new integration against mainnet, migrating a v0.5.3 dApp to v1 (directly or via `@yellow-org/sdk-compat`), or deciding which of the v1 method namespaces (`channels.v1.*`, `app_sessions.v1.*`, `apps.v1.*`, `user.v1.*`, `node.v1.*`) covers your call.
version: 2.0.0
sdk_version: "@yellow-org/sdk@^1.2.0"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://www.npmjs.com/package/@yellow-org/sdk
  - https://github.com/layer-3/nitrolite/blob/main/docs/api.yaml
  - https://docs.yellow.org/docs/0.5.x/build/quick-start
  - https://docs.yellow.org/docs/0.5.x/guides/migration-guide
---

# Yellow SDK v1 (`@yellow-org/sdk`)

**Status**: Active, current mainnet SDK · **Version**: 1.2.0
**Supersedes**: `@erc7824/nitrolite@0.5.3` (frozen)
**Runtime**: Node.js ≥ 20; deps: `viem`, `decimal.js`, `zod`, `abitype`

## Install

```bash
npm install @yellow-org/sdk viem
```

## Why v1 (vs v0.5.3)

v0.5.3 was bring-your-own-socket: you dialled the WebSocket, signed every
`createXxxMessage`, parsed every response. v1 collapses this into a single
`Client` that owns the socket, the state-signer, and the on-chain tx-signer.
Roughly 14 dimensions of wire format, auth, units, and asset resolution
changed — direct migration touches many files, which is why
`@yellow-org/sdk-compat` exists (many helpers are **noop stubs**, so don't
assume compat is a drop-in).

## Quickstart — `Client.create()`

```typescript
import { Client, createSigners, withBlockchainRPC, withHandshakeTimeout }
  from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(
  process.env.PRIVATE_KEY as `0x${string}`,
);

const client = await Client.create(
  'wss://clearnode.yellow.com/ws',
  stateSigner,        // signs channel states (EIP-191 prefixed)
  txSigner,           // signs on-chain transactions
  withBlockchainRPC(1n, process.env.MAINNET_RPC!),
  withHandshakeTimeout(10_000),
);
```

There is **no `new Client(...)`** — only the async factory. `Client.create`
does the WebSocket handshake and auth in one call.

## v1 method namespaces (dotted, versioned)

All RPC methods in v1 use dotted namespaces:

| Namespace | Example methods |
|---|---|
| `channels.v1.*` | `get_home_channel`, `request_creation`, `submit_state` |
| `app_sessions.v1.*` | `submit_deposit_state`, `submit_session_key_state` |
| `apps.v1.*` | `get_apps`, `submit_app_version` |
| `user.v1.*` | `get_balances`, `get_transactions`, `get_action_allowances` |
| `node.v1.*` | `ping`, `get_config`, `get_assets` |

Don't mix v0 (flat `get_balances`) and v1 (`user.v1.get_balances`) method
names in the same client — the wire format is incompatible.

## Common operations

v1 takes **positional args** (not object params) and amounts are `Decimal`
from `decimal.js`:

```ts
import Decimal from 'decimal.js';

// Deposits → unified balance — (blockchainId, asset, amount)
await client.deposit(137n, '0xUSDCAddress', new Decimal('100'));

// Transfers — (recipientWallet, asset, amount)
await client.transfer('0xRecipient', 'usdc', new Decimal('50'));

// App session — (definition, sessionData, quorumSigs, opts?)
const { appSessionId } = await client.createAppSession(definition, sessionData, [sig]);

// Channel lifecycle — single-arg forms
await client.closeHomeChannel('usdc');
await client.challenge(state);
await client.checkpoint('usdc');

// Queries
await client.ping();
const balances = await client.getBalances(ownerAddress);
```

See `reference.md` for the full method catalogue (Client + signers + options
+ every v1 namespace).

## Signers

- **`EthereumMsgSigner`** — signs channel states (EIP-191-prefixed personal sign)
- **`EthereumRawSigner`** — signs raw on-chain transaction hashes

`createSigners(privateKey)` returns both. Don't share keys across signers
other than via this helper — v1 expects the two to agree on the wallet
identity.

## App-session prefix bytes (v1-specific)

Quorum signatures in v1 app sessions carry a 1-byte type prefix:

- `0xA1` = `AppSessionWalletSignerV1`
- `0xA2` = `AppSessionKeySignerV1`

Raw 65-byte sigs without the prefix are rejected. v0.5.3 sigs have no
prefix → v0.5.3 and v1 quorum payloads are wire-incompatible.

## Notifications → polling

v1 **removes** push notifications (`bu`/`cu`/`tr`/`asu`). Poll `get*`
methods instead, or use `@yellow-org/sdk-compat`'s `EventPoller` shim that
re-synthesizes the v0.5.3 push shape.

## Navigation Guide

### When to read supporting files

**reference.md** — read when you need:

- Full `Client.create()` signature + every option factory
- Complete method catalogue by v1 namespace (channels / app_sessions / apps / user / node)
- Module layout (`./client`, `./signers`, `./config`, `./asset_store`, `./utils`, `./core`, `./app`, `./blockchain`, `./rpc`)
- Signer class internals (`StateSigner` / `TransactionSigner` interfaces)
- Core enums (`ChannelStatus`, `TransitionType`, `INTENT_*` constants)
- `@yellow-org/sdk-compat` behaviour table (wired vs noop helpers)
- Configuration options (handshake timeout, error handler, per-chain RPC)

## Related

- `yellow-nitro-rpc` — wire format v1 hides
- `yellow-clearnode-auth` — the auth Client.create runs internally
- `yellow-state-channels` — on-chain security layer
- `yellow-sdk-api` — legacy v0.5.3 SDK (reference only)
