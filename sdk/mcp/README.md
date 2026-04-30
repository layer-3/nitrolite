# Yellow SDK MCP Server

An MCP (Model Context Protocol) server that exposes the **Yellow SDK / Nitrolite SDK** knowledge base to AI coding tools — covering both the **TypeScript SDK** (`@yellow-org/sdk`) and the **Go SDK** (`github.com/layer-3/nitrolite/sdk/go`).

The npm package ships with a release-time content snapshot, so external users do **not** need to clone the Nitrolite repository. When running from this monorepo, the server reads local source files first and falls back to the packaged snapshot only when source files are unavailable.

## Quick Start

Choose the client you use. All published-package examples run the server from npm, so no Nitrolite repo clone is required.

### Claude Code

For your current project:

```bash
claude mcp add --transport stdio nitrolite -- npx -y @yellow-org/sdk-mcp@^1
```

For a shareable project config, add this to `.mcp.json` at the repository root:

```json
{
  "mcpServers": {
    "nitrolite": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@yellow-org/sdk-mcp@^1"]
    }
  }
}
```

### Claude Desktop

Add this to `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS, or `%APPDATA%\Claude\claude_desktop_config.json` on Windows, then restart Claude Desktop:

```json
{
  "mcpServers": {
    "nitrolite": {
      "command": "npx",
      "args": ["-y", "@yellow-org/sdk-mcp@^1"]
    }
  }
}
```

### Codex

```bash
codex mcp add nitrolite -- npx -y @yellow-org/sdk-mcp@^1
```

### Cursor

Add this to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "nitrolite": {
      "command": "npx",
      "args": ["-y", "@yellow-org/sdk-mcp@^1"]
    }
  }
}
```

### VS Code

Add this to `.vscode/mcp.json` in your workspace, or to your VS Code user MCP config:

```json
{
  "servers": {
    "nitrolite": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@yellow-org/sdk-mcp@^1"]
    }
  }
}
```

You can also install it from the VS Code command line:

```bash
code --add-mcp '{"name":"nitrolite","type":"stdio","command":"npx","args":["-y","@yellow-org/sdk-mcp@^1"]}'
```

### Local Development

From this repository:

```bash
cd sdk/mcp && npm install && cd ../..
```

Add this to `.mcp.json`:

```json
{
  "mcpServers": {
    "nitrolite": {
      "type": "stdio",
      "command": "npm",
      "args": ["--prefix", "sdk/mcp", "exec", "--", "tsx", "sdk/mcp/src/index.ts"]
    }
  }
}
```

Any MCP-compatible client that supports stdio servers can launch the package with the same `npx -y @yellow-org/sdk-mcp@^1` command.

## What's Inside

- **30 resources** — API reference (TS + Go), protocol docs, security patterns, examples (TS + Go), use cases, migration
- **9 tools** — server info, method lookup, type lookup, search, RPC format, import validation, concept explanation, scaffolding (TS + Go)
- **3 prompts** — guided workflows for building apps (covers both TS and Go), migrating from v0.5.3, building AI agents

## Tools

### `server_info`

Return package, SDK, compat, Go module, protocol, and content mode metadata. Use this in bug reports.

```text
> server_info()
```

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

## Publishing

The package is designed for npm distribution:

```bash
cd sdk/mcp
npm ci
npm run build
npm pack --dry-run
npm publish --access public
```

`npm run build` copies the release source/docs snapshot into `content/`, compiles to `dist/`, and makes the binary executable. The npm tarball includes only `dist`, `content`, `README.md`, `package.json`, and `server.json`.

Release tags matching `mcp-v*` run `.github/workflows/publish-sdk-mcp.yml`, which publishes the npm package and then publishes `server.json` to the MCP Registry.

Version policy: the MCP package mirrors the SDK release it documents. If `@yellow-org/sdk` is `1.2.1`, publish `@yellow-org/sdk-mcp@1.2.1` from the same release commit. Consumers can use `@^1` to track compatible v1 SDK docs, or pin an exact MCP version for audited builds.

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
