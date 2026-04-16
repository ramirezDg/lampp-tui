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

// ConfigPaths returns the paths to the main XAMPP config files.
// Order matches the service list: Apache, MySQL, PHP (FTP not standard on Windows).
func ConfigPaths() []string {
	base := ActiveXAMPPPath()
	return []string{
		filepath.Join(base, "apache", "conf", "httpd.conf"),
		filepath.Join(base, "mysql", "bin", "my.ini"),
		filepath.Join(base, "php", "php.ini"),
	}
}

// ApacheLogPath returns the path to the Apache error log.
func ApacheLogPath() string {
	return filepath.Join(ActiveXAMPPPath(), "apache", "logs", "error.log")
}

// SwitchVersionNote returns a short UI note shown when switching versions.
func SwitchVersionNote() string {
	return "Only active.txt config is updated.\nNo PATH or registry is modified."
}

// PHPBin returns the path to the PHP CLI binary inside a given XAMPP install.
func PHPBin(basePath string) string { return filepath.Join(basePath, "php", "php.exe") }

// MySQLBin returns the path to the MySQL client binary inside a given XAMPP install.
func MySQLBin(basePath string) string { return filepath.Join(basePath, "mysql", "bin", "mysql.exe") }
