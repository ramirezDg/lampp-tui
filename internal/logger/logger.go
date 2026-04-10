package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// logDir returns the absolute path for the application log directory,
// following the XDG Base Directory convention (~/.local/share/xampp-tui/logs).
func logDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "logs") // last-resort fallback
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "xampp-tui", "logs")
}

// Write appends msg to the application log file, creating the log directory
// when it does not exist. Errors are printed to stderr rather than panicking
// so that a logging failure never crashes the application.
func Write(msg string) {
	dir := logDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "logger: cannot create log dir: %v\n", err)
		return
	}
	path := filepath.Join(dir, "app.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: cannot open log file: %v\n", err)
		return
	}
	defer f.Close()
	f.WriteString(msg + "\n") //nolint:errcheck
}
