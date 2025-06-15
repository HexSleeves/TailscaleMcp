// tailscale-mcp-server/internal/config/config.go
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the server
type Config struct {
	// Tailscale configuration
	TailscaleAPIKey     string `json:"tailscale_api_key"`
	TailscaleTailnet    string `json:"tailscale_tailnet"`
	TailscaleAPIBaseURL string `json:"tailscale_api_base_url"`

	// Logging configuration
	LogLevel int    `json:"log_level"`
	LogFile  string `json:"log_file"`

	// Server configuration
	ServerMode string `json:"server_mode"`
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("invalid %s value %q: %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file if it exists (ignore errors)
	_ = godotenv.Load()

	cfg := &Config{
		// Default values
		TailscaleAPIBaseURL: "https://api.tailscale.com",
		LogLevel:            1, // INFO level
		ServerMode:          "stdio",
	}

	// Load from environment variables
	if apiKey := strings.TrimSpace(os.Getenv("TAILSCALE_API_KEY")); apiKey != "" {
		cfg.TailscaleAPIKey = apiKey
	}

	if tailnet := strings.TrimSpace(os.Getenv("TAILSCALE_TAILNET")); tailnet != "" {
		cfg.TailscaleTailnet = tailnet
	}

	if baseURL := strings.TrimSpace(os.Getenv("TAILSCALE_API_BASE_URL")); baseURL != "" {
		cfg.TailscaleAPIBaseURL = baseURL
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		if level, err := strconv.Atoi(logLevel); err == nil {
			if level >= 0 && level <= 3 {
				cfg.LogLevel = level
			}
		}
	}

	if logFile := strings.TrimSpace(os.Getenv("MCP_SERVER_LOG_FILE")); logFile != "" {
		cfg.LogFile = logFile
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	var errors []error

	// Validate API base URL format
	if c.TailscaleAPIBaseURL != "" {
		if _, err := url.Parse(c.TailscaleAPIBaseURL); err != nil {
			errors = append(errors, &ValidationError{
				Field:   "TailscaleAPIBaseURL",
				Value:   c.TailscaleAPIBaseURL,
				Message: "must be a valid URL",
			})
		}
	}

	// Validate log level range
	if c.LogLevel < 0 || c.LogLevel > 3 {
		errors = append(errors, &ValidationError{
			Field:   "LogLevel",
			Value:   strconv.Itoa(c.LogLevel),
			Message: "must be between 0 (debug) and 3 (error)",
		})
	}

	// Validate server mode
	validModes := []string{"stdio", "http"}
	if !contains(validModes, c.ServerMode) {
		errors = append(errors, &ValidationError{
			Field:   "ServerMode",
			Value:   c.ServerMode,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")),
		})
	}

	// Validate log file path if specified
	if c.LogFile != "" {
		if !isValidLogPath(c.LogFile) {
			errors = append(errors, &ValidationError{
				Field:   "LogFile",
				Value:   c.LogFile,
				Message: "path appears invalid or not writable",
			})
		}
	}

	if len(errors) > 0 {
		return &MultiValidationError{Errors: errors}
	}

	return nil
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []error
}

func (e *MultiValidationError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// HasAPICredentials returns true if API credentials are configured
func (c *Config) HasAPICredentials() bool {
	return c.TailscaleAPIKey != "" && c.TailscaleTailnet != ""
}

// LogLevelString returns the log level as a string
func (c *Config) LogLevelString() string {
	switch c.LogLevel {
	case 0:
		return "debug"
	case 1:
		return "info"
	case 2:
		return "warn"
	case 3:
		return "error"
	default:
		return "info"
	}
}

// SanitizedCopy returns a copy of the config with sensitive fields redacted
// This is useful for logging configuration without exposing secrets
func (c *Config) SanitizedCopy() *Config {
	copy := *c
	if copy.TailscaleAPIKey != "" {
		copy.TailscaleAPIKey = redactSecret(copy.TailscaleAPIKey)
	}
	return &copy
}

// redactSecret redacts a secret string for safe logging
func redactSecret(secret string) string {
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isValidLogPath performs basic validation of log file path
func isValidLogPath(path string) bool {
	// Basic path validation - check if directory exists or can be created
	dir := strings.TrimSuffix(path, "/")
	if idx := strings.LastIndex(dir, "/"); idx > 0 {
		dir = dir[:idx]
	} else {
		dir = "."
	}

	// Check if directory exists or is writable
	if stat, err := os.Stat(dir); err != nil {
		return false
	} else if !stat.IsDir() {
		return false
	}

	return true
}
