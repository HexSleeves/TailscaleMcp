package config

import (
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
	HTTPPort   int    `json:"http_port"`
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
		HTTPPort:            8080,
	}

	// Load from environment variables
	if apiKey := os.Getenv("TAILSCALE_API_KEY"); apiKey != "" {
		cfg.TailscaleAPIKey = apiKey
	}

	if tailnet := os.Getenv("TAILSCALE_TAILNET"); tailnet != "" {
		cfg.TailscaleTailnet = tailnet
	}

	if baseURL := os.Getenv("TAILSCALE_API_BASE_URL"); baseURL != "" {
		cfg.TailscaleAPIBaseURL = baseURL
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		if level, err := strconv.Atoi(logLevel); err == nil {
			if level >= 0 && level <= 3 {
				cfg.LogLevel = level
			}
		}
	}

	if logFile := os.Getenv("MCP_SERVER_LOG_FILE"); logFile != "" {
		cfg.LogFile = logFile
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Note: API key and tailnet are not required for CLI-only operations
	// They will be validated when needed by specific tools
	return nil
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
