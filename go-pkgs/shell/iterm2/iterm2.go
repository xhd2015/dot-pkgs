// Package iterm2 opens a new iTerm2 window on macOS and changes directory
// to the requested path.
//
// Example:
//
//	if err := iterm2.Open("/path/to/project"); err != nil {
//	    log.Fatal(err)
//	}
package iterm2

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const AppPath = "/Applications/iTerm.app"

var (
	ErrUnsupportedPlatform = errors.New("iterm2: unsupported platform (macOS only)")
	ErrNotInstalled        = errors.New("iterm2: iTerm2 is not installed")
)

// Config customizes Open for tests or alternate runners.
type Config struct {
	// Osascript runs AppleScript. When nil, /usr/bin/osascript is used.
	Osascript func(script string) error
	// Installed reports whether iTerm2 is present. When nil, AppPath is checked.
	Installed func() bool
}

// Installed reports whether iTerm2.app exists at AppPath.
func IsInstalled() bool {
	_, err := os.Stat(AppPath)
	return err == nil
}

// EscapePathForAppleScript escapes a path embedded in an AppleScript string literal.
func EscapePathForAppleScript(dirPath string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	).Replace(dirPath)
}

// BuildScript returns AppleScript that creates a new iTerm2 window and cds
// to dirPath. The default login shell stays running so the window does not close.
func BuildScript(dirPath string) string {
	escaped := EscapePathForAppleScript(dirPath)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set newWindow to (create window with default profile)`,
		`  tell current session of newWindow`,
		`    set targetDir to "` + escaped + `"`,
		`    write text ("cd " & quoted form of targetDir)`,
		`  end tell`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// Open opens a new iTerm2 window and cds to dir.
func Open(dir string) error {
	return OpenConfig(dir, nil)
}

// OpenConfig is like Open but allows injecting dependencies via cfg.
func OpenConfig(dir string, cfg *Config) error {
	if runtime.GOOS != "darwin" {
		return ErrUnsupportedPlatform
	}
	if installed := installedCheck(cfg); !installed {
		return ErrNotInstalled
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("iterm2: resolve dir: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("iterm2: stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("iterm2: not a directory: %s", abs)
	}

	script := BuildScript(abs)
	if err := osascriptRunner(cfg)(script); err != nil {
		return fmt.Errorf("iterm2: osascript: %w", err)
	}
	return nil
}

func installedCheck(cfg *Config) bool {
	if cfg != nil && cfg.Installed != nil {
		return cfg.Installed()
	}
	return IsInstalled()
}

func osascriptRunner(cfg *Config) func(string) error {
	if cfg != nil && cfg.Osascript != nil {
		return cfg.Osascript
	}
	return defaultOsascript
}

func defaultOsascript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}