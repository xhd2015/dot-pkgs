package iterm2

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const envOsascriptStdout = "KOOL_ITERM2_OSASCRIPT_STDOUT"

// TitleTarget selects whether title operations apply to the session/tab or its window.
type TitleTarget int

const (
	// TitleTargetSession operates on the session/tab name (default).
	TitleTargetSession TitleTarget = iota
	// TitleTargetWindow operates on the containing window name.
	TitleTargetWindow
)

var (
	// ErrNotInSession is returned when ITERM_SESSION_ID is empty or missing.
	ErrNotInSession = errors.New("iterm2: not inside an iTerm2 session (ITERM_SESSION_ID empty)")
	// ErrEmptyTitle is returned when SetTitle is called with an empty title.
	ErrEmptyTitle = errors.New("iterm2: title must not be empty")
)

// SessionUUID extracts the UUID portion of an ITERM_SESSION_ID (text after the last ':').
// If no colon is present, the full id is returned.
func SessionUUID(sessionID string) string {
	if i := strings.LastIndex(sessionID, ":"); i >= 0 && i+1 < len(sessionID) {
		return sessionID[i+1:]
	}
	return sessionID
}

// currentSessionID returns ITERM_SESSION_ID or an empty string when unset.
func currentSessionID() string {
	return strings.TrimSpace(os.Getenv("ITERM_SESSION_ID"))
}

// BuildGetTitleScript returns AppleScript that locates the session by UUID and
// returns the session or containing window name.
func BuildGetTitleScript(sessionID string, target TitleTarget) string {
	uuid := SessionUUID(sessionID)
	var returnLine string
	switch target {
	case TitleTargetWindow:
		returnLine = `          return name of aWindow`
	default:
		returnLine = `          return name of aSession`
	}
	lines := []string{
		`tell application "iTerm2"`,
		`  repeat with aWindow in windows`,
		`    repeat with aTab in tabs of aWindow`,
		`      repeat with aSession in sessions of aTab`,
		`        try`,
		`          if id of aSession contains "` + EscapePathForAppleScript(uuid) + `" then`,
		returnLine,
		`          end if`,
		`        end try`,
		`      end repeat`,
		`    end repeat`,
		`  end repeat`,
		`  error "session not found: ` + EscapePathForAppleScript(uuid) + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// BuildSetTitleScript returns AppleScript that locates the session by UUID and
// sets the session or containing window name to title.
func BuildSetTitleScript(sessionID string, target TitleTarget, title string) string {
	uuid := SessionUUID(sessionID)
	escapedTitle := EscapePathForAppleScript(title)
	var setLine string
	switch target {
	case TitleTargetWindow:
		setLine = `          set name of aWindow to "` + escapedTitle + `"`
	default:
		setLine = `          set name of aSession to "` + escapedTitle + `"`
	}
	lines := []string{
		`tell application "iTerm2"`,
		`  repeat with aWindow in windows`,
		`    repeat with aTab in tabs of aWindow`,
		`      repeat with aSession in sessions of aTab`,
		`        try`,
		`          if id of aSession contains "` + EscapePathForAppleScript(uuid) + `" then`,
		setLine,
		`          return`,
		`          end if`,
		`        end try`,
		`      end repeat`,
		`    end repeat`,
		`  end repeat`,
		`  error "session not found: ` + EscapePathForAppleScript(uuid) + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// GetTitle returns the current session or window title for the iTerm2 session
// identified by ITERM_SESSION_ID.
func GetTitle(target TitleTarget) (string, error) {
	sid := currentSessionID()
	if sid == "" {
		return "", ErrNotInSession
	}
	script := BuildGetTitleScript(sid, target)
	out, err := runOsascriptOutput(script)
	if err != nil {
		return "", err
	}
	// osascript typically appends a trailing newline; strip one for callers.
	return strings.TrimRight(out, "\r\n"), nil
}

// SetTitle sets the session or window title and returns the previous and new titles.
func SetTitle(title string, target TitleTarget) (old, newTitle string, err error) {
	if title == "" {
		return "", "", ErrEmptyTitle
	}
	sid := currentSessionID()
	if sid == "" {
		return "", "", ErrNotInSession
	}

	old, err = GetTitle(target)
	if err != nil {
		return "", "", err
	}

	script := BuildSetTitleScript(sid, target, title)
	if _, err := runOsascriptOutput(script); err != nil {
		return "", "", err
	}
	return old, title, nil
}

// runOsascriptOutput runs AppleScript and returns stdout. When KOOL_ITERM2_SCRIPT_OUT
// is set (test hook), the script is written to that path and stdout is mocked from
// KOOL_ITERM2_OSASCRIPT_STDOUT (with optional forced exit via KOOL_ITERM2_OSASCRIPT_EXIT).
func runOsascriptOutput(script string) (string, error) {
	if outPath := os.Getenv(envScriptOut); outPath != "" {
		if err := os.WriteFile(outPath, []byte(script), 0644); err != nil {
			return "", fmt.Errorf("write script out: %w", err)
		}
		if code, ok := osascriptExitFromEnv(); ok && code != 0 {
			return "", fmt.Errorf("osascript exited with status %d", code)
		}
		// Mock stdout for get / old-title paths under tests.
		return os.Getenv(envOsascriptStdout), nil
	}

	cmd := exec.Command("osascript", "-e", script)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if code, ok := osascriptExitFromEnv(); ok && code != 0 {
		return "", fmt.Errorf("osascript exited with status %d", code)
	}
	return stdout.String(), nil
}
