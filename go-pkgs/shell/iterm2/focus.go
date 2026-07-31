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
// window matching ref.WindowID, and selects the 1-based tab at ref.TabIndex.
// It does not create windows or tabs.
func BuildFocusScript(ref SessionRef) string {
	escaped := EscapeCommandForAppleScript(ref.WindowID)
	tabIndex := ref.TabIndex
	if tabIndex < 1 {
		tabIndex = 1
	}
	return strings.Join([]string{
		`tell application "iTerm2"`,
		`  activate`,
		`  set targetWindow to missing value`,
		`  repeat with aWindow in windows`,
		`    try`,
		`      if (id of aWindow as string) is "` + escaped + `" then`,
		`        set targetWindow to aWindow`,
		`        exit repeat`,
		`      end if`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`  if targetWindow is not missing value then`,
		`    select targetWindow`,
		`    try`,
		fmt.Sprintf(`      select tab %d of targetWindow`, tabIndex),
		`    on error`,
		`    end try`,
		`  end if`,
		`end tell`,
	}, "\n")
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
