package tailscale

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	resp := cli.GetStatus()

	// The test should pass regardless of whether Tailscale is actually running
	// We're testing the structure and parsing, not the actual Tailscale state
	if resp.Success {
		status := resp.Data

		// If successful, verify the structure
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
		assert.NotEmpty(t, resp.Error)
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
			errorMsg:    "invalid target specified",
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
			errorMsg:    "invalid character ';'",
		},
		{
			name:        "target with pipe",
			target:      "host|evil",
			expectError: true,
			errorMsg:    "invalid character '|'",
		},
		{
			name:        "target with backtick",
			target:      "host`evil",
			expectError: true,
			errorMsg:    "invalid character '`'",
		},
		{
			name:        "target with dollar",
			target:      "host$evil",
			expectError: true,
			errorMsg:    "invalid character '$'",
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
			errorMsg:    "invalid character ';'",
		},
		{
			name:        "input with pipe",
			input:       "host|evil",
			fieldName:   "hostname",
			expectError: true,
			errorMsg:    "invalid character '|'",
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
