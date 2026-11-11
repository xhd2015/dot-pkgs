//go:build unix

package service

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Default stop budgets. A service gets 3s to exit on SIGTERM, then 2s after
// SIGKILL before StopProcess reports failure.
const (
	DefaultTermTimeout = 3 * time.Second
	DefaultKillTimeout = 2 * time.Second
)

// Alive reports whether pid exists and is running.
//
// A zombie pid still succeeds kill(pid, 0) on Linux until its parent reaps it;
// treating zombies as alive makes stop and restart fail with "process N did
// not exit", so they are reported dead here.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if Zombie(pid) {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// Zombie reports whether pid exists as a zombie. Only Linux exposes the state
// cheaply; elsewhere it reports false and Alive falls back to kill(pid, 0).
func Zombie(pid int) bool {
	state, ok := ProcessState(pid)
	return ok && state == 'Z'
}

// ProcessState reads the state character from /proc/<pid>/stat ("R", "S", "Z", …).
// ok is false when the file is unreadable, which includes every non-Linux host.
func ProcessState(pid int) (state byte, ok bool) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, false
	}
	return ParseProcStatState(data)
}

// GroupID returns the process group id of pid.
func GroupID(pid int) (int, error) {
	if pid <= 0 {
		return 0, fmt.Errorf("invalid pid %d", pid)
	}
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return 0, err
	}
	return pgid, nil
}

// StopProcess terminates pid and, when pid leads its own process group, every
// descendant in that group.
//
// Services are started with Setpgid (or Setsid when detached), so the common
// path signals the group: a shell that spawned a server leaves the server
// behind otherwise. When pid is not a group leader the signal goes to pid
// alone, because -pid would target whatever group happens to share that number
// — including the host's own group.
func StopProcess(pid int) error {
	return StopProcessWithTimeout(pid, DefaultTermTimeout, DefaultKillTimeout)
}

// StopProcessWithTimeout is StopProcess with explicit budgets.
func StopProcessWithTimeout(pid int, termTimeout, killTimeout time.Duration) error {
	if pid <= 0 {
		return nil
	}
	pgid, err := GroupID(pid)
	group := err == nil && pgid == pid
	return signalUntilExit(pid, pgid, group, termTimeout, killTimeout)
}

// KillProcessGroupNow sends SIGKILL without waiting for a graceful exit, then
// waits up to two seconds for the pid to disappear. Used to reclaim an orphan
// left by a previous host instance, where nothing is left to negotiate with.
func KillProcessGroupNow(pid int) {
	if pid <= 0 {
		return
	}
	pgid, err := GroupID(pid)
	if err == nil && pgid == pid {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	} else {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	WaitForExit(pid, DefaultKillTimeout)
}

// WaitForExit polls pid until it disappears or timeout elapses.
func WaitForExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !Alive(pid) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return !Alive(pid)
}

// signalUntilExit sends SIGTERM, waits termTimeout, then SIGKILL, waits
// killTimeout. group selects whether the signal targets -pgid or pid.
func signalUntilExit(pid, pgid int, group bool, termTimeout, killTimeout time.Duration) error {
	send := func(sig syscall.Signal) {
		if group && pgid > 0 {
			_ = syscall.Kill(-pgid, sig)
			if sig == syscall.SIGKILL {
				// A process that left the group still holds the port.
				_ = syscall.Kill(pid, sig)
			}
			return
		}
		_ = syscall.Kill(pid, sig)
	}
	send(syscall.SIGTERM)
	if WaitForExit(pid, termTimeout) {
		return nil
	}
	send(syscall.SIGKILL)
	if WaitForExit(pid, killTimeout) {
		return nil
	}
	return fmt.Errorf("process %d did not exit", pid)
}
