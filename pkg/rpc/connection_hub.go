package rpc

import (
	"fmt"
	"sync"
)

// ObserveConnectionsFn is invoked on connect and disconnect with the current per-application
// connection count. A count of 0 signals that the bucket is empty and the observer should
// shed any per-label state (e.g., delete the Prometheus gauge series) to bound cardinality.
type ObserveConnectionsFn func(applicationID string, count uint32)

// ConnectionHub provides centralized management of all active RPC connections.
// It maintains thread-safe mappings between connection IDs and Connection instances,
// as well as user IDs and their associated connections. This enables efficient
// message routing and connection lifecycle management.
//
// Key features:
//   - Thread-safe connection storage and retrieval
//   - User-to-connection mapping for authenticated sessions
//   - Automatic cleanup of auth mappings when connections close
//   - Support for re-authentication (updating user associations)
//   - Broadcast capabilities to all connections for a specific user
type ConnectionHub struct {
	// connections maps connection IDs to RPCConnection instances
	connections map[string]Connection
	// authMapping maps UserIDs to their active connections.
	authMapping map[string]map[string]bool
	// mu protects concurrent access to the maps
	mu sync.RWMutex

	// appConnCount tracks active connection counts keyed by application_id (may be empty string)
	appConnCount map[string]uint32
	// observeConnections is a callback function to report per-application connection counts
	observeConnections ObserveConnectionsFn
}

// NewConnectionHub creates a new ConnectionHub instance with initialized maps.
// The hub is typically used internally by Node implementations to manage
// the lifecycle of all active connections.
func NewConnectionHub(observeConnections ObserveConnectionsFn) *ConnectionHub {
	return &ConnectionHub{
		connections:        make(map[string]Connection),
		authMapping:        make(map[string]map[string]bool),
		appConnCount:       make(map[string]uint32),
		observeConnections: observeConnections,
	}
}

// Add registers a new connection with the hub.
// The connection is indexed by its ConnectionID for fast retrieval.
//
// Returns an error if:
//   - The connection is nil
//   - A connection with the same ID already exists
func (hub *ConnectionHub) Add(conn Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	connID := conn.ConnectionID()

	hub.mu.Lock()

	// If the connection already exists, return an error
	if _, exists := hub.connections[connID]; exists {
		hub.mu.Unlock()
		return fmt.Errorf("connection with ID %s already exists", connID)
	}

	hub.connections[connID] = conn

	appID := conn.ApplicationID()
	hub.appConnCount[appID]++
	count := hub.appConnCount[appID]
	hub.mu.Unlock()

	// Invoke the observer outside the lock: SetRPCConnections takes Prometheus-internal
	// mutexes, and holding hub.mu across that would serialize readers (including Publish).
	hub.observeConnections(appID, count)

	return nil
}

// Get retrieves a connection by its unique connection ID.
// Returns the Connection instance if found, or nil if no connection
// with the specified ID exists in the hub.
//
// This method is safe for concurrent access.
func (hub *ConnectionHub) Get(connID string) Connection {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	conn, ok := hub.connections[connID]
	if !ok {
		return nil
	}

	return conn
}

// Remove unregisters a connection from the hub.
// This method:
//   - Removes the connection from the main connection map
//   - Cleans up any user-to-connection mappings
//   - Removes empty user entries to prevent memory leaks
//
// If the connection doesn't exist, this method does nothing (no-op).
// This method is safe for concurrent access.
func (hub *ConnectionHub) Remove(connID string) {
	hub.mu.Lock()

	conn, exists := hub.connections[connID]
	if !exists {
		hub.mu.Unlock()
		return // No connection to remove
	}
	delete(hub.connections, connID)

	appID := conn.ApplicationID()
	count, tracked := hub.appConnCount[appID]
	changed := false
	if tracked && count > 0 {
		hub.appConnCount[appID]--
		count = hub.appConnCount[appID]
		if count == 0 {
			delete(hub.appConnCount, appID)
		}
		changed = true
	}
	hub.mu.Unlock()

	// Only notify the observer when the bucket actually changed; otherwise we would
	// emit DeleteLabelValues for an app_id the gauge never tracked. Invoke outside
	// hub.mu to avoid serializing readers behind Prometheus-internal locks.
	if changed {
		hub.observeConnections(appID, count)
	}
}

// Publish broadcasts a message to all active connections for a specific user.
// This enables server-initiated notifications to be sent to all of a user's
// connected clients (e.g., multiple browser tabs or devices).
//
// The method:
//   - Looks up all connections associated with the user
//   - Attempts to send the message to each connection
//   - Silently skips any connections that fail to accept the message
//
// If the user has no active connections, the message is silently dropped.
// This method is safe for concurrent access.
// TODO: refine with subscription topics capability
func (hub *ConnectionHub) Publish(userID string, response []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	connIDs, ok := hub.authMapping[userID]
	if !ok {
		return
	}

	// Iterate over all connections for this user and send the message
	for connID := range connIDs {
		conn := hub.connections[connID]
		if conn == nil {
			continue // Skip if connection is nil or write sink is not set
		}

		// Write the response to the connection's write sink
		conn.WriteRawResponse(response)
	}
}

