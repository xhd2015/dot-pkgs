package cloudflare

import (
	"os/exec"
	"testing"
	"time"
)

func TestStatIsZombie(t *testing.T) {
	if !statIsZombie([]byte("1261 (cloudflared) Z 737 1 1 0 -1 0")) {
		t.Fatal("zombie stat should be dead")
	}
	if statIsZombie([]byte("1261 (cloudflared) S 737 1 1 0 -1 0")) {
		t.Fatal("sleeping stat should be alive")
	}
	if statIsZombie([]byte("12 (name) with) paren) R 1")) {
		t.Fatal("state is the field after the last paren")
	}
}

func TestRestartDelay(t *testing.T) {
	if restartDelay(0) != 2*time.Second || restartDelay(1) != 4*time.Second || restartDelay(2) != 8*time.Second || restartDelay(4) != 30*time.Second {
		t.Fatalf("delays 0=%s 1=%s 2=%s 4=%s", restartDelay(0), restartDelay(1), restartDelay(2), restartDelay(4))
	}
}

func TestSuperviseRestartsUntilStopped(t *testing.T) {
	var starts int
	prev := startCommand
	startCommand = func(configPath, tunnelName string) *exec.Cmd {
		starts++
		return exec.Command("sh", "-c", "exit 0")
	}
	t.Cleanup(func() { startCommand = prev })

	p, err := StartProcess("cfg", "tunnel", nil)
	if err != nil {
		t.Fatal(err)
	}
	p.sleepFor = func(time.Duration) {}
	deadline := time.Now().Add(5 * time.Second)
	for starts < 3 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if starts < 2 {
		t.Fatalf("starts = %d, want a restart after exit", starts)
	}
	got := starts
	time.Sleep(50 * time.Millisecond)
	if starts != got {
		t.Fatalf("starts grew after Stop: %d -> %d", got, starts)
	}
}
