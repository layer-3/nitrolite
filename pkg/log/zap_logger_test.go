package log_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/log"
)

// TestZapLogger comprehensively tests the ZapLogger implementation.
// It verifies:
// 1. Correct log level output (Debug, Info, Warn, Error)
// 2. Logger naming hierarchy with WithName
// 3. Key-value pair propagation with WithKV
// 4. Caller information accuracy
// 5. AddCallerSkip functionality for wrapper functions
func TestZapLogger(t *testing.T) {
	// Create a test logger with JSON format for easier parsing
	cfg := log.Config{
		Format: "json",
		Level:  log.LevelDebug,
	}
	tws := &testWriteSyncer{}
	logger := log.NewZapLogger(cfg, tws)

	// Test basic logging functionality with different levels
	testName := "testLogger"
	logger = logger.WithName(testName)

	keysAndValues := []any{"key1", "value1", "key2", "value2"}
	testMessage := "test message"
	expectedCallerFilePath := "log/zap_logger_test.go"

	logger.Debug(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelDebug, testName, testMessage, expectedCallerFilePath, 38, keysAndValues...)

	logger.Info(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelInfo, testName, testMessage, expectedCallerFilePath, 41, keysAndValues...)

	logger.Warn(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelWarn, testName, testMessage, expectedCallerFilePath, 44, keysAndValues...)

	logger.Error(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelError, testName, testMessage, expectedCallerFilePath, 47, keysAndValues...)

	// Test logger naming hierarchy
	testSubsystem := "testSubsystem"
	newExpectedName := fmt.Sprintf("%s.%s", testName, testSubsystem)
	logger = logger.WithName(testSubsystem)
	assert.Equal(t, newExpectedName, logger.Name())

	// Test key-value pair propagation
	newK := "newKey"
	newV := "newValue"
	newPair := []any{newK, newV}
	logger = logger.WithKV(newK, newV)
	assert.Equal(t, newPair, logger.GetAllKV())
	allKeysAndValues := append(newPair, keysAndValues...)

	logger.Debug(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelDebug, newExpectedName, testMessage, expectedCallerFilePath, 64, allKeysAndValues...)

	logger.Info(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelInfo, newExpectedName, testMessage, expectedCallerFilePath, 67, allKeysAndValues...)

	logger.Warn(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelWarn, newExpectedName, testMessage, expectedCallerFilePath, 70, allKeysAndValues...)

	logger.Error(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelError, newExpectedName, testMessage, expectedCallerFilePath, 73, allKeysAndValues...)

	// Test AddCallerSkip functionality for wrapper functions
	wrapperWithLoggerInfo := func(msg string, keysAndValues ...any) {
		logger.AddCallerSkip(1).Info(msg, keysAndValues...)
	}

	wrapperWithLoggerInfo(testMessage, keysAndValues...)
	tws.AssertEntry(t, log.LevelInfo, newExpectedName, testMessage, expectedCallerFilePath, 81, allKeysAndValues...)
}

// testWriteSyncer is a mock zapcore.WriteSyncer that captures the last written log entry.
// It's used to verify the exact output of the ZapLogger in tests.
type testWriteSyncer struct {
	lastEntry []byte
}

// Write captures the log entry for later assertion.
func (tws *testWriteSyncer) Write(p []byte) (n int, err error) {
	tws.lastEntry = p
	return len(p), nil
}

// Sync is a no-op for this test implementation.
func (tws *testWriteSyncer) Sync() error {
	return nil
}

// AssertEntry verifies that the last written log entry matches expected values.
// It checks the log level, logger name, message, caller information, and all key-value pairs.
func (tws *testWriteSyncer) AssertEntry(t *testing.T, level log.Level, name, message, callerFilePath string, callerLineNum int, keysAndValues ...any) {
	entryMap := make(map[string]any)
	require.NoError(t, json.Unmarshal(tws.lastEntry, &entryMap), "Failed to unmarshal log entry: %s", string(tws.lastEntry))

	assert.Contains(t, entryMap, "ts")
	assert.Equal(t, name, entryMap["logger"])
	assert.Equal(t, string(level), entryMap["level"])
	assert.Equal(t, message, entryMap["msg"])
	assert.Equal(t, fmt.Sprintf("%s:%d", callerFilePath, callerLineNum), entryMap["caller"].(string))

	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		value := keysAndValues[i+1]
		assert.Equal(t, value, entryMap[key.(string)])
	}

	assert.Equal(t, len(keysAndValues)/2, len(entryMap)-5) // -5 for ts, level, logger, caller and msg
}
