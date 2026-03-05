package rpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/layer-3/nitrolite/pkg/log"
)

// Default values are carefully chosen to balance resource consumption and operational flexibility.
// Higher values would increase node resource usage, while lower values would reduce the ability
// to handle traffic spikes and incidents. These defaults provide a reasonable middle ground.
var (
	// defaultWsConnWriteTimeout is the default maximum duration to wait for a write to complete.
	defaultWsConnWriteTimeout = 5 * time.Second
	// defaultWsConnProcessBufferSize is the default size of the buffer for processing incoming messages.
	defaultWsConnProcessBufferSize = 64
	// defaultWsConnWriteBufferSize is the default size of the buffer for outgoing messages.
	defaultWsConnWriteBufferSize = 64
	// defaultWsConnPingInterval is the default interval for sending WebSocket ping frames to clients.
	defaultWsConnPingInterval = 5 * time.Second
	// defaultWsConnPongTimeout is the default timeout for receiving pong responses from clients.
	// If no pong is received within this duration after a ping, the connection is considered dead.
	defaultWsConnPongTimeout = 10 * time.Second
)

// Connection represents an active RPC connection that handles bidirectional communication.
// Implementations of this interface manage the connection lifecycle and message routing.
// The interface is designed to be transport-agnostic, though the primary implementation uses WebSocket.
type Connection interface {
	// ConnectionID returns the unique identifier for this connection.
	// This ID is generated when the connection is established and remains
	// constant throughout the connection's lifetime.
	ConnectionID() string

	// Origin returns the origin of the connection, such as the client's IP address or other identifying information.
	Origin() string

	// RawRequests returns a read-only channel for receiving incoming raw request messages.
	// Messages received on this channel are raw bytes that need to be unmarshaled
	// into Request objects for processing. The channel is closed when the
	// connection is terminated.
	RawRequests() <-chan []byte

	// WriteRawResponse attempts to send a raw response message to the client.
	// The method returns true if the message was successfully queued for sending,
	// or false if the operation timed out (indicating a potentially unresponsive client).
	// Messages that fail to send may trigger connection closure.
	WriteRawResponse(message []byte) bool

	// Serve starts the connection's lifecycle by spawning goroutines for reading and writing.
	// This method returns immediately after starting the goroutines. The handleClosure
	// callback will be invoked asynchronously when the connection terminates (with an
	// error if abnormal termination occurred). The parentCtx parameter enables
	// graceful shutdown of the connection.
	Serve(parentCtx context.Context, handleClosure func(error))
}

// GorillaWsConnectionAdapter abstracts the methods of a WebSocket connection needed by WebsocketConnection.
type GorillaWsConnectionAdapter interface {
	// ReadMessage reads a message from the WebSocket connection.
	ReadMessage() (messageType int, p []byte, err error)
	// NextWriter returns a writer for the next message to be sent on the WebSocket connection.
	NextWriter(messageType int) (io.WriteCloser, error)
	// Close closes the WebSocket connection.
	Close() error
	// WriteControl writes a control message (ping, pong, close) to the connection.
	WriteControl(messageType int, data []byte, deadline time.Time) error
	// SetPongHandler sets the handler for pong messages received from the peer.
	SetPongHandler(h func(appData string) error)
	// SetReadDeadline sets the deadline for future Read calls.
	// A zero value means reads will not time out.
	SetReadDeadline(t time.Time) error
}

// WebsocketConnection implements the Connection interface using WebSocket transport.
// It manages bidirectional communication, and provides thread-safe operations
// for concurrent message processing. The connection supports graceful shutdown
// and automatic cleanup of resources.
//
// Key features:
//   - Concurrent read/write operations with separate goroutines
//   - Configurable timeouts for write operations
//   - Buffered channels for message processing
//   - Thread-safe user authentication state management
//   - Graceful connection closure with proper resource cleanup
//   - Native WebSocket ping/pong keepalive detection
type WebsocketConnection struct {
	// ctx is the parent context for managing goroutines
	ctx context.Context
	// connectionID is a unique identifier for this connection
	connectionID string
	// origin is the origin of the connection, such as the client's IP address
	origin string
	// websocketConn is the underlying WebSocket connection
	websocketConn GorillaWsConnectionAdapter
	// writeTimeout is the maximum duration to wait for a write to complete
	writeTimeout time.Duration
	// pingInterval is how often to send ping frames to clients
	pingInterval time.Duration
	// pongTimeout is the maximum duration to wait for a pong response from the client
	pongTimeout time.Duration

	// logger is used for logging events related to this connection
	logger log.Logger
	// onMessageSentHandler is called when a message is sent
	onMessageSentHandler func([]byte)
	// writeSink is the channel for sending messages to this connection
	writeSink chan []byte
	// processSink is the channel for processing incoming messages
	processSink chan []byte
	// closeConnCh is a channel that can be used to signal connection closure
	closeConnCh chan struct{}

	// mu is a mutex to protect access to user-related data
	mu sync.RWMutex
}

// WebsocketConnectionConfig contains configuration options for creating a new WebsocketConnection.
// All fields except ConnectionID and WebsocketConn have sensible defaults.
type WebsocketConnectionConfig struct {
	// ConnectionID is the unique identifier for this connection (required)
	ConnectionID string
	// Origin is the origin of the connection, such as the client's IP address (optional)
	Origin string
	// WebsocketConn is the underlying WebSocket connection (required)
	WebsocketConn GorillaWsConnectionAdapter

	// WriteTimeout is the maximum duration to wait for a write operation (default: 5s)
	WriteTimeout time.Duration
	// WriteBufferSize is the capacity of the outgoing message buffer (default: 10)
	WriteBufferSize int
	// ProcessBufferSize is the capacity of the incoming message buffer (default: 10)
	ProcessBufferSize int
	// PingInterval is how often to send ping frames to clients (default: 5s).
	PingInterval time.Duration
	// PongTimeout is the maximum duration to wait for a pong response from the client (default: 10s).
	// If no pong is received within this duration, the connection is considered dead.
	PongTimeout time.Duration
	// Logger for connection events (default: no-op logger)
	Logger log.Logger
	// OnMessageSentHandler is called after a message is successfully sent (optional)
	OnMessageSentHandler func([]byte)
}

// NewWebsocketConnection creates a new WebsocketConnection instance with the provided configuration.
// Returns an error if required fields (ConnectionID, WebsocketConn) are missing.
// Optional fields are set to sensible defaults if not provided.
func NewWebsocketConnection(config WebsocketConnectionConfig) (*WebsocketConnection, error) {
	if config.ConnectionID == "" {
		return nil, fmt.Errorf("connection ID cannot be empty")
	}
	if config.Origin == "" {
		config.Origin = "unknown"
	}
	if config.WebsocketConn == nil {
		return nil, fmt.Errorf("websocket connection cannot be nil")
	}
	if config.Logger == nil {
		config.Logger = log.NewNoopLogger()
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = defaultWsConnWriteTimeout
	}
	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = defaultWsConnWriteBufferSize
	}
	if config.ProcessBufferSize <= 0 {
		config.ProcessBufferSize = defaultWsConnProcessBufferSize
	}
	if config.PingInterval <= 0 {
		config.PingInterval = defaultWsConnPingInterval
	}
	if config.PongTimeout <= 0 {
		config.PongTimeout = defaultWsConnPongTimeout
	}
	if config.OnMessageSentHandler == nil {
		config.OnMessageSentHandler = func([]byte) {}
	}

	return &WebsocketConnection{
		connectionID:  config.ConnectionID,
		origin:        config.Origin,
		websocketConn: config.WebsocketConn,
		writeTimeout:  config.WriteTimeout,
		pingInterval:  config.PingInterval,
		pongTimeout:   config.PongTimeout,

		logger:               config.Logger.WithKV("connectionID", config.ConnectionID),
		onMessageSentHandler: config.OnMessageSentHandler,
		writeSink:            make(chan []byte, config.WriteBufferSize),
		processSink:          make(chan []byte, config.ProcessBufferSize),
		closeConnCh:          make(chan struct{}, 1),
	}, nil
}

// Serve starts the connection's lifecycle by spawning concurrent goroutines.
// This method:
//   - Spawns three goroutines: one for reading messages, one for writing messages,
//     and one for monitoring connection closure signals
//   - Returns immediately after starting the goroutines
//   - Spawns an additional goroutine that waits for all operations to complete
//     and then invokes handleClosure with any error that occurred
//
// The handleClosure callback is guaranteed to be called exactly once when the
// connection terminates. The method is idempotent - calling it multiple times
// will immediately invoke handleClosure without starting duplicate goroutines.
func (conn *WebsocketConnection) Serve(parentCtx context.Context, handleClosure func(error)) {
	conn.mu.Lock()
	if conn.ctx != nil {
		conn.mu.Unlock()
		handleClosure(nil) // Connection is already running
		return
	}
	conn.ctx = parentCtx
	conn.mu.Unlock()

	// Set up pong handler to refresh read deadline when pong is received from client.
	// This enables detection of dead connections - if no pong arrives within the timeout
	// after sending a ping, the read will fail and the connection will be closed.
	conn.websocketConn.SetPongHandler(func(appData string) error {
		// Refresh read deadline on each pong received
		return conn.websocketConn.SetReadDeadline(time.Now().Add(conn.pongTimeout))
	})
	// Set initial read deadline
	_ = conn.websocketConn.SetReadDeadline(time.Now().Add(conn.pongTimeout))

	// Create a child context that can be cancelled to stop all goroutines
	childCtx, cancel := context.WithCancel(parentCtx)
	wg := &sync.WaitGroup{}
	wg.Add(3)

	var closureErr error
	var closureErrMu sync.Mutex
	childHandleClosure := func(err error) {
		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		// Capture the first error encountered
		if err != nil && closureErr == nil {
			closureErr = err
		}

		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	// Start reading messages from the WebSocket connection
	go conn.readMessages(childHandleClosure)

	// Start writing messages to the WebSocket connection
	go conn.writeMessages(childCtx, childHandleClosure)

	// Wait for the WebSocket connection to close
	go conn.waitForConnClose(childCtx, childHandleClosure)

	go func() {
		// Wait for all goroutines to finish
		wg.Wait()

		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		// Invoke the closure handler with any error that occurred
		handleClosure(closureErr)

		// Close the WebSocket connection
		if err := conn.websocketConn.Close(); err != nil {
			conn.logger.Error("error closing WebSocket connection", "error", err)
		}
	}()
}

// ConnectionID returns the unique identifier for this connection.
func (conn *WebsocketConnection) ConnectionID() string {
	return conn.connectionID
}

// Origin returns the origin of the connection, such as the client's IP address or other identifying information.
func (conn *WebsocketConnection) Origin() string {
	return conn.origin
}

// RawRequests returns the channel for processing incoming requests.
func (conn *WebsocketConnection) RawRequests() <-chan []byte {
	return conn.processSink
}

// WriteRawResponse attempts to queue a message for sending to the client.
// The method uses a timeout to prevent blocking on unresponsive clients.
// If the message cannot be queued within the timeout duration:
//   - The method returns false
//   - A close signal is sent to trigger connection shutdown
//   - This prevents resource exhaustion from slow or disconnected clients
//
// Returns true if the message was successfully queued for sending.
func (conn *WebsocketConnection) WriteRawResponse(message []byte) bool {
	timer := time.NewTimer(conn.writeTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		select {
		case conn.closeConnCh <- struct{}{}:
		default:
		}
		return false
	case conn.writeSink <- message:
		return true
	}
}

// readMessages listens for incoming messages on the WebSocket connection.
// It reads messages and sends them to the processSink channel for further processing.
func (conn *WebsocketConnection) readMessages(handleClosure func(error)) {
	defer close(conn.processSink) // Close the processing channel when done

	for {
		_, messageBytes, err := conn.websocketConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				conn.logger.Error("WebSocket connection closed with unexpected reason", "error", err)
				handleClosure(err)
			} else {
				handleClosure(nil) // Normal closure
			}
			return
		}

		if len(messageBytes) == 0 {
			conn.logger.Debug("received empty message, skipping")
			continue // Skip empty messages
		}

		select {
		case conn.processSink <- messageBytes:
			// ok
		default:
			conn.logger.Warn("incoming queue full; closing connection",
				"queue_len", len(conn.processSink),
				"queue_cap", cap(conn.processSink),
			)
			select {
			case conn.closeConnCh <- struct{}{}:
			default:
			}
			return
		}
	}
}

// writeMessages handles outgoing messages to the WebSocket connection.
// It reads from the message sink channel, writes to the WebSocket,
// and sends periodic ping frames to detect dead connections.
func (conn *WebsocketConnection) writeMessages(ctx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines

	pingTicker := time.NewTicker(conn.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			conn.logger.Debug("context done, stopping message writing")
			return
		case <-pingTicker.C:
			// Send native WebSocket ping frame
			if err := conn.websocketConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(conn.writeTimeout)); err != nil {
				conn.logger.Error("error sending ping", "error", err)
				return
			}
		case messageBytes, ok := <-conn.writeSink:
			if !ok {
				return // Channel closed, stop writing
			}
			if len(messageBytes) == 0 {
				continue // Skip empty messages
			}

			w, err := conn.websocketConn.NextWriter(websocket.TextMessage)
			if err != nil {
				conn.logger.Error("error getting writer for response", "error", err)
				continue
			}

			if _, err := w.Write(messageBytes); err != nil {
				conn.logger.Error("error writing response", "error", err)
				w.Close()
				continue
			}

			if err := w.Close(); err != nil {
				conn.logger.Error("error closing writer for response", "error", err)
				continue
			}

			conn.onMessageSentHandler(messageBytes)
		}
	}
}

// waitForConnClose waits for the WebSocket connection to close.
// It listens for the close signal and logs the closure event.
func (conn *WebsocketConnection) waitForConnClose(ctx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines when done

	select {
	case <-ctx.Done():
		conn.logger.Debug("context done, stopping connection close wait")
	case <-conn.closeConnCh:
		conn.logger.Info("WebSocket connection closed by server", "connectionID", conn.ConnectionID())
	}
}
