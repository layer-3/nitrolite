# Core Package

The `core` package defines the fundamental domain models, interfaces, and cryptographic utilities for the Clearnode protocol. It serves as the single source of truth for shared data structures between the node, client, and smart contract interactions.

## Overview

### Key Features

* **Domain Models**: Standardized structures for Channels, States, Transitions, and Ledgers.
* **On-chain Events**: Type-safe definitions for Home and Escrow channel lifecycle events.
* **Cryptographic Utilities**: Deterministic ID generation for channels, states, and transactions using Keccak256 and ABI packing.
* **Validation**: Interface and implementation for state transition logic (e.g., version increments, epoch tracking).
* **Abstractions**: Clean interfaces for Blockchain Clients and Event Listeners.

## Core Components

### Channel Lifecycle

The protocol distinguishes between two types of channels:

1. **Home Channel**: The primary settlement layer between a user and a node.
2. **Escrow Channel**: Temporary channels used for cross-chain or specific deposit/withdrawal operations.

### State Management

The `State` struct represents a snapshot of the off-chain ledger. Every state update must include:

* A version increment.
* At least one `Transition` (Transfer, Deposit, Withdrawal, etc.).
* Valid signatures from both the User and the Node.

### Identification & Hashing

The package provides deterministic hashing utilities to ensure consistency between off-chain logic and on-chain Smart Contracts:

| Method | Description |
| --- | --- |
| `GetHomeChannelID` | Hashes node, user, token, nonce, and challenge period. |
| `GetEscrowChannelID` | Derives an ID from a Home Channel ID and a state version. |
| `GetStateID` | Generates a unique hash for a specific state snapshot. |
| `GetTransactionID` | Creates a unique reference for individual transfers or adjustments. |

## Interface Definitions

### Client Interface

The `Client` interface abstracts the communication with the `ChannelsHub` smart contract:

* **Vault Operations**: `Deposit`, `Withdraw`, and `GetAccountsBalances`.
* **Channel Operations**: `Create`, `Checkpoint`, `Challenge`, and `Close`.
* **Escrow Operations**: Initiation and Finalization of Escrow deposits and withdrawals.

### Listener Interface

The `Listener` allows applications to react to on-chain state changes by registering handlers for events like `HomeChannelCreatedEvent` or `EscrowDepositFinalizedEvent`.

### State Advancer

The `StateAdvancer` ensures that off-chain state updates follow the protocol rules.

```go
advancer := core.NewStateAdvancerV1()
err := advancer.ValidateAdvancement(oldState, newState)
// Checks: Version increment, ledgers, epoch consistency, signature presence, etc.

```

## Data Structures

### Transaction Types

Transactions are categorized to handle specific ledger movements:

* **Home/Escrow**: Deposit, Withdrawal and Migration.
* **Operations**: Transfers, Commits, Releases.
* **Locking**: Escrow and Mutual locks for cross-chain safety.

### Ledger

The `Ledger` tracks balances and "Net Flow" (total funds inflow(+)/outflow(-) of the channel) for both the user and the node within a specific channel context.

## Usage Example: Generating IDs

```go
import "github.com/layer-3/nitrolite/core"

// Generate a Home Channel ID
channelID, err := core.GetHomeChannelID(
    nodeAddr, 
    userAddr, 
    tokenAddr, 
    nonce, 
    7*24*3600, // challenge
)

// Generate a State ID for a new update
stateID, err := core.GetStateID(userWallet, "eth", epoch, version)

```
