package iterm2

import (
	"errors"
	"fmt"
	"strings"
)

// TabSetRunMode selects how RunTabSet places or reuses windows.
type TabSetRunMode int

const (
	// TabSetRunSmart reuses found sessions when present, else creates a new window.
	TabSetRunSmart TabSetRunMode = iota
	// TabSetRunNewWindow always creates a new window.
	TabSetRunNewWindow
	// TabSetRunNoNewWindow syncs into FrontmostWindowID (errors if unset).
	TabSetRunNoNewWindow
)

// ErrNoITermWindow is returned when NoNewWindow mode has no target window.
var ErrNoITermWindow = errors.New("iterm2: no iTerm window available for NoNewWindow mode")

// TabSetConfig injects find/busy/exec for orchestration (and tests).
type TabSetConfig struct {
	Find              func(setName string) ([]TabSessionRef, error)
	Busy              func(ref TabSessionRef) BusyState
	Exec              func(script string) error
	FrontmostWindowID string
}

// TabSetRunOptions controls RunTabSet mode.
type TabSetRunOptions struct {
	Mode TabSetRunMode
}

// TabRunResult is the per-tab outcome of a RunTabSet call.
type TabRunResult struct {
	TabID  string
	Action string // "created" | "resent" | "skipped-busy" | "skipped-unknown"
}

// TabSetRunResult is the aggregate result of RunTabSet.
type TabSetRunResult struct {
	CreatedWindow bool
	FocusedWindow string
	Warning       string
	Tabs          []TabRunResult
}

// RunTabSet orchestrates create / resend / skip for a tab set using cfg hooks.
func RunTabSet(spec TabSetSpec, opts TabSetRunOptions, cfg *TabSetConfig) (*TabSetRunResult, error) {
	cfg = normalizeTabSetConfig(cfg)

	switch opts.Mode {
	case TabSetRunNewWindow:
		return runTabSetCreateWindow(spec, cfg)
	case TabSetRunNoNewWindow:
		if strings.TrimSpace(cfg.FrontmostWindowID) == "" {
			return nil, ErrNoITermWindow
		}
		refs, err := cfg.Find(spec.Name)
		if err != nil {
			return nil, err
		}
		return runTabSetSync(spec, cfg, cfg.FrontmostWindowID, refs, "")
	default: // TabSetRunSmart
		refs, err := cfg.Find(spec.Name)
		if err != nil {
			return nil, err
		}
		if len(refs) == 0 {
			return runTabSetCreateWindow(spec, cfg)
		}
		windowIDs := uniqueWindowIDs(refs)
		warning := ""
		if len(windowIDs) > 1 {
			warning = fmt.Sprintf("tab set %q found in %d windows; syncing most recent only", spec.Name, len(windowIDs))
		}
		focus := windowIDs[0]
		return runTabSetSync(spec, cfg, focus, refs, warning)
	}
}

func runTabSetCreateWindow(spec TabSetSpec, cfg *TabSetConfig) (*TabSetRunResult, error) {
	script := BuildTabSetNewWindowScript(spec)
	if err := cfg.Exec(script); err != nil {
		return nil, err
	}
	result := &TabSetRunResult{CreatedWindow: true}
	for _, tab := range spec.Tabs {
		result.Tabs = append(result.Tabs, TabRunResult{TabID: tab.ID, Action: "created"})
	}
	return result, nil
}

// runTabSetSync reconciles spec.Tabs into windowID using find refs in that window only.
func runTabSetSync(spec TabSetSpec, cfg *TabSetConfig, windowID string, allRefs []TabSessionRef, warning string) (*TabSetRunResult, error) {
	byTab := map[string]TabSessionRef{}
	for _, ref := range allRefs {
		if windowID == "" || ref.WindowID == windowID {
			// First occurrence wins (stable for duplicate TabIDs).
			if _, exists := byTab[ref.TabID]; !exists {
				byTab[ref.TabID] = ref
			}
		}
	}

	result := &TabSetRunResult{
		CreatedWindow: false,
		FocusedWindow: windowID,
		Warning:       warning,
	}

	// Always bring the target window forward on reuse (even if all tabs skipped).
	if windowID != "" {
		if err := cfg.Exec(buildFocusWindowScript(windowID)); err != nil {
			return nil, err
		}
	}

	for _, tab := range spec.Tabs {
		ref, found := byTab[tab.ID]
		if !found {
			script := buildCreateTabInWindowScript(windowID, spec.Name, tab)
			if err := cfg.Exec(script); err != nil {
				return nil, err
			}
			result.Tabs = append(result.Tabs, TabRunResult{TabID: tab.ID, Action: "created"})
			continue
		}

		state := BusyStateUnknown
		if cfg.Busy != nil {
			state = cfg.Busy(ref)
		}
		switch state {
		case BusyStateBusy:
			result.Tabs = append(result.Tabs, TabRunResult{TabID: tab.ID, Action: "skipped-busy"})
		case BusyStateUnknown:
			result.Tabs = append(result.Tabs, TabRunResult{TabID: tab.ID, Action: "skipped-unknown"})
		default: // Idle
			script := buildResendCommandScript(ref, tab)
			if err := cfg.Exec(script); err != nil {
				return nil, err
			}
			result.Tabs = append(result.Tabs, TabRunResult{TabID: tab.ID, Action: "resent"})
		}
	}
	return result, nil
}

// buildFocusWindowScript selects an iTerm2 window by id string and activates iTerm.
func buildFocusWindowScript(windowID string) string {
	escaped := EscapeCommandForAppleScript(windowID)
	return strings.Join([]string{
		tellHeaderResolved(),
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
		`  end if`,
		`end tell`,
	}, "\n")
}

// buildCreateTabInWindowScript creates one tab in an existing window and stamps markers.
// Never sends Ctrl+C.
func buildCreateTabInWindowScript(windowID, setName string, tab TabSpec) string {
	escapedSet := EscapeCommandForAppleScript(setName)
	escapedTabID := EscapeCommandForAppleScript(tab.ID)
	lines := []string{
		tellHeaderResolved(),
		`  activate`,
	}
	if windowID != "" {
		lines = append(lines,
			fmt.Sprintf(`  set targetWindow to missing value`),
			`  repeat with aWindow in windows`,
			fmt.Sprintf(`    try`),
			fmt.Sprintf(`      if (id of aWindow as string) is "%s" then`, EscapeCommandForAppleScript(windowID)),
			`        set targetWindow to aWindow`,
			`        exit repeat`,
			`      end if`,
			`    on error`,
			`    end try`,
			`  end repeat`,
			`  if targetWindow is missing value then`,
			`    set targetWindow to current window`,
			`  end if`,
			`  tell targetWindow`,
		)
	} else {
		lines = append(lines, `  tell current window`)
	}
	lines = append(lines,
		`    set newTab to (create tab with default profile)`,
		`    tell current session of newTab`,
		fmt.Sprintf(`      set variable named "%s" to "%s"`, TabSetVar, escapedSet),
		fmt.Sprintf(`      set variable named "%s" to "%s"`, TabSetTabVar, escapedTabID),
	)
	if tab.Name != "" {
		lines = append(lines, fmt.Sprintf(`      set name to "%s"`, EscapeCommandForAppleScript(tab.Name)))
	}
	if tab.Cwd != "" {
		// cd always executes with newline (NoSubmit does not apply to cwd).
		lines = append(lines, fmt.Sprintf(`      write text ("cd " & quoted form of "%s")`, EscapePathForAppleScript(tab.Cwd)))
	}
	if tab.Command != "" {
		lines = append(lines, fmt.Sprintf(`      %s`, writeTextCommand(tab.Command, tab.NoSubmit)))
	}
	lines = append(lines,
		`    end tell`,
		`  end tell`,
		`end tell`,
	)
	return strings.Join(lines, "\n")
}

// buildResendCommandScript writes the tab command into an existing idle session.
// Never sends Ctrl+C — write text only. Honors TabSpec.NoSubmit.
func buildResendCommandScript(ref TabSessionRef, tab TabSpec) string {
	lines := []string{
		tellHeaderResolved(),
		`  activate`,
	}
	// Prefer session id when available; fall back to TTY scan.
	if ref.SessionID != "" {
		sid := EscapeCommandForAppleScript(ref.SessionID)
		lines = append(lines,
			`  set targetSession to missing value`,
			`  set targetTab to missing value`,
			`  set targetWindow to missing value`,
			`  repeat with aWindow in windows`,
			`    repeat with aTab in tabs of aWindow`,
			`      repeat with aSession in sessions of aTab`,
			`        try`,
			fmt.Sprintf(`          if (id of aSession as string) is "%s" then`, sid),
			`            set targetSession to aSession`,
			`            set targetTab to aTab`,
			`            set targetWindow to aWindow`,
			`            exit repeat`,
			`          end if`,
			`        on error`,
			`        end try`,
			`      end repeat`,
			`      if targetSession is not missing value then exit repeat`,
			`    end repeat`,
			`    if targetSession is not missing value then exit repeat`,
			`  end repeat`,
			`  if targetSession is not missing value then`,
			// Select tab/session so write text hits an interactive prompt reliably.
			`    try`,
			`      select targetWindow`,
			`      select targetTab`,
			`      select targetSession`,
			`    on error`,
			`    end try`,
			`    tell targetSession`,
		)
		if tab.Command != "" {
			lines = append(lines, fmt.Sprintf(`      %s`, writeTextCommand(tab.Command, tab.NoSubmit)))
		}
		lines = append(lines,
			`    end tell`,
			`  end if`,
		)
	} else {
		// Best-effort: write into current session of matching window.
		lines = append(lines,
			`  tell current session of current window`,
		)
		if tab.Command != "" {
			lines = append(lines, fmt.Sprintf(`    %s`, writeTextCommand(tab.Command, tab.NoSubmit)))
		}
		lines = append(lines, `  end tell`)
	}
	lines = append(lines, `end tell`)
	return strings.Join(lines, "\n")
}

func uniqueWindowIDs(refs []TabSessionRef) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range refs {
		id := r.WindowID
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func normalizeTabSetConfig(cfg *TabSetConfig) *TabSetConfig {
	if cfg == nil {
		cfg = &TabSetConfig{}
	}
	// Copy so we don't mutate caller.
	out := *cfg
	if out.Find == nil {
		out.Find = defaultTabSetFind
	}
	if out.Busy == nil {
		out.Busy = defaultTabSetBusy
	}
	if out.Exec == nil {
		out.Exec = defaultOsascript
	}
	return &out
}

func defaultTabSetFind(setName string) ([]TabSessionRef, error) {
	script := BuildTabSetFindScript(setName)
	out, err := runOsascriptOutput(script)
	if err != nil {
		return nil, err
	}
	return ParseTabSetFindOutput(out)
}

func defaultTabSetBusy(ref TabSessionRef) BusyState {
	tty := strings.TrimSpace(ref.TTY)
	if tty == "" {
		return BusyStateUnknown
	}
	comm, ok := foregroundCommForTTY(tty)
	return ClassifyBusyFromComm(comm, ok)
}
