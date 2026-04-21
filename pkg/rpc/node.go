package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/log"
)

const (
	defaultNodeErrorMessage = "an error occurred while processing the request"
)

const (
	// nodeGroupHandlerPrefix is the prefix used for all handler group IDs
	nodeGroupHandlerPrefix = "group."
	// nodeGroupRoot is the identifier for the root handler group
	nodeGroupRoot = "root"
)

// Node represents an RPC server that manages client connections and routes
// messages to appropriate handlers. It provides a foundation for building
// RPC-based services with support for middleware, authentication, and
// server-initiated notifications. The interface is transport-agnostic,
// allowing for different implementations (WebSocket, HTTP/2, etc.).
type Node interface {
	// ServeHTTP handles incoming HTTP requests and upgrades them to
	// the appropriate transport protocol (e.g., WebSocket). It manages
	// connection lifecycle, including reading requests, routing to
	// handlers, and sending responses.
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Handle registers a handler function for a specific RPC method.
	// When a request with the matching method name is received,
	// the handler will be invoked with the request context.
	Handle(method string, handler Handler)

	// Notify sends a server-initiated notification to a specific user.
	// All active connections for the user will receive the notification.
	// If the user has no active connections, the notification is dropped.
	Notify(userID string, method string, params Payload)

	// Use adds global middleware that will be executed for all requests.
	// Middleware is executed in the order it was added, before any
	// method-specific handlers.
	Use(middleware Handler)

	// NewGroup creates a new handler group for organizing related endpoints.
	// Groups can have their own middleware and can be nested to create
	// hierarchical handler structures.
	NewGroup(name string) HandlerGroup
}

type HandlerGroup interface {
	Handle(method string, handler Handler)
	Use(middleware Handler)
	NewGroup(name string) HandlerGroup
}

var (
	_ Node         = &WebsocketNode{}
	_ http.Handler = &WebsocketNode{}

	_ HandlerGroup = &WebsocketHandlerGroup{}
)

// WebsocketNode implements the Node interface using WebSocket as the transport layer.
// It provides a complete RPC server implementation with the following features:
//
//   - WebSocket connection management with automatic cleanup
//   - Request routing based on method names
//   - Middleware support at global and group levels
//   - Cryptographic signing of all responses
//   - Connection authentication and re-authentication
//   - Server-initiated notifications to specific users
//   - Configurable timeouts and buffer sizes
//   - Structured logging for debugging and monitoring
//
// The node automatically handles connection lifecycle, including:
//   - WebSocket protocol upgrade
//   - Concurrent message processing
//   - Graceful connection shutdown
//   - Resource cleanup on disconnection
type WebsocketNode struct {
	// upgrader handles the HTTP to WebSocket protocol upgrade
	upgrader websocket.Upgrader
	// cfg contains configuration for the node
	cfg WebsocketNodeConfig
	// groupId identifies this node's handler group (defaults to "group.root")
	groupId string
	// handlerChain maps handler IDs to their middleware/handler chains
	handlerChain map[string][]Handler
	// routes maps RPC method names to their handler chain path (e.g., ["group.root", "group.private", "method"])
	routes map[string][]string
	// connHub manages all active WebSocket connections
	connHub *ConnectionHub
}

// WebsocketNodeConfig contains all configuration options for creating a WebsocketNode.
// The only required field is Logger; all others have sensible defaults.
type WebsocketNodeConfig struct {
	// Logger is used for structured logging throughout the node (required).
	Logger log.Logger

	// Connection lifecycle callbacks:
	// ObserveConnections is called with the current number of active connections whenever a connection is established or closed.
	ObserveConnections ObserveConnectionsFn

	// WebSocket upgrader configuration:

	// WsUpgraderReadBufferSize sets the read buffer size for the WebSocket upgrader (default: 1024).
	WsUpgraderReadBufferSize int
	// WsUpgraderWriteBufferSize sets the write buffer size for the WebSocket upgrader (default: 1024).
	WsUpgraderWriteBufferSize int
	// WsUpgraderCheckOrigin validates the origin of incoming WebSocket requests.
	// Default allows all origins; implement this for CORS protection.
	WsUpgraderCheckOrigin func(r *http.Request) bool

	// Connection-level configuration:

	// WsConnWriteTimeout is the maximum time to wait for a write operation (default: 5s).
	// Connections that exceed this timeout are considered unresponsive and closed.
	WsConnWriteTimeout time.Duration
	// WsConnWriteBufferSize is the capacity of each connection's outgoing message queue (default: 10).
	WsConnWriteBufferSize int
	// WsConnProcessBufferSize is the capacity of each connection's incoming message queue (default: 10).
	WsConnProcessBufferSize int
}

// NewWebsocketNode creates a new WebsocketNode instance with the provided configuration.
// The node is ready to accept WebSocket connections after creation.
//
// Required configuration:
//   - Logger: Used for structured logging
//
// Returns an error if required configuration is missing.
func NewWebsocketNode(config WebsocketNodeConfig) (*WebsocketNode, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	config.Logger = config.Logger.WithName("rpc-node")

	if config.ObserveConnections == nil {
		// Default implementation does nothing, but can be overridden for monitoring
		config.ObserveConnections = func(region, origin string, count uint32) {}
	}
	if config.WsUpgraderReadBufferSize <= 0 {
		// It's the optimal default value as recommended
		// by the library documentation for most use cases
		config.WsUpgraderReadBufferSize = 1024
	}
	if config.WsUpgraderWriteBufferSize <= 0 {
		// It's the optimal default value as recommended
		// by the library documentation for most use cases
		config.WsUpgraderWriteBufferSize = 1024
	}
	if config.WsUpgraderCheckOrigin == nil {
		// Default allows all origins as this application is designed to be public
		// and accept connections from multiple different applications without
		// origin restrictions by default
		config.WsUpgraderCheckOrigin = func(r *http.Request) bool {
			return true // Allow all origins by default
		}
	}

	node := &WebsocketNode{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.WsUpgraderReadBufferSize,
			WriteBufferSize: config.WsUpgraderWriteBufferSize,
			CheckOrigin:     config.WsUpgraderCheckOrigin,
		},
		cfg:          config,
		groupId:      nodeGroupHandlerPrefix + nodeGroupRoot,
		handlerChain: make(map[string][]Handler),
		routes:       make(map[string][]string),
		connHub:      NewConnectionHub(config.ObserveConnections),
	}

	return node, nil
}

// ServeHTTP implements http.Handler, making the node compatible with standard HTTP servers.
// This method:
//  1. Upgrades incoming HTTP requests to WebSocket connections
//  2. Creates a unique connection ID and manages connection state
//  3. Spawns goroutines for concurrent message processing
//  4. Invokes lifecycle callbacks (OnConnect, OnDisconnect, etc.)
//  5. Blocks until the connection is closed
//
// Each connection runs independently with its own goroutines for:
//   - Reading incoming messages
//   - Processing requests and routing to handlers
//   - Writing outgoing responses
//   - Monitoring connection health
//
// The method ensures proper cleanup when connections close, including
// removing the connection from the hub and invoking disconnect callbacks.
func (wn *WebsocketNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	applicationID := r.URL.Query().Get(ApplicationIDQueryParam)
	if applicationID != "" && !app.IsValidApplicationID(applicationID) {
		wn.cfg.Logger.Warn("rejecting connection with invalid application_id",
			"remoteAddr", r.RemoteAddr, "applicationID", applicationID)
		http.Error(w, fmt.Sprintf("invalid %s: must match %s", ApplicationIDQueryParam, app.ApplicationIDRegex.String()), http.StatusBadRequest)
		return
	}

	wsConnection, err := wn.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wn.cfg.Logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer wsConnection.Close()

	connectionID := uuid.NewString()

	connConfig := WebsocketConnectionConfig{
		ConnectionID:      connectionID,
		Origin:            r.Header.Get("Origin"),
		WebsocketConn:     wsConnection,
		Logger:            wn.cfg.Logger,
		ProcessBufferSize: wn.cfg.WsConnProcessBufferSize,
		WriteBufferSize:   wn.cfg.WsConnWriteBufferSize,
	}
	connection, err := NewWebsocketConnection(connConfig)
	if err != nil {
		wn.cfg.Logger.Error("failed to create WebSocket connection", "error", err, "connectionID", connectionID)
		return
	}
	if err := wn.connHub.Add(connection); err != nil {
		wn.cfg.Logger.Error("failed to add connection to hub", "error", err, "connectionID", connectionID)
		return
	}

	wn.cfg.Logger.Info("new WebSocket connection established", "connectionID", connectionID)

	// Cleanup function executed when connection closes
	defer func() {
		wn.connHub.Remove(connectionID)

		wn.cfg.Logger.Info("connection closed", "connectionID", connectionID)
	}()

	parentCtx, cancel := context.WithCancel(r.Context())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	childHandleClosure := func(_ error) {
		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	go connection.Serve(parentCtx, childHandleClosure)
	go wn.processRequests(connection, applicationID, parentCtx, childHandleClosure)

	wg.Wait()
}

// processRequests is the main request processing loop for a connection.
// It:
//  1. Reads raw messages from the connection's request channel
//  2. Unmarshals and validates incoming requests
//  3. Looks up the appropriate handler chain for the method
//  4. Creates a Context with the request and executes the handler chain
//  5. Sends the signed response back to the client
//  6. Handles re-authentication if the UserID changes
//
// The method runs until the connection closes or the context is cancelled.
// Each connection has its own SafeStorage instance for maintaining state
// across requests.
func (wn *WebsocketNode) processRequests(conn Connection, applicationID string, parentCtx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines when done
	safeStorage := NewSafeStorage()
	if applicationID != "" {
		safeStorage.Set(ApplicationIDQueryParam, applicationID)
	}

	for {
		var messageBytes []byte
		select {
		case <-parentCtx.Done():
			wn.cfg.Logger.Debug("context done, stopping message processing")
			return
		case messageBytes = <-conn.RawRequests():
			if len(messageBytes) == 0 {
				return // Exit if the message is empty (connection closed)
			}
		}

		req := Message{}
		if err := json.Unmarshal(messageBytes, &req); err != nil {
			wn.cfg.Logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
			wn.sendErrorResponse(conn, req.RequestID, "", "invalid message format")
			continue
		}

		methodRoute, ok := wn.routes[req.Method]
		if !ok || len(methodRoute) == 0 {
			wn.cfg.Logger.Debug("no handlers' route found for method", "method", req.Method)
			wn.sendErrorResponse(conn, req.RequestID, req.Method, fmt.Sprintf("unknown method: %s", req.Method))
			continue
		}

		var routeHandlers []Handler
		for _, handlersId := range methodRoute {
			handlers := wn.handlerChain[handlersId]
			if len(handlers) > 0 {
				routeHandlers = append(routeHandlers, handlers...)
			}
		}
		if len(routeHandlers) == 0 {
			wn.sendErrorResponse(conn, req.RequestID, req.Method, fmt.Sprintf("unknown method: %s", req.Method))
			continue
		}

		ctx := &Context{
			Context:  parentCtx,
			Request:  req,
			handlers: routeHandlers,
			Storage:  safeStorage,
		}
		ctx.Next() // Start processing the handlers

		// Marshal the response
		responseBytes, err := json.Marshal(ctx.Response)
		if err != nil {
			wn.sendErrorResponse(conn, req.RequestID, req.Method, defaultNodeErrorMessage)
			wn.cfg.Logger.Error("failed to marshal response", "error", err, "method", req.Method)
			continue
		}

		if !conn.WriteRawResponse(responseBytes) {
			wn.cfg.Logger.Warn("write queue timeout", "connectionID", conn.ConnectionID(), "method", req.Method)
			return // stop processing this conn; it's closing anyway
		}
	}
}

// NewGroup creates a new handler group with the specified name.
// Groups provide a way to organize related handlers and apply
// common middleware. Groups can be nested to create hierarchical
// structures.
//
// Example:
//
//	// Create a group for authenticated endpoints
//	privateGroup := node.NewGroup("private")
//	privateGroup.Use(authMiddleware)
//	privateGroup.Handle("get_balance", handleGetBalance)
//
//	// Create a nested group with additional middleware
//	adminGroup := privateGroup.NewGroup("admin")
//	adminGroup.Use(adminAuthMiddleware)
//	adminGroup.Handle("manage_users", handleManageUsers)
func (wn *WebsocketNode) NewGroup(name string) HandlerGroup {
	return &WebsocketHandlerGroup{
		groupId:     nodeGroupHandlerPrefix + name,
		routePrefix: []string{wn.groupId},
		root:        wn,
	}
}

// Handle registers a handler function for the specified RPC method.
// When a request with a matching method name is received, the handler
// will be invoked with a Context containing the request information.
//
// The handler executes after all global middleware registered with Use().
//
// Panics if:
//   - method is empty
//   - handler is nil
func (wn *WebsocketNode) Handle(method string, handler Handler) {
	wn.handle(method, handler)
	wn.routes[method] = []string{wn.groupId, method}
}

// handle is the internal method for registering handlers.
// It validates inputs and stores the handler in the handler chain.
func (wn *WebsocketNode) handle(method string, handler Handler) {
	if method == "" {
		panic("Websocket method cannot be empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("Websocket handler cannot be nil for method %s", method))
	}

	wn.handlerChain[method] = []Handler{handler}
}

// Use adds global middleware that executes for all requests.
// Middleware is executed in the order it was registered, before
// any method-specific handlers. Common middleware includes:
//   - Authentication checks
//   - Request logging
//   - Rate limiting
//   - Request validation
//
// Example:
//
//	node.Use(loggingMiddleware)
//	node.Use(rateLimitMiddleware)
//	node.Use(authMiddleware)
func (wn *WebsocketNode) Use(middleware Handler) {
	wn.use(wn.groupId, middleware)
}

// use is the internal method for adding middleware to a specific group.
// Middleware is appended to the group's handler chain.
func (wn *WebsocketNode) use(groupId string, middleware Handler) {
	if middleware == nil {
		panic("Websocket middleware handler cannot be nil for group")
	}

	if _, exists := wn.handlerChain[groupId]; !exists {
		wn.handlerChain[groupId] = []Handler{}
	}

	wn.handlerChain[groupId] = append(wn.handlerChain[groupId], middleware)
}

// Notify sends a server-initiated notification to all connections of a specific user.
// This enables the server to push updates to clients without a prior request.
// Common use cases include:
//   - Balance updates after transactions
//   - Status changes in long-running operations
//   - Real-time notifications for user events
//
// The notification is sent to all active connections for the user.
// If the user has no active connections, the notification is silently dropped.
//
// Notifications have RequestID=0 to distinguish them from responses.
func (wn *WebsocketNode) Notify(userID, method string, params Payload) {
	// message, err := prepareRawNotification(method, params)
	// if err != nil {
	// 	wn.cfg.Logger.Error("failed to prepare notification message", "error", err, "userID", userID, "method", method)
	// 	return
	// }

	// wn.connHub.Publish(userID, message)
}

// getSendResponseFunc creates a SendResponseFunc for a specific connection.
// The returned function can be used to send notifications to that connection.
func (wn *WebsocketNode) getSendResponseFunc(conn Connection) SendResponseFunc {
	return func(method string, params Payload) {
		responseBytes, err := prepareRawNotification(method, params)
		if err != nil {
			wn.cfg.Logger.Error("failed to prepare notification message", "error", err, "method", method)
			return
		}

		if conn == nil {
			wn.cfg.Logger.Error("RPCConnection is nil, cannot send message", "method", method)
			return
		}

		conn.WriteRawResponse(responseBytes)
	}
}

// sendErrorResponse sends an error response to a connection.
// It's used for protocol-level errors before request processing.
func (wn *WebsocketNode) sendErrorResponse(conn Connection, requestID uint64, method string, message string) {
	if conn == nil {
		wn.cfg.Logger.Error("connection is nil, cannot send error response", "requestID", requestID)
		return
	}

	res := NewErrorResponse(requestID, method, message)
	responseBytes, err := json.Marshal(res)
	if err != nil {
		wn.cfg.Logger.Error("failed to marshal error response", "error", err)
		return
	}

	conn.WriteRawResponse(responseBytes)
}

// prepareRawNotification creates a server-initiated notification message.
// Unlike responses, notifications don't correspond to a specific request.
func prepareRawNotification(method string, params Payload) ([]byte, error) {
	msg := NewEvent(0, method, params) // RequestID=0 for notifications

	responseBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %w", err)
	}

	return responseBytes, nil
}

// WebsocketHandlerGroup implements the HandlerGroup interface for organizing
// related handlers with shared middleware. Groups support nesting, allowing
// for hierarchical organization of endpoints with inherited middleware chains.
//
// When a request matches a handler in a group, the middleware chain is:
//  1. Global middleware (from Node.Use)
//  2. Parent group middleware (if nested)
//  3. This group's middleware
//  4. The method handler
//
// This enables fine-grained control over request processing pipelines.
type WebsocketHandlerGroup struct {
	// groupId is the unique identifier for this group
	groupId string
	// routePrefix contains the chain of group IDs leading to this group
	routePrefix []string
	// root is a reference to the Node this group belongs to
	root *WebsocketNode
}

// NewGroup creates a nested handler group within this group.
// The nested group inherits the middleware chain from all parent groups,
// allowing for progressive middleware application.
//
// Example:
//
//	api := node.NewGroup("api")
//	api.Use(apiVersionMiddleware)
//
//	v1 := api.NewGroup("v1")
//	v1.Use(v1AuthMiddleware)
//	// Handlers in v1 group will execute: global → api → v1 middleware
func (hg *WebsocketHandlerGroup) NewGroup(name string) HandlerGroup {
	return &WebsocketHandlerGroup{
		groupId:     fmt.Sprintf("%s.%s", hg.groupId, name),
		routePrefix: append(hg.routePrefix, hg.groupId),
		root:        hg.root,
	}
}

// Handle registers a handler for the specified RPC method within this group.
// The handler will execute after all applicable middleware:
//   - Global middleware
//   - Parent group middleware (for nested groups)
//   - This group's middleware
//
// The method parameter must be unique across the entire node.
func (hg *WebsocketHandlerGroup) Handle(method string, handler Handler) {
	hg.root.routes[method] = append(hg.routePrefix, hg.groupId, method)
	hg.root.handle(method, handler)
}

// Use adds middleware to this handler group.
// The middleware will execute for all handlers in this group
// and any nested groups. Middleware is executed in the order
// it was added.
//
// Panics if middleware is nil.
func (hg *WebsocketHandlerGroup) Use(middleware Handler) {
	hg.root.use(hg.groupId, middleware)
}
