# Nitrolite

Nitrolite is a state channel framework for Ethereum and EVM-compatible blockchains. It enables off-chain interactions (instant finality, low gas) while maintaining on-chain security guarantees.

## Repository Structure

| Directory | Description | Language |
|-----------|-------------|----------|
| `contracts/` | Solidity smart contracts (ChannelHub, ChannelEngine) | Solidity (Foundry) |
| `clearnode/` | Off-chain broker: ledger services, WebSocket JSON-RPC | Go |
| `sdk/ts/` | TypeScript SDK (`@yellow-org/sdk`) | TypeScript |
| `sdk/ts-compat/` | Compat layer (`@yellow-org/sdk-compat`) bridging v0.5.3 API to v1.0.0+ | TypeScript |
| `sdk/go/` | Go SDK for backend integrations | Go |
| `sdk/ts-mcp/` | TypeScript MCP server exposing SDK API surface to AI agents/IDEs | TypeScript |
| `sdk/mcp-go/` | Go MCP server exposing Go SDK API surface to AI agents/IDEs | Go |
| `cerebro/` | Interactive CLI for channel/asset management | Go |
| `pkg/` | Shared Go packages (core, sign, rpc, app, blockchain, log) | Go |
| `docs/` | Protocol specifications, architecture docs | Markdown |
| `test/integration/` | Integration tests against a live clearnode | TypeScript |

See stack-specific `CLAUDE.md` files in `sdk/ts/`, `sdk/ts-compat/`, and `sdk/go/` for detailed conventions.

## Build & Test Commands

### TypeScript SDK
```bash
cd sdk/ts && npm install                 # Install dependencies (first time)
cd sdk/ts && npm test                    # Unit tests (Jest)
cd sdk/ts && npm run build               # Tests + compile (runs tests first!)
cd sdk/ts && npm run typecheck           # Type check only
cd sdk/ts && npm run lint                # ESLint
```

### TypeScript SDK Compat
```bash
cd sdk/ts-compat && npm install          # Install dependencies (first time)
cd sdk/ts-compat && npm test             # Unit tests (Jest)
cd sdk/ts-compat && npm run build        # Compile
cd sdk/ts-compat && npm run typecheck    # Type check only
```

### Go SDK
```bash
go test ./sdk/go/... -v                  # SDK tests only (from repo root)
go build ./sdk/go/...                    # Build SDK
go test ./...                            # ALL Go tests (clearnode + pkg + sdk + cerebro)
go vet ./...                             # Lint all Go code
```

### Smart Contracts
```bash
cd contracts && forge build              # Compile
cd contracts && forge test               # Run tests
cd contracts && forge fmt                # Format
```

### Integration Tests
```bash
cd test/integration && npm test          # Requires a running clearnode
```

## Important Notes

- **Go module** is at repo root: `go.mod`, module `github.com/layer-3/nitrolite`, Go 1.25
- **Build order**: `sdk/ts` must build before `sdk/ts-compat` (has `"@yellow-org/sdk": "file:../ts"` dependency)
- **sdk/ts build runs tests first**: `npm run build` = `npm run test && tsc`. Avoid `npm test && npm run build` (double-tests).
- **Foundry** uses git submodules for deps (`forge-std`, `openzeppelin-contracts`). Use `--recurse-submodules` on clone.
- **TS MCP server** (`sdk/ts-mcp/`): run `cd sdk/ts-mcp && npm install` before first use.
- **Go MCP server** (`sdk/mcp-go/`): run `cd sdk/mcp-go && go build .` to verify.
- **Never** edit `.env` files or commit secrets.

## V1 Protocol Source of Truth

- **API definition**: `docs/api.yaml` — canonical list of all v1 RPC methods, types, and request/response schemas
- **Protocol spec**: `docs/protocol/` — state channels, transitions, enforcement, security
- **Contract invariants**: `contracts/SECURITY.md`
- **Contract design**: `contracts/suggested-contract-design.md`, entrypoint `contracts/src/ChannelHub.sol`
- **Clearnode docs**: `clearnode/readme.md`, `clearnode/docs/`

**Do NOT use `clearnode/docs/API.md` as v1 reference** — it documents the 0.5.x compat-layer method names (e.g., `transfer`, `create_channel`, `auth_request`). The v1 methods use grouped names (e.g., `channels.v1.submit_state`, `app_sessions.v1.create_app_session`).

### SDK vs Compat

- **`@yellow-org/sdk`** (`sdk/ts/`) — v1 protocol SDK. Use for all new code.
- **`@yellow-org/sdk-compat`** (`sdk/ts-compat/`) — bridges 0.5.x API surface to v1 runtime. Wraps `Client` with `NitroliteClient`, exposes legacy types and method names. For migration only.
- **`sdk/go/`** — Go v1 SDK. No compat layer exists for Go.

## Commit Convention

```text
feat|fix|chore|test|docs(scope): description

# Examples:
feat(sdk/ts): add transfer batching support
fix(sdk-compat): export missing generateRequestId
chore(contracts): update OpenZeppelin to v5.2
test(integration): add channel resize test
```

## CI Workflows

| Workflow | Trigger | What it tests |
|----------|---------|---------------|
| `test-go.yml` | PR / push | Go tests (`go test ./...`) |
| `test-forge.yml` | PR / push | Contract tests (`forge test`) |
| `test-sdk.yml` | push | TypeScript SDK tests |
| `test-integration.yml` | push | Integration tests |
| `publish-sdk.yml` | release | Publish SDK to npm |
