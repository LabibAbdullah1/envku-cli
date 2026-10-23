//go:build !linux

package installer

// IsRoot returns true on non-Linux systems for development and testing convenience
func IsRoot() bool {
	return true
}
