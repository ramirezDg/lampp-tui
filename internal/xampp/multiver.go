package xampp

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"xampp-tui/internal/platform"
)

// InstalledVersion describes a single XAMPP installation found on the system.
type InstalledVersion struct {
	Version      string
	Path         string
	PHPVersion   string
	MySQLVersion string
	IsActive     bool
}

// ScanInstalledVersions finds all XAMPP versions installed on the system.
// It scans platform.XAMPPBaseDir()/{version}/ directories and falls back to
// checking the active XAMPP path for legacy single-version installs.
func ScanInstalledVersions() []InstalledVersion {
	activePath := GetActivePath()
	var versions []InstalledVersion

	if entries, err := os.ReadDir(platform.XAMPPBaseDir()); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(platform.XAMPPBaseDir(), e.Name())
			if !isXAMPPDir(path) {
				continue
			}
			versions = append(versions, newInstalledVersion(e.Name(), path, activePath))
		}
	}

	// Legacy fallback: active path exists but is not under XAMPPBaseDir.
	if _, err := os.Stat(platform.ActiveXAMPPPath()); err == nil {
		real, _ := filepath.EvalSymlinks(platform.ActiveXAMPPPath())
		if !strings.HasPrefix(real, platform.XAMPPBaseDir()) {
			covered := false
			for _, v := range versions {
				vReal, _ := filepath.EvalSymlinks(v.Path)
				if vReal != "" && vReal == real {
					covered = true
					break
				}
			}
			if !covered {
				versions = append(versions, newInstalledVersion("default", platform.ActiveXAMPPPath(), activePath))
			}
		}
	}

	return versions
}

// GetActivePath returns the resolved filesystem path of the active XAMPP.
func GetActivePath() string {
	real, err := filepath.EvalSymlinks(platform.ActiveXAMPPPath())
	if err != nil {
		if _, err2 := os.Stat(platform.ActiveXAMPPPath()); err2 == nil {
			return platform.ActiveXAMPPPath()
		}
		return ""
	}
	return real
}

// SwitchVersion updates the active XAMPP path to point to the installation
// at targetPath.
func SwitchVersion(targetPath string) error {
	return platform.SetActiveVersion(targetPath)
}

// UninstallVersion permanently removes the XAMPP installation at path.
// It refuses to remove the currently active version.
func UninstallVersion(path string) error {
	activePath := GetActivePath()
	realPath, _ := filepath.EvalSymlinks(path)
	realActive, _ := filepath.EvalSymlinks(activePath)

	if path == activePath || (realPath != "" && realActive != "" && realPath == realActive) {
		return errActiveVersion
	}
	return platform.RemoveInstallation(path)
}

var errActiveVersion = errStr("cannot uninstall the active XAMPP version — switch to another version first")

type errStr string

func (e errStr) Error() string { return string(e) }

// GetVersionInfo returns PHP and MySQL version strings for the XAMPP
// installation rooted at basePath.
func GetVersionInfo(basePath string) (phpVer, mysqlVer string) {
	return phpVersion(basePath), mysqlVersion(basePath)
}

// ─── internal helpers ─────────────────────────────────────────────────────────

func isXAMPPDir(path string) bool {
	// Apache directory is "apache2" on Linux, "apache" on Windows.
	// PHP/MySQL binary layout also differs per platform.
	candidates := []string{
		"apache2",
		"apache",
		filepath.Join("bin", "php"),
		filepath.Join("php", "php.exe"),
		filepath.Join("bin", "mysql"),
		filepath.Join("mysql", "bin", "mysql.exe"),
	}
	for _, sub := range candidates {
		if _, err := os.Stat(filepath.Join(path, sub)); err == nil {
			return true
		}
	}
	return false
}

func newInstalledVersion(name, path, activePath string) InstalledVersion {
	real, _ := filepath.EvalSymlinks(path)
	activeReal, _ := filepath.EvalSymlinks(activePath)

	isActive := path == activePath ||
		(real != "" && activeReal != "" && real == activeReal) ||
		(activePath == "" && path == platform.ActiveXAMPPPath())

	php, mysql := GetVersionInfo(path)
	return InstalledVersion{
		Version:      name,
		Path:         path,
		PHPVersion:   php,
		MySQLVersion: mysql,
		IsActive:     isActive,
	}
}

func phpVersion(basePath string) string {
	bin := platform.PHPBin(basePath)
	out, err := exec.Command(bin, "-r",
		"echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION.'.'.PHP_RELEASE_VERSION;",
	).Output()
	if err != nil {
		return "—"
	}
	return strings.TrimSpace(string(out))
}

var mysqlVerRe = regexp.MustCompile(`Distrib\s+([\d.]+)`)

func mysqlVersion(basePath string) string {
	bin := platform.MySQLBin(basePath)
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return "—"
	}
	if m := mysqlVerRe.FindSubmatch(out); len(m) > 1 {
		return string(m[1])
	}
	return strings.TrimSpace(string(out))
}
