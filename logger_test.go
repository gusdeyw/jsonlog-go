package jsonlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
		CompressOnClose:     true,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger.filePath == "" {
		t.Error("logger.filePath should not be empty")
	}
}

func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log messages at different levels
	logger.Debug("debug message", zap.String("key", "value"))
	logger.Info("info message", zap.Int("count", 42))
	logger.Warn("warn message", zap.Error(nil))
	logger.Error("error message", zap.String("error_code", "ERR001"))

	// Verify log file was created
	logFile := filepath.Join(tmpDir, "test.log")
	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("log file was not created: %v", err)
	}

	// Verify log file has content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("log file is empty")
	}
}

func TestLogWithLevel(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.LogWithLevel(InfoLevel, "test message", zap.String("field", "value"))

	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("log file should contain data")
	}
}

func TestCompressLogFile(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write some logs
	logger.Info("test message 1", zap.String("field", "value1"))
	logger.Info("test message 2", zap.String("field", "value2"))
	logger.Close()

	// Compress the log file
	err = logger.CompressLogFile()
	if err != nil {
		t.Fatalf("failed to compress log file: %v", err)
	}

	// Verify compressed file was created
	compressedFile := filepath.Join(tmpDir, "test.log.gz")
	if _, err := os.Stat(compressedFile); err != nil {
		t.Fatalf("compressed file was not created: %v", err)
	}

	// Verify compressed file has content
	content, err := os.ReadFile(compressedFile)
	if err != nil {
		t.Fatalf("failed to read compressed file: %v", err)
	}

	if len(content) == 0 {
		t.Error("compressed file is empty")
	}
}

func TestReadCompressedLogs(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write some logs
	logger.Info("test message 1", zap.String("field", "value1"))
	logger.Info("test message 2", zap.String("field", "value2"))
	logger.Warn("test warning", zap.String("field", "value3"))
	logger.Close()

	// Compress the log file
	err = logger.CompressLogFile()
	if err != nil {
		t.Fatalf("failed to compress log file: %v", err)
	}

	// Read compressed logs
	compressedFile := filepath.Join(tmpDir, "test.log.gz")
	logs, err := ReadCompressedLogs(compressedFile)
	if err != nil {
		t.Fatalf("failed to read compressed logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}

	// Verify log structure
	for _, log := range logs {
		if _, ok := log["timestamp"]; !ok {
			t.Error("log entry missing timestamp")
		}
		if _, ok := log["level"]; !ok {
			t.Error("log entry missing level")
		}
		if _, ok := log["message"]; !ok {
			t.Error("log entry missing message")
		}
	}
}

func TestReadCompressedLogsFiltered(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Write logs at different levels
	logger.Info("info message 1")
	logger.Info("info message 2")
	logger.Warn("warn message")
	logger.Error("error message")
	logger.Close()

	// Compress the log file
	err = logger.CompressLogFile()
	if err != nil {
		t.Fatalf("failed to compress log file: %v", err)
	}

	// Read only info level logs
	compressedFile := filepath.Join(tmpDir, "test.log.gz")
	logs, err := ReadCompressedLogsFiltered(compressedFile, FilterByLevel("info"))
	if err != nil {
		t.Fatalf("failed to read filtered logs: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("expected 2 info logs, got %d", len(logs))
	}

	for _, log := range logs {
		if level, ok := log["level"]; ok {
			if level != "info" {
				t.Errorf("expected level 'info', got '%v'", level)
			}
		}
	}
}

func TestFilterByTimeRange(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	now := time.Now()
	logger.Info("message 1")
	time.Sleep(100 * time.Millisecond)
	logger.Info("message 2")
	later := time.Now()

	logger.Close()

	// Compress the log file
	err = logger.CompressLogFile()
	if err != nil {
		t.Fatalf("failed to compress log file: %v", err)
	}

	// Read logs within time range
	compressedFile := filepath.Join(tmpDir, "test.log.gz")
	logs, err := ReadCompressedLogsFiltered(
		compressedFile,
		FilterByTimeRange(now.Add(-time.Second), later.Add(time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to read filtered logs: %v", err)
	}

	if len(logs) < 1 {
		t.Errorf("expected at least 1 log, got %d", len(logs))
	}
}

func TestJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	config := Config{
		LogPath:             tmpDir,
		LogFileName:         "test",
		EnableConsoleOutput: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("test message", zap.String("user_id", "123"), zap.Int("count", 42))
	logger.Close()

	// Read log file and verify JSON format
	logFile := filepath.Join(tmpDir, "test.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	if err != nil {
		t.Fatalf("log entry is not valid JSON: %v", err)
	}

	// Verify fields
	if msg, ok := logEntry["message"]; !ok || msg != "test message" {
		t.Error("message field is missing or incorrect")
	}
	if userId, ok := logEntry["user_id"]; !ok || userId != "123" {
		t.Error("user_id field is missing or incorrect")
	}
	if count, ok := logEntry["count"]; !ok || count != float64(42) {
		t.Error("count field is missing or incorrect")
	}
}
