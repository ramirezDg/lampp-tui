package platform

import (
	"os"
	"path/filepath"
)

// AppDataDir returns the base application data directory for xampp-tui.
//
// Default: %APPDATA%\xampp-tui
func AppDataDir() string {
	if dir := os.Getenv("APPDATA"); dir != "" {
		return filepath.Join(dir, "xampp-tui")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "xampp-tui")
	}
	return filepath.Join(home, "AppData", "Roaming", "xampp-tui")
}
