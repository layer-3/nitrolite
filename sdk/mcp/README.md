# Nitrolite MCP Server (Unified)

An MCP (Model Context Protocol) server that exposes the **Nitrolite SDK** knowledge base to AI coding tools — covering both the **TypeScript SDK** (`@yellow-org/sdk`) and the **Go SDK** (`github.com/layer-3/nitrolite/sdk/go`). It reads SDK source code, protocol documentation, and Go type definitions at startup, making every method, type, enum, and protocol concept discoverable by AI agents.

> **Renamed in v1.3.0** — the off-chain broker was renamed from `clearnode` to `nitronode`. v1.2.0 and earlier ship as `clearnode`; v1.3.0 and later ship as `nitronode`. See [`MIGRATION-NITRONODE.md`](../../MIGRATION-NITRONODE.md) (also exposed as the `nitrolite://migration/nitronode` MCP resource).

## Quick Start

```bash
# From repo root:
cd sdk/mcp && npm install && cd ../..
```

Add to `.mcp.json` (already configured in this repo):

```json
{
  "mcpServers": {
    "nitrolite": {
      "command": "npm",
      "args": ["--prefix", "sdk/mcp", "exec", "--", "tsx", "sdk/mcp/src/index.ts"]
    }
  }
}
```

Any MCP-compatible tool (Claude Code, Cursor, Windsurf, VS Code Copilot) auto-discovers the server from `.mcp.json`.

## What's Inside

- **30 resources** — API reference (TS + Go), protocol docs, security patterns, examples (TS + Go), use cases, migration
- **8 tools** — method lookup, type lookup, search, RPC format, import validation, concept explanation, scaffolding (TS + Go)
- **3 prompts** — guided workflows for building apps (covers both TS and Go), migrating from v0.5.3, building AI agents

---

## Tools

### `lookup_method`

Look up a Client method by name. Supports an optional `language` parameter.

```text
> lookup_method({ name: "transfer" })                    # TypeScript (default)
> lookup_method({ name: "Transfer", language: "go" })    # Go SDK
> lookup_method({ name: "transfer", language: "both" })  # Both SDKs
```

### `lookup_type`

Look up a type, interface, struct, or enum by name.

```text
> lookup_type({ name: "AppDefinitionV1" })                          # TypeScript (default)
> lookup_type({ name: "AppSessionV1", language: "go" })             # Go (from pkg/app)
> lookup_type({ name: "ChannelStatus", language: "go" })            # Go enum with values
```

### `search_api`

Fuzzy search across methods and types.

```text
> search_api({ query: "transfer" })                        # TypeScript (default)
> search_api({ query: "session", language: "go" })         # Go methods + types
> search_api({ query: "channel", language: "both" })       # Both SDKs
```

### `scaffold_project`

Generate a starter project. TypeScript templates output `package.json` + `tsconfig.json` + `src/index.ts`. Go templates output `go.mod` + `main.go`.

| Template | Language | Description |
|----------|----------|-------------|
| `transfer-app` | TypeScript | Deposit, transfer, close channel |
| `app-session` | TypeScript | Multi-party app session |
| `ai-agent` | TypeScript | Autonomous payment agent |
| `go-transfer-app` | Go | Deposit, transfer, close channel |
| `go-app-session` | Go | Multi-party app session |
| `go-ai-agent` | Go | Autonomous payment agent with graceful shutdown |

### Other Tools

- `get_rpc_method` — 0.5.x compat method to v1 wire format
- `validate_import` — check if a symbol is in `@yellow-org/sdk-compat` or `@yellow-org/sdk`
- `explain_concept` — plain-English explanation of protocol concepts
- `lookup_rpc_method` — full v1 RPC method lookup from `docs/api.yaml`

---

## Resources

### TypeScript SDK
- `nitrolite://api/methods` — TS client methods by category
- `nitrolite://api/types` — TS interfaces and type aliases
- `nitrolite://api/enums` — TS enums

### Go SDK
- `nitrolite://go-api/methods` — Go client methods by category
- `nitrolite://go-api/types` — Go structs and enum types (from `pkg/` and `sdk/go/`)

### Examples
- `nitrolite://examples/full-transfer-script` — complete TypeScript transfer script
- `nitrolite://examples/full-app-session-script` — complete TypeScript app session script
- `nitrolite://go-examples/full-transfer-script` — complete Go transfer script
- `nitrolite://go-examples/full-app-session-script` — complete Go app session script

### Protocol & Security
- `nitrolite://protocol/{overview,terminology,cryptography,wire-format,rpc-methods,auth-flow,channel-lifecycle,state-model,enforcement,cross-chain,interactions}`
- `nitrolite://security/{overview,app-session-patterns,state-invariants}`

### Use Cases
- `nitrolite://use-cases`, `nitrolite://use-cases/ai-agents`

### Migration
- `nitrolite://migration/overview`

---

## Prompts

- `create-channel-app` — step-by-step guide covering both TypeScript and Go SDKs
- `migrate-from-v053` — migration guide from `@layer-3/nitrolite` v0.5.3 to the compat layer
- `build-ai-agent-app` — AI agent payments guide for both TypeScript and Go
