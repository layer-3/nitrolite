package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/layer-3/nitrolite/pkg/log"
)

// TestContextLogger tests the context-based logger functionality.
// It verifies:
// 1. Default behavior returns a NoopLogger when no logger is set in context
// 2. SetContextLogger properly stores a logger in the context
// 3. FromContext retrieves the correct logger type
// 4. When a valid span is present in context, SetContextLogger wraps the logger with SpanLogger
func TestContextLogger(t *testing.T) {
	ctx := context.Background()

	// Test 1: When no logger is in context, FromContext returns a NoopLogger
	logger := log.FromContext(ctx)
	assert.NotNil(t, logger)

	_, isNoop := logger.(log.NoopLogger)
	assert.True(t, isNoop)

	// Test 2: SetContextLogger stores a logger that can be retrieved by FromContext
	cfg := log.Config{}
	logger = log.NewZapLogger(cfg)
	ctx = log.SetContextLogger(ctx, logger)

	logger = log.FromContext(ctx)
	assert.NotNil(t, logger)

	_, isZapLogger := logger.(*log.ZapLogger)
	assert.True(t, isZapLogger)

	// Test 3: When a valid span is in context, SetContextLogger wraps the logger with SpanLogger
	ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: [16]byte{1},
		SpanID:  [8]byte{1},
	}))
	ctx = log.SetContextLogger(ctx, logger)

	logger = log.FromContext(ctx)
	assert.NotNil(t, logger)

	_, isSpanLogger := logger.(*log.SpanLogger)
	assert.True(t, isSpanLogger)
}
