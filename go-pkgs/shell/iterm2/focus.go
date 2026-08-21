package iterm2

import (
	"fmt"
	"strings"
)

// FocusConfig injects Exec for Focus (tests and default osascript).
type FocusConfig struct {
	// Exec runs the AppleScript. When nil, the package default osascript runner is used.
	Exec func(script string) error
}

// BuildFocusScript returns AppleScript that activates iTerm2, selects the
// window matching ref.WindowID, then selects the pane's containing tab by its
// stable session ID when available. It never creates windows or tabs.
func BuildFocusScript(ref SessionRef) string {
	escaped := EscapeCommandForAppleScript(ref.WindowID)
	escapedSession := EscapeCommandForAppleScript(ref.SessionID)
	tabIndex := ref.TabIndex
	if tabIndex < 1 {
		tabIndex = 1
	}
	lines := []string{
		tellHeaderResolved(),
		`  activate`,
		`  set targetWindow to missing value`,
		`  set targetTab to missing value`,
		`  set targetSession to missing value`,
		`  repeat with aWindow in windows`,
		`    try`,
		`      if (id of aWindow as string) is "` + escaped + `" then`,
		`        set targetWindow to aWindow`,
		`        exit repeat`,
		`      end if`,
		`    on error`,
		`    end try`,
		`  end repeat`,
	}
	if ref.SessionID != "" {
		lines = append(lines,
			`  if targetWindow is not missing value then`,
			`    repeat with aTab in tabs of targetWindow`,
			`      repeat with aSession in sessions of aTab`,
			`        try`,
			`          if (id of aSession as string) is "`+escapedSession+`" then`,
			`            set targetTab to aTab`,
			`            set targetSession to aSession`,
			`            exit repeat`,
			`          end if`,
			`        on error`,
			`        end try`,
			`      end repeat`,
			`      if targetSession is not missing value then exit repeat`,
			`    end repeat`,
			`  end if`,
		)
	}
	lines = append(lines,
		`  if targetWindow is not missing value then`,
		`    try`,
		`      select targetWindow`,
		`      if targetTab is not missing value then`,
		`        select targetTab`,
		`      else`,
		fmt.Sprintf(`        select tab %d of targetWindow`, tabIndex),
		`      end if`,
		`      if targetSession is not missing value then`,
		`        select targetSession`,
		`      end if`,
		`    on error`,
		`    end try`,
		`  end if`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// Focus builds the focus AppleScript for ref and runs it via cfg.Exec
// (or the package default osascript runner when cfg/Exec is nil).
// Exec errors propagate unchanged.
func Focus(ref SessionRef, cfg *FocusConfig) error {
	script := BuildFocusScript(ref)
	execFn := defaultOsascript
	if cfg != nil && cfg.Exec != nil {
		execFn = cfg.Exec
	}
	return execFn(script)
}
