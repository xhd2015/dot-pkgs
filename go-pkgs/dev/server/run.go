package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/open"
)

// runOnce builds the configured frontend and backend binary once, starts the
// backend as a child speaking the shared --serve-only contract, and waits
// until Ctrl-C or an unexpected child exit. Artifacts live in the system temp
// dir and are removed on exit, so the repository stays clean.
func runOnce(ctx context.Context, cfg Config, opts options) error {
	port, err := selectPort(opts.port, cfg.BackendPort, 0, cfg, "backend")
	if err != nil {
		return err
	}
	opts.port = port
	runDir, err := os.MkdirTemp("", "dot-pkgs-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(runDir)
	if build := cfg.FrontendBuild; build != nil {
		if _, err := os.Stat(filepath.Join(cfg.Root, build.Dir, "node_modules")); os.IsNotExist(err) && len(build.Install) > 0 {
			if err := runCommand(ctx, filepath.Join(cfg.Root, build.Dir), build.Install, cfg); err != nil {
				return fmt.Errorf("frontend install: %w", err)
			}
		}
		message(cfg.Stdout, "90", "Building frontend...\n")
		if err := runCommand(ctx, filepath.Join(cfg.Root, build.Dir), build.Command, cfg); err != nil {
			return fmt.Errorf("frontend build: %w", err)
		}
	}
	message(cfg.Stdout, "90", "Building backend on %d...\n", opts.port)
	bin := filepath.Join(runDir, "backend")
	if err := runCommand(ctx, cfg.Root, []string{"go", "build", "-o", bin, cfg.BuildPackage}, cfg); err != nil {
		return fmt.Errorf("backend build: %w", err)
	}
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	id := hex.EncodeToString(bytes)
	appURL := fmt.Sprintf("http://127.0.0.1:%d", opts.port)
	backend, err := startChild(cfg.Root, []string{bin, "--serve-only", "--port", strconv.Itoa(opts.port), "--no-open"},
		[]string{runEnv + "=" + id}, cfg)
	if err != nil {
		return fmt.Errorf("backend start: %w", err)
	}
	backendExited := make(chan struct{})
	go func() { <-backend.done; close(backendExited) }()
	// Acknowledge only after application cleanup, then stop the owned group.
	defer func() { stopOwnedBackend(appURL, id); backend.stop() }()
	readyCtx, cancelReady := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancelReady()
	ready := make(chan error, 1)
	go func() { ready <- waitReady(readyCtx, appURL, id) }()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-backendExited:
			return fmt.Errorf("backend exited: %v", backend.err)
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
