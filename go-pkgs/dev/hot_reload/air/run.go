package air

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Options configures Start / Run.
type Options struct {
	// Dir is the working directory for air (usually module root). Required.
	Dir string
	// BuildCmd is passed to --build.cmd. Required.
	BuildCmd string
	// Entrypoint is passed to --build.entrypoint (relative path preferred). Required.
	Entrypoint string
	// ArgsBin are repeated --build.args_bin values for the rebuilt binary.
	ArgsBin []string
	// DelayMs is --build.delay (default 500).
	DelayMs int
	// IncludeExt defaults to ["go"].
	IncludeExt []string
	// IncludeDir is optional --build.include_dir (comma-joined).
	IncludeDir []string
	// ExcludeDir is optional --build.exclude_dir (comma-joined).
	ExcludeDir []string
	// ExcludeRegex defaults to `_test\.go`.
	ExcludeRegex string
	// ExcludeUnchanged defaults to true when nil.
	ExcludeUnchanged *bool

	// Bin overrides the air binary path. Empty → LookPath("air") after Ensure.
	Bin string
	// SkipEnsure skips Ensure when Bin is empty (caller must ensure air exists).
	SkipEnsure bool
	Ensure     EnsureOpts

	Stdout io.Writer
	Stderr io.Writer

	// Command builds the air exec.Cmd. Nil → default with Setpgid.
	// Tests inject a fake command here.
	Command func(ctx context.Context, bin string, args []string, dir string, stdout, stderr io.Writer) *exec.Cmd

	// StopTimeout is how long to wait after SIGTERM before SIGKILL (default 5s).
	StopTimeout time.Duration
	// OnStop runs once after the air process group is stopped (e.g. kill port, remove bin).
	OnStop func()
}

// Process is a running air instance.
type Process struct {
	// Exited receives when air Wait returns (buffered; receive at most once).
	Exited <-chan error

	cmd *exec.Cmd
	// done is closed after Wait returns; Stop waits on done (safe if Exited was already read).
	done   chan struct{}
	stderr io.Writer

	stopOnce    sync.Once
	stopTimeout time.Duration
	onStop      func()
}

// Start ensures air (unless SkipEnsure), builds args, and starts the process group.
// Caller should select on Exited and call Stop on shutdown.
func Start(ctx context.Context, opts Options) (*Process, error) {
	if strings.TrimSpace(opts.Dir) == "" {
		return nil, fmt.Errorf("air: Dir is required")
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	bin := strings.TrimSpace(opts.Bin)
	if bin == "" {
		if !opts.SkipEnsure {
			ensureOpts := opts.Ensure
			if ensureOpts.Stdout == nil {
				ensureOpts.Stdout = stdout
			}
			if ensureOpts.Stderr == nil {
				ensureOpts.Stderr = stderr
			}
			res, err := Ensure(ctx, ensureOpts)
			if err != nil {
				return nil, err
			}
			bin = res.BinPath
		} else {
			lookPath := opts.Ensure.LookPath
			if lookPath == nil {
				lookPath = exec.LookPath
			}
			var err error
			bin, err = lookPath("air")
			if err != nil {
				return nil, fmt.Errorf("air: %w", err)
			}
		}
	}

	args, err := BuildArgs(opts)
	if err != nil {
		return nil, err
	}

	buildCmd := func(ctx context.Context, bin string, args []string, dir string, stdout, stderr io.Writer) *exec.Cmd {
		if opts.Command != nil {
			return opts.Command(ctx, bin, args, dir, stdout, stderr)
		}
		c := exec.CommandContext(ctx, bin, args...)
		c.Dir = dir
		c.Stdout = stdout
		c.Stderr = stderr
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return c
	}

	cmd := buildCmd(ctx, bin, args, opts.Dir, stdout, stderr)
	fmt.Fprintf(stderr, "starting air (BE watch)...\n")
	fmt.Fprintf(stderr, "  build: %s\n", opts.BuildCmd)
	fmt.Fprintf(stderr, "  run:   %s %s\n", opts.Entrypoint, strings.Join(opts.ArgsBin, " "))

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start air: %w", err)
	}

	exited := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		exited <- cmd.Wait()
		close(done)
	}()

	stopTimeout := opts.StopTimeout
	if stopTimeout <= 0 {
		stopTimeout = 5 * time.Second
	}

	return &Process{
		Exited:      exited,
		cmd:         cmd,
		done:        done,
		stderr:      stderr,
		stopTimeout: stopTimeout,
		onStop:      opts.OnStop,
	}, nil
}

// Stop terminates the air process group (SIGTERM, then SIGKILL) and runs OnStop once.
func (p *Process) Stop() {
	if p == nil {
		return
	}
	p.stopOnce.Do(func() {
		fmt.Fprintln(p.stderr, "stopping air...")
		if p.cmd != nil && p.cmd.Process != nil {
			pgid := p.cmd.Process.Pid
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
			select {
			case <-p.done:
			case <-time.After(p.stopTimeout):
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
				select {
				case <-p.done:
				case <-time.After(2 * time.Second):
				}
			}
		}
		if p.onStop != nil {
			p.onStop()
		}
	})
}

// Run starts air and blocks until ctx is cancelled or air exits.
// On ctx cancel it stops air and returns nil. On air exit it returns a wrapped error.
func Run(ctx context.Context, opts Options) error {
	p, err := Start(ctx, opts)
	if err != nil {
		return err
	}
	defer p.Stop()

	select {
	case <-ctx.Done():
		return nil
	case err := <-p.Exited:
		if err != nil {
			return fmt.Errorf("air exited: %w", err)
		}
		return fmt.Errorf("air exited")
	}
}
