package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"xampp-tui/internal/logger"
	"xampp-tui/internal/platform"
)

// InstallProgressFunc receives status messages during an installation.
type InstallProgressFunc func(msg string)

// RunInstaller executes the previously-downloaded installer for the given
// version, installing XAMPP to platform.XAMPPBaseDir()/{version}/.
// If the installer fails, any partially-created target directory is removed.
func RunInstaller(version string, onProgress InstallProgressFunc) error {
	runFile := filepath.Join(downloadDir(), platform.InstallerFilename(version))

	if _, err := os.Stat(runFile); err != nil {
		return fmt.Errorf("installer file not found: %s", runFile)
	}

	if err := os.Chmod(runFile, 0o755); err != nil {
		return fmt.Errorf("chmod installer: %w", err)
	}

	targetDir := filepath.Join(platform.XAMPPBaseDir(), version)
	logger.Write(fmt.Sprintf("running XAMPP installer %s → %s", version, targetDir))

	err := platform.ExecuteInstaller(runFile, targetDir, func(msg string) {
		if onProgress != nil {
			onProgress(msg)
		}
	})

	if err != nil {
		logger.Write("installer error: " + err.Error())
		// Clean up any partial installation.
		exec.Command("sudo", "rm", "-rf", targetDir).Run() //nolint:errcheck
		return err
	}

	logger.Write(fmt.Sprintf("XAMPP %s installed at %s", version, targetDir))
	return nil
}
