package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RunServiceCmd executes `sudo /opt/lampp/lampp {arg}` and returns its output.
// A 2-second timeout is applied automatically.
func RunServiceCmd(ctx context.Context, arg string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sudo", LamppBinPath(), arg).Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("timeout querying XAMPP status")
	}
	return string(out), err
}

// ListeningPorts runs `ss -lptn` and returns its raw output.
// Returns an empty string if ss is not available.
func ListeningPorts(ctx context.Context) string {
	out, _ := exec.CommandContext(ctx, "ss", "-lptn").Output()
	return string(out)
}

// PIDsForProcess returns the PIDs of all running processes matching any of
// the given names, using pidof(8).
func PIDsForProcess(ctx context.Context, names ...string) []string {
	for _, name := range names {
		out, err := exec.CommandContext(ctx, "pidof", name).Output()
		if err == nil {
			if pids := strings.Fields(strings.TrimSpace(string(out))); len(pids) > 0 {
				return pids
			}
		}
	}
	return nil
}

// PIDFilePath returns the filesystem path of the PID file for the given
// XAMPP service, or an empty string if not applicable.
func PIDFilePath(service string) string {
	switch service {
	case "apache":
		return ActiveXAMPPPath() + "/logs/httpd.pid"
	}
	return ""
}
