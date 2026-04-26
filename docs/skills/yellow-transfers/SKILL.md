---
name: yellow-transfers
description: |
  Off-chain peer-to-peer transfers on Yellow Network — instant, gasless, cross-asset. Covers the `transfer` Nitro RPC method, allowances enforcement, destination vs destination_user_tag, multi-asset in one call, unified balance semantics, delivery receipts, notifications (asu, bu), and idempotency. Use when: sending funds between agents, implementing tipping, paying for a marketplace task, routing a swap leg, or debugging why a transfer didn't show up.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/transfers
---

# Yellow Transfers

A **transfer** moves value from one unified balance to another in a single
Nitro RPC. No gas, no app session, sub-200ms. The primitive behind tips,
payments, marketplace settlements, and swap legs.

Spec: <https://docs.yellow.org/docs/protocol/off-chain/transfers>.

## Params

```json
{
  "destination": "0x8B3192f2F7b1b34f2e4e7B8C9D1E0F2A3B4C5D6E",
  "allocations": [
    { "asset": "usdc", "amount": "50.0" },
    { "asset": "eth",  "amount": "0.01" }
  ]
}
```

| Field | Required | Notes |
|---|---|---|
| `destination` | one-of | Recipient wallet address (0x hex). |
| `destination_user_tag` | one-of | Recipient user tag (alphanumeric). Required if `destination` omitted. |
| `allocations[]` | ✓ | At least one `{asset, amount}`. Multi-asset in one call is atomic. |

### Allocation semantics

- `asset`: **lowercase** symbol (`usdc`, `eth`, `weth`, `btc`, `yellow`, …).
  Use `get_assets` to enumerate the current network's supported symbols.
- `amount`: decimal string in human-readable units (e.g., `"50.0"` for 50
  USDC, not `50000000`). ClearNode handles the conversion to smallest unit
  internally.
- All allocations are debited from the sender and credited to the recipient
  in a single atomic operation. Partial failure is impossible — either the
  whole multi-asset transfer succeeds or none of it does.

## Signatures

Signed by the **session key** of the sender, carried in the envelope `sig[]`
(standard keccak256-of-req-bytes signature, see `yellow-nitro-rpc`). The
recipient does not sign.

## Response

Returns an object wrapping an array of `LedgerTransaction` records — one
per asset moved (same shape as the `tr` notification and
`get_ledger_transactions` rows):

```json
{
  "transactions": [
    {
      "id": 7023,
      "tx_type": "transfer",
      "from_account":     "0x742d35Cc...",
      "from_account_tag": "@alice",
      "to_account":       "0x8B3192f2...",
      "to_account_tag":   "@bob",
      "asset":  "usdc",
      "amount": "50.0",
      "created_at": "2026-04-23T12:00:00Z"
    }
  ]
}
```

If you passed two allocations in one `transfer` call, the response array has
two entries (one per asset). There's no single `transaction_id` / `status`
wrapper — each asset transfer is its own row. Use `id` as your stable
reference for receipts.

Failures don't arrive here — they come as an `error` response with
descriptive text (see `yellow-errors`).

## Notifications

After a successful transfer, both sides receive unsolicited `res` messages
with `id = 0`:

```json
{ "res": [0, "bu", { "balance_updates": [{ "asset": "usdc", "amount": "150.0" }] }, <ts>] }
// "bu" = balance update. `amount` is the NEW TOTAL unified balance (not a delta).

{ "res": [0, "tr", { "transactions": [{ "id": 7023, "tx_type": "transfer", ... }] }, <ts>] }
// "tr" = transfer. Fired on both sender and recipient sockets; same row shape as the response.
```

Subscribe to these to keep your local balance UI in sync without polling.
See `yellow-notifications` for the full payload catalog.

## Allowances enforcement

The sender's session key must have an allowance ≥ the amount for each asset.
Aggregate spend across multiple transfers counts against the same allowance.

```json
// Session key has allowances: [{asset: "usdc", amount: "100.0"}]
// Successful first transfer:   50 USDC → used 50, remaining 50
// Successful second transfer:  30 USDC → used 80, remaining 20
// Attempted third:             50 USDC → REJECTED (exceeds remaining 20)
```

Rotation with a new allowance resets the meter. See `yellow-session-keys`.

## Idempotency

Each row in the response has a stable numeric `id`. If your client crashed
between send and receive, do NOT just retry — query first via
`get_ledger_transactions`:

```ts
const { ledger_transactions } = await client.send('get_ledger_transactions', {
  account_id: myAddress,
  tx_type: 'transfer',
  limit: 50,
});
// Look for a matching (receiver, amount, ~timestamp) tuple.
```

Or, if you generated a client-side unique ID (e.g., an invoice ID), include
it in a follow-up metadata call so you can correlate on retry.

## `destination` vs `destination_user_tag`

- **destination** (wallet): direct, canonical. Use when you know the address.
- **destination_user_tag**: user-level alias (e.g., `@alice`). Use for UX
  flows where users pick friends by handle. Tag → address resolution happens
  server-side.

## Multi-asset atomicity

Multiple allocations in one call are settled atomically. This is the simple
way to implement swap legs, combined fees, or payment + tip:

```json
{
  "destination": "0xSeller",
  "allocations": [
    { "asset": "usdc", "amount": "100.0" },     // payment
    { "asset": "yellow", "amount": "1.0" }       // tip
  ]
}
```

If the sender has insufficient balance on **any** asset, the entire transfer
is rejected — no partial credit.

## Common failures

| Symptom | Cause |
|---|---|
| `error: insufficient balance` | Sender unified balance < amount for one of the assets. |
| `error: allowance exceeded` | Cumulative session key spend exceeded the allowance cap. Rotate. |
| `error: unsupported asset` | Asset symbol not on this network. Check `get_assets`. Case matters — use lowercase. |
| `error: invalid destination` | Neither `destination` nor `destination_user_tag` was valid. |
| Transfer "succeeds" but recipient balance didn't change | You're reading from a stale cache or wrong chain. Subscribe to `bu` notifications. |
| `error: self-transfer` | `destination` equals the authenticated sender. Not permitted. |

## Related

- `yellow-clearnode-auth` — connect first
- `yellow-session-keys` — allowances + scopes
- `yellow-app-sessions` — multi-party escrow when you need more than a raw transfer
