//go:build !windows

package tailscale

import "os/exec"

// setWinAttrs is a no-op outside Windows so the code compiles everywhere.
func setWinAttrs(cmd *exec.Cmd) {}
