package xampp

import (
	"os"
	"regexp"
	"strings"
	"xampp-tui/internal/platform"
)

// apacheLineRe matches Apache error log lines:
//
//	[Day Mon DD HH:MM:SS.usec YYYY] [module:level] [pid N] message
var apacheLineRe = regexp.MustCompile(
	`\[\w+ \w+ +\d+ (\d{2}:\d{2}:\d{2})\.\d+ \d+\] \[[\w./-]+:(\w+)\] \[pid \d+\] (.*)`,
)

// RecentLogs returns the last n formatted lines from the XAMPP Apache error
// log. Each returned string has the form "HH:MM:SS  [level]  message".
// Returns nil if the log file cannot be read.
func RecentLogs(n int) []string {
	raw := tailFile(platform.ApacheLogPath(), n)
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		out = append(out, formatLogLine(line))
	}
	return out
}

// formatLogLine converts a raw Apache error log line into a compact string.
// Falls back to returning the original line when the format is not recognised.
func formatLogLine(line string) string {
	m := apacheLineRe.FindStringSubmatch(line)
	if len(m) == 4 {
		return m[1] + "  [" + m[2] + "]  " + m[3]
	}
	return line
}

// tailFile reads the last n lines of the file at path efficiently by seeking
// to the end and reading backwards in chunks. Returns nil on error.
func tailFile(path string, n int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return nil
	}

	fileSize := fi.Size()

	// Estimate: ~200 bytes per log line (Apache lines are long).
	chunkSize := int64(n * 200)
	if chunkSize > fileSize {
		chunkSize = fileSize
	}

	buf := make([]byte, chunkSize)
	offset := fileSize - chunkSize
	_, err = f.ReadAt(buf, offset)
	if err != nil {
		return nil
	}

	// Normalize Windows CRLF so the regex and display work on all platforms.
	normalized := strings.ReplaceAll(string(buf), "\r\n", "\n")
	lines := strings.Split(strings.TrimRight(normalized, "\n"), "\n")

	// The first line may be a partial read — drop it unless we read from BOF.
	if offset > 0 && len(lines) > 1 {
		lines = lines[1:]
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
