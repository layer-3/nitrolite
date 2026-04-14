# MCP Vetting Results — TypeScript Server

**77 passed, 0 failed, 77 total**

---

## Inventory

### 0.1: 30 resources registered

**Status:** PASS

```
nitrolite://api/enums
nitrolite://api/methods
nitrolite://api/types
nitrolite://examples/app-sessions
nitrolite://examples/auth
nitrolite://examples/channels
nitrolite://examples/full-app-session-script
nitrolite://examples/full-transfer-script
nitrolite://examples/transfers
nitrolite://go-api/methods
nitrolite://go-api/types
nitrolite://go-examples/full-app-session-script
nitrolite://go-examples/full-transfer-script
nitrolite://migration/overview
nitrolite://protocol/auth-flow
nitrolite://proto
... (981 chars total, truncated)
```

### 0.2: 8 tools: explain_concept, get_rpc_method, lookup_method, lookup_rpc_method, lookup_type, scaffold_project, search_api, validate_import

**Status:** PASS

```
explain_concept, get_rpc_method, lookup_method, lookup_rpc_method, lookup_type, scaffold_project, search_api, validate_import
```

### 0.3: 3 prompts: build-ai-agent-app, create-channel-app, migrate-from-v053

**Status:** PASS

```
build-ai-agent-app, create-channel-app, migrate-from-v053
```

## Transfer

### B.1: search_api("transfer") mentions Decimal

**Status:** PASS

```
# Search results for "transfer"

## Methods (6 matches)
- `create(wsURL: string,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    ...opts: Option[]): Promise<Client>` — Other
- `setHomeBlockchain(asset: string, blockchainId: bigint): Promise<void>` — Node & Queries
- `signState(state: core.State): Promise<Hex>` — Signing
- `validateAndSignState(currentState: core.State, proposedState: core.State): Promise<Hex>` — Signing
- `transfer(recipientWallet: string, asset: string, amoun
... (946 chars total, truncated)
```

### B.2: lookup_method("create") — correct signature

**Status:** PASS

```
### create
**Signature:** `create(wsURL: string,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    ...opts: Option[]): Promise<Client>`
**Category:** Other
**Description:** Main Nitrolite SDK Client Provides a unified interface for interacting with Nitrolite payment channels. Combines both high-level operations (Deposit, Withdraw, Transfer) and low-level RPC access for advanced use cases. /

import { Address, Hex, createPublicClient, createWalletClient, http, custom, verifyMessa
... (6842 chars total, truncated)
```

### B.3: lookup_method("deposit") — Decimal amount

**Status:** PASS

```
### deposit
**Signature:** `deposit(blockchainId: bigint, asset: string, amount: Decimal): Promise<core.State>`
**Category:** Transactions
**Description:** SignAndSubmitState is a helper that validates, signs a state and submits it to the node. Returns the node's signature. /
  private async signAndSubmitState(currentState: core.State, proposedState: core.State): Promise<Hex> {
    // Validate and sign state
    const sig = await this.validateAndSignState(currentState, proposedState);
    propos
... (2759 chars total, truncated)
```

### B.4: lookup_method("transfer") — Decimal amount

**Status:** PASS

```
### transfer
**Signature:** `transfer(recipientWallet: string, asset: string, amount: Decimal): Promise<core.State>`
**Category:** Transactions
**Description:** Transfer prepares a transfer state to send funds to another wallet address. This method handles two scenarios automatically: 1. If no channel exists: Creates a new channel with the transfer transition 2. If channel exists: Advances the state with a transfer send transition  The returned state is signed by both the user and the node. For 
... (1032 chars total, truncated)
```

### B.5: lookup_method("approveToken") — 3 params

**Status:** PASS

```
### approveToken
**Signature:** `approveToken(chainId: bigint, asset: string, amount: Decimal): Promise<string>`
**Category:** Transactions
**Description:** Approve the ChannelHub contract to spend tokens on behalf of the user. This is required before depositing ERC-20 tokens. Native tokens (e.g., ETH) do not require approval and will throw an error if attempted.  @param chainId - The blockchain network ID (e.g., 11155111n for Sepolia) @param asset - The asset symbol to approve (e.g., "usdc") @p
... (602 chars total, truncated)
```

### B.6: lookup_method("checkpoint")

**Status:** PASS

```
### checkpoint
**Signature:** `checkpoint(asset: string): Promise<string>`
**Category:** Other
**Description:** Checkpoint executes the blockchain transaction for the latest signed state. It fetches the latest co-signed state and, based on the transition type and on-chain channel status, calls the appropriate blockchain method.  This is the only method that interacts with the blockchain. It should be called after any state-building method (deposit, withdraw, closeHomeChannel, etc.) to settle the
... (1217 chars total, truncated)
```

### B.7: lookup_method("getBalances") — wallet param

**Status:** PASS

```
### getBalances
**Signature:** `getBalances(wallet: Address): Promise<core.BalanceEntry[]>`
**Category:** Node & Queries
**Description:** GetBalances retrieves the balance information for a user's wallet.  @param wallet - The user's wallet address @returns Array of balance entries for each asset  @example ```typescript const balances = await client.getBalances('0x1234...'); for (const entry of balances) { console.log(`${entry.asset}: ${entry.balance}`); } ```
```

### B.8: lookup_method("closeHomeChannel")

**Status:** PASS

```
### closeHomeChannel
**Signature:** `closeHomeChannel(asset: string): Promise<core.State>`
**Category:** Channels
**Description:** CloseHomeChannel prepares a finalize state to close the user's channel for a specific asset. This creates a final state with zero user balance and submits it to the node.  The returned state is signed by both the user and the node, but has not yet been submitted to the blockchain. Use {@link checkpoint} to execute the on-chain close.  @param asset - The asset symbol 
... (758 chars total, truncated)
```

### B.9: scaffold_project("transfer-app")

**Status:** PASS

```
# Scaffold: transfer-app

## package.json
```json
{
  "name": "nitrolite-transfer-app",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "start": "npx tsx src/index.ts",
    "build": "tsc",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@yellow-org/sdk": "^1.2.0",
    "decimal.js": "^10.4.0",
    "viem": "^2.46.0"
  },
  "devDependencies": {
    "typescript": "^5.7.0",
    "tsx": "^4.19.0",
    "@types/node": "^22.0.0"
  }
}
```

## tsconfig.json
```
... (2002 chars total, truncated)
```

### B.10: full transfer script resource

**Status:** PASS

```
# Complete Transfer Script

A fully working TypeScript script that connects to a clearnode, opens a channel, deposits funds, transfers tokens, and closes the channel.

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// --- Configuration ---
const PRIVATE_KEY = process.env.PRIVATE_KEY as `0x${string}`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_URL
... (2505 chars total, truncated)
```

### B.11: get_rpc_method("transfer") — object params

**Status:** PASS

```
## RPC: transfer

**V1 Wire Method:** `channels.v1.submit_state`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "transfer", { destination, allocations: [{ asset, amount }] }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "transfer", { state }] }
```
```

## AppSession

### C.1: search_api("app session")

**Status:** PASS

```
# Search results for "app session"

## Methods (8 matches)
- `getAppSessions(options?: {
    appSessionId?: string;
    wallet?: Address;
    status?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ sessions: app.AppSessionInfoV1[]; metadata: core.PaginationMetadata }>` — App Sessions
- `getAppDefinition(appSessionId: string): Promise<app.AppDefinitionV1>` — Other
- `createAppSession(definition: app.AppDefinitionV1,
    sessionData: string,
    quorumSigs: string[],
    opts?: 
... (1233 chars total, truncated)
```

### C.2: createAppSession signature

**Status:** PASS

```
### createAppSession
**Signature:** `createAppSession(definition: app.AppDefinitionV1,
    sessionData: string,
    quorumSigs: string[],
    opts?: { ownerSig?: string }): Promise<{ appSessionId: string; version: string; status: string }>`
**Category:** App Sessions
**Description:** CreateAppSession creates a new application session between participants.  @param definition - The app definition with participants, quorum, application ID @param sessionData - Optional JSON stringified session data 
... (1027 chars total, truncated)
```

### C.3: submitAppState signature

**Status:** PASS

```
### submitAppState
**Signature:** `submitAppState(appStateUpdate: app.AppStateUpdateV1,
    quorumSigs: string[]): Promise<void>`
**Category:** App Sessions
**Description:** SubmitAppState submits an app session state update. This method handles operate, withdraw, and close intents. For deposits, use submitAppSessionDeposit instead.  @param appStateUpdate - The app state update (intent: operate, withdraw, or close) @param quorumSigs - Participant signatures for the app state update  @example ```
... (889 chars total, truncated)
```

### C.4: closeAppSession — correctly absent

**Status:** PASS
**Notes:** Correctly absent

```
No method matching "closeAppSession" found. Available categories: Other, Node & Queries, Signing, Transactions, Channels, App Sessions
```

### C.5: lookup_type AppDefinitionV1

**Status:** PASS

```
### AppDefinitionV1 (interface)
**Source:** sdk/ts (app)
```typescript
applicationId: string;
  participants: AppParticipantV1[];
  quorum: number; // uint8
  nonce: bigint; // uint64
```
```

### C.6: lookup_type AppStateUpdateV1

**Status:** PASS

```
### AppStateUpdateV1 (interface)
**Source:** sdk/ts (app)
```typescript
appSessionId: string;
  intent: AppStateUpdateIntent;
  version: bigint; // uint64
  allocations: AppAllocationV1[];
  sessionData: string;
```

---

### SignedAppStateUpdateV1 (interface)
**Source:** sdk/ts (app)
```typescript
appStateUpdate: AppStateUpdateV1;
  quorumSigs: string[];
```
```

### C.7: scaffold_project("app-session")

**Status:** PASS

```
# Scaffold: app-session

## package.json
```json
{
  "name": "nitrolite-app-session",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "start": "npx tsx src/index.ts",
    "build": "tsc",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@yellow-org/sdk": "^1.2.0",
    "decimal.js": "^10.4.0",
    "viem": "^2.46.0"
  },
  "devDependencies": {
    "typescript": "^5.7.0",
    "tsx": "^4.19.0",
    "@types/node": "^22.0.0"
  }
}
```

## tsconfig.json
```js
... (3099 chars total, truncated)
```

### C.8: app-sessions resource

**Status:** PASS

```
# Nitrolite SDK — App Session Examples

## Creating an App Session

```typescript
import { app } from '@yellow-org/sdk';

// 1. Define the app session
const definition: app.AppDefinitionV1 = {
    applicationId: 'my-game-app',
    participants: [
        { walletAddress: '0xAlice...', signatureWeight: 50 },
        { walletAddress: '0xBob...', signatureWeight: 50 },
    ],
    quorum: 100, // Both must agree
    nonce: BigInt(Date.now()),
};

// 2. Collect quorum signatures from participants (of
... (1402 chars total, truncated)
```

### C.9: full app-session script

**Status:** PASS

```
# Complete App Session Script

A fully working TypeScript script that creates a multi-party app session, submits state updates, and closes with final allocations.

```typescript
import { Client, createSigners, withBlockchainRPC, app } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

// --- Configuration ---
const PRIVATE_KEY = process.env.PRIVATE_KEY as `0x${string}`;
const CLEARNODE_URL = process.env.CLEARNODE_URL || 'wss://clearnode.example.com/ws';
const RPC_URL = process.env.RPC_UR
... (3403 chars total, truncated)
```

### C.10: get_rpc_method("create_app_session")

**Status:** PASS

```
## RPC: create_app_session

**V1 Wire Method:** `app_sessions.v1.create_app_session`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "create_app_session", { definition, session_data, quorum_sigs, owner_sig? }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "create_app_session", { app_session_id, version, status }] }
```
```

### C.11: get_rpc_method("submit_app_state")

**Status:** PASS

```
## RPC: submit_app_state

**V1 Wire Method:** `app_sessions.v1.submit_app_state`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "submit_app_state", { app_state_update, quorum_sigs }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "submit_app_state", { accepted: boolean }] }
```
```

## AIAgent

### D.1: scaffold_project("ai-agent")

**Status:** PASS

```
# Scaffold: ai-agent

## package.json
```json
{
  "name": "nitrolite-ai-agent",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "start": "npx tsx src/index.ts",
    "build": "tsc",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@yellow-org/sdk": "^1.2.0",
    "decimal.js": "^10.4.0",
    "viem": "^2.46.0"
  },
  "devDependencies": {
    "typescript": "^5.7.0",
    "tsx": "^4.19.0",
    "@types/node": "^22.0.0"
  }
}
```

## tsconfig.json
```json
{
 
... (2598 chars total, truncated)
```

### D.2: AI agents use-case resource

**Status:** PASS

```
# AI Agent Use Cases

How to use Nitrolite for AI agent payments and agent-to-agent interactions.

## Why State Channels for AI Agents?

AI agents need to make frequent, small payments — often thousands per session. On-chain transactions are too slow and expensive. State channels provide:
- **Instant finality** — no waiting for block confirmations
- **Near-zero cost** — gas only on channel open/close, not per-transfer
- **Programmable** — agents manage channels autonomously via the SDK

## Agent
... (2312 chars total, truncated)
```

## Migration

### E.1: validate_import("NitroliteClient")

**Status:** PASS

```
**NitroliteClient** is exported from `@yellow-org/sdk-compat`.

```typescript
import { NitroliteClient } from '@yellow-org/sdk-compat';
```
```

### E.2: validate_import("createAuthRequestMessage")

**Status:** PASS

```
**createAuthRequestMessage** is exported from `@yellow-org/sdk-compat`.

```typescript
import { createAuthRequestMessage } from '@yellow-org/sdk-compat';
```
```

### E.3: validate_import("Client") — in SDK, not compat

**Status:** PASS

```
**Client** is NOT in `@yellow-org/sdk-compat` but IS in `@yellow-org/sdk`.

```typescript
import { Client } from '@yellow-org/sdk';
```

> Note: SDK classes should not be re-exported from compat (SSR risk). Import directly from `@yellow-org/sdk`.
```

### E.4: validate_import("RPCMethod")

**Status:** PASS

```
**RPCMethod** is exported from `@yellow-org/sdk-compat`.

```typescript
import { RPCMethod } from '@yellow-org/sdk-compat';
```
```

### E.5: validate_import("Cli") — must NOT match

**Status:** PASS

```
**Cli** was not found in either `@yellow-org/sdk-compat` or `@yellow-org/sdk` barrel exports. It may be a deep import or may not exist.
```

### E.6: validate_import("FakeSymbol")

**Status:** PASS

```
**FakeSymbol** was not found in either `@yellow-org/sdk-compat` or `@yellow-org/sdk` barrel exports. It may be a deep import or may not exist.
```

### E.7: auth example resource

**Status:** PASS

```
# Nitrolite SDK — Authentication Examples

## Compat Layer Auth Flow (Legacy v0.5.3 Pattern)

```typescript
import {
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createEIP712AuthMessageSigner,
    parseAnyRPCResponse,
    type AuthRequestParams,
} from '@yellow-org/sdk-compat';

// 1. Create auth request
const authParams: AuthRequestParams = {
    address: account.address,
    session_key: '0x0000000000000000000000000000000000000000',
    application: 'My App',
    expires_at: 
... (1380 chars total, truncated)
```

### E.8: migration overview resource

**Status:** PASS

```
# Migrating from v0.5.3 to Compat Layer

This guide explains how to migrate your Nitrolite dApp from the v0.5.3 SDK to the **compat layer**, which bridges the old API to the v1.0.0 runtime with minimal code changes.

## 1. Why Use the Compat Layer

The v1.0.0 protocol introduces breaking changes across wire format, authentication, WebSocket lifecycle, unit system, asset resolution, and more. A direct migration touches **20+ files** per app with deep, scattered rewrites.

The compat layer central
... (3736 chars total, truncated)
```

### E.9: migrate-from-v053 prompt

**Status:** PASS

```
I need to migrate my app from `@layer-3/nitrolite` v0.5.3 to the new SDK. Help me step by step.

Here is the official migration guide:

# Migrating from v0.5.3 to Compat Layer

This guide explains how to migrate your Nitrolite dApp from the v0.5.3 SDK to the **compat layer**, which bridges the old API to the v1.0.0 runtime with minimal code changes.

## 1. Why Use the Compat Layer

The v1.0.0 protocol introduces breaking changes across wire format, authentication, WebSocket lifecycle, unit syste
... (4215 chars total, truncated)
```

## Protocol

### F.1: explain_concept("state channel")

**Status:** PASS

```
**Channel**

A state container shared between a user and a node that allows off-chain state updates while maintaining on-chain security guarantees. Each channel operates on a single unified asset.

---

**State**

An abstract data structure representing the current configuration of a protocol entity at a specific version.
```

### F.2: explain_concept("app session")

**Status:** PASS

```
No exact match for "app session". Related concepts:

**Session Key**

A delegated signing key authorized by a participant's primary key to sign specific types of state updates on their behalf. Session key authorization MUST be associated with the same address as the channel's user or node participant.

---

**Application Session**

An extension that enables off-chain application functionality. Application sessions hold committed assets and maintain their own state.

---

**Application State**

T
... (685 chars total, truncated)
```

### F.3: explain_concept("challenge period")

**Status:** PASS

```
**Challenge**

An on-chain operation where a participant disputes the current enforced state by submitting a signed state along with a challenger signature. Initiates the challenge duration, during which other participants MAY respond with a higher-version state.
```

### F.4: explain_concept("clearnode")

**Status:** PASS

```
No concept matching "clearnode" found. 34 concepts indexed from protocol terminology. Try broader terms like "channel", "state", "session", "escrow", "transfer".
```

### F.5: explain_concept("made up thing") — graceful

**Status:** PASS

```
No concept matching "made up thing" found. 34 concepts indexed from protocol terminology. Try broader terms like "channel", "state", "session", "escrow", "transfer".
```

### F.6: resource nitrolite://protocol/overview

**Status:** PASS

```
# Nitrolite Protocol Overview

Nitrolite is a state channel protocol that enables high-speed off-chain interactions between users while preserving on-chain security guarantees.

Users exchange signed state updates off-chain with Nodes that act as a hub connecting network participants. Any user can enforce the latest agreed state on the blockchain layer at any time.

## Table of Contents

1. [Overview](overview.md) — high-level protocol description and design goals
2. [Terminology](terminology.md
... (5058 chars total, truncated)
```

### F.7: resource nitrolite://protocol/terminology

**Status:** PASS

```
# Terminology

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)

---

This document defines all protocol terms used throughout the Nitrolite protocol documentation.

Each term is defined once. All other documents MUST use these terms consistently.

## Naming Conventions

- Protocol entities use CamelCase (e.g., ChannelState, AppSession)
- Field names use CamelCase (e.g., ChannelId, StateVersion)
- Operations use lowercase with hyphens in document references (e.g., state-
... (7514 chars total, truncated)
```

### F.8: resource nitrolite://protocol/wire-format

**Status:** PASS

```
# Interaction Model

Previous: [Cross-Chain and Assets](cross-chain-and-assets.md) | Next: [Security and Limitations](security-and-limitations.md)

---

This document defines the logical communication protocol between participants.

All operations are defined as semantic protocol operations, independent of transport technologies such as WebSocket or gRPC.

## Purpose

Participants exchange protocol messages to advance state, manage channels, and coordinate operations. This document defines the s
... (5043 chars total, truncated)
```

### F.9: resource nitrolite://security/overview

**Status:** PASS

```
# Security and Limitations

Previous: [Interactions](interactions.md) | Next: [Extensions Overview](extensions/overview.md)

---

This document describes the security guarantees of the Nitrolite protocol, its current trust assumptions, and the known limitations of the present version.

## Protocol Maturity

The core protocol functionality is implemented and operational. A user MAY operate over a unified asset, deposit and withdraw on any supported blockchain, and conduct the majority of interact
... (5370 chars total, truncated)
```

### F.10: resource nitrolite://security/app-session-patterns

**Status:** PASS

```
# App Session Security Patterns

Best practices for building secure, decentralization-ready app sessions on Nitrolite.

## Quorum Design

App sessions use a weight-based quorum system for governance:

```typescript
interface AppDefinitionV1 {
  applicationId: string;
  participants: AppParticipantV1[];  // each has walletAddress + signatureWeight
  quorum: number;                     // minimum total weight to authorize actions (uint8)
  nonce: bigint;
}
```

### Recommended Patterns

**Equal 2-
... (2933 chars total, truncated)
```

### F.11: resource nitrolite://security/state-invariants

**Status:** PASS

```
# State Invariants

Critical invariants that MUST hold across all state transitions. Violating these will cause on-chain enforcement to fail.

## Ledger Invariant (Fund Conservation)

```
UserAllocation + NodeAllocation == UserNetFlow + NodeNetFlow
```

This ensures no assets can be created or destroyed through state transitions. The total distributable balance always equals the total cumulative flows.

## Allocation Non-Negativity

All allocation values (UserAllocation, NodeAllocation) MUST be 
... (1618 chars total, truncated)
```

## RPC

### G.1: get_ledger_balances — wallet param

**Status:** PASS

```
## RPC: get_ledger_balances

**V1 Wire Method:** `user.v1.get_balances`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "get_ledger_balances", { wallet }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "get_ledger_balances", { balances: RPCBalance[] }] }
```
```

### G.2: create_channel — chain_id/token

**Status:** PASS

```
## RPC: create_channel

**V1 Wire Method:** `channels.v1.request_creation`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "create_channel", [{ chain_id, token }], timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "create_channel", [{ channel_id, channel, state, server_signature }], timestamp], sig: [...] }
```
```

### G.3: close_channel — channel_id/funds_destination

**Status:** PASS

```
## RPC: close_channel

**V1 Wire Method:** `channels.v1.submit_state`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "close_channel", { channel_id, funds_destination }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "close_channel", { channel_id, state, server_signature }] }
```
```

### G.4: resize_channel — all 4 fields

**Status:** PASS

```
## RPC: resize_channel

**V1 Wire Method:** `channels.v1.submit_state`

**Request format (v0.5.3 compat):**
```
{ req: [requestId, "resize_channel", { channel_id, resize_amount, allocate_amount, funds_destination }, timestamp], sig: [...] }
```

**Response format:**
```
{ res: [requestId, "resize_channel", { channel_id, state, server_signature }] }
```
```

### G.5: nonexistent — lists available

**Status:** PASS

```
Unknown RPC method "nonexistent". Available: ping, get_channels, get_ledger_balances, transfer, create_channel, close_channel, create_app_session, submit_app_state, get_app_sessions, get_app_definition, get_ledger_transactions, resize_channel
```

## Types

### I.1: api/methods resource

**Status:** PASS

```
# Nitrolite SDK — Client Methods

## App Sessions

### `getAppSessions(options?: {
    appSessionId?: string;
    wallet?: Address;
    status?: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ sessions: app.AppSessionInfoV1[]; metadata: core.PaginationMetadata }>`
GetAppSessions retrieves application sessions for the user.  @param options - Optional filters (appSessionId, wallet, status, pagination) @returns Array of app session info and pagination metadata  @example ```typescr
... (30807 chars total, truncated)
```

### I.2: api/types resource

**Status:** PASS

```
# Nitrolite SDK — Types & Interfaces

## Interfaces (79)

### `ChannelDefinition` (sdk/ts (core))
```typescript
nonce: bigint; // uint64 - A unique number to prevent replay attacks
  challenge: number; // uint32 - Challenge period for the channel in seconds
  approvedSigValidators: string; // Hex string bitmap of approved signature validators
```

### `Channel` (sdk/ts (core))
```typescript
channelId: string; // Unique identifier for the channel
  userWallet: Address; // User wallet address
  as
... (19872 chars total, truncated)
```

### I.3: api/enums resource

**Status:** PASS

```
# Nitrolite SDK — Enums

## `ChannelType` (sdk/ts (core))
```typescript
Home = 1,
  Escrow = 2,
```

## `ChannelParticipant` (sdk/ts (core))
```typescript
User = 0,
  Node = 1,
```

## `ChannelSignerType` (sdk/ts (core))
```typescript
Default = 0x00,
  SessionKey = 0x01,
```

## `ChannelStatus` (sdk/ts (core))
```typescript
Void = 0,
  Open = 1,
  Challenged = 2,
  Closed = 3,
```

## `TransitionType` (sdk/ts (core))
```typescript
Void = 0,
  Acknowledgement = 1,
  HomeDeposit = 10,
  HomeWithdr
... (2839 chars total, truncated)
```

### I.4: lookup_type AppDefinitionV1

**Status:** PASS

```
### AppDefinitionV1 (interface)
**Source:** sdk/ts (app)
```typescript
applicationId: string;
  participants: AppParticipantV1[];
  quorum: number; // uint8
  nonce: bigint; // uint64
```
```

### I.5: lookup_type AppStateUpdateV1

**Status:** PASS

```
### AppStateUpdateV1 (interface)
**Source:** sdk/ts (app)
```typescript
appSessionId: string;
  intent: AppStateUpdateIntent;
  version: bigint; // uint64
  allocations: AppAllocationV1[];
  sessionData: string;
```

---

### SignedAppStateUpdateV1 (interface)
**Source:** sdk/ts (app)
```typescript
appStateUpdate: AppStateUpdateV1;
  quorumSigs: string[];
```
```

### I.6: lookup_type NonexistentType — graceful

**Status:** PASS

```
No type matching "NonexistentType" found. 99 TS types and 65 Go types indexed.
```

## Prompts

### J.1: create-channel-app prompt

**Status:** PASS

```
Guide me through building a Nitrolite state channel application. Cover:

1. **Setup** — Install dependencies (@yellow-org/sdk, viem), create Client with config
2. **Authentication** — Connect wallet, establish WebSocket, authenticate with clearnode
3. **Channel Lifecycle** — Deposit (auto-creates channel), query channels, close channel
4. **Transfers** — Send tokens to another participant via state channels
5. **App Sessions** — Create sessions for multi-party apps, submit state, close
6. **Erro
... (1675 chars total, truncated)
```

### J.2: migrate-from-v053 prompt

**Status:** PASS

```
I need to migrate my app from `@layer-3/nitrolite` v0.5.3 to the new SDK. Help me step by step.

Here is the official migration guide:

# Migrating from v0.5.3 to Compat Layer

This guide explains how to migrate your Nitrolite dApp from the v0.5.3 SDK to the **compat layer**, which bridges the old API to the v1.0.0 runtime with minimal code changes.

## 1. Why Use the Compat Layer

The v1.0.0 protocol introduces breaking changes across wire format, authentication, WebSocket lifecycle, unit syste
... (4215 chars total, truncated)
```

### J.3: build-ai-agent-app prompt

**Status:** PASS

```
I want to build an AI agent that uses Nitrolite state channels for payments. Guide me through:

1. **Agent Wallet Setup** — Create a wallet for the agent, configure the SDK client
2. **Channel Management** — Open a channel, deposit funds for the agent to use
3. **Automated Payments** — Implement a payment function the agent can call autonomously
4. **Session Key Delegation** — Set up a session key with spending caps for security
5. **Agent-to-Agent Payments** — Transfer funds between two autonom
... (1708 chars total, truncated)
```

## Edge

### K.1: lookup_method("nonexistent") — no crash

**Status:** PASS

```
No method matching "nonexistent" found. Available categories: Other, Node & Queries, Signing, Transactions, Channels, App Sessions
```

### K.2: get_rpc_method("nonexistent") — lists methods

**Status:** PASS

```
Unknown RPC method "nonexistent". Available: ping, get_channels, get_ledger_balances, transfer, create_channel, close_channel, create_app_session, submit_app_state, get_app_sessions, get_app_definition, get_ledger_transactions, resize_channel
```

### K.3: search_api("") — no crash

**Status:** PASS

```
# Search results for ""

## Methods (47 matches)
- `create(wsURL: string,
    stateSigner: StateSigner,
    txSigner: TransactionSigner,
    ...opts: Option[]): Promise<Client>` — Other
- `setHomeBlockchain(asset: string, blockchainId: bigint): Promise<void>` — Node & Queries
- `close(): Promise<void>` — Other
- `waitForClose(): void` — Other
- `signState(state: core.State): Promise<Hex>` — Signing
- `getUserAddress(): Address` — Other
- `validateAndSignState(currentState: core.State, proposedSt
... (1305 chars total, truncated)
```

### K.4: scaffold includes decimal.js

**Status:** PASS

```
# Scaffold: transfer-app

## package.json
```json
{
  "name": "nitrolite-transfer-app",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "start": "npx tsx src/index.ts",
    "build": "tsc",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@yellow-org/sdk": "^1.2.0",
    "decimal.js": "^10.4.0",
    "viem": "^2.46.0"
  },
  "devDependencies": {
    "typescript": "^5.7.0",
    "tsx": "^4.19.0",
    "@types/node": "^22.0.0"
  }
}
```

## tsconfig.json
```
... (2002 chars total, truncated)
```

## Sweep

### L.1: resource nitrolite://examples/channels

**Status:** PASS

```
# Nitrolite SDK — Channel Examples

## Creating a Channel & Depositing

```typescript
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const { stateSigner, txSigner } = createSigners('0xYourPrivateKey...');
const client = await Client.create(
    'wss://clearnode.example.com/ws',
    stateSigner,
    txSigner,
    withBlockchainRPC(80002n, 'https://rpc.amoy.polygon.technology'),
);

// Deposit creates channel if needed
const state = a
... (1144 chars total, truncated)
```

### L.2: resource nitrolite://examples/transfers

**Status:** PASS

```
# Nitrolite SDK — Transfer Examples

## Simple Transfer

```typescript
import Decimal from 'decimal.js';

const state = await client.transfer('0xRecipient...', 'usdc', new Decimal('5.0'));
console.log('Transfer tx ID:', state.transition.txId);
```

## Using the Compat Layer

```typescript
import { NitroliteClient } from '@yellow-org/sdk-compat';

const client = await NitroliteClient.create({
    wsURL: 'wss://clearnode.example.com/ws',
    walletClient,
    chainId: 11155111,
    blockchainRPCs:
... (628 chars total, truncated)
```

### L.3: resource nitrolite://protocol/rpc-methods

**Status:** PASS

```
# V1 RPC Methods

All v1 RPC methods defined in `docs/api.yaml`. Methods use grouped naming: `{group}.v1.{method}`.

## channels

| Method | Description | Request Fields | Response Fields |
|---|---|---|---|
| `channels.v1.get_home_channel` | Retrieve current on-chain home channel information | wallet, asset | channel |
| `channels.v1.get_escrow_channel` | Retrieve current on-chain escrow channel information | escrow_channel_id | channel |
| `channels.v1.get_channels` | Retrieve all channels for
... (4628 chars total, truncated)
```

### L.4: resource nitrolite://protocol/auth-flow

**Status:** PASS

```
# Request Signing & Authorization

In v1, every RPC request includes a `sig` field — the client's signature over the entire `req` tuple. This is the authorization mechanism. There is no separate authentication handshake; request signatures are the identity proof.

## Session Keys

Session keys enable delegated signing with scoped permissions. They are managed via:
- `channels.v1.submit_session_key_state` — register/update channel session keys
- `app_sessions.v1.submit_session_key_state` — regist
... (1288 chars total, truncated)
```

### L.5: resource nitrolite://protocol/cryptography

**Status:** PASS

```
# Cryptography

Previous: [Terminology](terminology.md) | Next: [State Model](state-model.md)

---

This document defines how protocol objects are encoded, hashed, and signed.

All rules are described as algorithms and canonical procedures, independent of any specific programming language.

## Purpose

Cryptography in the Nitrolite protocol serves three functions:

1. **Authentication** — proving that a specific participant authorized a state update
2. **Integrity** — ensuring that signed data h
... (6153 chars total, truncated)
```

### L.6: resource nitrolite://protocol/channel-lifecycle

**Status:** PASS

```
# Channel Protocol

Previous: [State Model](state-model.md) | Next: [Enforcement and Settlement](enforcement.md)

---

This document describes how channels operate and how states evolve through off-chain state advancement.

## Purpose

Channels are the primary mechanism for off-chain interaction in the Nitrolite protocol. They allow participants to exchange assets and update state without on-chain transactions.

## Channel Definition

A channel is defined by a set of immutable parameters fixed a
... (16448 chars total, truncated)
```

### L.7: resource nitrolite://protocol/state-model

**Status:** PASS

```
# State Model

Previous: [Cryptography](cryptography.md) | Next: [Channel Protocol](channel-protocol.md)

---

This document describes the abstract structure of protocol states.

It explains how states are defined and structured. Operational flows are described in separate documents.

## Purpose

States represent the current agreed configuration of protocol entities. The state model defines:

- what information a state contains
- how states are identified and versioned
- how states are represent
... (10458 chars total, truncated)
```

### L.8: resource nitrolite://protocol/enforcement

**Status:** PASS

```
# State Enforcement

Previous: [Channel Protocol](channel-protocol.md) | Next: [Cross-Chain and Assets](cross-chain-and-assets.md)

---

This document describes how channel states are enforced on the blockchain layer.

## Purpose

Enforcement is the mechanism by which off-chain state is reflected on-chain. It serves two complementary roles:

1. **Regular state synchronization** — participants submit signed states to the blockchain layer to keep the on-chain record up-to-date with the latest off-
... (13071 chars total, truncated)
```

### L.9: resource nitrolite://protocol/cross-chain

**Status:** PASS

```
# Cross-Chain and Asset Model

Previous: [Enforcement and Settlement](enforcement.md) | Next: [Interactions](interactions.md)

---

This document describes the unified asset model and cross-chain functionality.

## Purpose

The unified asset model allows participants to operate on assets from multiple blockchains within a single channel. This eliminates the need for separate channels per blockchain and enables cross-chain interactions.

## Unified Asset Concept

Assets in the Nitrolite protocol 
... (10465 chars total, truncated)
```

### L.10: resource nitrolite://protocol/interactions

**Status:** PASS

```
# Interaction Model

Previous: [Cross-Chain and Assets](cross-chain-and-assets.md) | Next: [Security and Limitations](security-and-limitations.md)

---

This document defines the logical communication protocol between participants.

All operations are defined as semantic protocol operations, independent of transport technologies such as WebSocket or gRPC.

## Purpose

Participants exchange protocol messages to advance state, manage channels, and coordinate operations. This document defines the s
... (5043 chars total, truncated)
```

### L.11: resource nitrolite://use-cases

**Status:** PASS

```
# Nitrolite Use Cases

What you can build with the Nitrolite SDK and state channels.

## Peer-to-Peer Payments
Instant, gas-free token transfers between users. Open a channel, transfer any amount instantly, settle on-chain only when needed.
**SDK methods:** `client.deposit()`, `client.transfer()`, `client.closeHomeChannel()`

## Gaming (Real-Time Wagering)
Turn-based or real-time games where players wager tokens. App sessions track game state; winners receive payouts automatically.
**SDK methods
... (2039 chars total, truncated)
```

### L.17: lookup_rpc_method alias for transfer

**Status:** PASS

```
No v1 RPC method matching "transfer". Available methods:
channels.v1.get_home_channel, channels.v1.get_escrow_channel, channels.v1.get_channels, channels.v1.get_latest_state, channels.v1.get_states, channels.v1.request_creation, channels.v1.submit_state, channels.v1.submit_session_key_state, channels.v1.get_last_key_states, channels.v1.home_channel_created, app_sessions.v1.submit_deposit_state, app_sessions.v1.submit_app_state, app_sessions.v1.rebalance_app_sessions, app_sessions.v1.get_app_defi
... (834 chars total, truncated)
```
