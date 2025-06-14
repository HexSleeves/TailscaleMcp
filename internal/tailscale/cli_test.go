package tailscale

import (
	"context"
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

func TestCLIIntegration(t *testing.T) {
	cli := setupCliTest(t)

	t.Run("Security Validation", func(t *testing.T) {
		tests := []struct {
			name        string
			args        []string
			expectError bool
			errorMsg    string
		}{
			{
				name:        "empty command",
				args:        []string{},
				expectError: true,
				errorMsg:    "no command specified",
			},
			{
				name:        "disallowed command",
				args:        []string{"rm", "-rf", "/"},
				expectError: true,
				errorMsg:    "command 'rm' not allowed",
			},
			{
				name:        "command injection attempt",
				args:        []string{"status", "; rm -rf /"},
				expectError: true,
				errorMsg:    "command argument contains invalid characters",
			},
			{
				name:        "pipe injection attempt",
				args:        []string{"status", "| cat /etc/passwd"},
				expectError: true,
				errorMsg:    "command argument contains invalid characters",
			},
			{
				name:        "backtick injection attempt",
				args:        []string{"status", "`whoami`"},
				expectError: true,
				errorMsg:    "command argument contains invalid characters",
			},
			{
				name:        "dollar injection attempt",
				args:        []string{"status", "$(whoami)"},
				expectError: true,
				errorMsg:    "command argument contains invalid characters",
			},
			{
				name:        "argument too long",
				args:        []string{"status", strings.Repeat("a", 1001)},
				expectError: true,
				errorMsg:    "command argument too long",
			},
			{
				name:        "valid command",
				args:        []string{"version"},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp := cli.ExecuteCommand(context.Background(), tt.args)

				if tt.expectError {
					assert.False(t, resp.Success)
					assert.Contains(t, resp.Error, tt.errorMsg)
				} else {
					// For valid commands, we don't care if they succeed or fail
					// (depends on Tailscale being installed/running)
					// We just care that they pass validation
					if !resp.Success {
						// If it failed, it should be due to execution, not validation
						assert.NotContains(t, resp.Error, "not allowed")
						assert.NotContains(t, resp.Error, "invalid characters")
						assert.NotContains(t, resp.Error, "too long")
					}
				}
			})
		}
	})

	t.Run("CLI Method Functionality", func(t *testing.T) {
		t.Run("GetVersion", func(t *testing.T) {
			resp := cli.GetVersion()
			// Should either succeed or fail with a reasonable error
			if !resp.Success {
				assert.NotEmpty(t, resp.Error)
			}
		})

		t.Run("Ping Validation", func(t *testing.T) {
			// Test empty target validation
			resp := cli.Ping("", 4) // Pass count explicitly
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, "invalid target specified")

			// Test with valid target (may fail if not connected, but should pass validation)
			resp = cli.Ping("100.64.0.1", 1) // Ping once
			if !resp.Success {
				// Should fail due to execution, not validation
				assert.NotContains(t, resp.Error, "invalid target specified")
			}

			// Test ping count validation
			resp = cli.Ping("100.64.0.1", 0)
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, "count must be an integer between 1 and 100")

			resp = cli.Ping("100.64.0.1", 101)
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, "count must be an integer between 1 and 100")
		})

		t.Run("IP", func(t *testing.T) {
			resp := cli.IP()
			if !resp.Success {
				assert.NotEmpty(t, resp.Error)
			}
		})

		t.Run("Netcheck", func(t *testing.T) {
			resp := cli.Netcheck()
			if !resp.Success {
				assert.NotEmpty(t, resp.Error)
			}
		})

		t.Run("Up", func(t *testing.T) {
			// Test with nil options (should attempt to bring up, expect external error)
			resp := cli.Up(nil)
			assert.False(t, resp.Success)
			assert.NotEmpty(t, resp.Error)
			// Expect errors related to daemon connection or authentication, not validation
			assert.NotContains(t, resp.Error, "invalid character")
			assert.NotContains(t, resp.Error, "too long")

			// Test with structured options (should attempt to bring up, expect external error)
			resp = cli.Up(&UpOptions{AcceptRoutes: true})
			assert.False(t, resp.Success)
			assert.NotEmpty(t, resp.Error)
			// Expect errors related to daemon connection or authentication, not validation
			assert.NotContains(t, resp.Error, "invalid character")
			assert.NotContains(t, resp.Error, "too long")

			// Test Up with hostname validation
			resp = cli.Up(&UpOptions{Hostname: "host;evil"})
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, "invalid character ';'")

			// Test Up with advertise-routes validation
			resp = cli.Up(&UpOptions{AdvertiseRoutes: []string{"invalid-route"}})
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, "invalid route format")

			// Test Up with authKey - validation only, actual execution will likely fail
			resp = cli.Up(&UpOptions{AuthKey: "tskey-dummykey-abcdef"})
			assert.False(t, resp.Success)
			assert.NotEmpty(t, resp.Error)
			assert.NotContains(t, resp.Error, "invalid character")
			assert.NotContains(t, resp.Error, "too long")
		})

		t.Run("Down", func(t *testing.T) {
			resp := cli.Down()
			if !resp.Success {
				assert.NotEmpty(t, resp.Error)
			}
		})

		t.Run("Logout", func(t *testing.T) {
			resp := cli.Logout()
			if !resp.Success {
				assert.NotEmpty(t, resp.Error)
			}
		})
	})

	t.Run("Command Whitelist", func(t *testing.T) {
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
	})

	t.Run("ExecuteCommand Limits", func(t *testing.T) {
		// Test that the function handles context properly
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		resp := cli.ExecuteCommand(ctx, []string{"version"})
		// Should handle the cancelled context gracefully
		if !resp.Success {
			assert.NotEmpty(t, resp.Error)
		}
	})
}
