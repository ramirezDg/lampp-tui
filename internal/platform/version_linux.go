package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetActiveVersion updates the /opt/lampp symlink to point to targetPath.
// Requires sudo.
func SetActiveVersion(targetPath string) error {
	out, err := exec.Command("sudo", "ln", "-sfn", targetPath, ActiveXAMPPPath()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set active version: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// RemoveInstallation permanently deletes the XAMPP installation at path.
// Requires sudo.
func RemoveInstallation(path string) error {
	out, err := exec.Command("sudo", "rm", "-rf", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("remove installation: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
