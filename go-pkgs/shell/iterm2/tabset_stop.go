package iterm2

import (
	"fmt"
	"strings"
)

// TabSetStopResult reports how many windows/tabs StopTabSet closed.
type TabSetStopResult struct {
	ClosedWindows int
	ClosedTabs    int
	Warning       string
}

// StopTabSet closes windows (or tabs) that host sessions for setName.
// When no sessions are found, returns a not-running warning and zero closes.
func StopTabSet(setName string, cfg *TabSetConfig) (*TabSetStopResult, error) {
	cfg = normalizeTabSetConfig(cfg)
	refs, err := cfg.Find(setName)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return &TabSetStopResult{
			Warning: fmt.Sprintf("tab set %q is not running", setName),
		}, nil
	}

	windowIDs := uniqueWindowIDs(refs)
	// Prefer closing whole windows when WindowIDs are known and non-empty.
	nonEmptyWindows := make([]string, 0, len(windowIDs))
	for _, id := range windowIDs {
		if id != "" {
			nonEmptyWindows = append(nonEmptyWindows, id)
		}
	}

	result := &TabSetStopResult{}
	if len(nonEmptyWindows) > 0 {
		for _, wid := range nonEmptyWindows {
			script := buildCloseWindowScript(wid)
			if err := cfg.Exec(script); err != nil {
				return nil, err
			}
			result.ClosedWindows++
		}
		return result, nil
	}

	// Fall back: close individual sessions/tabs.
	for _, ref := range refs {
		script := buildCloseSessionScript(ref)
		if err := cfg.Exec(script); err != nil {
			return nil, err
		}
		result.ClosedTabs++
	}
	return result, nil
}

func buildCloseWindowScript(windowID string) string {
	escaped := EscapeCommandForAppleScript(windowID)
	lines := []string{
		`tell application "iTerm2"`,
		`  repeat with aWindow in windows`,
		`    try`,
		fmt.Sprintf(`      if (id of aWindow as string) is "%s" then`, escaped),
		`        close aWindow`,
		`        exit repeat`,
		`      end if`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

func buildCloseSessionScript(ref TabSessionRef) string {
	lines := []string{
		`tell application "iTerm2"`,
	}
	if ref.SessionID != "" {
		escaped := EscapeCommandForAppleScript(ref.SessionID)
		lines = append(lines,
			`  repeat with aWindow in windows`,
			`    repeat with aTab in tabs of aWindow`,
			`      repeat with aSession in sessions of aTab`,
			`        try`,
			fmt.Sprintf(`          if (id of aSession as string) is "%s" or (id of aSession as string) contains "%s" then`, escaped, escaped),
			`            close aTab`,
			`            return`,
			`          end if`,
			`        on error`,
			`        end try`,
			`      end repeat`,
			`    end repeat`,
			`  end repeat`,
		)
	} else {
		lines = append(lines, `  close current session of current window`)
	}
	lines = append(lines, `end tell`)
	return strings.Join(lines, "\n")
}
