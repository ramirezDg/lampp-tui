package platform

import (
	"os/exec"
	"syscall"
)

// OpenBrowser opens url using the Windows shell (cmd /c start).
// CREATE_NEW_PROCESS_GROUP detaches the child from the TUI's console.
func OpenBrowser(url string) {
	cmd := exec.Command("cmd", "/c", "start", url)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Start() //nolint:errcheck
}
