package platform

import (
	"os/exec"
	"syscall"
)

// OpenBrowser opens url in the system browser. The child process is detached
// from the TUI's controlling terminal (raw mode) using a new session so that
// xdg-open doesn't interfere with Bubble Tea's input handling.
func OpenBrowser(url string) {
	for _, browser := range []string{"xdg-open", "sensible-browser", "x-www-browser"} {
		if _, err := exec.LookPath(browser); err != nil {
			continue
		}
		cmd := exec.Command(browser, url)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		// Create a new session so the child is not attached to the TUI terminal.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if cmd.Start() == nil {
			return
		}
	}
}
