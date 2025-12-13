# Logger Package

Shadowforge's global logging package using [logrus](https://github.com/sirupsen/logrus)
with support for both CLI (colorized text) and API (JSON) modes.

## Features

- **Dual Mode Support**: CLI mode with colorized output, API mode with structured JSON
- **Global Logger**: Easy access via `logger.Log` or convenience functions
- **Environment Configuration**: Configure via `SHADOWFORGE_LOG_MODE` and `SHADOWFORGE_LOG_LEVEL`
- **Structured Logging**: Full support for fields and context
- **Production Ready**: JSON output for log aggregation, text output for debugging

## Quick Start

```go
import "github.com/greysquirr3l/shadowforge/pkg/logger"

func main() {
    // Simple logging
    logger.Info("Application started")
    logger.Debugf("Debug info: %s", debugData)

    // Structured logging
    logger.WithFields(logrus.Fields{
        "user_id": "123",
        "action": "login",
    }).Info("User action")

    // Error logging
    if err != nil {
        logger.WithError(err).Error("Operation failed")
    }
}
```

## Modes

### CLI Mode (Default)

Perfect for interactive command-line tools. Produces colorized, human-readable output.

```go
logger.SetMode(logger.ModeCLI)
// Output: 2024-01-15 10:30:45 INFO  Application started
```

**Environment Variable**:

```bash
export SHADOWFORGE_LOG_MODE=cli
```

### API Mode

Perfect for API servers and production deployments. Produces structured JSON logs.

```go
logger.SetMode(logger.ModeAPI)
// Output: {"timestamp":"2024-01-15T10:30:45.000Z","level":"info","message":"Application started"}
```

**Environment Variable**:

```bash
export SHADOWFORGE_LOG_MODE=api
```

## Configuration

### Log Level

Set via environment variable:

```bash
export SHADOWFORGE_LOG_LEVEL=debug   # trace, debug, info, warn, error, fatal, panic
```

Or programmatically:

```go
logger.SetLevel(logrus.DebugLevel)
```

### Log Output

Change output destination:

```go
import "os"

// Write to file
file, _ := os.Create("app.log")
logger.SetOutput(file)

// Write to stderr
logger.SetOutput(os.Stderr)
```

## Usage Examples

### Basic Logging

```go
// Simple messages
logger.Debug("Debug message")
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message")

// Formatted messages
logger.Infof("Processing file: %s", filename)
logger.Errorf("Failed to connect to %s: %v", host, err)
```

### Structured Logging

```go
// Single field
logger.WithField("user_id", "123").Info("User logged in")

// Multiple fields
logger.WithFields(logrus.Fields{
    "request_id": "req-123",
    "method": "POST",
    "path": "/api/embed",
    "status": 201,
    "duration_ms": 45,
}).Info("Request completed")

// With error
if err != nil {
    logger.WithError(err).WithFields(logrus.Fields{
        "operation": "encrypt",
        "file": filename,
    }).Error("Encryption failed")
}
```

### CLI Application Example

```go
package main

import (
    "github.com/greysquirr3l/shadowforge/pkg/logger"
    "github.com/spf13/cobra"
)

func main() {
    // CLI mode is default
    logger.Info("🚀 Shadowforge CLI starting...")

    rootCmd := &cobra.Command{
        Use: "shadowforge",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // Set log level from flag
            if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
                logger.SetLevel(logrus.DebugLevel)
            }
        },
    }

    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
    rootCmd.Execute()
}
```

### API Server Example

```go
package main

import (
    "github.com/greysquirr3l/shadowforge/pkg/logger"
    "github.com/labstack/echo/v4"
)

func main() {
    // Switch to JSON mode for API server
    logger.SetMode(logger.ModeAPI)

    e := echo.New()

    // Logging middleware
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            logger.WithFields(logrus.Fields{
                "method": c.Request().Method,
                "path":   c.Request().URL.Path,
                "ip":     c.RealIP(),
            }).Info("API request")

            return next(c)
        }
    })

    e.Start(":8080")
}
```

### Domain Service Example

```go
package crypto

import (
    "context"
    "github.com/greysquirr3l/shadowforge/pkg/logger"
)

type CryptoService struct {
    // ... fields
}

func (s *CryptoService) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
    logger.WithFields(logrus.Fields{
        "operation": "encrypt",
        "size": len(data),
        "algorithm": "kyber1024",
    }).Debug("Starting encryption")

    // Encryption logic...

    if err != nil {
        logger.WithError(err).WithFields(logrus.Fields{
            "operation": "encrypt",
            "size": len(data),
        }).Error("Encryption failed")
        return nil, err
    }

    logger.WithField("encrypted_size", len(encrypted)).Info("Encryption successful")
    return encrypted, nil
}
```

### Testing with Custom Logger

```go
func TestSomething(t *testing.T) {
    // Capture logs during testing
    var buf bytes.Buffer
    logger.SetOutput(&buf)
    logger.SetLevel(logrus.DebugLevel)

    // Run test
    result := doSomething()

    // Check logs
    output := buf.String()
    assert.Contains(t, output, "expected log message")
}
```

## Best Practices

### DO ✅

- **Use structured fields** for queryable data:

  ```go
  logger.WithFields(logrus.Fields{
      "user_id": userID,
      "action": "login",
  }).Info("User action")
  ```

- **Include context** in error logs:

  ```go
  logger.WithError(err).WithField("operation", "encrypt").Error("Operation failed")
  ```

- **Use appropriate levels**:
  - `Debug`: Detailed diagnostic info
  - `Info`: General informational messages
  - `Warn`: Warning messages (recoverable issues)
  - `Error`: Error messages (operation failed)
  - `Fatal`: Fatal errors (application exits)

### DON'T ❌

- **Don't log sensitive data**:

  ```go
  // ❌ NEVER do this
  logger.WithField("password", password).Debug("Auth attempt")
  logger.WithField("private_key", key).Info("Key generated")

  // ✅ Do this instead
  logger.WithField("key_id", keyID).Info("Key generated")
  ```

- **Don't use string concatenation** in hot paths:

  ```go
  // ❌ Slower
  logger.Info("Processing " + filename + " with size " + strconv.Itoa(size))

  // ✅ Faster
  logger.WithFields(logrus.Fields{
      "filename": filename,
      "size": size,
  }).Info("Processing file")
  ```

- **Don't overuse Fatal/Panic** - they exit the application:

  ```go
  // ❌ Don't use Fatal for recoverable errors
  if err != nil {
      logger.Fatal(err)  // Exits entire application!
  }

  // ✅ Use Error and return
  if err != nil {
      logger.WithError(err).Error("Operation failed")
      return err
  }
  ```

## Environment Configuration Reference

| Variable | Values | Default | Description |
|----------|--------|---------|-------------|
| `SHADOWFORGE_LOG_MODE` | `cli`, `api`, `json` | `cli` | Logging output format |
| `SHADOWFORGE_LOG_LEVEL` | `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic` | `info` | Minimum log level |

## JSON Field Mapping

When using API mode, fields are mapped as follows:

| Logrus Field | JSON Field |
|--------------|------------|
| `time` | `timestamp` |
| `level` | `level` |
| `msg` | `message` |
| `func` | `caller` |

Custom fields are preserved as-is.

## Integration with Other Components

### With Echo (API Server)

```go
import (
    "github.com/greysquirr3l/shadowforge/pkg/logger"
    "github.com/labstack/echo/v4"
)

// Create custom middleware
func LoggerMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            req := c.Request()
            res := c.Response()

            logger.WithFields(logrus.Fields{
                "method": req.Method,
                "uri": req.RequestURI,
                "ip": c.RealIP(),
                "status": res.Status,
            }).Info("HTTP request")

            return next(c)
        }
    }
}
```

### With Cobra (CLI)

```go
import (
    "github.com/greysquirr3l/shadowforge/pkg/logger"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use: "shadowforge",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
            logger.SetLevel(logrus.DebugLevel)
            logger.Debug("Verbose logging enabled")
        }
    },
}

func init() {
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
```

## Migration Guide

If you're currently using `slog`, you can gradually migrate:

```go
// Old (slog)
slog.Info("message", slog.String("key", "value"))

// New (logrus via logger package)
logger.WithField("key", "value").Info("message")
```

## Performance Considerations

- **Structured logging** with fields is more efficient than string formatting
- **JSON mode** has slightly more overhead than text mode (negligible for most use cases)
- **Debug/Trace logs** can be completely disabled in production by setting level to Info or higher
- Use **lazy evaluation** for expensive operations:

  ```go
  if logger.Log.Level >= logrus.DebugLevel {
      logger.Debugf("Expensive operation: %s", expensiveFunction())
  }
  ```

## License

Part of Shadowforge - see root LICENSE file.
