# jsonlog-go - Complete Documentation

A simple and powerful Go logging package built on top of [Zap](https://github.com/uber-go/zap) that provides JSON-formatted logs with built-in compression and retrieval capabilities.

## Features

- **Zap-based Logging**: Leverages the high-performance Zap logger from Uber
- **JSON Output**: All logs are stored in JSON format for easy parsing and analysis
- **Gzip Compression**: Automatically compress log files to reduce storage size
- **Log Retrieval**: Read and parse compressed JSON logs programmatically
- **Log Filtering**: Filter logs by level, time range, or custom functions
- **Simple API**: Clean and intuitive interface for logging operations
- **File-based Storage**: Specify custom paths for log storage
- **Console Output**: Optional console logging alongside file logging

## Installation

```bash
go get github.com/gusdeyw/jsonlog-go
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/gusdeyw/jsonlog-go"
    "go.uber.org/zap"
)

func main() {
    // Create logger with config
    config := jsonlog.Config{
        LogPath:             "./logs",
        LogFileName:         "app",
        EnableConsoleOutput: true,
        CompressOnClose:     true,
    }

    logger, err := jsonlog.NewLogger(config)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()

    // Log messages at different levels
    logger.Info("Application started", zap.String("version", "1.0.0"))
    logger.Error("Something went wrong", zap.Error(someError))
}
```

## Usage Examples

### Basic Logging

```go
// Log at specific levels
logger.Debug("Debug message", zap.String("key", "value"))
logger.Info("Info message", zap.Int("count", 42))
logger.Warn("Warning message", zap.Float64("rate", 3.14))
logger.Error("Error message", zap.Error(err))

// Log with dynamic level
logger.LogWithLevel(jsonlog.InfoLevel, "Dynamic message")
```

### Supported Field Types

The package supports all Zap field types:

```go
logger.Info("Log with various fields",
    zap.String("user", "john"),
    zap.Int("age", 30),
    zap.Float64("score", 95.5),
    zap.Bool("active", true),
    zap.Duration("elapsed", time.Second),
    zap.Error(err),
)
```

### Compression

```go
// Write logs
logger.Info("Message 1")
logger.Info("Message 2")
logger.Close()

// Compress the log file
err := logger.CompressLogFile()
if err != nil {
    log.Fatal(err)
}
// Result: logs/app.log.gz
```

### Reading Compressed Logs

```go
import "path/filepath"

// Read all logs from compressed file
logs, err := jsonlog.ReadCompressedLogs("./logs/app.log.gz")
if err != nil {
    log.Fatal(err)
}

for _, logEntry := range logs {
    fmt.Printf("%s: %s\n", logEntry["level"], logEntry["message"])
}
```

### Filtering Logs

#### Filter by Level

```go
logs, err := jsonlog.ReadCompressedLogsFiltered(
    "./logs/app.log.gz",
    jsonlog.FilterByLevel("error"),
)
```

#### Filter by Time Range

```go
import "time"

start := time.Now().Add(-24 * time.Hour)
end := time.Now()

logs, err := jsonlog.ReadCompressedLogsFiltered(
    "./logs/app.log.gz",
    jsonlog.FilterByTimeRange(start, end),
)
```

#### Custom Filters

```go
// Filter for logs with specific message
customFilter := func(log map[string]interface{}) bool {
    if msg, ok := log["message"].(string); ok {
        return strings.Contains(msg, "error")
    }
    return false
}

logs, err := jsonlog.ReadCompressedLogsFiltered(
    "./logs/app.log.gz",
    customFilter,
)

// Combine multiple filters
logs, err := jsonlog.ReadCompressedLogsFiltered(
    "./logs/app.log.gz",
    func(log map[string]interface{}) bool {
        levelOk := log["level"] == "error"
        messageOk := strings.Contains(
            log["message"].(string), 
            "database",
        )
        return levelOk && messageOk
    },
)
```

## Configuration

```go
type Config struct {
    // LogPath is the directory where logs will be saved (required)
    LogPath string

    // LogFileName is the name of the log file without extension
    // (optional, defaults to "app")
    LogFileName string

    // EnableConsoleOutput determines if logs should also be printed to console
    // (optional, defaults to false)
    EnableConsoleOutput bool

    // CompressOnClose enables gzip compression when closing
    // (optional, defaults to false)
    CompressOnClose bool
}
```

## Log Entry Structure

Each log entry is stored as JSON with the following structure:

```json
{
  "timestamp": "2025-01-15T10:30:45.123456Z",
  "level": "info",
  "caller": "main.go:25",
  "message": "User logged in",
  "user_id": "12345",
  "session_id": "abc123",
  "duration_ms": 145
}
```

## Performance Considerations

- **Compression**: Typical JSON logs compress at 10:1 ratio or better
- **I/O**: File writes are buffered for efficiency
- **Memory**: Logs are streamed when read from compressed files
- **Concurrency**: Thread-safe operations with internal locking

## API Reference

### Logger Methods

```go
// Core logging methods
logger.Debug(message string, fields ...zap.Field)
logger.Info(message string, fields ...zap.Field)
logger.Warn(message string, fields ...zap.Field)
logger.Error(message string, fields ...zap.Field)
logger.Fatal(message string, fields ...zap.Field)
logger.Panic(message string, fields ...zap.Field)

// Dynamic level logging
logger.LogWithLevel(level LogLevel, message string, fields ...zap.Field)

// Lifecycle
logger.Close() error
logger.CompressLogFile() error
```

### Package Functions

```go
// Read logs
ReadCompressedLogs(filePath string) ([]map[string]interface{}, error)

// Read filtered logs
ReadCompressedLogsFiltered(
    filePath string,
    filter FilterFunc,
) ([]map[string]interface{}, error)

// Built-in filters
FilterByLevel(level string) FilterFunc
FilterByTimeRange(start, end time.Time) FilterFunc
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## Best Practices

1. **Always defer Close()**: Ensure buffers are flushed and synced
   ```go
   logger, err := jsonlog.NewLogger(config)
   if err != nil {
       log.Fatal(err)
   }
   defer logger.Close()
   ```

2. **Compress old logs**: Periodically compress log files to save space
   ```go
   logger.CompressLogFile()
   ```

3. **Use structured fields**: Leverage structured logging for queryable data
   ```go
   logger.Info("User action",
       zap.String("user_id", userID),
       zap.String("action", "login"),
       zap.Time("timestamp", time.Now()),
   )
   ```

4. **Handle compression errors**: Always check for errors when compressing
   ```go
   if err := logger.CompressLogFile(); err != nil {
       logger.Error("Failed to compress logs", zap.Error(err))
   }
   ```

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Dependencies

- [go.uber.org/zap](https://github.com/uber-go/zap) - High-performance logging

## Support

For issues or questions, please open an issue on GitHub.
