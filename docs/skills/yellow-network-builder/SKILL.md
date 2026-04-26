---
name: yellow-network-builder
description: |
  Getting-started guide for Yellow Network on mainnet. Covers what Yellow solves, three-layer architecture (on-chain Custody + off-chain Nitro RPC + P2P YNP), $YELLOW token, running a ClearNode locally, fund flow, key concepts, and a suite index pointing to every specialist skill. Use when: starting a new Yellow Network integration, understanding the protocol, setting up environments, or deciding which specialist skill covers your current task.
version: 4.0.0
sdk_version: "@yellow-org/sdk@^1.2.0"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org
  - https://docs.yellow.org/whitepaper
  - https://docs.yellow.org/docs/0.5.x/guides/manuals/running-clearnode-locally
---

# Yellow Network Builder Guide

**Protocol**: NitroRPC/0.4 (wire) + VirtualApp:Custody v0.3.0 (on-chain)
**Active SDK**: `@yellow-org/sdk@^1.2.0` (legacy: `@erc7824/nitrolite@^0.5.3`, frozen)

Yellow Network is a Layer-3 overlay for decentralised clearing and
settlement across EVM chains. Lock funds on-chain once; transact unlimited
times off-chain via signed state updates; settle on-chain once. Finality
< 1 second, zero gas per operation, funds always recoverable via the
Custody contract.

## "Brokers" — clarifying a confusing term

Yellow's docs call ClearNode operators "brokers" because they hold and
update ledger state. They are **not** market-makers. There is no public
broker registry, no broker matching engine, no `submit_order` /
`get_quote` RPC. Order books, quotes, matching — all of it lives in the
**application layer** that you build.

Public ClearNodes (sandbox + mainnet) are reachable by anyone — no
partnership, no whitelisting. Running your own ClearNode requires locking
**250,000 $YELLOW** as operator collateral; using one does not.

For asset exchange (swaps), see `yellow-swap-design`.

## Yellow Network repo map (github.com/layer-3)

| Repo | What it is | Use for |
|---|---|---|
| [`nitrolite`](https://github.com/layer-3/nitrolite) | Protocol + Go/TS SDK + ClearNode + smart contracts | Default starting point; everything builders need |
| [`docs`](https://github.com/layer-3/docs) | Source of `docs.yellow.org` | Cite the docs site, not the repo, for stable URLs |
| [`clearsync`](https://github.com/layer-3/clearsync) | B2B inter-exchange clearing protocol (Solidity + Go) | Algo-trading firms / exchanges, not retail apps |
| [`broker-contracts`](https://github.com/layer-3/broker-contracts) | Early broker token + LiteVault primitives | Reference only; dormant since 2025 |
| [`cosign-demo`](https://github.com/layer-3/cosign-demo) | Shared-approval demo on Nitrolite | Learn app sessions by example |

`yellow.pro` is a Yellow-built consumer **trading product** (UI). It is
**not** publicly callable infrastructure — don't try to integrate against
it as a backend.

## Three-layer architecture

| Layer | Components | Speed | Cost |
|---|---|---|---|
| **Application** | Your business logic | Instant | Zero |
| **Off-Chain** | ClearNode + Nitro RPC (WebSocket) | < 1 s | Zero gas |
| **On-Chain** | Custody + Adjudicator contracts | Block time | Gas fees |

The blockchain is touched only when: opening a channel, funding/resizing,
cooperative close, or dispute. Everything else — transfers, app-session
state updates, queries — is off-chain.

## Environments

| Environment | WebSocket | Faucet | Assets |
|---|---|---|---|
| Sandbox | `wss://clearnet-sandbox.yellow.com/ws` | Yes | `ytest.usd` |
| Mainnet | `wss://clearnet.yellow.com/ws` | No | USDC, YELLOW, ETH, etc. |

$YELLOW token (Ethereum mainnet): `0x236eb848c95b231299b4aa9f56c73d6893462720`
— 10 B fixed supply, ERC-20, deploy 2026-03-05. Used for fees, tier staking,
governance. Full tokenomics in `reference.md`.

## Quick start

```bash
npm install @yellow-org/sdk viem
```

```ts
import { Client, createSigners } from '@yellow-org/sdk';

const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);
const client = await Client.create(
  'wss://clearnet.yellow.com/ws',
  stateSigner, txSigner,
);

// v1 Client methods are positional. blockchainId is bigint, amount is Decimal.
import Decimal from 'decimal.js';
await client.deposit(137n, 'usdc', new Decimal('100'));
await client.transfer('0x…', 'usdc', new Decimal('50'));
```

See `yellow-sdk-v1` for the full Client API.

## Builder suite index

Drop into the right specialist skill for the task:

**Protocol foundations**
- `yellow-nitro-rpc` — wire format `{req, sig}` envelope, signatures, timestamps
- `yellow-clearnode-auth` — 3-step auth + EIP-712 Policy + JWT reuse
- `yellow-session-keys` — hot-wallet delegation, allowances, rotation
- `yellow-errors` — descriptive-string error model + recovery

**Operational primitives**
- `yellow-transfers` — instant P2P transfers over unified balance
- `yellow-app-sessions` — multi-party escrow / stake / game sessions
- `yellow-state-channels` — channel lifecycle (off-chain + on-chain)
- `yellow-deposits-withdrawals` — on-chain funding in/out
- `yellow-queries` — 12 read-only RPC methods
- `yellow-swap-design` — three patterns for off-chain asset exchange when there's no native swap protocol
- `yellow-notifications` — `bu`/`cu`/`tr`/`asu` server-push

**On-chain**
- `yellow-custody-contract` — Solidity ABI, events, structs

**SDKs**
- `yellow-sdk-v1` — `@yellow-org/sdk@1.x` (**active**)
- `yellow-sdk-api` — legacy `@erc7824/nitrolite@0.5.3` (frozen)

## Navigation Guide

### When to read supporting files

**reference.md** — read when you need:

- Full `$YELLOW` tokenomics (allocation table, vesting, uses)
- **Run-a-ClearNode-locally recipe** (docker compose steps, ports table, env vars)
- Full architecture breakdown (what's each layer's responsibility)
- Supported chains list with chain IDs
- Deeper quick-start with raw Nitro RPC (the legacy / no-SDK path)
- Fund-flow diagrams across layers
- Glossary of terms (unified balance, app session, challenge window, etc.)

## Related

Start with `yellow-sdk-v1` for the active SDK, then follow the suite
index above for specialist topics.
