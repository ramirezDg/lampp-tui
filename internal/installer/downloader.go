package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"xampp-tui/internal/logger"
	"xampp-tui/internal/platform"
)

// ProgressFunc is called periodically during a download with the number of
// bytes received so far and the total file size (-1 if unknown).
type ProgressFunc func(downloaded, total int64)

// downloadDir returns the absolute path used for storing downloaded installers.
func downloadDir() string {
	return filepath.Join(platform.AppDataDir(), "downloads")
}

// Download downloads the XAMPP installer for the given version. Partial files
// are removed automatically if the download fails or the file looks invalid.
func Download(version string, onProgress ProgressFunc) error {
	dir := downloadDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating download dir: %w", err)
	}

	url := platform.InstallerDownloadURL(version)
	logger.Write(fmt.Sprintf("downloading XAMPP %s from %s", version, url))

	resp, err := http.Get(url) //nolint:gosec // URL constructed from a known-safe pattern
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	total := resp.ContentLength

	dest := filepath.Join(dir, platform.InstallerFilename(version))
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	buf := make([]byte, 32*1024)
	var downloaded int64
	var writeErr error

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				writeErr = fmt.Errorf("writing file: %w", werr)
				break
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
			writeErr = fmt.Errorf("reading response: %w", readErr)
			break
		}
	}

	f.Close()

	if writeErr != nil {
		os.Remove(dest)
		return writeErr
	}

	// Sanity check: XAMPP installers are always > 50 MB.
	if downloaded < 50*1024*1024 {
		os.Remove(dest)
		return fmt.Errorf("downloaded file too small (%d bytes) — possible server error", downloaded)
	}

	logger.Write(fmt.Sprintf("download complete: %s (%d bytes)", dest, downloaded))
	return nil
}

// Install is the fire-and-forget wrapper used outside of the TUI flow.
func Install(version string) error {
	return Download(version, nil)
}

// DownloadedVersions returns the version strings of all XAMPP installers that
// have been downloaded and are ready to install. It scans both the current
// data directory and legacy locations from older builds.
func DownloadedVersions() []string {
	prefix := platform.InstallerFilePrefix()
	suffix := platform.InstallerFileSuffix()

	seen := make(map[string]bool)
	var versions []string

	for _, dir := range downloadSearchDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
				continue
			}
			ver := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
			if ver != "" && !seen[ver] {
				seen[ver] = true
				versions = append(versions, ver)
			}
		}
	}
	return versions
}

// downloadSearchDirs returns all directories to scan for downloaded installers,
// including legacy paths from older builds.
func downloadSearchDirs() []string {
	dirs := []string{downloadDir()}

	if cwd, err := os.Getwd(); err == nil {
		if legacy := filepath.Join(cwd, "downloads"); legacy != downloadDir() {
			dirs = append(dirs, legacy)
		}
	}

	if exe, err := os.Executable(); err == nil {
		if legacy := filepath.Join(filepath.Dir(exe), "downloads"); legacy != downloadDir() {
			dirs = append(dirs, legacy)
		}
	}

	return dirs
}
