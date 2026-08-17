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

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

// AppPath is the system-wide default install path (/Applications/iTerm.app).
// Prefer ResolveAppPath() for detection: it also checks ~/Applications first
// and honors ITERM2_APP_PATH.
const AppPath = "/Applications/iTerm.app"

// EnvITerm2AppPath is the env var for an explicit iTerm2.app bundle override.
// When set and usable, ResolveAppPath returns it; when set but unusable,
// resolve returns "" with no fallthrough to home/system installs.
const EnvITerm2AppPath = "ITERM2_APP_PATH"

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
	// SafeInputIgnore prefixes each session write text with Ctrl-U (ASCII 21)
	// so leftover keystrokes stolen by a newly focused window are discarded.
	SafeInputIgnore bool
}

// ResolveAppPathOpts injects env/home/IsApp for parallel-safe resolve tests.
// Zero value uses production defaults (os.Getenv, os.UserHomeDir, Stat+IsDir).
type ResolveAppPathOpts struct {
	// Getenv reads env; nil => os.Getenv. Tests inject a closure.
	Getenv func(key string) string
	// Home returns user home for ~/Applications candidate; nil => os.UserHomeDir.
	// Empty home skips the home candidate.
	Home func() string
	// IsApp reports whether path is a usable iTerm.app bundle directory.
	// nil => os.Stat + IsDir (same idea as resolveAppPathAmong).
	IsApp func(path string) bool
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

// defaultIsApp reports whether path is an existing directory (usable .app bundle).
func defaultIsApp(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// resolveAppPathAmong returns the first existing .app bundle path among candidates, or "".
func resolveAppPathAmong(candidates []string) string {
	for _, p := range candidates {
		if defaultIsApp(p) {
			return p
		}
	}
	return ""
}

// ResolveAppPathWith returns the preferred existing iTerm2.app path using opts.
//
// Order:
//  1. ITERM2_APP_PATH if set (non-empty after trim) and IsApp → that path
//  2. ITERM2_APP_PATH set but unusable → "" (no fallthrough)
//  3. ~/Applications/iTerm.app if present
//  4. /Applications/iTerm.app if present
//  5. ""
func ResolveAppPathWith(opts ResolveAppPathOpts) string {
	getenv := opts.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	isApp := opts.IsApp
	if isApp == nil {
		isApp = defaultIsApp
	}

	if envPath := strings.TrimSpace(getenv(EnvITerm2AppPath)); envPath != "" {
		if isApp(envPath) {
			return envPath
		}
		// Strict: intentional override that is unusable must not fall through.
		return ""
	}

	var home string
	if opts.Home != nil {
		home = opts.Home()
	} else if h, err := os.UserHomeDir(); err == nil {
		home = h
	}
	if home != "" {
		homeApp := filepath.Join(home, "Applications", "iTerm.app")
		if isApp(homeApp) {
			return homeApp
		}
	}
	if isApp(AppPath) {
		return AppPath
	}
	return ""
}

// ResolveAppPath returns the preferred existing iTerm2.app path using production
// defaults (real env, home, and filesystem). Equivalent to ResolveAppPathWith(zero opts).
func ResolveAppPath() string {
	return ResolveAppPathWith(ResolveAppPathOpts{})
}

// IsInstalled reports whether ResolveAppPath finds a usable iTerm2.app bundle.
func IsInstalled() bool {
	return ResolveAppPath() != ""
}

// TellApplicationHeader returns the AppleScript tell line for appPath.
//
// Non-empty path → path-bound string literal:
//
//	tell application "/path/to/iTerm.app"
//
// Empty path → bare name fallback:
//
//	tell application "iTerm2"
//
// The bare "iTerm2" fallback is intentional, not a bug. Callers that prefer a
// concrete install use ResolveAppPath() first (home then /Applications) and pass
// that path. When no path is known, Launch Services name resolution is the
// correct last resort — do not remove the bare name or "fix" it to always
// error without an alternate product decision.
//
// Use a string-literal path (not `POSIX file "…" as text`). A runtime
// expression target prevents AppleScript from loading iTerm's dictionary at
// compile time, so iTerm terms like `create window with default profile` fail
// with "Expected "," but found class name" (-2741).
func TellApplicationHeader(appPath string) string {
	if appPath == "" {
		// INTENTIONAL: bare LS name when appPath is empty (see comment above).
		return `tell application "iTerm2"`
	}
	return `tell application "` + EscapePathForAppleScript(appPath) + `"`
}

// tellHeaderResolved is TellApplicationHeader(ResolveAppPath()) for package scripts.
func tellHeaderResolved() string {
	return TellApplicationHeader(ResolveAppPath())
}

// EscapePathForAppleScript escapes a path embedded in an AppleScript string literal.
// Delegates to applescript.EscapeString (backslash and double-quote).
func EscapePathForAppleScript(dirPath string) string {
	return applescript.EscapeString(dirPath)
}

// EscapeCommandForAppleScript escapes a command for AppleScript write text literals.
// Delegates to applescript.EscapeString. For long FollowUp bodies, see
// applescript.CheckWriteText and DocumentWriteTextLimitation.
func EscapeCommandForAppleScript(command string) string {
	return applescript.EscapeString(command)
}

func buildSessionCommandLines(followUpCommands []string, safeInputIgnore bool) []string {
	cd := `"cd " & quoted form of targetDir`
	if safeInputIgnore {
		cd = `(ASCII character 21) & ` + cd
	}
	lines := []string{
		`        write text (` + cd + `)`,
	}
	for _, command := range followUpCommands {
		escaped := EscapeCommandForAppleScript(command)
		if safeInputIgnore {
			lines = append(lines, fmt.Sprintf(`        write text ((ASCII character 21) & "%s")`, escaped))
		} else {
			lines = append(lines, fmt.Sprintf(`        write text "%s"`, escaped))
		}
	}
	return lines
}

// BuildPathScanSmokeScriptApp returns path-scan smoke AppleScript targeting appPath.
func BuildPathScanSmokeScriptApp(appPath string) string {
	lines := []string{
		TellApplicationHeader(appPath),
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

// BuildPathScanSmokeScript returns AppleScript that probes session path variables.
func BuildPathScanSmokeScript() string {
	return BuildPathScanSmokeScriptApp(ResolveAppPath())
}

// BuildScriptApp returns smart-open AppleScript targeting appPath and cds to dirPath.
func BuildScriptApp(appPath, dirPath string, followUps ...string) string {
	return buildScriptApp(appPath, dirPath, false, followUps)
}

func buildScriptApp(appPath, dirPath string, safeInputIgnore bool, followUps []string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps, safeInputIgnore)
	lines := []string{
		TellApplicationHeader(appPath),
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

// BuildScript returns AppleScript that smart-opens iTerm2 and cds to dirPath.
func BuildScript(dirPath string, followUps ...string) string {
	return buildScriptApp(ResolveAppPath(), dirPath, false, followUps)
}

// BuildReuseCurrentSessionScript returns AppleScript that scans session paths like smart-open.
// On match it focuses the tab/session at targetDir without cd; on miss it creates a window and cds.
func BuildReuseCurrentSessionScript(dirPath string, followUps ...string) string {
	return buildReuseCurrentSessionScript(dirPath, false, followUps)
}

func buildReuseCurrentSessionScript(dirPath string, safeInputIgnore bool, followUps []string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps, safeInputIgnore)
	lines := []string{
		tellHeaderResolved(),
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

// BuildForceNewWindowScriptApp returns force-new-window AppleScript targeting appPath.
func BuildForceNewWindowScriptApp(appPath, dirPath string, followUps ...string) string {
	return buildForceNewWindowScriptApp(appPath, dirPath, false, followUps)
}

func buildForceNewWindowScriptApp(appPath, dirPath string, safeInputIgnore bool, followUps []string) string {
	escaped := EscapePathForAppleScript(dirPath)
	sessionCommandLines := buildSessionCommandLines(followUps, safeInputIgnore)
	lines := []string{
		TellApplicationHeader(appPath),
		`  set targetDir to "` + escaped + `"`,
		`  set newWindow to (create window with default profile)`,
		`  tell current session of newWindow`,
	}
	lines = append(lines, sessionCommandLines...)
	lines = append(lines,
		`      set variable named "`+KoolTargetDirVar+`" to targetDir`,
		`  end tell`,
		`  activate`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// BuildForceNewWindowScript returns AppleScript that opens a new iTerm2 window,
// cds to dirPath, runs follow-up commands, and sets the koolTargetDir variable.
// Unlike BuildScript and BuildReuseCurrentSessionScript, it skips session scanning
// entirely — always creating a new window. Create happens before activate so the
// window lands on the current Space instead of following an existing iTerm window.
func BuildForceNewWindowScript(dirPath string, followUps ...string) string {
	return buildForceNewWindowScriptApp(ResolveAppPath(), dirPath, false, followUps)
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
	var safeInputIgnore bool
	if cfg != nil {
		if len(cfg.FollowUpCommands) > 0 {
			followUps = cfg.FollowUpCommands
		}
		mode = cfg.Mode
		safeInputIgnore = cfg.SafeInputIgnore
	}
	var script string
	if mode == ModeReuseCurrent {
		script = buildReuseCurrentSessionScript(target, safeInputIgnore, followUps)
	} else if mode == ModeForceNew {
		script = buildForceNewWindowScriptApp(ResolveAppPath(), target, safeInputIgnore, followUps)
	} else {
		script = buildScriptApp(ResolveAppPath(), target, safeInputIgnore, followUps)
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