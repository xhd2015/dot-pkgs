package iterm2

import (
	"path/filepath"
	"strings"
)

// BusyState classifies whether a tab-set session looks idle, busy, or unknown.
type BusyState int

const (
	// BusyStateIdle means the foreground process is a login shell (waiting for input).
	BusyStateIdle BusyState = iota
	// BusyStateBusy means a non-shell foreground process is running.
	BusyStateBusy
	// BusyStateUnknown means the probe failed or the process name was empty.
	BusyStateUnknown
)

// ClassifyBusyFromComm maps a foreground process name (and probe success) to BusyState.
//
//	!ok or empty/whitespace fgComm → BusyStateUnknown
//	basename in {zsh, bash, fish, sh} (optional leading "-" for login shells) → BusyStateIdle
//	any other non-empty name → BusyStateBusy
//
// Login shells often appear as "-bash" / "-zsh" in ps COMM; those must count as idle
// so Ctrl-C back to the prompt can re-run via write text.
func ClassifyBusyFromComm(fgComm string, ok bool) BusyState {
	if !ok {
		return BusyStateUnknown
	}
	name := strings.TrimSpace(fgComm)
	if name == "" {
		return BusyStateUnknown
	}
	base := filepath.Base(name)
	// login shell: argv0 often starts with '-'
	base = strings.TrimPrefix(base, "-")
	switch base {
	case "zsh", "bash", "fish", "sh":
		return BusyStateIdle
	default:
		return BusyStateBusy
	}
}
