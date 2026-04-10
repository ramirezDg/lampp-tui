package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xampp-tui/internal/logger"
)

// XAMPPBaseDir is where versioned XAMPP installations are placed.
const XAMPPBaseDir = "/opt/xampp"

// InstallProgressFunc receives status messages during an installation.
type InstallProgressFunc func(msg string)

// RunInstaller executes the previously-downloaded .run file for the given
// version, installing XAMPP to /opt/xampp/{version}/ in unattended mode.
// If the installer fails, any partially-created target directory is removed.
func RunInstaller(version string, onProgress InstallProgressFunc) error {
	runFile := filepath.Join(downloadDir(),
		fmt.Sprintf("xampp-linux-x64-%s-0-installer.run", version))

	if _, err := os.Stat(runFile); err != nil {
		return fmt.Errorf("installer file not found: %s", runFile)
	}

	if err := os.Chmod(runFile, 0o755); err != nil {
		return fmt.Errorf("chmod installer: %w", err)
	}

	targetDir := filepath.Join(XAMPPBaseDir, version)
	if onProgress != nil {
		onProgress(fmt.Sprintf("Installing XAMPP %s to %s …", version, targetDir))
	}
	logger.Write(fmt.Sprintf("running XAMPP installer %s → %s", version, targetDir))

	cmd := exec.Command("sudo", runFile,
		"--mode", "unattended",
		"--prefix", targetDir,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		logger.Write("installer error: " + msg)
		// Clean up any partial installation so the directory doesn't appear
		// as a valid (but broken) XAMPP version in subsequent scans.
		exec.Command("sudo", "rm", "-rf", targetDir).Run() //nolint:errcheck
		return fmt.Errorf("installer failed: %w", err)
	}

	logger.Write(fmt.Sprintf("XAMPP %s installed at %s", version, targetDir))
	return nil
}
