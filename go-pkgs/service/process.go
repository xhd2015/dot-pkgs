package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ErrUnsupported reports a process primitive the current platform cannot
// provide. Every caller treats it as "cannot manage this service".
var ErrUnsupported = errors.New("service management is not supported on this platform")

// ParseProcStatState extracts the state field from a /proc/<pid>/stat line.
// Format: pid (comm) state ... — comm may contain spaces and parentheses, so
// the split point is the last ") ".
func ParseProcStatState(stat []byte) (state byte, ok bool) {
	s := string(stat)
	i := strings.LastIndex(s, ") ")
	if i < 0 || i+2 >= len(s) {
		return 0, false
	}
	return s[i+2], true
}

// CommandMatches reports whether the live command line still belongs to the
// recorded one. Used before reclaiming a pid from a previous run: a recycled
// pid must never be killed. The recorded command may be a fragment (the first
// words of the shell line), so this is a containment test.
func CommandMatches(recorded, live string) bool {
	recorded = strings.TrimSpace(recorded)
	live = strings.TrimSpace(live)
	if recorded == "" || live == "" {
		return false
	}
	return strings.Contains(live, recorded)
}

// ProcessCommandLine returns the command line of pid, or "" when it is gone or
// cannot be read.
//
// Linux reads /proc/<pid>/cmdline; elsewhere it shells out to ps, whose output
// includes the interpreter and arguments.
func ProcessCommandLine(pid int) string {
	if pid <= 0 {
		return ""
	}
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		if len(data) == 0 {
			return ""
		}
		parts := strings.Split(string(data), "\x00")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return strings.TrimSpace(strings.Join(parts, " "))
	}
	out, err := exec.Command("ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ListeningPIDs returns the pids listening on TCP port.
//
// It requires lsof, the only portable tool that reports the listener (not every
// socket) on both macOS and Linux. A port with no listener returns nil, nil.
func ListeningPIDs(port int) ([]int, error) {
	if port <= 0 {
		return nil, nil
	}
	out, err := exec.Command("lsof", "-tiTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-n", "-P").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) == 0 {
			// lsof exits 1 with no stderr when nothing matches.
			return nil, nil
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			return nil, fmt.Errorf("lsof not installed: required for service port management")
		}
		return nil, err
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, convErr := strconv.Atoi(line)
		if convErr == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

// KillPortOwner terminates whatever listens on port, so a restart can bind it
// again. A port with no listener is a no-op.
//
// This is the last resort before starting a service whose previous instance
// leaked: it is scoped to the listener and is never called for a port the user
// did not configure for this service.
func KillPortOwner(port int) error {
	pids, err := ListeningPIDs(port)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		if err := StopProcess(pid); err != nil {
			return err
		}
	}
	return nil
}
