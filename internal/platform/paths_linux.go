package platform

import "path/filepath"

// XAMPPBaseDir is the parent directory for all versioned XAMPP installations.
func XAMPPBaseDir() string { return "/opt/xampp" }

// ActiveXAMPPPath is the canonical path to the currently active XAMPP version.
// On Linux this is a symlink maintained by xampp-tui.
func ActiveXAMPPPath() string { return "/opt/lampp" }

// LamppBinPath returns the path to the lampp service-control script.
func LamppBinPath() string { return filepath.Join(ActiveXAMPPPath(), "lampp") }

// LamppBinDir returns the directory that should be in PATH for XAMPP CLI tools.
func LamppBinDir() string { return filepath.Join(ActiveXAMPPPath(), "bin") }
