package installer

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
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
}

// regexes for extracting fields from a single <tr>…</tr> block.
var (
	reVersionName = regexp.MustCompile(`<span class="name">(\d[^<]+)</span>`)
	reDownloads   = regexp.MustCompile(`<span class="count">([\d,]+)</span>`)
)

// reDirLink is built at init time using the platform-specific path prefix so
// that the same binary always fetches from the right SourceForge category.
var reDirLink = regexp.MustCompile(
	`href="(/projects/xampp/files/` + platform.VersionDirPrefix() + `/[\d][^"]+/)"`,
)

// FetchVersions fetches the XAMPP file listing from SourceForge and returns
// all versions that have more than 5 recorded downloads.
//
// Pure Go — no shell, gawk, or curl dependency.
func FetchVersions() ([]Version, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Get(platform.VersionListURL())
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
		})
	}

	return versions
}

func looksLikeVersion(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9' && strings.Contains(s, ".")
}
