package xampp

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	lamppBinPath    = "/opt/lampp/bin"
	lamppBinExport  = `export PATH="/opt/lampp/bin:$PATH"`
	lamppBinComment = `# Added by xampp-tui — makes php/mysql point to the active XAMPP version`
)

// DetectShellConfig returns the absolute path to the user's primary shell
// startup file. It checks $SHELL first, then falls back to ~/.profile.
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
		// Prefer .bash_profile on login shells, but .bashrc is more universal.
		rc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rc); err == nil {
			return rc
		}
		return filepath.Join(home, ".bash_profile")
	default:
		return filepath.Join(home, ".profile")
	}
}

// LamppAlreadyInPATH returns true if the runtime PATH already contains
// /opt/lampp/bin (i.e. it was added in a previous shell session).
func LamppAlreadyInPATH() bool {
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		if p == lamppBinPath {
			return true
		}
	}
	return false
}

// EnsureLamppInPATH appends the /opt/lampp/bin export to configPath if it is
// not already present in that file. Returns true if the line was actually
// added, false if it was already there.
func EnsureLamppInPATH(configPath string) (added bool, err error) {
	// Check whether the file already contains a reference to /opt/lampp/bin.
	f, openErr := os.Open(configPath)
	if openErr == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), lamppBinPath) {
				f.Close()
				return false, nil // already configured — nothing to do
			}
		}
		f.Close()
	} else if !os.IsNotExist(openErr) {
		return false, fmt.Errorf("reading %s: %w", configPath, openErr)
	}

	// Fish shell uses a different syntax.
	var line string
	if strings.Contains(configPath, "fish") {
		line = fmt.Sprintf("\n%s\nfish_add_path %s\n", lamppBinComment, lamppBinPath)
	} else {
		line = fmt.Sprintf("\n%s\n%s\n", lamppBinComment, lamppBinExport)
	}

	out, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("opening %s for write: %w", configPath, err)
	}
	defer out.Close()

	if _, err := out.WriteString(line); err != nil {
		return false, fmt.Errorf("writing to %s: %w", configPath, err)
	}
	return true, nil
}
