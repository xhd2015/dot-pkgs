package proc

import "syscall"

// Alive reports whether pid is live. pid <= 0 always returns false (before
// inject/live). When opts.Alive != nil and pid > 0, the inject is used;
// otherwise a Unix signal-0 probe is performed.
func Alive(pid int, opts Options) bool {
	if pid <= 0 {
		return false
	}
	if opts.Alive != nil {
		return opts.Alive(pid)
	}
	return aliveLive(pid)
}

func aliveLive(pid int) bool {
	// Signal 0 checks existence / permission without delivering a real signal.
	return syscall.Kill(pid, 0) == nil
}
