package rpc_test

import (
	"context"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

// MockCallHandler is a function type that handles RPC calls in the mock dialer.
// It receives the request parameters and a notification publisher function,
// and returns response.
type MockCallHandler func(params rpc.Payload, publishNotification MockNotificationPublisher) (*rpc.Message, error)

// MockNotificationPublisher is a function type that allows handlers to publish
// asynchronous notifications to the client.
type MockNotificationPublisher func(event rpc.Event, notification rpc.Payload)

// Ensure MockDialer implements the Dialer interface
var _ rpc.Dialer = (*MockDialer)(nil)

// MockDialer is a test implementation of the rpc.Dialer interface.
// It allows registering handlers for specific RPC methods and simulating
// server responses and notifications without actual network connections.
type MockDialer struct {
	// handlers maps RPC methods to their mock handlers
	handlers map[rpc.Method]MockCallHandler
	// eventCh is the channel for publishing notifications to the client
	eventCh chan *rpc.Message
}

// NewMockDialer creates a new mock dialer for testing.
// The event channel is buffered to prevent blocking when publishing notifications.
func NewMockDialer() *MockDialer {
	return &MockDialer{
		handlers: make(map[rpc.Method]MockCallHandler),
		eventCh:  make(chan *rpc.Message, 10),
	}
}

// RegisterHandler registers a mock handler for a specific RPC method.
// The handler will be called when the client makes a request to that method.
func (d *MockDialer) RegisterHandler(method rpc.Method, handler MockCallHandler) {
	d.handlers[method] = handler
}

// Dial is a no-op for the mock dialer since no actual connection is made.
// It simulates an always-connected state.
func (d *MockDialer) Dial(ctx context.Context, url string, handleClosure func(err error)) error {
	// No-op for mock dialer - always connected
	return nil
}

// IsConnected always returns true for the mock dialer.
func (d *MockDialer) IsConnected() bool {
	return true
}

// Call handles RPC calls by routing them to registered mock handlers.
// If no handler is registered for the method, it returns a "method not found" error.
// Handler errors are converted to RPC error responses.
func (d *MockDialer) Call(ctx context.Context, req *rpc.Message) (*rpc.Message, error) {
	if req == nil {
		return nil, rpc.ErrNilRequest
	}

	// Find the handler for this method
	handler, exists := d.handlers[rpc.Method(req.Method)]
	if !exists {
		res := rpc.NewErrorResponse(req.RequestID, req.Method, "method not found")
		return &res, nil
	}

	// Call the handler with a notification publisher
	res, err := handler(req.Payload, d.publishNotification)
	if err != nil {
		res := rpc.NewErrorResponse(req.RequestID, req.Method, err.Error())
		return &res, nil
	}

	return res, nil
}

// EventCh returns the channel for receiving notifications.
func (d *MockDialer) EventCh() <-chan *rpc.Message {
	return d.eventCh
}

// publishNotification is a helper method that handlers can use to send
// asynchronous notifications to the client through the event channel.
func (d *MockDialer) publishNotification(event rpc.Event, notification rpc.Payload) {
	res := rpc.NewEvent(0, string(event), notification)

	// Non-blocking send to prevent handlers from getting stuck
	select {
	case d.eventCh <- &res:
		// Notification sent successfully
	default:
		// Channel full, notification dropped
		// In a real implementation, you might want to log this
	}
}

// CloseEventChannel closes the event channel to simulate connection closure.
// This is useful for testing client behavior when the connection is lost.
func (d *MockDialer) CloseEventChannel() {
	close(d.eventCh)
}
