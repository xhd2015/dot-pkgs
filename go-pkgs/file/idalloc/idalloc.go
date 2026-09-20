// Package idalloc allocates sequential integer ids from a persistent state
// file, serialized across processes with an advisory lock on a sibling lock
// file so concurrent binaries never observe or reuse the same id.
package idalloc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	// stateFileMode is the permission for the state file and lock file.
	stateFileMode os.FileMode = 0o600
	// dirMode is the permission for the state file's directory when the
	// package creates it.
	dirMode os.FileMode = 0o700
	// lockSuffix replaces the state file extension to derive the lock file.
	lockSuffix = ".lock"
	// tempPattern is the CreateTemp pattern for atomic state replacement.
	tempPattern = ".id-*.tmp"
)

// Option customizes an Allocator.
type Option func(*Allocator)

// state is the on-disk counter format: {"last": N}.
type state struct {
	Last int64 `json:"last"`
}

// Allocator allocates sequential ids from a single state file. The zero value
// is not usable; construct with New. Methods are safe for concurrent use: the
// allocator holds no file descriptors between calls and every Next call takes
// the lock for the duration of one read-modify-write.
type Allocator struct {
	stateFile string
	lockFile  string
}

// New returns an allocator for stateFile, for example "~/.learning/id.json".
// The lock file is inferred from the state file by replacing its extension
// with ".lock" ("id.json" -> "id.lock"; extensionless names simply gain the
// suffix: "ids" -> "ids.lock").
func New(stateFile string, opts ...Option) *Allocator {
	a := &Allocator{
		stateFile: stateFile,
		lockFile:  inferLockFile(stateFile),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func inferLockFile(stateFile string) string {
	return strings.TrimSuffix(stateFile, filepath.Ext(stateFile)) + lockSuffix
}

// Next returns one past the persisted counter, raised to at least min+1, and
// persists the new value. Passing the largest id already present in the
// caller's data as min makes a deleted or stale state file recoverable
// without ever reusing an id. The state file is replaced atomically via a
// temporary file and rename, so the lock file is never renamed and readers
// observe either the old or the new complete state.
func (a *Allocator) Next(min int64) (int64, error) {
	if err := ensureDir(filepath.Dir(a.stateFile)); err != nil {
		return 0, err
	}
	lock, err := os.OpenFile(a.lockFile, os.O_CREATE|os.O_RDWR, stateFileMode)
	if err != nil {
		return 0, fmt.Errorf("open id lock: %w", err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return 0, fmt.Errorf("lock id allocator: %w", err)
	}
	defer func() { _ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) }()

	last, err := a.readLast()
	if err != nil {
		return 0, err
	}
	next := last
	if min > next {
		next = min
	}
	next++
	if err := a.writeLast(next); err != nil {
		return 0, err
	}
	return next, nil
}

// Last returns the persisted counter without allocating. It returns 0 when
// the state file does not exist yet. Because the state file is only ever
// replaced atomically, the lock is not needed to read a consistent value.
func (a *Allocator) Last() (int64, error) {
	return a.readLast()
}

func (a *Allocator) readLast() (int64, error) {
	data, err := os.ReadFile(a.stateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read id state: %w", err)
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return 0, fmt.Errorf("corrupt %s: %w (delete it to reseed from existing data)", filepath.Base(a.stateFile), err)
	}
	return s.Last, nil
}

func (a *Allocator) writeLast(last int64) error {
	data, err := json.MarshalIndent(state{Last: last}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.CreateTemp(filepath.Dir(a.stateFile), tempPattern)
	if err != nil {
		return fmt.Errorf("create id temp: %w", err)
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(stateFileMode); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return fmt.Errorf("write id state: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close id state: %w", closeErr)
	}
	return os.Rename(f.Name(), a.stateFile)
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("create id dir: %w", err)
	}
	return nil
}

// MaxNumeric returns the largest value in ids that parses as a decimal
// integer. Non-numeric entries, such as legacy hex ids, are ignored.
func MaxNumeric(ids []string) int64 {
	var max int64
	for _, id := range ids {
		if n, err := strconv.ParseInt(id, 10, 64); err == nil && n > max {
			max = n
		}
	}
	return max
}
