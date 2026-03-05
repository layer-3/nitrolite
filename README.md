[![Release Policy](https://img.shields.io/badge/release%20policy-v1.0-blue)](https://github.com/layer-3/release-process/blob/master/README.md)
[![codecov](https://codecov.io/github/layer-3/nitrolite/graph/badge.svg?token=XASM4CIEFO)](https://codecov.io/github/layer-3/nitrolite)
[![Go Reference](https://pkg.go.dev/badge/github.com/layer-3/nitrolite.svg)](https://pkg.go.dev/github.com/layer-3/nitrolite)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/layer-3/nitrolite)](https://github.com/layer-3/nitrolite/releases)


# Nitrolite: State Channel Framework

Nitrolite is a lightweight, efficient state channel framework for Ethereum and other EVM-compatible blockchains, enabling off-chain interactions while maintaining on-chain security guarantees.

## Overview

Nitrolite is a complete state channel infrastructure consisting of several main components:

1.  **Smart Contracts**: On-chain infrastructure for state channel management (ChannelHub, ChannelEngine).
2.  **Clearnode**: A broker providing ledger services and off-chain state management.
3.  **Go SDK**: Comprehensive Go library for building backend and CLI applications.
4.  **TypeScript SDK**: Client-side library for building web and mobile applications.
5.  **Cerebro (CLI)**: Interactive command-line interface for managing channels and assets.

### Key Benefits

- **Instant Finality**: Transactions settle immediately between parties off-chain.
- **Reduced Gas Costs**: Most interactions happen off-chain, with minimal on-chain footprint.
- **High Throughput**: Support for thousands of transactions per second.
- **Security Guarantees**: Same security as on-chain, backed by cryptographic proofs and challenge periods.
- **Chain Agnostic**: Works with any EVM-compatible blockchain.

## Project Structure

This repository contains:

- **[`/contracts`](/contracts)**: Solidity smart contracts for the state channel framework.
- **[`/clearnode`](/clearnode)**: Implementation of the Clearnode service.
- **[`/sdk/go`](/sdk/go)**: Official Go SDK.
- **[`/sdk/ts`](/sdk/ts)**: Official TypeScript SDK.
- **[`/cerebro`](/cerebro)**: Clearnode CLI tool.
- **[`/pkg`](/pkg)**: Shared Go packages (core state machine, signing, etc.).
- **[`/docs`](/docs)**: Protocol specifications and architectural documentation.

## Protocol

Nitrolite implements a state channel protocol that enables secure off-chain communication with minimal on-chain operations. The protocol includes:

- **Channel Creation**: A funding protocol where participants lock assets in the `ChannelHub`.
- **Off-Chain Updates**: A mechanism for exchanging and signing versioned state updates off-chain.
- **Channel Closure**: Multiple resolution paths including cooperative close and unilateral challenge-response.
- **Checkpointing**: The ability to record valid states on-chain without closing the channel.
- **App Sessions**: Support for multi-participant application sessions with arbitrary logic.
- **Session Keys**: Delegated signing authority for automated or restricted operations.

See the [protocol description](/protocol-description.md) for complete details.

## Smart Contracts

The Nitrolite contract system (built with Foundry) provides:

- **ChannelHub**: The central entry point for all channel operations.
- **ChannelEngine**: Logic for state verification and transition validation.
- **Escrow Engines**: Specialized logic for cross-chain deposits and withdrawals.
- **SigValidators**: Pluggable signature validation modules (e.g., Session Keys).

### Deployments

For the most up-to-date contract addresses on all supported networks, see the [contract deployments directory](/contracts/deployments/).

See the [contract README](/contracts/README.md) for detailed contract documentation.

## Clearnode

Clearnode is a message broker and state manager that enables efficient off-chain payment channels.

### Features

- **Multi-Chain Support**: Connect to multiple EVM blockchains (Polygon, Celo, Base, etc.).
- **Off-Chain Payments**: High-throughput transfers with near-zero latency.
- **App Sessions**: Manage complex multi-party application states.
- **Flexible Storage**: Support for SQLite (embedded) and PostgreSQL.
- **Quorum-Based Signatures**: Weight-based quorum requirements for state updates.

See the [Clearnode Documentation](/clearnode/README.md) for more details.

## SDKs

### Go SDK

The official Go SDK for building performant backend integrations or CLI tools.

```bash
go get github.com/layer-3/nitrolite/sdk/go
```
See [SDK Go README](/sdk/go/README.md).

### TypeScript SDK

The official TypeScript SDK for web-based applications.

```bash
npm install @layer-3/nitrolite
```
See [SDK TS README](/sdk/ts/README.md).

## Cerebro (CLI)

Interactive CLI for interacting with Clearnode.

```bash
# From the root directory
go build -o nitrolite-cli ./cerebro
./nitrolite-cli wss://node.example.com/ws
```
See [Cerebro README](/cerebro/README.md).

## Quick Start with Docker Compose

Get started quickly with the local development environment:

```bash
# Start the environment
docker-compose up -d

# This will:
# 1. Start a local Anvil blockchain on port 8545
# 2. Deploy core Nitrolite contracts
# 3. Seed test tokens and configuration
# 4. Start the Clearnode service.
```

## Development

```bash
# Run contract tests
cd contracts && forge test

# Run Go tests (requires GOCACHE redirection if restricted)
export GOCACHE=/tmp/gocache && go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
