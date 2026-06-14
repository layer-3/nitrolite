# Core Package

The `core` package defines the fundamental domain models, interfaces, and cryptographic utilities for the Nitronode protocol. It serves as the single source of truth for shared data structures between the node, client, and smart contract interactions.

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

The `Listener` exposes events via a **two-handler model**. A `liveHandler` receives live events plus any historical events still within the reorg window, while a `historicalEventHandler` receives mature historical events past the configured `confirmationDelay`. Per-event routing is decided by the listener itself: it compares `eventLog.BlockTimestamp` against `confirmationDelay` to choose which handler an event flows into. This makes the listener delay-aware rather than pushing that decision down to consumers.

The typical `liveHandler` is the **`ConfirmationGate`**, which implements the reorg-protection window. The gate buffers each event for `confirmation_delay_secs` before forwarding it to the reactor; if the event's block is reorged out within that window, the gate silently drops it instead of committing it downstream. With the gate in place, the reactor only ever sees events whose blocks have survived the configured confirmation window.

To make this work, the listener owns timestamp population. **`ensureBlockTimestamp`** guarantees `BlockTimestamp` is set on every non-removed event before it is forwarded: it uses `eventLog.BlockTimestamp` directly when present, and otherwise falls back to a cached `HeaderByHash` lookup. The gate relies on this to compute each event's `arrivedAt` correctly. **`Removed: true`** logs are handled exclusively at the listener boundary: in the live (Phase 2) path with a gate, removed logs are forwarded so the gate can cancel a pending timer; with no gate configured (`confirmation_delay_secs == 0`), the listener drops removed logs at Phase 2 and the reactor never sees them. Historical (Phase 1) replays use `eth_getLogs`, which never emits removals, so that path is simpler by construction.

On startup, the listener reconciles against possible reorgs that happened while the node was down. **`findCommonAncestor`** walks stored block hashes backward to locate a still-canonical resume point. If every stored block has been reorged out, it returns the orphaned-latest height so `eth_getLogs` re-fetches canonical replacements from that range; the orphan hash itself is discarded — only the height matters because `eth_getLogs` is a canonical-chain range query.

See [`nitronode/docs/reorg-fix.md`](../../nitronode/docs/reorg-fix.md) for the full design.

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
