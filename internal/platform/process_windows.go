package platform

import "os/exec"

// KillProcess terminates the process with the given PID using taskkill.
func KillProcess(pid string) {
	exec.Command("taskkill", "/PID", pid, "/F").Run() //nolint:errcheck
}
