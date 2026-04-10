package platform

import (
	"os"
	"path/filepath"
)

// AppDataDir returns the base application data directory for xampp-tui,
// following the XDG Base Directory specification.
//
// Default: ~/.local/share/xampp-tui
// Override: $XDG_DATA_HOME/xampp-tui
func AppDataDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", ".xampp-tui")
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "xampp-tui")
}
