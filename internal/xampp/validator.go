package xampp

import "os"

// requiredPaths are the filesystem entries that must exist for XAMPP to be
// considered installed. If any one of them is absent, IsInstalled returns false.
var requiredPaths = []string{
	"/opt/lampp/apache2",
	"/opt/lampp/mysql",
	"/opt/lampp/sbin/proftpd",
}

// IsInstalled reports whether XAMPP appears to be installed on the system by
// checking for the presence of the binaries it ships.
func IsInstalled() bool {
	for _, path := range requiredPaths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}
