// Package jsonlog provides a structured JSON logging system built on top of Uber's zap library.
// It supports file rotation, compression, filtering, and flexible output configuration.
//
// Example usage:
//
//	config := jsonlog.Config{
//		LogPath:             "./logs",
//		LogFileName:         "app",
//		EnableConsoleOutput: true,
//		CompressOnClose:     true,
//	}
//
//	logger, err := jsonlog.NewLogger(config)
//	if err != nil {
//		panic(err)
//	}
//	defer logger.Close()
//
//	// Log messages with structured fields
//	logger.Info("User login", zap.String("user_id", "user123"))
//	logger.Error("Database error", zap.String("error", "connection timeout"))
//
//	// Compress logs after closing
//	logger.CompressLogFile()
//
//	// Read compressed logs with filtering
//	logs, err := jsonlog.ReadCompressedLogsFiltered(
//		"./logs/app.log.gz",
//		jsonlog.FilterByLevel("error"),
//	)
package jsonlog

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// Logger is the main logging service
type Logger struct {
	zapLogger  *zap.Logger
	filePath   string
	fileLogger *lumberjack.Logger
	mu         sync.Mutex
}

// Config holds the logger configuration
type Config struct {
	// LogPath is the directory where logs will be saved
	LogPath string

	// LogFileName is the name of the log file (without extension)
	LogFileName string

	// EnableConsoleOutput determines if logs should also be printed to console
	EnableConsoleOutput bool

	// CompressOnClose enables gzip compression when closing
	CompressOnClose bool

	// RotationSize is the max size in bytes before rotation (0 = no rotation)
	RotationSize int64
}

// NewLogger creates a new logger instance
func NewLogger(config Config) (*Logger, error) {
	// Validate config
	if config.LogPath == "" {
		return nil, fmt.Errorf("LogPath is required")
	}

	if config.LogFileName == "" {
		config.LogFileName = "app"
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(config.LogPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Build file path
	logFilePath := filepath.Join(config.LogPath, config.LogFileName+".log")

	// Create Zap logger configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// File output (always JSON) - using lumberjack for proper file handle management
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	}
	fileSync := zapcore.AddSync(fileLogger)
	fileCore := zapcore.NewCore(fileEncoder, fileSync, zapcore.DebugLevel)
	cores = append(cores, fileCore)

	// Console output (if enabled)
	if config.EnableConsoleOutput {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	// Create combined logger
	combinedCore := zapcore.NewTee(cores...)
	zapLogger := zap.New(combinedCore, zap.AddCaller())

	logger := &Logger{
		zapLogger:  zapLogger,
		filePath:   logFilePath,
		fileLogger: fileLogger,
	}

	return logger, nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...zap.Field) {
	l.zapLogger.Debug(message, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...zap.Field) {
	l.zapLogger.Info(message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.zapLogger.Warn(message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...zap.Field) {
	l.zapLogger.Error(message, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...zap.Field) {
	l.zapLogger.Fatal(message, fields...)
}

// Panic logs a panic message
func (l *Logger) Panic(message string, fields ...zap.Field) {
	l.zapLogger.Panic(message, fields...)
}

// LogWithLevel logs a message with specified level
func (l *Logger) LogWithLevel(level LogLevel, message string, fields ...zap.Field) {
	switch level {
	case DebugLevel:
		l.Debug(message, fields...)
	case InfoLevel:
		l.Info(message, fields...)
	case WarnLevel:
		l.Warn(message, fields...)
	case ErrorLevel:
		l.Error(message, fields...)
	case FatalLevel:
		l.Fatal(message, fields...)
	case PanicLevel:
		l.Panic(message, fields...)
	default:
		l.Info(message, fields...)
	}
}

// Close closes the logger and flushes buffers
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Sync zap logger
	if err := l.zapLogger.Sync(); err != nil && err.Error() != "sync /dev/stdout: The handle is invalid" {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	// Close lumberjack logger to release file handle
	if l.fileLogger != nil {
		if err := l.fileLogger.Close(); err != nil {
			return fmt.Errorf("failed to close file logger: %w", err)
		}
	}

	return nil
}

// CompressLogFile compresses the log file with gzip
func (l *Logger) CompressLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := os.Stat(l.filePath); err != nil {
		return fmt.Errorf("log file not found: %w", err)
	}

	// Create compressed file path
	compressedPath := l.filePath + ".gz"

	// Open source file
	source, err := os.Open(l.filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	// Create destination file
	destination, err := os.Create(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer destination.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(destination)
	defer gzipWriter.Close()

	// Copy content
	if _, err := io.Copy(gzipWriter, source); err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	// Flush gzip writer
	if err := gzipWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush gzip writer: %w", err)
	}

	return nil
}

// ReadCompressedLogs reads and decompresses logs from a gzip file
func ReadCompressedLogs(filePath string) ([]map[string]interface{}, error) {
	// Open compressed file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Decode JSON lines directly from gzip stream
	var logs []map[string]interface{}
	decoder := json.NewDecoder(gzipReader)

	for decoder.More() {
		var logEntry map[string]interface{}
		if err := decoder.Decode(&logEntry); err != nil {
			continue // Skip malformed lines
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

// ReadCompressedLogsFiltered reads and filters logs from a gzip file
func ReadCompressedLogsFiltered(filePath string, filter FilterFunc) ([]map[string]interface{}, error) {
	logs, err := ReadCompressedLogs(filePath)
	if err != nil {
		return nil, err
	}

	var filtered []map[string]interface{}
	for _, log := range logs {
		if filter(log) {
			filtered = append(filtered, log)
		}
	}

	return filtered, nil
}

// FilterFunc is a function type for filtering logs
type FilterFunc func(log map[string]interface{}) bool

// FilterByLevel creates a filter for a specific log level
func FilterByLevel(level string) FilterFunc {
	return func(log map[string]interface{}) bool {
		if l, ok := log["level"]; ok {
			return l == level
		}
		return false
	}
}

// FilterByTimeRange creates a filter for logs within a time range
func FilterByTimeRange(start, end time.Time) FilterFunc {
	return func(log map[string]interface{}) bool {
		if ts, ok := log["timestamp"].(string); ok {
			// Try RFC3339Nano first (with colon in timezone)
			t, err := time.Parse(time.RFC3339Nano, ts)
			if err != nil {
				// Try the format without colon in timezone: 2025-12-02T15:59:57.317+0800
				t, err = time.Parse("2006-01-02T15:04:05.000-0700", ts)
				if err != nil {
					// Try with 3-digit milliseconds
					t, err = time.Parse("2006-01-02T15:04:05.000Z0700", ts)
					if err != nil {
						return false
					}
				}
			}
			return t.After(start) && t.Before(end)
		}
		return false
	}
}
