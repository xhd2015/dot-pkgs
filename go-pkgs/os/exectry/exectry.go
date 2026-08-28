// Package exectry helps with Linux Docker overlayfs ETXTBSY when a process
// fork/execs a binary that was just written (even after write→rename).
package exectry

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// IsTextFileBusy reports whether err is (or wraps) ETXTBSY / "text file busy".
func IsTextFileBusy(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.ETXTBSY) {
		return true
	}
	return strings.Contains(err.Error(), "text file busy")
}

// Output runs name with args, retrying briefly on ETXTBSY.
func Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, _, err := OutputStderr(ctx, name, args...)
	return out, err
}

// OutputStderr runs name with args, retrying briefly on ETXTBSY.
// stderr is the captured stderr from the last attempt.
func OutputStderr(ctx context.Context, name string, args ...string) (stdout []byte, stderr string, err error) {
	var buf bytes.Buffer
	for attempt := 0; attempt < 8; attempt++ {
		buf.Reset()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stderr = &buf
		stdout, err = cmd.Output()
		stderr = buf.String()
		if err == nil || !IsTextFileBusy(err) {
			return stdout, stderr, err
		}
		time.Sleep(time.Duration(20*(attempt+1)) * time.Millisecond)
	}
	return stdout, stderr, err
}

// WriteExecutable writes body to path via temp file + fsync + rename so an
// immediate fork/exec is less likely to hit ETXTBSY on overlayfs.
func WriteExecutable(path string, body []byte) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := f.Write(body); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if dirf, err := os.Open(filepath.Dir(path)); err == nil {
		_ = dirf.Sync()
		dirf.Close()
	}
	return nil
}
