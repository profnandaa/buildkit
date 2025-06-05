//go:build !windows

package local

import (
	"context"
	gofs "io/fs"

	"github.com/tonistiigi/fsutil"
)

func fsWalk(fs fsutil.FS, ctx context.Context, s string, walkFn gofs.WalkDirFunc) error {
	return fs.Walk(ctx, s, walkFn)
}
