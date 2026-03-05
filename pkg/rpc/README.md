# RPC Package

The `pkg/rpc` package provides the core data structures and utilities for the Nitrolite Node RPC protocol. This package implements a secure RPC communication protocol designed for blockchain and distributed systems with strong typing, efficient encoding, and clear error handling.

## Overview

### Protocol Features
- **Versioned API**: Clear API versioning with `{group}.v{version}.{action}` method naming
- **Type Safety**: Strongly-typed request/response structures for all operations
- **WebSocket Transport**: Persistent connections with automatic reconnection
- **Error Handling**: Explicit error types with `MsgTypeRespErr`

### Client Features
- **High-Level Client**: Type-safe methods for all RPC operations
- **WebSocket Transport**: Persistent connection with keep-alive
- **Channel Management**: Query and submit channel states
- **Application Sessions**: Multi-party state channel applications
- **Session Keys**: Register and manage session keys

### Server Features
- **RPC Server**: Complete server implementation with WebSocket transport
- **Handler Registration**: Simple method-based request routing
- **Middleware Support**: Composable request processing pipeline
- **Handler Groups**: Organize endpoints with shared middleware
- **Connection Management**: Automatic connection lifecycle handling

## Core Components

### Messages

The protocol uses four message types:

- **MsgTypeReq (1)**: Request message from client
- **MsgTypeResp (2)**: Success response from server
- **MsgTypeEvent (3)**: Server-initiated event notification
- **MsgTypeRespErr (4)**: Error response from server

All messages share a common structure with Type, RequestID, Method, Payload, and Timestamp.

### Message Structure

Messages are the fundamental unit of communication:

```go
type Message struct {
    Type      MsgType  // Message type: 1=Req, 2=Resp, 3=Event, 4=RespErr
    RequestID uint64   // Unique request identifier
    Method    string   // RPC method name (e.g., "node.v1.ping")
    Payload   Payload  // Method parameters or response data
    Timestamp uint64   // Unix milliseconds timestamp
}
```

Messages use a compact JSON array encoding for efficiency:

```json
[1, 12345, "node.v1.get_config", {}, 1634567890123]
```

This format: `[Type, RequestID, Method, Payload, Timestamp]`

### API Versioning

All RPC methods follow a versioned naming convention:

```
{group}.v{version}.{action}
```

Examples:
- `node.v1.ping` - Node group, version 1, ping action
- `channels.v1.get_channels` - Channels group, version 1, get channels action
- `user.v1.get_balances` - User group, version 1, get balances action

API groups:
- **node**: Node configuration and connectivity
- **channels**: Payment channel management and state
- **app_sessions**: Application session operations
- **session_keys**: Session key management
- **user**: User balances and transactions

### Type Naming

All V1 API types follow the convention: `{Group}V{Version}{Name}{Request|Response}`

Examples:
- `NodeV1PingRequest` / `NodeV1PingResponse`
- `ChannelsV1GetChannelsRequest` / `ChannelsV1GetChannelsResponse`
- `UserV1GetBalancesRequest` / `UserV1GetBalancesResponse`

### Error Handling

Error responses use `MsgTypeRespErr` and preserve the original request method:

```go
// Success response
Message{Type: MsgTypeResp, Method: "node.v1.ping", ...}

// Error response
Message{Type: MsgTypeRespErr, Method: "node.v1.ping", Payload: {"error": "..."}}
```

The package provides explicit error types for client communication:

```go
// Client-facing error - will be sent in response
if amount < 0 {
    return rpc.Errorf("invalid amount: cannot be negative")
}

// Internal error - generic message sent to client
if err := db.Save(); err != nil {
    return fmt.Errorf("database error: %w", err)
}
```

## Installation

```go
import "github.com/layer-3/nitrolite/pkg/rpc"
```

## Server Usage

### Creating an RPC Server

```go
import (
    "github.com/layer-3/nitrolite/pkg/rpc"
    "github.com/layer-3/nitrolite/pkg/log"
)

// Create server configuration
config := rpc.WebsocketNodeConfig{
    Logger: logger,  // Required: for structured logging

    // Connection lifecycle callbacks
    OnConnectHandler: func(send rpc.SendResponseFunc) {
        log.Info("New connection established")
    },
    OnDisconnectHandler: func(wallet string) {
        log.Info("Connection closed", "wallet", wallet)
    },
}

// Create the RPC node
node, err := rpc.NewWebsocketNode(config)
if err != nil {
    log.Fatal("Failed to create node", "error", err)
}

// Register handlers (node.v1.ping is built-in)
node.Handle("node.v1.get_config", handleGetConfig)
node.Handle("user.v1.get_balances", handleGetBalances)

// Add global middleware
node.Use(loggingMiddleware)
node.Use(rateLimitMiddleware)

// Create groups for organized endpoints
channelsV1Group := node.NewGroup("channels.v1")
channelsV1Group.Handle("channels.v1.submit_state", handleSubmitState)

// Start the server
http.Handle("/ws", node)
log.Fatal(http.ListenAndServe(":8080", nil))
```

### Writing Handlers

Handlers process RPC requests and generate responses:

```go
func handleGetBalances(c *rpc.Context) {

    // Extract and validate parameters
    var req rpc.UserV1GetBalancesRequest
    if err := c.Request.Payload.Translate(&req); err != nil {
        c.Fail(nil, "invalid parameters")
        return
    }

    // Access connection storage
    if lastCheck, ok := c.Storage.Get("last_balance_check"); ok {
        log.Debug("Last balance check", "time", lastCheck)
    }
    c.Storage.Set("last_balance_check", time.Now())

    // Process the request
    balances, err := ledger.GetBalances(req.Wallet)
    if err != nil {
        log.Error("Failed to get balances", "error", err)
        c.Fail(err, "failed to retrieve balances")
        return
    }

    // Create response
    resp := rpc.UserV1GetBalancesResponse{
        Balances: balances,
    }
    respPayload, _ := rpc.NewPayload(resp)

    // Send successful response
    c.Succeed("user.v1.get_balances", respPayload)
}
```

### Writing Middleware

Middleware can process requests before/after handlers:

```go
func loggingMiddleware(c *rpc.Context) {
    start := time.Now()
    method := c.Request.Method

    // Pre-processing
    log.Info("Request started",
        "method", method,
        "requestID", c.Request.RequestID)

    // Continue to next handler
    c.Next()

    // Post-processing
    duration := time.Since(start)
    log.Info("Request completed",
        "method", method,
        "requestID", c.Request.RequestID,
        "duration", duration)
}

func rateLimitMiddleware(c *rpc.Context) {
    key := fmt.Sprintf("rate_limit_%s", userID)

    // Get current count
    count := 0
    if val, ok := c.Storage.Get(key); ok {
        count = val.(int)
    }

    if count >= 100 {
        c.Fail(nil, "rate limit exceeded")
        return
    }

    // Increment and continue
    c.Storage.Set(key, count+1)
    c.Next()
}
```

## Client Usage

### Quick Start

```go
import "github.com/layer-3/nitrolite/pkg/rpc"

// Create client
dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
client := rpc.NewClient(dialer)

// Set up event handlers (optional)
go func() {
    for event := range dialer.EventCh() {
        if event == nil {
            break
        }
        log.Info("Received event", "method", event.Method)
    }
}()

// Connect to server
ctx := context.Background()
err := client.Start(ctx, "wss://node.example.com/ws", func(err error) {
    if err != nil {
        log.Error("Connection closed", "error", err)
    }
})
if err != nil {
    log.Fatal("Failed to start client", "error", err)
}

// Make RPC calls
err = client.NodeV1Ping(ctx)
if err != nil {
    log.Error("Ping failed", "error", err)
}

config, err := client.NodeV1GetConfig(ctx)
if err != nil {
    log.Fatal(err)
}
log.Info("Connected to node", "address", config.NodeAddress)
```

### Available Client Methods

#### Node Methods
```go
// Ping the server
err := client.NodeV1Ping(ctx)

// Get node configuration
config, err := client.NodeV1GetConfig(ctx)

// Get supported assets (optional chain filter)
chainID := uint32(1)
assets, err := client.NodeV1GetAssets(ctx, rpc.NodeV1GetAssetsRequest{
    ChainID: &chainID,
})
```

#### User Methods
```go
// Get user balances
balances, err := client.UserV1GetBalances(ctx, rpc.UserV1GetBalancesRequest{
    Wallet: walletAddress,
})

// Get transactions
txs, err := client.UserV1GetTransactions(ctx, rpc.UserV1GetTransactionsRequest{
    Wallet: walletAddress,
})
```

#### Channel Methods
```go
// Get home channel
homeChannel, err := client.ChannelsV1GetHomeChannel(ctx, rpc.ChannelsV1GetHomeChannelRequest{
    Wallet: walletAddress,
    Asset:  "usdc",
})

// Get escrow channel
escrowChannel, err := client.ChannelsV1GetEscrowChannel(ctx, rpc.ChannelsV1GetEscrowChannelRequest{
    Wallet: walletAddress,
    Asset:  "usdc",
})

// Get all channels
channels, err := client.ChannelsV1GetChannels(ctx, rpc.ChannelsV1GetChannelsRequest{
    Wallet: walletAddress,
})

// Get latest state
state, err := client.ChannelsV1GetLatestState(ctx, rpc.ChannelsV1GetLatestStateRequest{
    Wallet: walletAddress,
    Asset:  "usdc",
})

// Get states with filters
states, err := client.ChannelsV1GetStates(ctx, rpc.ChannelsV1GetStatesRequest{
    Wallet: walletAddress,
    Asset:  &asset,
})

// Request channel creation
creation, err := client.ChannelsV1RequestCreation(ctx, rpc.ChannelsV1RequestCreationRequest{
    Wallet: walletAddress,
    Asset:  "usdc",
})

// Submit state
submitResp, err := client.ChannelsV1SubmitState(ctx, rpc.ChannelsV1SubmitStateRequest{
    State: stateData,
})
```

#### App Session Methods
```go
// Get app definition
appDef, err := client.AppSessionsV1GetAppDefinition(ctx, rpc.AppSessionsV1GetAppDefinitionRequest{
    AppSessionID: sessionID,
})

// Get app sessions
sessions, err := client.AppSessionsV1GetAppSessions(ctx, rpc.AppSessionsV1GetAppSessionsRequest{
    Wallet: walletAddress,
})

// Create app session
session, err := client.AppSessionsV1CreateAppSession(ctx, rpc.AppSessionsV1CreateAppSessionRequest{
    Definition: definition,
})

// Close app session
closeResp, err := client.AppSessionsV1CloseAppSession(ctx, rpc.AppSessionsV1CloseAppSessionRequest{
    AppSessionID: sessionID,
})

// Submit deposit state
depositResp, err := client.AppSessionsV1SubmitDepositState(ctx, rpc.AppSessionsV1SubmitDepositStateRequest{
    AppSessionID: sessionID,
    State:        stateData,
})

// Submit app state
appStateResp, err := client.AppSessionsV1SubmitAppState(ctx, rpc.AppSessionsV1SubmitAppStateRequest{
    AppSessionID: sessionID,
    State:        stateData,
})
```

#### Session Key Methods
```go
// Register session key
registerResp, err := client.SessionKeysV1Register(ctx, rpc.SessionKeysV1RegisterRequest{
    Wallet:     walletAddress,
    SessionKey: sessionKeyAddress,
})

// Get session keys
keys, err := client.SessionKeysV1GetSessionKeys(ctx, rpc.SessionKeysV1GetSessionKeysRequest{
    Wallet: walletAddress,
})

// Revoke session key
revokeResp, err := client.SessionKeysV1RevokeSessionKey(ctx, rpc.SessionKeysV1RevokeSessionKeyRequest{
    Wallet:     walletAddress,
    SessionKey: sessionKeyToRevoke,
})
```

## Low-Level Usage

### Creating Messages Directly

```go
// Create parameters
params, err := rpc.NewPayload(rpc.NodeV1GetAssetsRequest{})
if err != nil {
    return err
}

// Create a request
request := rpc.NewRequest(
    12345,                      // Request ID
    "node.v1.get_assets",       // Method name
    params,                     // Parameters
)

// Send via dialer
response, err := dialer.Call(ctx, &request)
if err != nil {
    return err
}

// Check for errors
if err := response.Error(); err != nil {
    return fmt.Errorf("RPC error: %w", err)
}

// Process response
var result rpc.NodeV1GetAssetsResponse
if err := response.Payload.Translate(&result); err != nil {
    return err
}
```

### Working with Payloads

```go
// Creating payload from a struct
type MyRequest struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}

req := MyRequest{
    Field1: "value",
    Field2: 42,
}

payload, err := rpc.NewPayload(req)
if err != nil {
    return err
}

// Extracting payload into a struct
var received MyRequest
if err := payload.Translate(&received); err != nil {
    return rpc.Errorf("invalid parameters: %v", err)
}
```

### Error Handling

```go
// Check response for errors
response, err := client.NodeV1GetAssets(ctx, req)
if err != nil {
    // This handles both transport errors and RPC errors
    return err
}

// In handlers, use rpc.Errorf for client-facing errors
if amount < 0 {
    return rpc.Errorf("invalid amount: cannot be negative")
}

// Regular errors are treated as internal errors
if err := db.Query(); err != nil {
    return fmt.Errorf("database error: %w", err)
}
```

## Advanced Usage

### WebSocket Dialer Configuration

```go
cfg := rpc.WebsocketDialerConfig{
    // Duration to wait for WebSocket handshake (default: 5s)
    HandshakeTimeout: 5 * time.Second,

    // How often to send ping messages (default: 5s)
    PingInterval: 5 * time.Second,

    // Request ID used for ping messages (default: 100)
    PingRequestID: 100,

    // Buffer size for event channel (default: 100)
    EventChanSize: 100,
}

dialer := rpc.NewWebsocketDialer(cfg)
```

### Concurrent RPC Calls

```go
// The client supports concurrent calls from multiple goroutines
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()

        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        err := client.NodeV1Ping(ctx)
        if err != nil {
            log.Error("Ping failed", "id", id, "error", err)
            return
        }

        log.Info("Ping succeeded", "id", id)
    }(i)
}

wg.Wait()
```

## Security Considerations

When using this protocol:

1. **Timestamp Validation**: Validate timestamps to prevent replay attacks
2. **Parameter Validation**: Thoroughly validate all parameters
3. **Error Messages**: Use `rpc.Errorf()` for safe client-facing errors
4. **Request IDs**: Use unique request IDs to prevent duplicate processing
5. **Rate Limiting**: Implement rate limiting middleware

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./pkg/rpc

# Run with race detector
go test -race ./pkg/rpc

# Run with verbose output
go test -v ./pkg/rpc
```

Test coverage includes:
- Message encoding/decoding
- Payload handling
- WebSocket dialer functionality
- All V1 client methods
- Context and middleware
- Error handling

## Dependencies

- Standard library: `encoding/json`, `errors`, `fmt`, `time`, `context`, `sync`
- WebSocket: `github.com/gorilla/websocket`
- UUID: `github.com/google/uuid`

## See Also

- [API Documentation](api.yaml) - OpenAPI specification for V1 API
- Package documentation: `go doc github.com/layer-3/nitrolite/pkg/rpc`
