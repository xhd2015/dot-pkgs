//go:build windows

package service

import "time"

// Windows process control is not implemented. Service management on Windows is
// out of scope: the eluc host is POSIX-only. The stubs keep the package
// compiling so hosts that also target Windows still build.

const (
	// DefaultTermTimeout is unused on Windows.
	DefaultTermTimeout = 3 * time.Second
	// DefaultKillTimeout is unused on Windows.
	DefaultKillTimeout = 2 * time.Second
)

// Alive always reports false on Windows.
func Alive(pid int) bool { return false }

// Zombie always reports false on Windows.
func Zombie(pid int) bool { return false }

// ProcessState returns an empty state on Windows.
func ProcessState(pid int) (byte, bool) { return 0, false }

// GroupID returns the pid on Windows.
func GroupID(pid int) (int, error) { return pid, nil }

// StopProcess returns ErrUnsupported on Windows.
func StopProcess(pid int) error { return ErrUnsupported }

// StopProcessWithTimeout returns ErrUnsupported on Windows.
func StopProcessWithTimeout(pid int, termTimeout, killTimeout time.Duration) error {
	return ErrUnsupported
}

// StopProcessGroup returns ErrUnsupported on Windows.
func StopProcessGroup(pid int) error { return ErrUnsupported }

// StopProcessGroupWithTimeout returns ErrUnsupported on Windows.
func StopProcessGroupWithTimeout(pid int, termTimeout, killTimeout time.Duration) error {
	return ErrUnsupported
}

// KillProcessGroupNow returns ErrUnsupported on Windows.
func KillProcessGroupNow(pid int) error { return ErrUnsupported }

// WaitForExit returns false on Windows.
func WaitForExit(pid int, timeout time.Duration) bool { return false }
