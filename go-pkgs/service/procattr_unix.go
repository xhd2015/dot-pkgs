//go:build unix

package service

import "syscall"

// procAttr puts the child in its own process group (Setpgid) so stop can signal
// the whole tree, or in a new session (Setsid) when the child must outlive the
// host, which also detaches it from the controlling terminal.
func procAttr(detached bool) *syscall.SysProcAttr {
	if detached {
		return &syscall.SysProcAttr{Setsid: true}
	}
	return &syscall.SysProcAttr{Setpgid: true}
}
