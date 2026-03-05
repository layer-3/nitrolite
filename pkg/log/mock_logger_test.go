package log_test

import "github.com/layer-3/nitrolite/pkg/log"

var _ log.Logger = &MockLogger{}

// MockLogger is a test double implementation of the Logger interface.
// It captures log entries and tracks logger state for verification in tests.
// This is particularly useful for testing components that depend on logging behavior.
type MockLogger struct {
	lastEntry MockLogEntry

	name          string
	keysAndValues []any
	callerSkip    int
}

// NewMockLogger creates a new mock logger with default values.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		name:          "mock",
		keysAndValues: []any{},
		callerSkip:    0,
	}
}

// MockLogEntry represents a captured log entry with all its metadata.
// It's used to verify that the correct log level, message, and key-value pairs were logged.
type MockLogEntry struct {
	Level         log.Level
	Message       string
	KeysAndValues []any
}

// Debug captures a debug level log entry.
func (ml *MockLogger) Debug(msg string, keysAndValues ...any) {
	ml.updateLastEntry(log.LevelDebug, msg, keysAndValues...)
}

// Info captures an info level log entry.
func (ml *MockLogger) Info(msg string, keysAndValues ...any) {
	ml.updateLastEntry(log.LevelInfo, msg, keysAndValues...)
}

// Warn captures a warning level log entry.
func (ml *MockLogger) Warn(msg string, keysAndValues ...any) {
	ml.updateLastEntry(log.LevelWarn, msg, keysAndValues...)
}

// Error captures an error level log entry.
func (ml *MockLogger) Error(msg string, keysAndValues ...any) {
	ml.updateLastEntry(log.LevelError, msg, keysAndValues...)
}

// Fatal captures a fatal level log entry.
func (ml *MockLogger) Fatal(msg string, keysAndValues ...any) {
	ml.updateLastEntry(log.LevelFatal, msg, keysAndValues...)
}

// WithKV adds a key-value pair to the logger's context.
// The key-value pair will be included in all subsequent log entries.
func (ml *MockLogger) WithKV(key string, value any) log.Logger {
	ml.keysAndValues = append(ml.keysAndValues, key, value)
	return ml
}

// GetAllKV returns all key-value pairs that have been added to this logger.
func (ml *MockLogger) GetAllKV() []any { return ml.keysAndValues }

// WithName sets the logger's name.
func (ml *MockLogger) WithName(name string) log.Logger {
	ml.name = name
	return ml
}

// Name returns the current logger name.
func (ml *MockLogger) Name() string {
	return ml.name
}

// AddCallerSkip increases the caller skip count.
// This is used to test that wrappers correctly adjust the skip level.
func (ml *MockLogger) AddCallerSkip(skip int) log.Logger {
	ml.callerSkip += skip
	return ml
}

// CallerSkip returns the current caller skip count.
// This is a test-specific method to verify AddCallerSkip behavior.
func (ml *MockLogger) CallerSkip() int {
	return ml.callerSkip
}

// LastEntry returns the most recently captured log entry.
// This is used in tests to verify logging behavior.
func (ml *MockLogger) LastEntry() MockLogEntry {
	return ml.lastEntry
}

// updateLastEntry is a helper method that captures log entry details.
// It combines the logger's context key-value pairs with the provided ones.
func (ml *MockLogger) updateLastEntry(level log.Level, msg string, keysAndValues ...any) {
	ml.lastEntry = MockLogEntry{
		Level:         level,
		Message:       msg,
		KeysAndValues: append(ml.keysAndValues, keysAndValues...),
	}
}
