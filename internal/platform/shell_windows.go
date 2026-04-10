package platform

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DetectShellConfig returns an empty string on Windows because PATH is managed
// via the registry, not a shell startup file.
func DetectShellConfig() string { return "" }

// EnsureInPATH adds binDir to the user's PATH in the Windows Registry
// (HKCU\Environment) using the `reg` command-line tool. Returns true if the
// entry was actually added.
func EnsureInPATH(_, binDir string) (bool, error) {
	// Read current user PATH from registry.
	out, err := exec.Command("reg", "query",
		`HKCU\Environment`, "/v", "PATH").Output()
	if err != nil && !strings.Contains(string(out), "ERROR") {
		return false, fmt.Errorf("reading registry PATH: %w", err)
	}

	current := parseRegPath(string(out))

	// Check if already present.
	for _, p := range strings.Split(current, ";") {
		if strings.EqualFold(strings.TrimSpace(p), binDir) {
			return false, nil
		}
	}

	// Append and write back.
	newPath := strings.TrimRight(current, ";") + ";" + binDir
	_, err = exec.Command("reg", "add",
		`HKCU\Environment`, "/v", "PATH",
		"/t", "REG_EXPAND_SZ",
		"/d", newPath, "/f",
	).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("writing registry PATH: %w", err)
	}
	return true, nil
}

// IsInCurrentPATH reports whether dir is already in the running process's PATH.
func IsInCurrentPATH(dir string) bool {
	for _, p := range strings.Split(os.Getenv("PATH"), ";") {
		if strings.EqualFold(strings.TrimSpace(p), dir) {
			return true
		}
	}
	return false
}

// parseRegPath extracts the PATH value from `reg query` output.
func parseRegPath(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "PATH") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}
