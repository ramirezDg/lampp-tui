package xampp

import (
	"context"
	"fmt"
	"os"
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

	// Apache: read master PID from PID file (most reliable) so ss lookup
	// matches the socket-owning process, not a worker.
	apacheMaster := readPIDFile("/opt/lampp/logs/httpd.pid")
	if apacheMaster == "" {
		apacheMaster = findProcessPID(ctx, "httpd", "apache2")
	}
	apacheAll := findAllPIDs(ctx, "httpd", "apache2")

	mysqlPID := findProcessPID(ctx, "mysqld")
	ftpPID := findProcessPID(ctx, "proftpd")

	return map[string]ServiceInfo{
		"Apache": {
			Name:  "Apache",
			PID:   apacheMaster,
			Port:  findPort(ssOutput, apacheMaster, apacheAll, "httpd", "apache2"),
			State: strings.Contains(statusText, "Apache is running"),
		},
		"MySQL": {
			Name:  "MySQL",
			PID:   mysqlPID,
			Port:  findPort(ssOutput, mysqlPID, []string{mysqlPID}, "mysqld"),
			State: strings.Contains(statusText, "MySQL is running"),
		},
		"FTP": {
			Name:  "FTP",
			PID:   ftpPID,
			Port:  findPort(ssOutput, ftpPID, []string{ftpPID}, "proftpd"),
			State: strings.Contains(statusText, "ProFTPD is running"),
		},
	}, nil
}

// readPIDFile reads a PID from a file, returning an empty string on error.
func readPIDFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// findProcessPID returns a single PID for display purposes (first match).
func findProcessPID(ctx context.Context, names ...string) string {
	pids := findAllPIDs(ctx, names...)
	if len(pids) == 0 {
		return ""
	}
	return pids[0]
}

// findAllPIDs returns every running PID for any of the given process names.
func findAllPIDs(ctx context.Context, names ...string) []string {
	for _, name := range names {
		// pidof without -s returns all PIDs separated by spaces.
		out, err := exec.CommandContext(ctx, "pidof", name).Output()
		if err == nil {
			if pids := strings.Fields(strings.TrimSpace(string(out))); len(pids) > 0 {
				return pids
			}
		}
	}
	return nil
}

// runSSCommand executes ss -lptn and returns its raw output. Errors are
// silently swallowed so that a missing ss(8) binary degrades gracefully
// (ports simply show as empty).
func runSSCommand(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "ss", "-lptn").Output()
	return string(out), err
}

var portRe = regexp.MustCompile(`:(\d+)`)

// findPort looks up the listening port for a service in ss output.
//
// Search order:
//  1. Lines containing masterPID (socket-owning process, e.g. Apache master).
//  2. Lines containing any PID in allPIDs (catches single-process daemons).
//  3. Lines containing any process name (last-resort name-only fallback).
//
// Returns "N/A" when no port can be determined.
func findPort(ssOutput, masterPID string, allPIDs []string, names ...string) string {
	if ssOutput == "" {
		return "N/A"
	}
	lines := strings.Split(ssOutput, "\n")

	extractPort := func(line string) string {
		if m := portRe.FindStringSubmatch(line); len(m) == 2 {
			return m[1]
		}
		return ""
	}

	// 1. Master PID (most precise).
	if masterPID != "" {
		for _, line := range lines {
			if strings.Contains(line, masterPID) {
				if p := extractPort(line); p != "" {
					return p
				}
			}
		}
	}

	// 2. Any known PID for this service.
	for _, pid := range allPIDs {
		if pid == masterPID {
			continue // already tried
		}
		for _, line := range lines {
			if strings.Contains(line, "pid="+pid) || strings.Contains(line, ","+pid+",") {
				if p := extractPort(line); p != "" {
					return p
				}
			}
		}
	}

	// 3. Process name only.
	for _, line := range lines {
		for _, name := range names {
			if strings.Contains(line, name) {
				if p := extractPort(line); p != "" {
					return p
				}
			}
		}
	}

	return "N/A"
}
