package xampp

import "xampp-tui/internal/platform"

// DetectShellConfig returns the path to the user's primary shell startup file.
func DetectShellConfig() string { return platform.DetectShellConfig() }

// EnsureLamppInPATH adds the active XAMPP bin directory to the user's shell
// config (or Windows Registry) if not already present.
// Returns true when the entry was actually added.
func EnsureLamppInPATH(configPath string) (bool, error) {
	return platform.EnsureInPATH(configPath, platform.LamppBinDir())
}

// LamppAlreadyInPATH reports whether the XAMPP bin directory is in the
// current process's PATH.
func LamppAlreadyInPATH() bool {
	return platform.IsInCurrentPATH(platform.LamppBinDir())
}
