package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/dev/hot_reload/air"
	"github.com/xhd2015/dot-pkgs/go-pkgs/net/port"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/open"
)

func selectPort(requested, first, reserved int, cfg Config, label string) (int, error) {
	start, count := requested, 1
	if requested == 0 {
		start, count = first, 100
	}
	for p := start; p < start+count && p <= 65535; p++ {
		if p == reserved {
			continue
		}
		if _, err := port.FindAvailablePort(p, 1); err == nil {
			if p != start {
				message(cfg.Stderr, "33", "warning: %s port %d is occupied; using %d\n", label, start, p)
			}
			return p, nil
		}
	}
	return 0, fmt.Errorf("requested %s port %d is unavailable", label, start)
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }

func supervise(ctx context.Context, cfg Config, opts options) error {
	var err error
	opts.port, err = selectPort(opts.port, cfg.BackendPort, opts.vitePort, cfg, "backend")
	if err != nil {
		return err
	}
	if cfg.Frontend != nil {
		opts.vitePort, err = selectPort(opts.vitePort, cfg.FrontendPort, opts.port, cfg, "frontend")
		if err != nil {
			return err
		}
	}
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	id := hex.EncodeToString(bytes)
	if err := os.MkdirAll(filepath.Join(cfg.Root, "tmp"), 0755); err != nil {
		return err
	}
	runDir, err := os.MkdirTemp(filepath.Join(cfg.Root, "tmp"), "dev-")
	if err != nil {
		return err
	}
	// Only this invocation writes here; cleanup follows all process stops.
	defer os.RemoveAll(runDir)
	var frontendDone <-chan struct{}
	var frontend *child
	if cfg.Frontend != nil {
		v := cfg.Frontend
		if _, err := os.Stat(filepath.Join(cfg.Root, v.Dir, "node_modules")); os.IsNotExist(err) && len(v.Install) > 0 {
			if err := runCommand(ctx, filepath.Join(cfg.Root, v.Dir), v.Install, cfg); err != nil {
				return fmt.Errorf("frontend install: %w", err)
			}
		}
		message(cfg.Stdout, "90", "Starting frontend on %d...\n", opts.vitePort)
		frontend, err = startVite(cfg, runDir, id, opts.vitePort)
		if err != nil {
			return fmt.Errorf("frontend start: %w", err)
		}
		defer frontend.stop()
		frontendDone = frontend.done
		readyCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
		ready := make(chan error, 1)
		go func() { ready <- waitReady(readyCtx, fmt.Sprintf("http://127.0.0.1:%d", opts.vitePort), id) }()
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-frontendDone:
			err = fmt.Errorf("frontend exited: %v", frontend.err)
		case err = <-ready:
		}
		cancel()
		if err != nil {
			return fmt.Errorf("frontend readiness: %w", err)
		}
	}
	appURL := fmt.Sprintf("http://127.0.0.1:%d", opts.port)
	backendDone := make(chan error, 1)
	backendExited := make(chan struct{})
	stopBackend := func() {}
	if opts.noAir {
		backendCtx, cancel := context.WithCancel(ctx)
		go func() { backendDone <- serve(backendCtx, cfg, opts, id) }()
		stopBackend = func() { cancel(); <-backendExited }
	} else {
		message(cfg.Stdout, "90", "Building backend on %d...\n", opts.port)
		bin := filepath.Join(runDir, "backend")
		excluded := []string{"tmp", ".git", "node_modules", "external", "vendor"}
		if cfg.Frontend != nil {
			excluded = append(excluded, filepath.ToSlash(cfg.Frontend.Dir))
		}
		airBin, err := air.Ensure(ctx, air.EnsureOpts{Stdout: cfg.Stdout, Stderr: cfg.Stderr})
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		// The supervisor first asks its identified backend to clean up, then
		// stops Air. Keep signal cancellation from racing that ordering.
		proc, err := air.Start(context.WithoutCancel(ctx), air.Options{
			Bin:           airBin.BinPath,
			TmpDir:        runDir,
			SendInterrupt: true, KillDelay: 2 * time.Second,
			Dir: cfg.Root, BuildCmd: "go build -o " + shellQuote(bin) + " " + shellQuote(cfg.BuildPackage),
			Entrypoint: bin,
			ArgsBin:    []string{"--serve-only", "--port", strconv.Itoa(opts.port), "--vite-port", strconv.Itoa(opts.vitePort), "--no-open"},
			Env:        []string{runEnv + "=" + id}, IncludeDir: cfg.WatchDirs,
			ExcludeDir: excluded,
			IncludeExt: []string{"go", "mod", "sum"}, Stdout: cfg.Stdout, Stderr: cfg.Stderr,
		})
		if err != nil {
			return err
		}
		go func() { backendDone <- <-proc.Exited }()
		stopBackend = func() { stopOwnedBackend(appURL, id); proc.Stop() }
	}
	// The no-Air completion channel is observed once, but shutdown must also wait.
	var backendErr error
	go func() { backendErr = <-backendDone; close(backendExited) }()
	defer stopBackend()
	readyCtx, cancelReady := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancelReady()
	ready := make(chan error, 1)
	go func() { ready <- waitReady(readyCtx, appURL, id) }()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-frontendDone:
			return fmt.Errorf("frontend exited: %v", frontend.err)
		case <-backendExited:
			return fmt.Errorf("backend runner exited: %v", backendErr)
		case err := <-ready:
			if err != nil {
				return fmt.Errorf("backend did not become ready within %s: %w", cfg.StartupTimeout, err)
			}
			message(cfg.Stdout, "32", "%s ready: %s%s\n", cfg.Name, appURL, cfg.BrowserPath)
			if !opts.noOpen {
				if _, err := open.URL(appURL + cfg.BrowserPath); err != nil {
					message(cfg.Stderr, "33", "warning: open browser: %v\n", err)
				}
			}
			ready = nil
		}
	}
}
