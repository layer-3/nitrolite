[![Go Reference](https://pkg.go.dev/badge/github.com/layer-3/nitrolite/sdk/go.svg)](https://pkg.go.dev/github.com/layer-3/nitrolite/sdk/go)

# Clearnode Go SDK

Go SDK for Clearnode payment channels providing both high-level and low-level operations in a unified client:
- **State Operations**: `Deposit`, `Withdraw`, `Transfer`, `CloseHomeChannel`, `Acknowledge` - build and co-sign states off-chain
- **Blockchain Settlement**: `Checkpoint` - the single entry point for all on-chain transactions
- **Low-Level Operations**: Direct RPC access for custom flows and advanced use cases

## Method Cheat Sheet

### State Operations (Off-Chain)
```go
client.Deposit(ctx, blockchainID, asset, amount)      // Prepare deposit state
client.Withdraw(ctx, blockchainID, asset, amount)     // Prepare withdrawal state
client.Transfer(ctx, recipientWallet, asset, amount)  // Prepare transfer state
client.CloseHomeChannel(ctx, asset)                   // Prepare finalize state
client.Acknowledge(ctx, asset)                        // Acknowledge received state
```

### Blockchain Settlement
```go
client.Checkpoint(ctx, asset)                         // Settle latest state on-chain
client.Challenge(ctx, state)                          // Submit on-chain challenge
client.ApproveToken(ctx, chainID, asset, amount)      // Approve ChannelHub to spend tokens
client.GetOnChainBalance(ctx, chainID, asset, wallet) // Query on-chain token balance
```

### Node Information
```go
client.Ping(ctx)                    // Health check
client.GetConfig(ctx)               // Node configuration
client.GetBlockchains(ctx)          // Supported blockchains
client.GetAssets(ctx, blockchainID) // Supported assets
```

### User Queries
```go
client.GetBalances(ctx, wallet)             // User balances
client.GetTransactions(ctx, wallet, opts)   // Transaction history
```

### Channel Queries
```go
client.GetHomeChannel(ctx, wallet, asset)       // Home channel info
client.GetEscrowChannel(ctx, escrowChannelID)   // Escrow channel info
client.GetLatestState(ctx, wallet, asset, onlySigned) // Latest state
```

### App Registry
```go
client.GetApps(ctx, opts)                              // List registered apps
client.RegisterApp(ctx, appID, metadata, approvalNotRequired) // Register new app
```

### App Sessions
```go
client.GetAppSessions(ctx, opts)                              // List sessions
client.GetAppDefinition(ctx, appSessionID)                    // Session definition
client.CreateAppSession(ctx, definition, sessionData, sigs)   // Create session
client.SubmitAppSessionDeposit(ctx, update, sigs, asset, amount) // Deposit to session
client.SubmitAppState(ctx, update, sigs)                      // Update session
client.RebalanceAppSessions(ctx, signedUpdates)               // Atomic rebalance
```

### Session Keys — App Sessions
```go
client.SignSessionKeyState(state)                                   // Sign an app session key state
client.SubmitAppSessionKeyState(ctx, state)                         // Register/update app session key
client.GetLastAppKeyStates(ctx, userAddress, opts)                  // Get active app session key states
```

### Session Keys — Channels
```go
client.SignChannelSessionKeyState(state)                            // Sign a channel session key state
client.SubmitChannelSessionKeyState(ctx, state)                     // Register/update channel session key
client.GetLastChannelKeyStates(ctx, userAddress, opts)              // Get active channel session key states
```

### Shared Utilities
```go
client.Close()                          // Close connection
client.WaitCh()                         // Connection monitor channel
client.SignState(state)                 // Sign a state (advanced)
client.GetUserAddress()                 // Get signer's address
client.SetHomeBlockchain(asset, chainID) // Set default blockchain for asset
```

## Quick Start

### Unified Client (High-Level + Low-Level)

```go
package main

import (
    "context"
    "github.com/layer-3/nitrolite/pkg/core"
    "github.com/layer-3/nitrolite/pkg/sign"
    sdk "github.com/layer-3/nitrolite/sdk/go"
    "github.com/shopspring/decimal"
)

func main() {
    // Create signers from private key
    msgSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
    stateSigner, _ := core.NewChannelDefaultSigner(msgSigner)
    txSigner, _ := sign.NewEthereumRawSigner(privateKeyHex)

    // Create unified client
    client, _ := sdk.NewClient(
        "wss://clearnode.example.com/ws",
        stateSigner,
        txSigner,
        sdk.WithBlockchainRPC(80002, "https://polygon-amoy.alchemy.com/v2/KEY"),
    )
    defer client.Close()

    ctx := context.Background()

    // Step 1: Build and co-sign states off-chain
    state, _ := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
    fmt.Printf("Deposit state version: %d\n", state.Version)

    // Step 2: Settle on-chain via Checkpoint
    txHash, _ := client.Checkpoint(ctx, "usdc")
    fmt.Printf("On-chain tx: %s\n", txHash)

    // Transfer (off-chain only, no Checkpoint needed for existing channels)
    state, _ = client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))

    // Low-level operations - same client
    config, _ := client.GetConfig(ctx)
    balances, _ := client.GetBalances(ctx, walletAddress)
}
```

## Architecture

```
sdk/go/
├── client.go         # Core client, constructors, high-level operations
├── node.go           # Node information methods
├── user.go           # User query methods
├── channel.go        # Channel and state management
├── app_registry.go   # App registry methods
├── app_session.go    # App session methods
├── asset_cache.go    # Asset lookup and caching
├── config.go         # Configuration options
├── doc.go            # Package documentation
└── utils.go          # Type conversions
```

## Client API

### Creating a Client

```go
// Step 1: Create signers from private key
msgSigner, err := sign.NewEthereumMsgSigner("0x1234...")
if err != nil {
    log.Fatal(err)
}

// Wrap with ChannelDefaultSigner (prepends 0x00 type byte)
stateSigner, err := core.NewChannelDefaultSigner(msgSigner)
if err != nil {
    log.Fatal(err)
}

txSigner, err := sign.NewEthereumRawSigner("0x1234...")
if err != nil {
    log.Fatal(err)
}

// Step 2: Create unified client
client, err := sdk.NewClient(
    wsURL,
    stateSigner,  // core.ChannelSigner for signing channel states
    txSigner,     // sign.Signer for signing blockchain transactions
    sdk.WithBlockchainRPC(chainID, rpcURL), // Required for Checkpoint
    sdk.WithHandshakeTimeout(10*time.Second),
    sdk.WithPingInterval(5*time.Second),
)

// Step 3: (Optional) Set home blockchain for assets
// Required for Transfer operations that may trigger channel creation
err = client.SetHomeBlockchain("usdc", 80002)
if err != nil {
    log.Fatal(err)
}
```

### Configuring Home Blockchain

#### `SetHomeBlockchain(asset, blockchainID) error`

Sets the default blockchain network for a specific asset. This is required for `Transfer()` operations that may trigger channel creation, as Transfer doesn't accept a blockchain ID parameter.

```go
err := client.SetHomeBlockchain("usdc", 80002)
if err != nil {
    log.Fatal(err)
}
```

**Important Notes:**
- This mapping is immutable once set for the client instance
- The asset must be supported on the specified blockchain
- Required before calling `Transfer()` on a new channel

### State Operations

All state operations build and co-sign a state off-chain. They return `(*core.State, error)`. Use `Checkpoint` to settle the state on-chain.

#### `Deposit(ctx, blockchainID, asset, amount) (*core.State, error)`

Prepares a deposit state. Creates a new channel if none exists, otherwise advances the existing state.

```go
state, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
txHash, err := client.Checkpoint(ctx, "usdc") // settle on-chain
```

**Requirements:**
- Sufficient token balance (checked on-chain during Checkpoint)

#### `Withdraw(ctx, blockchainID, asset, amount) (*core.State, error)`

Prepares a withdrawal state to remove funds from the channel.

```go
state, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
txHash, err := client.Checkpoint(ctx, "usdc") // settle on-chain
```

**Requirements:**
- Existing channel with sufficient balance

#### `Transfer(ctx, recipientWallet, asset, amount) (*core.State, error)`

Prepares an off-chain transfer to another wallet. For existing channels, no Checkpoint is needed.

```go
state, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
```

**Requirements:**
- Existing channel with sufficient balance OR
- Home blockchain configured via `SetHomeBlockchain()` (for new channels)

#### `CloseHomeChannel(ctx, asset) (*core.State, error)`

Prepares a finalize state to close the user's channel.

```go
state, err := client.CloseHomeChannel(ctx, "usdc")
txHash, err := client.Checkpoint(ctx, "usdc") // close on-chain
```

**Requirements:**
- Existing channel (user must have deposited first)

#### `Acknowledge(ctx, asset) (*core.State, error)`

Acknowledges a received state (e.g., after receiving a transfer).

```go
state, err := client.Acknowledge(ctx, "usdc")
```

**Requirements:**
- Home blockchain configured via `SetHomeBlockchain()` when no channel exists

### Blockchain Settlement

#### `Checkpoint(ctx, asset) (txHash, error)`

Settles the latest co-signed state on-chain. This is the single entry point for all blockchain transactions. Based on the transition type and on-chain channel status, it calls the appropriate blockchain method:

- **Channel not on-chain** (status Void): Creates the channel
- **Deposit/Withdrawal on existing channel**: Checkpoints the state
- **Finalize**: Closes the channel

```go
txHash, err := client.Checkpoint(ctx, "usdc")
```

**Requirements:**
- Blockchain RPC configured via `WithBlockchainRPC`
- A co-signed state must exist (call Deposit, Withdraw, etc. first)
- Sufficient gas for the blockchain transaction

#### `Challenge(ctx, state) (txHash, error)`

Submits an on-chain challenge for a channel using a co-signed state. A challenge initiates a dispute period on-chain. If the counterparty does not respond with a higher-versioned state before the challenge period expires, the channel can be closed with the challenged state.

```go
state, err := client.GetLatestState(ctx, wallet, "usdc", true)
txHash, err := client.Challenge(ctx, *state)
```

**Requirements:**
- Blockchain RPC configured via `WithBlockchainRPC`
- State must have both user and node signatures
- State must have a HomeChannelID

#### `ApproveToken(ctx, chainID, asset, amount) (txHash, error)`

Approves the ChannelHub contract to spend ERC-20 tokens on behalf of the user. This is required before depositing ERC-20 tokens. Native tokens (e.g., ETH) do not require approval.

```go
txHash, err := client.ApproveToken(ctx, 80002, "usdc", decimal.NewFromInt(1000))
```

**Requirements:**
- Blockchain RPC configured via `WithBlockchainRPC`

#### `GetOnChainBalance(ctx, chainID, asset, wallet) (decimal.Decimal, error)`

Queries the on-chain token balance (ERC-20 or native) for a wallet on a specific blockchain. The returned value is already adjusted for token decimals.

```go
balance, err := client.GetOnChainBalance(ctx, 80002, "usdc", "0x1234...")
fmt.Printf("On-chain balance: %s\n", balance)
```

**Requirements:**
- Blockchain RPC configured via `WithBlockchainRPC`

## Low-Level API

All low-level RPC methods are available on the same Client instance.

### Node Information

```go
err := client.Ping(ctx)
config, err := client.GetConfig(ctx)
blockchains, err := client.GetBlockchains(ctx)
assets, err := client.GetAssets(ctx, &blockchainID) // or nil for all
```

### User Data

```go
balances, err := client.GetBalances(ctx, wallet)
txs, meta, err := client.GetTransactions(ctx, wallet, opts)
```

### Channel Queries

```go
channel, err := client.GetHomeChannel(ctx, wallet, asset)
channel, err := client.GetEscrowChannel(ctx, escrowChannelID)
state, err := client.GetLatestState(ctx, wallet, asset, onlySigned)
```

**Note:** State submission and channel creation are handled internally by state operations (Deposit, Withdraw, Transfer). On-chain settlement is handled by Checkpoint.

### App Registry

```go
// List registered applications with optional filtering
apps, meta, err := client.GetApps(ctx, &sdk.GetAppsOptions{
    AppID:       &appID,
    OwnerWallet: &wallet,
})

// Register a new application
err := client.RegisterApp(ctx, "my-app", `{"name": "My App"}`, false)
```

### App Sessions (Low-Level)

```go
sessions, meta, err := client.GetAppSessions(ctx, opts)
def, err := client.GetAppDefinition(ctx, appSessionID)
sessionID, version, status, err := client.CreateAppSession(ctx, def, data, sigs)
nodeSig, err := client.SubmitAppSessionDeposit(ctx, update, sigs, asset, amount)
err := client.SubmitAppState(ctx, update, sigs)
batchID, err := client.RebalanceAppSessions(ctx, signedUpdates)
```

### Session Keys — App Sessions

```go
// Sign and submit an app session key state
state := app.AppSessionKeyStateV1{
    UserAddress:    client.GetUserAddress(),
    SessionKey:     "0xSessionKey...",
    Version:        1,
    ApplicationIDs: []string{"app1"},
    AppSessionIDs:  []string{},
    ExpiresAt:      time.Now().Add(24 * time.Hour),
}
sig, err := client.SignSessionKeyState(state)
state.UserSig = sig
err = client.SubmitAppSessionKeyState(ctx, state)

// Query active app session key states
states, err := client.GetLastAppKeyStates(ctx, userAddress, nil)
states, err := client.GetLastAppKeyStates(ctx, userAddress, &sdk.GetLastKeyStatesOptions{
    SessionKey: &sessionKeyAddr,
})
```

### Session Keys — Channels

```go
// Sign and submit a channel session key state
state := core.ChannelSessionKeyStateV1{
    UserAddress: client.GetUserAddress(),
    SessionKey:  "0xSessionKey...",
    Version:     1,
    Assets:      []string{"usdc", "weth"},
    ExpiresAt:   time.Now().Add(24 * time.Hour),
}
sig, err := client.SignChannelSessionKeyState(state)
state.UserSig = sig
err = client.SubmitChannelSessionKeyState(ctx, state)

// Query active channel session key states
states, err := client.GetLastChannelKeyStates(ctx, userAddress, nil)
states, err := client.GetLastChannelKeyStates(ctx, userAddress, &sdk.GetLastChannelKeyStatesOptions{
    SessionKey: &sessionKeyAddr,
})
```

## Key Concepts

### State Management

Payment channels use versioned states signed by both user and node. The SDK uses a two-step pattern:

```go
// Step 1: Build and co-sign state off-chain
state, _ := client.Deposit(...)   // Returns *core.State
state, _ = client.Withdraw(...)   // Returns *core.State
state, _ = client.Transfer(...)   // Returns *core.State

// Step 2: Settle on-chain (when needed)
txHash, _ := client.Checkpoint(ctx, "usdc")
```

**State Flow (Internal):**
1. Get latest state with `GetLatestState()`
2. Create next state with `state.NextState()`
3. Apply transition (deposit, withdraw, transfer, etc.)
4. Sign state with `SignState()`
5. Submit to node for co-signing
6. Return co-signed state

On-chain settlement is handled separately by `Checkpoint`.

### Signing

States are signed using ECDSA with EIP-155 via `pkg/sign`:

```go
// Create signers from private key
stateSigner, err := sign.NewEthereumMsgSigner(privateKeyHex)  // For channel states
txSigner, err := sign.NewEthereumRawSigner(privateKeyHex)     // For blockchain transactions

// Get address
address := txSigner.PublicKey().Address().String()
```

**Signing Process:**
1. State -> ABI Encode (via `core.PackState`)
2. Packed State -> Keccak256 Hash
3. Hash -> ECDSA Sign (via `signer.Sign`)
4. Result: 65-byte signature (R || S || V)

**Two Signer Types:**
- `EthereumMsgSigner`: Signs channel state updates (off-chain signatures)
- `EthereumRawSigner`: Signs blockchain transactions (on-chain operations)

### Channel Signers (`pkg/core`)

The SDK wraps raw signers with a `ChannelSigner` interface that prepends a type byte to every signature. This allows the on-chain contract to dispatch signature verification to the correct validator.

```go
// ChannelSigner interface (in pkg/core)
type ChannelSigner interface {
    sign.Signer
    Type() ChannelSignerType
}
```

**Two channel signer types:**

| Type | Byte | Struct | Usage |
|------|------|--------|-------|
| Default | `0x00` | `core.ChannelDefaultSigner` | Main wallet signs directly. Signature = `0x00 \|\| EIP-191 sig`. |
| Session Key | `0x01` | `core.ChannelSessionKeySignerV1` | Delegated session key signs on behalf of main wallet. Signature = `0x01 \|\| ABI-encoded auth + session key sig`. |

**Creating a channel signer:**

```go
// Default signer (wraps EthereumMsgSigner with 0x00 prefix)
msgSigner, _ := sign.NewEthereumMsgSigner(privateKeyHex)
channelSigner, _ := core.NewChannelDefaultSigner(msgSigner)

// Pass to NewClient as the stateSigner parameter
client, _ := sdk.NewClient(wsURL, channelSigner, txSigner, opts...)
```

The `NewClient` constructor expects a `core.ChannelSigner` for the `stateSigner` parameter. When using `sign.NewEthereumMsgSigner` directly, it must first be wrapped with `core.NewChannelDefaultSigner` (or `core.ChannelSessionKeySignerV1` for session key operation).

### Channel Lifecycle

1. **Void**: No channel exists
2. **Create**: Deposit creates channel on-chain
3. **Open**: Channel active, can deposit/withdraw/transfer
4. **Challenged**: Dispute initiated (advanced)
5. **Closed**: Channel finalized (advanced)

## When to Use State Operations vs Low-Level Operations

### Use State Operations When:
- Building user-facing applications
- Need simple deposit/withdraw/transfer
- Want automatic state management with two-step pattern
- Don't need custom flows

### Use Low-Level Operations When:
- Building infrastructure/tooling
- Implementing custom state transitions
- Need fine-grained control
- Working with app sessions directly

## Error Handling

All errors include context:

```go
state, err := client.Deposit(ctx, 80002, "usdc", amount)
if err != nil {
    log.Printf("State error: %v", err)
}

txHash, err := client.Checkpoint(ctx, "usdc")
if err != nil {
    // Error: "failed to create channel on blockchain: insufficient balance"
    log.Printf("Checkpoint error: %v", err)
}
```

Common errors:
- `"home blockchain not set for asset"` - Missing `SetHomeBlockchain` for new channel creation
- `"blockchain RPC not configured for chain"` - Missing `WithBlockchainRPC` (for Checkpoint)
- `"no channel exists for asset"` - Checkpoint called without a co-signed state
- `"insufficient balance"` - Not enough funds in channel/wallet
- `"failed to sign state"` - Invalid private key or state
- `"transition type ... does not require a blockchain operation"` - Checkpoint called on unsupported transition

## Configuration Options

```go
sdk.WithBlockchainRPC(chainID, rpcURL)    // Configure blockchain RPC (required for Checkpoint)
sdk.WithHandshakeTimeout(duration)         // Connection timeout (default: 5s)
sdk.WithPingInterval(duration)             // Keepalive interval (default: 5s)
sdk.WithErrorHandler(func(error))          // Connection error handler
```

## Examples

### App Sessions Example

Comprehensive example demonstrating app session lifecycle and operations.

See [examples/app_sessions/lifecycle.go](examples/app_sessions/lifecycle.go)

```bash
cd examples/app_sessions
go run lifecycle.go
```

This example demonstrates:
- Creating app sessions with multiple participants
- Depositing assets into app sessions
- Operating on app session state (redistributing allocations)
- Atomic rebalancing across multiple app sessions
- Withdrawing from app sessions
- Closing app sessions

The example walks through a complete multi-party app session scenario with three wallets.

## Types

All types are imported from `pkg/core` and `pkg/app`:

```go
// Core types
core.State           // Channel state
core.Channel         // Channel info
core.Transition      // State transition
core.Transaction     // Transaction record
core.Asset           // Asset info
core.Token           // Token implementation
core.Blockchain      // Blockchain info

// Core channel session key types
core.ChannelSessionKeyStateV1  // Channel session key state
// Fields: UserAddress, SessionKey, Version (uint64), Assets []string,
//         ExpiresAt (time.Time), UserSig string

// App registry types
app.AppV1                 // Application definition
app.AppInfoV1             // Application info with timestamps

// App session types
app.AppSessionInfoV1      // Session info
app.AppDefinitionV1       // Session definition
app.AppStateUpdateV1      // Session update
app.AppSessionKeyStateV1  // App session key state
// Fields: UserAddress, SessionKey, Version (uint64), ApplicationIDs []string,
//         AppSessionIDs []string, ExpiresAt (time.Time), UserSig string
```

## Operation Internals

For understanding how operations work under the hood:

### Deposit Flow (New Channel)
1. Create channel definition
2. Create void state
3. Set home ledger (token, chain)
4. Apply deposit transition
5. Sign state
6. Request channel creation from node (co-sign)
7. Return co-signed state

### Deposit Flow (Existing Channel)
1. Get latest state
2. Create next state
3. Apply deposit transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### Withdraw Flow
1. Get latest state
2. Create next state
3. Apply withdrawal transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### Transfer Flow
1. Get latest state
2. Create next state
3. Apply transfer transition
4. Sign state
5. Submit to node (co-sign)
6. Return co-signed state

### CloseHomeChannel Flow
1. Get latest state
2. Verify channel exists
3. Create next state
4. Apply finalize transition
5. Sign state
6. Submit to node (co-sign)
7. Return co-signed state

### Checkpoint Flow
1. Get latest signed state (both signatures)
2. Determine blockchain ID from state's home ledger
3. Get on-chain channel status
4. Route based on transition type + status:
   - Void channel -> `blockchainClient.Create()`
   - Existing channel -> `blockchainClient.Checkpoint()`
   - Finalize -> `blockchainClient.Close()`
5. Return transaction hash

## Requirements

- Go 1.21+
- Running Clearnode instance
- Blockchain RPC endpoint (for Checkpoint settlement)

## License

Part of the Nitrolite project.
