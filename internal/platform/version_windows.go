package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetActiveVersion writes targetPath to the active.txt config file so that
// ActiveXAMPPPath() returns the new value on subsequent calls.
// On Windows we cannot use symlinks without admin rights, so we use a file.
func SetActiveVersion(targetPath string) error {
	cfgPath := filepath.Join(AppDataDir(), "active.txt")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	return os.WriteFile(cfgPath, []byte(targetPath+"\n"), 0o644)
}

// RemoveInstallation permanently removes the directory at path using
// `rmdir /s /q` (does not require elevation for directories the user owns).
func RemoveInstallation(path string) error {
	out, err := exec.Command("cmd", "/c", "rmdir", "/s", "/q", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("remove installation: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
