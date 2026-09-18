package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	lessflags "github.com/xhd2015/less-flags"
)

type options struct {
	port, vitePort           int
	noOpen, noAir, serveOnly bool
}

// Main parses the shared command contract and supervises until Ctrl-C.
func Main(args []string, cfg Config) error {
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	var outputMu sync.Mutex
	cfg.Stdout = &synchronizedWriter{mu: &outputMu, dst: cfg.Stdout}
	cfg.Stderr = &synchronizedWriter{mu: &outputMu, dst: cfg.Stderr}
	help := fmt.Sprintf(`Usage: go run %s [options]
Start %s with Go/Air hot reload and optional Vite HMR.
Missing Air is installed via brew install go-air, then go install as fallback.

Options:
  --port <n>       Backend port (default: first free from %d)
  --vite-port <n>  Frontend port (default: first free from %d)
  --no-open       Do not open a browser
  --use-air       Enable Air (default; compatibility alias)
  --no-use-air    Disable backend hot reload
  --serve-only    Internal backend mode
  -h, --help      Show help
`, cfg.BuildPackage, cfg.Name, cfg.BackendPort, cfg.FrontendPort)
	var opts options
	var useAir bool
	remain, err := lessflags.Int("--port", &opts.port).Int("--vite-port", &opts.vitePort).
		Bool("--no-open", &opts.noOpen).Bool("--use-air", &useAir).
		Bool("--no-use-air", &opts.noAir).Bool("--serve-only", &opts.serveOnly).
		HelpFunc("-h,--help", func() { fmt.Fprint(cfg.Stdout, help) }).HelpNoExit().Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remain, " "))
	}
	if useAir && (opts.noAir || opts.serveOnly) {
		return fmt.Errorf("--use-air conflicts with --no-use-air and --serve-only")
	}
	if opts.port < 0 || opts.port > 65535 || opts.vitePort < 0 || opts.vitePort > 65535 {
		return fmt.Errorf("ports must be between 1 and 65535; 0 selects automatically")
	}
	if cfg.Root == "" || cfg.BackendHandler == nil || cfg.BuildPackage == "" {
		return fmt.Errorf("Root, BuildPackage and BackendHandler are required")
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return err
	}
	cfg.Root = root
	if cfg.StartupTimeout == 0 {
		cfg.StartupTimeout = 60 * time.Second
	}
	if cfg.BackendPort == 0 {
		cfg.BackendPort = 8080
	}
	if cfg.FrontendPort == 0 {
		cfg.FrontendPort = 5173
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if opts.serveOnly {
		if opts.port == 0 || os.Getenv(runEnv) == "" {
			return fmt.Errorf("--serve-only is internal; start the parent dev command")
		}
		return serve(ctx, cfg, opts, os.Getenv(runEnv))
	}
	err = supervise(ctx, cfg, opts)
	if ctx.Err() != nil {
		return nil
	}
	return err
}
