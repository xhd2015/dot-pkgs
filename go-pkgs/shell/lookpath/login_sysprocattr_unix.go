//go:build unix

package lookpath

import "syscall"

// loginSysProcAttr starts the login probe in a new session with no controlling
// TTY. Interactive shells (-i) otherwise inherit the caller's controlling
// terminal and try tcsetpgrp → SIGTTOU → Stopped under job control.
func loginSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
