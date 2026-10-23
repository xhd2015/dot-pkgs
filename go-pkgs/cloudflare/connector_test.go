package cloudflare

import (
	"syscall"
	"testing"
	"time"
)

func TestStopAndWaitPIDSIGKILLsAfterTimeout(t *testing.T) {
	prevNow := connectorNow
	prevSleep := connectorSleep
	prevAlive := connectorAlive
	prevKill := connectorKill
	t.Cleanup(func() {
		connectorNow = prevNow
		connectorSleep = prevSleep
		connectorAlive = prevAlive
		connectorKill = prevKill
	})

	now := time.Date(2026, 10, 23, 16, 0, 0, 0, time.UTC)
	connectorNow = func() time.Time { return now }
	connectorSleep = func(d time.Duration) { now = now.Add(d) }
	connectorAlive = func(pid int) bool { return pid == 8765 }
	var sigs []syscall.Signal
	connectorKill = func(pid int, sig syscall.Signal) error {
		if pid == 8765 {
			sigs = append(sigs, sig)
		}
		return nil
	}

	stopAndWaitPID(8765)
	if len(sigs) < 2 {
		t.Fatalf("sigs = %v, want TERM then KILL", sigs)
	}
	if sigs[0] != syscall.SIGTERM {
		t.Fatalf("first = %v, want SIGTERM", sigs[0])
	}
	var sawKill bool
	for _, s := range sigs {
		if s == syscall.SIGKILL {
			sawKill = true
		}
	}
	if !sawKill {
		t.Fatalf("sigs = %v, want SIGKILL after 2s", sigs)
	}
}

func TestStopAndWaitPIDReturnsAfterDeath(t *testing.T) {
	prevNow := connectorNow
	prevSleep := connectorSleep
	prevAlive := connectorAlive
	prevKill := connectorKill
	t.Cleanup(func() {
		connectorNow = prevNow
		connectorSleep = prevSleep
		connectorAlive = prevAlive
		connectorKill = prevKill
	})

	now := time.Date(2026, 10, 23, 16, 0, 0, 0, time.UTC)
	connectorNow = func() time.Time { return now }
	connectorSleep = func(d time.Duration) { now = now.Add(d) }
	alive := true
	connectorAlive = func(pid int) bool { return pid == 8765 && alive }
	connectorKill = func(pid int, sig syscall.Signal) error {
		if pid == 8765 && sig == syscall.SIGTERM {
			alive = false
		}
		return nil
	}

	stopAndWaitPID(8765)
	if alive {
		t.Fatal("expected process marked dead after SIGTERM")
	}
}

func TestSweepConfigConnectorsSkipsKeepPID(t *testing.T) {
	prevList := listConfigPIDs
	prevAlive := connectorAlive
	prevKill := connectorKill
	prevSleep := connectorSleep
	t.Cleanup(func() {
		listConfigPIDs = prevList
		connectorAlive = prevAlive
		connectorKill = prevKill
		connectorSleep = prevSleep
	})

	listConfigPIDs = func(string) ([]int, error) { return []int{8765, 82434}, nil }
	connectorAlive = func(int) bool { return false }
	connectorSleep = func(time.Duration) {}
	var stopped []int
	connectorKill = func(pid int, sig syscall.Signal) error {
		if pid > 0 && sig == syscall.SIGTERM {
			stopped = append(stopped, pid)
		}
		return nil
	}

	sweepConfigConnectors("/tmp/config.yml", 82434)
	if len(stopped) != 1 || stopped[0] != 8765 {
		t.Fatalf("stopped = %v, want [8765]", stopped)
	}
}

func TestConfigFromArgs(t *testing.T) {
	got := configFromArgs([]string{
		"cloudflared", "tunnel", "--config", "/tmp/managed/config.yml", "run", "ai-critic-vq35xx0d03",
	})
	if got != "/tmp/managed/config.yml" {
		t.Fatalf("got %q", got)
	}
}
