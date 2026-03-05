package log_test

import (
	"testing"

	"github.com/layer-3/nitrolite/pkg/log"
	"github.com/stretchr/testify/assert"
)

// TestSpanLogger tests the SpanLogger implementation that bridges logging with span recording.
// It verifies:
// 1. Log messages are forwarded to both the wrapped logger and span event recorder
// 2. Trace and span IDs are automatically added to log entries
// 3. Error and Fatal levels are recorded as errors in the span
// 4. Logger attributes and configuration are properly propagated
// 5. Caller skip is correctly adjusted for the wrapper
func TestSpanLogger(t *testing.T) {
	// Create mocks for testing
	mockLogger := NewMockLogger()
	mockSer := NewMockSpanEventRecorder("trace-id-123", "span-id-456")
	logger := log.NewSpanLogger(mockLogger, mockSer)
	// Verify caller skip is incremented to account for the wrapper
	assert.Equal(t, 1, mockLogger.CallerSkip())

	// Helper function to convert key-value slice to map for easier assertion
	kvSliceToMap := func(kv []any) map[string]any {
		kvMap := make(map[string]any)
		for i := 0; i < len(kv); i += 2 {
			key, ok := kv[i].(string)
			if !ok || i+1 >= len(kv) {
				continue
			}
			kvMap[key] = kv[i+1]
		}
		return kvMap
	}

	// Helper function to assert both logger and span recorder received correct data
	assertEntry := func(
		t *testing.T,
		expectedLevel log.Level,
		expectedName, expectedMsg string,
		expectedKeysAndValues []any,
	) {
		// Verify the wrapped logger received the correct data
		mockEntry := mockLogger.LastEntry()
		assert.Equal(t, expectedLevel, mockEntry.Level)
		assert.Equal(t, expectedMsg, mockEntry.Message)

		// Verify trace and span IDs are automatically added
		expectedKVMap := kvSliceToMap(expectedKeysAndValues)
		actualKVMap := kvSliceToMap(mockEntry.KeysAndValues)
		for k, v := range expectedKVMap {
			assert.Equal(t, v, actualKVMap[k])
		}
		assert.Equal(t, len(expectedKVMap)+2, len(actualKVMap)) // +2 for traceId and spanId
		assert.Equal(t, mockSer.TraceID(), actualKVMap["traceId"])
		assert.Equal(t, mockSer.SpanID(), actualKVMap["spanId"])

		// Verify error levels are recorded as errors in the span
		shouldHaveError := expectedLevel == log.LevelError || expectedLevel == log.LevelFatal
		assert.Equal(t, shouldHaveError, mockSer.HasError(), "SpanEventRecorder HasError() mismatch")

		// Verify span recorder received correct metadata
		actualKVMap = kvSliceToMap(mockSer.LastEventMetadata())
		for k, v := range expectedKVMap {
			assert.Equal(t, v, actualKVMap[k])
		}
		assert.Equal(t, len(expectedKVMap)+3, len(actualKVMap)) // +3 for level, msg and component
		assert.Equal(t, string(expectedLevel), actualKVMap["level"])
		assert.Equal(t, expectedMsg, actualKVMap["msg"])
		assert.Equal(t, expectedName, actualKVMap["component"])
	}

	// Test basic logging with all levels
	testName := "testLogger"
	logger = logger.WithName(testName)

	keysAndValues := []any{"key1", "value1", "key2", "value2"}
	testMessage := "test message"

	logger.Debug(testMessage, keysAndValues...)
	assertEntry(t, log.LevelDebug, testName, testMessage, keysAndValues)

	logger.Info(testMessage, keysAndValues...)
	assertEntry(t, log.LevelInfo, testName, testMessage, keysAndValues)

	logger.Warn(testMessage, keysAndValues...)
	assertEntry(t, log.LevelWarn, testName, testMessage, keysAndValues)

	logger.Error(testMessage, keysAndValues...)
	assertEntry(t, log.LevelError, testName, testMessage, keysAndValues)

	// Test logger naming
	testSubsystem := "testSubsystem"
	logger = logger.WithName(testSubsystem)
	assert.Equal(t, testSubsystem, logger.Name())

	// Test key-value propagation
	newK := "newKey"
	newV := "newValue"
	newPair := []any{newK, newV}
	logger = logger.WithKV(newK, newV)
	assert.Equal(t, newPair, logger.GetAllKV())
	allKeysAndValues := append(newPair, keysAndValues...)

	// Test AddCallerSkip functionality
	wrapperWithLoggerInfo := func(msg string, keysAndValues ...any) {
		logger.AddCallerSkip(1).Error(msg, keysAndValues...)
	}

	wrapperWithLoggerInfo(testMessage, keysAndValues...)
	assertEntry(t, log.LevelError, testSubsystem, testMessage, allKeysAndValues)
	assert.Equal(t, 2, mockLogger.CallerSkip())
}
