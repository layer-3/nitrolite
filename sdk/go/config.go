package sdk

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

// Config holds the configuration options for the Clearnode client.
type Config struct {
	// URL is the WebSocket URL of the clearnode server
	URL string

	// HandshakeTimeout is the maximum time to wait for initial connection
	HandshakeTimeout time.Duration

	// PingTimeout is how long to wait for a ping from the server before considering the connection dead.
	// The server sends periodic pings to detect dead clients.
	PingTimeout time.Duration

	// ErrorHandler is called when connection errors occur
	ErrorHandler func(error)

	// BlockchainRPCs maps blockchain IDs to their RPC endpoints
	// Used by SDKClient for on-chain operations
	BlockchainRPCs map[uint64]string

	// ApplicationID is an advisory origin tag the client sends to the clearnode as
	// the "app_id" WebSocket query parameter. The clearnode stamps this value on
	// records produced by requests from this client. Empty means no tag.
	ApplicationID string
}

// Option is a functional option for configuring the Client.
type Option func(*Config)

// DefaultConfig returns the default configuration with sensible defaults.
var DefaultConfig = Config{
	HandshakeTimeout: 5 * time.Second,
	PingTimeout:      15 * time.Second,
	ErrorHandler:     defaultErrorHandler,
}

// appendApplicationIDQueryParam returns wsURL with the app_id query parameter set
// to applicationID. If applicationID is empty, wsURL is returned unchanged. Any
// existing app_id value is overwritten. Returns an error if wsURL cannot be parsed.
func appendApplicationIDQueryParam(wsURL, applicationID string) (string, error) {
	if applicationID == "" {
		return wsURL, nil
	}
	parsed, err := url.Parse(wsURL)
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	q.Set(rpc.ApplicationIDQueryParam, applicationID)
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

// defaultErrorHandler logs errors to stderr.
func defaultErrorHandler(err error) {
	if err != nil {
		log.New(os.Stderr, "[clearnode] ", log.LstdFlags).Printf("connection error: %v", err)
	}
}

// WithHandshakeTimeout sets the maximum time to wait for initial connection.
func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.HandshakeTimeout = d
	}
}

// WithPingTimeout sets how long to wait for a ping from the server before considering the connection dead.
func WithPingTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.PingTimeout = d
	}
}

// WithApplicationID sets the application ID sent to the clearnode as the
// "app_id" WebSocket query parameter. The clearnode treats this as an advisory
// origin tag on records produced by requests from this connection.
func WithApplicationID(appID string) Option {
	return func(c *Config) {
		c.ApplicationID = appID
	}
}

// WithErrorHandler sets a custom error handler for connection errors.
// The handler is called when the connection encounters an error or is closed.
func WithErrorHandler(fn func(error)) Option {
	return func(c *Config) {
		c.ErrorHandler = fn
	}
}
