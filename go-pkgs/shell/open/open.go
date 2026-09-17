// Package open wraps the desktop "open" primitive: hand a path, a URL, or a
// path plus an application to the operating system and let it launch.
//
// It exists so callers stop shelling out to `open` / `xdg-open` / `cmd start`
// by hand. The argv shape is a pure function of the target and the platform,
// and the process runner is injectable, so the whole surface is testable
// without launching anything.
//
// Opening a directory in a terminal is a different concern: see
// shell/openterm2 (iTerm2, else Terminal.app) and shell/iterm2.
package open

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Result is one launch: the argv that ran and the runner's captured output.
type Result struct {
	Args []string
	Out  string
}

// Config injects the platform and the process runner. A nil Config, or a nil
// hook on it, means the production default.
type Config struct {
	// GOOS overrides runtime.GOOS.
	GOOS string
	// Run executes argv and returns its combined output. nil → exec.
	Run func(args []string) (string, error)
}

func (c *Config) goos() string {
	if c != nil && c.GOOS != "" {
		return c.GOOS
	}
	return runtime.GOOS
}

func (c *Config) run() func(args []string) (string, error) {
	if c != nil && c.Run != nil {
		return c.Run
	}
	return execRun
}

func execRun(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("open: %w: %s", err, text)
		}
		return text, fmt.Errorf("open: %w", err)
	}
	return text, nil
}

func launch(cfg *Config, build func(goos string) ([]string, error)) (*Result, error) {
	argv, err := build(cfg.goos())
	if err != nil {
		return nil, err
	}
	out, err := cfg.run()(argv)
	if err != nil {
		return nil, err
	}
	return &Result{Args: argv, Out: out}, nil
}

// Args returns the argv that opens target with the platform's default handler.
// It does not exec, and an empty target is an error on every platform.
func Args(target, goos string) ([]string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, fmt.Errorf("open: target is empty")
	}
	switch goos {
	case "darwin":
		return []string{"open", target}, nil
	case "linux":
		return []string{"xdg-open", target}, nil
	case "windows":
		// `start` reads the first quoted argument as a window title, so an
		// empty one must come before the target.
		return []string{"cmd", "/c", "start", "", target}, nil
	default:
		return nil, fmt.Errorf("open: unsupported platform %s", goos)
	}
}

// URLArgs is Args for a URL. The argv shape is identical; the separate name
// documents the intent at the call site.
func URLArgs(url, goos string) ([]string, error) {
	return Args(url, goos)
}

// AppArgs returns the argv that opens path with app: `open -a <app> [path]`.
// app is an application name ("Google Chrome") or an absolute .app path; an
// empty path launches the application itself. `open -a` is macOS only.
func AppArgs(app, path, goos string) ([]string, error) {
	app = strings.TrimSpace(app)
	if app == "" {
		return nil, fmt.Errorf("open: app is empty")
	}
	if goos != "darwin" {
		return nil, fmt.Errorf("open: open -a is unsupported on %s", goos)
	}
	argv := []string{"open", "-a", app}
	if p := strings.TrimSpace(path); p != "" {
		argv = append(argv, p)
	}
	return argv, nil
}

// Path opens a file, directory, or .app bundle with the default handler.
func Path(path string) (*Result, error) {
	return PathConfig(path, nil)
}

// PathConfig is Path with an injectable platform and runner.
func PathConfig(path string, cfg *Config) (*Result, error) {
	target := strings.TrimSpace(path)
	return launch(cfg, func(goos string) ([]string, error) { return Args(target, goos) })
}

// URL opens a URL with the default handler (the browser).
func URL(url string) (*Result, error) {
	return URLConfig(url, nil)
}

// URLConfig is URL with an injectable platform and runner.
func URLConfig(url string, cfg *Config) (*Result, error) {
	target := strings.TrimSpace(url)
	return launch(cfg, func(goos string) ([]string, error) { return URLArgs(target, goos) })
}

// App opens path with the named or bundled application. An empty path launches
// the application itself.
func App(app, path string) (*Result, error) {
	return AppConfig(app, path, nil)
}

// AppConfig is App with an injectable platform and runner.
func AppConfig(app, path string, cfg *Config) (*Result, error) {
	return launch(cfg, func(goos string) ([]string, error) { return AppArgs(app, path, goos) })
}
