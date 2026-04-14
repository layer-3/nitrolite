# MCP Server Test Cases

Manual vetting checklist for both the TypeScript and Go MCP servers.
Derived from `sdk/mcp-test-plan.md`. Run each case by connecting to the server through an MCP client or the runner script.

---

## How to Run

All commands from **repo root** (`/path/to/nitrolite`):

```bash
# Setup (one-time)
cd sdk/mcp-test-results && npm install && cd ..
cd sdk/mcp && npm install && cd ..

# Run the vetting script
npm --prefix sdk/mcp-test-results exec -- tsx sdk/mcp-test-results/run.ts
```

The script connects to both servers using the same cwd and commands as `.mcp.json`.

---

## 0. Surface Inventory

Verify the registered surface matches expectations. If anything is missing or renamed, the server has a regression.

| # | Server | Check | Expected |
|---|--------|-------|----------|
| 0.1 | Unified | `listResources()` | 30 resources (see plan for full URI list) |
| 0.2 | Unified | `listTools()` | 8 tools: `lookup_method`, `lookup_type`, `search_api`, `get_rpc_method`, `validate_import`, `explain_concept`, `scaffold_project`, `lookup_rpc_method` |
| 0.3 | Unified | `listPrompts()` | 3 prompts: `create-channel-app`, `migrate-from-v053`, `build-ai-agent-app` |

---

## A. .mcp.json Smoke Check

| # | Check | Expected | Pass? |
|---|-------|----------|-------|
| A.1 | `.mcp.json` has `nitrolite` entry | command: `npm`, args include `--prefix sdk/mcp` | |
| A.2 | Unified server connects and responds to `listTools()` | No error, returns 8 tools | |

---

## B. Transfer App (TS)

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| B.1 | `search_api` | `"transfer"` | Mentions `Decimal`, not string amounts | |
| B.2 | `lookup_method` | `"create"` | `Client.create(wsURL, stateSigner, txSigner, ...opts)` | |
| B.3 | `lookup_method` | `"deposit"` | `deposit(blockchainId: bigint, asset: string, amount: Decimal)` | |
| B.4 | `lookup_method` | `"transfer"` | `transfer(recipientWallet, asset, amount: Decimal)` | |
| B.5 | `lookup_method` | `"approveToken"` | 3 params including `amount: Decimal` | |
| B.6 | `lookup_method` | `"checkpoint"` | `checkpoint(asset: string)` | |
| B.7 | `lookup_method` | `"getBalances"` | `getBalances(wallet: Address)` | |
| B.8 | `lookup_method` | `"closeHomeChannel"` | `closeHomeChannel(asset: string)` | |
| B.9 | `scaffold_project` | `"transfer-app"` | Has `createSigners`, `Decimal`, `checkpoint`. NOT `walletClient`/`DefaultConfig`/string amounts | |
| B.10 | Resource | `nitrolite://examples/full-transfer-script` | Correct constructor, `approveToken` with 3 args, deposit+checkpoint, getBalances with wallet arg, close+checkpoint | |
| B.11 | `get_rpc_method` | `"transfer"` | Params: `{destination, allocations}`, NOT `[recipient, ...]` | |

---

## C. App Session Game (TS)

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| C.1 | `search_api` | `"app session"` | Finds `createAppSession`, `submitAppState`. NOT `closeAppSession` | |
| C.2 | `lookup_method` | `"createAppSession"` | `(definition, sessionData, quorumSigs, opts?)` | |
| C.3 | `lookup_method` | `"submitAppState"` | `(appStateUpdate, quorumSigs)` | |
| C.4 | `lookup_method` | `"closeAppSession"` | "not found" | |
| C.5 | `lookup_type` | `"AppDefinitionV1"` | Shows `applicationId`, `quorum`, `nonce`. NOT `protocol`, `appName`, `weights` | |
| C.6 | `lookup_type` | `"AppStateUpdateV1"` | Shows `appSessionId`, `intent`, `version`, `allocations` | |
| C.7 | `scaffold_project` | `"app-session"` | Correct definition shape, quorum sigs, versioned updates, close via Close intent | |
| C.8 | Resource | `nitrolite://examples/app-sessions` | `applicationId`/`quorum`/`nonce`, no `closeAppSession` call | |
| C.9 | Resource | `nitrolite://examples/full-app-session-script` | Quorum sigs, version 2n/3n, close via intent | |
| C.10 | `get_rpc_method` | `"create_app_session"` | Params include `definition`, `session_data`, `quorum_sigs` | |
| C.11 | `get_rpc_method` | `"submit_app_state"` | Params include `app_state_update`, `quorum_sigs` | |

---

## D. AI Payment Agent (TS)

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| D.1 | `scaffold_project` | `"ai-agent"` | Uses `createSigners`, `Decimal`. NOT `walletClient`/`DefaultConfig` | |
| D.2 | Resource | `nitrolite://use-cases/ai-agents` | Correct constructor, `Decimal` amounts | |
| D.3 | `lookup_method` | `"approveToken"` | 3-arg signature confirmed | |

---

## E. Migration from v0.5.3 (TS only)

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| E.1 | `validate_import` | `"NitroliteClient"` | Found in `@yellow-org/sdk-compat` | |
| E.2 | `validate_import` | `"createAuthRequestMessage"` | Found in compat | |
| E.3 | `validate_import` | `"Client"` | Found in `@yellow-org/sdk`, NOT compat | |
| E.4 | `validate_import` | `"RPCMethod"` | Found in compat | |
| E.5 | `validate_import` | `"Cli"` | NOT found — must not substring-match | |
| E.6 | `validate_import` | `"FakeSymbol"` | Not found in either | |
| E.7 | Resource | `nitrolite://examples/auth` | Correct imports, `AuthRequestParams` object shape | |
| E.8 | Resource | `nitrolite://migration/overview` | Loads migration docs | |
| E.9 | Prompt | `migrate-from-v053` | Non-empty, mentions compat layer | |

---

## F. Protocol Understanding (both servers)

| # | Server | Tool/Resource | Input | Expected | Pass? |
|---|--------|--------------|-------|----------|-------|
| F.1 | Both | `explain_concept` | `"state channel"` | Meaningful explanation | |
| F.2 | Both | `explain_concept` | `"app session"` | Meaningful | |
| F.3 | Both | `explain_concept` | `"challenge period"` | Meaningful | |
| F.4 | Both | `explain_concept` | `"clearnode"` | Meaningful | |
| F.5 | Both | `explain_concept` | `"made up thing"` | Graceful "not found" | |
| F.6 | Both | Resource | `nitrolite://protocol/overview` | Non-empty (>50 chars) | |
| F.7 | Both | Resource | `nitrolite://protocol/terminology` | Has definitions | |
| F.8 | Both | Resource | `nitrolite://protocol/wire-format` | Non-empty | |
| F.9 | Both | Resource | `nitrolite://security/overview` | Non-empty | |
| F.10 | Both | Resource | `nitrolite://security/app-session-patterns` | Non-empty | |
| F.11 | Both | Resource | `nitrolite://security/state-invariants` | Non-empty | |

---

## G. RPC Wire Formats (TS)

| # | Tool | Input | Assert contains | Assert NOT contains | Pass? |
|---|------|-------|-----------------|---------------------|-------|
| G.1 | `get_rpc_method` | `"get_ledger_balances"` | `wallet` | — | |
| G.2 | `get_rpc_method` | `"create_channel"` | `chain_id, token` | `[token, chainId, amount]` | |
| G.3 | `get_rpc_method` | `"close_channel"` | `channel_id, funds_destination` | `[channelId]` | |
| G.4 | `get_rpc_method` | `"resize_channel"` | `resize_amount, allocate_amount` | `[channelId, amount]` | |
| G.5 | `get_rpc_method` | `"nonexistent"` | Lists available methods | — | |
| G.6 | `lookup_rpc_method` | `"channels.v1.get_home_channel"` | Has description | — | |
| G.7 | `lookup_rpc_method` | `"app_sessions.v1.create_app_session"` | Exists | — | |
| G.8 | `lookup_rpc_method` | `"user.v1.get_balances"` | Exists | — | |

---

## H. Go SDK

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| H.1 | `lookup_method` | `"Deposit"` | Exists, has `ctx context.Context` | |
| H.2 | `lookup_method` | `"Transfer"` | Exists | |
| H.3 | `lookup_method` | `"CreateAppSession"` | Exists, mentions quorumSigs | |
| H.4 | `lookup_method` | `"SubmitAppState"` | Exists, mentions quorumSigs | |
| H.5 | `lookup_method` | `"CloseHomeChannel"` | Exists | |
| H.6 | `lookup_method` | `"Checkpoint"` | Exists | |
| H.7 | `lookup_method` | `"GetBalances"` | Exists | |
| H.8 | `lookup_method` | `"GetConfig"` | Exists | |
| H.9 | `lookup_method` | `"CloseAppSession"` | NOT found | |
| H.10 | `scaffold_project` | `"go-transfer-app"` | Has `Checkpoint` after `Deposit` | |
| H.11 | `scaffold_project` | `"go-app-session"` | Quorum sigs (not nil), versioned, close via Close intent | |
| H.12 | `scaffold_project` | `"go-ai-agent"` | Reasonable template | |
| H.13 | Resource | `nitrolite://go-examples/full-transfer-script` | `CloseHomeChannel` + `Checkpoint` | |
| H.14 | Resource | `nitrolite://go-examples/full-app-session-script` | Quorum sigs, `initVersion + 1`, close via intent | |
| H.15 | Resource | `nitrolite://use-cases` | "close via Close intent", NOT `CloseAppSession()` | |
| H.16 | `lookup_type` | `"AppSessionV1"`, `language: "go"` | Found, non-empty | |
| H.17 | `lookup_type` | `"ChannelStatus"`, `language: "go"` | Enum found | |

---

## I. Type Discovery (TS)

| # | Tool/Resource | Input | Expected | Pass? |
|---|--------------|-------|----------|-------|
| I.1 | Resource | `nitrolite://api/methods` | Lists 46+ methods with signatures | |
| I.2 | Resource | `nitrolite://api/types` | Lists interfaces/types with fields | |
| I.3 | Resource | `nitrolite://api/enums` | Lists enums with values | |
| I.4 | `lookup_type` | `"AppDefinitionV1"` | `applicationId`, `participants`, `quorum`, `nonce` | |
| I.5 | `lookup_type` | `"AppStateUpdateV1"` | `appSessionId`, `intent`, `version`, `allocations` | |
| I.6 | `lookup_type` | `"NonexistentType"` | Graceful not-found | |

---

## J. Prompts (both servers)

| # | Server | Prompt | Expected | Pass? |
|---|--------|--------|----------|-------|
| J.1 | TS | `create-channel-app` | Non-empty, mentions channel lifecycle | |
| J.2 | TS | `migrate-from-v053` | Loads migration doc content | |
| J.3 | TS | `build-ai-agent-app` | Non-empty, mentions session keys | |
| J.4 | Go | `create-channel-app` | Mentions `Checkpoint`, NOT `CloseAppSession` | |
| J.5 | Go | `build-ai-agent-app` | Non-empty | |

---

## K. Edge Cases

| # | Server | Tool | Input | Expected | Pass? |
|---|--------|------|-------|----------|-------|
| K.1 | TS | `lookup_method` | `"nonexistent"` | Graceful message, no crash | |
| K.2 | TS | `get_rpc_method` | `"nonexistent"` | Lists available methods | |
| K.3 | TS | `search_api` | `""` | Does not crash | |
| K.4 | TS | `scaffold_project` | `"transfer-app"` | package.json includes `decimal.js` | |
| K.5 | Go | `scaffold_project` | `"nonexistent"` | Error listing templates | |
| K.6 | Go | `search_api` | `"zzzzz"` | "no matches" gracefully | |

---

## L. Resource Sweep (completeness)

Reads every remaining resource not covered by scenarios A–K. Verifies non-empty content.

### TS resources not covered above

| # | Resource URI | Check | Pass? |
|---|-------------|-------|-------|
| L.1 | `nitrolite://examples/channels` | Non-empty, has code | |
| L.2 | `nitrolite://examples/transfers` | Non-empty, has code | |
| L.3 | `nitrolite://protocol/rpc-methods` | Non-empty | |
| L.4 | `nitrolite://protocol/auth-flow` | Non-empty | |
| L.5 | `nitrolite://protocol/channel-lifecycle` | Non-empty | |
| L.6 | `nitrolite://protocol/state-model` | Non-empty | |
| L.7 | `nitrolite://use-cases` | Non-empty | |

### Go resources not covered above

| # | Resource URI | Check | Pass? |
|---|-------------|-------|-------|
| L.11 | `nitrolite://go-api/methods` | Non-empty, 44+ methods | |
| L.12 | `nitrolite://go-api/types` | Non-empty, Go structs/enums | |
| L.13 | `nitrolite://go-examples/full-transfer-script` | Non-empty, Go code | |
| L.14 | `nitrolite://go-examples/full-app-session-script` | Non-empty, Go code | |
| L.15 | `nitrolite://protocol/enforcement` | Non-empty | |
| L.16 | `nitrolite://protocol/auth-flow` | Non-empty | |

### TS tool alias

| # | Tool | Input | Check | Pass? |
|---|------|-------|-------|-------|
| L.17 | `lookup_rpc_method` (TS) | `"transfer"` | Returns same result as `get_rpc_method("transfer")` | |

---

## P1 Gate (must-pass before push)

- [ ] 0.1–0.6 — Surface inventory matches (no missing/renamed tools/resources/prompts)
- [ ] A.1–A.4 — `.mcp.json` launches both servers
- [ ] B.1–B.11 — Transfer app scenario all correct
- [ ] C.1–C.11 — App session scenario all correct
- [ ] E.3, E.5 — `validate_import` exact match (no false positives)
- [ ] G.1–G.4 — RPC formats match API.md
- [ ] H.1–H.9 — Go method coverage (all exist, `CloseAppSession` doesn't)
- [ ] I.4–I.5 — Type discovery returns correct shapes
