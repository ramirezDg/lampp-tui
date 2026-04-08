package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"xampp-tui/internal/logger"
)

const downloadDir = "./downloads"

// ProgressFunc is called periodically during a download with the number of
// bytes received so far and the total file size (-1 if unknown).
type ProgressFunc func(downloaded, total int64)

// Download downloads the XAMPP installer for the given version into the local
// downloads directory via HTTP. If onProgress is non-nil it is called each
// time a chunk is written to disk.
func Download(version string, onProgress ProgressFunc) error {
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return fmt.Errorf("creating download dir: %w", err)
	}

	url := fmt.Sprintf(
		"https://sourceforge.net/projects/xampp/files/XAMPP%%20Linux/%s/xampp-linux-x64-%s-0-installer.run/download",
		version, version,
	)
	logger.Write(fmt.Sprintf("downloading XAMPP %s from %s", version, url))

	resp, err := http.Get(url) //nolint:gosec // URL is constructed from a known safe pattern
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	total := resp.ContentLength // -1 when unknown

	dest := filepath.Join(downloadDir, fmt.Sprintf("xampp-linux-x64-%s-0-installer.run", version))
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	var downloaded int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return fmt.Errorf("writing file: %w", werr)
			}
			downloaded += int64(n)
			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("reading response: %w", readErr)
		}
	}

	logger.Write(fmt.Sprintf("download complete: %s (%d bytes)", dest, downloaded))
	return nil
}

// Install is the fire-and-forget wrapper used outside of the TUI flow.
func Install(version string) error {
	return Download(version, nil)
}
