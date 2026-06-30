// Package iterm2 opens iTerm2 on macOS with smart window/tab reuse and cds
// to the requested path, optionally running follow-up shell commands.
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
	"strconv"
	"strings"
)

const AppPath = "/Applications/iTerm.app"

const (
	envInstalled     = "KOOL_ITERM2_INSTALLED"
	envScriptOut     = "KOOL_ITERM2_SCRIPT_OUT"
	envOsascriptExit = "KOOL_ITERM2_OSASCRIPT_EXIT"
	envGOOS          = "KOOL_ITERM2_GOOS"
)

var (
	ErrUnsupportedPlatform = errors.New("iterm2: unsupported platform (macOS only)")
	ErrNotInstalled        = errors.New("iterm2: iTerm2 is not installed")
)

var testGOOS string

// SetGOOSForTest overrides the platform check for tests. Pass "" to reset.
func SetGOOSForTest(goos string) {
	testGOOS = goos
}

func effectiveGOOS() string {
	if testGOOS != "" {
		return testGOOS
	}
	if v := os.Getenv(envGOOS); v != "" {
		return v
	}
	return runtime.GOOS
}

// OpenMode selects smart-open vs reuse-current-session behavior.
type OpenMode int

const (
	// ModeSmart scans session paths and reuses window with new tab or new window.
	ModeSmart OpenMode = iota
	// ModeReuseCurrent cds in current session of current tab/window (kool -r).
	ModeReuseCurrent
)

// Config customizes Open for tests or alternate runners.
type Config struct {
	// Osascript runs AppleScript. When nil, default runner is used.
	Osascript func(script string) error
	// Installed reports whether iTerm2 is present. When nil, AppPath / env is checked.
	Installed func() bool
	// FollowUpCommands are shell commands written after cd (OpenConfig only).
	FollowUpCommands []string
	// Mode defaults to ModeSmart when zero.
	Mode OpenMode
}

// IsInstalled reports whether iTerm2.app exists at AppPath.
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

// EscapeCommandForAppleScript escapes a command for AppleScript write text literals.
func EscapeCommandForAppleScript(command string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	).Replace(command)
}

func buildSessionCommandLines(followUpCommands []string) []string {
	lines := []string{
		`        write text ("cd " & quoted form of targetDir)`,
	}
	for _, command := range followUpCommands {
		lines = append(lines, fmt.Sprintf(`        write text "%s"`, EscapeCommandForAppleScript(command)))
	}
	return lines
}

// BuildPathScanSmokeScript returns AppleScript that probes session path variables.
func BuildPathScanSmokeScript() string {
	lines := []string{
		`tell application "iTerm2"`,
		`  repeat with aWindow in windows`,
		`    repeat with aTab in tabs of aWindow`,
		`      repeat with aSession in sessions of aTab`,
		`        try`,
		`          tell aSession`,
		`            set sessionPath to variable named "path"`,
		`          end tell`,
		`        on error`,
		`        end try`,
		`      end repeat`,
		`    end repeat`,
		`  end repeat`,
		`  return "ok"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// BuildScript returns AppleScript that smart-opens iTerm2 and cds to dirPath.
func BuildScript(dirPath string, followUps ...string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set targetDir to "` + escaped + `"`,
		`  set matchingWindow to missing value`,
		`  repeat with aWindow in windows`,
		`    if not (is hotkey window of aWindow) then`,
		`      repeat with aTab in tabs of aWindow`,
		`        repeat with aSession in sessions of aTab`,
		`          try`,
		`            tell aSession`,
		`              set sessionPath to variable named "path"`,
		`            end tell`,
		`            if sessionPath is targetDir then`,
		`              set matchingWindow to aWindow`,
		`              exit repeat`,
		`            end if`,
		`          on error`,
		`          end try`,
		`        end repeat`,
		`        if matchingWindow is not missing value then exit repeat`,
		`      end repeat`,
		`      if matchingWindow is not missing value then exit repeat`,
		`    end if`,
		`  end repeat`,
		`  if matchingWindow is not missing value then`,
		`    tell matchingWindow`,
		`      create tab with default profile`,
		`      tell current session of current tab`,
	}
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`      end tell`,
		`      select`,
		`    end tell`,
		`  else`,
		`    set newWindow to (create window with default profile)`,
		`    tell current session of newWindow`,
	)
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`    end tell`,
		`  end if`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// BuildReuseCurrentSessionScript returns AppleScript that cds in the current iTerm2 session.
func BuildReuseCurrentSessionScript(dirPath string, followUps ...string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set targetDir to "` + escaped + `"`,
		`  if (count of windows) is 0 then`,
		`    create window with default profile`,
		`  end if`,
		`  tell current session of current tab of current window`,
	}
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`  end tell`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

func normalizeTargetDirectory(dirPath string) (string, error) {
	abs, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return abs, nil
		}
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil
	}
	return real, nil
}

// Open opens iTerm2 with smart reuse and cds to dir.
func Open(dir string) error {
	return OpenConfig(dir, nil)
}

// OpenConfig is like Open but allows injecting dependencies via cfg.
func OpenConfig(dir string, cfg *Config) error {
	if effectiveGOOS() != "darwin" {
		return ErrUnsupportedPlatform
	}
	if !installedCheck(cfg) {
		return fmt.Errorf("%w. Install it from https://iterm2.com/", ErrNotInstalled)
	}

	target, err := normalizeTargetDirectory(dir)
	if err != nil {
		return fmt.Errorf("iterm2: resolve dir: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("iterm2: stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("iterm2: not a directory: %s", target)
	}

	var followUps []string
	mode := ModeSmart
	if cfg != nil {
		if len(cfg.FollowUpCommands) > 0 {
			followUps = cfg.FollowUpCommands
		}
		mode = cfg.Mode
	}
	var script string
	if mode == ModeReuseCurrent {
		script = BuildReuseCurrentSessionScript(target, followUps...)
	} else {
		script = BuildScript(target, followUps...)
	}
	if err := osascriptRunner(cfg)(script); err != nil {
		return fmt.Errorf("iterm2: osascript: %w", err)
	}
	return nil
}

func installedCheck(cfg *Config) bool {
	if cfg != nil && cfg.Installed != nil {
		return cfg.Installed()
	}
	switch os.Getenv(envInstalled) {
	case "0":
		return false
	case "1":
		return true
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
	if outPath := os.Getenv(envScriptOut); outPath != "" {
		if err := os.WriteFile(outPath, []byte(script), 0644); err != nil {
			return fmt.Errorf("write script out: %w", err)
		}
		if code, ok := osascriptExitFromEnv(); ok && code != 0 {
			return fmt.Errorf("osascript exited with status %d", code)
		}
		return nil
	}

	cmd := exec.Command("osascript", "-e", script)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if code, ok := osascriptExitFromEnv(); ok && code != 0 {
		return fmt.Errorf("osascript exited with status %d", code)
	}
	return nil
}

func osascriptExitFromEnv() (int, bool) {
	raw := os.Getenv(envOsascriptExit)
	if raw == "" {
		return 0, false
	}
	code, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return code, true
}