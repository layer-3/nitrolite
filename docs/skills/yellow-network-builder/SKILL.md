---
name: yellow-network-builder
description: |
  Getting-started guide for building against the current Nitrolite v1 protocol and @yellow-org/sdk. Covers the repo-local sources of truth, architecture layers, ChannelHub on-chain entrypoint, v1 RPC namespaces, setup commands, and when to use the TypeScript SDK. Use when starting a new Nitrolite integration or checking whether documentation matches the v1 monorepo surface.
version: 5.0.0
sdk_version: "@yellow-org/sdk@^1.2.0"
network: mainnet
last_verified: 2026-05-06
user-invocable: true
source_urls:
  - https://github.com/layer-3/nitrolite/blob/main/docs/api.yaml
  - https://github.com/layer-3/nitrolite/blob/main/sdk/ts/src/rpc/methods.ts
  - https://github.com/layer-3/nitrolite/blob/main/contracts/src/ChannelHub.sol
---

# Nitrolite Builder Guide

Nitrolite is a state channel framework for Ethereum and EVM-compatible chains. It lets applications move most interaction off-chain while preserving an on-chain recovery path through signed states and the `ChannelHub` contract.

## Sources of Truth

For v1 integrations, verify behavior against the repository itself:

- `docs/api.yaml` — canonical v1 RPC methods, request schemas, response schemas, and protocol types.
- `sdk/ts/src/rpc/methods.ts` — TypeScript constants for v1 method names.
- `sdk/ts/src/client.ts` — high-level `Client` API exposed by `@yellow-org/sdk`.
- `contracts/src/ChannelHub.sol` — v1 on-chain entrypoint for channel settlement and challenges.
- `contracts/src/interfaces/Types.sol` — on-chain structs, statuses, intents, and channel definitions.

Do not use legacy flat RPC method names as v1 references. They belong to older compatibility surfaces, not the canonical v1 API.

## Architecture

| Layer | Main components | Use for |
|---|---|---|
| Application | Your app and `@yellow-org/sdk` client | Product logic and UX |
| Off-chain | Clearnode WebSocket JSON-RPC | Fast state negotiation and reads |
| On-chain | `ChannelHub`, engines, validators | Deposits, checkpoints, challenges, closes |

The normal app path is: create a `Client`, produce off-chain states with high-level SDK methods, and submit on-chain checkpoints or challenges only when settlement or recovery is needed.

## Current TypeScript SDK

Install the active SDK:

```bash
npm install @yellow-org/sdk viem decimal.js
```

Create a client with the async factory:

```ts
import Decimal from 'decimal.js';
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(
  process.env.PRIVATE_KEY as `0x${string}`,
);

const client = await Client.create(
  'wss://clearnode-sandbox.yellow.org/v1/ws',
  stateSigner,
  txSigner,
  withBlockchainRPC(80002n, process.env.POLYGON_AMOY_RPC!),
);

await client.deposit(80002n, 'usdc', new Decimal('100'));
await client.checkpoint('usdc');
await client.transfer('0xRecipient...', 'usdc', new Decimal('50'));
```

Amounts are `Decimal` human amounts in the v1 SDK. The high-level SDK handles state construction and signing; call `checkpoint(asset)` when the latest off-chain state needs to be submitted on-chain.

## v1 RPC Namespaces

The v1 wire API uses dotted namespaces. Confirm exact method names in `sdk/ts/src/rpc/methods.ts` and schemas in `docs/api.yaml`.

| Namespace | Examples |
|---|---|
| `channels.v1.*` | `channels.v1.get_home_channel`, `channels.v1.request_creation`, `channels.v1.submit_state` |
| `app_sessions.v1.*` | `app_sessions.v1.create_app_session`, `app_sessions.v1.submit_app_state`, `app_sessions.v1.get_app_sessions` |
| `apps.v1.*` | `apps.v1.get_apps`, `apps.v1.submit_app_version` |
| `user.v1.*` | `user.v1.get_balances`, `user.v1.get_transactions`, `user.v1.get_action_allowances` |
| `node.v1.*` | `node.v1.ping`, `node.v1.get_config`, `node.v1.get_assets` |

Prefer the high-level `Client` methods for application code. Use low-level RPC names when implementing SDK internals, debugging wire behavior, or cross-checking generated API docs.

## Repo Map

| Path | Purpose |
|---|---|
| `sdk/ts` | Active TypeScript SDK package, published as `@yellow-org/sdk` |
| `sdk/ts-compat` | Compatibility package for older API shapes; not the v1 source of truth |
| `contracts` | Solidity contracts, including `ChannelHub` and supporting engines |
| `clearnode` | Off-chain broker/node implementation |
| `pkg` | Shared Go protocol, RPC, signing, and app packages |
| `docs` | Protocol and API documentation |

## Local Development

Common commands from the repository root:

```bash
cd sdk/ts && npm install
cd sdk/ts && npm test
cd sdk/ts && npm run build
cd contracts && forge build
cd contracts && forge test
go test ./...
```

## Related

- `yellow-sdk-v1` — detailed quick reference for `@yellow-org/sdk` v1.
