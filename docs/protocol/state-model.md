# State Model

Previous: [Cryptography](cryptography.md) | Next: [Channel Protocol](channel-protocol.md)

---

This document describes the abstract structure of protocol states.

It explains how states are defined and structured. Operational flows are described in separate documents.

## Purpose

States represent the current agreed configuration of protocol entities. The state model defines:

- what information a state contains
- how states are identified and versioned
- how states are represented for off-chain and on-chain use

## Common State Fields

All protocol states share the following common properties:

| Field    | Description                                                    |
| -------- | -------------------------------------------------------------- |
| EntityId | 32-byte unique identifier of the entity this state belongs to  |
| Version  | 64-bit unsigned integer, monotonically increasing              |

In addition to these common fields, each state contains entity-specific data whose structure varies depending on the entity type and use case. The entity-specific data is defined by the respective entity specification.

## State Identification and Versioning

Each state is identified by the combination of its entity identifier and version number.

Rules:

- The entity identifier is derived from the entity definition and is immutable
- The version MUST start at 1 for the initial state
- Versions are strictly increasing; the exact increment rule depends on the context:
  - Off-chain state advancement requires each new version to be exactly the previous version plus one
  - On-chain enforcement requires only that the submitted version be strictly greater than the currently recorded on-chain version

## Channel State

The channel state is the primary protocol state. It represents the current configuration of a channel.

| Field         | Description                                            |
| ------------- | ------------------------------------------------------ |
| ChannelId     | 32-byte identifier derived from the channel definition |
| Metadata      | 32-byte  Hash of channel metadata                      |
| Version       | 64-bit unsigned integer, state version                 |
| HomeLedger    | Asset allocations on the home chain                    |
| NonHomeLedger | Asset allocations on the non-home chain                |
| Transition    | Describes the operation that produced this state       |
| UserSig       | User signature for the state                           |
| NodeSig       | Node signature for the state                           |

The channel identifier encodes a protocol version byte as its first byte, followed by the hash of the channel definition parameters. This ensures uniqueness across protocol deployments.

### Ledger

A ledger records asset allocations for a specific blockchain within a channel. Each channel state contains exactly two ledgers: a home ledger and a non-home ledger.

| Field          | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| ChainId        | Identifier of the blockchain this ledger is associated with  |
| Token          | Token contract address on this chain                         |
| Decimals       | Decimal precision of the token on this chain                 |
| UserAllocation | Amount allocated to the user                                 |
| UserNetFlow    | Cumulative net flow for the user (may be negative)           |
| NodeAllocation | Amount allocated to the node                                 |
| NodeNetFlow    | Cumulative net flow for the node (may be negative)           |

**Ledger invariant:** A ledger MUST satisfy the following invariant at all times:

```
UserAllocation + NodeAllocation == UserNetFlow + NodeNetFlow
```

UserNetFlow tracks the cumulative net amount that has flowed into or out of the user's position through deposits, withdrawals, and cross-chain operations. NodeNetFlow tracks the cumulative net amount that has flowed through the node's position, including transfers, commits, and releases. Allocations represent the current distributable balances. The invariant ensures that the total distributable balance always equals the total cumulative flows — no assets can be created or destroyed through state transitions.

All allocation values MUST be non-negative. Net flow values MAY be negative, reflecting outbound transfers or withdrawals that exceed inbound flows.

### Empty Non-Home Ledger

When a channel state does not involve cross-chain operations, the non-home ledger MUST be empty. An empty non-home ledger is defined as a ledger where all fields are set to their zero values:

| Field          | Value                                      |
| -------------- | ------------------------------------------ |
| ChainId        | 0                                          |
| Token          | Zero address (0x0000...0000)               |
| Decimals       | 0                                          |
| UserAllocation | 0                                          |
| UserNetFlow    | 0                                          |
| NodeAllocation | 0                                          |
| NodeNetFlow    | 0                                          |

An empty non-home ledger is structurally present but zeroed. A non-home ledger with metadata (non-zero ChainId or Token) but zero balances is NOT considered empty.

## Off-Chain Representation

The off-chain representation is the primary operational format of a channel state. It is the representation exchanged between participants during state advancement, and it is the representation that is signed.

The off-chain representation contains all channel state fields directly, including the full transition data (type, transaction identifier, account identifier, and amount). This representation is optimized for human readability, ease of validation, and efficient signature generation.

## Enforcement Representation

The off-chain and on-chain (enforcement) representations depict the **same logical state**. The on-chain (enforcement) representation is derived deterministically from the off-chain one — no additional information is required.

When a state is submitted to the blockchain layer, it uses an enforcement representation optimized for on-chain verification, gas efficiency, and deterministic encoding.

The following fields are preserved exactly from the off-chain representation:

- Version
- Home and non-home ledger fields (ChainId, Token, Decimals, UserAllocation, UserNetFlow, NodeAllocation, NodeNetFlow)

The following fields are derived:

- **Intent** — derived from the transition type via the intent mapping table
- **MetadataHash** — the Keccak-256 hash of the ABI-encoded transition data (type, transaction identifier, account identifier, and amount). This captures all off-chain transition information in a single hash, ensuring that the enforcement representation is bound to the specific transition without transmitting the full transition data on-chain.

The enforcement representation is constructed by packing these fields into an ABI-encoded structure:

```
SignablePayload = AbiEncode(ChannelId, AbiEncode(Version, Intent, MetadataHash, HomeLedger, NonHomeLedger))
```

Where each ledger is encoded as a tuple of (chain identifier, token address, decimals, user allocation, user net flow, node allocation, node net flow).

Because the mapping is deterministic, both the off-chain and enforcement representations produce the same message digest when signed, ensuring that a signature over the off-chain state is valid for enforcement and vice versa.

## Intent Mapping

Each transition type maps to an intent value used in the enforcement representation. The intent determines how the blockchain layer processes the state.

| On-chain Intent                    | Transition                  |
| -------------------------- | --------------------------- |
| OPERATE                    | TransferSend, TransferReceive, Commit, Release, Acknowledgement |
| CLOSE                      | Finalize                    |
| DEPOSIT                    | Home Deposit                |
| WITHDRAW                   | Home Withdrawal             |
| INITIATE_ESCROW_DEPOSIT    | Escrow Deposit Initiate     |
| FINALIZE_ESCROW_DEPOSIT    | Escrow Deposit Finalize     |
| INITIATE_ESCROW_WITHDRAWAL | Escrow Withdrawal Initiate  |
| FINALIZE_ESCROW_WITHDRAWAL | Escrow Withdrawal Finalize  |
| INITIATE_MIGRATION         | Migration Initiate          |
| FINALIZE_MIGRATION         | Migration Finalize          |

Transitions that map to the OPERATE intent do not require on-chain checkpointing under normal operation.

## Transition Field

Each state update includes a transition that describes the operation that produced the new state.

| Field     | Description                                                      |
| --------- | ---------------------------------------------------------------- |
| Type      | Transition type identifier                                       |
| TxId      | Transaction identifier hash                                      |
| AccountId | Context-dependent account identifier (varies by transition type) |
| Amount    | Amount involved in the transition                                |

The transition type determines the validation rules applied to the state update. The account identifier carries different semantics depending on the transition type — for example, it references the channel identifier for deposit and withdrawal operations, the counterparty address for transfers, or the application session identifier for commit and release operations.

## State Consistency Rules

State validity requirements differ between off-chain advancement and on-chain enforcement contexts. Off-chain advancement rules are defined in the [Channel Protocol](channel-protocol.md) document, and on-chain enforcement rules are defined in the [State Enforcement](enforcement.md) document.

In both contexts, the following invariants MUST hold:

- The entity identifier MUST match the entity definition
- The version MUST be strictly greater than the previously accepted version
- Ledger invariants MUST be satisfied (allocations equal net flows, allocation values non-negative)

---

Previous: [Cryptography](cryptography.md) | Next: [Channel Protocol](channel-protocol.md)
