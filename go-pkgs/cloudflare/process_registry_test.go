package cloudflare

import (
	"os/exec"
	"testing"
	"time"
)

func TestRegisterProcessStopsPreviousSupervisor(t *testing.T) {
	prev := startCommand
	var starts int
	startCommand = func(configPath, tunnelName string) *exec.Cmd {
		starts++
		return exec.Command("sh", "-c", "sleep 60")
	}
	t.Cleanup(func() {
		startCommand = prev
		stopRegisteredProcess("cfg-registry-test")
	})

	p1, err := StartProcess("cfg-registry-test", "tunnel", nil)
	if err != nil {
		t.Fatal(err)
	}
	registerProcess("cfg-registry-test", p1)

	p2, err := StartProcess("cfg-registry-test", "tunnel", nil)
	if err != nil {
		t.Fatal(err)
	}
	registerProcess("cfg-registry-test", p2)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		p1.mu.Lock()
		stopped := p1.stopped
		p1.mu.Unlock()
		if stopped {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	p1.mu.Lock()
	stopped := p1.stopped
	p1.mu.Unlock()
	if !stopped {
		t.Fatal("first Process was not Stop()'d; supervisor would respawn")
	}
	if registeredProcess("cfg-registry-test") != p2 {
		t.Fatal("registry should hold the new Process")
	}
	got := starts
	time.Sleep(80 * time.Millisecond)
	if starts != got {
		t.Fatalf("starts grew after replace: %d -> %d (old supervisor respawned)", got, starts)
	}
}
