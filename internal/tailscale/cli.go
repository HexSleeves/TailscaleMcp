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
	"github.com/hexsleeves/tailscale-mcp-server/pkg/cli"
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

func NewTailscaleCLI() (*TailscaleCLI, error) {
	path := os.Getenv("TAILSCALE_PATH")

	switch {
	// 1. Env-provided path: expand & validate.
	case path != "":
		abs, err := filepath.Abs(path)
		if err != nil {
			logger.Errorf("invalid TAILSCALE_PATH: %w", err)
			return nil, err
		}

		// Make sure it actually exists & is executable.
		if st, err := os.Stat(abs); err != nil || st.IsDir() ||
			st.Mode()&0111 == 0 {
			logger.Errorf("TAILSCALE_PATH is not an executable: %s", abs)
			return nil, err
		}
		path = abs

	// 2. Fallback: search `$PATH`.
	default:
		look, err := exec.LookPath("tailscale") // same as your getCommandPath
		if err != nil {
			logger.Errorf("tailscale binary not found in PATH (and TAILSCALE_PATH unset)")
			return nil, err
		}
		path = look
	}

	return &TailscaleCLI{tailscalePath: path}, nil
}

// ExecuteCommand runs the Tailscale CLI with validation, timeout, buffer limit
// and Windows-window hiding just like the TS version.
func (c *TailscaleCLI) ExecuteCommand(
	parent context.Context,
	args []string,
	env []string,
) CLIResponse[string] {
	// --- command validation --------------------------------------------------
	if len(args) == 0 {
		return CLIResponse[string]{
			Success: false,
			Error:   "no command specified",
		}
	}

	// Validate the command is in our whitelist
	command := args[0]
	if !allowedCommands[command] {
		return CLIResponse[string]{
			Success: false,
			Error:   fmt.Sprintf("command '%s' not allowed", command),
		}
	}

	// --- argument validation -------------------------------------------------
	for _, a := range args {
		if len(a) > maxArgLen {
			return CLIResponse[string]{
				Success: false,
				Error:   "command argument too long",
			}
		}

		// Basic injection prevention - reject arguments with suspicious characters
		if strings.ContainsAny(a, ";&|`$(){}[]<>") {
			return CLIResponse[string]{
				Success: false,
				Error:   "command argument contains invalid characters",
			}
		}
	}

	logger.Debugf("Executing: %s %s", c.tailscalePath, strings.Join(args, " "))

	// --- build exec.Command --------------------------------------------------
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.tailscalePath, args...)
	setWinAttrs(cmd) // hides console on Windows, no-op elsewhere

	// Apply additional environment variables
	if len(env) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, env...)
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
		logger.Warnf("CLI stderr: %s", stderrStr)
	}

	// --- build response ------------------------------------------------------
	if err != nil {
		// ctx.Err() will be non-nil if we hit the timeout.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			err = fmt.Errorf("command timed out after %s", timeout)
		}

		logger.Errorf("CLI command failed: %v", err)
		return CLIResponse[string]{
			Success: false,
			Error:   err.Error(),
			Stderr:  stderrStr,
		}
	}

	return CLIResponse[string]{
		Success: true,
		Stderr:  stderrStr,
		Data:    strings.TrimSpace(outBuf.String()),
	}
}

////////////////////////////////////////////////////////////////////////////////
// helpers
////////////////////////////////////////////////////////////////////////////////

// limitWriter stops writing once the limit is reached and returns an error to
// the caller (propagates to exec.Command.Run()).
type limitWriter struct {
	w     *bytes.Buffer
	n     int
	limit int
}

func newLimitWriter(w *bytes.Buffer, limit int) *limitWriter {
	return &limitWriter{w: w, limit: limit}
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.n+len(p) > l.limit {
		// write only the remaining bytes so we don't exceed the slice capacity
		remaining := l.limit - l.n
		if remaining > 0 {
			l.w.Write(p[:remaining])
			l.n += remaining
		}
		return remaining, fmt.Errorf("output exceeds %d bytes", l.limit)
	}
	n, err := l.w.Write(p)
	l.n += n
	return n, err
}

// coalesce returns the first non-empty string—handy for defaulting errors.
func coalesce(s, fallback string) string {
	if s != "" {
		return s
	}
	return fallback
}

////////////////////////////////////////////////////////////////////////////////
// CLI methods
////////////////////////////////////////////////////////////////////////////////

// Get Tailscale status
func (c *TailscaleCLI) GetStatus() CLIResponse[cli.TailscaleStatus] {
	raw := c.ExecuteCommand(context.Background(), []string{"status", "--json"}, nil)
	if !raw.Success {
		return CLIResponse[cli.TailscaleStatus]{
			Success: false,
			Stderr:  raw.Stderr,
			Error:   coalesce(raw.Error, "unknown error"),
		}
	}

	st, err := cli.ParseSchema[cli.TailscaleStatus](raw.Data)
	if err != nil {
		logger.Errorf("failed to parse status JSON: %v", err)

		return CLIResponse[cli.TailscaleStatus]{
			Success: false,
			Error:   fmt.Sprintf("failed to parse status data: %v", err),
		}
	}

	return CLIResponse[cli.TailscaleStatus]{
		Success: true,
		Data:    st,
	}
}

// GetVersion gets the Tailscale version
func (c *TailscaleCLI) GetVersion() CLIResponse[string] {
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
func (c *TailscaleCLI) Up(options *UpOptions) CLIResponse[string] {
	args := []string{"up"}
	env := []string{}

	if options != nil {
		if options.LoginServer != "" {
			if err := c.validateStringInput(options.LoginServer, "loginServer"); err != nil {
				return CLIResponse[string]{
					Success: false,
					Error:   err.Error(),
				}
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
				return CLIResponse[string]{
					Success: false,
					Error:   err.Error(),
				}
			}
			args = append(args, "--hostname", options.Hostname)
		}

		if len(options.AdvertiseRoutes) > 0 {
			if err := c.validateRoutes(options.AdvertiseRoutes); err != nil {
				return CLIResponse[string]{
					Success: false,
					Error:   err.Error(),
				}
			}
			args = append(args, "--advertise-routes", strings.Join(options.AdvertiseRoutes, ","))
		}

		if options.AuthKey != "" {
			if err := c.validateStringInput(options.AuthKey, "authKey"); err != nil {
				return CLIResponse[string]{
					Success: false,
					Error:   err.Error(),
				}
			}
			// Pass auth key securely via environment variable
			logger.Debugf("Auth key passed securely via TS_AUTHKEY environment variable")
			env = append(env, "TS_AUTHKEY="+options.AuthKey)
		}

		if options.Timeout > 0 {
			args = append(args, "--timeout", fmt.Sprintf("%ds", int(options.Timeout.Seconds())))
		}
	}

	return c.ExecuteCommand(context.Background(), args, env)
}

// UpSimple brings Tailscale up with optional string arguments (for backward compatibility)
func (c *TailscaleCLI) UpSimple(args ...string) CLIResponse[string] {
	cmdArgs := append([]string{"up"}, args...)
	return c.ExecuteCommand(context.Background(), cmdArgs, nil)
}

// Down brings Tailscale down
func (c *TailscaleCLI) Down() CLIResponse[string] {
	return c.ExecuteCommand(context.Background(), []string{"down"}, nil)
}

// Logout logs out of Tailscale
func (c *TailscaleCLI) Logout() CLIResponse[string] {
	return c.ExecuteCommand(context.Background(), []string{"logout"}, nil)
}

// Ping pings a Tailscale peer with an optional count
func (c *TailscaleCLI) Ping(target string, count int) CLIResponse[string] {
	if err := c.validateTarget(target); err != nil {
		return CLIResponse[string]{
			Success: false,
			Error:   err.Error(),
		}
	}

	if count < minPingCount || count > maxPingCount {
		return CLIResponse[string]{
			Success: false,
			Error:   fmt.Sprintf("count must be an integer between %d and %d", minPingCount, maxPingCount),
		}
	}

	cmdArgs := []string{"ping", target, "-c", fmt.Sprintf("%d", count)}
	return c.ExecuteCommand(context.Background(), cmdArgs, nil)
}

// IP gets the Tailscale IP addresses
func (c *TailscaleCLI) IP() CLIResponse[string] {
	return c.ExecuteCommand(context.Background(), []string{"ip"}, nil)
}

// Netcheck runs network connectivity check
func (c *TailscaleCLI) Netcheck() CLIResponse[string] {
	return c.ExecuteCommand(context.Background(), []string{"netcheck"}, nil)
}

// SetExitNode sets or clears the exit node
func (c *TailscaleCLI) SetExitNode(nodeID string) CLIResponse[string] {
	args := []string{"set"}

	if nodeID != "" {
		if err := c.validateTarget(nodeID); err != nil {
			return CLIResponse[string]{
				Success: false,
				Error:   err.Error(),
			}
		}
		args = append(args, "--exit-node", nodeID)
	} else {
		args = append(args, "--exit-node=") // Clear exit node
	}

	return c.ExecuteCommand(context.Background(), args, nil)
}

// SetShieldsUp enables or disables shields up mode
func (c *TailscaleCLI) SetShieldsUp(enabled bool) CLIResponse[string] {
	val := "false"
	if enabled {
		val = "true"
	}
	return c.ExecuteCommand(context.Background(), []string{"set", "--shields-up", val}, nil)
}

// IsAvailable checks if the Tailscale CLI is available
func (c *TailscaleCLI) IsAvailable() CLIResponse[bool] {
	resp := c.ExecuteCommand(context.Background(), []string{"version"}, nil)
	return CLIResponse[bool]{
		Success: resp.Success,
		Error:   resp.Error,
		Stderr:  resp.Stderr,
		Data:    resp.Success, // If version command succeeds, CLI is available
	}
}

// validateTarget validates target format (hostname, IP, or Tailscale node name)
func (c *TailscaleCLI) validateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("invalid target specified")
	}

	// Comprehensive validation to prevent command injection
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "[", "]", "<", ">", "\\", "'", "\""}
	for _, char := range dangerousChars {
		if strings.Contains(target, char) {
			return fmt.Errorf("invalid character '%s' in target", char)
		}
	}

	// Additional validation for common patterns
	if strings.Contains(target, "..") || strings.HasPrefix(target, "/") || strings.Contains(target, "~") {
		return fmt.Errorf("invalid path patterns in target")
	}

	// Validate target format using regex
	if !validTargetPattern.MatchString(target) {
		return fmt.Errorf("target contains invalid characters")
	}

	// Length validation (DNS hostname max length)
	if len(target) > maxHostnameLen {
		return fmt.Errorf("target too long")
	}

	return nil
}

// validateStringInput validates general string inputs
func (c *TailscaleCLI) validateStringInput(input, fieldName string) error {
	// Check for dangerous characters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "<", ">", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(input, char) {
			return fmt.Errorf("invalid character '%s' in %s", char, fieldName)
		}
	}

	// Length validation
	if len(input) > maxArgLen {
		return fmt.Errorf("%s too long", fieldName)
	}

	return nil
}

// validateRoutes validates CIDR routes
func (c *TailscaleCLI) validateRoutes(routes []string) error {
	for _, route := range routes {
		// Basic CIDR validation
		if route == "0.0.0.0/0" || route == "::/0" {
			continue // Allow default routes
		}

		if !cidrPattern.MatchString(route) {
			return fmt.Errorf("invalid route format: %s", route)
		}

		// Additional validation using net package
		_, _, err := net.ParseCIDR(route)
		if err != nil {
			return fmt.Errorf("invalid CIDR route %s: %v", route, err)
		}
	}

	return nil
}

// ListPeers gets a list of peer hostnames
func (c *TailscaleCLI) ListPeers() CLIResponse[[]string] {
	statusResult := c.GetStatus()

	if !statusResult.Success {
		return CLIResponse[[]string]{
			Success: false,
			Error:   statusResult.Error,
			Stderr:  statusResult.Stderr,
		}
	}

	var peers []string
	if statusResult.Data.Peer != nil {
		for _, peer := range statusResult.Data.Peer {
			if peer.HostName != "" {
				peers = append(peers, peer.HostName)
			}
		}
	}

	return CLIResponse[[]string]{
		Success: true,
		Data:    peers,
	}
}
