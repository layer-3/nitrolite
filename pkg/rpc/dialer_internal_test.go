package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebsocketDialer_Call_MarshalFailureNoSinkLeak verifies that a request
// which fails to marshal does not leave a dangling entry in responseSinks.
// Regression test: marshalling must happen before the response sink is
// registered, otherwise repeated malformed calls grow the map unboundedly.
func TestWebsocketDialer_Call_MarshalFailureNoSinkLeak(t *testing.T) {
	t.Parallel()

	d := NewWebsocketDialer(DefaultWebsocketDialerConfig)

	// Malformed raw JSON in the payload makes json.Marshal fail.
	badPayload := Payload{"bad": json.RawMessage("{invalid")}

	const calls = 100
	for i := uint64(1); i <= calls; i++ {
		req := NewRequest(i, "noop", badPayload)
		_, err := d.Call(context.Background(), &req)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrMarshalingRequest), "expected ErrMarshalingRequest, got %v", err)
	}

	d.mu.RLock()
	pending := len(d.responseSinks)
	d.mu.RUnlock()
	assert.Equal(t, 0, pending, "marshal failures must not leak response sinks")
}
