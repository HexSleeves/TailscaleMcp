package tailscale

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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
		return remaining, fmt.Errorf("output exceeds %d bytes: %w", l.limit, io.ErrShortWrite)
	}

	n, err := l.w.Write(p)
	l.n += n
	return n, err
}

// getTailscaleFallbackPaths returns platform-specific fallback paths for the Tailscale binary
func getTailscaleFallbackPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join("C:", "Program Files", "Tailscale", "tailscale.exe"),
			filepath.Join("C:", "Program Files (x86)", "Tailscale", "tailscale.exe"),
		}
	case "darwin":
		return []string{
			"/usr/local/bin/tailscale",
			"/opt/homebrew/bin/tailscale",
			"/usr/bin/tailscale",
		}
	default: // Linux and other Unix-like systems
		return []string{
			"/usr/bin/tailscale",
			"/usr/local/bin/tailscale",
			"/opt/tailscale/bin/tailscale",
			"/snap/bin/tailscale",
		}
	}
}

// isExecutableFile checks if the given path is an executable file
func isExecutableFile(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		return !st.IsDir() && strings.HasSuffix(strings.ToLower(path), ".exe")
	}

	// Check if it's a regular file (not directory) and is executable
	return !st.IsDir() && st.Mode()&0111 != 0
}
