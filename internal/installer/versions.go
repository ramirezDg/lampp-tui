package installer

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version holds the display name and SourceForge directory URL for a single
// XAMPP release.
type Version struct {
	Name        string
	DownloadURL string
}

const sourceForgeListURL = "https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/"

// regexes for extracting fields from a single <tr>…</tr> block.
var (
	reVersionName = regexp.MustCompile(`<span class="name">(\d[^<]+)</span>`)
	reDownloads   = regexp.MustCompile(`<span class="count">([\d,]+)</span>`)
	reDirLink     = regexp.MustCompile(`href="(/projects/xampp/files/XAMPP%20Linux/[\d][^"]+/)"`)
)

// FetchVersions fetches the XAMPP Linux file listing from SourceForge and
// returns all versions that have more than 5 recorded downloads.
//
// Pure Go implementation — no shell, gawk, or curl required.
func FetchVersions() ([]Version, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	resp, err := client.Get(sourceForgeListURL)
	if err != nil {
		return nil, fmt.Errorf("fetching version list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sourceforge returned HTTP %d", resp.StatusCode)
	}

	// Cap at 2 MB — the page is typically ~100–200 KB.
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
// Version per valid row. Exported for unit testing.
func parseVersionRows(body string) []Version {
	var versions []Version

	// Split on <tr so each chunk represents one table row.
	for chunk := range strings.SplitSeq(body, "<tr") {
		// Trim to the end of this row.
		if end := strings.Index(chunk, "</tr>"); end >= 0 {
			chunk = chunk[:end]
		}

		// Extract version name — must begin with a digit.
		nm := reVersionName.FindStringSubmatch(chunk)
		if nm == nil {
			continue
		}
		name := strings.TrimSpace(nm[1])
		if !looksLikeVersion(name) {
			continue
		}

		// Extract download count — filter low-traffic entries.
		dl := reDownloads.FindStringSubmatch(chunk)
		if dl == nil {
			continue
		}
		count, err := strconv.Atoi(strings.ReplaceAll(dl[1], ",", ""))
		if err != nil || count <= 5 {
			continue
		}

		// Extract the SourceForge directory link.
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

// looksLikeVersion returns true when s starts with a digit and contains at
// least one dot — e.g. "8.2.12" or "7.4.33".
func looksLikeVersion(s string) bool {
	return len(s) > 0 && s[0] >= '0' && s[0] <= '9' && strings.Contains(s, ".")
}
