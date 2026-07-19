package iterm2

import (
	"bytes"
	"os/exec"
	"strings"
)

// foregroundCommForTTY returns the command name of the foreground process on
// tty (e.g. "/dev/ttys001") using `ps`. ok is false when the probe fails.
//
// On macOS/BSD, STAT contains '+' for the process group associated with the
// terminal (foreground). When no '+' is found, the last non-empty comm on that
// TTY is used as a weak fallback.
func foregroundCommForTTY(tty string) (comm string, ok bool) {
	tty = strings.TrimSpace(tty)
	if tty == "" {
		return "", false
	}
	short := strings.TrimPrefix(tty, "/dev/")
	if short == "" {
		return "", false
	}

	cmd := exec.Command("ps", "-t", short, "-o", "stat=,comm=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", false
	}

	var last string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: STAT may be like "Ss+" or "S+" then whitespace then COMM.
		// Split on first run of spaces after the stat token.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		stat := fields[0]
		name := fields[len(fields)-1]
		if name == "" {
			continue
		}
		last = name
		if strings.Contains(stat, "+") {
			return name, true
		}
	}
	if last != "" {
		return last, true
	}
	return "", false
}
