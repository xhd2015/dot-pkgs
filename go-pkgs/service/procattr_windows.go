//go:build windows

package service

import "syscall"

// procAttr returns nil: Windows process separation is out of scope, so the
// child inherits the host's console and group.
func procAttr(detached bool) *syscall.SysProcAttr { return nil }
