//go:build linux

package installer

import "os"

// IsRoot checks if the application is running under root UID (0) on Linux
func IsRoot() bool {
	return os.Geteuid() == 0
}
