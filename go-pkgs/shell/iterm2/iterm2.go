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

// AppPath is the system-wide default install path (/Applications/iTerm.app).
// Prefer ResolveAppPath() for detection: it also checks ~/Applications first.
const AppPath = "/Applications/iTerm.app"

// KoolTargetDirVar is the iTerm2 user session variable used to track kool-opened dirs.
const KoolTargetDirVar = "user.koolTargetDir"

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
	// ModeReuseCurrent focuses a session at targetDir when found, else new window + cd (kool -r).
	ModeReuseCurrent
	// ModeForceNew always opens a new window, skipping session scan (kool -n).
	ModeForceNew
)

// Config customizes Open for tests or alternate runners.
type Config struct {
	// Osascript runs AppleScript. When nil, default runner is used.
	Osascript func(script string) error
	// Installed reports whether iTerm2 is present. When nil, ResolveAppPath / env is checked.
	Installed func() bool
	// FollowUpCommands are shell commands written after cd (OpenConfig only).
	FollowUpCommands []string
	// Mode defaults to ModeSmart when zero.
	Mode OpenMode
}

// appCandidates returns install paths in preference order: ~/Applications then /Applications.
func appCandidates() []string {
	cands := make([]string, 0, 2)
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		cands = append(cands, filepath.Join(home, "Applications", "iTerm.app"))
	}
	cands = append(cands, AppPath)
	return cands
}

// resolveAppPathAmong returns the first existing .app bundle path among candidates, or "".
func resolveAppPathAmong(candidates []string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			continue
		}
		return p
	}
	return ""
}

// ResolveAppPath returns the preferred existing iTerm2.app path.
// Order: ~/Applications/iTerm.app, then /Applications/iTerm.app.
// Returns "" when none exist.
func ResolveAppPath() string {
	return resolveAppPathAmong(appCandidates())
}

// IsInstalled reports whether an iTerm2.app bundle exists under ~/Applications or /Applications.
func IsInstalled() bool {
	return ResolveAppPath() != ""
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
		`          set sessionPath to ""`,
		`          set koolTargetDir to ""`,
		`          try`,
		`            tell aSession`,
		`              set sessionPath to variable named "path"`,
		`            end tell`,
		`          on error`,
		`          end try`,
		`          try`,
		`            tell aSession`,
		`              set koolTargetDir to variable named "` + KoolTargetDirVar + `"`,
		`            end tell`,
		`          on error`,
		`          end try`,
		`          if sessionPath is targetDir or koolTargetDir is targetDir then`,
		`            set matchingWindow to aWindow`,
		`            exit repeat`,
		`          end if`,
		`        end repeat`,
		`        if matchingWindow is not missing value then exit repeat`,
		`      end repeat`,
		`      if matchingWindow is not missing value then exit repeat`,
		`    end if`,
		`  end repeat`,
		`  if matchingWindow is not missing value then`,
		`    tell matchingWindow`,
		`      set newTab to (create tab with default profile)`,
		`      tell current session of newTab`,
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
		`        set variable named "`+KoolTargetDirVar+`" to targetDir`,
		`    end tell`,
		`  end if`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// BuildReuseCurrentSessionScript returns AppleScript that scans session paths like smart-open.
// On match it focuses the tab/session at targetDir without cd; on miss it creates a window and cds.
func BuildReuseCurrentSessionScript(dirPath string, followUps ...string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set targetDir to "` + escaped + `"`,
		`  set matchingWindow to missing value`,
		`  set matchingTab to missing value`,
		`  set matchingSession to missing value`,
		`  repeat with aWindow in windows`,
		`    if not (is hotkey window of aWindow) then`,
		`      repeat with aTab in tabs of aWindow`,
		`        repeat with aSession in sessions of aTab`,
		`          set sessionPath to ""`,
		`          set koolTargetDir to ""`,
		`          try`,
		`            tell aSession`,
		`              set sessionPath to variable named "path"`,
		`            end tell`,
		`          on error`,
		`          end try`,
		`          try`,
		`            tell aSession`,
		`              set koolTargetDir to variable named "` + KoolTargetDirVar + `"`,
		`            end tell`,
		`          on error`,
		`          end try`,
		`          if sessionPath is targetDir or koolTargetDir is targetDir then`,
		`            set matchingWindow to aWindow`,
		`            set matchingTab to aTab`,
		`            set matchingSession to aSession`,
		`            exit repeat`,
		`          end if`,
		`        end repeat`,
		`        if matchingWindow is not missing value then exit repeat`,
		`      end repeat`,
		`      if matchingWindow is not missing value then exit repeat`,
		`    end if`,
		`  end repeat`,
		`  if matchingWindow is not missing value then`,
		`    select matchingWindow`,
		`    tell matchingWindow`,
		`      select matchingTab`,
		`    end tell`,
		`    tell matchingSession`,
		`      select`,
		`    end tell`,
		`  else`,
		`    set newWindow to (create window with default profile)`,
		`    tell current session of newWindow`,
	}
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`        set variable named "`+KoolTargetDirVar+`" to targetDir`,
		`    end tell`,
		`  end if`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// BuildForceNewWindowScript returns AppleScript that opens a new iTerm2 window,
// cds to dirPath, runs follow-up commands, and sets the koolTargetDir variable.
// Unlike BuildScript and BuildReuseCurrentSessionScript, it skips session scanning
// entirely — always creating a new window.
func BuildForceNewWindowScript(dirPath string, followUps ...string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set targetDir to "` + escaped + `"`,
		`  set newWindow to (create window with default profile)`,
		`  tell current session of newWindow`,
	}
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`      set variable named "`+KoolTargetDirVar+`" to targetDir`,
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
	} else if mode == ModeForceNew {
		script = BuildForceNewWindowScript(target, followUps...)
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