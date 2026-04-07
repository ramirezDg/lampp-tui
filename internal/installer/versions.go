package installer

import (
	"fmt"
	"os/exec"
	"strings"
)

// Version holds the name and direct download page URL for a single XAMPP
// release scraped from SourceForge.
type Version struct {
	Name        string
	DownloadURL string
}

// FetchVersions scrapes the XAMPP SourceForge listing and returns all
// releases that have more than 5 recorded downloads, ordered as they appear
// on the page.
func FetchVersions() ([]Version, error) {
	// The awk script extracts version name, download count, and directory link
	// from each <tr> on the SourceForge file listing page.
	bashScript := `
		curl -s https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/ | \
		gawk '
			BEGIN { ver=""; count=""; link=""; }
			/<tr title=/ { ver=""; count=""; link=""; }
			/<a href="\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\// {
				match($0, /<a href="(\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\/)"/, arr)
				if (arr[1] != "") link=arr[1]
			}
			/<span class="name">/ {
				match($0, /<span class="name">([^<]+)<\/span>/, arr)
				if (arr[1] != "") ver=arr[1]
			}
			/<span class="count">/ {
				match($0, /<span class="count">([0-9,]+)<\/span>/, arr)
				gsub(",", "", arr[1])
				count=arr[1]
			}
			/<\/tr>/ {
				if (ver != "" && count != "" && count+0 > 5 && link != "") {
					print ver "|https://sourceforge.net" link
				}
			}'
		`

	out, err := exec.Command("bash", "-c", bashScript).Output()
	if err != nil {
		return nil, fmt.Errorf("version scrape failed: %w", err)
	}

	var versions []Version
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		idx := strings.IndexByte(line, '|')
		if idx <= 0 {
			continue
		}
		versions = append(versions, Version{
			Name:        line[:idx],
			DownloadURL: line[idx+1:],
		})
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no XAMPP versions found with more than 5 downloads")
	}
	return versions, nil
}
