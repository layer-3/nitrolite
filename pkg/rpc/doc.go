// Package rpc provides the core data structures and utilities for the Nitrolite Node RPC protocol.
//
// This package implements a secure RPC communication protocol designed for
// blockchain and distributed systems. It provides strong typing, efficient encoding, and
// clear error handling with versioned API endpoints.
//
// # Protocol Overview
//
// The protocol uses a request-response pattern with compact JSON encoding:
//
//   - Messages use typed communication (request, response, event, error response)
//   - All messages include timestamps for replay protection
//   - Methods follow versioned naming: {group}.v{version}.{action}
//   - Responses preserve the request method for easy correlation
//
// # Core Types
//
// Messages are the fundamental unit of communication:
//
//	type Message struct {
//	    Type      MsgType // Message type: Req, Resp, Event, or RespErr
//	    RequestID uint64  // Unique request identifier
//	    Method    string  // RPC method name (e.g., "node.v1.ping")
//	    Payload   Payload // Method parameters or response data
//	    Timestamp uint64  // Unix milliseconds timestamp
//	}
//
// Message types:
//
//	MsgTypeReq     = 1  // Request message
//	MsgTypeResp    = 2  // Success response message
//	MsgTypeEvent   = 3  // Server-initiated event notification
//	MsgTypeRespErr = 4  // Error response message
//
// # JSON Encoding
//
// Messages use a compact array encoding for efficiency. A message like:
//
//	Message{
//	    Type:      MsgTypeReq,
//	    RequestID: 12345,
//	    Method:    "node.v1.get_config",
//	    Payload:   {},
//	    Timestamp: 1634567890123,
//	}
//
// Encodes to:
//
//	[1, 12345, "node.v1.get_config", {}, 1634567890123]
//
// This format reduces message size while maintaining readability and compatibility.
//
// # API Versioning
//
// All RPC methods follow a versioned naming convention:
//
//	{group}.v{version}.{action}
//
// Examples:
//
//	node.v1.ping              // Node group, version 1, ping action
//	channels.v1.get_channels  // Channels group, version 1, get channels action
//	user.v1.get_balances      // User group, version 1, get balances action
//
// API groups:
//
//   - node: Node configuration and connectivity
//   - channels: Payment channel management and state
//   - app_sessions: Application session operations
//   - session_keys: Session key management
//   - user: User balances and transactions
//
// # Error Handling
//
// Error responses use MsgTypeRespErr and preserve the original request method:
//
//	// Success response
//	Message{Type: MsgTypeResp, Method: "node.v1.ping", ...}
//
//	// Error response
//	Message{Type: MsgTypeRespErr, Method: "node.v1.ping", Payload: {"error": "..."}}
//
// Creating error responses:
//
//	// In a handler
//	if err := validate(request); err != nil {
//	    return NewErrorResponse(request.RequestID, request.Method, err.Error())
//	}
//
// Checking for errors in responses:
//
//	response, err := client.Call(ctx, &request)
//	if err != nil {
//	    return err
//	}
//	if err := response.Error(); err != nil {
//	    // Handle RPC error
//	    return fmt.Errorf("RPC error: %w", err)
//	}
//
// The package provides explicit error types for client communication:
//
//	// Client-facing error - will be sent in response
//	if amount < 0 {
//	    return rpc.Errorf("invalid amount: cannot be negative")
//	}
//
//	// Internal error - generic message sent to client
//	if err := db.Save(); err != nil {
//	    return fmt.Errorf("database error: %w", err)
//	}
//
// # Parameter Handling
//
// The Payload type provides flexible parameter handling with type safety:
//
//	// Creating payload from a struct
//	payload, err := rpc.NewPayload(NodeV1GetAssetsRequest{
//	    ChainID: &chainID,
//	})
//
//	// Extracting payload into a struct
//	var resp NodeV1GetConfigResponse
//	err := response.Payload.Translate(&resp)
//
// # Client Communication
//
// The Client type provides convenient methods for all V1 RPC operations:
//
//	// Create client with WebSocket dialer
//	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
//	client := rpc.NewClient(dialer)
//
//	// Connect to server
//	err := client.Start(ctx, "wss://clearnode-sandbox.yellow.org/v1/ws", func(err error) {
//	    if err != nil {
//	        log.Error("Connection closed", "error", err)
//	    }
//	})
//	if err != nil {
//	    log.Fatal("Failed to start client", "error", err)
//	}
//
//	// Make RPC calls - Node methods
//	err = client.NodeV1Ping(ctx)
//	if err != nil {
//	    log.Error("Ping failed", "error", err)
//	}
//
//	config, err := client.NodeV1GetConfig(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	log.Info("Connected to node", "address", config.NodeAddress)
//
//	// User methods
//	balances, err := client.UserV1GetBalances(ctx, rpc.UserV1GetBalancesRequest{
//	    Wallet: walletAddress,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, balance := range balances.Balances {
//	    log.Info("Balance", "asset", balance.Asset, "amount", balance.Amount)
//	}
//
//	// Channel methods
//	channels, err := client.ChannelsV1GetChannels(ctx, rpc.ChannelsV1GetChannelsRequest{
//	    Wallet: walletAddress,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ## Low-Level Dialer
//
// For direct RPC communication without the convenience methods:
//
//	params, _ := rpc.NewPayload(map[string]string{"key": "value"})
//	request := rpc.NewRequest(12345, "node.v1.ping", params)
//
//	response, err := dialer.Call(ctx, &request)
//	if err != nil {
//	    log.Error("RPC call failed", "error", err)
//	}
//
//	// Handle events manually
//	go func() {
//	    for event := range dialer.EventCh() {
//	        if event == nil {
//	            break
//	        }
//	        log.Info("Received event", "method", event.Method)
//	    }
//	}()
//
// # API Types
//
// The package includes comprehensive type definitions for the Nitrolite Node V1 RPC API:
//
// Request/Response Types (organized by group):
//   - Node: NodeV1PingRequest/Response, NodeV1GetConfigRequest/Response, etc.
//   - Channels: ChannelsV1GetHomeChannelRequest/Response, etc.
//   - AppSessions: AppSessionsV1CreateAppSessionRequest/Response, etc.
//   - SessionKeys: SessionKeysV1RegisterRequest/Response, etc.
//   - User: UserV1GetBalancesRequest/Response, etc.
//
// Common Types:
//   - ChannelV1: On-chain channel information
//   - StateV1: Channel state with transitions and ledgers
//   - TransactionV1: Transaction records
//   - AssetV1: Supported asset information
//   - BlockchainInfoV1: Supported network information
//
// All types follow the naming convention: {Group}V{Version}{Name}{Request|Response}
//
// # Server Implementation
//
// The package provides a complete RPC server implementation through the WebsocketNode:
//
//	// Create and configure the server
//	config := rpc.WebsocketNodeConfig{
//	    Logger: logger,
//	}
//
//	node, err := rpc.NewWebsocketNode(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register handlers (built-in handlers are registered automatically)
//	// The node.v1.ping handler is registered by default
//
//	node.Handle("user.v1.get_balances", handleGetBalances)
//	node.Handle("channels.v1.get_channels", handleGetChannels)
//
//	// Add middleware
//	node.Use(loggingMiddleware)
//	node.Use(authMiddleware)
//
//	// Create handler groups
//	privateGroup := node.NewGroup("private")
//	privateGroup.Use(requireAuthMiddleware)
//	privateGroup.Handle("channels.v1.request_creation", handleCreateChannel)
//
//	// Start the server
//	http.Handle("/ws", node)
//	http.ListenAndServe(":8080", nil)
//
// Writing handlers:
//
//	func handleGetBalances(c *rpc.Context) {
//	    // Extract parameters
//	    var req UserV1GetBalancesRequest
//	    if err := c.Request.Payload.Translate(&req); err != nil {
//	        c.Fail(nil, "invalid parameters")
//	        return
//	    }
//
//	    // Process request
//	    balances := getBalancesForWallet(req.Wallet)
//
//	    // Create response
//	    resp := UserV1GetBalancesResponse{
//	        Balances: balances,
//	    }
//	    respPayload, _ := rpc.NewPayload(resp)
//
//	    // Send response
//	    c.Succeed("user.v1.get_balances", respPayload)
//	}
//
// Writing middleware:
//
//	func authMiddleware(c *rpc.Context) {
//	    // Check if connection is authenticated
//	    if c.UserID == "" {
//	        // Try to authenticate from request
//	        token := extractToken(c.Request)
//	        userID, err := validateToken(token)
//	        if err != nil {
//	            c.Fail(nil, "authentication required")
//	            return
//	        }
//	        c.UserID = userID
//	    }
//
//	    // Continue to next handler
//	    c.Next()
//	}
//
// # Security Considerations
//
// When using this protocol:
//
//  1. Always validate timestamps to prevent replay attacks
//  2. Use rpc.Errorf() for safe client-facing errors
//  3. Thoroughly validate all parameters
//  4. Use unique request IDs to prevent duplicate processing
//  5. Implement proper authentication middleware
//  6. Rate limit requests to prevent abuse
//
// # Example Usage
//
// Creating and sending a request:
//
//	// Create request
//	params, _ := rpc.NewPayload(NodeV1GetAssetsRequest{})
//	request := rpc.NewRequest(12345, "node.v1.get_assets", params)
//
//	// Marshal and send
//	data, _ := json.Marshal(request)
//	// ... send data over WebSocket ...
//
// Processing a request:
//
//	// Unmarshal request
//	var request rpc.Message
//	err := json.Unmarshal(data, &request)
//
//	// Check message type
//	if request.Type != rpc.MsgTypeReq {
//	    return rpc.NewErrorResponse(request.RequestID, request.Method, "invalid message type")
//	}
//
//	// Process based on method
//	switch request.Method {
//	case "node.v1.ping":
//	    return rpc.NewResponse(request.RequestID, "node.v1.ping", rpc.Payload{})
//	case "user.v1.get_balances":
//	    var params UserV1GetBalancesRequest
//	    if err := request.Payload.Translate(&params); err != nil {
//	        return rpc.NewErrorResponse(request.RequestID, request.Method, "invalid parameters")
//	    }
//	    // ... handle get balances ...
//	}
//
// # Testing
//
// The package includes a comprehensive test suite:
//
//   - client_test.go: Unit tests for all V1 client methods
//   - dialer_test.go: Tests for WebSocket dialer functionality
//   - message_test.go: Tests for message encoding/decoding
//   - payload_test.go: Tests for payload handling
//
// Run tests with:
//
//	go test github.com/layer-3/nitrolite/pkg/rpc
package rpc
