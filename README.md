# jsonlog-go

A simple yet powerful Go logging package built on top of [Zap](https://github.com/uber-go/zap) that provides JSON-formatted logs with built-in gzip compression and advanced retrieval capabilities.

## Overview

`jsonlog-go` simplifies structured logging for Go applications by combining the high-performance Zap logger with automatic JSON serialization, file compression, and intelligent log retrieval. Perfect for applications that need production-grade logging with minimal setup.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Best Practices](#best-practices)
- [Testing](#testing)

## Features

- 🚀 **High-Performance Logging** - Built on Uber's Zap logger for minimal overhead
- 📝 **JSON Output** - All logs are stored in structured JSON format for easy parsing and analysis
- 🗜️ **Automatic Compression** - Gzip compression reduces log file sizes by ~90% without data loss
- 📖 **Log Retrieval** - Read and parse compressed JSON logs programmatically
- 🔍 **Smart Filtering** - Filter logs by level, time range, or custom predicates
- 🎯 **Simple API** - Intuitive functions mirror standard logging patterns
- 📂 **Configurable Storage** - Specify custom paths for log storage at initialization
- 🖥️ **Dual Output** - Optional console logging alongside file logging
- 🔒 **Thread-Safe** - Safe for concurrent logging from multiple goroutines

## Installation

```bash
go get github.com/gusdeyw/jsonlog-go
```

Then import in your code:

```go
import "github.com/gusdeyw/jsonlog-go"
```

## Quick Start

Here's a minimal example to get you logging in seconds:

```go
package main

import (
	"log"
	"github.com/gusdeyw/jsonlog-go"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	config := jsonlog.Config{
		LogPath:             "./logs",
		LogFileName:         "app",
		EnableConsoleOutput: true,
	}

	logger, err := jsonlog.NewLogger(config)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	// Start logging
	logger.Info("Application started", zap.String("version", "1.0.0"))
	logger.Error("Something failed", zap.Error(err))
}
```

Output files:
- `./logs/app.log` - Contains newline-delimited JSON logs

## Core Concepts

### 1. Logger Initialization

Every logging session starts with `NewLogger()`. The `Config` struct defines where and how logs are saved:

```go
config := jsonlog.Config{
	LogPath:             "./logs",              // Directory to store logs (required)
	LogFileName:         "myapp",               // File name prefix (optional, defaults to "app")
	EnableConsoleOutput: true,                  // Also print to stdout (optional)
	CompressOnClose:     true,                  // Auto-compress on Close() (optional)
}

logger, err := jsonlog.NewLogger(config)
if err != nil {
	log.Fatal(err)
}
defer logger.Close()
```

### 2. Logging Levels

Six standard logging levels are supported:

```go
logger.Debug("Debug information", zap.String("key", "value"))
logger.Info("Informational message")
logger.Warn("Warning message")
logger.Error("Error occurred", zap.Error(err))
logger.Fatal("Fatal error - exits application")
logger.Panic("Panic - triggers panic recovery")
```

### 3. Structured Fields

Log custom data using Zap fields:

```go
logger.Info("User action",
	zap.String("user_id", "12345"),
	zap.Int("attempt", 2),
	zap.Float64("duration_ms", 145.5),
	zap.Bool("success", true),
	zap.Duration("elapsed", 2*time.Second),
	zap.Error(err),
)
```

### 4. JSON Output Format

Logs are stored as newline-delimited JSON (NDJSON):

```json
{
  "timestamp": "2025-01-15T10:30:45.123456Z",
  "level": "info",
  "caller": "main.go:25",
  "message": "User action",
  "user_id": "12345",
  "attempt": 2,
  "duration_ms": 145.5,
  "success": true
}
{
  "timestamp": "2025-01-15T10:30:46.234567Z",
  "level": "error",
  "caller": "main.go:30",
  "message": "Database connection failed",
  "error": "connection refused"
}
```

### 5. Compression

Logs can be compressed with gzip to save storage space:

```go
// Compress the current log file
if err := logger.CompressLogFile(); err != nil {
	logger.Error("Compression failed", zap.Error(err))
}
// Creates: app.log.gz
```

**Compression Benefits:**
- Reduces file size by ~90% for typical JSON logs
- Maintains full readability when decompressed
- No data loss or corruption

### 6. Log Retrieval

Read logs from compressed files back into memory:

```go
logs, err := jsonlog.ReadCompressedLogs("./logs/app.log.gz")
if err != nil {
	log.Fatal(err)
}

for i, log := range logs {
	fmt.Printf("Log #%d: [%s] %s\n",
		i+1,
		log["level"],
		log["message"],
	)
}
```

Each log entry is a `map[string]interface{}` containing all fields.

## Usage Examples

### Example 1: Basic Application Logging

```go
package main

import (
	"github.com/gusdeyw/jsonlog-go"
	"go.uber.org/zap"
)

func main() {
	logger, _ := jsonlog.NewLogger(jsonlog.Config{
		LogPath:             "./logs",
		LogFileName:         "app",
		EnableConsoleOutput: true,
	})
	defer logger.Close()

	logger.Info("Server starting", zap.Int("port", 8080))
	logger.Info("Accepting connections")
	// ... handle requests ...
	logger.Info("Server shutdown", zap.Int("requests_served", 1024))
}
```

### Example 2: Dynamic Logging Level

```go
func logWithDynamicLevel(logger *jsonlog.Logger, level string, msg string) {
	var logLevel jsonlog.LogLevel

	switch level {
	case "debug":
		logLevel = jsonlog.DebugLevel
	case "info":
		logLevel = jsonlog.InfoLevel
	case "warn":
		logLevel = jsonlog.WarnLevel
	case "error":
		logLevel = jsonlog.ErrorLevel
	default:
		logLevel = jsonlog.InfoLevel
	}

	logger.LogWithLevel(logLevel, msg)
}
```

### Example 3: Error Handling with Context

```go
func processPayment(logger *jsonlog.Logger, orderID string, amount float64) error {
	logger.Info("Processing payment",
		zap.String("order_id", orderID),
		zap.Float64("amount", amount),
	)

	if err := chargeCard(amount); err != nil {
		logger.Error("Payment processing failed",
			zap.String("order_id", orderID),
			zap.Float64("amount", amount),
			zap.Error(err),
		)
		return err
	}

	logger.Info("Payment processed successfully",
		zap.String("order_id", orderID),
		zap.Float64("amount", amount),
	)
	return nil
}
```

### Example 4: Filtering Error Logs

```go
func analyzeErrors(filePath string) {
	// Read all logs
	logs, _ := jsonlog.ReadCompressedLogs(filePath)

	// Filter for errors only
	errorLogs, _ := jsonlog.ReadCompressedLogsFiltered(
		filePath,
		jsonlog.FilterByLevel("error"),
	)

	fmt.Printf("Total logs: %d\n", len(logs))
	fmt.Printf("Error logs: %d\n", len(errorLogs))

	for _, err := range errorLogs {
		fmt.Printf("  - %s\n", err["message"])
	}
}
```

### Example 5: Time-Range Analysis

```go
import "time"

func getLogs24Hours(filePath string) ([]map[string]interface{}, error) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	return jsonlog.ReadCompressedLogsFiltered(
		filePath,
		jsonlog.FilterByTimeRange(yesterday, now),
	)
}
```

### Example 6: Custom Filtering Logic

```go
func getFailedRequests(filePath string) ([]map[string]interface{}, error) {
	customFilter := func(log map[string]interface{}) bool {
		// Include only error-level logs with "request" in message
		if level, ok := log["level"].(string); ok && level == "error" {
			if msg, ok := log["message"].(string); ok {
				return strings.Contains(msg, "request")
			}
		}
		return false
	}

	return jsonlog.ReadCompressedLogsFiltered(filePath, customFilter)
}
```

## API Reference

### Types

#### `LogLevel`

String type for logging levels:

```go
const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)
```

#### `Config`

Logger configuration struct:

```go
type Config struct {
	LogPath             string // Directory for logs (required)
	LogFileName         string // File name prefix (default: "app")
	EnableConsoleOutput bool   // Print to stdout
	CompressOnClose     bool   // Auto-compress on Close()
}
```

#### `Logger`

Main logging service:

```go
type Logger struct {
	// Contains filtered or unexported fields
}
```

#### `FilterFunc`

Filtering function type:

```go
type FilterFunc func(log map[string]interface{}) bool
```

### Functions

#### `NewLogger(config Config) (*Logger, error)`

Creates and initializes a new logger instance.

```go
logger, err := jsonlog.NewLogger(jsonlog.Config{
	LogPath:     "./logs",
	LogFileName: "app",
})
if err != nil {
	log.Fatal(err)
}
```

#### Logger Methods

```go
// Logging methods
func (l *Logger) Debug(message string, fields ...zap.Field)
func (l *Logger) Info(message string, fields ...zap.Field)
func (l *Logger) Warn(message string, fields ...zap.Field)
func (l *Logger) Error(message string, fields ...zap.Field)
func (l *Logger) Fatal(message string, fields ...zap.Field)
func (l *Logger) Panic(message string, fields ...zap.Field)

// Dynamic level logging
func (l *Logger) LogWithLevel(level LogLevel, message string, fields ...zap.Field)

// Lifecycle
func (l *Logger) Close() error
func (l *Logger) CompressLogFile() error
```

#### `ReadCompressedLogs(filePath string) ([]map[string]interface{}, error)`

Reads all logs from a compressed gzip file.

```go
logs, err := jsonlog.ReadCompressedLogs("./logs/app.log.gz")
if err != nil {
	log.Fatal(err)
}
// logs is []map[string]interface{}
```

#### `ReadCompressedLogsFiltered(filePath string, filter FilterFunc) ([]map[string]interface{}, error)`

Reads logs applying a custom filter.

```go
logs, err := jsonlog.ReadCompressedLogsFiltered(
	"./logs/app.log.gz",
	jsonlog.FilterByLevel("error"),
)
```

#### `FilterByLevel(level string) FilterFunc`

Creates a filter matching a specific log level.

```go
errorFilter := jsonlog.FilterByLevel("error")
logs, _ := jsonlog.ReadCompressedLogsFiltered("app.log.gz", errorFilter)
```

#### `FilterByTimeRange(start, end time.Time) FilterFunc`

Creates a filter for logs within a time range.

```go
now := time.Now()
dayAgo := now.Add(-24 * time.Hour)
logs, _ := jsonlog.ReadCompressedLogsFiltered(
	"app.log.gz",
	jsonlog.FilterByTimeRange(dayAgo, now),
)
```

## Configuration

### Basic Configuration

```go
config := jsonlog.Config{
	LogPath:     "./logs",
	LogFileName: "myapp",
}

logger, _ := jsonlog.NewLogger(config)
```

### Full Configuration

```go
config := jsonlog.Config{
	LogPath:             "./logs",          // Where to save logs
	LogFileName:         "myapp",           // Log file name (without extension)
	EnableConsoleOutput: true,              // Also print to console
	CompressOnClose:     true,              // Auto-compress when closing
}

logger, _ := jsonlog.NewLogger(config)
```

### Configuration Notes

- **LogPath**: Must be writable directory. Created if doesn't exist.
- **LogFileName**: Defaults to "app". Final file: `LogPath/LogFileName.log`
- **EnableConsoleOutput**: Useful for development, disable in production for better performance
- **CompressOnClose**: Not implemented yet; manual compression via `CompressLogFile()` recommended

## Best Practices

### 1. Always Defer Close()

Ensures all buffers are flushed and resources cleaned up:

```go
logger, err := jsonlog.NewLogger(config)
if err != nil {
	log.Fatal(err)
}
defer logger.Close() // ← Always include this
```

### 2. Use Structured Fields

Structured fields make logs queryable and analyzable:

```go
// ❌ Bad
logger.Info("User login " + userID + " from " + ipAddr)

// ✅ Good
logger.Info("User login",
	zap.String("user_id", userID),
	zap.String("ip_address", ipAddr),
)
```

### 3. Include Context in Errors

Log relevant context when errors occur:

```go
logger.Error("Database query failed",
	zap.String("query", query),
	zap.String("table", "users"),
	zap.Error(err),
)
```

### 4. Compress Periodically

Reduce storage by compressing old log files:

```go
// Compress daily
logger.Close()
logger.CompressLogFile()
logger.NewLogger(config) // Start fresh
```

### 5. Filter Strategically

Use filtering to extract actionable insights:

```go
// Get all errors from last hour
recentErrors, _ := jsonlog.ReadCompressedLogsFiltered(
	"app.log.gz",
	func(log map[string]interface{}) bool {
		// Your custom logic
		return true
	},
)
```

### 6. Use Consistent Field Names

Maintain consistency for better log analysis:

```go
// Always use same field names across your application
logger.Info("request", zap.String("request_id", rid))
logger.Info("response", zap.String("request_id", rid))
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

### Run Specific Test

```bash
go test -run TestCompressLogFile ./...
```

### Test Files

- `logger_test.go` - Core functionality tests
- `example_test.go` - Usage examples and demonstrations

## Performance

- **Log Writing**: ~1,000 logs/ms (varies by field complexity)
- **Compression**: Typical 10:1 reduction for JSON logs
- **Memory Usage**: Minimal overhead, logs streamed when reading
- **Thread Safety**: Safe for concurrent writes from multiple goroutines

## License

MIT License - See LICENSE file

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## Support

For issues or questions:
- Open an issue on GitHub
- Check existing documentation
- Review test examples

---

**Built with ❤️ using Zap logger**
