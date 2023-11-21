package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

var ErrRequirements = errors.Errorf("missing requirements")

func Tmpdir(t *testing.T, appliers ...fstest.Applier) string {
	t.Helper()

	// We cannot use t.TempDir() to create a temporary directory here because
	// appliers might contain fstest.CreateSocket. If the test name is too long,
	// t.TempDir() could return a path that is longer than 108 characters. This
	// would result in "bind: invalid argument" when we listen on the socket.
	tmpdir, err := os.MkdirTemp("", "buildkit")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tmpdir))
	})

	err = fstest.Apply(appliers...).Apply(tmpdir)
	require.NoError(t, err)

	return tmpdir
}

func RunCmd(cmd *exec.Cmd, logs map[string]*bytes.Buffer) error {
	if logs != nil {
		setCmdLogs(cmd, logs)
	}
	fmt.Fprintf(cmd.Stderr, "> RunCmd %v %+v\n", time.Now(), cmd.String())
	return cmd.Run()
}

func StartCmd(cmd *exec.Cmd, logs map[string]*bytes.Buffer) (func() error, error) {
	if logs != nil {
		setCmdLogs(cmd, logs)
	}

	fmt.Fprintf(cmd.Stderr, "> StartCmd %v %+v\n", time.Now(), cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	eg, ctx := errgroup.WithContext(context.TODO())

	stopped := make(chan struct{})
	stop := make(chan struct{})
	eg.Go(func() error {
		err := cmd.Wait()
		fmt.Fprintf(cmd.Stderr, "> stopped %v %+v %v\n", time.Now(), cmd.ProcessState, cmd.ProcessState.ExitCode())
		close(stopped)
		select {
		case <-stop:
			return nil
		default:
			return err
		}
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
		case <-stopped:
		case <-stop:
			fmt.Fprintf(cmd.Stderr, "> sending sigterm %v\n", time.Now())
			cmd.Process.Signal(syscall.SIGTERM)
			go func() {
				select {
				case <-stopped:
				case <-time.After(20 * time.Second):
					cmd.Process.Kill()
				}
			}()
		}
		return nil
	})

	return func() error {
		close(stop)
		return eg.Wait()
	}, nil
}

// WaitSocket will dial a socket opened by a command passed in as cmd.
// On Linux this socket is typically a Unix socket,
// while on Windows this will be a named pipe.
func WaitSocket(address string, d time.Duration, cmd *exec.Cmd) error {
	return waitSocket(address, d, cmd)
}

func LookupBinary(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return errors.Wrapf(ErrRequirements, "failed to lookup %s binary", name)
	}
	return nil
}

type MultiCloser struct {
	fns []func() error
}

func (mc *MultiCloser) F() func() error {
	return func() error {
		var err error
		for i := range mc.fns {
			if err1 := mc.fns[len(mc.fns)-1-i](); err == nil {
				err = err1
			}
		}
		mc.fns = nil
		return err
	}
}

func (mc *MultiCloser) Append(f func() error) {
	mc.fns = append(mc.fns, f)
}

func setCmdLogs(cmd *exec.Cmd, logs map[string]*bytes.Buffer) {
	b := new(bytes.Buffer)
	logs["stdout: "+cmd.String()] = b
	cmd.Stdout = &lockingWriter{Writer: b}
	b = new(bytes.Buffer)
	logs["stderr: "+cmd.String()] = b
	cmd.Stderr = &lockingWriter{Writer: b}
}

type lockingWriter struct {
	mu sync.Mutex
	io.Writer
}

func (w *lockingWriter) Write(dt []byte) (int, error) {
	w.mu.Lock()
	n, err := w.Writer.Write(dt)
	w.mu.Unlock()
	return n, err
}
