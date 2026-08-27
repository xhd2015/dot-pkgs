//go:build unix && !linux

package ptywrap

import "syscall"

func applyPdeathsig(attr *syscall.SysProcAttr) {}
