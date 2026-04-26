---
name: yellow-notifications
description: |
  The five server-to-client Nitro RPC notifications ClearNode pushes over the same WebSocket — `assets` (connection-hydration push fired once on connect), `bu` (balance update), `cu` (channel update), `tr` (transfer), `asu` (app session update). Covers when each fires, exact payload field names, how to route them client-side (requestId = 0), and common patterns (balance UI hydration, live channel monitoring, transfer receipts, app-session reactive updates). Use when: implementing real-time UI, reconciling state without polling, or debugging why a notification isn't arriving.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/queries
---

# Yellow Notifications — Server-push over Nitro RPC

ClearNode sends unsolicited asynchronous messages on the same WebSocket used
for RPC. Subscribing replaces polling for balances, channel state, transfer
receipts, and app-session updates.

Spec: <https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/queries>
(the "Notifications (Server-to-Client)" section).

## v1 SDK caveat

`@yellow-org/sdk@1.x` **replaces the push-notification model with
polling**. The `Client` class exposes `get*` methods and you poll them
on a timer. The v0.5.3-compat layer (`@yellow-org/sdk-compat`) ships
an `EventPoller` that re-synthesizes the `bu`/`cu`/`tr`/`asu` push shape
for backwards compat, but under the hood it's still polling. If you're
building new against v1, design for polling (or use the `EventPoller`
shim) — the notifications documented below are the **v0.5.3 / raw Nitro
RPC** model.

## Wire format (v0.5.3 / raw RPC)

Same envelope as a response, with:
- `requestId = 0` (distinguishes it from a reply to a pending call)
- `method` = the notification name (`bu` / `cu` / `tr` / `asu`)

```json
{ "res": [0, "bu", { "balance_updates": [...] }, 1776953834665], "sig": ["0x..."] }
```

Router rule (see `yellow-nitro-rpc`):

```ts
if (id === 0 || !pendingCallbacks.has(id)) {
  dispatchNotification(method, result);
} else {
  resolvePending(id, method, result);
}
```

## The five notifications

### `assets` — connection-hydration push

Fired **once, automatically, on every fresh WebSocket connect** — before
`auth_request`. Same payload shape as `get_assets` so you can warm an asset
cache without an explicit query:

```json
{ "assets": [
  { "token": "0x...", "chain_id": 137, "symbol": "usdc", "decimals": 6 },
  ...
]}
```

Treat this as "free" and use it to pre-populate any per-asset lookups.

### `bu` — Balance Update

Fires whenever account balances change (transfers, app-session
deposits/withdraws, channel resize, etc.).

Payload:
```json
{
  "balance_updates": [
    { "asset": "usdc", "amount": "150.0" },
    { "asset": "eth",  "amount": "0.05"  }
  ]
}
```

Each `LedgerBalance`:
- `asset` (string, lowercase symbol)
- `amount` (string, decimal, **new total** — not delta)

Uses: real-time balance widget, animated counters, balance history log.

### `cu` — Channel Update

Fires when a channel's state changes — opened, resizing, challenged, closed.

Payload is a full `Channel` object (same shape as an entry returned by
`get_channels`):
```json
{
  "channel_id": "0x...",
  "participant": "0x...",
  "status": "open",            // open | resizing | challenged | closed
  "token": "0x...",
  "wallet": "0x...",
  "amount": "100.0",
  "chain_id": 137,
  "adjudicator": "0x...",
  "challenge": 3600,
  "nonce": 42,
  "version": 7,
  "created_at": "…",
  "updated_at": "…"
}
```

Uses: monitor disputes (`status: "challenged"` → alert user), prevent
premature closes, show channel version on UI.

### `tr` — Transfer

Fires when a transfer credits or debits the user's account — both incoming
and outgoing.

Payload:
```json
{
  "transactions": [
    {
      "id": 123,
      "tx_type": "transfer",
      "from_account": "0xA...",
      "from_account_tag": "@alice",
      "to_account": "0xB...",
      "to_account_tag": "@bob",
      "asset": "usdc",
      "amount": "50.0",
      "created_at": "2026-04-23T12:00:00Z"
    }
  ]
}
```

Each entry has the same shape as a `get_ledger_transactions` row. Use
`from_account` / `to_account` (not `sender`/`receiver` — those are older
names).

Uses: toast "payment received", update activity feed optimistically, trigger
downstream actions (e.g., ship the product once `to_account === me`).

### `asu` — App Session Update

Fires when an app session state changes — new `submit_app_state`, close,
deposit, withdraw.

Payload:
```json
{
  "app_session": {
    "app_session_id": "0x...",
    "application": "my-game",
    "status": "open",
    "participants": ["0xA", "0xB"],
    "weights": [50, 50],
    "quorum": 100,
    "protocol": "NitroRPC/0.4",
    "challenge": 86400,
    "version": 5,
    "nonce": 17,
    "session_data": "{\"round\": 3}",
    "created_at": "…",
    "updated_at": "…"
  },
  "participant_allocations": [
    { "participant": "0xA", "asset": "usdc", "amount": "120.0" },
    { "participant": "0xB", "asset": "usdc", "amount": "80.0"  }
  ]
}
```

Uses: reactive multi-party UIs — opponent made a move, pot just grew,
judge's ruling posted.

## Subscription

Notifications arrive the moment you authenticate — no explicit subscribe
call. Just register a handler at the WebSocket layer:

```ts
client.onNotification('bu',  (p) => updateBalancesUI(p.balance_updates));
client.onNotification('cu',  (p) => updateChannelUI(p));
client.onNotification('tr',  (p) => showTransferToast(p.transactions));
client.onNotification('asu', (p) => updateSessionUI(p.app_session, p.participant_allocations));
```

## Patterns

### Balance cache invalidation

```ts
let balanceCache: Balance[] = [];

async function warmCache() {
  const { ledger_balances } = await client.send('get_ledger_balances', {});
  balanceCache = ledger_balances;
}

client.onNotification('bu', ({ balance_updates }) => {
  for (const b of balance_updates) {
    const idx = balanceCache.findIndex((x) => x.asset === b.asset);
    if (idx >= 0) balanceCache[idx] = b;
    else balanceCache.push(b);
  }
  renderBalances(balanceCache);
});
```

### Outgoing transfer confirmation (race-free)

```ts
const pending = new Map<string, {resolve: Function, timer: any}>();

client.onNotification('tr', ({ transactions }) => {
  for (const t of transactions) {
    const key = `${t.to_account}-${t.asset}-${t.amount}`;
    const waiter = pending.get(key);
    if (waiter) {
      clearTimeout(waiter.timer);
      pending.delete(key);
      waiter.resolve(t);
    }
  }
});

async function sendAndConfirm(destination, allocations) {
  const res = await client.send('transfer', { destination, allocations }, signer);
  // Wait for the matching `tr` notification to confirm the other side saw it
  return new Promise((resolve, reject) => {
    const key = `${destination}-${allocations[0].asset}-${allocations[0].amount}`;
    const timer = setTimeout(() => { pending.delete(key); reject(new Error('tr timeout')); }, 5000);
    pending.set(key, { resolve, timer });
  });
}
```

### Dispute watcher

```ts
client.onNotification('cu', (channel) => {
  if (channel.status === 'challenged') {
    alertOperator(`Channel ${channel.channel_id} is under challenge`);
    pageOncall();
  }
});
```

## Gotchas

- **Don't route `id = 0` to pending callbacks.** A notification is NOT a response.
- **`bu.amount` is the new total**, not a delta. If your UI shows deltas, compute them by diffing against your last known balance.
- **`tr` fires for both sides.** Filter by `from_account === me` (outgoing) vs `to_account === me` (incoming) before showing toasts — otherwise you'll double-notify.
- **`cu` fires on EVERY state change**, including intermediate states during resize. Debounce if your UI is expensive to render.
- **No ordering guarantees across notifications of different types**, but same-type notifications arrive in order. A `bu` and `tr` for the same event may arrive in either order — don't assume.
- **Reconnects drop queued notifications.** After a reconnect, call `get_ledger_balances`, `get_channels`, and `get_app_sessions` to hydrate — you may have missed events while disconnected.

## Related

- `yellow-nitro-rpc` — envelope format (why `requestId = 0`)
- `yellow-queries` — polling counterparts; shapes match 1:1 with notification payloads
- `yellow-transfers` — the source of `tr` events
- `yellow-app-sessions` — the source of `asu` events
- `yellow-state-channels` — the source of `cu` events
