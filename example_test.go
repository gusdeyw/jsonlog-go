package jsonlog

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Example demonstrating basic logger usage
func TestExampleNewLogger(t *testing.T) {
	tmpDir := "example"

	// Create logger with config
	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "myapp",
		EnableConsoleOutput: true,
		CompressOnClose:     true,
	}

	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Log messages at different levels
	logger.Debug("Starting application", zap.String("version", "1.0.0"))
	logger.Info("User login", zap.String("user_id", "user123"))
	logger.Warn("High memory usage", zap.Int("memory_mb", 800))
	logger.Error("Database connection failed", zap.String("host", "localhost"))

	logger.Close()
}

// Example demonstrating log compression
func TestExampleCompression(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "app",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Generate some logs
	for i := 0; i < 100; i++ {
		logger.Info("Processing request",
			zap.Int("request_id", i),
			zap.String("status", "success"),
		)
	}

	logger.Close()

	// Compress logs
	if err := logger.CompressLogFile(); err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	// Check file sizes
	logFile := filepath.Join(tmpDir, "app.log")
	compressedFile := filepath.Join(tmpDir, "app.log.gz")

	info1, _ := os.Stat(logFile)
	info2, _ := os.Stat(compressedFile)

	fmt.Printf("Original: %d bytes, Compressed: %d bytes\n", info1.Size(), info2.Size())
	fmt.Printf("Compression ratio: %.2f%%\n", float64(info2.Size())/float64(info1.Size())*100)
}

// Example demonstrating reading compressed logs
func TestExampleReadCompressed(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "app",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write logs
	logger.Info("User registered", zap.String("email", "user@example.com"))
	logger.Warn("Suspicious activity detected", zap.String("ip", "192.168.1.1"))
	logger.Error("Payment processing failed", zap.String("order_id", "ORD123"))
	logger.Close()

	// Compress
	if err := logger.CompressLogFile(); err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	// Read compressed logs
	compressedFile := filepath.Join(tmpDir, "app.log.gz")
	logs, err := ReadCompressedLogs(compressedFile)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	fmt.Printf("Total logs: %d\n", len(logs))
	for i, log := range logs {
		fmt.Printf("Log %d: %s - %s\n", i+1, log["level"], log["message"])
	}
}

// Example demonstrating log filtering
func TestExampleLogFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "app",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Log various messages
	logger.Debug("Debug info")
	logger.Info("Operation successful")
	logger.Info("Another success")
	logger.Warn("Warning issued")
	logger.Error("Error occurred")
	logger.Error("Another error")
	logger.Close()

	// Compress
	if err := logger.CompressLogFile(); err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	compressedFile := filepath.Join(tmpDir, "app.log.gz")

	// Filter for error logs only
	errorLogs, err := ReadCompressedLogsFiltered(
		compressedFile,
		FilterByLevel("error"),
	)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	fmt.Printf("Error logs found: %d\n", len(errorLogs))
	for _, log := range errorLogs {
		fmt.Printf("  - %s\n", log["message"])
	}

	// Filter for time range
	now := time.Now()
	timeLogs, err := ReadCompressedLogsFiltered(
		compressedFile,
		FilterByTimeRange(now.Add(-time.Hour), now.Add(time.Hour)),
	)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	fmt.Printf("Logs in time range: %d\n", len(timeLogs))
}

// Example demonstrating custom field logging
func TestExampleCustomFields(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "app",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Log with various field types
	logger.Info("API Request",
		zap.String("endpoint", "/api/users"),
		zap.String("method", "POST"),
		zap.Int("status_code", 201),
		zap.Duration("response_time", 125*time.Millisecond),
		zap.Float64("memory_usage_mb", 234.5),
		zap.Bool("success", true),
	)

	logger.Close()

	// Read back
	logger.CompressLogFile()
	compressedFile := filepath.Join(tmpDir, "app.log.gz")
	logs, _ := ReadCompressedLogs(compressedFile)

	if len(logs) > 0 {
		fmt.Printf("Logged fields: %v\n", logs[0])
	}
}
