// tailscale-mcp-server/internal/tailscale/cli.go
package tailscale

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/pkg/schema"
)

const (
	maxArgLen  = 1000
	maxBufSize = 10 * 1024 * 1024 // 10 MB
	timeout    = 30 * time.Second

	// DNS hostname max length
	maxHostnameLen = 253

	// Ping count limits
	minPingCount = 1
	maxPingCount = 100
)

// Validation patterns
var (
	// VALID_TARGET_PATTERN validates hostname, IP, or Tailscale node name
	// Hostname/IP pattern: no leading/trailing dots or hyphens, no consecutive dots
	validTargetPattern = regexp.MustCompile(`^(([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*)|([0-9a-fA-F:]+))$`)

	// CIDR validation pattern
	cidrPattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$|^([0-9a-fA-F:]+)\/\d{1,3}$`)
)

// allowedCommands defines the whitelist of allowed Tailscale CLI commands for security
var allowedCommands = map[string]bool{
	"status":    true,
	"up":        true,
	"down":      true,
	"logout":    true,
	"switch":    true,
	"configure": true,
	"netcheck":  true,
	"ip":        true,
	"ping":      true,
	"ssh":       true,
	"version":   true,
	"update":    true,
	"web":       true,
	"file":      true,
	"bugreport": true,
	"cert":      true,
	"lock":      true,
	"licenses":  true,
	"exit-node": true,
	"set":       true,
	"unset":     true,
}

type TailscaleCLI struct {
	tailscalePath string
}

// CLIError represents an error that occurred during CLI execution
type CLIError struct {
	Command    string
	Args       []string
	ExitCode   int
	Stderr     string
	Underlying error
}

func (e *CLIError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("tailscale %s failed: %v", e.Command, e.Underlying)
	}
	return fmt.Sprintf("tailscale %s failed with exit code %d: %s", e.Command, e.ExitCode, e.Stderr)
}

func (e *CLIError) Unwrap() error {
	return e.Underlying
}

func NewTailscaleCLI() (*TailscaleCLI, error) {
	path := os.Getenv("TAILSCALE_PATH")

	switch {
	// 1. Env-provided path: expand & validate.
	case path != "":
		abs, err := filepath.Abs(path)
		if err != nil {
			logger.Error("invalid TAILSCALE_PATH", "path", path, "error", err)
			return nil, fmt.Errorf("invalid TAILSCALE_PATH %q: %w", path, err)
		}

		// Make sure it actually exists & is executable.
		if st, err := os.Stat(abs); err != nil {
			logger.Error("TAILSCALE_PATH does not exist", "path", abs, "error", err)
			return nil, fmt.Errorf("TAILSCALE_PATH does not exist: %s", abs)
		} else if st.IsDir() {
			err := fmt.Errorf("TAILSCALE_PATH is a directory, not an executable: %s", abs)
			logger.Error("TAILSCALE_PATH validation failed", "error", err)
			return nil, err
		} else if st.Mode()&0111 == 0 {
			err := fmt.Errorf("TAILSCALE_PATH is not executable: %s", abs)
			logger.Error("TAILSCALE_PATH validation failed", "error", err)
			return nil, err
		}
		path = abs

	// 2. Fallback: search `$PATH` first, then common installation paths.
	default:
		if look, err := exec.LookPath("tailscale"); err == nil {
			path = look
		} else {
			// Try platform-specific fallback paths
			fallbackPaths := getTailscaleFallbackPaths()
			var found bool
			for _, fallbackPath := range fallbackPaths {
				if isExecutableFile(fallbackPath) {
					path = fallbackPath
					found = true
					logger.Debug("Found tailscale binary at fallback path", "path", fallbackPath)
					break
				}
			}
			if !found {
				logger.Error("tailscale binary not found in PATH or common installation paths")
				return nil, fmt.Errorf("tailscale binary not found")
			}
		}
	}

	return &TailscaleCLI{tailscalePath: path}, nil
}

// ExecuteCommand runs the Tailscale CLI with validation, timeout, buffer limit
// and Windows-window hiding just like the TS version.
func (c *TailscaleCLI) ExecuteCommand(
	ctx context.Context,
	args []string,
	env []string,
) (string, error) {
	// --- command validation --------------------------------------------------
	if len(args) == 0 {
		return "", errors.New("no command specified")
	}

	// Validate the command is in our whitelist
	command := args[0]
	if !allowedCommands[command] {
		return "", fmt.Errorf("command %q not allowed", command)
	}

	// --- argument validation -------------------------------------------------
	for i, a := range args {
		if len(a) > maxArgLen {
			return "", fmt.Errorf("argument %d too long (%d chars)", i, len(a))
		}

		// Basic injection prevention - reject arguments with suspicious characters
		if strings.ContainsAny(a, ";&|`$(){}[]<>") {
			return "", fmt.Errorf("argument %d contains invalid characters: %q", i, a)
		}
	}

	logger.Debug("Executing tailscale command", "path", c.tailscalePath, "args", args)

	// --- build exec.Command --------------------------------------------------
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.tailscalePath, args...)
	setWinAttrs(cmd) // hides console on Windows, no-op elsewhere

	// Apply additional environment variables
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	// Capture + limit stdout/stderr to 10 MB each.
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	cmd.Stdout = newLimitWriter(&outBuf, maxBufSize)
	cmd.Stderr = newLimitWriter(&errBuf, maxBufSize)

	// --- execute -------------------------------------------------------------
	err := cmd.Run()

	stderrStr := strings.TrimSpace(errBuf.String())
	if stderrStr != "" {
		logger.Warn("CLI stderr", "stderr", stderrStr)
	}

	// --- handle errors -------------------------------------------------------
	if err != nil {
		cliErr := &CLIError{
			Command:    command,
			Args:       args,
			Stderr:     stderrStr,
			Underlying: err,
		}

		// Check for timeout
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			cliErr.Underlying = fmt.Errorf("command timed out after %s", timeout)
		}

		// Extract exit code if available
		if exitErr, ok := err.(*exec.ExitError); ok {
			cliErr.ExitCode = exitErr.ExitCode()
		}

		logger.Error("CLI command failed", "command", command, "args", args, "error", err)
		return "", cliErr
	}

	return strings.TrimSpace(outBuf.String()), nil
}

////////////////////////////////////////////////////////////////////////////////
// CLI methods
////////////////////////////////////////////////////////////////////////////////

// GetStatus gets Tailscale status
func (c *TailscaleCLI) GetStatus() (*schema.TailscaleStatus, error) {
	output, err := c.ExecuteCommand(context.Background(), []string{"status", "--json"}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	status, err := schema.ParseSchema[schema.TailscaleStatus](output)
	if err != nil {
		logger.Error("failed to parse status JSON", "error", err, "output", output)
		return nil, fmt.Errorf("failed to parse status data: %w", err)
	}

	return &status, nil
}

// GetVersion gets the Tailscale version
func (c *TailscaleCLI) GetVersion() (string, error) {
	return c.ExecuteCommand(context.Background(), []string{"version"}, nil)
}

// UpOptions defines options for the Up command
type UpOptions struct {
	LoginServer     string
	AcceptRoutes    bool
	AcceptDNS       bool
	Hostname        string
	AdvertiseRoutes []string
	AuthKey         string
	Timeout         time.Duration
}

// Up brings Tailscale up with structured options
func (c *TailscaleCLI) Up(options *UpOptions) error {
	args := []string{"up"}
	env := []string{}

	if options != nil {
		if options.LoginServer != "" {
			if err := c.validateStringInput(options.LoginServer, "loginServer"); err != nil {
				return fmt.Errorf("invalid login server: %w", err)
			}
			args = append(args, "--login-server", options.LoginServer)
		}

		if options.AcceptRoutes {
			args = append(args, "--accept-routes")
		}

		if options.AcceptDNS {
			args = append(args, "--accept-dns")
		}

		if options.Hostname != "" {
			if err := c.validateStringInput(options.Hostname, "hostname"); err != nil {
				return fmt.Errorf("invalid hostname: %w", err)
			}
			args = append(args, "--hostname", options.Hostname)
		}

		if len(options.AdvertiseRoutes) > 0 {
			if err := c.validateRoutes(options.AdvertiseRoutes); err != nil {
				return fmt.Errorf("invalid routes: %w", err)
			}
			args = append(args, "--advertise-routes", strings.Join(options.AdvertiseRoutes, ","))
		}

		if options.AuthKey != "" {
			if err := c.validateStringInput(options.AuthKey, "authKey"); err != nil {
				return fmt.Errorf("invalid auth key: %w", err)
			}
			// Pass auth key securely via environment variable
			logger.Debug("Auth key passed securely via TS_AUTHKEY environment variable")
			env = append(env, "TS_AUTHKEY="+options.AuthKey)
		}

		if options.Timeout > 0 {
			args = append(args, "--timeout", fmt.Sprintf("%ds", int(options.Timeout.Seconds())))
		}
	}

	_, err := c.ExecuteCommand(context.Background(), args, env)
	return err
}

// Down brings Tailscale down
func (c *TailscaleCLI) Down() error {
	_, err := c.ExecuteCommand(context.Background(), []string{"down"}, nil)
	return err
}

// Logout logs out of Tailscale
func (c *TailscaleCLI) Logout() error {
	_, err := c.ExecuteCommand(context.Background(), []string{"logout"}, nil)
	return err
}

// Ping pings a Tailscale peer with an optional count
func (c *TailscaleCLI) Ping(target string, count int) (string, error) {
	if err := c.validateTarget(target); err != nil {
		return "", fmt.Errorf("invalid target: %w", err)
	}

	if count < minPingCount || count > maxPingCount {
		return "", fmt.Errorf("count must be an integer between %d and %d", minPingCount, maxPingCount)
	}

	cmdArgs := []string{"ping", target, "-c", fmt.Sprintf("%d", count)}
	return c.ExecuteCommand(context.Background(), cmdArgs, nil)
}

// IP gets the Tailscale IP addresses
func (c *TailscaleCLI) IP() (string, error) {
	return c.ExecuteCommand(context.Background(), []string{"ip"}, nil)
}

// Netcheck runs network connectivity check
func (c *TailscaleCLI) Netcheck() (string, error) {
	return c.ExecuteCommand(context.Background(), []string{"netcheck"}, nil)
}

// SetExitNode sets or clears the exit node
func (c *TailscaleCLI) SetExitNode(nodeID string) error {
	args := []string{"set"}

	if nodeID != "" {
		if err := c.validateTarget(nodeID); err != nil {
			return fmt.Errorf("invalid node ID: %w", err)
		}
		args = append(args, "--exit-node", nodeID)
	} else {
		args = append(args, "--exit-node=") // Clear exit node
	}

	_, err := c.ExecuteCommand(context.Background(), args, nil)
	return err
}

// SetShieldsUp enables or disables shields up mode
func (c *TailscaleCLI) SetShieldsUp(enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := c.ExecuteCommand(context.Background(), []string{"set", "--shields-up", val}, nil)
	return err
}

// IsAvailable checks if the Tailscale CLI is available
func (c *TailscaleCLI) IsAvailable() bool {
	_, err := c.ExecuteCommand(context.Background(), []string{"version"}, nil)
	return err == nil
}

// validateTarget validates target format (hostname, IP, or Tailscale node name)
func (c *TailscaleCLI) validateTarget(target string) error {
	if target == "" {
		return errors.New("target cannot be empty")
	}

	// Comprehensive validation to prevent command injection
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "[", "]", "<", ">", "\\", "'", "\""}
	for _, char := range dangerousChars {
		if strings.Contains(target, char) {
			return fmt.Errorf("invalid character %q in target", char)
		}
	}

	// Additional validation for common patterns
	if strings.Contains(target, "..") || strings.HasPrefix(target, "/") || strings.Contains(target, "~") {
		return errors.New("invalid path patterns in target")
	}

	// Validate target format using regex
	if !validTargetPattern.MatchString(target) {
		return errors.New("target contains invalid characters")
	}

	// Length validation (DNS hostname max length)
	if len(target) > maxHostnameLen {
		return fmt.Errorf("target too long (max %d chars)", maxHostnameLen)
	}

	return nil
}

// validateStringInput validates general string inputs
func (c *TailscaleCLI) validateStringInput(input, fieldName string) error {
	// Check for dangerous characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "<", ">", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(input, char) {
			return fmt.Errorf("invalid character %q in %s", char, fieldName)
		}
	}

	// Length validation
	if len(input) > maxArgLen {
		return fmt.Errorf("%s too long (max %d chars)", fieldName, maxArgLen)
	}

	return nil
}

// validateRoutes validates CIDR routes
func (c *TailscaleCLI) validateRoutes(routes []string) error {
	for i, route := range routes {
		// Basic CIDR validation
		if route == "0.0.0.0/0" || route == "::/0" {
			continue // Allow default routes
		}

		if !cidrPattern.MatchString(route) {
			return fmt.Errorf("invalid route format at index %d: %s", i, route)
		}

		// Additional validation using net package
		_, _, err := net.ParseCIDR(route)
		if err != nil {
			return fmt.Errorf("invalid CIDR route at index %d (%s): %w", i, route, err)
		}
	}

	return nil
}

// ListPeers gets a list of peer hostnames
func (c *TailscaleCLI) ListPeers() ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var peers []string
	if status.Peer != nil {
		for _, peer := range status.Peer {
			if peer.HostName != "" {
				peers = append(peers, peer.HostName)
			}
		}
	}

	return peers, nil
}
