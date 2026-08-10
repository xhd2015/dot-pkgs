package iterm2

import (
	"strconv"
	"strings"
)

// SessionRef is one iTerm2 session from a full session-list scan.
type SessionRef struct {
	WindowID   string // iTerm window id string
	WindowName string // optional
	TabIndex   int    // 1-based tab index in that window
	SessionID  string // iTerm session UUID (optional)
	TTY        string // e.g. /dev/ttys148
	Name       string // session name (optional)
}

// NormalizeTTY maps bare and /dev TTY forms to a comparable non-empty string.
// Empty or whitespace-only input yields "".
//
// Examples: "ttys148" and "/dev/ttys148" normalize equal.
func NormalizeTTY(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "/dev/") {
		return s
	}
	return "/dev/" + s
}

// BuildSessionListScript returns AppleScript that scans every iTerm2
// window/tab/session and emits a line-oriented TSV dump:
//
//	WindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName
//
// Field separator is ASCII character 9 (not bare AppleScript `tab`, which is
// an iTerm element name inside the tell block).
func BuildSessionListScript() string {
	sep := fieldSepAS
	lines := []string{
		tellHeaderResolved(),
		`  set fieldSep to ` + sep,
		`  set outLines to {}`,
		`  repeat with aWindow in windows`,
		`    set windowID to ""`,
		`    set windowName to ""`,
		`    try`,
		`      set windowID to id of aWindow as string`,
		`    on error`,
		`    end try`,
		`    try`,
		`      set windowName to name of aWindow as string`,
		`    on error`,
		`    end try`,
		`    set tabIndex to 0`,
		`    repeat with aTab in tabs of aWindow`,
		`      set tabIndex to tabIndex + 1`,
		`      set sessionIndex to 0`,
		`      repeat with aSession in sessions of aTab`,
		`        set sessionIndex to sessionIndex + 1`,
		`        set sessionID to ""`,
		`        set sessionTTY to ""`,
		`        set sessionName to ""`,
		`        try`,
		`          set sessionID to id of aSession as string`,
		`        on error`,
		`        end try`,
		`        try`,
		`          set sessionTTY to tty of aSession`,
		`        on error`,
		`        end try`,
		`        try`,
		`          set sessionName to name of aSession as string`,
		`        on error`,
		`        end try`,
		`        set end of outLines to windowID & fieldSep & windowName & fieldSep & (tabIndex as string) & fieldSep & (sessionIndex as string) & fieldSep & sessionID & fieldSep & sessionTTY & fieldSep & sessionName`,
		`      end repeat`,
		`    end repeat`,
		`  end repeat`,
		`  set AppleScript's text item delimiters to linefeed`,
		`  set joined to outLines as text`,
		`  set AppleScript's text item delimiters to ""`,
		`  return joined`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// ParseSessionListOutput parses osascript stdout from BuildSessionListScript
// into session refs. Format (tab-separated, one session per line):
//
//	WindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName
//
// Blank lines and lines starting with # are ignored. Empty/whitespace-only
// input yields an empty slice and a nil error. SessionIndex is accepted and
// discarded when filling SessionRef.
func ParseSessionListOutput(stdout string) ([]SessionRef, error) {
	if strings.TrimSpace(stdout) == "" {
		return []SessionRef{}, nil
	}
	var refs []SessionRef
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		// Require at least WindowID-ish content; pure whitespace fields after split of junk.
		if len(parts) == 0 {
			continue
		}
		ref := SessionRef{}
		if len(parts) > 0 {
			ref.WindowID = parts[0]
		}
		if len(parts) > 1 {
			ref.WindowName = parts[1]
		}
		if len(parts) > 2 {
			if n, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
				ref.TabIndex = n
			}
		}
		// parts[3] = SessionIndex — discarded
		if len(parts) > 4 {
			ref.SessionID = parts[4]
		}
		if len(parts) > 5 {
			ref.TTY = parts[5]
		}
		if len(parts) > 6 {
			ref.Name = parts[6]
		}
		// Skip lines that are not real records (e.g. no fields of interest).
		if ref.WindowID == "" && ref.SessionID == "" && ref.TTY == "" {
			continue
		}
		refs = append(refs, ref)
	}
	if refs == nil {
		return []SessionRef{}, nil
	}
	return refs, nil
}

// FindByTTY returns refs whose TTY matches any of the query TTYs after
// NormalizeTTY on both sides. Order follows the input refs (stable). Empty
// queries or no matches yield an empty slice.
func FindByTTY(refs []SessionRef, ttys []string) []SessionRef {
	want := make(map[string]struct{}, len(ttys))
	for _, t := range ttys {
		n := NormalizeTTY(t)
		if n == "" {
			continue
		}
		want[n] = struct{}{}
	}
	if len(want) == 0 {
		return []SessionRef{}
	}
	var out []SessionRef
	for _, ref := range refs {
		n := NormalizeTTY(ref.TTY)
		if n == "" {
			continue
		}
		if _, ok := want[n]; ok {
			out = append(out, ref)
		}
	}
	if out == nil {
		return []SessionRef{}
	}
	return out
}
