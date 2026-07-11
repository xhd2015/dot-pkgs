package cloudflare

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Process wraps a started cloudflared tunnel process (process group).
type Process struct {
	cmd *exec.Cmd
}

// StartProcess launches `cloudflared tunnel --config <config> run <tunnel>`
// in its own process group so Stop can kill the whole group.
func StartProcess(configPath, tunnelName string, log *os.File) (*Process, error) {
	cmd := exec.Command("cloudflared", "tunnel", "--config", configPath, "run", tunnelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if log != nil {
		cmd.Stdout = log
		cmd.Stderr = log
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cloudflared: %w", err)
	}
	return &Process{cmd: cmd}, nil
}

// Stop kills the process group and waits.
func (p *Process) Stop() error {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	pgid := p.cmd.Process.Pid
	// Negative PID kills the process group on Unix.
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	_ = p.cmd.Wait()
	return nil
}

// PID returns the process id, or 0 if unknown.
func (p *Process) PID() int {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func lookPath(file string) (string, error) {
	return exec.LookPath(file)
}
