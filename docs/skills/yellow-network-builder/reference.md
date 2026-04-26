# network-builder — Complete Reference

> Extended documentation. Read the `SKILL.md` quick-start first; this file contains the full details moved out for progressive disclosure.


# Yellow Network Builder Guide

**Status**: Production Ready
**Version**: 3.0.0
**Last Verified**: 2026-04-23
**Dependencies**: `@yellow-org/sdk@^1.2.0` (active) or `@erc7824/nitrolite@^0.5.3` (legacy/frozen), `viem`
**Protocol**: NitroRPC/0.4 (spec) + VirtualApp:Custody v0.3.0 (on-chain)


## $YELLOW tokenomics (whitepaper)

| Fact | Value |
|---|---|
| Total supply | 10,000,000,000 $YELLOW (fixed) |
| Standard | ERC-20 |
| Mainnet deploy | 2026-03-05 |
| Public sale price | $0.011 |

Allocation:

| Allocation | % |
|---|---|
| Founders (6mo cliff, 60mo vest) | 10% |
| Token sales | 12.5% |
| Community Treasury | 30% |
| Foundation Treasury | 20% |
| Network Growth | 25% |
| Ecosystem Reserve | 2.5% |

Mainnet $YELLOW token: `0x236eb848c95b231299b4aa9f56c73d6893462720`.

## Run a ClearNode locally

From the `guides/manuals/running-clearnode-locally` page:

```bash
git clone <clearnode repo>
cd clearnode
cp -r config/compose/example config/compose/local
docker compose up
```

| Port | Service |
|---|---|
| 8000 | WS + HTTP (Nitro RPC) |
| 4242 | Prometheus metrics |
| 5432 | Postgres (if using postgres driver) |

Key env vars (verified against `clearnode/runtime.go` and `clearnode/main.go`):

| Var | Meaning |
|---|---|
| `BROKER_PRIVATE_KEY` | ClearNode operator key |
| `CLEARNODE_DATABASE_URL` | DB connection string |
| `CLEARNODE_CONFIG_DIR_PATH` | Config override dir |
| `CLEARNODE_LOG_LEVEL` | `debug` / `info` / `warn` / `error` |
| `<CHAIN>_BLOCKCHAIN_RPC` | Per-chain RPC URL (e.g. `ETHEREUM_BLOCKCHAIN_RPC`) |

For ports + database driver + message-expiry settings, edit the YAML
config under `CLEARNODE_CONFIG_DIR_PATH` — the runtime does not read
`HTTP_PORT`, `METRICS_PORT`, `DATABASE_DRIVER`, or `MSG_EXPIRY_TIME`
as environment variables. Earlier versions of these docs listed them as
env vars; that was wrong.

## Builder Suite Index

Start here, then drop into the specialist skill for whatever you're building.
All skills are verified against `docs.yellow.org` v0.5.x.

**Protocol foundations**
- `yellow-nitro-rpc` — wire format: `{req, sig}` envelope, signatures, timestamps
- `yellow-clearnode-auth` — 3-step auth handshake with EIP-712 Policy, JWT reuse
- `yellow-session-keys` — hot-wallet delegation, allowances, scopes, rotation
- `yellow-errors` — descriptive-string error model (no numeric codes) + recovery

**Operational primitives**
- `yellow-transfers` — instant P2P transfers over unified balance
- `yellow-app-sessions` — multi-party escrow / stake / game sessions (OPERATE / DEPOSIT / WITHDRAW)
- `yellow-state-channels` — channel lifecycle (off-chain RPC + on-chain)
- `yellow-deposits-withdrawals` — on-chain funding in and out of unified balance
- `yellow-queries` — all 12 read-only RPC methods with verbatim field names
- `yellow-notifications` — the 4 server-push events (bu / cu / tr / asu)

**On-chain reference**
- `yellow-custody-contract` — Solidity interfaces, EIP-712 typehashes, events, structs

**SDKs**
- `yellow-sdk-v1` — **`@yellow-org/sdk@1.x`** (active upstream SDK)
- `yellow-sdk-api` — legacy `@erc7824/nitrolite@0.5.3` (frozen; reference only)


## What Yellow Network Solves

Every blockchain transaction requires global consensus. This creates three fundamental constraints:

| Challenge | Impact |
|-----------|--------|
| **High latency** | 15 seconds to minutes per confirmation |
| **Gas costs** | Fees spike during congestion; microtransactions impractical |
| **Limited throughput** | Ethereum processes ~15-30 TPS |

Yellow Network removes these constraints for applications by moving high-frequency operations **off-chain** while preserving blockchain-level security.

### The Core Model: Lock Once, Transact Unlimited, Settle Once

Instead of putting every operation on-chain, Yellow Network uses **state channels**:

1. **Lock funds** in a smart contract (one on-chain transaction)
2. **Transact instantly** off-chain with cryptographic signatures (zero cost, sub-second)
3. **Settle** the final state (one on-chain transaction)

Consider a 10 USDC wager chess game: on-chain = 40+ transactions = $100s in fees. State channel = 2 transactions = minimal fees, regardless of number of moves.

### What You Get

| Feature | Value |
|---------|-------|
| **Instant finality** | Sub-second (< 1 second typical) |
| **Zero gas per operation** | Off-chain operations have no blockchain fees |
| **Unlimited throughput** | No consensus bottleneck during active phase |
| **Blockchain security** | Funds always recoverable via on-chain contracts |


## Three-Layer Architecture

| Layer | Components | Speed | Cost |
|-------|-----------|-------|------|
| **Application** | Business logic, UI, agent services | Instant (local) | Zero |
| **Off-Chain** | ClearNode + Nitro RPC state channel coordination | < 1 second | Zero gas |
| **On-Chain** | Custody contracts, adjudicators, NodeRegistry, AppRegistry | Block time | Gas fees |

The blockchain is only touched in 4 scenarios:
1. Opening a channel (lock funds)
2. Resizing a channel (add or remove funds)
3. Closing a channel cooperatively (distribute funds)
4. Disputing a state (challenge-response)

Everything else — transfers, app state updates, session operations — is off-chain.


## $YELLOW Token

| Property | Value |
|----------|-------|
| **Address** | `0x236eb848c95b231299b4aa9f56c73d6893462720` |
| **Standard** | ERC-20 |
| **Network** | Ethereum mainnet (bridgeable to other chains) |

### Uses

| Use | Description |
|-----|-------------|
| **Fees** | Platform fees denominated in $YELLOW |
| **Staking** | Agent tier collateral (0 → 100 → 1,000 → 10,000 → 100,000 YELLOW) |
| **Governance** | Node-tier agents participate in protocol governance |
| **Payment abstraction** | Unified account abstraction across supported chains |


## Environments

| Environment | WebSocket | Faucet | Assets |
|-------------|-----------|--------|--------|
| **Sandbox** | `wss://clearnet-sandbox.yellow.com/ws` | Yes | Test tokens (`ytest.usd`) |
| **Mainnet** | `wss://clearnet.yellow.com/ws` | No | Real USDC, YELLOW, ETH, etc. |

Configure via `CLEARNODE_WS_URL` and `CLEARNODE_NETWORK` env vars. Typical setup:

```typescript
const CLEARNODE_URLS = {
  sandbox: 'wss://clearnet-sandbox.yellow.com/ws',
  mainnet: 'wss://clearnet.yellow.com/ws',
};
```


## Quick Start

### 1. Install

```bash
npm install @yellow-org/sdk viem
# legacy/frozen: npm install @erc7824/nitrolite viem
```

### 2. Connect

```typescript
import { Client } from '@yellow-org/sdk';
// or roll your own WebSocket wrapper — see yellow-sdk-v1

const CLEARNODE_URLS = {
  sandbox: 'wss://clearnet-sandbox.yellow.com/ws',
  mainnet: 'wss://clearnet.yellow.com/ws',
};

const client = await Client.create(CLEARNODE_URLS.mainnet, stateSigner, txSigner);
// client.connected === true
```

### 3. Authenticate

```typescript
import { privateKeyToAccount, generatePrivateKey } from 'viem/accounts';

// Generate ephemeral session key (stored in memory, never persisted)
const sessionKey = privateKeyToAccount(generatePrivateKey());

// Use @yellow-org/sdk's Client.create() which authenticates inline,
// or build the 3-step handshake — see yellow-clearnode-auth
//
// Auth Policy payload (canonical v0.5.3+ field names):
//   address:     agentAddress              // Main wallet (EOA)
//   session_key: sessionKey.address
//   application: 'my-app'
//   allowances:  [
//     { asset: 'usdc',   amount: '10000.0' },
//     { asset: 'yellow', amount: '100000.0' },
//   ]
//   expires_at:  Date.now() + 86_400_000    // milliseconds
// Signed with the main wallet via EIP-712.
// After verify, use sessionKey.sign() for all subsequent operations.
```

### 4. Use the Network

```typescript
// Use client.transfer(...) from @yellow-org/sdk,
// or send a raw 'transfer' Nitro RPC — see yellow-transfers

// Instant off-chain transfer — no gas. v1 Client.transfer takes positional
// args: (recipientWallet, asset, Decimal amount).
import Decimal from 'decimal.js';
await client.transfer(recipientAddress, 'usdc', new Decimal('50.0'));
```


## Fund Flow

```text
User Wallet (ERC-20 on any chain)
  |
  | deposit (on-chain tx)
  v
Available Balance (Custody Contract — on-chain)
  |
  | resize / channel open (on-chain tx)
  v
Channel-Locked (Custody Contract — on-chain)
  |
  | ClearNode coordination
  v
Unified Balance (ClearNode — off-chain)
  |                                    |
  | create app session                 | transfer (instant)
  v                                    v
App Session (off-chain)         Recipient Unified Balance
  |
  | close app session
  v
Unified Balance (returned)
  |
  | resize / channel close (on-chain tx)
  v
Available Balance
  |
  | withdraw (on-chain tx)
  v
User Wallet
```

**Cross-chain aggregation**: Deposit 50 USDC on Polygon + 50 USDC on Base → Unified Balance shows 100 USDC. Withdraw all 100 to Arbitrum in a single transaction.


## Supported Chains

| Chain | Chain ID | Key Assets |
|-------|----------|-----------|
| Ethereum | 1 | USDC, YELLOW |
| BNB Smart Chain | 56 | USDC, USDT, BNB |
| Polygon | 137 | USDC, USDT, WETH |
| World Chain | 480 | USDC |
| Base | 8453 | USDC, USDT, ETH |
| Linea | 59144 | USDC, USDT, ETH |
| XRPL EVM | 1440000 | XRP |


## Key Concepts

### Unified Balance

An off-chain account maintained by ClearNode. It aggregates funds from all channels across all chains. The source for transfers and app session deposits. Backed by on-chain locked funds.

### App Sessions

Multi-party off-chain channels with custom governance rules. Built on top of unified balance. Use case: escrow, prediction markets, staking, gaming. See `yellow-app-sessions` for the protocol details.

Key properties:
- Protocol: `NitroRPC/0.4` (always use this for new sessions)
- Governance: quorum-based (weights + threshold)
- `app_session_id` = `keccak256(JSON.stringify(definition))`
- Version must increment by exactly 1 per state update
- Allocations represent **final state**, not deltas

### Session Keys

Ephemeral signing keys generated locally. After the one-time EIP-712 auth, all subsequent messages are signed with the session key. Allowances cap spending per asset. Sessions expire at a configured Unix timestamp.

### Challenge-Response

The security guarantee. If ClearNode goes offline or acts maliciously, any participant can submit their latest signed state to the blockchain. After the challenge period, funds are distributed according to that state. No one can steal funds — only a valid signed state can change allocations.


## Protocol Version Reference

| Version | Status | Features |
|---------|--------|----------|
| NitroRPC/0.2 | Legacy | Basic state updates only |
| **NitroRPC/0.4** | **Current** | Intent system (OPERATE/DEPOSIT/WITHDRAW), dynamic fund management |

Always use `NitroRPC/0.4` for new app sessions. The protocol version is set at session creation and cannot be changed.


## Reference

- Protocol docs: <https://docs.yellow.org>
- Active SDK: `@yellow-org/sdk` (see `yellow-sdk-v1`)
- Legacy SDK (frozen): `@erc7824/nitrolite@0.5.3` (see `yellow-sdk-api`)
