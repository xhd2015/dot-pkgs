package iterm2

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/window"
)

// Status model types live in package window; aliases keep existing iterm2 importers stable.
type (
	CurrentLocation = window.CurrentLocation
	TabStatusRow    = window.TabStatusRow
	WindowStatus    = window.WindowStatus
	TabStatus       = window.TabStatus
)

// CurrentStatusConfig injects resolve inputs for tests.
type CurrentStatusConfig struct {
	// SessionID returns ITERM_SESSION_ID (or equivalent). Nil uses the process env.
	SessionID func() string
	// ListSessions returns a live session scan. Nil runs BuildSessionListScript via osascript.
	ListSessions func() ([]SessionRef, error)
	// ControllingTTY returns the process controlling TTY (e.g. /dev/ttys143). Nil probes.
	ControllingTTY func() string
	// AncestorTTYs returns ancestor process TTYs in walk order (nearest first). Nil probes via ps.
	AncestorTTYs func() []string
}

func normalizeCurrentStatusConfig(cfg *CurrentStatusConfig) CurrentStatusConfig {
	out := CurrentStatusConfig{}
	if cfg != nil {
		out = *cfg
	}
	if out.SessionID == nil {
		out.SessionID = currentSessionID
	}
	if out.ListSessions == nil {
		out.ListSessions = ListSessions
	}
	if out.ControllingTTY == nil {
		out.ControllingTTY = probeControllingTTY
	}
	if out.AncestorTTYs == nil {
		out.AncestorTTYs = probeAncestorTTYs
	}
	return out
}

// ListSessions runs BuildSessionListScript and parses the result.
// Retries once on AppleScript Invalid index (-1719) mid-scan races.
func ListSessions() ([]SessionRef, error) {
	return runSessionListScript(BuildSessionListScript())
}

// listRefsForStatus prefers a UUID-scoped window dump when ITERM_SESSION_ID is
// set and ListSessions was not injected (narrower race window). Falls back to
// the full session list for TTY/ancestor resolve.
func listRefsForStatus(c CurrentStatusConfig, cfg *CurrentStatusConfig) ([]SessionRef, error) {
	useNarrow := cfg == nil || cfg.ListSessions == nil
	if useNarrow {
		if sid := strings.TrimSpace(c.SessionID()); sid != "" {
			refs, err := ListSessionsInWindowByUUID(sid)
			if err == nil && len(FindBySessionUUID(refs, sid)) > 0 {
				return refs, nil
			}
			// Miss or transient error: fall through to full scan.
		}
	}
	return c.ListSessions()
}

// FindBySessionUUID returns refs whose SessionID equals or contains uuid
// (case-insensitive). Empty uuid yields an empty slice.
func FindBySessionUUID(refs []SessionRef, uuid string) []SessionRef {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return []SessionRef{}
	}
	want := strings.ToLower(SessionUUID(uuid))
	if want == "" {
		return []SessionRef{}
	}
	var out []SessionRef
	for _, ref := range refs {
		sid := strings.ToLower(strings.TrimSpace(ref.SessionID))
		if sid == "" {
			continue
		}
		if sid == want || strings.Contains(sid, want) {
			out = append(out, ref)
		}
	}
	if out == nil {
		return []SessionRef{}
	}
	return out
}

// ResolveCurrentLocation finds the iTerm2 pane that parents this process.
//
// Order: ITERM_SESSION_ID UUID → controlling TTY in the session list → first
// ancestor TTY that appears in the session list. Returns ErrNotInSession when
// none match.
func ResolveCurrentLocation() (CurrentLocation, error) {
	return ResolveCurrentLocationWith(nil)
}

// ResolveCurrentLocationWith is ResolveCurrentLocation with injectable probes.
func ResolveCurrentLocationWith(cfg *CurrentStatusConfig) (CurrentLocation, error) {
	c := normalizeCurrentStatusConfig(cfg)
	refs, err := listRefsForStatus(c, cfg)
	if err != nil {
		return CurrentLocation{}, err
	}
	return resolveCurrentLocationFrom(refs, c)
}

func resolveCurrentLocationFrom(refs []SessionRef, c CurrentStatusConfig) (CurrentLocation, error) {
	if sid := strings.TrimSpace(c.SessionID()); sid != "" {
		matches := FindBySessionUUID(refs, sid)
		if loc, ok := uniqueLocation(matches); ok {
			return loc, nil
		}
	}

	if tty := strings.TrimSpace(c.ControllingTTY()); tty != "" {
		matches := FindByTTY(refs, []string{tty})
		if loc, ok := uniqueLocation(matches); ok {
			return loc, nil
		}
	}

	for _, tty := range c.AncestorTTYs() {
		tty = strings.TrimSpace(tty)
		if tty == "" || tty == "??" || tty == "?" {
			continue
		}
		matches := FindByTTY(refs, []string{tty})
		if loc, ok := uniqueLocation(matches); ok {
			return loc, nil
		}
	}

	return CurrentLocation{}, ErrNotInSession
}

func uniqueLocation(matches []SessionRef) (CurrentLocation, bool) {
	if len(matches) == 0 {
		return CurrentLocation{}, false
	}
	// Prefer a single match; if several share the same window/tab/session, take the first.
	first := matches[0]
	for _, m := range matches[1:] {
		if m.WindowID != first.WindowID || m.TabIndex != first.TabIndex || SessionUUID(m.SessionID) != SessionUUID(first.SessionID) {
			// Ambiguous: still return the first (stable) — callers rarely hit this when UUID/TTY unique.
			break
		}
	}
	return locationFromRef(first), true
}

func locationFromRef(ref SessionRef) CurrentLocation {
	return CurrentLocation{
		WindowID:    ref.WindowID,
		WindowName:  ref.WindowName,
		TabIndex:    ref.TabIndex,
		SessionID:   ref.SessionID,
		TTY:         ref.TTY,
		SessionName: ref.Name,
	}
}

// BuildWindowStatus lists every tab in loc's window, marking the current tab.
func BuildWindowStatus(refs []SessionRef, loc CurrentLocation) WindowStatus {
	return window.BuildWindowStatus(sessionRefsToPaneRefs(refs), loc)
}

// BuildTabStatus returns the current-tab summary for loc.
func BuildTabStatus(loc CurrentLocation) TabStatus {
	return window.BuildTabStatus(loc)
}

func sessionRefsToPaneRefs(refs []SessionRef) []window.PaneRef {
	out := make([]window.PaneRef, len(refs))
	for i, ref := range refs {
		out[i] = window.PaneRef{
			WindowID:   ref.WindowID,
			WindowName: ref.WindowName,
			TabIndex:   ref.TabIndex,
			SessionID:  ref.SessionID,
			TTY:        ref.TTY,
			Name:       ref.Name,
		}
	}
	return out
}

// CurrentWindowStatus resolves the parent window and lists its tabs.
func CurrentWindowStatus() (WindowStatus, error) {
	return CurrentWindowStatusWith(nil)
}

// CurrentWindowStatusWith is CurrentWindowStatus with injectable probes.
func CurrentWindowStatusWith(cfg *CurrentStatusConfig) (WindowStatus, error) {
	c := normalizeCurrentStatusConfig(cfg)
	refs, err := listRefsForStatus(c, cfg)
	if err != nil {
		return WindowStatus{}, err
	}
	loc, err := resolveCurrentLocationFrom(refs, c)
	if err != nil {
		return WindowStatus{}, err
	}
	return BuildWindowStatus(refs, loc), nil
}

// CurrentTabStatus resolves the parent tab summary.
func CurrentTabStatus() (TabStatus, error) {
	return CurrentTabStatusWith(nil)
}

// CurrentTabStatusWith is CurrentTabStatus with injectable probes.
func CurrentTabStatusWith(cfg *CurrentStatusConfig) (TabStatus, error) {
	loc, err := ResolveCurrentLocationWith(cfg)
	if err != nil {
		return TabStatus{}, err
	}
	return BuildTabStatus(loc), nil
}

// FormatWindowStatus renders WindowStatus for human CLI stdout.
func FormatWindowStatus(st WindowStatus) string {
	return window.FormatWindowStatus(st)
}

// FormatTabStatus renders TabStatus for human CLI stdout.
func FormatTabStatus(st TabStatus) string {
	return window.FormatTabStatus(st)
}

func probeControllingTTY() string {
	// Prefer the process controlling terminal via ps; falls back to empty.
	out, err := exec.Command("ps", "-p", strconv.Itoa(os.Getpid()), "-o", "tty=").Output()
	if err != nil {
		return ""
	}
	tty := strings.TrimSpace(string(out))
	if tty == "" || tty == "??" || tty == "?" {
		return ""
	}
	return NormalizeTTY(tty)
}

func probeAncestorTTYs() []string {
	out, err := exec.Command("ps", "-ax", "-o", "pid=,ppid=,tty=").Output()
	if err != nil {
		return nil
	}
	type row struct {
		ppid int
		tty  string
	}
	byPID := map[int]row{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		byPID[pid] = row{ppid: ppid, tty: fields[2]}
	}
	var ttys []string
	seen := map[string]struct{}{}
	pid := os.Getpid()
	for hops := 0; hops < 64; hops++ {
		r, ok := byPID[pid]
		if !ok {
			break
		}
		tty := strings.TrimSpace(r.tty)
		if tty != "" && tty != "??" && tty != "?" {
			n := NormalizeTTY(tty)
			if _, exists := seen[n]; !exists {
				seen[n] = struct{}{}
				ttys = append(ttys, n)
			}
		}
		if r.ppid <= 0 || r.ppid == pid {
			break
		}
		pid = r.ppid
	}
	return ttys
}
