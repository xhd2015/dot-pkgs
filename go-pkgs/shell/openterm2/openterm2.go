// Package openterm2 opens a directory in iTerm2 when resolvable, else Terminal.app.
package openterm2

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	ViaITerm2   = "iterm2"
	ViaTerminal = "terminal"

	defaultTerminalApp = "/Applications/Utilities/Terminal.app"
)

// Result is the opener that ran and the .app path handed to it.
type Result struct {
	Via     string
	AppPath string
}

// Config injects resolve/open hooks. Nil cfg or nil hooks use production defaults.
type Config struct {
	ResolveITerm func() string
	OpenITerm    func(dir string) error
	OpenTerminal func(dir string) error
	TerminalApp  string
}

// Open is OpenConfig(dir, nil).
func Open(dir string) (*Result, error) {
	return OpenConfig(dir, nil)
}

// OpenConfig validates dir, then opens iTerm2 or Terminal.app.
// A failed iTerm2 open does not fall through to Terminal.
func OpenConfig(dir string, cfg *Config) (*Result, error) {
	if err := validateDir(dir); err != nil {
		return nil, err
	}

	resolve := iterm2.ResolveAppPath
	openITerm := iterm2.Open
	terminalApp := defaultTerminalApp
	var openTerminal func(dir string) error
	if cfg != nil {
		if cfg.ResolveITerm != nil {
			resolve = cfg.ResolveITerm
		}
		if cfg.OpenITerm != nil {
			openITerm = cfg.OpenITerm
		}
		if cfg.OpenTerminal != nil {
			openTerminal = cfg.OpenTerminal
		}
		if cfg.TerminalApp != "" {
			terminalApp = cfg.TerminalApp
		}
	}
	if openTerminal == nil {
		app := terminalApp
		openTerminal = func(d string) error {
			return openWithArgs(TerminalOpenArgs(app, d))
		}
	}

	if appPath := resolve(); appPath != "" {
		if err := openITerm(dir); err != nil {
			return nil, err
		}
		return &Result{Via: ViaITerm2, AppPath: appPath}, nil
	}

	if err := openTerminal(dir); err != nil {
		return nil, err
	}
	return &Result{Via: ViaTerminal, AppPath: terminalApp}, nil
}

// TerminalOpenArgs returns argv for `open -a <appPath> <absDir>`. It does not exec.
func TerminalOpenArgs(appPath, dir string) []string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	return []string{"open", "-a", appPath, absDir}
}

func validateDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("openterm2: dir is empty")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("openterm2: stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("openterm2: not a directory: %s", dir)
	}
	return nil
}

func openWithArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("openterm2: empty open args")
	}
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("openterm2: %w: %s", err, strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("openterm2: %w", err)
	}
	return nil
}
