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
	port, vitePort        int
	noOpen, noAir, useAir bool
	serveOnly, kill       bool
	remain                []string
}

type mode int

const (
	devMode mode = iota
	runMode
)

func (m mode) String() string {
	if m == runMode {
		return "run"
	}
	return "dev"
}

// Main parses the shared command contract and supervises until Ctrl-C.
func Main(args []string, cfg Config) error {
	return dispatch(args, cfg, devMode)
}

// RunMain parses the one-shot command contract, builds the configured frontend
// and the backend binary once, serves, and supervises until Ctrl-C. It shares
// the dev supervisor's directory control, so run and dev in the same directory
// stop each other.
func RunMain(args []string, cfg Config) error {
	return dispatch(args, cfg, runMode)
}

func dispatch(args []string, cfg Config, mode mode) error {
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	var outputMu sync.Mutex
	cfg.Stdout = &synchronizedWriter{mu: &outputMu, dst: cfg.Stdout}
	cfg.Stderr = &synchronizedWriter{mu: &outputMu, dst: cfg.Stderr}
	opts, err := parseArgs(args, cfg, mode)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(opts.remain) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(opts.remain, " "))
	}
	if mode == devMode && opts.useAir && (opts.noAir || opts.serveOnly) {
		return fmt.Errorf("--use-air conflicts with --no-use-air and --serve-only")
	}
	if opts.port < 0 || opts.port > 65535 || (mode == devMode && (opts.vitePort < 0 || opts.vitePort > 65535)) {
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
			return fmt.Errorf("--serve-only is internal; start the parent %s command", mode)
		}
		return serve(ctx, cfg, opts, os.Getenv(runEnv))
	}
	control, err := openDirectoryControl(cfg.Root)
	if err != nil {
		return err
	}
	defer control.unlock()
	found, err := control.stopExisting(cfg)
	if err != nil {
		return err
	}
	if opts.kill {
		if !found {
			message(cfg.Stdout, "90", "No supervisor found for %s.\n", cfg.Root)
		}
		return nil
	}
	if err := control.start(cancel); err != nil {
		return err
	}
	defer control.close()
	control.unlock()
	if mode == devMode {
		err = supervise(ctx, cfg, opts)
	} else {
		err = runOnce(ctx, cfg, opts)
	}
	if ctx.Err() != nil {
		return nil
	}
	return err
}

func parseArgs(args []string, cfg Config, mode mode) (options, error) {
	var opts options
	if mode == runMode {
		if err := rejectDevFlags(args); err != nil {
			return opts, err
		}
	}
	less := lessflags.Int("--port", &opts.port)
	if mode == devMode {
		less = less.Int("--vite-port", &opts.vitePort).
			Bool("--use-air", &opts.useAir).
			Bool("--no-use-air", &opts.noAir)
	}
	help := helpText(cfg, mode)
	remain, err := less.Bool("--no-open", &opts.noOpen).Bool("--serve-only", &opts.serveOnly).
		Bool("--kill", &opts.kill).
		HelpFunc("-h,--help", func() { fmt.Fprint(cfg.Stdout, help) }).HelpNoExit().Parse(args)
	if err != nil {
		return opts, err
	}
	opts.remain = remain
	return opts, nil
}

// rejectDevFlags gives run users a targeted message for hot-reload flags
// instead of a generic unknown-argument error.
func rejectDevFlags(args []string) error {
	for _, arg := range args {
		if arg == "--" {
			return nil
		}
		name := arg
		if eq := strings.Index(arg, "="); eq >= 0 {
			name = arg[:eq]
		}
		switch name {
		case "--vite-port":
			return fmt.Errorf("--vite-port is a dev flag; run serves the built frontend instead")
		case "--use-air", "--no-use-air", "--no-air":
			return fmt.Errorf("%s is a dev flag; run has no watch", name)
		}
	}
	return nil
}

func helpText(cfg Config, mode mode) string {
	if mode == runMode {
		return fmt.Sprintf(`Usage: go run %s [options]
Start %s with a one-shot build (no watch): build the configured frontend and
the backend binary once, then serve the built app until Ctrl-C.
Each nonzero port flag is strict; unspecified ports are selected automatically.

Options:
  --port <n>    Backend port (default: first free from %d)
  --kill        Stop the supervisor for this directory
  --no-open     Do not open a browser
  -h, --help    Show help
`, cfg.BuildPackage, cfg.Name, cfg.BackendPort)
	}
	return fmt.Sprintf(`Usage: go run %s [options]
Start %s with Go/Air hot reload and optional Vite HMR.
Missing Air is installed via brew install go-air, then go install as fallback.
Each nonzero port flag is strict; unspecified ports are selected automatically.

Options:
  --port <n>       Backend port (default: first free from %d)
  --vite-port <n>  Frontend port (default: first free from %d)
  --kill           Stop the supervisor for this directory
  --no-open        Do not open a browser
  --use-air        Enable Air (default; compatibility alias)
  --no-use-air     Disable backend hot reload
  --serve-only     Internal backend mode
  -h, --help       Show help
`, cfg.BuildPackage, cfg.Name, cfg.BackendPort, cfg.FrontendPort)
}
