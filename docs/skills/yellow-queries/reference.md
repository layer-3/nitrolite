# queries — Complete Reference

> Extended documentation. Read the `SKILL.md` quick-start first; this file contains the full details moved out for progressive disclosure.


# Yellow Queries — Read-only Nitro RPC methods

All queries are signed with the session key (same envelope format as any
other request — see `yellow-nitro-rpc`). They never mutate state. Use them
to hydrate UI, reconcile accounting, and verify protocol health.

## Method index

| Method | Auth | Purpose |
|---|---|---|
| `ping` | public | Liveness check + latency measurement |
| `get_config` | public | ClearNode configuration (supported chains, protocol version) |
| `get_assets` | public | Supported tokens and their decimals per chain |
| `get_app_definition` | public | Resolve a known app by id to its full `definition` |
| `get_channels` | public | List payment channels for a participant |
| `get_app_sessions` | public | List app sessions (escrow, stake, games) |
| `get_ledger_entries` | public | Double-entry bookkeeping audit trail |
| `get_ledger_transactions` | public | User-facing transaction history |
| `get_ledger_balances` | **auth** | Current balances per asset (unified or per app session) |
| `get_user_tag` | **auth** | Returns the authenticated user's own tag |
| `get_session_keys` | **auth** | Your active session keys + per-asset allowance usage |
| `get_rpc_history` | **auth** | Your recent RPC calls to this ClearNode |

**Pagination defaults** for methods that accept them: `offset = 0`,
`limit = 10` (max **100**), `sort = "desc"` (by `created_at`).


## `ping`

```json
req:  { /* no params */ }
res:  null  // response method is "pong"; result is literally null. A non-error response is the signal
```

No params, no body. Use on a 30s timer to keep the connection healthy and
measure latency:

```ts
const start = Date.now();
await client.send('ping', {});
console.log(`latency: ${Date.now() - start}ms`);
```

Also doubles as a cheap auth check — if the session expired, `ping` errors
with `"authentication required"`.

## `get_config`

Returns the ClearNode's self-description. Verbatim shape:

```json
{
  "broker_address": "0x...",
  "networks": [
    {
      "chain_id": 137,
      "name": "Polygon",
      "custody_address":    "0x...",
      "adjudicator_address": "0x..."
    }
  ]
}
```

Cache for the session — this rarely changes.

## `get_assets`

```ts
// All assets across all chains
const { assets } = await client.send('get_assets', {});

// Filtered to one chain
const { assets } = await client.send('get_assets', { chain_id: 137 });
```

Response:

```json
{
  "assets": [
    { "token": "0x2791Bca1...8A84174", "chain_id": 137, "symbol": "usdc", "decimals": 6 },
    { "token": "0x...",                 "chain_id": 1,   "symbol": "eth",  "decimals": 18 }
  ]
}
```

Asset `symbol` is **lowercase**. Always use `get_assets` before populating a
dropdown — don't hardcode token addresses in client code.

## `get_channels`

List on-chain payment channels associated with a participant:

```ts
const { channels } = await client.send('get_channels', {
  participant: '0x...',     // optional; defaults to caller
  chain_id: 137,            // optional filter
  status: 'open',           // optional filter: open | resizing | challenged | closed
  offset: 0, limit: 50,     // pagination
});
```

Each `Channel` includes `channel_id`, `chain_id`, `token`, `status`,
`version`, `allocations[]`, and on-chain contract addresses.

## `get_app_sessions`

```ts
const { app_sessions } = await client.send('get_app_sessions', {
  participant: '0x...',     // optional
  status: 'open',           // optional: open | closed
  offset: 0, limit: 50,
});
```

Each entry: `app_session_id`, `status`, `version`, `allocations`,
`session_data`, `definition`, plus `participants[]`. Query before signing a
`submit_app_state` or `close_app_session` to confirm the current version.

## `get_app_definition`

If you know an `app_session_id` but not its `definition`:

```ts
const { definition } = await client.send('get_app_definition', { app_session_id });
```

Returns the protocol, participants, weights, quorum, challenge window, and
nonce from the original `create_app_session`.

## `get_ledger_balances`

The single source of truth for "how much do I have?":

```ts
// Unified balance (default = caller)
const { ledger_balances } = await client.send('get_ledger_balances', {});

// Balance inside an app session
const { ledger_balances } = await client.send('get_ledger_balances', {
  account_id: '0xAppSessionId',
});
```

Response (note the field name — **`ledger_balances`**, not `balances`):

```json
{
  "ledger_balances": [
    { "asset": "usdc", "amount": "150.0" },
    { "asset": "eth",  "amount": "0.05"  }
  ]
}
```

Returns **all tracked assets**, including ones that currently evaluate to
zero.

## `get_ledger_entries`

Double-entry bookkeeping audit trail — each entry is one side (debit OR
credit) of a transaction. Use for reconciliation, not for end-user display.

```ts
const { ledger_entries } = await client.send('get_ledger_entries', {
  account_id: '0x...',       // optional
  wallet: '0x...',           // optional
  asset: 'usdc',             // optional
  offset: 0, limit: 100,
  sort: 'desc',              // asc | desc
});
```

Each entry: `id`, `account_id`, `account_type` (integer — observed `0`
on the sandbox for the default asset ledger), `asset`, `participant`,
`credit`, `debit`, `created_at` (ISO 8601).

## `get_ledger_transactions`

Higher-level transaction history for end-user display (sender, receiver,
amount, type):

```ts
const { ledger_transactions } = await client.send('get_ledger_transactions', {
  account_id: '0x...',       // optional
  asset: 'usdc',             // optional
  tx_type: 'transfer',       // optional: transfer | deposit | withdrawal |
                             //            app_deposit | app_withdrawal |
                             //            escrow_lock | escrow_unlock
  offset: 0, limit: 100,
  sort: 'desc',
});
```

Each record (verbatim field names):

```ts
{
  id:               number,
  tx_type:          'transfer' | 'deposit' | 'withdrawal' | 'app_deposit' |
                    'app_withdrawal' | 'escrow_lock' | 'escrow_unlock',
  from_account:     address,      // NOT `sender`
  from_account_tag: string,       // optional @handle
  to_account:       address,      // NOT `receiver`
  to_account_tag:   string,
  asset:            string,
  amount:           string,
  created_at:       string,
}
```

Sort newest first by default. Use this (not `get_ledger_entries`) to power a
user-facing activity feed.

## `get_rpc_history`

Your own recent RPC calls to this ClearNode. Useful for debugging.
**Response field is `rpc_entries`** (not `rpc_history`):

```ts
const { rpc_entries } = await client.send('get_rpc_history', {
  offset: 0, limit: 10, sort: 'desc',
});
// rpc_entries[] → { id, sender, req_id, method, params, timestamp,
//                   req_sig[], response, res_sig[] }
```

## `get_user_tag`

Returns the **authenticated user's** tag (if one is set). Takes no params:

```ts
const { tag } = await client.send('get_user_tag', {});
// tag === 'UX123D' or null
```

To resolve someone else's tag → address, use `get_ledger_transactions` and
read the `from_account_tag` / `to_account_tag` fields on their transactions
(or discover via an external directory).

## `get_session_keys`

Audit live session keys and their remaining allowances. Response shape
(verbatim per spec — note **`session_keys`** and per-asset allowance
breakdown):

```ts
const { session_keys } = await client.send('get_session_keys', {
  offset: 0, limit: 10, sort: 'desc',  // all optional
});
// session_keys[] → {
//   id:          number,
//   session_key: address,
//   application: string,
//   allowances:  Array<{ asset: string, allowance: string, used: string }>,
//   scope:       string,
//   expires_at:  string,
//   created_at:  string,
// }
```

Compute `remaining = allowance - used` per allowance entry. Use in an agent
dashboard or to decide when to rotate.


## Patterns

### Hydrate a balances widget

```ts
const [{ ledger_balances }, { assets }] = await Promise.all([
  client.send('get_ledger_balances', {}),
  client.send('get_assets', {}),
]);
// Join by symbol → render with correct decimals
```

### Render activity feed

```ts
const { ledger_transactions } = await client.send('get_ledger_transactions', {
  limit: 50, sort: 'desc',
});
for (const tx of ledger_transactions) {
  // schema fields are from_account / to_account (not sender / receiver)
  render(tx.tx_type, tx.from_account, tx.to_account, tx.amount, tx.asset);
}
```

### Confirm before closing an app session

```ts
const { app_sessions } = await client.send('get_app_sessions', { status: 'open' });
const session = app_sessions.find((s) => s.app_session_id === id);
if (!session) throw new Error('Session not open or not found');
// Use session.version + 1, session.allocations as the basis for close payload
```

## Related

- `yellow-nitro-rpc` — envelope format for every query
- `yellow-transfers` — queries pair with transfer to verify settlement
- `yellow-app-sessions` — queries pair with app-session mutations
- `yellow-notifications` — push complements polling (bu, cu, tr, asu)
