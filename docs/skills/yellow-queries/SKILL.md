---
name: yellow-queries
description: |
  Complete catalog of the 12 read-only Nitro RPC query methods on ClearNode — get_config, get_assets, get_app_definition, get_channels, get_app_sessions, get_ledger_balances, get_ledger_entries, get_ledger_transactions, get_rpc_history, get_user_tag, get_session_keys, and ping. Use when: displaying balances or transaction history, listing channels / app sessions, checking the network config, measuring latency, or auditing which session keys are active.
version: 2.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/queries
---

# Yellow Queries — Read-only Nitro RPC methods

All queries are signed with the session key (same envelope as any other
request — see `yellow-nitro-rpc`). They never mutate state. Use them to
hydrate UI, reconcile accounting, and verify protocol health.

## The 12 methods

| Method | Auth | Purpose |
|---|---|---|
| `ping` | public | Liveness + latency |
| `get_config` | public | ClearNode config (chains, addresses) |
| `get_assets` | public | Supported tokens per chain |
| `get_app_definition` | public | Resolve an app by id |
| `get_channels` | public | List payment channels |
| `get_app_sessions` | public | List app sessions (escrow/stake/games) |
| `get_ledger_entries` | public | Double-entry audit trail |
| `get_ledger_transactions` | public | User-facing transaction history |
| `get_ledger_balances` | **auth** | Unified balance per asset |
| `get_user_tag` | **auth** | Caller's own user tag |
| `get_session_keys` | **auth** | Active session keys + allowance usage |
| `get_rpc_history` | **auth** | Caller's recent RPC calls |

**Pagination defaults** (for methods that accept them): `offset=0`,
`limit=10` (max **100**), `sort="desc"` by `created_at`.

## Response field names — verbatim

Get these wrong and your client breaks:

| Method | Response wrapper |
|---|---|
| `get_ledger_balances` | `{ ledger_balances: [{asset, amount}] }` (NOT `balances`) |
| `get_ledger_transactions` | `{ ledger_transactions: [...] }` |
| `get_ledger_entries` | `{ ledger_entries: [...] }` |
| `get_channels` | `{ channels: [...] }` |
| `get_app_sessions` | `{ app_sessions: [...] }` (plural) |
| `get_rpc_history` | `{ rpc_entries: [...] }` (NOT `rpc_history`) |
| `get_session_keys` | `{ session_keys: [...] }` with per-asset allowance breakdown |
| `get_user_tag` | `{ tag }` — takes **no params**, returns caller's tag |
| `ping` | method in response is `pong`, result is literally `null`. A non-error response is the signal |

Transaction rows use `from_account` / `to_account` (NOT `sender`/`receiver`).

## Quick examples

```ts
// Liveness + latency
const t0 = Date.now();
await client.send('ping', {});
const latencyMs = Date.now() - t0;

// Balances
const { ledger_balances } = await client.send('get_ledger_balances', {});
// ledger_balances[] → { asset: 'usdc', amount: '150.0' }

// Assets for a chain
const { assets } = await client.send('get_assets', { chain_id: 137 });
// assets[] → { token, chain_id, symbol, decimals }

// Transaction history (paginated)
const { ledger_transactions } = await client.send('get_ledger_transactions', {
  tx_type: 'transfer',
  limit: 50,
  sort: 'desc',
});
```

## Navigation Guide

### When to read supporting files

**reference.md** — read when you need:

- Per-method detailed params + response shapes (all 12 methods)
- Channel / AppSession row schemas in full
- Ledger entry vs ledger transaction — when to use which
- `get_rpc_history` record shape (id, sender, req_id, method, timestamp, sigs)
- `get_session_keys` full breakdown (allowance / used / remaining per asset)
- Hydration patterns (balances widget, activity feed, pre-close verification)
- `tx_type` enum values + filter combinations

## Related

- `yellow-nitro-rpc` — envelope format for every query
- `yellow-transfers` — `get_ledger_transactions` pairs with `transfer`
- `yellow-app-sessions` — `get_app_sessions` before `submit_app_state`/`close`
- `yellow-notifications` — push counterparts (`bu`/`cu`/`tr`/`asu`)
