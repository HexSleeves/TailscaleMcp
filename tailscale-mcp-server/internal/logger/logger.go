package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	initialized  bool
	globalLogger *zap.Logger
	loggerMutex  sync.RWMutex
)

// Initialize sets up the global logger with the specified level and optional file output
func Initialize(level int, logFile string) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found or unable to load:", err)
	}

	// Determine if we're in development mode
	isDev := strings.ToLower(os.Getenv("ENVIRONMENT")) == "development"

	var config zap.Config
	if isDev {
		config = zap.NewDevelopmentConfig()
		// Custom time format for development
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
	} else {
		config = zap.NewProductionConfig()
	}

	// Set log level from environment variable or parameter
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		var zapLevel zapcore.Level
		if err := zapLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
			fmt.Printf("Invalid LOG_LEVEL environment variable '%s', defaulting to info: %v\n", logLevelStr, err)
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		} else {
			config.Level = zap.NewAtomicLevelAt(zapLevel)
		}
	} else {
		// Use the passed level parameter
		var zapLevel zapcore.Level
		switch level {
		case 0:
			zapLevel = zap.DebugLevel
		case 1:
			zapLevel = zap.InfoLevel
		case 2:
			zapLevel = zap.WarnLevel
		case 3:
			zapLevel = zap.ErrorLevel
		default:
			zapLevel = zap.InfoLevel
		}
		config.Level = zap.NewAtomicLevelAt(zapLevel)
	}

	// ---------------------------------------------------------------------
	// Encoder / format configuration
	// ---------------------------------------------------------------------
	// Human-readable logs are now the default using zap's console encoder.
	// If structured JSON output is explicitly desired, set LOG_FORMAT=json.
	// Any other value (including empty) results in console encoding.

	switch strings.ToLower(os.Getenv("LOG_FORMAT")) {
	case "json":
		config.Encoding = "json"
	default:
		config.Encoding = "console"
	}

	// Improve console readability with colored, capital levels irrespective
	// of dev / prod mode.go test ./... | cat
	if config.Encoding == "console" {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// Align time formatting across modes for consistency.
		if config.EncoderConfig.EncodeTime == nil {
			config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
		}
	}

	// Set output paths
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}

	if logFile != "" {
		config.OutputPaths = append(config.OutputPaths, logFile)
		config.ErrorOutputPaths = append(config.ErrorOutputPaths, logFile)
	}

	// ---------------------------------------------------------------------
	// Build the final logger
	// ---------------------------------------------------------------------
	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	globalLogger = logger
	initialized = true

	return nil
}

// Cleanup properly closes the logger and flushes any buffered log entries
func Cleanup() error {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// isInitialized checks if the logger has been initialized
func isInitialized() bool {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return initialized && globalLogger != nil
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if !isInitialized() {
		fmt.Fprintf(os.Stderr, "DEBUG (logger not initialized): %s\n", msg)
		return
	}
	globalLogger.Debug(msg, convertArgsToZapFields(args)...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	if !isInitialized() {
		fmt.Fprintf(os.Stderr, "INFO (logger not initialized): %s\n", msg)
		return
	}
	globalLogger.Info(msg, convertArgsToZapFields(args)...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if !isInitialized() {
		fmt.Fprintf(os.Stderr, "WARN (logger not initialized): %s\n", msg)
		return
	}
	globalLogger.Warn(msg, convertArgsToZapFields(args)...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	if !isInitialized() {
		fmt.Fprintf(os.Stderr, "ERROR (logger not initialized): %s\n", msg)
		return
	}
	globalLogger.Error(msg, convertArgsToZapFields(args)...)
}

// Fatal logs a fatal error message and exits
func Fatal(msg string, args ...any) {
	if isInitialized() {
		globalLogger.Fatal(msg, convertArgsToZapFields(args)...)
		// Fatal calls os.Exit(1) internally
	} else {
		// Fallback if logger is nil
		fmt.Fprintf(os.Stderr, "FATAL (logger not initialized): %s\n", msg)
		os.Exit(1)
	}
}

// With returns a logger with the given attributes
func With(args ...any) *zap.Logger {
	if isInitialized() {
		return globalLogger.With(convertArgsToZapFields(args)...)
	}
	return zap.NewNop().With(convertArgsToZapFields(args)...)
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if isInitialized() {
		return globalLogger
	}
	return zap.NewNop()
}

// Sugar returns a sugared logger for more flexible logging
func Sugar() *zap.SugaredLogger {
	if isInitialized() {
		return globalLogger.Sugar()
	}
	return zap.NewNop().Sugar()
}

// Sugared logger convenience methods
func Debugf(template string, args ...any) {
	Sugar().Debugf(template, args...)
}

func Infof(template string, args ...any) {
	Sugar().Infof(template, args...)
}

func Warnf(template string, args ...any) {
	Sugar().Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	Sugar().Errorf(template, args...)
}

func Fatalf(template string, args ...any) {
	Sugar().Fatalf(template, args...)
}

// Debugw logs a debug message with key-value pairs (sugared)
func Debugw(msg string, keysAndValues ...any) {
	Sugar().Debugw(msg, keysAndValues...)
}

// Infow logs an info message with key-value pairs (sugared)
func Infow(msg string, keysAndValues ...any) {
	Sugar().Infow(msg, keysAndValues...)
}

// Warnw logs a warning message with key-value pairs (sugared)
func Warnw(msg string, keysAndValues ...any) {
	Sugar().Warnw(msg, keysAndValues...)
}

// Errorw logs an error message with key-value pairs (sugared)
func Errorw(msg string, keysAndValues ...any) {
	Sugar().Errorw(msg, keysAndValues...)
}

// Fatalw logs a fatal error message with key-value pairs (sugared)
func Fatalw(msg string, keysAndValues ...any) {
	Sugar().Fatalw(msg, keysAndValues...)
}

// Helper function to convert variadic arguments to zap.Field slices
func convertArgsToZapFields(args []any) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	var fields []zap.Field
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			// Odd number of args - log the orphaned key
			fields = append(fields, zap.Any("orphaned_key", args[i]))
			break
		}

		key, ok := args[i].(string)
		if !ok {
			// Non-string key - convert to string
			key = fmt.Sprintf("%v", args[i])
		}

		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}
