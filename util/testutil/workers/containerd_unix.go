//go:build !windows

package workers

import "os"

func forceRemoveAll(path string) error {
	return os.RemoveAll(path)
}
