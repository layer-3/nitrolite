# Nitrolite SDK MCP Server (TypeScript)

An MCP (Model Context Protocol) server that exposes the **Nitrolite TypeScript SDK** (`@yellow-org/sdk`) knowledge base to AI coding tools. It reads TS SDK source code, protocol documentation, and clearnode API specs at startup, making every method, type, enum, and protocol concept discoverable by AI agents. For the Go SDK equivalent, see `sdk/mcp-go/`.

## Quick Start

```bash
# From repo root:
cd sdk/ts-mcp && npm install && cd ../..

# Add to .mcp.json (already configured in this repo):
```

```json
{
  "mcpServers": {
    "nitrolite-ts": {
      "command": "npm",
      "args": ["--prefix", "sdk/ts-mcp", "exec", "--", "tsx", "sdk/ts-mcp/src/index.ts"]
    }
  }
}
```

Any MCP-compatible tool (Claude Code, Cursor, Windsurf, VS Code Copilot) auto-discovers the server from `.mcp.json`.

## What's Inside

- **25 resources** — API reference, protocol docs, security patterns, examples, use cases, clearnode docs
- **8 tools** — method lookup, type lookup, search, RPC format, import validation, concept explanation, scaffolding
- **3 prompts** — guided workflows for building apps, migrating from v0.5.3, building AI agents

---

## Tools — Expected Output

### `lookup_method`

Look up a specific SDK Client method by name. Returns signature, category, and JSDoc description.

```
> lookup_method({ name: "transfer" })

## transfer
**Signature:** `transfer(recipientWallet: string, asset: string, amount: Decimal): Promise<core.State>`
**Category:** Transactions
**Description:** Transfer prepares a transfer state to send funds to another wallet address.
This method handles two scenarios automatically:
1. If no channel exists: Creates a new channel with the transfer transition
2. If channel exists: Advances the state with a transfer send transition
```

```
> lookup_method({ name: "create" })

## create
**Signature:** `create(wsURL: string, stateSigner: StateSigner, txSigner: TransactionSigner,
    ...opts: Option[]): Promise<Client>`
**Category:** Other
**Description:** Main Nitrolite SDK Client. Provides a unified interface for interacting
with Nitrolite payment channels.
```

```
> lookup_method({ name: "closeAppSession" })

No method matching "closeAppSession" found.
Available categories: Other, Node & Queries, Signing, Transactions, Channels, App Sessions
```

### `lookup_type`

Look up a type, interface, or enum definition. Returns fields and source file.

```
> lookup_type({ name: "AppDefinitionV1" })

## AppDefinitionV1 (interface) — sdk/ts (app)
applicationId: string;
participants: AppParticipantV1[];
quorum: number;
nonce: bigint;
```

```
> lookup_type({ name: "AppStateUpdateV1" })

## AppStateUpdateV1 (interface) — sdk/ts (app)
appSessionId: string;
intent: AppStateUpdateIntent;
version: bigint;
allocations: AppAllocationV1[];
sessionData: string;
```

### `search_api`

Fuzzy search across all SDK methods and types.

```
> search_api({ query: "transfer" })

# Search results for "transfer"

## Methods (5 matches)
- `transfer(recipientWallet: string, asset: string, amount: Decimal): Promise<core.State>` — Transactions
- `create(wsURL: string, stateSigner: StateSigner, ...): Promise<Client>` — Other
- ...

## Types (3 matches)
- `TransferAllocation` (interface) — sdk-compat
- ...
```

```
> search_api({ query: "app session" })

# Search results for "app session"

## Methods (8 matches)
- `createAppSession(definition: app.AppDefinitionV1, sessionData: string, quorumSigs: string[], ...)`
- `submitAppState(appStateUpdate: app.AppStateUpdateV1, quorumSigs: string[]): Promise<void>`
- `submitAppSessionDeposit(appStateUpdate: app.AppStateUpdateV1, quorumSigs: string[], ...)`
- `getAppSessions(options?): Promise<{ sessions: AppSessionInfoV1[]; ... }>`
- ...
```

### `get_rpc_method`

Get the RPC wire format for a clearnode method. Matches `clearnode/docs/API.md`.

```
> get_rpc_method({ method: "auth_request" })

## RPC: auth_request

**V1 Wire Method:** `auth_request`

**Request format (v0.5.3 compat):**
{ req: [requestId, "auth_request", { address, session_key?, application?,
  allowances?, scope?, expires_at? }, timestamp], sig: [...] }

**Response format:**
{ res: [requestId, "auth_challenge", { challenge_message }] }
```

```
> get_rpc_method({ method: "create_channel" })

## RPC: create_channel

**V1 Wire Method:** `channels.v1.request_creation`

**Request format (v0.5.3 compat):**
{ req: [requestId, "create_channel", [{ chain_id, token }], timestamp], sig: [...] }

**Response format:**
{ res: [requestId, "create_channel", [{ channel_id, channel, state, server_signature }], ...] }
```

```
> get_rpc_method({ method: "resize_channel" })

## RPC: resize_channel

**Request format (v0.5.3 compat):**
{ req: [requestId, "resize_channel", { channel_id, resize_amount,
  allocate_amount, funds_destination }, timestamp], sig: [...] }
```

### `lookup_rpc_method`

Look up a clearnode RPC method by name. Returns description and access level (from `clearnode/docs/API.md`). Distinct from `get_rpc_method` which returns the wire format.

```
> lookup_rpc_method({ method: "create_channel" })

## RPC: `create_channel`
**Description:** Returns data and Broker signature to open a channel
**Access:** Private
```

```
> lookup_rpc_method({ method: "transfer" })

## RPC: `transfer`
**Description:** Transfers funds from user's unified balance to another account
**Access:** Private
```

### `validate_import`

Check if a symbol is exported from `@yellow-org/sdk-compat` or `@yellow-org/sdk`.

```
> validate_import({ symbol: "NitroliteClient" })

**NitroliteClient** is exported from `@yellow-org/sdk-compat`.

import { NitroliteClient } from '@yellow-org/sdk-compat';
```

```
> validate_import({ symbol: "Client" })

**Client** is NOT in `@yellow-org/sdk-compat` but IS in `@yellow-org/sdk`.

import { Client } from '@yellow-org/sdk';

> Note: SDK classes should not be re-exported from compat (SSR risk).
```

```
> validate_import({ symbol: "Cli" })

**Cli** was not found in either `@yellow-org/sdk-compat` or `@yellow-org/sdk` barrel exports.
```

### `explain_concept`

Plain-English explanation of a Nitrolite protocol concept. Draws from `docs/protocol/terminology.md`.

```
> explain_concept({ concept: "state channel" })

**State**
An abstract data structure representing the current configuration of a protocol entity
at a specific version.

**Channel**
A state container shared between a user and a node that allows off-chain state updates
while maintaining on-chain security guarantees.
```

```
> explain_concept({ concept: "made up thing" })

No concept matching "made up thing" found. 34 concepts indexed.
```

### `scaffold_project`

Generate a starter project. Templates: `transfer-app`, `app-session`, `ai-agent`.

```
> scaffold_project({ template: "transfer-app" })

# Scaffold: transfer-app

## package.json
{ "dependencies": { "@yellow-org/sdk": "^1.2.0", "decimal.js": "^10.4.0", "viem": "^2.46.0" } }

## src/index.ts
import { Client, createSigners, withBlockchainRPC } from '@yellow-org/sdk';
import Decimal from 'decimal.js';

const { stateSigner, txSigner } = createSigners(PRIVATE_KEY);
const client = await Client.create(CLEARNODE_URL, stateSigner, txSigner, ...);

await client.deposit(CHAIN_ID, 'usdc', new Decimal(10));
await client.checkpoint('usdc');
await client.transfer(recipient, 'usdc', new Decimal(5));
```

```
> scaffold_project({ template: "app-session" })

## src/index.ts
import { Client, createSigners, withBlockchainRPC, app } from '@yellow-org/sdk';

const definition: app.AppDefinitionV1 = {
    applicationId: 'my-app',
    participants: [...],
    quorum: 100,
    nonce: BigInt(Date.now()),
};
const session = await client.createAppSession(definition, '{}', quorumSigs);
await client.submitAppState(update, ['0xMySig...', '0xPeerSig...']);

// Close via submitAppState with Close intent (no closeAppSession method)
```

---

## Resources

### API Reference
| URI | Content |
|-----|---------|
| `nitrolite://api/methods` | All 46+ SDK Client methods with signatures, organized by category |
| `nitrolite://api/types` | All interfaces, type aliases from core, RPC, app, and compat |
| `nitrolite://api/enums` | All enums with values |

### Protocol
| URI | Content |
|-----|---------|
| `nitrolite://protocol/overview` | State channels, design goals, system roles |
| `nitrolite://protocol/terminology` | Canonical definitions of all protocol terms |
| `nitrolite://protocol/wire-format` | Message envelope structure |
| `nitrolite://protocol/rpc-methods` | All clearnode RPC methods (from API.md) |
| `nitrolite://protocol/auth-flow` | Challenge-response authentication |
| `nitrolite://protocol/channel-lifecycle` | Channel states and transitions |
| `nitrolite://protocol/state-model` | State structure, versioning, invariants |

### Security
| URI | Content |
|-----|---------|
| `nitrolite://security/overview` | Security guarantees and trust assumptions |
| `nitrolite://security/app-session-patterns` | Quorum design, challenge periods |
| `nitrolite://security/state-invariants` | Fund conservation, version ordering, signature rules |

### Examples
| URI | Content |
|-----|---------|
| `nitrolite://examples/channels` | Create, query, close channels |
| `nitrolite://examples/transfers` | Transfers with SDK and compat layer |
| `nitrolite://examples/app-sessions` | Create sessions, submit state, close |
| `nitrolite://examples/auth` | EIP-712 auth flow |
| `nitrolite://examples/full-transfer-script` | Complete end-to-end transfer script |
| `nitrolite://examples/full-app-session-script` | Complete multi-party app session |

### Clearnode
| URI | Content |
|-----|---------|
| `nitrolite://clearnode/entities` | Channel, State, AppSession entity schemas |
| `nitrolite://clearnode/session-keys` | Session key delegation docs |
| `nitrolite://clearnode/protocol` | Clearnode protocol specification |

### Use Cases & Migration
| URI | Content |
|-----|---------|
| `nitrolite://use-cases` | Payments, gaming, escrow, DeFi, streaming |
| `nitrolite://use-cases/ai-agents` | Agent payments, session keys, frameworks |
| `nitrolite://migration/overview` | v0.5.3 to v1.x migration guide |

## Prompts

| Prompt | Description |
|--------|-------------|
| `create-channel-app` | Step-by-step guide to build a state channel app |
| `migrate-from-v053` | Interactive migration assistant from v0.5.3 |
| `build-ai-agent-app` | Guided conversation for building an AI payment agent |

## Architecture

Single TypeScript file (`src/index.ts`) that at startup reads and indexes:
- SDK client methods from `sdk/ts/src/client.ts` (regex-based static analysis)
- Types from `sdk/ts/src/core/types.ts`, `sdk/ts/src/rpc/types.ts`, `sdk/ts/src/app/types.ts`, `sdk/ts-compat/src/types.ts`
- Compat barrel exports from `sdk/ts-compat/src/index.ts`
- Protocol docs from `docs/protocol/*.md`
- Clearnode API from `clearnode/docs/API.md`

Transport: stdio (JSON-RPC over stdin/stdout). No network calls, no runtime SDK imports.

## Testing

Start the server manually to verify it's working:

```bash
# Should print to stderr: "Nitrolite SDK MCP server running on stdio"
npm --prefix sdk/ts-mcp exec -- tsx sdk/ts-mcp/src/index.ts
```
