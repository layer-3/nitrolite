---
name: yellow-sdk-v1
description: |
  Reference for the active @yellow-org/sdk v1 TypeScript SDK in this monorepo. Covers Client.create, signer setup, option factories, high-level channel/app-session/user/node methods, and the canonical v1 RPC namespaces from docs/api.yaml and sdk/ts/src/rpc/methods.ts. Use when building new integrations or verifying SDK usage against the current v1 API.
version: 3.0.0
sdk_version: "@yellow-org/sdk@^1.2.0"
network: mainnet
last_verified: 2026-05-06
user-invocable: true
source_urls:
  - https://www.npmjs.com/package/@yellow-org/sdk
  - https://github.com/layer-3/nitrolite/blob/main/docs/api.yaml
  - https://github.com/layer-3/nitrolite/blob/main/sdk/ts/src/client.ts
  - https://github.com/layer-3/nitrolite/blob/main/sdk/ts/src/rpc/methods.ts
---

# Yellow SDK v1 (`@yellow-org/sdk`)

`@yellow-org/sdk` is the active TypeScript SDK for Nitrolite v1. It exposes a single async `Client.create(...)` factory, signer helpers, high-level channel operations, app-session operations, and typed access to the v1 RPC namespace.

## Install

```bash
npm install @yellow-org/sdk viem decimal.js
```

## Create a Client

```ts
import { Client, createSigners, withBlockchainRPC, withHandshakeTimeout } from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(
  process.env.PRIVATE_KEY as `0x${string}`,
);

const client = await Client.create(
  'wss://clearnode-sandbox.yellow.org/v1/ws',
  stateSigner,
  txSigner,
  withBlockchainRPC(80002n, process.env.POLYGON_AMOY_RPC!),
  withHandshakeTimeout(10_000),
);
```

There is no public `new Client(...)` constructor. `Client.create` establishes the WebSocket connection and returns a ready client.

## Signers

`createSigners(privateKey)` returns both signer roles expected by the v1 client:

- `stateSigner` — signs channel and app-session state messages.
- `txSigner` — signs and sends on-chain transactions.

The package also exports signer classes such as `EthereumMsgSigner`, `EthereumRawSigner`, `ChannelDefaultSigner`, `ChannelSessionKeyStateSigner`, `AppSessionWalletSignerV1`, and `AppSessionKeySignerV1` for advanced flows.

## Option Factories

The package root exports these client option factories from `sdk/ts/src/config.ts`:

- `withHandshakeTimeout(ms: number)`
- `withErrorHandler(handler: (error: Error) => void)`
- `withBlockchainRPC(chainId: bigint, rpcUrl: string)`
- `withApplicationID(applicationId: string)`

Use `withBlockchainRPC` before calling methods that need chain access, such as deposits, withdrawals, approvals, checkpoints, and challenges.

## Common Operations

The v1 `Client` takes positional arguments. Amounts are `Decimal` human amounts, not raw smallest-unit integers.

```ts
import Decimal from 'decimal.js';

await client.deposit(80002n, 'usdc', new Decimal('100'));
await client.withdraw(80002n, 'usdc', new Decimal('25'));
await client.transfer('0xRecipient...', 'usdc', new Decimal('50'));
await client.acknowledge('usdc');
await client.closeHomeChannel('usdc');
await client.checkpoint('usdc');
```

`checkpoint(asset)` is the on-chain settlement entrypoint used after off-chain state changes when the state needs to be submitted to `ChannelHub`.

## Query Methods

High-level query methods include:

- `ping()`
- `getConfig()`
- `getBlockchains()`
- `getAssets(blockchainId?)`
- `getBalances(wallet)`
- `getTransactions(wallet, options?)`
- `getActionAllowances(wallet)`
- `getChannels(wallet, options?)`
- `getHomeChannel(wallet, asset)`
- `getEscrowChannel(escrowChannelId)`
- `getLatestState(wallet, asset, onlySigned)`

## App Sessions

High-level app-session methods include:

- `getAppSessions({ appSessionId?, wallet?, status?, page?, pageSize? })`
- `getAppDefinition(appSessionId)`
- `createAppSession(definition, sessionData, quorumSigs, opts?)`
- `submitAppSessionDeposit(update, quorumSigs, asset, depositAmount)`
- `submitAppState(update, quorumSigs)`
- `rebalanceAppSessions(signedUpdates)`

There is no dedicated v1 close-session RPC method. Closing an app session is represented as an app-state submission with the close intent, per the `app_state_update.intent` schema in `docs/api.yaml`.

## v1 RPC Namespaces

The canonical v1 RPC method names are in `sdk/ts/src/rpc/methods.ts` and schemas are in `docs/api.yaml`.

| Namespace | Method constants |
|---|---|
| `channels.v1.*` | `ChannelsV1GetHomeChannelMethod`, `ChannelsV1RequestCreationMethod`, `ChannelsV1SubmitStateMethod`, plus channel query/session-key methods |
| `app_sessions.v1.*` | `AppSessionsV1CreateAppSessionMethod`, `AppSessionsV1SubmitAppStateMethod`, `AppSessionsV1GetAppSessionsMethod`, plus deposit/rebalance/session-key methods |
| `apps.v1.*` | `AppsV1GetAppsMethod`, `AppsV1SubmitAppVersionMethod` |
| `user.v1.*` | `UserV1GetBalancesMethod`, `UserV1GetTransactionsMethod`, `UserV1GetActionAllowancesMethod` |
| `node.v1.*` | `NodeV1PingMethod`, `NodeV1GetConfigMethod`, `NodeV1GetAssetsMethod` |

Do not mix these with legacy flat RPC names in new integrations.

## On-Chain Boundary

The v1 on-chain entrypoint in this repository is `contracts/src/ChannelHub.sol`. SDK operations produce signed states; `checkpoint(asset)` and `challenge(state)` are the high-level methods that cross the on-chain boundary.

## Compatibility Note

`sdk/ts-compat` exists for older application code that still uses legacy API shapes. Treat it as a migration aid, not as the v1 protocol source of truth.
