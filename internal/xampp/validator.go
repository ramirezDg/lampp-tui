package xampp

import (
	"os"
	"path/filepath"
	"xampp-tui/internal/platform"
)

// IsInstalled reports whether XAMPP appears to be installed at the active
// path by checking for the presence of its core subdirectories.
func IsInstalled() bool {
	base := platform.ActiveXAMPPPath()
	for _, sub := range []string{"apache2", "mysql"} {
		if _, err := os.Stat(filepath.Join(base, sub)); err != nil {
			return false
		}
	}
	return true
}
