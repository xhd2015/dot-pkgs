package iterm2

import (
	"fmt"
	"strconv"
	"strings"
)

// DirectoryFocusCandidate is an existing iTerm2 session that advertises a
// working directory through either its path or user.koolTargetDir variable.
// TabIndex is one-based.
type DirectoryFocusCandidate struct {
	WindowID      string
	WindowTitle   string
	TabIndex      int
	SessionID     string
	Path          string
	KoolTargetDir string
}

// BuildDirectoryFocusFindScript returns AppleScript that lists every existing
// session matching targetDir. It never creates an iTerm window, tab, or session.
func BuildDirectoryFocusFindScript(targetDir string) string {
	target := EscapePathForAppleScript(targetDir)
	lines := []string{
		tellHeaderResolved(),
		`  set targetDir to "` + target + `"`,
		`  set fieldSep to ` + fieldSepAS,
		`  set outLines to {}`,
		`  repeat with aWindow in windows`,
		`    set windowID to ""`,
		`    set windowTitle to ""`,
		`    try`, `      set windowID to id of aWindow as string`, `    end try`,
		`    try`, `      set windowTitle to name of aWindow as string`, `    end try`,
		`    set tabIndex to 0`,
		`    repeat with aTab in tabs of aWindow`,
		`      set tabIndex to tabIndex + 1`,
		`      repeat with aSession in sessions of aTab`,
		`        set sessionPath to ""`, `        set koolTargetDir to ""`, `        set sessionID to ""`,
		`        try`, `          tell aSession`, `            set sessionPath to variable named "path"`, `          end tell`, `        on error`, `        end try`,
		`        try`, `          tell aSession`, `            set koolTargetDir to variable named "` + KoolTargetDirVar + `"`, `          end tell`, `        on error`, `        end try`,
		`        try`, `          set sessionID to id of aSession as string`, `        on error`, `        end try`,
		`        if sessionPath is targetDir or koolTargetDir is targetDir then`,
		`          set end of outLines to windowID & fieldSep & windowTitle & fieldSep & (tabIndex as string) & fieldSep & sessionID & fieldSep & sessionPath & fieldSep & koolTargetDir`,
		`        end if`,
		`      end repeat`, `    end repeat`, `  end repeat`,
		`  set AppleScript's text item delimiters to linefeed`, `  set joined to outLines as text`, `  set AppleScript's text item delimiters to ""`, `  return joined`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// ParseDirectoryFocusFindOutput parses BuildDirectoryFocusFindScript output.
func ParseDirectoryFocusFindOutput(stdout string) ([]DirectoryFocusCandidate, error) {
	if strings.TrimSpace(stdout) == "" {
		return []DirectoryFocusCandidate{}, nil
	}
	var candidates []DirectoryFocusCandidate
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			return nil, fmt.Errorf("iterm2: malformed directory focus result: %q", line)
		}
		tabIndex, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			return nil, fmt.Errorf("iterm2: malformed tab index: %q", parts[2])
		}
		candidate := DirectoryFocusCandidate{WindowID: parts[0], WindowTitle: parts[1], TabIndex: tabIndex, SessionID: parts[3]}
		if len(parts) > 4 {
			candidate.Path = parts[4]
		}
		if len(parts) > 5 {
			candidate.KoolTargetDir = parts[5]
		}
		candidates = append(candidates, candidate)
	}
	if candidates == nil {
		return []DirectoryFocusCandidate{}, nil
	}
	return candidates, nil
}

// FindDirectoryFocusCandidates discovers all existing exact directory matches.
func FindDirectoryFocusCandidates(targetDir string) ([]DirectoryFocusCandidate, error) {
	out, err := runOsascriptOutput(BuildDirectoryFocusFindScript(targetDir))
	if err != nil {
		return nil, err
	}
	return ParseDirectoryFocusFindOutput(out)
}

// FocusDirectoryCandidate brings an existing candidate's window and tab forward.
func FocusDirectoryCandidate(candidate DirectoryFocusCandidate) error {
	return Focus(SessionRef{WindowID: candidate.WindowID, TabIndex: candidate.TabIndex, SessionID: candidate.SessionID}, nil)
}
