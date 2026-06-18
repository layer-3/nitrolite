# Nitronode

Nitronode (formerly Clearnode) is the off-chain node implementation for the Nitrolite protocol. It manages state channels, processes off-chain transactions, and coordinates state updates between users and applications to enable fast, low-cost payment channels and complex application sessions.

## Overview

Nitronode provides a WebSocket-based RPC service that allows users and applications to:
- Create and manage home and escrow channels on multiple blockchains
- Perform instant off-chain transfers between users
- Execute multi-party application sessions with arbitrary logic
- Delegate signing authority via session keys
- Track balances and transaction history across all supported assets

The node monitors blockchain events, validates state transitions using a monotonic sequence logic, and ensures secure coordination between on-chain channels and off-chain state updates.

## Architecture

Nitronode is built with a modular architecture:

- **RPC Server**: WebSocket-based JSON-RPC server handling client requests.
- **Blockchain Listeners**: Monitors on-chain events from Nitrolite `ChannelHub` contracts across multiple chains.
- **Confirmation Gate**: Per-chain reorg-protection buffer between the listener and event handlers. Delays event delivery by `confirmation_delay_secs` so that events whose blocks are reorged out before the window elapses are dropped instead of committed. See [docs/reorg-fix.md](docs/reorg-fix.md).
- **Event Handlers**: Processes blockchain events to update internal channel and user states.
- **Storage Layer**:
  - **Database Store**: Persistent storage for channels, states, and transactions (supports SQLite and PostgreSQL).
  - **Memory Store**: Fast in-memory access for node configuration, blockchains, and assets.
- **Blockchain Workers**: Coordinates on-chain operations such as automated settlement or rebalancing.
- **Metrics**: Built-in Prometheus exporter for monitoring node health and protocol performance.

### API Groups

The WebSocket RPC service exposes several API groups:

1. **channel_v1**: Core payment channel management (Creation, State Submission, Latest State).
2. **app_session_v1**: Advanced application session management (Creation, Deposits).
3. **user_v1**: User-specific queries (Balances, Transaction History).
4. **node_v1**: Node-level information (Config, Supported Assets).

For detailed API specifications, see [../docs/api.yaml](../docs/api.yaml).

### Event handler version monotonicity

Channel-lifecycle events from the blockchain listener are applied with per-channel version monotonicity. If an event whose `StateVersion` is lower than the row's current `StateVersion` arrives (possible under contract reentrancy, indexer mis-order, or reorg replay), the handler logs a structured warning and drops the event without mutating the row. Implementation lives in [`event_handlers/service.go`](event_handlers/service.go).

For home-channel events (`ChannelChallenged`, `ChannelCheckpointed`, `ChannelClosed`), a dropped event additionally triggers a chain-state refresh: the Node fetches the authoritative on-chain channel state via `getChannelData` on the bound `ChannelHub` contract and overwrites the row's `Status`, `StateVersion`, and `ChallengeExpiresAt`. The refresher implementation is [`pkg/blockchain/evm/chain_state_refresher.go`](../pkg/blockchain/evm/chain_state_refresher.go), bound through the [`core.ChainStateRefresher`](../pkg/core/interface.go) interface.

The refresh runs inside the event-processing transaction. On RPC failure the transaction rolls back and the listener replays the event, so convergence is never silently lost. Escrow event handlers enforce the guard without the refresh hook — cross-chain RPC plumbing for escrow refresh is a deferred follow-up item. Pending its arrival, escrow rows can remain divergent from chain across an interim window until the next on-chain event arrives.

Operators may see `"event state version is less than current channel state version, ignoring"` warn logs during channel-lifecycle events; these indicate the defense-in-depth path fired and the Node converged with chain.

## Configuration

Nitronode uses YAML files for core configuration and environment variables for sensitive data and runtime overrides.

### Blockchain Configuration

Define supported chains in `config/blockchains.yaml`:

```yaml
default_contract_address: "0x019B65A265EB3363822f2752141b3dF16131b262"

blockchains:
  - name: polygon_amoy
    id: 80002
    contract_address: "0x9d1E88627884e066B81A02d69BCB2437a520534C"
    block_step: 1000
    confirmation_delay_secs: 10   # reorg-protection window; 0 disables. See docs/reorg-fix.md.

  - name: base_sepolia
    id: 84532
    contract_address: "0x33e57a8900882B8D5A038eC3Aa844c19Acfc539A"
```

### Asset Configuration

Define supported assets and their multi-chain token deployments in `config/assets.yaml`:

> **Warning:** all tokens grouped under one `symbol` are treated as fully fungible 1:1 representations of the same asset — off-chain credit denominated in that asset can be redeemed from any of these token inventories. Group only economically equivalent (1:1 redeemable) tokens under one symbol; mixing non-equivalent tokens (e.g. a test token and production USDC) lets credit sourced from the cheap inventory be redeemed against the valuable one. Equivalence cannot be verified programmatically and is an operator responsibility. See [Asset-symbol equivalence](../docs/protocol/security-and-limitations.md).

```yaml
assets:
  - symbol: "USDC"
    name: "USD Coin"
    decimals: 6
    suggested_blockchain_id: 80002
    tokens:
      - blockchain_id: 80002
        address: "0xDB9F293e3898c9E5536A3be1b0C56c89d2b32DEb"
        decimals: 6
      - blockchain_id: 84532
        address: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831"
        decimals: 6
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NITRONODE_SIGNER_KEY` | Private key for signing node state updates | (Required) |
| `NITRONODE_DATABASE_DRIVER` | `sqlite` or `postgres` | `sqlite` |
| `NITRONODE_DATABASE_URL` | Postgres DSN/URL or sqlite file path. When set for `postgres`, used verbatim and overrides the individual host/user/password/sslmode fields | `nitronode.db` |
| `NITRONODE_DATABASE_SSLMODE` | Postgres SSL mode: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` | `require` |
| `NITRONODE_LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `NITRONODE_BLOCKCHAIN_RPC_<NAME>` | RPC endpoint for a specific blockchain | (Required) |

## Running Nitronode

### Prerequisites

- Go 1.25 or later
- Access to blockchain RPC endpoints (e.g., Alchemy, Infura, or local Anvil)

### Quick Start (Local)

1. Set up your configuration files in a `./config` directory.
2. Set the required environment variables:
   ```bash
   export NITRONODE_SIGNER_KEY=0x...
   export NITRONODE_BLOCKCHAIN_RPC_POLYGON_AMOY=https://...
   ```
3. Run the node:
   ```bash
   go run . --config-dir ./config
   ```

The node will be available at `ws://localhost:7824/ws`.

### Docker

```bash
docker build -t nitronode .
docker run -p 7824:7824 -e NITRONODE_SIGNER_KEY=... nitronode
```

## Development

### Project Structure

```
nitronode/
├── api/             # JSON-RPC request handlers
├── config/          # Default configurations and migrations
├── docs/            # Component design notes (e.g. reorg-fix.md)
├── event_handlers/  # Logic for reacting to blockchain events
├── metrics/         # Prometheus telemetry implementation
├── store/           # Persistence layer (SQL and Memory)
├── main.go          # Entry point
└── runtime.go       # System initialization logic
```

### Testing

```bash
# Run all tests (requires GOCACHE redirection if in restricted environment)
export GOCACHE=/tmp/gocache && go test -v ./...
```

## Protocol Features Not Yet Active

The following protocol operations are fully specified in [protocol-description.md](../protocol-description.md) and implemented in the `ChannelHub` smart contract, but are **not yet active in the Nitronode off-chain implementation**. Submitting these transition types via `channels.v1.submit_state` returns an error.

| Feature | Transition types | Status |
|---------|-----------------|--------|
| Home chain migration | `migrate` | Off-chain flow not implemented. On-chain `initiateMigration()` / `finalizeMigration()` are functional. Event handlers for `MigrationInInitiated`, `MigrationOutFinalized`, etc. are stubs. |
| Cross-chain escrow deposit | `mutual_lock`, `escrow_lock`, `escrow_deposit` | Off-chain flow implemented, but not fully tested. |
| Cross-chain escrow withdrawal | `escrow_withdraw` | Off-chain flow implemented, but not fully tested. |

## Documentation

- [Nitrolite Protocol Overview](../protocol-description.md)
- [Communication Flows](../docs/communication_flows/)
- [API Reference](../docs/api.yaml)
- [Reorg-Protection Confirmation Gate](docs/reorg-fix.md)

## License

Part of the Nitrolite project. Licensed under the MIT License.
