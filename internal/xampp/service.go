package xampp

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"xampp-tui/internal/platform"
)

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

// Snapshot bundles status + per-service details into a single value.
type Snapshot struct {
	Status  ServiceStatus
	Details map[string]ServiceInfo
}

// GetSnapshot returns the full service picture in one call: running states,
// PIDs, and ports.
func GetSnapshot(ctx context.Context) (Snapshot, error) {
	statusText, err := platform.RunServiceCmd(ctx, "status")
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
func GetServiceStatus(ctx context.Context) (ServiceStatus, error) {
	text, err := platform.RunServiceCmd(ctx, "status")
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
// and "restart".
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

	_, err := platform.RunServiceCmd(context.Background(), arg)
	return err
}

// ─── internal helpers ─────────────────────────────────────────────────────────

func buildDetails(ctx context.Context, status ServiceStatus, statusText string) (map[string]ServiceInfo, error) {
	portsOutput := platform.ListeningPorts(ctx)

	// Apache: read master PID from PID file (most reliable).
	apacheMaster := readPIDFile(platform.PIDFilePath("apache"))
	if apacheMaster == "" {
		apacheMaster = firstPID(platform.PIDsForProcess(ctx, "httpd", "apache2"))
	}
	apacheAll := platform.PIDsForProcess(ctx, "httpd", "apache2")

	mysqlPID := firstPID(platform.PIDsForProcess(ctx, "mysqld"))
	ftpPID := firstPID(platform.PIDsForProcess(ctx, "proftpd"))

	return map[string]ServiceInfo{
		"Apache": {
			Name:  "Apache",
			PID:   apacheMaster,
			Port:  findPort(portsOutput, apacheMaster, apacheAll, "httpd", "apache2"),
			State: strings.Contains(statusText, "Apache is running"),
		},
		"MySQL": {
			Name:  "MySQL",
			PID:   mysqlPID,
			Port:  findPort(portsOutput, mysqlPID, []string{mysqlPID}, "mysqld"),
			State: strings.Contains(statusText, "MySQL is running"),
		},
		"FTP": {
			Name:  "FTP",
			PID:   ftpPID,
			Port:  findPort(portsOutput, ftpPID, []string{ftpPID}, "proftpd"),
			State: strings.Contains(statusText, "ProFTPD is running"),
		},
	}, nil
}

func readPIDFile(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func firstPID(pids []string) string {
	if len(pids) > 0 {
		return pids[0]
	}
	return ""
}

var portRe = regexp.MustCompile(`:(\d+)`)

func findPort(portsOutput, masterPID string, allPIDs []string, names ...string) string {
	if portsOutput == "" {
		return "N/A"
	}
	lines := strings.Split(portsOutput, "\n")

	extractPort := func(line string) string {
		if m := portRe.FindStringSubmatch(line); len(m) == 2 {
			return m[1]
		}
		return ""
	}

	if masterPID != "" {
		for _, line := range lines {
			if strings.Contains(line, masterPID) {
				if p := extractPort(line); p != "" {
					return p
				}
			}
		}
	}

	for _, pid := range allPIDs {
		if pid == masterPID {
			continue
		}
		for _, line := range lines {
			if strings.Contains(line, "pid="+pid) || strings.Contains(line, ","+pid+",") {
				if p := extractPort(line); p != "" {
					return p
				}
			}
		}
	}

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
