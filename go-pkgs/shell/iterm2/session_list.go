package iterm2

import (
	"strconv"
	"strings"
	"time"
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
//
// Iteration uses count snapshots and per-item try/on error so mid-scan
// window/tab/session churn (AppleScript -1719 Invalid index) skips the
// missing item instead of aborting the whole dump.
func BuildSessionListScript() string {
	sep := fieldSepAS
	lines := []string{
		tellHeaderResolved(),
		`  set fieldSep to ` + sep,
		`  set outLines to {}`,
		`  set windowCount to 0`,
		`  try`,
		`    set windowCount to count of windows`,
		`  on error`,
		`  end try`,
		`  repeat with wi from 1 to windowCount`,
		`    try`,
		`      set aWindow to window wi`,
		`      set windowID to ""`,
		`      set windowName to ""`,
		`      try`,
		`        set windowID to id of aWindow as string`,
		`      on error`,
		`      end try`,
		`      try`,
		`        set windowName to name of aWindow as string`,
		`      on error`,
		`      end try`,
		`      set tabCount to 0`,
		`      try`,
		`        set tabCount to count of tabs of aWindow`,
		`      on error`,
		`      end try`,
		`      repeat with ti from 1 to tabCount`,
		`        try`,
		`          set aTab to tab ti of aWindow`,
		`          set sessionCount to 0`,
		`          try`,
		`            set sessionCount to count of sessions of aTab`,
		`          on error`,
		`          end try`,
		`          repeat with si from 1 to sessionCount`,
		`            try`,
		`              set aSession to session si of aTab`,
		`              set sessionID to ""`,
		`              set sessionTTY to ""`,
		`              set sessionName to ""`,
		`              try`,
		`                set sessionID to id of aSession as string`,
		`              on error`,
		`              end try`,
		`              try`,
		`                set sessionTTY to tty of aSession`,
		`              on error`,
		`              end try`,
		`              try`,
		`                set sessionName to name of aSession as string`,
		`              on error`,
		`              end try`,
		`              set end of outLines to windowID & fieldSep & windowName & fieldSep & (ti as string) & fieldSep & (si as string) & fieldSep & sessionID & fieldSep & sessionTTY & fieldSep & sessionName`,
		`            on error`,
		`            end try`,
		`          end repeat`,
		`        on error`,
		`        end try`,
		`      end repeat`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`  set AppleScript's text item delimiters to linefeed`,
		`  set joined to outLines as text`,
		`  set AppleScript's text item delimiters to ""`,
		`  return joined`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// BuildSessionsInWindowByUUIDScript returns AppleScript that finds the session
// whose id contains sessionUUID, then dumps every session in that window only
// (same TSV shape as BuildSessionListScript). Soft-skips invalid indexes.
func BuildSessionsInWindowByUUIDScript(sessionID string) string {
	uuid := EscapePathForAppleScript(SessionUUID(sessionID))
	sep := fieldSepAS
	lines := []string{
		tellHeaderResolved(),
		`  set targetUUID to "` + uuid + `"`,
		`  set fieldSep to ` + sep,
		`  set outLines to {}`,
		`  set windowCount to 0`,
		`  try`,
		`    set windowCount to count of windows`,
		`  on error`,
		`  end try`,
		`  repeat with wi from 1 to windowCount`,
		`    try`,
		`      set aWindow to window wi`,
		`      set windowID to ""`,
		`      set windowName to ""`,
		`      try`,
		`        set windowID to id of aWindow as string`,
		`      on error`,
		`      end try`,
		`      try`,
		`        set windowName to name of aWindow as string`,
		`      on error`,
		`      end try`,
		`      set tabCount to 0`,
		`      try`,
		`        set tabCount to count of tabs of aWindow`,
		`      on error`,
		`      end try`,
		`      set foundInWindow to false`,
		`      repeat with ti from 1 to tabCount`,
		`        try`,
		`          set aTab to tab ti of aWindow`,
		`          set sessionCount to 0`,
		`          try`,
		`            set sessionCount to count of sessions of aTab`,
		`          on error`,
		`          end try`,
		`          repeat with si from 1 to sessionCount`,
		`            try`,
		`              set aSession to session si of aTab`,
		`              set sessionID to ""`,
		`              try`,
		`                set sessionID to id of aSession as string`,
		`              on error`,
		`              end try`,
		`              if sessionID contains targetUUID then`,
		`                set foundInWindow to true`,
		`                exit repeat`,
		`              end if`,
		`            on error`,
		`            end try`,
		`          end repeat`,
		`          if foundInWindow then exit repeat`,
		`        on error`,
		`        end try`,
		`      end repeat`,
		`      if foundInWindow then`,
		`        set tabCount to 0`,
		`        try`,
		`          set tabCount to count of tabs of aWindow`,
		`        on error`,
		`        end try`,
		`        repeat with ti from 1 to tabCount`,
		`          try`,
		`            set aTab to tab ti of aWindow`,
		`            set sessionCount to 0`,
		`            try`,
		`              set sessionCount to count of sessions of aTab`,
		`            on error`,
		`            end try`,
		`            repeat with si from 1 to sessionCount`,
		`              try`,
		`                set aSession to session si of aTab`,
		`                set sessionID to ""`,
		`                set sessionTTY to ""`,
		`                set sessionName to ""`,
		`                try`,
		`                  set sessionID to id of aSession as string`,
		`                on error`,
		`                end try`,
		`                try`,
		`                  set sessionTTY to tty of aSession`,
		`                on error`,
		`                end try`,
		`                try`,
		`                  set sessionName to name of aSession as string`,
		`                on error`,
		`                end try`,
		`                set end of outLines to windowID & fieldSep & windowName & fieldSep & (ti as string) & fieldSep & (si as string) & fieldSep & sessionID & fieldSep & sessionTTY & fieldSep & sessionName`,
		`              on error`,
		`              end try`,
		`            end repeat`,
		`          on error`,
		`          end try`,
		`        end repeat`,
		`        exit repeat`,
		`      end if`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`  set AppleScript's text item delimiters to linefeed`,
		`  set joined to outLines as text`,
		`  set AppleScript's text item delimiters to ""`,
		`  return joined`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// appleScriptInvalidIndex reports whether err looks like AppleScript -1719
// (collection mutated mid-scan: "Invalid index").
func appleScriptInvalidIndex(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid index") || strings.Contains(msg, "(-1719)") || strings.Contains(msg, "-1719")
}

// runSessionListScript runs script via osascript and parses TSV, retrying once
// when AppleScript reports Invalid index mid-scan.
func runSessionListScript(script string) ([]SessionRef, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(50 * time.Millisecond)
		}
		out, err := runOsascriptOutput(script)
		if err != nil {
			lastErr = err
			if appleScriptInvalidIndex(err) {
				continue
			}
			return nil, err
		}
		return ParseSessionListOutput(out)
	}
	return nil, lastErr
}

// ListSessionsInWindowByUUID dumps sessions for the window that hosts sessionID.
func ListSessionsInWindowByUUID(sessionID string) ([]SessionRef, error) {
	uuid := strings.TrimSpace(SessionUUID(sessionID))
	if uuid == "" {
		return []SessionRef{}, nil
	}
	return runSessionListScript(BuildSessionsInWindowByUUIDScript(sessionID))
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

// IsFullSessionUUID reports whether id is a full iTerm session unique ID
// (8-4-4-4-12 hex), after stripping an ITERM_SESSION_ID "w…t…p…:" prefix.
func IsFullSessionUUID(id string) bool {
	id = strings.TrimSpace(SessionUUID(id))
	if len(id) != 36 {
		return false
	}
	for i := 0; i < 36; i++ {
		c := id[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHexByte(c) {
				return false
			}
		}
	}
	return true
}

func isHexByte(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// FindSessionRefsByRef matches refs by full/prefix session UUID (≥8 chars) or
// TTY (ttysN or /dev/ttysN). Case-insensitive for IDs. Empty ref → empty.
func FindSessionRefsByRef(refs []SessionRef, ref string) []SessionRef {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return []SessionRef{}
	}
	refLower := strings.ToLower(libSessionUUIDLower(ref))
	refTTY := NormalizeTTY(ref)

	var out []SessionRef
	seen := map[string]struct{}{}
	add := func(r SessionRef) {
		key := strings.ToLower(strings.TrimSpace(r.SessionID))
		if key == "" {
			key = NormalizeTTY(r.TTY) + "|" + r.WindowID + "|" + strconv.Itoa(r.TabIndex)
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}

	for _, r := range refs {
		sid := strings.ToLower(strings.TrimSpace(SessionUUID(r.SessionID)))
		if sid != "" && refLower != "" {
			if sid == refLower || (len(refLower) >= 8 && strings.HasPrefix(sid, refLower)) {
				add(r)
				continue
			}
		}
		if refTTY != "" && NormalizeTTY(r.TTY) == refTTY {
			add(r)
		}
	}
	if out == nil {
		return []SessionRef{}
	}
	return out
}

func libSessionUUIDLower(ref string) string {
	return strings.ToLower(strings.TrimSpace(SessionUUID(ref)))
}
