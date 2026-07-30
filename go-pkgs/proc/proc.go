// Package proc provides shared process primitives: parse/list processes,
// parse/list open files, build children index and BFS descendants, and probe
// Alive — all testable with pure parsers and injectable Options.
package proc

// Proc is a single process table row.
type Proc struct {
	PID  int
	PPID int
	Cmd  string
}

// Options holds injectable hooks for List, OpenFiles, and Alive.
// When a hook is nil, the corresponding function uses a live system probe.
// Parallel-safe: no package-level mutable state.
type Options struct {
	List      func() []Proc
	OpenFiles func(pid int) []string
	Alive     func(pid int) bool
}
