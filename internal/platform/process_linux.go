package platform

import "os/exec"

// KillProcess sends SIGTERM to the process with the given PID string.
func KillProcess(pid string) {
	exec.Command("kill", pid).Run() //nolint:errcheck
}
