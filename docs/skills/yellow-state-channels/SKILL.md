---
name: yellow-state-channels
description: |
  On-chain state channel protocol patterns for Yellow Network (ERC-7824 / Nitrolite). Covers channel lifecycle (open, operate, close), cooperative and unilateral close paths, on-chain contracts, data structures, challenge-response, app sessions vs payment channels, and the security model. Use when: working with channel creation/closing, debugging disputes, understanding the on-chain security layer, or building custom adjudicators.
version: 3.0.0
sdk_version: "@erc7824/nitrolite@^0.5.3"
network: mainnet
last_verified: 2026-04-26
user-invocable: true
source_urls:
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/off-chain/channel-methods
  - https://docs.yellow.org/docs/0.5.x/protocol/app-layer/on-chain/data-structures
  - https://docs.yellow.org/docs/0.5.x/guides/migration-guide
---

# Yellow Network State Channels

**Protocol**: Nitrolite v0.5.0 (ERC-7824)
**On-chain domain**: `VirtualApp:Custody` v0.3.0

State channels are a Layer-2 scaling technique: lock funds on-chain, transact
freely off-chain with signed state updates, settle final state on-chain. The
Custody contract is the security anchor; disputes always have an on-chain
escape hatch.

## When to use state channels

Parties are **identified**, interactions are **frequent**, and **zero gas per
operation** is required. A 10-USDC chess game off-chain costs $0 in gas
regardless of move count; on-chain the same game would cost $100s.

| Property | Value |
|---|---|
| Finality | < 1 second (signature verification) |
| Cost per off-chain op | $0 |
| Throughput | Unlimited during active phase |
| Security | Funds recoverable via on-chain dispute |

## ⚠️ v0.5.0 breaking changes (read first if migrating)

1. **Channels open ZERO-balance.** Call `resize_channel` with positive
   `resize_amount` to fund. `create_channel` no longer accepts an initial
   amount.
2. **State signatures switched participants.** Pre-v0.5: participant was
   the session key. Post-v0.5: participant is the **wallet**; signed by the
   **wallet**. Pre-v0.5 signing code is rejected on-chain.
3. **`Signature` struct replaced by `Hex`** to support EIP-1271 (smart
   wallets) + EIP-6492 (pre-deployment sigs).
4. **Channel-locked funds are separate from unified balance** — users with
   a non-zero channel amount cannot `transfer` or deposit into app-sessions
   directly from that locked amount.
5. **Session-key field renames**: `app_name → application`, `expire →
   expires_at`, `allowances[]` now `{asset, amount}` objects.
6. **Protocol wire format** (0.3 → 0.5): params moved from positional array
   to named object inside the `req` envelope.

## Channel lifecycle (on-chain state machine)

```
VOID ──create() (2 sigs) ─────────────▶ ACTIVE   (implicit join, recommended)
VOID ──create() (1 sig) ──────────────▶ INITIAL ──join() ──▶ ACTIVE  (legacy)

ACTIVE ──off-chain state updates ─────▶ ACTIVE   (unlimited, zero gas)
ACTIVE ──resize() ────────────────────▶ ACTIVE   (add/remove funds)
ACTIVE ──checkpoint() ────────────────▶ ACTIVE   (anchor state on-chain)
ACTIVE ──close() (cooperative) ───────▶ FINAL    (both sign, instant)
ACTIVE ──challenge() ─────────────────▶ DISPUTE  (unilateral)

DISPUTE ──checkpoint(newer state) ────▶ ACTIVE   (dispute resolved)
DISPUTE ──close() after timeout ──────▶ FINAL    (challenged state wins)
```

## Two enums people confuse — keep them straight

**`StateIntent`** — "what kind of state is this?" (ABI-integer order matters)

| Value | Name | Meaning |
|---|---|---|
| 0 | `OPERATE` | Normal update |
| 1 | `INITIALIZE` | Channel creation |
| 2 | `RESIZE` | Allocation delta |
| 3 | `FINALIZE` | Closure |

**`Status`** — "what phase is the channel in?"

- On-chain (5-valued): `VOID=0, INITIAL=1, ACTIVE=2, DISPUTE=3, FINAL=4`
- `@yellow-org/sdk` v1 (4-valued): `Void=0, Open=1, Challenged=2, Closed=3`
  (INITIAL+ACTIVE collapse to `Open`, DISPUTE→`Challenged`, FINAL→`Closed`)

v1 SDK also uses a different **intent vocabulary**: `INTENT_OPERATE=0,
INTENT_CLOSE=1, INTENT_DEPOSIT=2, INTENT_WITHDRAW=3`, plus 4–9 for
escrow/migration. Don't mix with the on-chain `StateIntent` enum above.

## Closing a channel

**Cooperative (preferred)**: all participants sign a state with `intent =
FINALIZE`, one participant calls `Custody.close()`. Funds distribute
immediately, one transaction.

**Unilateral (dispute path)**:
1. Participant A calls `challenge(channelId, latestState, sigs)` — `ACTIVE →
   DISPUTE`, timer starts (minimum 1 hour, typically 24 h).
2. Within the window, any party may submit a **newer** state via
   `checkpoint()` — contract accepts higher `version`, reverts to `ACTIVE`.
3. If the window elapses with no newer state, any party calls `close()` —
   the challenged state wins; `DISPUTE → FINAL`.

**Iron rule**: higher `version` always wins. Withholding a newer state
doesn't help; the honest party submits their signed copy.

## Quick start

```solidity
// 1. Create (implicit join — 2 sigs collected off-chain first)
Custody.create(channel, initialState, userSig, clearnodeSig);

// 2. Fund via resize (v0.5+)
Custody.resize(channelId, resizeState, [userSig, clearnodeSig]);

// 3. Operate — off-chain, signed state exchange (Nitro RPC)

// 4. Close cooperatively
Custody.close(channelId, finalState, userSig, clearnodeSig);
```

## Navigation Guide

### When to read supporting files

**reference.md** — read when you need:

- Full Solidity interface signatures (`IChannel`, `IDeposit`, `IChannelReader`,
  `IAdjudicator`)
- Complete struct layouts with byte-level encoding rules
- Packed-state canonical signing payload
- Challenge-response protocol step-by-step with state-transition invariants
- Security model, threat matrix, liveness requirements
- Adjudicator customization for app-specific state validation
- Supported signature formats (Raw ECDSA / EIP-191 / EIP-712 / EIP-1271 /
  EIP-6492)

## Related skills

- `yellow-custody-contract` — Solidity ABI and events
- `yellow-deposits-withdrawals` — funding the unified balance
- `yellow-app-sessions` — off-chain multi-party account primitive
- `yellow-nitro-rpc` — wire format for off-chain state updates
- `yellow-sdk-v1` — high-level `Client.create()` API for v1 channel ops
