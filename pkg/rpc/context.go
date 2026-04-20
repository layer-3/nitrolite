package rpc

import (
	"context"
	"sync"
)

// ApplicationIDQueryParam is the URL query parameter clients use to declare
// their application identity during the WebSocket upgrade (e.g.
// ws://host/?app_id=my-app). The same key is used to store the value in
// per-connection SafeStorage.
const ApplicationIDQueryParam = "app_id"

// GetApplicationID returns the application identifier associated with the current
// connection (supplied by the client as the ApplicationIDQueryParam query parameter
// during the WebSocket upgrade). Returns an empty string if no app_id was provided.
//
// The value is an advisory origin tag — it is self-declared by the client and
// must not be used for authentication or access control.
func GetApplicationID(c *Context) string {
	if c == nil || c.Storage == nil {
		return ""
	}
	v, ok := c.Storage.Get(ApplicationIDQueryParam)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// Handler defines the function signature for RPC request processors.
// Handlers receive a Context containing the request and all necessary
// information to process it. They can call c.Next() to delegate to
// the next handler in the middleware chain, enabling composable
// request processing pipelines.
type Handler func(c *Context)

// SendResponseFunc is a function type for sending server-initiated RPC notifications.
// Unlike regular responses that reply to client requests, these functions enable
// the server to push unsolicited messages to clients (e.g., balance updates,
// connection events). The method parameter specifies the notification type,
// and params contains the notification data.
type SendResponseFunc func(method string, params Payload)

// Context encapsulates all information related to an RPC request and provides
// methods for handlers to process and respond. It implements a middleware
// pattern where handlers can be chained together, each having the ability
// to process the request, modify the context, or delegate to the next handler.
//
// The Context serves multiple purposes:
//   - Request/response container: Holds the incoming request and outgoing response
//   - Middleware chain management: Tracks and executes the handler chain
//   - Session state: Provides per-connection storage for maintaining state
//   - Response helpers: Convenient methods for success and error responses
type Context struct {
	// Context is the standard Go context for the request
	Context context.Context
	// Request is the original RPC request message
	Request Message
	// Response is the response message to be sent back to the client
	Response Message
	// Storage provides per-connection storage for session data
	Storage *SafeStorage

	// handlers is the remaining handler chain to execute
	handlers []Handler
}

// Next advances the middleware chain by executing the next handler.
// This enables handlers to perform pre-processing, call Next() to
// delegate to subsequent handlers, then perform post-processing.
// If there are no more handlers in the chain, Next() returns
// immediately without error.
//
// Example middleware pattern:
//
//	func invalidRequest(c *Context) {
//	    // Pre-processing: check request message type
//	    if c.Request.Type != MsgTypeReq {
//	        c.Fail(nil, "invalid request")
//	        return
//	    }
//	    c.Next() // Continue to next handler
//	}
func (c *Context) Next() {
	if len(c.handlers) == 0 {
		return
	}

	handler := c.handlers[0]
	c.handlers = c.handlers[1:]
	handler(c)
}

// Succeed sets a successful response for the RPC request.
// This method should be called by handlers when the request has been
// processed successfully. The method parameter typically matches the
// request method, and params contains the result data.
//
// Example:
//
//	func handleGetBalance(c *Context) {
//	    balance := getBalanceForUser(c.Request.Payload{"userWallet"})
//	    c.Succeed("get_balance", Payload{"balance": balance})
//	}
func (c *Context) Succeed(method string, payload Payload) {
	c.Response = NewResponse(
		c.Request.RequestID,
		method,
		payload,
	)
}

// Fail sets an error response for the RPC request. This method should be called by handlers
// when an error occurs during request processing.
//
// Error handling behavior:
//   - If err is an RPCError: The exact error message is sent to the client
//   - If err is any other error type: The fallbackMessage is sent to the client
//   - If both err is nil/non-RPCError AND fallbackMessage is empty: A generic error message is sent
//
// This design allows handlers to control what error information is exposed to clients:
//   - Use RPCError for client-safe, descriptive error messages
//   - Use regular errors with a fallbackMessage to hide internal error details
//
// Usage examples:
//
//	// Hide internal error details from client
//	balance, err := ledger.GetBalance(account)
//	if err != nil {
//		c.Fail(err, "failed to retrieve balance")
//		return
//	}
//
//	// Validation error with no internal error
//	if len(params) < 3 {
//		c.Fail(nil, "invalid parameters: expected at least 3")
//		return
//	}
//
// The response will have Method="error" and Params containing the error message.
func (c *Context) Fail(err error, fallbackMessage string) {
	message := fallbackMessage
	if _, ok := err.(Error); ok {
		message = err.Error()
	}
	if message == "" {
		message = defaultNodeErrorMessage
	}

	c.Response = NewErrorResponse(
		c.Request.RequestID,
		c.Request.Method,
		message,
	)
}

// SafeStorage provides thread-safe key-value storage for connection-specific data.
// Each connection gets its own SafeStorage instance that persists for the
// connection's lifetime. This enables handlers to store and retrieve session
// state, authentication tokens, rate limiting counters, or any other
// per-connection data across multiple requests.
//
// Common use cases:
//   - Storing authentication state and policies
//   - Caching frequently accessed data
//   - Maintaining request counters for rate limiting
//   - Storing connection-specific configuration
type SafeStorage struct {
	// mu protects concurrent access to the storage map
	mu sync.RWMutex
	// storage holds the key-value pairs
	storage map[string]any
}

// NewSafeStorage creates a new thread-safe storage instance.
// The storage starts empty and can be used immediately for
// storing connection-specific data.
func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		storage: make(map[string]any),
	}
}

// Set stores a value with the given key in the storage.
// If the key already exists, its value is overwritten.
// The value can be of any type. This method is thread-safe
// and can be called concurrently from multiple goroutines.
//
// Example:
//
//	storage.Set("auth_token", "bearer-xyz123")
//	storage.Set("rate_limit_count", 42)
//	storage.Set("user_preferences", userPrefs)
func (s *SafeStorage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

// Get retrieves a value by key from the storage.
// Returns the value and true if the key exists, or nil and false
// if the key is not found. The caller must type-assert the returned
// value to the expected type.
//
// Example:
//
//	if val, ok := storage.Get("auth_token"); ok {
//	    token := val.(string)
//	    // Use token...
//	}
func (s *SafeStorage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.storage[key]
	return value, exists
}
