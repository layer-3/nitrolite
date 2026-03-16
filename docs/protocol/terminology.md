# Terminology

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)

---

This document defines all protocol terms used throughout the Nitrolite protocol documentation.

Each term is defined once. All other documents MUST use these terms consistently.

## Naming Conventions

- Protocol entities use CamelCase (e.g., ChannelState, AppSession)
- Field names use CamelCase (e.g., ChannelId, StateVersion)
- Operations use lowercase with hyphens in document references (e.g., state-advancement)

## Core Entities

### Channel

A state container shared between a user and a node that allows off-chain state updates while maintaining on-chain security guarantees. Each channel operates on a single unified asset.

### Channel Definition

The immutable parameters that define a channel: user, node, asset, nonce, challenge duration, and approved signature validators. A channel definition is fixed at creation time and MUST NOT change during the channel lifecycle.

### Channel State

The current agreed configuration of a channel, including home and non-home ledger allocations, a version number, and a transition field. Channel state evolves through off-chain state advancement.

### Participant

An entity that holds a signing key and participates in a channel. Each channel has exactly two participants: a user and a node.

### Asset

A representation of value within the protocol, identified by a human-readable symbol and decimal precision. Assets are identified independently of any specific blockchain; the same logical asset MAY exist on multiple chains with different token addresses.

## State Concepts

### State

An abstract data structure representing the current configuration of a protocol entity at a specific version.

### State Version

A monotonically increasing integer that identifies the order of state updates. During off-chain advancement, each new state MUST have a version exactly one greater than the previous state.

### State Advancement

The process of updating a protocol entity's state off-chain through signed transitions exchanged between participants.

### State Enforcement

The process of submitting a signed state to the blockchain layer for on-chain validation and enforcement.

### Transition

A typed operation that describes the reason and parameters for a state update. Each transition carries a type, transaction identifier, account identifier, and amount.

### Intent

A value derived from the transition type that determines how the blockchain layer processes an enforced state. Intents include OPERATE, CLOSE, DEPOSIT, WITHDRAW, and various escrow and migration intents.

## Cryptographic Concepts

### Signature

A cryptographic proof that a specific key holder authorized a specific message. The protocol uses ECDSA over secp256k1.

### Signer

An entity capable of producing signatures. Each signer is associated with a specific key.

### Session Key

A delegated signing key authorized by a participant's primary key to sign specific types of state updates on their behalf. Session key authorization MUST be associated with the same address as the channel's user or node participant.

### Signature Validation Mode

A mechanism that determines how a signature is verified. The protocol currently defines two modes: default (0x00) for standard ECDSA validation and session key (0x01) for delegated validation.

## Ledger Concepts

### Ledger

A record of asset allocations within a channel, associated with a specific blockchain. Each ledger tracks user and node allocations and net flows, and MUST satisfy the invariant that allocations equal net flows.

### Home Ledger

The primary ledger of a channel state, associated with the blockchain where the state is enforced. The home ledger is the authoritative source for channel state enforcement.

### Non-Home Ledger

A secondary ledger tracking asset allocations on a blockchain other than the home chain. Used for cross-chain escrow operations and migrations.

### Home Chain

The blockchain identified by the home ledger's chain identifier. The home chain determines where enforcement operations are executed. It MAY change through a migration operation.

### Locked Funds

The total assets held by the blockchain enforcement contract on behalf of a specific channel. Unless the channel is being closed, the sum of UserAllocation and NodeAllocation MUST equal the locked funds.

### Vault

A pool of available funds maintained by the node on a specific blockchain, separate from any specific channel. The vault is used to cover required fund locking when a transition requires the node to lock additional assets into a channel.

### WAD Normalization

The process of scaling chain-specific asset amounts to the asset's configured decimal precision for exact, lossless cross-chain comparisons:

```
NormalizedAmount = Amount * 10^(18 - ChainDecimals)
```

Each unified asset defines a canonical decimal precision (e.g. 6 for USDC) that is used during User <> Clearnode interactions (e.g. on-chain deposit, on-chain state submission requests, transfers, app session operations etc.). The maximum supported decimal precision is 18.

## State Signing Categories

### Mutually Signed State

A state that carries valid signatures from both the user and the node. Only mutually signed states are enforceable on-chain.

### Node-Issued Pending State

A state produced by the node that carries only the node's signature. A pending state is NOT enforceable on-chain and becomes mutually signed only after the user acknowledges it.

### Channel Status

A specific on-chain channel data configuration, which changes throughout channel lifecycle, and includes *operating*, *disputed*, *migrating-in*, *migrated out*, etc. This can be thought of as a Finite State-Machine State (do not confuse with State Channel State).

### Escrow Channel Identifier

A 32-byte hash derived deterministically from the home channel identifier and the state version. Used to uniquely identify each escrow operation.

## Protocol Operations

### Checkpoint

The operation of submitting a signed state to the blockchain layer. A checkpoint records the latest agreed state on-chain.

### Challenge

An on-chain operation where a participant disputes the current enforced state by submitting a signed state along with a challenger signature. Initiates the challenge duration, during which other participants MAY respond with a higher-version state.

### Commit

The operation of moving assets from a channel into an extension, such as an application session. Decreases the user's allocation and the node's net flow.

### Release

The operation of returning assets from an extension back to the channel. Increases the user's allocation and the node's net flow.

### Escrow

A two-phase mechanism for cross-chain operations. An "escrow initiate" locks funds, and an "escrow finalize" releases them upon cooperative completion or after a timeout period.

## Extension Concepts

### Extension

An additional protocol module that provides functionality beyond the core channel protocol. Extensions interact with channels through commit and release transitions.

### Application Session

An extension that enables off-chain application functionality. Application sessions hold committed assets and maintain their own state.

### Application State

The state associated with an application session, tracking committed assets and application-specific data.

---

Previous: [Overview](overview.md) | Next: [Cryptography](cryptography.md)
