package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// logDir is resolved once relative to the running binary so the path is
// stable regardless of the working directory the user launches the TUI from.
func logDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "logs"
	}
	return filepath.Join(filepath.Dir(exe), "logs")
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
	f.WriteString(msg + "\n")
}
