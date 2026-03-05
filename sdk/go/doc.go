// Package sdk provides the official Go client for the Nitrolite Clearnode API.
//
// The SDK offers a unified interface for interacting with Clearnode payment channels,
// supporting both high-level state operations and low-level RPC access. It simplifies
// the process of managing channel states, performing off-chain transactions, and
// settling on-chain when necessary.
//
// # Key Features
//
//   - Unified Client: A single `Client` struct for all operations.
//   - State Operations: High-level methods (`Deposit`, `Withdraw`, `Transfer`, `CloseHomeChannel`, `Acknowledge`)
//     to build and co-sign channel states off-chain.
//   - Blockchain Settlement: A single `Checkpoint` method to settle the latest state on-chain
//     (creating channels, depositing, withdrawing, or finalizing).
//   - Low-Level Access: Direct access to Clearnode RPC methods for advanced use cases
//     (e.g., querying node config, balances, channel info).
//   - App Sessions: Comprehensive support for creating and managing application sessions.
//   - Session Keys: Support for registering and using session keys for delegated signing.
//
// # Usage
//
// To use the SDK, create a `Client` instance with your Clearnode WebSocket URL and signers.
// You can configure blockchain RPCs for on-chain operations.
//
//	package main
//
//	import (
//		"context"
//		"log"
//		"time"
//
//		"github.com/layer-3/nitrolite/pkg/sign"
//		sdk "github.com/layer-3/nitrolite/sdk/go"
//		"github.com/shopspring/decimal"
//	)
//
//	func main() {
//		// Initialize signers
//		privateKeyHex := "YOUR_PRIVATE_KEY_HEX"
//		stateSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)
//		if err != nil {
//			log.Fatal(err)
//		}
//		txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Create Client
//		client, err := sdk.NewClient(
//			"wss://clearnode.example.com/ws",
//			stateSigner,
//			txSigner,
//			sdk.WithBlockchainRPC(80002, "https://rpc-endpoint.example.com"),
//			sdk.WithHandshakeTimeout(10*time.Second),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer client.Close()
//
//		ctx := context.Background()
//
//		// 1. Deposit (Off-chain state preparation)
//		// This creates a new state reflecting the deposit intent.
//		state, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("Deposit prepared. New state version: %d", state.Version)
//
//		// 2. Settlement (On-chain execution)
//		// This submits the transaction to the blockchain to create the channel or deposit funds.
//		txHash, err := client.Checkpoint(ctx, "usdc")
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("On-chain transaction hash: %s", txHash)
//
//		// 3. Transfer (Off-chain transaction)
//		// Transfers are instant and don't require immediate on-chain settlement.
//		_, err = client.Transfer(ctx, "0xRecipientAddress...", "usdc", decimal.NewFromInt(50))
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println("Transfer completed off-chain")
//	}
//
// # Client Configuration
//
// The `NewClient` function accepts variadic `ClientOption` functions to customize behavior:
//
//   - `WithBlockchainRPC(chainID, url)`: Registers an RPC endpoint for on-chain settlement.
//   - `WithHandshakeTimeout(duration)`: Sets the timeout for the initial WebSocket handshake.
//   - `WithPingInterval(duration)`: Sets the interval for WebSocket ping/pong keepalives.
//   - `WithErrorHandler(func(error))`: Sets a callback for handling background connection errors.
//
// # Error Handling
//
// The SDK methods return standard Go errors. Common errors to check for include connection issues,
// insufficient balances, or invalid state transitions. Errors from RPC calls often contain
// detailed messages from the Clearnode server.
package sdk
