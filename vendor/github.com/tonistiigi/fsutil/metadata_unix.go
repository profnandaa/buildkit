//go:build !windows
// +build !windows

package fsutil

// no special files on unix
func isMetadataFile(path string) bool {
	return false
}
