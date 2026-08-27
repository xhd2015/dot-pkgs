//go:build windows

package ptywrap

import (
	"os"
	"os/exec"
)

func preparePTYCmd(cmd *exec.Cmd) {}

func startPdeathWatcher(ppid, child int) {}

func killProcessGroup(pid int) {
	if pid <= 0 {
		return
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Kill()
	}
}
