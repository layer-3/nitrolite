# Nitrolite V1 Clearnode Specifications

This directory contains Clearnode architecture, models and communication flows that facilitate communication between user, SDK client, Node and Blockchains — the core off-chain engine for the Nitrolite V1 Protocol.

## Contents

- **[api.yaml](api.yaml)** - API definitions including types, state transitions, and RPC methods
- **[data_models.mmd](data_models.mmd)** - Data model diagrams

### Communication Flows

- **[home_chan_creation_from_scratch.mmd](communication_flows/home_chan_creation_from_scratch.mmd)** - Home channel creation with initial deposit
- **[home_chan_deposit.mmd](communication_flows/home_chan_deposit.mmd)** - Home channel deposit (existing channel)
- **[home_chan_withdraw.mmd](communication_flows/home_chan_withdraw.mmd)** - Home channel withdrawal (existing channel)
- **[home_chan_withdraw_on_create_from_state.mmd](communication_flows/home_chan_withdraw_on_create_from_state.mmd)** - Channel creation with withdrawal from pending state
- **[transfer.mmd](communication_flows/transfer.mmd)** - Off-chain transfer (sender + automatic receiver state creation)
- **[app_session_deposit.mmd](communication_flows/app_session_deposit.mmd)** - Application session deposit with quorum verification
- **[escrow_chan_deposit.mmd](communication_flows/escrow_chan_deposit.mmd)** - Cross-chain escrow deposit (mutual lock → on-chain → finalize)
- **[escrow_chan_withdrawal.mmd](communication_flows/escrow_chan_withdrawal.mmd)** - Cross-chain escrow withdrawal (escrow lock → on-chain → finalize)

#### Remaining Flows

The following communication flows are not yet documented:

- **home chain migration** - Cross-chain state migration between home channels
- **app session create / operate / withdraw / close** - Full app session lifecycle beyond deposits

---

## Project Structure

```text
cerebro/                    # Cerebro Testing Client
clearnode/
    action_gateway/         # Rate limiting via gated actions
    api/
        app_session_v1/     # App session endpoints (create, deposit, operate, withdraw, close)
        apps_v1/            # Application registry endpoints
        channel_v1/         # Channel endpoints (create, submit_state, get_state, transfer)
        node_v1/            # Node info endpoints
        user_v1/            # User endpoints (balances, staking)
    config/
        migrations/
            postgres/       # Goose SQL migrations (embedded at compile time)
    event_handlers/         # Blockchain event processing (channel events, locking events)
    metrics/                # Prometheus metrics + lifespan metric aggregation
    store/
        database/           # GORM-based DB store
        memory/             # In-memory store for assets, blockchains, config
    blockchain_worker.go    # Processes pending BlockchainAction records
    runtime.go              # Embeds migrations, initializes services
    main.go                 # Entry point, EVM listeners, metric exporters
contracts/                  # Smart contracts (ChannelHub, Locking, etc.)
docs/                       # This directory
pkg/
    app/                    # App session types (AppSessionStatus, quorum, allocations)
    blockchain/
        evm/                # EVM client implementations
    core/                   # Core types: Channel, State, Transaction, Signer, Transition
    log/                    # Structured logging
    rpc/                    # RPC protocol: messages, requests, responses, errors
    sign/                   # Signer implementations (EthereumMsgSigner, EthereumRawSigner)
sdk/
    go/                     # Go SDK client
    ts/                     # TypeScript SDK client
    ts-compat/              # TypeScript compatibility SDK
test/                       # Integration test scenarios
```
