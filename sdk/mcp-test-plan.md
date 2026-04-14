# MCP Server Vetting — First-Time User Walkthrough

## What This Is

A hands-on test of both MCP servers as a developer would experience them. We connect to each server, exercise the high-value user-facing surface through realistic scenarios, verify the full inventory is registered, and sweep all remaining resources for non-empty content. Then we check: does the output actually help me build on Nitrolite? Is anything wrong, misleading, or missing?

Not an automated test suite. A script that exercises the servers, dumps raw output for review, then a human-written findings report.

---

## Canonical Setup

Everything runs from **repo root**. One cwd, one set of commands.

```bash
# From repo root:
cd /Users/maharshimishra/Documents/nitrolite

# Install the test harness deps (one-time)
cd sdk/mcp-test-results && npm install && cd ..

# Ensure MCP server deps are installed (one-time)
cd sdk/mcp && npm install && cd ..

# Run the vetting script (writes results to sdk/mcp-test-results/)
npm --prefix sdk/mcp-test-results exec -- tsx sdk/mcp-test-results/run.ts
```

The runner script uses `cwd: REPO_ROOT` for both servers, matching how `.mcp.json` launches them.

---

## Scenario 0: Surface Inventory

**Before testing any content, verify the registered surface matches what we expect.**

### Unified Server — Expected Inventory

**30 Resources:**
| URI | Name |
|-----|------|
| `nitrolite://api/methods` | api-methods |
| `nitrolite://api/types` | api-types |
| `nitrolite://api/enums` | api-enums |
| `nitrolite://go-api/methods` | go-api-methods |
| `nitrolite://go-api/types` | go-api-types |
| `nitrolite://examples/channels` | examples-channels |
| `nitrolite://examples/transfers` | examples-transfers |
| `nitrolite://examples/app-sessions` | examples-app-sessions |
| `nitrolite://examples/auth` | examples-auth |
| `nitrolite://examples/full-transfer-script` | examples-full-transfer |
| `nitrolite://examples/full-app-session-script` | examples-full-app-session |
| `nitrolite://go-examples/full-transfer-script` | go-examples-full-transfer |
| `nitrolite://go-examples/full-app-session-script` | go-examples-full-app-session |
| `nitrolite://migration/overview` | migration-overview |
| `nitrolite://protocol/overview` | protocol-overview |
| `nitrolite://protocol/terminology` | protocol-terminology |
| `nitrolite://protocol/wire-format` | protocol-wire-format |
| `nitrolite://protocol/rpc-methods` | protocol-rpc-methods |
| `nitrolite://protocol/cryptography` | protocol-cryptography |
| `nitrolite://protocol/auth-flow` | protocol-auth-flow |
| `nitrolite://protocol/channel-lifecycle` | protocol-channel-lifecycle |
| `nitrolite://protocol/state-model` | protocol-state-model |
| `nitrolite://protocol/enforcement` | protocol-enforcement |
| `nitrolite://protocol/cross-chain` | protocol-cross-chain |
| `nitrolite://protocol/interactions` | protocol-interactions |
| `nitrolite://security/overview` | security-overview |
| `nitrolite://security/app-session-patterns` | security-app-session-patterns |
| `nitrolite://security/state-invariants` | security-state-invariants |
| `nitrolite://use-cases` | use-cases |
| `nitrolite://use-cases/ai-agents` | use-cases-ai-agents |

**8 Tools:** `lookup_method`, `lookup_type`, `search_api`, `get_rpc_method`, `validate_import`, `explain_concept`, `scaffold_project`, `lookup_rpc_method`

**3 Prompts:** `create-channel-app`, `migrate-from-v053`, `build-ai-agent-app`

### Acceptance

- `listResources()` returns exactly the expected URIs (none missing, none extra)
- `listTools()` returns exactly the expected tool names
- `listPrompts()` returns exactly the expected prompt names

---

## Scenario A: ".mcp.json smoke check"

The real first-touch. Read `.mcp.json`, verify the command would launch the unified server.

1. Read `.mcp.json` — confirm `nitrolite` entry exists
2. The command (`npm --prefix sdk/mcp exec -- tsx sdk/mcp/src/index.ts`) must resolve from repo root
3. Server connects and responds to `listTools()` without error

---

## Scenario B: "I want to build a transfer app" (TS)

1. `search_api("transfer")` — find the right methods. Output should mention `Decimal`, not string amounts
2. `lookup_method("create")` — returns `Client.create(wsURL, stateSigner, txSigner, ...opts)`
3. `lookup_method("deposit")` — returns `deposit(blockchainId: bigint, asset: string, amount: Decimal)`
4. `lookup_method("transfer")` — returns `transfer(recipientWallet: string, asset: string, amount: Decimal)`
5. `lookup_method("approveToken")` — returns 3 params including `amount: Decimal`
6. `lookup_method("checkpoint")` — returns `checkpoint(asset: string)`
7. `lookup_method("getBalances")` — returns `getBalances(wallet: Address)`
8. `lookup_method("closeHomeChannel")` — returns `closeHomeChannel(asset: string)`
9. `scaffold_project("transfer-app")` — check: `createSigners`, `Decimal`, `checkpoint` after deposit, NOT `walletClient`/`DefaultConfig`/string amounts
10. Read `nitrolite://examples/full-transfer-script` — complete runnable script with correct constructor, `approveToken` with 3 args, deposit+checkpoint, getBalances with wallet arg, close+checkpoint
11. `get_rpc_method("transfer")` — params are `{destination, allocations}`, not `[recipient, ...]`

---

## Scenario C: "I want to build an app session game" (TS)

1. `search_api("app session")` — finds `createAppSession`, `submitAppState`. Does NOT mention `closeAppSession`
2. `lookup_method("createAppSession")` — `(definition: AppDefinitionV1, sessionData: string, quorumSigs: string[], opts?)`
3. `lookup_method("submitAppState")` — `(appStateUpdate: AppStateUpdateV1, quorumSigs: string[])`
4. `lookup_method("closeAppSession")` — returns "not found"
5. `lookup_type("AppDefinitionV1")` — shows `applicationId`, `participants`, `quorum`, `nonce`. NOT `protocol`, `appName`, `weights`
6. `lookup_type("AppStateUpdateV1")` — shows `appSessionId`, `intent`, `version`, `allocations`, `sessionData`
7. `scaffold_project("app-session")` — correct definition shape, quorum sigs, versioned updates, close via Close intent, NOT `closeAppSession()`
8. Read `nitrolite://examples/app-sessions` — correct `applicationId`/`quorum`/`nonce` shape, no `closeAppSession` call
9. Read `nitrolite://examples/full-app-session-script` — complete flow with quorum sigs, version 2n/3n, close via intent
10. `get_rpc_method("create_app_session")` — params include `definition`, `session_data`, `quorum_sigs`
11. `get_rpc_method("submit_app_state")` — params include `app_state_update`, `quorum_sigs`

---

## Scenario D: "I want to build an AI payment agent" (TS)

1. `scaffold_project("ai-agent")` — uses `createSigners`, `Decimal`, NOT `walletClient`/`DefaultConfig`
2. Read `nitrolite://use-cases/ai-agents` — correct constructor (`createSigners` + `Client.create(wsURL, ...)`), `Decimal` amounts
3. `lookup_method("approveToken")` — 3-arg signature confirmed

---

## Scenario E: "I'm migrating from v0.5.3" (TS only)

1. `validate_import("NitroliteClient")` — found in `@yellow-org/sdk-compat`
2. `validate_import("createAuthRequestMessage")` — found in compat
3. `validate_import("Client")` — found in `@yellow-org/sdk`, NOT compat (SSR safety)
4. `validate_import("RPCMethod")` — found in compat
5. `validate_import("Cli")` — NOT found (must not substring-match `Client`)
6. `validate_import("FakeSymbol")` — not found in either package
7. Read `nitrolite://examples/auth` — correct imports (`createAuthRequestMessage`, `AuthRequestParams`), correct call shape
8. Read `nitrolite://migration/overview` — loads migration docs content
9. Prompt `migrate-from-v053` — returns non-empty migration guide content

---

## Scenario F: "I need to understand the protocol"

1. `explain_concept("state channel")` — meaningful explanation (both servers)
2. `explain_concept("app session")` — meaningful (both servers)
3. `explain_concept("challenge period")` — meaningful (both servers)
4. `explain_concept("clearnode")` — meaningful (both servers)
5. `explain_concept("made up thing")` — graceful "not found" (both servers)
6. Read `nitrolite://protocol/overview` — non-empty content from docs (both servers)
7. Read `nitrolite://protocol/terminology` — has definitions (both servers)
8. Read `nitrolite://protocol/wire-format` — non-empty (both servers)
9. Read `nitrolite://security/overview` — non-empty (both servers)
10. Read `nitrolite://security/app-session-patterns` — non-empty (both servers)
11. Read `nitrolite://security/state-invariants` — non-empty (both servers)

---

## Scenario G: "I'm checking RPC wire formats" (TS)

1. `get_rpc_method("get_ledger_balances")` — params include `wallet`
2. `get_rpc_method("create_channel")` — `[{chain_id, token}]`, NOT `[token, chainId, amount]`
3. `get_rpc_method("close_channel")` — `{channel_id, funds_destination}`, NOT `[channelId]`
4. `get_rpc_method("resize_channel")` — `{channel_id, resize_amount, allocate_amount, funds_destination}`
5. `get_rpc_method("transfer")` — `{destination, allocations}`, NOT `[recipient, ...]`
6. `get_rpc_method("nonexistent")` — lists available methods, no crash
7. `lookup_rpc_method("channels.v1.get_home_channel")` — returns description
8. `lookup_rpc_method("app_sessions.v1.create_app_session")` — exists
9. `lookup_rpc_method("user.v1.get_balances")` — exists

---

## Scenario H: "I'm using the Go SDK"

1. `lookup_method("Deposit")` — exists, has `ctx context.Context`
2. `lookup_method("Transfer")` — exists
3. `lookup_method("CreateAppSession")` — exists, mention quorumSigs
4. `lookup_method("SubmitAppState")` — exists, mention quorumSigs
5. `lookup_method("CloseHomeChannel")` — exists
6. `lookup_method("Checkpoint")` — exists
7. `lookup_method("GetBalances")` — exists
8. `lookup_method("GetConfig")` — exists
9. `lookup_method("CloseAppSession")` — NOT found
10. `scaffold_project("go-transfer-app")` — has `Checkpoint` after `Deposit`
11. `scaffold_project("go-app-session")` — has quorum sigs (not `nil`), versioned updates, close via Close intent
12. `scaffold_project("go-ai-agent")` — reasonable agent template
13. Read `nitrolite://go-examples/full-transfer-script` — `CloseHomeChannel` + `Checkpoint`
14. Read `nitrolite://go-examples/full-app-session-script` — quorum sigs, `initVersion + 1`, close via Close intent
15. Read `nitrolite://use-cases` — "close via Close intent", NOT `CloseAppSession()`
16. `lookup_type("AppSessionV1", language: "go")` — found, non-empty
17. `lookup_type("ChannelStatus", language: "go")` — enum found

---

## Scenario I: "Type discovery" (TS)

1. Read `nitrolite://api/methods` — lists all indexed methods with signatures. Count should be 46+
2. Read `nitrolite://api/types` — lists interfaces/types with fields
3. Read `nitrolite://api/enums` — lists enums with values
4. `lookup_type("AppDefinitionV1")` — returns `applicationId`, `participants`, `quorum`, `nonce`
5. `lookup_type("AppStateUpdateV1")` — returns `appSessionId`, `intent`, `version`, `allocations`, `sessionData`
6. `lookup_type("BalanceEntry")` — returns something reasonable
7. `lookup_type("NonexistentType")` — graceful not-found

---

## Scenario J: "Prompts" (both servers)

### TS Prompts
1. `create-channel-app` — non-empty guide mentioning channel lifecycle, transfers, app sessions
2. `migrate-from-v053` — loads migration doc content, mentions compat layer
3. `build-ai-agent-app` — non-empty guide mentioning session keys, agent payments

### Go Prompts
4. `create-channel-app` — mentions `Checkpoint` after `CloseHomeChannel`, NOT `CloseAppSession`
5. `build-ai-agent-app` — non-empty agent guide

---

## Scenario K: "Edge cases"

1. TS: `lookup_method("nonexistent")` — graceful message, no crash
2. TS: `get_rpc_method("nonexistent")` — lists available methods
3. TS: `search_api("")` — does not crash
4. TS: `scaffold_project("transfer-app")` — output includes `decimal.js` in package.json dependencies
5. Go: `scaffold_project("nonexistent")` — error listing available templates
6. Go: `search_api("zzzzz")` — "no matches" gracefully

---

## Scenario L: Resource Sweep (completeness check)

Scenarios A–K exercise the high-value surface through realistic user flows. This sweep reads every remaining resource not covered above to verify non-empty content. The runner script reads ALL resource URIs from `listResources()` and logs output length + first 200 chars.

### TS resources not covered by scenarios A–K:

| URI | Check |
|-----|-------|
| `nitrolite://examples/channels` | Non-empty, has code examples |
| `nitrolite://examples/transfers` | Non-empty, has code examples |
| `nitrolite://protocol/rpc-methods` | Non-empty (from docs/api.yaml) |
| `nitrolite://protocol/auth-flow` | Non-empty |
| `nitrolite://protocol/cryptography` | Non-empty (from docs/protocol/) |
| `nitrolite://protocol/channel-lifecycle` | Non-empty |
| `nitrolite://protocol/state-model` | Non-empty |
| `nitrolite://protocol/enforcement` | Non-empty |
| `nitrolite://protocol/cross-chain` | Non-empty |
| `nitrolite://protocol/interactions` | Non-empty |
| `nitrolite://use-cases` | Non-empty |

### Go resources not covered by scenarios A–K:

| URI | Check |
|-----|-------|
| `nitrolite://go-api/methods` | Non-empty, 44+ methods listed |
| `nitrolite://go-api/types` | Non-empty, Go structs/enums |
| `nitrolite://go-examples/full-transfer-script` | Non-empty, Go code |
| `nitrolite://go-examples/full-app-session-script` | Non-empty, Go code |

### TS tool alias check:

The TS server registers both `get_rpc_method` and `lookup_rpc_method`. Scenarios only exercise `get_rpc_method`. Verify `lookup_rpc_method` also works:

| Tool | Input | Check |
|------|-------|-------|
| `lookup_rpc_method` (TS) | `"transfer"` | Returns same result as `get_rpc_method("transfer")` |

---

## Deliverables

| File | What it contains |
|------|------------------|
| `sdk/mcp-test-results/run.ts` | Script that connects to the unified server and dumps all output |
| `sdk/mcp-test-results/ts-server.md` | Raw output from every TS tool call, resource read, and prompt fetch |
| `sdk/mcp-test-results/go-server.md` | Raw output from every Go tool call and resource read (same unified server) |
| `sdk/mcp-test-results/findings.md` | Pass/fail per scenario with notes on anything wrong/misleading/missing |

---

## Implementation Order

1. Create `sdk/mcp-test-results/` directory
2. Write `run.ts` — spawns each server, calls all tools/resources/prompts per scenario, writes markdown output files
3. Run it from repo root
4. Review raw output, write `findings.md`
