package iterm2

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Canonical iTerm2 install tags for ContentsResult.App (same forms as kool save).
const (
	CanonicalITermAppSystem = "/Applications/iTerm.app"
	CanonicalITermAppHome   = "~/Applications/iTerm.app"
)

// ErrSessionNotFound is returned when no running iTerm2 install has the session.
var ErrSessionNotFound = errors.New("session not found")

// ContentsResult is one pane dump. App is a canonical install tag when known.
type ContentsResult struct {
	SessionID string
	App       string
	Contents  string
}

// ContentsConfig injects Exec / app discovery for tests.
type ContentsConfig struct {
	// Exec runs AppleScript and returns stdout. Nil uses runOsascriptOutput.
	Exec func(script string) (string, error)
	// Apps, when non-nil, is the exact tell list (already running). Empty means none.
	Apps   []ContentsApp
	Getenv func(key string) string
	Home   func() string
	IsApp  func(path string) bool
	// Running reports whether abs bundle is a live process. Nil uses pgrep.
	Running func(abs string) bool
}

// ContentsApp is one iTerm2 install to tell.
type ContentsApp struct {
	// Canonical is ContentsResult.App (home or system tag, or a custom path).
	Canonical string
	// Path is the filesystem path used in tell application "…".
	Path string
}

// BuildContentsScript returns AppleScript that locates sessionID and returns
// contents of that session. It does not activate or select. appPath is a
// filesystem path for TellApplicationHeader (empty → bare "iTerm2").
func BuildContentsScript(sessionID, appPath string) string {
	uuid := SessionUUID(sessionID)
	escaped := EscapePathForAppleScript(uuid)
	lines := []string{
		TellApplicationHeader(appPath),
		`  repeat with aWindow in windows`,
		`    repeat with aTab in tabs of aWindow`,
		`      repeat with aSession in sessions of aTab`,
		`        try`,
		`          if id of aSession contains "` + escaped + `" then`,
		`            return contents of aSession`,
		`          end if`,
		`        end try`,
		`      end repeat`,
		`    end repeat`,
		`  end repeat`,
		`  error "session not found: ` + escaped + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// Contents dumps currently visible text for sessionID from a running iTerm2.
// Search order: ITERM2_APP_PATH (if set and usable), ~/Applications/iTerm.app,
// /Applications/iTerm.app. Skips bundles that are not running (does not launch).
// First UUID hit wins.
func Contents(sessionID string, cfg *ContentsConfig) (ContentsResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ContentsResult{}, fmt.Errorf("session id is required")
	}
	uuid := SessionUUID(sessionID)
	if strings.TrimSpace(uuid) == "" {
		return ContentsResult{}, fmt.Errorf("session id is required")
	}

	execFn := contentsOsascript
	if cfg != nil && cfg.Exec != nil {
		execFn = cfg.Exec
	}

	apps := contentsSearchApps(cfg)
	var lastNotFound error
	for _, app := range apps {
		script := BuildContentsScript(uuid, app.Path)
		out, err := execFn(script)
		if err != nil {
			if isSessionNotFound(err) {
				lastNotFound = err
				continue
			}
			return ContentsResult{}, err
		}
		return ContentsResult{
			SessionID: uuid,
			App:       app.Canonical,
			Contents:  strings.TrimRight(out, "\r\n"),
		}, nil
	}
	if lastNotFound != nil {
		return ContentsResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, uuid)
	}
	return ContentsResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, uuid)
}

func contentsSearchApps(cfg *ContentsConfig) []ContentsApp {
	if cfg != nil && cfg.Apps != nil {
		return cfg.Apps
	}

	getenv := os.Getenv
	homeFn := func() string {
		h, _ := os.UserHomeDir()
		return h
	}
	isApp := defaultIsApp
	running := defaultITermAppRunning
	if cfg != nil {
		if cfg.Getenv != nil {
			getenv = cfg.Getenv
		}
		if cfg.Home != nil {
			homeFn = cfg.Home
		}
		if cfg.IsApp != nil {
			isApp = cfg.IsApp
		}
		if cfg.Running != nil {
			running = cfg.Running
		}
	}

	home := homeFn()
	homeAbs := ""
	if home != "" {
		homeAbs = filepath.Join(home, "Applications", "iTerm.app")
	}

	var raw []ContentsApp
	if envPath := strings.TrimSpace(getenv(EnvITerm2AppPath)); envPath != "" {
		if isApp(envPath) {
			raw = append(raw, ContentsApp{
				Canonical: CanonicalITermApp(envPath, home),
				Path:      envPath,
			})
		}
		// Usable env is extra/first; still consider the two canonical installs.
	}
	if homeAbs != "" && isApp(homeAbs) {
		raw = append(raw, ContentsApp{
			Canonical: CanonicalITermAppHome,
			Path:      homeAbs,
		})
	}
	if isApp(AppPath) {
		raw = append(raw, ContentsApp{
			Canonical: CanonicalITermAppSystem,
			Path:      AppPath,
		})
	}

	seen := map[string]struct{}{}
	out := make([]ContentsApp, 0, len(raw))
	for _, a := range raw {
		key := filepath.Clean(a.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if running != nil && !running(a.Path) {
			continue
		}
		out = append(out, a)
	}
	return out
}

// CanonicalITermApp maps a filesystem path to ~/Applications/iTerm.app or
// /Applications/iTerm.app. Custom paths (ITERM2_APP_PATH) are returned cleaned.
func CanonicalITermApp(path, home string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if path == CanonicalITermAppSystem || path == CanonicalITermAppHome {
		return path
	}
	expanded := path
	if strings.HasPrefix(path, "~/") {
		if home != "" {
			expanded = filepath.Join(home, path[2:])
		}
	}
	expanded = filepath.Clean(expanded)
	if home != "" {
		homeApps := filepath.Join(home, "Applications")
		if expanded == filepath.Join(homeApps, "iTerm.app") ||
			strings.HasPrefix(expanded, homeApps+string(os.PathSeparator)) {
			if strings.EqualFold(filepath.Base(expanded), "iTerm.app") ||
				strings.EqualFold(filepath.Base(expanded), "iTerm2.app") {
				return CanonicalITermAppHome
			}
		}
	}
	if strings.HasPrefix(expanded, "/Applications/") {
		return CanonicalITermAppSystem
	}
	return path
}

func isSessionNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrSessionNotFound) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "session not found")
}

func contentsOsascript(script string) (string, error) {
	if os.Getenv(envScriptOut) != "" {
		return runOsascriptOutput(script)
	}
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if strings.Contains(strings.ToLower(msg), "session not found") {
			return "", fmt.Errorf("%w: %s", ErrSessionNotFound, msg)
		}
		return "", fmt.Errorf("osascript: %s", msg)
	}
	return stdout.String(), nil
}

func defaultITermAppRunning(abs string) bool {
	abs = strings.TrimSpace(abs)
	if abs == "" {
		return false
	}
	if strings.HasPrefix(abs, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			abs = filepath.Join(home, abs[2:])
		}
	}
	bin := filepath.Join(abs, "Contents", "MacOS", "iTerm2")
	if exec.Command("pgrep", "-f", bin).Run() == nil {
		return true
	}
	return exec.Command("pgrep", "-f", abs).Run() == nil
}
