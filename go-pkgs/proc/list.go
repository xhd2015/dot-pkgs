package proc

import (
	"os/exec"
	"strconv"
)

// List returns a process snapshot. When opts.List != nil, the inject is used
// as-is (no re-parse). Otherwise live `ps -ax -o pid=,ppid=,command=` is run
// (fallback `ps -x …`). On live failure returns empty (not panic).
func List(opts Options) []Proc {
	if opts.List != nil {
		return opts.List()
	}
	return listLive()
}

func listLive() []Proc {
	// macOS/BSD and Linux: pid, ppid, full command line.
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=,command=")
	out, err := cmd.Output()
	if err != nil {
		// Try without -a (some environments).
		cmd = exec.Command("ps", "-x", "-o", "pid=,ppid=,command=")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
	}
	return ParsePSOutput(out)
}

// OpenFiles returns open-file paths for pid. When opts.OpenFiles != nil, the
// inject result is returned as-is. Otherwise live `lsof -p <pid> -Fn` is run.
// On failure returns empty.
func OpenFiles(pid int, opts Options) []string {
	if opts.OpenFiles != nil {
		return opts.OpenFiles(pid)
	}
	return openFilesLive(pid)
}

func openFilesLive(pid int) []string {
	if pid <= 0 {
		return nil
	}
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-Fn")
	out, err := cmd.Output()
	if err != nil {
		// lsof exits non-zero when no files / no process; still may have stdout.
		if len(out) == 0 {
			return nil
		}
	}
	return ParseLsofFn(out)
}
