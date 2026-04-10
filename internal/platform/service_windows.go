package platform

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunServiceCmd controls an XAMPP service on Windows using individual binaries.
// Supported args: "status", "startapache", "stopapache", "startmysql", "stopmysql",
// "start" (all), "stop" (all), "restart" (all).
func RunServiceCmd(ctx context.Context, arg string) (string, error) {
	base := ActiveXAMPPPath()
	apacheBin := filepath.Join(base, "apache", "bin", "httpd.exe")
	mysqlBin := filepath.Join(base, "mysql", "bin", "mysqladmin.exe")

	switch arg {
	case "status":
		return windowsStatus(ctx, apacheBin, mysqlBin)
	case "startapache":
		return runOut(ctx, apacheBin, "-k", "start")
	case "stopapache":
		return runOut(ctx, apacheBin, "-k", "stop")
	case "startmysql":
		return runOut(ctx, filepath.Join(base, "mysql", "bin", "mysqld.exe"), "--standalone")
	case "stopmysql":
		return runOut(ctx, mysqlBin, "-u", "root", "shutdown")
	case "start":
		runOut(ctx, apacheBin, "-k", "start")          //nolint:errcheck
		runOut(ctx, filepath.Join(base, "mysql", "bin", "mysqld.exe"), "--standalone") //nolint:errcheck
		return "all started", nil
	case "stop":
		runOut(ctx, apacheBin, "-k", "stop")            //nolint:errcheck
		runOut(ctx, mysqlBin, "-u", "root", "shutdown") //nolint:errcheck
		return "all stopped", nil
	case "restart":
		runOut(ctx, apacheBin, "-k", "restart") //nolint:errcheck
		return "apache restarted", nil
	}
	return "", fmt.Errorf("unsupported service command: %s", arg)
}

func windowsStatus(ctx context.Context, apacheBin, _ string) (string, error) {
	var sb strings.Builder

	// Check Apache via tasklist
	out, _ := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq httpd.exe", "/NH").Output()
	if strings.Contains(string(out), "httpd.exe") {
		sb.WriteString("Apache is running.\n")
	} else {
		sb.WriteString("Apache is not running.\n")
	}

	// Check MySQL
	out2, _ := exec.CommandContext(ctx, "tasklist", "/FI", "IMAGENAME eq mysqld.exe", "/NH").Output()
	if strings.Contains(string(out2), "mysqld.exe") {
		sb.WriteString("MySQL is running.\n")
	} else {
		sb.WriteString("MySQL is not running.\n")
	}

	// FTP not standard on Windows XAMPP
	sb.WriteString("ProFTPD is not running.\n")
	_ = apacheBin
	return sb.String(), nil
}

func runOut(ctx context.Context, bin string, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, bin, args...).CombinedOutput()
	return string(out), err
}

// ListeningPorts runs `netstat -ano` and returns its raw output.
func ListeningPorts(ctx context.Context) string {
	out, _ := exec.CommandContext(ctx, "netstat", "-ano").Output()
	return string(out)
}

// PIDsForProcess returns PIDs for the given process names using tasklist.
func PIDsForProcess(ctx context.Context, names ...string) []string {
	var pids []string
	for _, name := range names {
		out, err := exec.CommandContext(ctx, "tasklist", "/FI",
			"IMAGENAME eq "+name, "/NH", "/FO", "CSV").Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Split(line, ",")
			if len(fields) >= 2 {
				pid := strings.Trim(fields[1], `" `)
				if pid != "" {
					pids = append(pids, pid)
				}
			}
		}
	}
	return pids
}

// PIDFilePath returns empty on Windows — PID files are not used.
func PIDFilePath(_ string) string { return "" }
