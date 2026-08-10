package iterm2

import (
	"strings"
)

// TabSessionRef is one session discovered by a tab-set find scan.
type TabSessionRef struct {
	SetName   string
	TabID     string
	WindowID  string
	SessionID string
	TTY       string
}

// fieldSepAS is an ASCII TAB (character id 9) for line fields in find output.
// Do NOT use bare AppleScript `tab` inside `tell application "iTerm2"` — there
// `tab` refers to an iTerm tab element and stringifies as the word "tab", which
// breaks ParseTabSetFindOutput (fields never split) and makes smart re-run create
// duplicate tabs in the current window.
const fieldSepAS = `ASCII character 9`

// BuildTabSetFindScript returns AppleScript that scans iTerm2 sessions for
// user.koolTabSet / user.koolTabSetTab markers matching setName and emits a
// line-oriented dump suitable for ParseTabSetFindOutput:
//
//	SetName\tTabID\tWindowID\tSessionID\tTTY
func BuildTabSetFindScript(setName string) string {
	escaped := EscapeCommandForAppleScript(setName)
	// fieldSep must not be bare `tab` (iTerm class name in this tell block).
	sep := fieldSepAS
	lines := []string{
		tellHeaderResolved(),
		`  set targetSet to "` + escaped + `"`,
		`  set fieldSep to ` + sep,
		`  set outLines to {}`,
		`  repeat with aWindow in windows`,
		`    set windowID to ""`,
		`    try`,
		`      set windowID to id of aWindow as string`,
		`    on error`,
		`    end try`,
		`    repeat with aTab in tabs of aWindow`,
		`      repeat with aSession in sessions of aTab`,
		`        set setName to ""`,
		`        set tabID to ""`,
		`        set sessionID to ""`,
		`        set sessionTTY to ""`,
		`        try`,
		`          tell aSession`,
		`            set setName to variable named "` + TabSetVar + `"`,
		`            set tabID to variable named "` + TabSetTabVar + `"`,
		`          end tell`,
		`        on error`,
		`        end try`,
		`        if setName is targetSet and setName is not "" then`,
		`          try`,
		`            set sessionID to id of aSession as string`,
		`          on error`,
		`          end try`,
		`          try`,
		`            set sessionTTY to tty of aSession`,
		`          on error`,
		`          end try`,
		`          set end of outLines to setName & fieldSep & tabID & fieldSep & windowID & fieldSep & sessionID & fieldSep & sessionTTY`,
		`        end if`,
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

// ParseTabSetFindOutput parses osascript stdout from BuildTabSetFindScript into
// session refs. Format (tab-separated, one session per line):
//
//	SetName\tTabID\tWindowID\tSessionID\tTTY
//
// Blank lines and lines starting with # are ignored. Empty/whitespace-only
// input yields an empty slice and a nil error. Trailing fields may be omitted.
func ParseTabSetFindOutput(stdout string) ([]TabSessionRef, error) {
	if strings.TrimSpace(stdout) == "" {
		return nil, nil
	}
	var refs []TabSessionRef
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Keep empty fields: split on tab without collapsing.
		parts := strings.Split(line, "\t")
		ref := TabSessionRef{}
		if len(parts) > 0 {
			ref.SetName = parts[0]
		}
		if len(parts) > 1 {
			ref.TabID = parts[1]
		}
		if len(parts) > 2 {
			ref.WindowID = parts[2]
		}
		if len(parts) > 3 {
			ref.SessionID = parts[3]
		}
		if len(parts) > 4 {
			ref.TTY = parts[4]
		}
		// Require at least a set name field to count as a record.
		if ref.SetName == "" && ref.TabID == "" {
			continue
		}
		refs = append(refs, ref)
	}
	if refs == nil {
		return []TabSessionRef{}, nil
	}
	return refs, nil
}
