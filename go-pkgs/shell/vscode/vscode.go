// Package vscode opens a directory or a file in Visual Studio Code.
//
// The `code` CLI is often missing from a thin PATH (a menu-bar app, a daemon,
// or a login shell's environment), so resolution runs through
// shell/lookpath: PATH first, then the well-known app bundles, then login
// shells. Opening a directory reuses the current window by default, which is
// what a "Open in VS Code" button should do.
package vscode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// Result is one launch: the resolved CLI, the argv that ran, and the runner's
// captured output (empty for the detached default runner).
type Result struct {
	CodePath string
	Via      string
	Args     []string
	Out      string
}

// Config injects resolution and the process runner. A nil Config, or a nil
// hook on it, means the production default.
type Config struct {
	// Home resolves the per-user app bundles. Empty → os.UserHomeDir.
	Home string
	// NewWindow opens a new window (-n) instead of reusing one (-r).
	NewWindow bool
	// Resolve finds the `code` CLI. nil → Resolve(lookpath.Options{Home}).
	Resolve func(home string) (lookpath.Result, error)
	// Run executes argv; args[0] is the resolved CLI. nil → detached start.
	Run func(args []string) (string, error)
}

func (c *Config) home() string {
	if c != nil && strings.TrimSpace(c.Home) != "" {
		return strings.TrimSpace(c.Home)
	}
	home, _ := os.UserHomeDir()
	return home
}

func (c *Config) resolve() func(home string) (lookpath.Result, error) {
	if c != nil && c.Resolve != nil {
		return c.Resolve
	}
	return func(home string) (lookpath.Result, error) {
		return Resolve(lookpath.Options{Home: home})
	}
}

func (c *Config) run() func(args []string) (string, error) {
	if c != nil && c.Run != nil {
		return c.Run
	}
	return DetachRunner
}

func (c *Config) newWindow() bool {
	return c != nil && c.NewWindow
}

// DetachRunner starts argv and lets it go: VS Code outlives the CLI that
// handed the request over, so the caller must not wait for it. It is the
// default Config.Run and reports only a failure to start.
func DetachRunner(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("vscode: %w", err)
	}
	_ = cmd.Process.Release()
	return "", nil
}

// WaitRunner runs argv to completion with inherited stdio and returns any exit
// error. A CLI caller that must report a failed launch picks this as
// Config.Run.
func WaitRunner(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("vscode: %w", err)
	}
	return "", nil
}

// WellKnownCandidates returns the absolute `code` CLI paths inside the Visual
// Studio Code and VS Code Insiders bundles: system-wide first, then per-user.
func WellKnownCandidates(home string) []string {
	apps := []string{
		"Visual Studio Code.app",
		"Visual Studio Code - Insiders.app",
	}
	var out []string
	for _, name := range apps {
		out = append(out, filepath.Join("/Applications", name, "Contents", "Resources", "app", "bin", "code"))
	}
	if home = strings.TrimSpace(home); home != "" {
		for _, name := range apps {
			out = append(out, filepath.Join(home, "Applications", name, "Contents", "Resources", "app", "bin", "code"))
		}
	}
	return out
}

// Resolve finds the `code` CLI: PATH first, then the app bundles, then login
// shells. A caller that needs its own search order can pass ExtraCandidates.
func Resolve(opts lookpath.Options) (lookpath.Result, error) {
	if opts.ExtraCandidates == nil {
		home := opts.Home
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		opts.ExtraCandidates = WellKnownCandidates(home)
	}
	return lookpath.Look("code", opts)
}

// Args returns the argv that opens dir, the resolved CLI first. Reuse (-r) is
// the default so a button press does not pile up windows.
func Args(codePath, dir string, newWindow bool) []string {
	flag := "-r"
	if newWindow {
		flag = "-n"
	}
	return []string{codePath, flag, dir}
}

// FileArgs returns the argv that opens file, at line when line is positive.
func FileArgs(codePath, file string, line int) ([]string, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, fmt.Errorf("vscode: file is empty")
	}
	if line > 0 {
		return []string{codePath, "--goto", fmt.Sprintf("%s:%d", file, line)}, nil
	}
	return []string{codePath, file}, nil
}

// Open opens dir in VS Code, reusing an existing window.
func Open(dir string) (*Result, error) {
	return OpenConfig(dir, nil)
}

// OpenConfig is Open with an injectable resolver and runner.
//
// The path is not stat-ed: `code` creates a new buffer for a path that is not
// there yet, and the existing callers (a workspace root, a configured repo
// directory) already hold a real one.
func OpenConfig(dir string, cfg *Config) (*Result, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("vscode: dir is empty")
	}
	resolved, err := cfg.resolve()(cfg.home())
	if err != nil {
		return nil, fmt.Errorf("vscode: %w", err)
	}
	return launch(cfg, resolved, Args(resolved.Path, dir, cfg.newWindow()))
}

// OpenFile opens file, at line when line is positive.
func OpenFile(file string, line int) (*Result, error) {
	return OpenFileConfig(file, line, nil)
}

// OpenFileConfig is OpenFile with an injectable resolver and runner.
func OpenFileConfig(file string, line int, cfg *Config) (*Result, error) {
	if strings.TrimSpace(file) == "" {
		return nil, fmt.Errorf("vscode: file is empty")
	}
	resolved, err := cfg.resolve()(cfg.home())
	if err != nil {
		return nil, fmt.Errorf("vscode: %w", err)
	}
	argv, err := FileArgs(resolved.Path, file, line)
	if err != nil {
		return nil, err
	}
	return launch(cfg, resolved, argv)
}

func launch(cfg *Config, resolved lookpath.Result, argv []string) (*Result, error) {
	out, err := cfg.run()(argv)
	if err != nil {
		return nil, err
	}
	return &Result{CodePath: resolved.Path, Via: resolved.Via, Args: argv, Out: out}, nil
}
