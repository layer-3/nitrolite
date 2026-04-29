# Clearnode

Clearnode is the off-chain node implementation for the Nitrolite protocol. It manages state channels, processes off-chain transactions, and coordinates state updates between users and applications to enable fast, low-cost payment channels and complex application sessions.

## Overview

Clearnode provides a WebSocket-based RPC service that allows users and applications to:
- Create and manage home and escrow channels on multiple blockchains
- Perform instant off-chain transfers between users
- Execute multi-party application sessions with arbitrary logic
- Delegate signing authority via session keys
- Atomically rebalance funds across multiple application sessions
- Track balances and transaction history across all supported assets

The node monitors blockchain events, validates state transitions using a monotonic sequence logic, and ensures secure coordination between on-chain channels and off-chain state updates.

## Architecture

Clearnode is built with a modular architecture:

- **RPC Server**: WebSocket-based JSON-RPC server handling client requests.
- **Blockchain Listeners**: Monitors on-chain events from Nitrolite `ChannelHub` contracts across multiple chains.
- **Event Handlers**: Processes blockchain events to update internal channel and user states.
- **Storage Layer**:
  - **Database Store**: Persistent storage for channels, states, and transactions (supports SQLite and PostgreSQL).
  - **Memory Store**: Fast in-memory access for node configuration, blockchains, and assets.
- **Blockchain Workers**: Coordinates on-chain operations such as automated settlement or rebalancing.
- **Metrics**: Built-in Prometheus exporter for monitoring node health and protocol performance.

### API Groups

The WebSocket RPC service exposes several API groups:

1. **channel_v1**: Core payment channel management (Creation, State Submission, Latest State).
2. **app_session_v1**: Advanced application session management (Creation, Deposits, Rebalancing).
3. **user_v1**: User-specific queries (Balances, Transaction History).
4. **node_v1**: Node-level information (Config, Supported Assets).

For detailed API specifications, see [../docs/api.yaml](../docs/api.yaml).

## Configuration

Clearnode uses YAML files for core configuration and environment variables for sensitive data and runtime overrides.

### Blockchain Configuration

Define supported chains in `config/blockchains.yaml`:

```yaml
default_contract_address: "0x019B65A265EB3363822f2752141b3dF16131b262"

blockchains:
  - name: polygon_amoy
    id: 80002
    contract_address: "0x9d1E88627884e066B81A02d69BCB2437a520534C"
    block_step: 1000

  - name: base_sepolia
    id: 84532
    contract_address: "0x33e57a8900882B8D5A038eC3Aa844c19Acfc539A"
```

### Asset Configuration

Define supported assets and their multi-chain token deployments in `config/assets.yaml`:

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
| `CLEARNODE_SIGNER_KEY` | Private key for signing node state updates | (Required) |
| `CLEARNODE_DATABASE_DRIVER` | `sqlite` or `postgres` | `sqlite` |
| `CLEARNODE_DATABASE_URL` | Connection string or file path | `clearnode.db` |
| `CLEARNODE_LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `CLEARNODE_BLOCKCHAIN_RPC_<NAME>` | RPC endpoint for a specific blockchain | (Required) |

## Running Clearnode

### Prerequisites

- Go 1.25 or later
- Access to blockchain RPC endpoints (e.g., Alchemy, Infura, or local Anvil)

### Quick Start (Local)

1. Set up your configuration files in a `./config` directory.
2. Set the required environment variables:
   ```bash
   export CLEARNODE_SIGNER_KEY=0x...
   export CLEARNODE_BLOCKCHAIN_RPC_POLYGON_AMOY=https://...
   ```
3. Run the node:
   ```bash
   go run . --config-dir ./config
   ```

The node will be available at `ws://localhost:7824/ws`.

### Docker

```bash
docker build -t clearnode .
docker run -p 7824:7824 -e CLEARNODE_SIGNER_KEY=... clearnode
```

## Development

### Project Structure

```
clearnode/
├── api/             # JSON-RPC request handlers
├── config/          # Default configurations and migrations
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

## Documentation

- [Nitrolite Protocol Overview](../protocol-description.md)
- [Communication Flows](../docs/communication_flows/)
- [API Reference](../docs/api.yaml)

## License

Part of the Nitrolite project. Licensed under the MIT License.
