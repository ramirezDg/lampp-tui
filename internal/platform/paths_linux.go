package platform

import "path/filepath"

// XAMPPBaseDir is the parent directory for all versioned XAMPP installations.
func XAMPPBaseDir() string { return "/opt/xampp" }

// ActiveXAMPPPath is the canonical path to the currently active XAMPP version.
// On Linux this is a symlink maintained by xampp-tui.
func ActiveXAMPPPath() string { return "/opt/lampp" }

// LamppBinPath returns the path to the lampp service-control script.
func LamppBinPath() string { return filepath.Join(ActiveXAMPPPath(), "lampp") }

// LamppBinDir returns the directory that should be in PATH for XAMPP CLI tools.
func LamppBinDir() string { return filepath.Join(ActiveXAMPPPath(), "bin") }

// ConfigPaths returns the paths to the main XAMPP config files.
// Order matches the service list: Apache, MySQL, FTP.
func ConfigPaths() []string {
	base := ActiveXAMPPPath()
	return []string{
		filepath.Join(base, "etc", "httpd.conf"),
		filepath.Join(base, "etc", "my.cnf"),
		filepath.Join(base, "etc", "proftpd.conf"),
	}
}

// ApacheLogPath returns the path to the Apache error log.
func ApacheLogPath() string { return filepath.Join(ActiveXAMPPPath(), "logs", "error_log") }

// SwitchVersionNote returns a short UI note shown when switching versions.
func SwitchVersionNote() string {
	return "Only /opt/lampp symlink is updated.\nNo PATH or shell config is modified."
}

// PHPBin returns the path to the PHP CLI binary inside a given XAMPP install.
func PHPBin(basePath string) string { return filepath.Join(basePath, "bin", "php") }

// MySQLBin returns the path to the MySQL client binary inside a given XAMPP install.
func MySQLBin(basePath string) string { return filepath.Join(basePath, "bin", "mysql") }
