package cloudflare

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

var (
	processRegistryMu sync.Mutex
	processRegistry   = map[string]*Process{}
)

func registryKey(configPath string) string {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return configPath
	}
	return abs
}

func stopRegisteredProcess(configPath string) {
	key := registryKey(configPath)
	processRegistryMu.Lock()
	p := processRegistry[key]
	delete(processRegistry, key)
	processRegistryMu.Unlock()
	if p != nil {
		_ = p.Stop()
	}
}

func registerProcess(configPath string, p *Process) {
	if p == nil {
		return
	}
	key := registryKey(configPath)
	processRegistryMu.Lock()
	old := processRegistry[key]
	processRegistry[key] = p
	processRegistryMu.Unlock()
	if old != nil && old != p {
		_ = old.Stop()
	}
}

func registeredProcess(configPath string) *Process {
	processRegistryMu.Lock()
	defer processRegistryMu.Unlock()
	return processRegistry[registryKey(configPath)]
}

// startRegisteredConnector Stop()s any supervisor for this config, sweeps OS
// leftovers, then StartProcess and records the Process so a discarded Session
// cannot leave a respawning goroutine behind.
func startRegisteredConnector(st *TunnelState, configPath, name, managedDir string) (*Process, error) {
	stopRegisteredProcess(configPath)
	if st != nil && st.ConnectorPID > 0 {
		stopAndWaitPID(st.ConnectorPID)
		st.ConnectorPID = 0
	}
	sweepConfigConnectors(configPath, 0)

	logPath := filepath.Join(managedDir, "cloudflared.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	proc, err := StartProcess(configPath, name, logFile)
	_ = logFile.Close()
	if err != nil {
		return nil, err
	}
	registerProcess(configPath, proc)
	proc.SetOnPID(func(pid int) {
		tryPersistConnectorPID(managedDir, pid)
	})
	if st != nil {
		st.ConnectorPID = proc.PID()
	}
	if err := waitConnectorAlive(proc.PID(), logPath); err != nil {
		stopRegisteredProcess(configPath)
		return nil, err
	}
	return proc, nil
}

func tryPersistConnectorPID(managedDir string, pid int) {
	if pid <= 0 || managedDir == "" {
		return
	}
	lockPath := filepath.Join(managedDir, "lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		// Attach still holds the managed-dir flock and will SaveTunnelState.
		return
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	st, err := LoadTunnelState(managedDir)
	if err != nil || st == nil {
		return
	}
	if st.ConnectorPID == pid {
		return
	}
	st.ConnectorPID = pid
	_ = SaveTunnelState(managedDir, st)
}
