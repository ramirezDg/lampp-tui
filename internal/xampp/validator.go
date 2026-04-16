package xampp

import (
	"os"
	"path/filepath"
	"xampp-tui/internal/platform"
)

// IsInstalled reports whether XAMPP appears to be installed at the active
// path by checking for the presence of its core subdirectories.
// Apache directory is "apache2" on Linux and "apache" on Windows.
func IsInstalled() bool {
	base := platform.ActiveXAMPPPath()
	apacheFound := false
	for _, sub := range []string{"apache2", "apache"} {
		if _, err := os.Stat(filepath.Join(base, sub)); err == nil {
			apacheFound = true
			break
		}
	}
	if !apacheFound {
		return false
	}
	if _, err := os.Stat(filepath.Join(base, "mysql")); err != nil {
		return false
	}
	return true
}
