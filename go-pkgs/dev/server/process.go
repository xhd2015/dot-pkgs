package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type child struct {
	cmd  *exec.Cmd
	done chan struct{}
	err  error
	once sync.Once
}

func startChild(dir string, argv, env []string, cfg Config) (*child, error) {
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty child command")
	}
	c := exec.Command(argv[0], argv[1:]...)
	c.Dir, c.Stdout, c.Stderr = dir, cfg.Stdout, cfg.Stderr
	c.Env = append(os.Environ(), env...)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := c.Start(); err != nil {
		return nil, err
	}
	p := &child{cmd: c, done: make(chan struct{})}
	go func() { p.err = c.Wait(); close(p.done) }()
	return p, nil
}

func (p *child) stop() {
	p.once.Do(func() {
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-p.done:
		case <-time.After(3 * time.Second):
		}
		// A wrapper can exit before its descendants. Clean the owned group too.
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-p.done:
		case <-time.After(2 * time.Second):
		}
	})
}

func runCommand(ctx context.Context, dir string, argv []string, cfg Config) error {
	p, err := startChild(dir, argv, nil, cfg)
	if err != nil {
		return err
	}
	defer p.stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return p.err
	}
}

func waitReady(ctx context.Context, address, id string) error {
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, "GET", address+readyPath, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			if readErr == nil && resp.StatusCode == 200 && string(body) == id {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func stopOwnedBackend(address, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Do not send a shutdown request to any other invocation at this address.
	if err := waitReady(ctx, address, id); err != nil {
		return
	}
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, _ := http.NewRequestWithContext(ctx, "POST", address+"/__dev/stop", nil)
	req.Header.Set("X-Dev-Run-ID", id)
	if resp, err := client.Do(req); err == nil {
		resp.Body.Close()
	}
}
