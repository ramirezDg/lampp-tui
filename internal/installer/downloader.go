package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xampp-tui/internal/logger"
	"xampp-tui/internal/platform"
)

// ProgressFunc is called periodically during a download with the number of
// bytes received so far and the total file size (-1 if unknown).
type ProgressFunc func(downloaded, total int64)

// minInstallerBytes is the minimum acceptable size for a downloaded XAMPP
// installer. Files smaller than this are considered server-error responses
// (HTML pages, redirects, etc.) and are removed automatically.
const minInstallerBytes = 50 * 1024 * 1024 // 50 MB

// downloadDir returns the absolute path used for storing downloaded installers.
func downloadDir() string {
	return filepath.Join(platform.AppDataDir(), "downloads")
}

// DownloadDir returns the absolute path used for storing downloaded installers.
// Exported so that UI layers can display the destination to the user.
func DownloadDir() string {
	return downloadDir()
}

// removeFile attempts to delete a file, retrying a few times on Windows where
// file-system locks can delay deletion briefly after a handle is closed.
func removeFile(path string) {
	for range 3 {
		if os.Remove(path) == nil {
			return
		}
		time.Sleep(150 * time.Millisecond)
	}
}

// Download downloads the XAMPP installer for the given version.
// dirURL is the SourceForge directory URL for the version (from FetchVersions);
// when non-empty it is used to resolve the exact installer URL regardless of
// the filename variant (e.g. -VS16-). Falls back to the platform-constructed
// URL when dirURL is empty.
// Partial or invalid files are removed automatically. The saved file always
// uses InstallerFilename(version) so that DownloadedVersions() can find it.
func Download(version string, dirURL string, onProgress ProgressFunc) error {
	dir := downloadDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating download dir: %w", err)
	}

	var url string
	if dirURL != "" {
		resolved, err := ResolveInstallerURL(dirURL)
		if err != nil {
			logger.Write(fmt.Sprintf("resolve failed (%v), falling back to platform URL", err))
			url = platform.InstallerDownloadURL(version)
		} else {
			url = resolved
		}
	} else {
		url = platform.InstallerDownloadURL(version)
	}
	logger.Write(fmt.Sprintf("downloading XAMPP %s from %s", version, url))

	resp, err := sfGet(url)
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
		removeFile(dest)
		return writeErr
	}

	if downloaded < minInstallerBytes {
		removeFile(dest)
		return fmt.Errorf("downloaded file too small (%d bytes) — SourceForge may have returned an error page", downloaded)
	}

	logger.Write(fmt.Sprintf("download complete: %s (%d bytes)", dest, downloaded))
	return nil
}

// Install is the fire-and-forget wrapper used outside of the TUI flow.
func Install(version string) error {
	return Download(version, "", nil)
}

// DownloadedVersions returns the version strings of all XAMPP installers that
// have been downloaded and are ready to install. Files smaller than
// minInstallerBytes are treated as stale/corrupt and removed automatically.
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

			// Validate size: stale files from failed downloads are tiny HTML pages.
			info, err := e.Info()
			if err != nil || info.Size() < minInstallerBytes {
				removeFile(filepath.Join(dir, name))
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
