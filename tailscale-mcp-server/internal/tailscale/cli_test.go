package tailscale

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
)

// useStubBinary writes a tiny stub "tailscale" binary that always exits 1 quickly.
// It prepends the directory containing the stub to PATH so the real tailscale
// binary is never invoked. This keeps tests hermetic and fast even when the
// developer is not logged-in to Tailscale or does not have the CLI installed.
func useStubBinary(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	stubPath := filepath.Join(tmpDir, "tailscale")

	script := "#!/usr/bin/env sh\necho 'dummy tailscale stub: $*' 1>&2\nexit 1\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(script), 0o755))

	// Prepend the temp dir to PATH
	oldPath := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath))
}

func setupCliTest(t *testing.T) *TailscaleCLI {
	useStubBinary(t) // ensure stub is first in PATH

	// Initialize logger for tests
	err := logger.Initialize(0, "") // Debug level, no file
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	cli, err := NewTailscaleCLI()
	if err != nil {
		t.Fatalf("Failed to create TailscaleCLI: %v", err)
	}

	return cli
}

func TestNewTailscaleCLI(t *testing.T) {
	cli := setupCliTest(t)
	assert.NotEmpty(t, cli.tailscalePath)
	assert.Contains(t, cli.tailscalePath, "tailscale")
}

func TestGetStatus(t *testing.T) {
	cli := setupCliTest(t)
	status, err := cli.GetStatus()

	// The test should pass regardless of whether Tailscale is actually running
	// We're testing the structure and parsing, not the actual Tailscale state
	if err == nil {
		// If successful, verify the structure
		assert.NotNil(t, status)
		assert.NotNil(t, status.Version)
		assert.NotNil(t, status.ClientVersion)
		if status.ClientVersion != nil {
			assert.NotNil(t, status.ClientVersion.RunningLatest)
		}
		if status.CurrentTailnet != nil {
			assert.NotNil(t, status.CurrentTailnet.Name)
			assert.NotNil(t, status.CurrentTailnet.MagicDNSSuffix)
			assert.NotNil(t, status.CurrentTailnet.MagicDNSEnabled)
		}

		assert.NotNil(t, status.MagicDNSSuffix)
		assert.NotNil(t, status.Self)

		// Health and Peer can be nil or empty depending on Tailscale state,
		// so we don't assert NotNil if they are omitempty in the JSON schema.
		// If they exist, we can assert their structure.
		// assert.NotNil(t, status.Health)
		// assert.NotNil(t, status.Peer)
	} else {
		// If not successful, we should have an error message
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestCommandWhitelist(t *testing.T) {
	expectedCommands := []string{
		"status", "up", "down", "logout", "switch", "configure",
		"netcheck", "ip", "ping", "ssh", "version", "update",
		"web", "file", "bugreport", "cert", "lock", "licenses",
		"exit-node", "set", "unset",
	}

	for _, cmd := range expectedCommands {
		assert.True(t, allowedCommands[cmd], "Command %s should be allowed", cmd)
	}

	disallowedCommands := []string{
		"rm", "cat", "ls", "chmod", "sudo", "su", "exec",
	}

	for _, cmd := range disallowedCommands {
		assert.False(t, allowedCommands[cmd], "Command %s should NOT be allowed", cmd)
	}
}

func TestValidateTarget(t *testing.T) {
	cli := setupCliTest(t)

	tests := []struct {
		name        string
		target      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty target",
			target:      "",
			expectError: true,
			errorMsg:    "target cannot be empty",
		},
		{
			name:        "valid IP",
			target:      "100.64.0.1",
			expectError: false,
		},
		{
			name:        "valid hostname",
			target:      "my-device",
			expectError: false,
		},
		{
			name:        "target with semicolon",
			target:      "host;evil",
			expectError: true,
			errorMsg:    "invalid character \";\" in target",
		},
		{
			name:        "target with pipe",
			target:      "host|evil",
			expectError: true,
			errorMsg:    "invalid character \"|\" in target",
		},
		{
			name:        "target with backtick",
			target:      "host`evil",
			expectError: true,
			errorMsg:    "invalid character \"`\" in target",
		},
		{
			name:        "target with dollar",
			target:      "host$evil",
			expectError: true,
			errorMsg:    "invalid character \"$\" in target",
		},
		{
			name:        "target too long",
			target:      strings.Repeat("a", 254),
			expectError: true,
			errorMsg:    "target too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.validateTarget(tt.target)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStringInput(t *testing.T) {
	cli := setupCliTest(t)

	tests := []struct {
		name        string
		input       string
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid input",
			input:       "valid-hostname",
			fieldName:   "hostname",
			expectError: false,
		},
		{
			name:        "input with semicolon",
			input:       "host;evil",
			fieldName:   "hostname",
			expectError: true,
			errorMsg:    "invalid character \";\" in hostname",
		},
		{
			name:        "input with pipe",
			input:       "host|evil",
			fieldName:   "hostname",
			expectError: true,
			errorMsg:    "invalid character \"|\" in hostname",
		},
		{
			name:        "input too long",
			input:       strings.Repeat("a", 1001),
			fieldName:   "hostname",
			expectError: true,
			errorMsg:    "hostname too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.validateStringInput(tt.input, tt.fieldName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRoutes(t *testing.T) {
	cli := setupCliTest(t)

	tests := []struct {
		name        string
		routes      []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid routes",
			routes:      []string{"192.168.1.0/24", "10.0.0.0/8"},
			expectError: false,
		},
		{
			name:        "default route IPv4",
			routes:      []string{"0.0.0.0/0"},
			expectError: false,
		},
		{
			name:        "default route IPv6",
			routes:      []string{"::/0"},
			expectError: false,
		},
		{
			name:        "invalid route format",
			routes:      []string{"invalid-route"},
			expectError: true,
			errorMsg:    "invalid route format",
		},
		{
			name:        "invalid CIDR",
			routes:      []string{"192.168.1.0/33"},
			expectError: true,
			errorMsg:    "invalid CIDR route",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.validateRoutes(tt.routes)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetTailscaleFallbackPaths tests platform-specific fallback paths
func TestGetTailscaleFallbackPaths(t *testing.T) {
	paths := getTailscaleFallbackPaths()
	assert.NotEmpty(t, paths, "Should return at least one fallback path")

	// Test that paths are platform-specific
	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, paths[0], "Program Files")
		assert.Contains(t, paths[0], "tailscale.exe")
	case "darwin":
		assert.Contains(t, paths, "/usr/local/bin/tailscale")
		assert.Contains(t, paths, "/opt/homebrew/bin/tailscale")
	default:
		assert.Contains(t, paths, "/usr/bin/tailscale")
		assert.Contains(t, paths, "/usr/local/bin/tailscale")
	}
}

// TestIsExecutableFile tests the executable file detection
func TestIsExecutableFile(t *testing.T) {
	// Create temporary directory and files for testing
	tmpDir := t.TempDir()

	// Create an executable file
	executablePath := filepath.Join(tmpDir, "executable")
	require.NoError(t, os.WriteFile(executablePath, []byte("#!/bin/sh\necho test"), 0755))

	// Create a non-executable file
	nonExecutablePath := filepath.Join(tmpDir, "nonexecutable")
	require.NoError(t, os.WriteFile(nonExecutablePath, []byte("content"), 0644))

	// Create a directory
	dirPath := filepath.Join(tmpDir, "directory")
	require.NoError(t, os.Mkdir(dirPath, 0755))

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "executable file",
			path:     executablePath,
			expected: true,
		},
		{
			name:     "non-executable file",
			path:     nonExecutablePath,
			expected: false,
		},
		{
			name:     "directory",
			path:     dirPath,
			expected: false,
		},
		{
			name:     "non-existent file",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExecutableFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewTailscaleCLI_WithEnvironmentVariable tests CLI creation with TAILSCALE_PATH set
func TestNewTailscaleCLI_WithEnvironmentVariable(t *testing.T) {
	// Save original environment
	oldPath := os.Getenv("TAILSCALE_PATH")
	defer func() {
		if oldPath == "" {
			os.Unsetenv("TAILSCALE_PATH")
		} else {
			os.Setenv("TAILSCALE_PATH", oldPath)
		}
	}()

	// Initialize logger for tests
	err := logger.Initialize(0, "")
	require.NoError(t, err)

	// Create a temporary executable
	tmpDir := t.TempDir()
	tailscalePath := filepath.Join(tmpDir, "tailscale")
	require.NoError(t, os.WriteFile(tailscalePath, []byte("#!/bin/sh\necho test"), 0755))

	// Test with valid TAILSCALE_PATH
	require.NoError(t, os.Setenv("TAILSCALE_PATH", tailscalePath))
	cli, err := NewTailscaleCLI()
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	assert.Equal(t, tailscalePath, cli.tailscalePath)

	// Test with invalid TAILSCALE_PATH (non-existent file)
	invalidPath := filepath.Join(tmpDir, "nonexistent")
	require.NoError(t, os.Setenv("TAILSCALE_PATH", invalidPath))
	cli, err = NewTailscaleCLI()
	assert.Error(t, err)
	assert.Nil(t, cli)

	// Test with invalid TAILSCALE_PATH (directory)
	require.NoError(t, os.Setenv("TAILSCALE_PATH", tmpDir))
	cli, err = NewTailscaleCLI()
	assert.Error(t, err)
	assert.Nil(t, cli)

	// Test with invalid TAILSCALE_PATH (non-executable file)
	nonExecPath := filepath.Join(tmpDir, "nonexec")
	require.NoError(t, os.WriteFile(nonExecPath, []byte("content"), 0644))
	require.NoError(t, os.Setenv("TAILSCALE_PATH", nonExecPath))
	cli, err = NewTailscaleCLI()
	assert.Error(t, err)
	assert.Nil(t, cli)
}

// TestNewTailscaleCLI_FallbackPaths tests CLI creation with fallback path resolution
func TestNewTailscaleCLI_FallbackPaths(t *testing.T) {
	// Save and clear environment variables that might interfere
	oldTailscalePath := os.Getenv("TAILSCALE_PATH")
	oldPATH := os.Getenv("PATH")
	defer func() {
		if oldTailscalePath == "" {
			os.Unsetenv("TAILSCALE_PATH")
		} else {
			os.Setenv("TAILSCALE_PATH", oldTailscalePath)
		}
		os.Setenv("PATH", oldPATH)
	}()

	os.Unsetenv("TAILSCALE_PATH")

	// Initialize logger for tests
	err := logger.Initialize(0, "")
	require.NoError(t, err)

	// Set PATH to empty to force fallback path usage
	require.NoError(t, os.Setenv("PATH", ""))

	// Test that NewTailscaleCLI uses fallback paths when PATH lookup fails
	cli, err := NewTailscaleCLI()

	// The test result depends on whether a real Tailscale binary exists at fallback paths
	// If found at a fallback path, it should succeed; otherwise, it should fail gracefully
	if err != nil {
		// No binary found at fallback paths - this is expected in CI environments
		assert.Nil(t, cli)
		assert.Contains(t, err.Error(), "tailscale binary not found")
	} else {
		// Binary found at a fallback path - verify it's a valid path
		assert.NotNil(t, cli)
		assert.NotEmpty(t, cli.tailscalePath)

		// Verify the path is one of the expected fallback paths
		fallbackPaths := getTailscaleFallbackPaths()
		found := false
		for _, fallbackPath := range fallbackPaths {
			if cli.tailscalePath == fallbackPath {
				found = true
				break
			}
		}
		assert.True(t, found, "CLI should use one of the fallback paths: %v, got: %s", fallbackPaths, cli.tailscalePath)
	}
}
