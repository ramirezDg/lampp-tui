package xampp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Multi-version path constants.
const (
	// XAMPPBaseDir is the parent directory for versioned XAMPP installations.
	XAMPPBaseDir = "/opt/xampp"

	// LamppLink is the canonical path (possibly a symlink) to the active XAMPP.
	LamppLink = "/opt/lampp"
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
// It scans /opt/xampp/{version}/ directories and falls back to checking
// /opt/lampp for legacy single-version installations.
func ScanInstalledVersions() []InstalledVersion {
	activePath := GetActivePath()
	var versions []InstalledVersion

	// Scan versioned installs under /opt/xampp/
	if entries, err := os.ReadDir(XAMPPBaseDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(XAMPPBaseDir, e.Name())
			if !isXAMPPDir(path) {
				continue
			}
			versions = append(versions, newInstalledVersion(e.Name(), path, activePath))
		}
	}

	// Legacy fallback: /opt/lampp exists and is not a versioned symlink
	if _, err := os.Stat(LamppLink); err == nil {
		real, _ := filepath.EvalSymlinks(LamppLink)

		if !strings.HasPrefix(real, XAMPPBaseDir) {
			covered := false
			for _, v := range versions {
				vReal, _ := filepath.EvalSymlinks(v.Path)
				if vReal != "" && vReal == real {
					covered = true
					break
				}
			}
			if !covered {
				versions = append(versions, newInstalledVersion("default", LamppLink, activePath))
			}
		}
	}

	return versions
}

// GetActivePath returns the resolved filesystem path of the active XAMPP.
func GetActivePath() string {
	real, err := filepath.EvalSymlinks(LamppLink)
	if err != nil {
		if _, err2 := os.Stat(LamppLink); err2 == nil {
			return LamppLink
		}
		return ""
	}
	return real
}

// SwitchVersion atomically updates /opt/lampp to point to the installation
// at path. Requires sudo.
func SwitchVersion(path string) error {
	out, err := exec.Command("sudo", "ln", "-sfn", path, LamppLink).CombinedOutput()
	if err != nil {
		return fmt.Errorf("switch version: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// GetVersionInfo returns PHP and MySQL version strings for the XAMPP
// installation rooted at basePath.
func GetVersionInfo(basePath string) (phpVer, mysqlVer string) {
	return phpVersion(basePath), mysqlVersion(basePath)
}

// ─── internal helpers ─────────────────────────────────────────────────────────

func isXAMPPDir(path string) bool {
	for _, sub := range []string{"apache2", filepath.Join("bin", "php"), filepath.Join("bin", "mysql")} {
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
		(activePath == "" && path == LamppLink)

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
	bin := filepath.Join(basePath, "bin", "php")
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
	bin := filepath.Join(basePath, "bin", "mysql")
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return "—"
	}
	if m := mysqlVerRe.FindSubmatch(out); len(m) > 1 {
		return string(m[1])
	}
	return strings.TrimSpace(string(out))
}
