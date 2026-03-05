package sdk

import (
	"context"
	"fmt"
	"sync"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

type MockDialer struct {
	mu        sync.Mutex
	responses map[string]interface{}
	connected bool
	eventCh   chan *rpc.Message
}

func NewMockDialer() *MockDialer {
	return &MockDialer{
		responses: make(map[string]interface{}),
		eventCh:   make(chan *rpc.Message, 10),
	}
}

func (m *MockDialer) Dial(ctx context.Context, url string, handleClosure func(err error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	return nil
}

func (m *MockDialer) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *MockDialer) Call(ctx context.Context, req *rpc.Message) (*rpc.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}

	respData, ok := m.responses[req.Method]
	if !ok {
		return nil, fmt.Errorf("no mock response for method %s", req.Method)
	}

	// Create payload from response data
	payload, err := rpc.NewPayload(respData)
	if err != nil {
		return nil, err
	}

	return &rpc.Message{
		Type:      rpc.MsgTypeResp,
		RequestID: req.RequestID,
		Method:    req.Method,
		Payload:   payload,
	}, nil
}

func (m *MockDialer) EventCh() <-chan *rpc.Message {
	return m.eventCh
}

func (m *MockDialer) RegisterResponse(method string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method] = response
}

func (m *MockDialer) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	close(m.eventCh)
}
