package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

// VersionListURL returns the SourceForge directory listing for XAMPP Windows.
func VersionListURL() string {
	return "https://sourceforge.net/projects/xampp/files/XAMPP%20Windows/"
}

// VersionDirPrefix returns the SourceForge path component used in directory
// links on the listing page. Used to build the link-matching regex.
func VersionDirPrefix() string { return "XAMPP%20Windows" }

// InstallerFilename returns the .exe filename for a given XAMPP version.
func InstallerFilename(version string) string {
	return fmt.Sprintf("xampp-windows-x64-%s-0-installer.exe", version)
}

// InstallerFilePrefix is the part of the filename that precedes the version.
func InstallerFilePrefix() string { return "xampp-windows-x64-" }

// InstallerFileSuffix is the part of the filename that follows the version.
func InstallerFileSuffix() string { return "-0-installer.exe" }

// InstallerDownloadURL returns the direct SourceForge download URL for a version.
func InstallerDownloadURL(version string) string {
	return fmt.Sprintf(
		"https://sourceforge.net/projects/xampp/files/XAMPP%%20Windows/%s/%s/download",
		version, InstallerFilename(version),
	)
}

// ExecuteInstaller runs the XAMPP .exe installer in unattended mode.
//
// On Windows the installer executable has a requireAdministrator manifest, so
// it must run as Administrator. Direct exec.Command calls from a non-elevated
// process cannot capture the exit code of an elevated child (Windows launches
// it through the UAC broker, making it a sibling rather than a child).
//
// We work around this by delegating to PowerShell Start-Process with
// -Verb RunAs -Wait, which:
//   1. Triggers the UAC elevation dialog (user must accept).
//   2. Waits for the installer to finish.
//   3. Captures and forwards the exit code so we can detect failures.
func ExecuteInstaller(runFile, targetDir string, onProgress func(string)) error {
	if onProgress != nil {
		onProgress(fmt.Sprintf(
			"Installing to %s …\n\nA UAC (User Account Control) prompt will appear.\nPlease accept it to allow the installation to proceed.",
			targetDir,
		))
	}

	// Escape single quotes in paths for PowerShell string literals.
	psFile := strings.ReplaceAll(runFile, "'", "''")
	psDir := strings.ReplaceAll(targetDir, "'", "''")

	psScript := fmt.Sprintf(
		"$p = Start-Process -FilePath '%s' -ArgumentList '--mode','unattended','--prefix','%s' -Verb RunAs -Wait -PassThru; exit $p.ExitCode",
		psFile, psDir,
	)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = "installation failed — make sure you accepted the UAC prompt and that you have administrator rights"
		}
		return fmt.Errorf("installer: %s: %w", msg, err)
	}
	return nil
}
