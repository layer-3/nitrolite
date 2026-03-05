# Blockchain-Agnostic Signing Library

A blockchain-agnostic library for cryptographic operations.

## Core Design

This library separates generic interfaces from specific blockchain implementations.

## Features

* **Blockchain-Agnostic Interfaces**: Defines a standard set of interfaces for cryptographic operations.
* **EVM Implementation**: Includes a ready-to-use implementation for Ethereum and other EVM-compatible chains.
* **Easily Extensible**: Simple to add support for new blockchains like Solana, Bitcoin, etc.
* **Type-Safe**: Provides distinct types for `Address`, `Signature`, `PublicKey`, and `PrivateKey`.

## Usage

See the Go package documentation and examples by running:
```bash
go doc -all github.com/layer-3/nitrolite/pkg/sign