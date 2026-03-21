//go:build !windows

package deps

import "os/exec"

// HideWindow is a no-op on non-Windows platforms
func HideWindow(cmd *exec.Cmd) {
	// No-op on non-Windows
}
