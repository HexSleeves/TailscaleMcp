package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var globalLogger *slog.Logger

// LogLevel represents the logging level
type LogLevel int

const (
	LevelDebug LogLevel = 0
	LevelInfo  LogLevel = 1
	LevelWarn  LogLevel = 2
	LevelError LogLevel = 3
)

// Initialize sets up the global logger with the specified level and optional file output
func Initialize(level int, logFile string) error {
	var slogLevel slog.Level
	switch LogLevel(level) {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time to be more readable
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	var writer io.Writer = os.Stderr

	// If log file is specified, write to both file and stderr
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logFile, err)
		}
		writer = io.MultiWriter(os.Stderr, file)
	}

	handler := slog.NewJSONHandler(writer, opts)
	globalLogger = slog.New(handler)

	return nil
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Warn(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

// Fatal logs a fatal error message and exits
func Fatal(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
	os.Exit(1)
}

// With returns a logger with the given attributes
func With(args ...any) *slog.Logger {
	if globalLogger != nil {
		return globalLogger.With(args...)
	}
	return slog.Default().With(args...)
}

// GetLogger returns the global logger instance
func GetLogger() *slog.Logger {
	if globalLogger != nil {
		return globalLogger
	}
	return slog.Default()
}
