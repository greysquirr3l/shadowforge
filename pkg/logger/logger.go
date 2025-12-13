// Package logger provides a global logrus-based logger for Shadowforge.
// It supports both CLI (colorized text) and API (JSON) modes.
package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogMode defines the logging mode (CLI or API)
type LogMode string

const (
	// ModeCLI uses colorized text output for interactive CLI
	ModeCLI LogMode = "cli"
	// ModeAPI uses JSON output for API server logs
	ModeAPI LogMode = "api"
)

var (
	// Log is the global logger instance
	Log *logrus.Logger
	// currentMode tracks the current logging mode
	currentMode LogMode = ModeCLI
)

func init() {
	// Initialize with CLI mode by default
	Log = NewCLILogger()
}

// NewLogger creates a new logger based on the environment.
// Checks SHADOWFORGE_LOG_MODE environment variable (cli|api).
// Defaults to CLI mode.
func NewLogger() *logrus.Logger {
	mode := os.Getenv("SHADOWFORGE_LOG_MODE")
	switch strings.ToLower(mode) {
	case "api", "json":
		return NewAPILogger()
	default:
		return NewCLILogger()
	}
}

// NewCLILogger creates a new logger configured for CLI usage with colorized output.
func NewCLILogger() *logrus.Logger {
	logger := logrus.New()
	currentMode = ModeCLI

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
	logger.SetLevel(getLogLevel())

	return logger
}

// NewAPILogger creates a new logger configured for API server usage with JSON output.
func NewAPILogger() *logrus.Logger {
	logger := logrus.New()
	currentMode = ModeAPI

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Use JSON formatter for API/server logs
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   "2006-01-02T15:04:05.000Z07:00",
		DisableTimestamp:  false,
		DisableHTMLEscape: true,
		PrettyPrint:       false, // Set to true for development if needed
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	// Set default log level to Info
	logger.SetLevel(getLogLevel())

	return logger
}

// getLogLevel reads the log level from environment variable.
// Checks SHADOWFORGE_LOG_LEVEL (debug|info|warn|error|fatal|panic).
// Defaults to Info.
func getLogLevel() logrus.Level {
	level := os.Getenv("SHADOWFORGE_LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// SetMode switches the logger to the specified mode (CLI or API).
func SetMode(mode LogMode) {
	switch mode {
	case ModeAPI:
		Log = NewAPILogger()
	case ModeCLI:
		Log = NewCLILogger()
	default:
		Log = NewCLILogger()
	}
}

// GetMode returns the current logging mode.
func GetMode() LogMode {
	return currentMode
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
