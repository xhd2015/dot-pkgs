package cloudflare

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	restartBackoffMin = 2 * time.Second
	restartBackoffCap = 30 * time.Second
)

// Process wraps a started cloudflared tunnel process (process group).
// If cloudflared exits on its own, including after an auto-update, the
// supervisor starts it again. Stop is the only path that does not respawn.
type Process struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	configPath string
	tunnelName string
	logPath    string
	stopped    bool
	attempt    int
	startedAt  time.Time
	pid        atomic.Int64
	done       chan struct{}
	sleepFor   func(time.Duration)
	onPID      func(int)
}

// DefaultTunnelProtocol pins the cloudflared edge transport for HTTP origins.
// The "auto" default tries QUIC first, which is blocked or degraded on many
// corporate networks: QUIC dial timeouts delay connector startup for minutes
// and long-lived h2 fallback connections still decay on lossy paths. http2
// matches the ai-critic unified_tunnel default.
const DefaultTunnelProtocol = "http2"

// tunnelRunArgs is the canonical `tunnel run` argv. All connector spawn sites
// (process registry, attach, restart-after-ingress-change) must go through it
// so every connector gets the pinned protocol.
func tunnelRunArgs(configPath, tunnelName string) []string {
	return []string{"tunnel", "--config", configPath, "--protocol", DefaultTunnelProtocol, "run", tunnelName}
}

// tunnelRunArgsAlt is the fallback argv shape used when the primary form is
// rejected (older cloudflared builds expect `run` before `--config`).
func tunnelRunArgsAlt(configPath, tunnelName string) []string {
	return []string{"tunnel", "run", "--config", configPath, "--protocol", DefaultTunnelProtocol, tunnelName}
}

// startCommand is replaced in tests.
var startCommand = func(configPath, tunnelName string) *exec.Cmd {
	return exec.Command("cloudflared", tunnelRunArgs(configPath, tunnelName)...)
}

// StartProcess launches `cloudflared tunnel --config <config> run <tunnel>`
// in its own process group so Stop can kill the whole group.
func StartProcess(configPath, tunnelName string, log *os.File) (*Process, error) {
	p := &Process{
		configPath: configPath,
		tunnelName: tunnelName,
		done:       make(chan struct{}),
	}
	if log != nil {
		p.logPath = log.Name()
	}
	if err := p.spawn(log); err != nil {
		return nil, err
	}
	go p.supervise()
	return p, nil
}

func (p *Process) spawn(log *os.File) error {
	cmd := startCommand(p.configPath, p.tunnelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if log != nil {
		cmd.Stdout = log
		cmd.Stderr = log
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start cloudflared: %w", err)
	}
	p.cmd = cmd
	p.startedAt = time.Now()
	if cmd.Process != nil {
		p.pid.Store(int64(cmd.Process.Pid))
		if p.onPID != nil {
			p.onPID(cmd.Process.Pid)
		}
	}
	return nil
}

func (p *Process) supervise() {
	defer close(p.done)
	for {
		p.mu.Lock()
		cmd := p.cmd
		p.mu.Unlock()
		if cmd == nil {
			return
		}
		_ = cmd.Wait()

		p.mu.Lock()
		if p.stopped || p.cmd != cmd {
			p.mu.Unlock()
			return
		}
		p.cmd = nil
		attempt := p.attempt
		if !p.startedAt.IsZero() && time.Since(p.startedAt) >= restartBackoffCap {
			attempt = 0
		}
		p.attempt = attempt + 1
		p.mu.Unlock()

		p.sleep(restartDelay(attempt))

		p.mu.Lock()
		if p.stopped {
			p.mu.Unlock()
			return
		}
		log, err := p.openLog()
		if err != nil {
			p.mu.Unlock()
			continue
		}
		err = p.spawn(log)
		if log != nil {
			_ = log.Close()
		}
		p.mu.Unlock()
		if err != nil {
			continue
		}
	}
}

func (p *Process) openLog() (*os.File, error) {
	if p.logPath == "" {
		return nil, nil
	}
	return os.OpenFile(p.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

func (p *Process) sleep(delay time.Duration) {
	if p.sleepFor != nil {
		p.sleepFor(delay)
		return
	}
	time.Sleep(delay)
}

func restartDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := restartBackoffMin
	for i := 0; i < attempt; i++ {
		if delay >= restartBackoffCap/2 {
			return restartBackoffCap
		}
		delay *= 2
	}
	if delay > restartBackoffCap {
		return restartBackoffCap
	}
	return delay
}

// Stop kills the process group and waits. It does not respawn.
func (p *Process) Stop() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	p.stopped = true
	pid := int(p.pid.Load())
	p.mu.Unlock()
	if pid > 0 {
		_ = syscall.Kill(-pid, syscall.SIGTERM)
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	if p.done != nil {
		select {
		case <-p.done:
		case <-time.After(5 * time.Second):
			if pid > 0 {
				_ = syscall.Kill(-pid, syscall.SIGKILL)
			}
			<-p.done
		}
	}
	return nil
}

// PID returns the process id, or 0 if unknown.
func (p *Process) PID() int {
	if p == nil {
		return 0
	}
	return int(p.pid.Load())
}

// SetOnPID reports each started pid, including restarts.
func (p *Process) SetOnPID(fn func(int)) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.onPID = fn
	pid := p.PID()
	p.mu.Unlock()
	if fn != nil && pid > 0 {
		fn(pid)
	}
}

func lookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if err := syscall.Kill(pid, 0); err != nil {
		return false
	}
	return !processZombie(pid)
}

func processZombie(pid int) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	return statIsZombie(b)
}

func statIsZombie(stat []byte) bool {
	end := bytes.LastIndex(stat, []byte(")"))
	if end < 0 || end+1 >= len(stat) {
		return false
	}
	rest := bytes.TrimSpace(stat[end+1:])
	return len(rest) > 0 && rest[0] == 'Z'
}
