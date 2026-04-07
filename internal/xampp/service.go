package xampp

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const lamppBin = "/opt/lampp/lampp"

// ServiceStatus holds the running/stopped state of each XAMPP service.
type ServiceStatus struct {
	Apache bool
	MySQL  bool
	FTP    bool
}

// ServiceInfo holds the runtime details for a single XAMPP service.
type ServiceInfo struct {
	Name  string
	PID   string
	Port  string
	State bool
}

// Snapshot bundles status + per-service details into a single value so callers
// can issue one subprocess call instead of two.
type Snapshot struct {
	Status  ServiceStatus
	Details map[string]ServiceInfo
}

// GetSnapshot returns the full service picture in one call: running states,
// PIDs, and ports. Callers that previously called GetServiceStatus and
// GetServiceDetails separately (causing lampp status to run twice) should use
// this instead.
func GetSnapshot(ctx context.Context) (Snapshot, error) {
	statusText, err := runLamppStatus(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	status := ServiceStatus{
		Apache: strings.Contains(statusText, "Apache is running"),
		MySQL:  strings.Contains(statusText, "MySQL is running"),
		FTP:    strings.Contains(statusText, "ProFTPD is running"),
	}

	details, err := buildDetails(ctx, status, statusText)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{Status: status, Details: details}, nil
}

// GetServiceStatus returns only the boolean running state of each service.
// Prefer GetSnapshot when you also need PID/port data.
func GetServiceStatus(ctx context.Context) (ServiceStatus, error) {
	text, err := runLamppStatus(ctx)
	if err != nil {
		return ServiceStatus{}, err
	}
	return ServiceStatus{
		Apache: strings.Contains(text, "Apache is running"),
		MySQL:  strings.Contains(text, "MySQL is running"),
		FTP:    strings.Contains(text, "ProFTPD is running"),
	}, nil
}

// Control starts or stops a XAMPP service. Accepted service values are
// "apache", "mysql", "ftp", and "all". Accepted actions are "start", "stop",
// and "restart". Name normalisation (e.g. "Apache" → "apache") happens here
// so callers in the UI layer do not need to know lampp's naming conventions.
func Control(service, action string) error {
	normalized := strings.ToLower(service)
	var arg string
	switch normalized {
	case "apache", "mysql", "ftp":
		arg = action + normalized
	case "all":
		arg = action
	default:
		return fmt.Errorf("unsupported service: %s", service)
	}

	cmd := exec.Command("sudo", lamppBin, arg)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("lampp %s %s: %w", action, service, err)
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// internal helpers
// ────────────────────────────────────────────────────────────────────────────

func runLamppStatus(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sudo", lamppBin, "status").Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("timeout querying XAMPP status")
	}
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// buildDetails collects PID and port for each service. It accepts the already-
// fetched statusText so no second lampp call is needed.
func buildDetails(ctx context.Context, status ServiceStatus, statusText string) (map[string]ServiceInfo, error) {
	ssOutput, _ := runSSCommand(ctx)

	apachePID := findProcessPID(ctx, "httpd", "apache2")
	mysqlPID := findProcessPID(ctx, "mysqld")
	ftpPID := findProcessPID(ctx, "proftpd")

	return map[string]ServiceInfo{
		"Apache": {
			Name:  "Apache",
			PID:   apachePID,
			Port:  findPortByPID(ssOutput, apachePID, "httpd", "apache2"),
			State: strings.Contains(statusText, "Apache is running"),
		},
		"MySQL": {
			Name:  "MySQL",
			PID:   mysqlPID,
			Port:  findPortByPID(ssOutput, mysqlPID, "mysqld"),
			State: strings.Contains(statusText, "MySQL is running"),
		},
		"FTP": {
			Name:  "FTP",
			PID:   ftpPID,
			Port:  findPortByPID(ssOutput, ftpPID, "proftpd"),
			State: strings.Contains(statusText, "ProFTPD is running"),
		},
	}, nil
}

// findProcessPID tries each process name in order and returns the first PID
// found, or an empty string when none are running.
func findProcessPID(ctx context.Context, names ...string) string {
	for _, name := range names {
		out, err := exec.CommandContext(ctx, "pidof", "-s", name).Output()
		if err == nil {
			if pid := strings.TrimSpace(string(out)); pid != "" {
				return pid
			}
		}
	}
	return ""
}

// runSSCommand executes ss -lptn and returns its raw output. Errors are
// silently swallowed so that a missing ss(8) binary degrades gracefully
// (ports simply show as empty).
func runSSCommand(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "ss", "-lptn").Output()
	return string(out), err
}

var portRe = regexp.MustCompile(`:(\d+)`)

// findPortByPID scans ss output for a line that contains the given PID and
// any of the provided process names, then extracts the port number. Falls back
// to a name-only search when no PID is available or when the PID-based search
// yields nothing. Returns "N/A" when no port can be determined.
func findPortByPID(ssOutput, pid string, names ...string) string {
	if ssOutput == "" {
		return "N/A"
	}
	lines := strings.Split(ssOutput, "\n")

	matchesName := func(line string) bool {
		for _, name := range names {
			if strings.Contains(line, name) {
				return true
			}
		}
		return false
	}

	// Primary: match by PID and process name.
	if pid != "" {
		for _, line := range lines {
			if strings.Contains(line, pid) && matchesName(line) {
				if m := portRe.FindStringSubmatch(line); len(m) == 2 {
					return m[1]
				}
			}
		}
	}

	// Fallback: match by process name only.
	for _, line := range lines {
		if matchesName(line) {
			if m := portRe.FindStringSubmatch(line); len(m) == 2 {
				return m[1]
			}
		}
	}

	return "N/A"
}
