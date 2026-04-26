---
name: yellow-app-sessions
description: |
  App Sessions — the multi-party off-chain account primitive on Yellow Network. Covers create_app_session / submit_app_state / close_app_session, the `definition` (participants, weights, quorum, challenge, nonce, protocol), allocations math, and the three canonical patterns (escrow, stake, multi-player games). Use when: building 2-of-3 escrow for a marketplace, locking a stake with slashing authority, running a turn-based game with quorum signatures, or debugging why a session won't close.
version: 1.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/app-sessions
  - https://docs.yellow.org/docs/0.5.x/guides/multi-party-app-sessions
---

# Yellow App Sessions

An **App Session** is a shared off-chain account on top of the unified balance.
Multiple participants move funds in, update state under custom governance
(weights + quorum), and split the funds on close. Instant settlement, zero
gas, no on-chain interaction until/unless a dispute forces it.

Spec: <https://docs.yellow.org/docs/protocol/off-chain/app-sessions>.

## Lifecycle

```
create_app_session ──▶ [open]
                        │
                        │  submit_app_state (0..N updates)
                        ▼
                       [open, version=N]
                        │
                        │  close_app_session (quorum sigs)
                        ▼
                     [closed]   ──▶ funds released to unified balances
```

## `create_app_session` params

```json
{
  "definition": {
    "application": "my_app",
    "protocol":    "NitroRPC/0.4",
    "participants": ["0xA", "0xB", "0xC"],
    "weights":      [40, 40, 50],
    "quorum":       80,
    "challenge":    86400,
    "nonce":        1699123456789
  },
  "allocations": [
    { "participant": "0xA", "asset": "usdc", "amount": "100.0" },
    { "participant": "0xB", "asset": "usdc", "amount": "0.0" },
    { "participant": "0xC", "asset": "usdc", "amount": "0.0" }
  ],
  "session_data": "{\"initial\": true}"
}
```

### `definition` fields

| Field | Meaning |
|---|---|
| `application` | Your app identifier. |
| `protocol` | Nitro protocol version. Current: `NitroRPC/0.4`. |
| `participants[]` | Addresses that can sign state updates. Order is preserved and referenced by `weights` / `allocations`. |
| `weights[]` | Voting weight per participant for state updates & close. Must have same length as `participants`. |
| `quorum` | Minimum **sum of weights** required to approve a state update or close. If `sum(weights) = 100` and `quorum = 80`, you need any subset totaling ≥ 80. |
| `challenge` | Dispute window in **seconds** if anyone forces a unilateral close. |
| `nonce` | Unique per (participants) tuple; prevents replay. Usually `Date.now()`. |

### `allocations[]`

Initial funds locked in the session. **Sum of all asset amounts per asset =
total locked**. Participants' unified balances must cover their allocation.
All amounts are decimal strings.

### Signatures

`create_app_session` requires a signature from **every participant whose
allocation is non-zero** (they're moving funds). Put them in the envelope
`sig[]` array in the same order as `participants[]`.

### App registration (v1)

`@yellow-org/sdk@1.x` adds an app-registration step **before** the first
`create_app_session`. Call `client.registerApp(appID, metadata,
creationApprovalNotRequired)` once per app id. If
`creationApprovalNotRequired=false`, subsequent session creations require
explicit owner approval. This step doesn't exist in v0.5.3 — on v0.5.3 you
go straight to `create_app_session`.

### Quorum signature prefix bytes (v1)

In `@yellow-org/sdk@1.x`, quorum signatures carry a 1-byte type prefix:

- `0xA1` = `AppSessionWalletSignerV1` (wallet-scoped signature)
- `0xA2` = `AppSessionKeySignerV1` (session-key-scoped signature)

Raw 65-byte sigs without a type byte are rejected. v0.5.3 doesn't use
prefixes — so v0.5.3 and v1 quorum payloads are wire-incompatible.

### Session ID derivation

The `app_session_id` is deterministic:

```
keccak256(JSON.stringify({
  application, protocol, participants, weights, quorum, challenge, nonce
}))
```

`chainId` is **not** part of the hash — app sessions live entirely off-chain
on top of the unified balance, so the same `definition` produces the same
id regardless of which chain funded the unified balance.

## `submit_app_state` params

Update the session — change allocations or session_data — with quorum sigs.
**`allocations` is the FINAL state, not a delta.**

```json
{
  "app_session_id": "0xabc...",
  "intent": "operate",          // "operate" | "deposit" | "withdraw"  (v0.4 only)
  "version": 2,                  // MUST equal current_version + 1  (v0.4 only)
  "allocations": [...],          // FINAL distribution (not delta)
  "session_data": "{\"round\": 2}"
}
```

### Intent rules (NitroRPC/0.4)

| Intent | Total invariant | Signatures required |
|---|---|---|
| `operate` | sum before = sum after | combined signer weight ≥ `quorum` |
| `deposit` | sum after **≥** sum before | quorum **AND** depositing participant must sign |
| `withdraw` | sum after **≤** sum before | combined signer weight ≥ `quorum` |

(The spec's wording allows equality on deposit/withdraw — a same-total
deposit is a no-op but not rejected.)

A `deposit` bumps one participant's allocation above their previous one and
moves the delta from their unified balance; they must co-sign regardless of
whether their weight was needed for quorum.

For NitroRPC/0.2, omit `intent` and `version` entirely — the protocol
version dictates the payload shape.

## `close_app_session` params

```json
{
  "app_session_id": "0xabc...",
  "allocations": [
    { "participant": "0xA", "asset": "usdc", "amount": "180.0" },
    { "participant": "0xB", "asset": "usdc", "amount": "15.0" },
    { "participant": "0xC", "asset": "usdc", "amount": "5.0" }
  ],
  "session_data": "{\"result\": \"A wins\"}"
}
```

**Iron rule**: `sum(allocations.amount) per asset === total locked in session
for that asset`. One cent off and the server rejects. If funds were added/
removed via `submit_app_state`, use the current total.

Response: `{ app_session_id, status: "closed", version }`. Funds are released
instantly to each participant's unified balance.

## Canonical patterns

### Pattern A — 2-of-3 escrow (marketplace)

```
participants: [Buyer, Seller, Judge]
weights:      [40,    40,     50]     // sum = 130
quorum:       80                       // any 2 parties
```

Happy path: Buyer + Seller agree on delivery → they sign close together
(40+40=80, meets quorum). Judge is idle.

Dispute: Buyer or Seller pairs with Judge → 40+50=90, meets quorum. Judge's
weight gives either party + judge authority to close.

Locked: Buyer contributes the full amount. Seller + Judge contribute 0. On
close, allocations reflect the resolution (100/0, 0/100, or split).

### Pattern B — Stake (slashing)

```
participants: [Agent, Treasury, SlashAuthority]
weights:      [0,     0,        100]
quorum:       100
```

Agent locks collateral. Only the SlashAuthority can unilaterally redirect
funds (e.g., a slashing DAO or the protocol operator). Used for tier stakes.

Happy path: Agent wants to unstake → SlashAuthority signs a close returning
100% to Agent's balance.

Slash: SlashAuthority signs a close redirecting N% to Treasury.

### Pattern C — Turn-based game

```
participants: [Player1, Player2]
weights:      [50,       50]
quorum:       100                      // both must sign every state
```

Every move is a `submit_app_state` both players sign. Close on game over with
allocations matching the outcome (winner takes pot, ties split).

For games with unknown duration, set `challenge` large enough that an
unresponsive player can be forced into the on-chain challenge flow without
losing the pot.

## Reading an active session

```ts
const { app_sessions } = await client.send('get_app_sessions', {
  participant: myAddress,   // optional filter
  status: 'open',           // optional filter
});
// app_sessions[] → { app_session_id, status, version, allocations,
//                    session_data, definition, participants, ... }
```

Use before signing a close to confirm the total and current version. The
method is **plural** (`get_app_sessions`) and returns a list — filter or
match the `app_session_id` you care about.

## Common failures

| Symptom | Cause |
|---|---|
| `error: insufficient balance` on create | A participant's unified balance is below their allocation. |
| `error: quorum not met` | Envelope `sig[]` missing signatures, or summed weights < quorum. |
| `error: allocation mismatch` | Close total ≠ locked total for one or more assets. |
| `error: version conflict` | Two concurrent `submit_app_state` with the same version. Refetch, bump version. |
| Session appears open after close | Close was rejected (check the `res` / `error` message), not actually closed. |

## Related skills

- `yellow-clearnode-auth` — connect first
- `yellow-transfers` — direct transfers (no session)
- `yellow-state-channels` — the on-chain layer that backs unified balance
- `yellow-sdk-v1` / `yellow-sdk-api` — typed SDKs that wrap these calls
