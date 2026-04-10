package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"xampp-tui/internal/logger"
)

// downloadDir returns the absolute path used for storing downloaded XAMPP
// installers, following the XDG Base Directory convention.
func downloadDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "downloads") // last-resort fallback
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "xampp-tui", "downloads")
}

// ProgressFunc is called periodically during a download with the number of
// bytes received so far and the total file size (-1 if unknown).
type ProgressFunc func(downloaded, total int64)

// Download downloads the XAMPP installer for the given version into the
// user's data directory. Partial files are removed on failure.
func Download(version string, onProgress ProgressFunc) error {
	dir := downloadDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating download dir: %w", err)
	}

	url := fmt.Sprintf(
		"https://sourceforge.net/projects/xampp/files/XAMPP%%20Linux/%s/xampp-linux-x64-%s-0-installer.run/download",
		version, version,
	)
	logger.Write(fmt.Sprintf("downloading XAMPP %s from %s", version, url))

	resp, err := http.Get(url) //nolint:gosec // URL constructed from a known-safe pattern
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	total := resp.ContentLength // -1 when unknown

	dest := filepath.Join(dir, fmt.Sprintf("xampp-linux-x64-%s-0-installer.run", version))
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
		os.Remove(dest) // clean up partial file
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
// XDG data directory and the legacy ./downloads/ path so that files downloaded
// with older builds of xampp-tui are still discovered.
func DownloadedVersions() []string {
	const prefix = "xampp-linux-x64-"
	const suffix = "-0-installer.run"

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

// downloadSearchDirs returns all directories to scan when looking for
// already-downloaded XAMPP installers.
func downloadSearchDirs() []string {
	dirs := []string{downloadDir()}

	// Legacy path: relative to the working directory (used by older builds).
	if cwd, err := os.Getwd(); err == nil {
		legacy := filepath.Join(cwd, "downloads")
		if legacy != downloadDir() {
			dirs = append(dirs, legacy)
		}
	}

	// Legacy path: next to the running binary (how go run places things).
	if exe, err := os.Executable(); err == nil {
		legacy := filepath.Join(filepath.Dir(exe), "downloads")
		if legacy != downloadDir() {
			dirs = append(dirs, legacy)
		}
	}

	return dirs
}
