package rpc_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestNewWebsocketConnection(t *testing.T) {
	t.Parallel()

	cfg := rpc.WebsocketConnectionConfig{}
	_, err := rpc.NewWebsocketConnection(cfg)
	require.Equal(t, "connection ID cannot be empty", err.Error())

	cfg.ConnectionID = "conn1"
	_, err = rpc.NewWebsocketConnection(cfg)
	require.Equal(t, "websocket connection cannot be nil", err.Error())

	cfg.WebsocketConn = &websocket.Conn{}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, cfg.ConnectionID, conn.ConnectionID())
	require.Equal(t, 64, cap(conn.RawRequests()))

	cfg.ProcessBufferSize = 20
	conn, err = rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, cfg.ConnectionID, conn.ConnectionID())
	require.Equal(t, cfg.ProcessBufferSize, cap(conn.RawRequests()))
}

func TestWebsocketConnection_Serve(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	wsConnMock := newGorillaWsConnMock(ctx)

	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:  "conn1",
		WebsocketConn: wsConnMock,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)

	var closureErr error
	var closureErrMu sync.Mutex
	handleClosure := func(err error) {
		closureErrMu.Lock()
		defer closureErrMu.Unlock()

		closureErr = err
	}
	conn.Serve(ctx, handleClosure)
	conn.Serve(ctx, handleClosure) // Second call should be no-op

	msg := "message1"
	wsConnMock.addMessageToRead(msg)

	select {
	case processedMsg := <-conn.RawRequests():
		require.Equal(t, msg, string(processedMsg))
	case <-time.After(100 * time.Millisecond):
		t.Fatal("message was not processed in time")
	}

	ok := conn.WriteRawResponse([]byte(msg))
	require.True(t, ok)
	time.Sleep(100 * time.Millisecond) // Allow some time for the write to complete

	lastWritten := wsConnMock.getLastWrittenMessage()
	require.Equal(t, msg, lastWritten)
	require.Equal(t, 1, wsConnMock.getCalledCloseCount())

	cancel() // Cancel the context to stop the connection
	time.Sleep(100 * time.Millisecond)

	closureErrMu.Lock()
	defer closureErrMu.Unlock()

	require.NoError(t, closureErr)
	require.Equal(t, 2, wsConnMock.getCalledCloseCount())
}

func TestWebsocketConnection_Serve_AppliesReadLimit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsConnMock := newGorillaWsConnMock(ctx)
	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:   "conn-readlimit",
		WebsocketConn:  wsConnMock,
		MaxMessageSize: 64 * 1024,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)

	conn.Serve(ctx, func(error) {})

	// Read limit must be set before any read happens.
	require.Eventually(t, func() bool {
		return wsConnMock.getReadLimit() == 64*1024
	}, 200*time.Millisecond, 5*time.Millisecond)
}

// TestWebsocketConnection_LocalReadLimit_GracefulClose simulates the local
// SetReadLimit hitting on an inbound frame. Gorilla returns ErrReadLimit to
// the application (and best-effort sends close 1009 to the peer over the wire).
// The connection must treat this as a graceful close, not an abnormal error.
func TestWebsocketConnection_LocalReadLimit_GracefulClose(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsConnMock := newGorillaWsConnMock(ctx)
	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:  "conn-readlimit",
		WebsocketConn: wsConnMock,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)

	closureCh := make(chan error, 1)
	conn.Serve(ctx, func(err error) { closureCh <- err })

	wsConnMock.readErrCh <- websocket.ErrReadLimit

	select {
	case err := <-closureCh:
		require.NoError(t, err, "ErrReadLimit must close gracefully, not as abnormal error")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("connection did not close after local read-limit hit")
	}
}

// TestWebsocketConnection_PeerInitiated1009_GracefulClose covers the symmetric
// case: the peer initiates a close with code 1009 (their read limit hit).
// Gorilla surfaces a *CloseError{1009} on ReadMessage in that case.
func TestWebsocketConnection_PeerInitiated1009_GracefulClose(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsConnMock := newGorillaWsConnMock(ctx)
	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:  "conn-peer1009",
		WebsocketConn: wsConnMock,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)

	closureCh := make(chan error, 1)
	conn.Serve(ctx, func(err error) { closureCh <- err })

	wsConnMock.readErrCh <- &websocket.CloseError{
		Code: websocket.CloseMessageTooBig,
		Text: "peer-side read limit exceeded",
	}

	select {
	case err := <-closureCh:
		require.NoError(t, err, "peer-sent 1009 must close gracefully, not as abnormal error")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("connection did not close after peer-initiated 1009")
	}
}

// stubLimiter rejects after the Nth call. Records every call for assertions.
type stubLimiter struct {
	rejectAt int
	calls    int
}

func (s *stubLimiter) Allow(_ time.Time, _ int) bool {
	s.calls++
	return s.calls < s.rejectAt
}

func TestWebsocketConnection_RateLimitedFrame_ClosesConnection(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsConnMock := newGorillaWsConnMock(ctx)
	limiter := &stubLimiter{rejectAt: 2} // first frame admits, second rejects
	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:     "conn-ratelimit",
		WebsocketConn:    wsConnMock,
		FrameRateLimiter: limiter,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)

	closureCh := make(chan error, 1)
	conn.Serve(ctx, func(err error) { closureCh <- err })

	// First frame admitted, drained by reader.
	wsConnMock.addMessageToRead("ok")
	select {
	case got := <-conn.RawRequests():
		require.Equal(t, "ok", string(got))
	case <-time.After(200 * time.Millisecond):
		t.Fatal("first frame not delivered")
	}

	// Second frame rejected by limiter → connection should close.
	wsConnMock.addMessageToRead("blocked")

	select {
	case err := <-closureCh:
		require.NoError(t, err, "rate-limit close is graceful")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("connection did not close after rate-limited frame")
	}
	require.Equal(t, 2, limiter.calls, "limiter was consulted for both frames")
}

func TestWebsocketConnection_ConnectionID(t *testing.T) {
	t.Parallel()

	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:  "conn1",
		WebsocketConn: &websocket.Conn{},
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)
	require.Equal(t, cfg.ConnectionID, conn.ConnectionID())
}

func TestWebsocketConnection_WriteRawResponse(t *testing.T) {
	t.Parallel()

	wsConnMock := newGorillaWsConnMock(context.Background())
	cfg := rpc.WebsocketConnectionConfig{
		ConnectionID:    "conn1",
		WebsocketConn:   wsConnMock,
		WriteBufferSize: 1,
		WriteTimeout:    100 * time.Millisecond,
	}
	conn, err := rpc.NewWebsocketConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.True(t, conn.WriteRawResponse([]byte("msg1")))
	require.False(t, conn.WriteRawResponse([]byte("msg2"))) // This should block until the first message is sent
}

type gorillaWsConnMock struct {
	ctx                context.Context
	messageToReadCh    chan []byte
	readErrCh          chan error
	lastWrittenMessage []byte
	calledCloseCount   int
	readLimit          int64

	mu sync.Mutex
}

func newGorillaWsConnMock(ctx context.Context) *gorillaWsConnMock {
	return &gorillaWsConnMock{
		ctx:             ctx,
		messageToReadCh: make(chan []byte, 1),
		readErrCh:       make(chan error, 1),
	}
}

func (m *gorillaWsConnMock) ReadMessage() (messageType int, p []byte, err error) {
	select {
	case <-m.ctx.Done():
		return 0, nil, &websocket.CloseError{
			Code: websocket.CloseNormalClosure,
			Text: "context cancelled",
		}
	case err := <-m.readErrCh:
		return 0, nil, err
	case msg := <-m.messageToReadCh:
		// Simulate reading a message
		return websocket.TextMessage, msg, nil
	}
}

func (m *gorillaWsConnMock) NextWriter(messageType int) (io.WriteCloser, error) {
	return m, nil
}

func (m *gorillaWsConnMock) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastWrittenMessage = p
	return len(p), nil
}

func (m *gorillaWsConnMock) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calledCloseCount++
	return nil
}

func (m *gorillaWsConnMock) addMessageToRead(msg string) {
	m.messageToReadCh <- []byte(msg)
}

func (m *gorillaWsConnMock) getLastWrittenMessage() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return string(m.lastWrittenMessage)
}

func (m *gorillaWsConnMock) getCalledCloseCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calledCloseCount
}

func (m *gorillaWsConnMock) WriteControl(messageType int, data []byte, deadline time.Time) error {
	// No-op for mock
	return nil
}

func (m *gorillaWsConnMock) SetPongHandler(h func(appData string) error) {
	// No-op for mock
}

func (m *gorillaWsConnMock) SetReadDeadline(t time.Time) error {
	// No-op for mock
	return nil
}

func (m *gorillaWsConnMock) SetReadLimit(limit int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.readLimit = limit
}

func (m *gorillaWsConnMock) getReadLimit() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.readLimit
}
