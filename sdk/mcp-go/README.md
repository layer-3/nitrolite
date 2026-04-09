# Nitrolite Go SDK MCP Server

An MCP (Model Context Protocol) server that exposes the Nitrolite protocol and Go SDK knowledge base to AI coding tools. It reads protocol documentation, clearnode API specs, and Go SDK source at startup, making every method and protocol concept discoverable by AI agents.

## Quick Start

```bash
# From repo root:
cd sdk/mcp-go && go build . && cd ../..

# Add to .mcp.json (already configured in this repo):
```

```json
{
  "mcpServers": {
    "nitrolite-go": {
      "command": "go",
      "args": ["run", "./sdk/mcp-go"]
    }
  }
}
```

Any MCP-compatible tool (Claude Code, Cursor, Windsurf, VS Code Copilot) auto-discovers the server from `.mcp.json`.

## What's Inside

- **19 resources** — API reference, protocol docs, security patterns, examples, use cases, clearnode docs
- **5 tools** — method lookup, search, RPC lookup, concept explanation, scaffolding
- **2 prompts** — guided workflows for building channel apps and AI agents

---

## Tools — Expected Output

### `lookup_method`

Look up a specific Go SDK Client method by name. Returns signature, category, and GoDoc description.

```
> lookup_method({ name: "Deposit" })

## Deposit
**Signature:**
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string,
    amount decimal.Decimal) (*core.State, error)
**Category:** Channels & Transactions
**Description:** Deposit prepares a deposit state for the user's channel. This method
handles two scenarios automatically:
  1. If no channel exists: Creates a new channel with the initial deposit
  2. If channel exists: Advances the state with a deposit transition
```

```
> lookup_method({ name: "Transfer" })

## Transfer
**Signature:**
func (c *Client) Transfer(ctx context.Context, recipientWallet string, asset string,
    amount decimal.Decimal) (*core.State, error)
**Category:** Channels & Transactions
**Description:** Transfer prepares a transfer state to send funds to another wallet address.
```

```
> lookup_method({ name: "CreateAppSession" })

## CreateAppSession
**Signature:**
func (c *Client) CreateAppSession(ctx context.Context, definition app.AppDefinitionV1,
    sessionData string, quorumSigs []string,
    opts ...CreateAppSessionOptions) (string, string, string, error)
**Category:** App Sessions
**Description:** Creates a new application session between participants.
```

```
> lookup_method({ name: "SubmitAppState" })

## SubmitAppState
**Signature:**
func (c *Client) SubmitAppState(ctx context.Context,
    appStateUpdate app.AppStateUpdateV1, quorumSigs []string) error
**Category:** App Sessions
**Description:** Submits an app session state update. Handles operate, withdraw, and
close intents. For deposits, use SubmitAppSessionDeposit instead.
```

```
> lookup_method({ name: "Checkpoint" })

## Checkpoint
**Signature:**
func (c *Client) Checkpoint(ctx context.Context, asset string) (string, error)
**Category:** Channels & Transactions
**Description:** Executes the blockchain transaction for the latest signed state.
This is the only method that interacts with the blockchain. It should be called
after any state-building method (deposit, withdraw, closeHomeChannel, etc.).
```

```
> lookup_method({ name: "CloseAppSession" })

No method matching "closeappsession" found. 45 methods indexed.
```

### `search_api`

Fuzzy search across all Go SDK methods.

```
> search_api({ query: "channel" })

# Search results for "channel"

## Methods
- `Deposit` — Channels & Transactions
  func (c *Client) Deposit(ctx context.Context, blockchainID uint64, ...) (*core.State, error)
- `Transfer` — Channels & Transactions
  func (c *Client) Transfer(ctx context.Context, ...) (*core.State, error)
- `CloseHomeChannel` — Channels & Transactions
  func (c *Client) CloseHomeChannel(ctx context.Context, asset string) (*core.State, error)
- `Checkpoint` — Channels & Transactions
  ...
```

```
> search_api({ query: "zzzzz" })

No methods matching "zzzzz". Try broader terms.
```

### `lookup_rpc_method`

Look up a clearnode RPC method. Returns description and access level.

```
> lookup_rpc_method({ method: "auth_request" })

## RPC: `auth_request`
**Description:** Initiates authentication with the server
**Access:** Public
```

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

### `explain_concept`

Plain-English explanation of a Nitrolite protocol concept.

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
> explain_concept({ concept: "app session" })

No exact match for "app session". Related:

**Application Session**
An extension that enables off-chain application functionality.
Application sessions hold committed assets and maintain their own state.

**Application State**
The state associated with an application session, tracking committed assets
and application-specific data.
```

```
> explain_concept({ concept: "made up thing" })

No concept matching "made up thing" found. 34 concepts indexed.
```

### `scaffold_project`

Generate a starter Go project. Templates: `transfer-app`, `app-session`, `ai-agent`.

```
> scaffold_project({ template: "transfer-app" })

# Scaffold: transfer-app

## go.mod
module my-nitrolite-transfer-app
require (
    github.com/layer-3/nitrolite v0.0.0
    github.com/shopspring/decimal v1.4.0
)

## main.go
stateSigner, _ := sign.NewEthereumMsgSigner(os.Getenv("PRIVATE_KEY"))
txSigner, _ := sign.NewEthereumRawSigner(os.Getenv("PRIVATE_KEY"))

client, err := sdk.NewClient(os.Getenv("CLEARNODE_URL"), stateSigner, txSigner, ...)
// Deposit + checkpoint on-chain
_, err = client.Deposit(ctx, 11155111, "usdc", decimal.NewFromInt(10))
txHash, err := client.Checkpoint(ctx, "usdc")
// Transfer
_, err = client.Transfer(ctx, recipient, "usdc", decimal.NewFromInt(5))
```

```
> scaffold_project({ template: "app-session" })

## main.go
def := app.AppDefinitionV1{
    ApplicationID: "my-app",
    Participants: []app.AppParticipantV1{...},
    Quorum: 100,
    Nonce:  1,
}

// Collect quorum signatures from participants (off-band signing)
quorumSigs := []string{"0xMySig...", "0xPeerSig..."}

sessionID, versionStr, _, err := client.CreateAppSession(ctx, def, "{}", quorumSigs)
initVersion, _ := strconv.ParseUint(versionStr, 10, 64)

// Submit state update (version = initial + 1)
update := app.AppStateUpdateV1{
    AppSessionID: sessionID,
    Intent:       app.AppStateUpdateIntentOperate,
    Version:      initVersion + 1,
    ...
}
client.SubmitAppState(ctx, update, operateSigs)

// Close session — submit with Close intent (version incremented)
update.Intent = app.AppStateUpdateIntentClose
update.Version = initVersion + 2
client.SubmitAppState(ctx, update, closeSigs)
```

```
> scaffold_project({ template: "nonexistent" })

Unknown template "nonexistent". Available: transfer-app, app-session, ai-agent
```

---

## Resources

### API Reference
| URI | Content |
|-----|---------|
| `nitrolite://api/methods` | All 44 Go SDK Client methods with signatures, organized by category |

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
| `nitrolite://protocol/enforcement` | On-chain dispute resolution and checkpoints |

### Security
| URI | Content |
|-----|---------|
| `nitrolite://security/overview` | Security guarantees and trust assumptions |
| `nitrolite://security/app-session-patterns` | Quorum design, challenge periods |
| `nitrolite://security/state-invariants` | Fund conservation, version ordering, signature rules |

### Examples
| URI | Content |
|-----|---------|
| `nitrolite://examples/full-transfer-script` | Complete Go transfer: connect, deposit, checkpoint, transfer, close |
| `nitrolite://examples/full-app-session-script` | Complete Go app session: create with quorum sigs, versioned updates, close via intent |

### Clearnode
| URI | Content |
|-----|---------|
| `nitrolite://clearnode/entities` | Channel, State, AppSession entity schemas |
| `nitrolite://clearnode/session-keys` | Session key delegation docs |
| `nitrolite://clearnode/protocol` | Clearnode protocol specification |

### Use Cases
| URI | Content |
|-----|---------|
| `nitrolite://use-cases` | Payments, gaming, escrow, AI agents, streaming |
| `nitrolite://use-cases/ai-agents` | Agent payments, session keys, shutdown patterns |

## Prompts

| Prompt | Expected Output |
|--------|-----------------|
| `create-channel-app` | Step-by-step guide covering: Setup, Client Creation, Channel Lifecycle (Deposit → Transfer → Checkpoint → CloseHomeChannel + Checkpoint), App Sessions (CreateAppSession, SubmitAppState with Operate/Withdraw/Close intents), Error Handling, Testing |
| `build-ai-agent-app` | Guided conversation covering: Agent Wallet Setup, Channel Management, Automated Payments, Session Key Delegation, Agent-to-Agent Payments, Framework Integration, Error Handling |

## Architecture

Single Go file (`main.go`) that at startup reads and indexes:
- Go SDK client methods from `sdk/go/*.go` (regex-based parsing of all 44 exported Client methods)
- Protocol docs from `docs/protocol/*.md`
- Protocol terminology into a concept-definition map
- Clearnode API methods from `clearnode/docs/API.md`

Transport: stdio (JSON-RPC over stdin/stdout). Uses [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go). No network calls, no runtime SDK imports.

## Key Behaviors

- **`CloseAppSession` does not exist.** Closing an app session is done by submitting a state update with `AppStateUpdateIntentClose` via `SubmitAppState`. The MCP server correctly returns "not found" for this method.
- **`CloseHomeChannel` requires `Checkpoint`.** The full close flow is `CloseHomeChannel` (prepares finalize state) → `Checkpoint` (submits on-chain). All examples and scaffolds show this two-step pattern.
- **App session updates are versioned.** `CreateAppSession` returns an initial version; subsequent `SubmitAppState` calls must increment it. Scaffolds show `initVersion + 1` for operate, `initVersion + 2` for close.
- **Quorum signatures are required.** `CreateAppSession` and `SubmitAppState` both require `quorumSigs []string` — participants sign off-band and signatures are collected before submission.

## Testing

Build and verify the server starts:

```bash
cd sdk/mcp-go && go build . && echo 'Build OK'
```
