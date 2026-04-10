package platform

import (
	"os"
	"path/filepath"
	"strings"
)

// XAMPPBaseDir is the parent directory for all versioned XAMPP installations.
func XAMPPBaseDir() string { return `C:\xampp-versions` }

// ActiveXAMPPPath returns the path to the currently active XAMPP installation.
// On Windows there are no symlinks, so xampp-tui stores the active path in
// a small config file: %APPDATA%\xampp-tui\active.txt.
// Falls back to C:\xampp if the config file is missing.
func ActiveXAMPPPath() string {
	data, err := os.ReadFile(filepath.Join(AppDataDir(), "active.txt"))
	if err != nil {
		return `C:\xampp`
	}
	if path := strings.TrimSpace(string(data)); path != "" {
		return path
	}
	return `C:\xampp`
}

// LamppBinPath is not used on Windows — XAMPP services are controlled
// by individual binaries (httpd.exe, mysqld.exe).
func LamppBinPath() string { return "" }

// LamppBinDir returns the directory that should be in PATH for XAMPP CLI tools.
func LamppBinDir() string { return filepath.Join(ActiveXAMPPPath(), "php") }
