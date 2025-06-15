//go:build windows

package tailscale

import (
	"os/exec"
	"syscall"
)

// setWinAttrs hides the console window that would otherwise pop up.
func setWinAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
