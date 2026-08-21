//go:build windows

package lookpath

import "syscall"

// loginSysProcAttr: Windows has no setsid; login probes do not use Unix job control.
func loginSysProcAttr() *syscall.SysProcAttr {
	return nil
}
