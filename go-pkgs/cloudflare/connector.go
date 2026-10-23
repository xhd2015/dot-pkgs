package cloudflare

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const connectorStopWait = 2 * time.Second

var (
	connectorNow   = time.Now
	connectorSleep = time.Sleep
	connectorAlive = processAlive
	connectorKill  = func(pid int, sig syscall.Signal) error {
		if pid == 0 {
			return nil
		}
		return syscall.Kill(pid, sig)
	}
	listConfigPIDs = defaultListConfigPIDs
)

// stopAndWaitPID SIGTERMs a process group, waits up to 2s, then SIGKILLs.
func stopAndWaitPID(pid int) {
	if pid <= 0 {
		return
	}
	_ = connectorKill(-pid, syscall.SIGTERM)
	_ = connectorKill(pid, syscall.SIGTERM)
	deadline := connectorNow().Add(connectorStopWait)
	for connectorAlive(pid) && connectorNow().Before(deadline) {
		connectorSleep(50 * time.Millisecond)
	}
	if !connectorAlive(pid) {
		return
	}
	_ = connectorKill(-pid, syscall.SIGKILL)
	_ = connectorKill(pid, syscall.SIGKILL)
	deadline = connectorNow().Add(200 * time.Millisecond)
	for connectorAlive(pid) && connectorNow().Before(deadline) {
		connectorSleep(20 * time.Millisecond)
	}
}

// sweepConfigConnectors stops every cloudflared using configPath except keepPID.
// keepPID 0 means stop all of them.
func sweepConfigConnectors(configPath string, keepPID int) {
	pids, err := listConfigPIDs(configPath)
	if err != nil {
		return
	}
	for _, pid := range pids {
		if pid <= 0 || pid == keepPID {
			continue
		}
		stopAndWaitPID(pid)
	}
}

func defaultListConfigPIDs(configPath string) ([]int, error) {
	want, err := filepath.Abs(configPath)
	if err != nil {
		want = configPath
	}
	out, err := exec.Command("pgrep", "-f", "cloudflared.*tunnel").Output()
	if err != nil && len(out) == 0 {
		return nil, nil
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, convErr := strconv.Atoi(line)
		if convErr != nil || pid <= 0 {
			continue
		}
		args, argsErr := readProcessArgs(pid)
		if argsErr != nil {
			continue
		}
		got := configFromArgs(args)
		if got == "" {
			continue
		}
		gotAbs, absErr := filepath.Abs(got)
		if absErr != nil {
			gotAbs = got
		}
		if gotAbs == want {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func configFromArgs(args []string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--config" {
			return args[i+1]
		}
	}
	return ""
}

func readProcessArgs(pid int) ([]string, error) {
	if data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline"); err == nil {
		parts := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
		var args []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				args = append(args, part)
			}
		}
		if len(args) > 0 {
			return args, nil
		}
	}
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-ww", "-o", "args=").Output()
	if err != nil {
		return nil, err
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return nil, os.ErrNotExist
	}
	return strings.Fields(line), nil
}
