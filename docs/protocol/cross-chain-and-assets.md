# Cross-Chain and Asset Model

Previous: [Enforcement and Settlement](enforcement.md) | Next: [Interactions](interactions.md)

---

This document describes the unified asset model and cross-chain functionality.

## Purpose

The unified asset model allows participants to operate on assets from multiple blockchains within a single channel. This eliminates the need for separate channels per blockchain and enables cross-chain interactions.

## Unified Asset Concept

Assets in the Nitrolite protocol are identified independently of any specific blockchain.

A unified asset is defined by:

| Field    | Description                                        |
| -------- | -------------------------------------------------- |
| Symbol   | Human-readable canonical asset identifier (e.g. "USDC") |
| Decimals | Decimal precision of the asset                     |

### Canonical Asset Identification

The protocol identifies a unified asset by its symbol. Within channel metadata, the symbol is represented as the first 8 bytes of its Keccak-256 hash, providing a compact canonical identifier. Two chain-specific tokens are recognized as the same unified asset if they share the same symbol-derived identifier and are configured as such by the node.

Symbol collisions are prevented by the node's asset configuration. The protocol does not maintain a global on-chain registry of unified assets.

### Amount Normalization

Assets on different blockchains MAY have different decimal precisions (e.g. USDC has 6 decimals on Ethereum but may have different precision on other chains). The protocol normalizes amounts for cross-chain comparisons using WAD normalization, which scales chain-specific amounts as if a token had 18 decimals:

```
NormalizedAmount = Amount * 10^(18 - ChainDecimals)
```

Each unified asset defines a canonical decimal precision (e.g. 6 for USDC) that is used during User <> Clearnode interactions (e.g. on-chain deposit, on-chain state submission requests, transfers, app session operations etc.).

Rules:

- Normalization is used **only for cross-chain comparisons** (e.g. validating that escrow amounts match across chains). It is not used for storage or accounting — stored values remain in their chain-native precision.
- The asset's configured decimal precision acts as the base, whereas 18 is the target of the upscaling. The maximum supported decimal precision is 18.
- Normalization is exact and lossless when scaling up. No rounding or remainder occurs.
- The blockchain layer validates that declared decimals match the actual token decimals on the current chain.

## Home Chain

The home chain is the blockchain against which a given channel state is enforced. It is identified by the chain identifier in the home ledger of that state.

The home chain determines:

- where enforcement operations for that state are executed
- which blockchain holds the locked funds for the channel
- the authoritative source for state validation

The home chain MAY change over the lifetime of a channel through a migration operation. After migration, the new home chain becomes the authoritative enforcement target.

## Home and Non-Home Ledger Roles

**Home Ledger**
The home ledger is the primary record of asset allocations. It is associated with the home chain and is directly enforceable through the blockchain layer.

Responsibilities:

- tracks the authoritative asset allocations
- receives checkpoints for enforcement
- holds deposited assets in the enforcement contract

**Non-Home Ledger**
The non-home ledger tracks asset allocations on a blockchain other than the home chain. When no cross-chain operation is in progress, the non-home ledger MUST be empty (see [Empty Non-Home Ledger](state-model.md#empty-non-home-ledger)).

Responsibilities:

- tracks assets involved in cross-chain escrow operations
- reflects cross-chain deposit and withdrawal allocations
- coordinates with the home ledger for consistency

## Escrow Model

Cross-chain operations use an **escrow** mechanism to coordinate fund movements across two independent blockchains.

An escrow is a temporary on-chain record that locks funds on one chain while a corresponding state update is being finalized on the other chain. Each escrow is identified by an **escrow channel identifier**, derived deterministically from the home channel identifier and the state version at initiation.

| Property       | Description                                                     |
| -------------- | --------------------------------------------------------------- |
| Identifier     | 32-byte hash derived from the home channel identifier and state version |
| Hosting chain  | The non-home chain (for deposits: where the user's funds are locked; for withdrawals: where the node's funds are locked) |
| Tracked amount | The amount locked in escrow, corresponding to the non-home ledger allocations |
| Unlock delay   | Escrow deposits include an unlock delay after which funds are automatically unlocked to the node if not challenged |
| ChallengeDuration | A period after a challenge was initiated that allows resolution. If no finalization state was supplied, the initiate state is finalized, and funds are returned |

An escrow is not a separate protocol entity with its own state — it is an on-chain record derived from a channel state transition. The escrow exists only between initiation and finalization (or timeout).

## Cross-Chain Deposit

To deposit assets from a non-home chain into a channel, the protocol uses a two-phase escrow process:

1. **Initiate (Escrow Deposit Initiate)** — participants sign a state that creates an escrow. On the home chain, the node's allocation increases to reserve funds. On the non-home chain, the user's deposit is locked in an escrow record with an unlock delay.
2. **Finalize (Escrow Deposit Finalize)** — after the escrow is created, participants sign a state that completes the deposit. On the home chain, the user's allocation increases by the deposited amount. On the non-home chain, the escrowed funds are released to the node's vault.

If the escrow is not finalized within the unlock delay, the escrowed funds on the non-home chain are automatically unlocked to the Node. Either participant MAY challenge the escrow during the challenge period. Note that it is NOT possible to challenge a deposit escrow after unlock delay has passed as the funds were already unlocked to the Node.

Cross-chain amounts are validated using WAD normalization to ensure the home-chain node allocation matches the non-home-chain user deposit.

## Cross-Chain Withdrawal

To withdraw assets to a non-home chain, the protocol uses a similar two-phase escrow process:

1. **Initiate (Escrow Withdrawal Initiate)** — participants sign a state that creates an escrow. On the non-home chain, the node locks funds from its vault into the escrow record.
2. **Finalize (Escrow Withdrawal Finalize)** — participants sign a state that completes the withdrawal. On the home chain, the user's allocation decreases. On the non-home chain, the escrowed funds are released to the user.

If the escrow is not finalized cooperatively, either participant MAY challenge the escrow.

## Home Chain Migration

The home chain of a channel MAY be changed through a two-phase migration process:

1. **Initiate (Migration Initiate)** — participants sign a state that begins the migration. On the current home chain, the state records the target chain allocation. On the target chain, a new channel record is created with status "migrating in" and the node locks funds equal to the user's allocation (validated via WAD normalization).
2. **Finalize (Migration Finalize)** — participants sign a state that completes the migration. On the new home chain, the channel transitions to operating status. On the old home chain, all locked funds are released to the node and the channel is marked as migrated out.

After migration, the following changes take effect:

- **Home chain identifier** is updated to reflect the migration
- **Home token address** is updated to reflect the migration
- **Ledger roles** — the former non-home ledger becomes the home ledger; the former home ledger becomes the non-home ledger (and its allocations are zeroed out on finalization)
- **Enforcement target** — all subsequent enforcement operations execute against the new home chain
- **Balances** — the user's allocation is preserved (normalized by decimal precision); the node's allocation is recalculated for the new chain

VERSION NOTE: Migration transitions are functional but may be refined in future protocol versions.

## Cross-Chain Replay Protection

The protocol prevents cross-chain replay through multiple binding mechanisms:

- **Chain identifier binding** — each ledger is bound to a specific chain identifier. The blockchain layer validates that the home ledger chain identifier matches the current blockchain. This prevents a state signed for one chain from being enforced on another.
- **Channel identifier scoping** — channel identifiers incorporate a protocol version byte, preventing replay across smart contract deployments. The same channel definition on a different protocol version produces a different channel identifier.
- **Escrow identifier uniqueness** — escrow channel identifiers are derived from the home channel identifier and the state version at initiation. This ensures that each escrow operation produces a unique identifier, preventing a completed escrow from being replayed.
- **Ledger validation** — on-chain enforcement validates that both home and non-home ledger's declared decimals match the actual token decimals on the current execution chain, preventing states crafted for a different token from being accepted. Additionally, a specific set of invariants is enforced for security purposes.

## Current Version Notes

In the current protocol version:

- Cross-chain operations require trust in the node to relay state correctly between chains. The node is responsible for submitting escrow initiation and finalization transactions on the appropriate chains.
- Full cross-chain enforcement (trustless bridging) is a planned future improvement.
- Each channel state supports exactly two ledgers: one home ledger and one non-home ledger. This is a V1-specific design constraint; future protocol versions MAY support additional ledger configurations.

---

Previous: [Enforcement and Settlement](enforcement.md) | Next: [Interactions](interactions.md)
