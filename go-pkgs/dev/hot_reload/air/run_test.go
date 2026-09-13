package air

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestStart_Stop(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-air.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntrap 'exit 0' TERM\nwhile true; do sleep 0.1; done\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	var stopped bool
	p, err := Start(context.Background(), Options{
		Dir:        dir,
		BuildCmd:   "go build -o ./tmp/x ./cmd/x",
		Entrypoint: "./tmp/x",
		ArgsBin:    []string{"--dev"},
		Bin:        script,
		SkipEnsure: true,
		Stderr:     &stderr,
		Command: func(ctx context.Context, bin string, args []string, dir string, stdout, stderr io.Writer) *exec.Cmd {
			c := exec.CommandContext(ctx, bin)
			c.Dir = dir
			c.Stdout = stdout
			c.Stderr = stderr
			c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			return c
		},
		StopTimeout: 2 * time.Second,
		OnStop: func() {
			stopped = true
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	p.Stop()
	if !stopped {
		t.Fatal("OnStop not called")
	}
	// Stop already waited on Wait(); do not receive from Exited again.
	p.Stop() // idempotent
}
