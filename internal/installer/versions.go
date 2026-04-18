package installer

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"xampp-tui/internal/platform"
)

// Version holds the display name and SourceForge directory URL for a single
// XAMPP release.
type Version struct {
	Name        string
	DownloadURL string
	Downloads   int
}

// regexes for extracting fields from a single <tr>…</tr> block.
var (
	reVersionName = regexp.MustCompile(`<span class="name">(\d[^<]+)</span>`)
	reDownloads   = regexp.MustCompile(`<span class="count">([\d,]+)</span>`)
)

// installerFileExt returns ".exe" (Windows) or ".run" (Linux) by inspecting
// the platform suffix. Called once at package init.
func installerFileExt() string {
	s := platform.InstallerFileSuffix()
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		return s[i:]
	}
	return ""
}

// reDirLink matches SourceForge version-directory links in the listing page.
// Handles both relative (/projects/...) and absolute (https://sourceforge.net/projects/...) hrefs,
// since SourceForge serves either format depending on the page.
var reDirLink = regexp.MustCompile(
	`href="(?:https://sourceforge\.net)?(/projects/xampp/files/` +
		platform.VersionDirPrefix() + `/[\d][^"]+/)"`,
)

// reInstallerDownload matches the SourceForge download link for the platform's
// standard (non-portable) XAMPP installer inside a version directory page.
//   - Handles both absolute and relative hrefs.
//   - Uses InstallerFilePrefix() to exclude "portable" variants.
//   - Uses installerFileExt() so it matches .exe on Windows and .run on Linux.
var reInstallerDownload = regexp.MustCompile(
	`href="(?:https://sourceforge\.net)?(/projects/xampp/files/` +
		platform.VersionDirPrefix() + `/[^"]+/` +
		regexp.QuoteMeta(platform.InstallerFilePrefix()) +
		`[^"]+installer` +
		regexp.QuoteMeta(installerFileExt()) +
		`/download)"`,
)

// sfClient is a shared HTTP client used for all SourceForge requests.
// No custom User-Agent is set: Go's default UA works fine for the file listing
// pages and avoids triggering SourceForge's browser-specific responses
// (e.g. brotli encoding or JavaScript-rendered pages).
var sfClient = &http.Client{
	Timeout: 30 * time.Second,
}

func sfGet(url string) (*http.Response, error) {
	return sfClient.Get(url)
}

// ResolveInstallerURL fetches a SourceForge version directory page and returns
// the full download URL of the platform's standard installer found there.
// This resolves naming variants (e.g. -VS16-) that differ across releases,
// and selects the correct file type per platform (.exe or .run).
func ResolveInstallerURL(dirURL string) (string, error) {
	resp, err := sfGet(dirURL)
	if err != nil {
		return "", fmt.Errorf("fetching version directory: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("reading directory page: %w", err)
	}

	m := reInstallerDownload.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf(
			"installer not found in directory listing (prefix=%s ext=%s) — SourceForge page format may have changed",
			platform.InstallerFilePrefix(), installerFileExt(),
		)
	}
	return "https://sourceforge.net" + string(m[1]), nil
}

// topVersionsLimit is the maximum number of versions shown in the picker.
// Only the most-downloaded releases are kept so the list stays manageable.
const topVersionsLimit = 12

// FetchVersions fetches the XAMPP file listing from SourceForge and returns
// the top-downloaded versions (up to topVersionsLimit), sorted by download
// count descending.
//
// Pure Go — no shell, gawk, or curl dependency.
func FetchVersions() ([]Version, error) {
	resp, err := sfGet(platform.VersionListURL())
	if err != nil {
		return nil, fmt.Errorf("fetching version list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sourceforge returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	versions := parseVersionRows(string(body))
	if len(versions) == 0 {
		return nil, fmt.Errorf("no XAMPP versions found — SourceForge page format may have changed")
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Downloads > versions[j].Downloads
	})
	if len(versions) > topVersionsLimit {
		versions = versions[:topVersionsLimit]
	}
	return versions, nil
}

// parseVersionRows splits the HTML body on <tr boundaries and extracts one
// Version per valid row.
func parseVersionRows(body string) []Version {
	var versions []Version

	for chunk := range strings.SplitSeq(body, "<tr") {
		if end := strings.Index(chunk, "</tr>"); end >= 0 {
			chunk = chunk[:end]
		}

		nm := reVersionName.FindStringSubmatch(chunk)
		if nm == nil {
			continue
		}
		name := strings.TrimSpace(nm[1])
		if !looksLikeVersion(name) {
			continue
		}

		dl := reDownloads.FindStringSubmatch(chunk)
		if dl == nil {
			continue
		}
		count, err := strconv.Atoi(strings.ReplaceAll(dl[1], ",", ""))
		if err != nil || count <= 5 {
			continue
		}

		lk := reDirLink.FindStringSubmatch(chunk)
		if lk == nil {
			continue
		}

		versions = append(versions, Version{
			Name:        name,
			DownloadURL: "https://sourceforge.net" + lk[1],
			Downloads:   count,
		})
	}

	return versions
}

func looksLikeVersion(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9' && strings.Contains(s, ".")
}
