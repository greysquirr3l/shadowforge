// Package logger provides a global logrus-based logger for Shadowforge.
package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	// Log is the global logger instance
	Log *logrus.Logger
)

func init() {
	Log = NewLogger()
}

// NewLogger creates a new configured logrus logger instance.
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Use text formatter with colors for CLI
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
		PadLevelText:    true,
	})

	// Set default log level to Info
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

// SetLevel sets the global logger level.
func SetLevel(level logrus.Level) {
	Log.SetLevel(level)
}

// SetOutput sets the global logger output.
func SetOutput(output io.Writer) {
	Log.SetOutput(output)
}

// Convenience functions for common log levels

// Debug logs a debug message.
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Info logs an info message.
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Warn logs a warning message.
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// Error logs an error message.
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatal logs a fatal message and exits.
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits.
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// WithField creates a new logger entry with a single field.
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// WithFields creates a new logger entry with multiple fields.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// WithError creates a new logger entry with an error field.
func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}
