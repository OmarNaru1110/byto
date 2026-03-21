//go:build windows

package deps

import (
	"os/exec"
	"syscall"
)

// HideWindow sets the SysProcAttr to hide the console window on Windows
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
