package iterm2

import (
	"fmt"
	"strings"
)

// Session user-variable names stamped on every tab-set session.
const (
	TabSetVar    = "user.koolTabSet"
	TabSetTabVar = "user.koolTabSetTab"
)

// TabSpec describes one tab in a tab set.
type TabSpec struct {
	ID       string
	Name     string
	Command  string
	Cwd      string
	NoSubmit bool // when true, write text without newline (stage command; user submits)
}

// writeTextCommand returns an AppleScript write-text line for cmd.
// When noSubmit is true, appends " without newline" so Enter is not sent.
func writeTextCommand(cmd string, noSubmit bool) string {
	line := fmt.Sprintf(`write text "%s"`, EscapeCommandForAppleScript(cmd))
	if noSubmit {
		return line + ` without newline`
	}
	return line
}

// TabSetSpec describes a named set of tabs opened in one new iTerm2 window.
type TabSetSpec struct {
	Name       string
	WindowName string
	Tabs       []TabSpec
}

// BuildTabSetNewWindowScript returns AppleScript that opens one new window with
// N tabs (first = initial session; N−1 create tab), stamps set/tab markers,
// sets optional names, optional per-tab cwd, and runs each tab's command.
func BuildTabSetNewWindowScript(spec TabSetSpec) string {
	setName := EscapeCommandForAppleScript(spec.Name)
	lines := []string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set newWindow to (create window with default profile)`,
	}

	for i, tab := range spec.Tabs {
		if i == 0 {
			lines = append(lines, `  tell current session of newWindow`)
		} else {
			lines = append(lines,
				`  tell newWindow`,
				`    set newTab to (create tab with default profile)`,
				`    tell current session of newTab`,
			)
		}

		// Stamp set + per-tab markers on the session.
		lines = append(lines,
			fmt.Sprintf(`    set variable named "%s" to "%s"`, TabSetVar, setName),
			fmt.Sprintf(`    set variable named "%s" to "%s"`, TabSetTabVar, EscapeCommandForAppleScript(tab.ID)),
		)

		if tab.Name != "" {
			// set name of the current session (tests look for "set name" + "session")
			lines = append(lines,
				fmt.Sprintf(`    set name to "%s"`, EscapeCommandForAppleScript(tab.Name)),
			)
		}

		if tab.Cwd != "" {
			// Per-tab cwd only — do not inject Open-style shared targetDir cd.
			escapedCwd := EscapePathForAppleScript(tab.Cwd)
			lines = append(lines,
				fmt.Sprintf(`    write text ("cd " & quoted form of "%s")`, escapedCwd),
			)
		}

		if tab.Command != "" {
			lines = append(lines,
				fmt.Sprintf(`    %s`, writeTextCommand(tab.Command, tab.NoSubmit)),
			)
		}

		if i == 0 {
			lines = append(lines, `  end tell`)
		} else {
			lines = append(lines,
				`    end tell`,
				`  end tell`,
			)
		}
	}

	if spec.WindowName != "" {
		lines = append(lines,
			fmt.Sprintf(`  set name of newWindow to "%s"`, EscapeCommandForAppleScript(spec.WindowName)),
		)
	}

	lines = append(lines, `end tell`)
	return strings.Join(lines, "\n")
}
