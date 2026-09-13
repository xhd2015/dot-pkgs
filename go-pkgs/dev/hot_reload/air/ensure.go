// Package air ensures the air CLI is on PATH and runs it for Go rebuild-on-change.
//
// Callers (typically script/dev) own Vite/orchestration; this package owns air
// install and the air process-group lifecycle.
package air

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	// BrewFormula installs the `air` binary. Homebrew formula `air` is an R
	// formatter; use go-air instead.
	BrewFormula = "go-air"

	// GoInstallPath is the module path for `go install`.
	GoInstallPath = "github.com/air-verse/air@latest"
)

// EnsureOpts configures Ensure.
type EnsureOpts struct {
	LookPath func(file string) (string, error)
	// RunCmd runs name with args (stdout/stderr already attached by caller via
	// Stdout/Stderr when using defaults). Nil → exec.CommandContext.
	RunCmd func(ctx context.Context, name string, args ...string) error
	// GoEnv returns `go env <name>` trimmed. Nil → run go env.
	GoEnv  func(name string) (string, error)
	Stdout io.Writer
	Stderr io.Writer
}

// EnsureResult is the outcome of Ensure.
type EnsureResult struct {
	// Action is one of: noop | brew | go_install
	Action  string
	BinPath string
}

// Ensure makes the air binary available on PATH.
//
// If already present, returns Action=noop. Otherwise tries brew install go-air,
// then go install github.com/air-verse/air@latest.
func Ensure(ctx context.Context, opts EnsureOpts) (EnsureResult, error) {
	var result EnsureResult
	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	runCmd := opts.RunCmd
	if runCmd == nil {
		runCmd = func(ctx context.Context, name string, args ...string) error {
			c := exec.CommandContext(ctx, name, args...)
			c.Stdout = stdout
			c.Stderr = stderr
			return c.Run()
		}
	}

	if bin, err := lookPath("air"); err == nil {
		result.Action = "noop"
		result.BinPath = bin
		return result, nil
	}

	fmt.Fprintln(stderr, "air not found on PATH; installing...")

	if _, err := lookPath("brew"); err == nil {
		fmt.Fprintf(stderr, "installing air via: brew install %s\n", BrewFormula)
		if err := runCmd(ctx, "brew", "install", BrewFormula); err != nil {
			fmt.Fprintf(stderr, "warning: brew install %s failed: %v\n", BrewFormula, err)
		} else if bin, err := lookPath("air"); err == nil {
			fmt.Fprintln(stderr, "air installed (brew go-air)")
			result.Action = "brew"
			result.BinPath = bin
			return result, nil
		}
	}

	if _, err := lookPath("go"); err != nil {
		return result, fmt.Errorf("air not installed; install with: brew install %s\n  or: go install %s", BrewFormula, GoInstallPath)
	}
	fmt.Fprintf(stderr, "installing air via: go install %s\n", GoInstallPath)
	if err := runCmd(ctx, "go", "install", GoInstallPath); err != nil {
		return result, fmt.Errorf("go install air: %w\n  or: brew install %s", err, BrewFormula)
	}
	bin, err := lookPath("air")
	if err != nil {
		goEnv := opts.GoEnv
		if goEnv == nil {
			goEnv = defaultGoEnv
		}
		gopath, _ := goEnv("GOPATH")
		gobin, _ := goEnv("GOBIN")
		return result, fmt.Errorf("air installed but not on PATH (check GOBIN=%s GOPATH/bin=%s/bin)",
			gobin, gopath)
	}
	fmt.Fprintln(stderr, "air installed (go install)")
	result.Action = "go_install"
	result.BinPath = bin
	return result, nil
}

func defaultGoEnv(name string) (string, error) {
	out, err := exec.Command("go", "env", name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
