package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"xampp-tui/internal/platform"
)

// Write appends msg to the application log file. The log directory is created
// on first use. Errors are printed to stderr so a logging failure never
// crashes the application.
func Write(msg string) {
	dir := filepath.Join(platform.AppDataDir(), "logs")
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
