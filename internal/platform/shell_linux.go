package platform

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectShellConfig returns the absolute path to the user's primary shell
// startup file, based on the $SHELL environment variable.
func DetectShellConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	case strings.Contains(shell, "bash"):
		rc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rc); err == nil {
			return rc
		}
		return filepath.Join(home, ".bash_profile")
	default:
		return filepath.Join(home, ".profile")
	}
}

// EnsureInPATH appends an export line for binDir to configPath if it is not
// already present. Returns true when the line was actually added.
func EnsureInPATH(configPath, binDir string) (bool, error) {
	// Check whether the file already contains a reference to binDir.
	f, err := os.Open(configPath)
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), binDir) {
				f.Close()
				return false, nil
			}
		}
		f.Close()
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var line string
	if strings.Contains(configPath, "fish") {
		line = fmt.Sprintf("\n# Added by xampp-tui\nfish_add_path %s\n", binDir)
	} else {
		line = fmt.Sprintf("\n# Added by xampp-tui\nexport PATH=\"%s:$PATH\"\n", binDir)
	}

	out, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("opening %s: %w", configPath, err)
	}
	defer out.Close()

	if _, err := out.WriteString(line); err != nil {
		return false, fmt.Errorf("writing %s: %w", configPath, err)
	}
	return true, nil
}

// IsInCurrentPATH reports whether dir is already in the running process's PATH.
func IsInCurrentPATH(dir string) bool {
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		if p == dir {
			return true
		}
	}
	return false
}
