package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"xampp-tui/internal/logger"
)

const downloadDir = "./downloads"

// Install downloads the XAMPP installer for the given version name into the
// local downloads directory using wget.
func Install(version string) error {
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		logger.Write(fmt.Sprintf("error creating download dir: %v", err))
		return fmt.Errorf("error creating download dir: %w", err)
	}

	url := fmt.Sprintf(
		"https://sourceforge.net/projects/xampp/files/XAMPP%%20Linux/%s/xampp-linux-x64-%s-0-installer.run/download",
		version, version,
	)

	logger.Write(fmt.Sprintf("Installing XAMPP version: %s", version))
	logger.Write(fmt.Sprintf("Downloading from: %s", url))
	logger.Write(fmt.Sprintf("Destination: %s", downloadDir))
	fmt.Printf("Downloading XAMPP %s...\n", version)

	cmd := exec.Command("wget", "--content-disposition", "-P", downloadDir, url)
	if err := cmd.Run(); err != nil {
		logger.Write(fmt.Sprintf("download failed: %v", err))
		return fmt.Errorf("download failed: %w", err)
	}
	logger.Write("wget completed")

	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		logger.Write(fmt.Sprintf("error reading download dir: %v", err))
		return fmt.Errorf("error reading download dir: %w", err)
	}

	found := false
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".run") {
			logger.Write(fmt.Sprintf("downloaded file: %s", e.Name()))
			fmt.Printf("Downloaded: %s\n", e.Name())
			found = true
		}
	}
	if !found {
		msg := "warning: no .run file found in downloads folder after download"
		logger.Write(msg)
		fmt.Println(msg)
	} else {
		logger.Write("download completed")
		fmt.Println("Download completed.")
	}
	return nil
}
