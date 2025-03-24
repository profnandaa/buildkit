//go:build !windows

package workers

import "os"

func forceRemoveAll(path string) error {
	os.RemoveAll(path)
}
