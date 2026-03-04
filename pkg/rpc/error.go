package rpc

import (
	"encoding/json"
	"fmt"
)

const (
	// errorParamKey is the standard key used in Params to store error messages.
	// When a Payload contains an error, it will be stored under this key.
	errorParamKey = "error"
)

// Dialer error messages
var (
	// Connection errors
	ErrAlreadyConnected  = fmt.Errorf("already connected")
	ErrNotConnected      = fmt.Errorf("not connected to server")
	ErrConnectionTimeout = fmt.Errorf("websocket connection timeout")
	ErrReadingMessage    = fmt.Errorf("error reading message")

	// Request/Response errors
	ErrNilRequest           = fmt.Errorf("nil request")
	ErrInvalidRequestMethod = fmt.Errorf("invalid request method")
	ErrMarshalingRequest    = fmt.Errorf("error marshaling request")
	ErrSendingRequest       = fmt.Errorf("error sending request")
	ErrNoResponse           = fmt.Errorf("no response received")
	ErrSendingPing          = fmt.Errorf("error sending ping")

	// WebSocket-specific errors
	ErrDialingWebsocket = fmt.Errorf("error dialing websocket server")
)

// Error represents an error in the RPC protocol that should be sent back to the client
// in the RPC response. Unlike generic errors, Error messages are guaranteed to be
// included in the error response sent to the client.
//
// Use Error when you want to provide specific, user-facing error messages in RPC responses.
// For internal errors that should not be exposed to clients, use regular errors instead.
//
// Example:
//
//	// Client will receive this exact error message
//	return rpc.Errorf("invalid wallet address: %s", addr)
//
//	// Client will receive a generic error message
//	return fmt.Errorf("database connection failed")
type Error struct {
	err error
}

// NewError creates a new Error from an existing error. The message from the provided error will be included in the RPC response sent to the client.
//
// Use this function when you have an existing error that contains a user-friendly message
// that you want to send back to the client. If the original error message is not suitable
// for client consumption, consider using Errorf to create a new Error with a custom message.
//
// Example:
//
//	// Creating an Error from an existing error
//	if err := validateAddress(addr); err != nil {
//	    return rpc.NewError(err)
//	}
//
//	// This will send the original error message to the client, so ensure it's appropriate for exposure.
func NewError(err error) Error {
	return Error{err: err}
}

// Errorf creates a new Error with a formatted error message that will be sent
// to the client in the RPC response. This is the preferred way to create client-facing
// errors in RPC handlers.
//
// The error message should be clear, actionable, and safe to expose to external clients.
// Avoid including sensitive information like internal system details, file paths, or
// database specifics.
//
// Usage in RPC handlers:
//
//	// In a handler function that returns an error
//	if amount.IsNegative() {
//		return Errorf("invalid amount: cannot be negative")
//	}
//
//	// With formatting for specific details
//	if balance.LessThan(amount) {
//		return Errorf("insufficient balance: need %s but have %s", amount, balance)
//	}
func Errorf(format string, args ...any) Error {
	return Error{
		err: fmt.Errorf(format, args...),
	}
}

// Error implements the error interface for Error.
// It returns the underlying error message that will be sent to clients.
//
// This method allows Error to be used anywhere a standard Go error is expected,
// while maintaining the distinction that this error's message is safe for
// external client consumption.
func (e Error) Error() string {
	return e.err.Error()
}

// NewErrorPayload creates a Params map containing an error message.
// This is a convenience function for creating standardized error parameters
// that follow the protocol's error response convention.
//
// The error message is stored under the "error" key and properly JSON-encoded.
//
// Example usage:
//
//	// Creating error params for a validation failure
//	errParams := NewErrorPayload("invalid address format")
//	payload := NewPayload(requestID, method, errParams)
//
//	// Creating error params from an existing error
//	if err := validateAmount(amount); err != nil {
//	    errParams := NewErrorPayload(err.Error())
//	    response := NewResponse(NewPayload(id, method, errParams))
//	}
//
// The resulting Params will contain: {"error": "invalid address format"}
func NewErrorPayload(errMsg string) Payload {
	return Payload{errorParamKey: json.RawMessage(fmt.Sprintf(`"%s"`, errMsg))}
}
