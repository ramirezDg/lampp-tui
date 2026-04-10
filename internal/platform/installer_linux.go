package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

// VersionListURL returns the SourceForge directory listing for XAMPP Linux.
func VersionListURL() string {
	return "https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/"
}

// VersionDirPrefix returns the SourceForge path component used in directory
// links on the listing page. Used to build the link-matching regex.
func VersionDirPrefix() string { return "XAMPP%20Linux" }

// InstallerFilename returns the .run filename for a given XAMPP version.
func InstallerFilename(version string) string {
	return fmt.Sprintf("xampp-linux-x64-%s-0-installer.run", version)
}

// InstallerFilePrefix is the part of the filename that precedes the version.
func InstallerFilePrefix() string { return "xampp-linux-x64-" }

// InstallerFileSuffix is the part of the filename that follows the version.
func InstallerFileSuffix() string { return "-0-installer.run" }

// InstallerDownloadURL returns the direct SourceForge download URL for a version.
func InstallerDownloadURL(version string) string {
	return fmt.Sprintf(
		"https://sourceforge.net/projects/xampp/files/XAMPP%%20Linux/%s/%s/download",
		version, InstallerFilename(version),
	)
}

// ExecuteInstaller runs the XAMPP .run installer for version, installing to
// targetDir. Requires sudo. Progress messages are sent to onProgress.
func ExecuteInstaller(runFile, targetDir string, onProgress func(string)) error {
	if onProgress != nil {
		onProgress(fmt.Sprintf("Installing to %s …", targetDir))
	}
	cmd := exec.Command("sudo", runFile, "--mode", "unattended", "--prefix", targetDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
