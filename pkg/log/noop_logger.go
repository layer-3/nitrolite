package log

var _ Logger = NoopLogger{}

// NoopLogger is a logger implementation that discards all log messages.
// It implements the Logger interface but performs no actual logging operations.
// This is useful for testing or when logging needs to be disabled.
type NoopLogger struct{}

// NewNoopLogger creates a new NoopLogger instance.
// All logging operations on the returned logger will be silently discarded.
func NewNoopLogger() Logger {
	return NoopLogger{}
}

// Debug implements Logger.Debug but performs no operation.
func (n NoopLogger) Debug(msg string, keysAndValues ...any) {}

// Info implements Logger.Info but performs no operation.
func (n NoopLogger) Info(msg string, keysAndValues ...any) {}

// Warn implements Logger.Warn but performs no operation.
func (n NoopLogger) Warn(msg string, keysAndValues ...any) {}

// Error implements Logger.Error but performs no operation.
func (n NoopLogger) Error(msg string, keysAndValues ...any) {}

// Fatal implements Logger.Fatal but performs no operation.
func (n NoopLogger) Fatal(msg string, keysAndValues ...any) {}

// WithKV implements Logger.WithKV but returns the same NoopLogger instance.
func (n NoopLogger) WithKV(key string, value any) Logger { return n }

// GetAllKV implements Logger.GetAllKV and returns an empty slice.
func (n NoopLogger) GetAllKV() []any { return []any{} }

// WithName implements Logger.WithName but returns the same NoopLogger instance.
func (n NoopLogger) WithName(name string) Logger { return n }

// Name implements Logger.Name and always returns "noop".
func (n NoopLogger) Name() string { return "noop" }

// AddCallerSkip implements Logger.AddCallerSkip but returns the same NoopLogger instance.
func (n NoopLogger) AddCallerSkip(skip int) Logger { return n }
