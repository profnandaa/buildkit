package workers

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/Microsoft/hcsshim"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// code from https://github.com/containerd/containerd/blob/f20973337f4b2c0562fc5a2becfc9acdbdc7e7b6/integration/client/helpers_windows.go
// see more details here - https://github.com/containerd/containerd/pull/5133

// forceRemoveAll is the same as os.RemoveAll, but is aware of io.containerd.snapshotter.v1.windows
// and uses hcsshim to unmount and delete container layers contained therein, in the correct order,
// when passed a containerd root data directory (i.e. the `--root` directory for containerd).
func forceRemoveAll(path string) error {
	// snapshots/windows/windows.go init()
	const snapshotPlugin = "io.containerd.snapshotter.v1" + "." + "windows"
	// snapshots/windows/windows.go NewSnapshotter()
	snapshotDir := filepath.Join(path, snapshotPlugin, "snapshots")
	if stat, err := os.Stat(snapshotDir); err == nil && stat.IsDir() {
		if err := cleanupWCOWLayers(snapshotDir); err != nil {
			return errors.Wrapf(err, "failed to cleanup WCOW layers in %s", snapshotDir)
		}
	}

	return os.RemoveAll(path)
}

func cleanupWCOWLayers(root string) error {
	// See snapshots/windows/windows.go getSnapshotDir()
	var layerNums []int
	var rmLayerNums []int
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path != root && info.IsDir() {
			name := filepath.Base(path)
			if strings.HasPrefix(name, "rm-") {
				layerNum, err := strconv.Atoi(strings.TrimPrefix(name, "rm-"))
				if err != nil {
					return err
				}
				rmLayerNums = append(rmLayerNums, layerNum)
			} else {
				layerNum, err := strconv.Atoi(name)
				if err != nil {
					return err
				}
				layerNums = append(layerNums, layerNum)
			}
			return filepath.SkipDir
		}

		return nil
	}); err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.IntSlice(rmLayerNums)))
	for _, rmLayerNum := range rmLayerNums {
		if err := cleanupWCOWLayer(filepath.Join(root, "rm-"+strconv.Itoa(rmLayerNum))); err != nil {
			return err
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(layerNums)))
	for _, layerNum := range layerNums {
		if err := cleanupWCOWLayer(filepath.Join(root, strconv.Itoa(layerNum))); err != nil {
			return err
		}
	}

	return nil
}

func cleanupWCOWLayer(layerPath string) error {
	info := hcsshim.DriverInfo{
		HomeDir: filepath.Dir(layerPath),
	}

	// ERROR_DEV_NOT_EXIST is returned if the layer is not currently prepared or activated.
	// ERROR_FLT_INSTANCE_NOT_FOUND is returned if the layer is currently activated but not prepared.
	if err := hcsshim.UnprepareLayer(info, filepath.Base(layerPath)); err != nil {
		if hcserror, ok := err.(*hcsshim.HcsError); !ok || (hcserror.Err != windows.ERROR_DEV_NOT_EXIST && hcserror.Err != syscall.Errno(windows.ERROR_FLT_INSTANCE_NOT_FOUND)) {
			return errors.Wrapf(err, "failed to unprepare %s", layerPath)
		}
	}

	if err := hcsshim.DeactivateLayer(info, filepath.Base(layerPath)); err != nil {
		return errors.Wrapf(err, "failed to deactivate %s", layerPath)
	}

	if err := hcsshim.DestroyLayer(info, filepath.Base(layerPath)); err != nil {
		return errors.Wrapf(err, "failed to destroy %s", layerPath)
	}

	return nil
}
