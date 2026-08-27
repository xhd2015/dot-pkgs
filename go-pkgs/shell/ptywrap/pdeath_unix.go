//go:build unix

package ptywrap

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// watchScript polls the manager pid and the PTY child. When the manager
// disappears (including SIGKILL), it SIGKILLs the child's process group so
// hangup-immune grandchildren cannot keep a slave FD. $1=ppid $2=child.
const pdeathWatchScript = `
ppid="$1"
child="$2"
while kill -0 "$ppid" 2>/dev/null; do
	if ! kill -0 "$child" 2>/dev/null; then
		exit 0
	fi
	sleep 0.1
done
kill -KILL -"$child" 2>/dev/null || true
kill -KILL "$child" 2>/dev/null || true
exit 0
`

func preparePTYCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Do not Setpgid here: creack/pty StartWithSize already Setsid (new
	// session + process group). Setpgid+Setsid together yields EPERM.
	applyPdeathsig(cmd.SysProcAttr)
}

func startPdeathWatcher(ppid, child int) {
	if ppid <= 1 || child <= 0 {
		return
	}
	w := exec.Command("/bin/sh", "-c", pdeathWatchScript, "ptywrap-pdeath", strconv.Itoa(ppid), strconv.Itoa(child))
	w.Stdin = nil
	w.Stdout = nil
	w.Stderr = nil
	w.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := w.Start(); err != nil {
		return
	}
	go func() { _ = w.Wait() }()
}

func killProcessGroup(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Kill()
	}
}
