# Example: Two-Party Escrow

> Demonstrates that V1 app session functionality is fully preserved in V2.
> This app could be built identically with V1's fixed quorum and allocation model.

## Concept

Buyer and seller agree on a trade. Buyer deposits funds into escrow. Seller
delivers goods/services off-chain. Buyer confirms receipt, releasing funds to
seller. Either party can dispute.

## Definition

```
Metadata: { buyer: address, seller: address, amount: decimal, asset: string }
Module:   escrow_v1 (simple two-party quorum, deterministic state machine)
```

## Data Schema

```
{ status: "awaiting_deposit" | "funded" | "confirmed" | "disputed" }
```

## Lifecycle

```
v1  buyer creates session             Data: { status: "awaiting_deposit" }
    Signers: [buyer]

v2  buyer deposits                    Data: { status: "funded" }
    VouchersUsed: [voucher from buyer's channel]
    Signers: [buyer]

v3  buyer confirms delivery           Data: { status: "confirmed" }
    Signers: [buyer]
    Module issues voucher to seller → Close: true
```

## Module Logic

- v1: verify buyer signed, status transition valid.
- v2: verify buyer signed, exactly one voucher used matching the agreed amount
  and asset. Transition awaiting_deposit → funded.
- v3: verify buyer signed, transition funded → confirmed. Issue voucher to
  seller's channel for the full amount. Signal close.

Dispute path: if buyer sends `{ status: "disputed" }`, module issues voucher
back to buyer and closes. A more sophisticated module could require both
signatures for dispute resolution.

## Why This Works in V1 Too

The escrow is sequential, two-party, and follows a fixed state machine. V1's
participants + quorum + allocation model handles this directly. The V2 version
uses the same pattern — Metadata replaces participants, the module replaces the
fixed quorum check, and vouchers replace allocation diffs. No new capabilities
are exercised.

This confirms V2 is a strict superset: anything V1 can do, V2 can do with an
equivalent module.
