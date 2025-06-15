package integration

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestHTTPModeGracefulShutdownExitCode verifies that the HTTP server exits with code 0
// when it receives a graceful shutdown signal (SIGTERM or SIGINT)
func TestHTTPModeGracefulShutdownExitCode(t *testing.T) {
	// Build the server binary for testing
	buildCmd := exec.Command("go", "build", "-o", "test-server", "../../cmd/tailscale-mcp-server")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test server: %v", err)
	}
	defer os.Remove("test-server")

	tests := []struct {
		name   string
		signal os.Signal
	}{
		{"SIGTERM http mode", syscall.SIGTERM},
		{"SIGINT http mode", syscall.SIGINT},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use different ports for each test to avoid conflicts
			port := 9080 + i

			// Start the HTTP server
			cmd := exec.Command("./test-server", "serve", "--mode=http", "--port="+fmt.Sprintf("%d", port))

			// Start the process
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}

			// Give the server time to start
			time.Sleep(500 * time.Millisecond)

			// Send the signal
			if err := cmd.Process.Signal(tt.signal); err != nil {
				t.Fatalf("Failed to send signal: %v", err)
			}

			// Wait for the process to exit with a timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case err := <-done:
				// Check the exit code
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						exitCode := exitError.ExitCode()
						if exitCode != 0 {
							t.Errorf("Expected exit code 0 for graceful shutdown, got %d", exitCode)
						}
					} else {
						t.Errorf("Unexpected error type: %v", err)
					}
				}
				// If err is nil, the process exited with code 0, which is what we want
			case <-time.After(5 * time.Second):
				// Force kill if it doesn't exit gracefully
				cmd.Process.Kill()
				t.Error("Server did not exit within timeout after receiving signal")
			}
		})
	}
}

// TestStdioModeNormalExit verifies that stdio mode exits with code 0
// when stdin is closed (normal MCP client disconnect scenario)
func TestStdioModeNormalExit(t *testing.T) {
	// Build the server binary for testing
	buildCmd := exec.Command("go", "build", "-o", "test-server", "../../cmd/tailscale-mcp-server")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test server: %v", err)
	}
	defer os.Remove("test-server")

	// Start the stdio server with no stdin (simulates client disconnect)
	cmd := exec.Command("./test-server", "serve", "--mode=stdio")

	// Run the command and check exit code
	err := cmd.Run()

	// The server should exit with code 0 when stdin closes
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if exitCode != 0 {
				t.Errorf("Expected exit code 0 when stdin closes, got %d", exitCode)
			}
		} else {
			t.Errorf("Unexpected error type: %v", err)
		}
	}
	// If err is nil, the process exited with code 0, which is what we want
}

// TestErrorExitCode verifies that the server exits with code 1 for actual errors
func TestErrorExitCode(t *testing.T) {
	// Build the server binary for testing
	buildCmd := exec.Command("go", "build", "-o", "test-server", "../../cmd/tailscale-mcp-server")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test server: %v", err)
	}
	defer os.Remove("test-server")

	tests := []struct {
		name string
		args []string
	}{
		{"invalid mode", []string{"serve", "--mode=invalid"}},
		{"invalid port", []string{"serve", "--mode=http", "--port=99999"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./test-server", tt.args...)

			// Run the command
			err := cmd.Run()

			// We expect an error (non-zero exit code)
			if err == nil {
				t.Error("Expected command to fail with non-zero exit code, but it succeeded")
				return
			}

			// Check that it's an exit error with code 1
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != 1 {
					t.Errorf("Expected exit code 1 for error condition, got %d", exitError.ExitCode())
				}
			} else {
				t.Errorf("Expected ExitError, got: %v", err)
			}
		})
	}
}
