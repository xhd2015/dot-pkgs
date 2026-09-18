package air

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCancellationAllowsGracefulCleanup(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-air")
	// A graceful TERM must execute the trap; SIGKILL cannot produce the marker.
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntrap 'echo stopped > stopped; exit 0' TERM\necho ready > ready\nwhile :; do sleep 0.1; done\n"), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, err := Start(ctx, Options{Dir: dir, Bin: script, BuildCmd: "unused", Entrypoint: "unused", Stdout: io.Discard, Stderr: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(dir, "ready")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child not ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case <-p.Exited:
	case <-time.After(3 * time.Second):
		t.Fatal("cancellation did not stop Air")
	}
	if _, err := os.Stat(filepath.Join(dir, "stopped")); err != nil {
		t.Fatalf("child was not gracefully stopped: %v", err)
	}
}
