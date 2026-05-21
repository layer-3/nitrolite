// Package logger provides a thin structured-logging wrapper around logrus.
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the package-level logger. It is initialised with sane defaults at
// package init time so callers never encounter a nil dereference, even if
// Initialize has not yet been called.
var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logrus.InfoLevel)
}

// Initialize reconfigures the logger with the requested level.
func Initialize(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	Log.SetLevel(logLevel)
	return nil
}

// Info logs an info-level message.
func Info(args ...interface{}) { Log.Info(args...) }

// Infof logs a formatted info-level message.
func Infof(format string, args ...interface{}) { Log.Infof(format, args...) }

// Warn logs a warn-level message.
func Warn(args ...interface{}) { Log.Warn(args...) }

// Warnf logs a formatted warn-level message.
func Warnf(format string, args ...interface{}) { Log.Warnf(format, args...) }

// Error logs an error-level message.
func Error(args ...interface{}) { Log.Error(args...) }

// Errorf logs a formatted error-level message.
func Errorf(format string, args ...interface{}) { Log.Errorf(format, args...) }

// Debug logs a debug-level message.
func Debug(args ...interface{}) { Log.Debug(args...) }

// Debugf logs a formatted debug-level message.
func Debugf(format string, args ...interface{}) { Log.Debugf(format, args...) }

// Fatal logs a fatal-level message and exits.
func Fatal(args ...interface{}) { Log.Fatal(args...) }

// Fatalf logs a formatted fatal-level message and exits.
func Fatalf(format string, args ...interface{}) { Log.Fatalf(format, args...) }
