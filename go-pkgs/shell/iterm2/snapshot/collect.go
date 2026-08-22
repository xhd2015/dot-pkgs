package snapshot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

// TTYProc is one process row from ListTTYProcs, tagged with its tty field
// as reported by ps (short form after /dev/ trim when normalizing).
type TTYProc struct {
	ProcRow
	TTY string
}

// Collector gathers hierarchy + process enrichment. Fields may be overridden
// per instance (parallel-safe; no process-global required for production or tests).
type Collector struct {
	// RunAppleScript runs an AppleScript body and returns stdout.
	RunAppleScript func(script string) (string, error)
	// ListProcs returns processes on a short tty name (e.g. "ttys003").
	// Used by fixture inject and as fallback when ListAllProcs is nil.
	ListProcs func(ttyShort string) ([]ProcRow, error)
	// ListAllProcs returns all processes keyed by short tty (e.g. "ttys003").
	// When set (production default), capture uses one global ps instead of
	// per-TTY ps -t.
	ListAllProcs func() (map[string][]ProcRow, error)
	// ListTTYProcs returns one process listing for the requested terminal devices
	// (batched ps -t a,b,c). Used by live critical scans; optional for Capture.
	ListTTYProcs func(ttyShorts []string) ([]TTYProc, error)
	// ListCwds returns cwd paths keyed by pid for the given pids.
	// Only used when CaptureOpts.IncludeCwd is true.
	ListCwds func(pids []int) (map[int]string, error)
	// ITermRunning reports whether iTerm2 appears to be running.
	ITermRunning func() bool
	// Now is the clock (defaults to time.Now).
	Now func() time.Time
	// Hostname defaults to os.Hostname (short form).
	Hostname func() (string, error)
	// ResolveSpace maps an iTerm/CG window id to a 0-based Desktop index.
	// Nil → space.SpaceIndexForWindow. Tests inject; FixedSpace short-circuits.
	ResolveSpace func(windowID uint64) (int, error)

	// AppTell is the AppleScript application target for live capture.
	// Empty → "iTerm2". Absolute path (or ~/…) → tell that .app.
	AppTell string
	// AppTag is stamped on every window from this collector when App is empty
	// (canonical install path for multi-app callers).
	AppTag string

	// OnListWindows is an optional hook invoked at the start of ListWindows
	// (streaming / test probes).
	OnListWindows func()
	// OnListTabs is an optional hook invoked at the start of ListTabsAndSessions.
	OnListTabs func(windowIndex int)

	// fixtureEnabled + fixtureWindows drive ListWindows / ListTabs without AppleScript.
	fixtureEnabled bool
	fixtureWindows []SnapshotWindow
}

// NewCollector returns a Collector with production defaults (live osascript/ps).
// Cwd lsof is off unless CaptureOpts.IncludeCwd is set.
func NewCollector() *Collector {
	return &Collector{
		RunAppleScript: defaultRunAppleScript,
		ListAllProcs:   defaultListAllProcs,
		ListTTYProcs:   defaultListTTYProcs,
		ListCwds:       defaultListCwds,
		ITermRunning:   defaultITermRunning,
		ResolveSpace:   defaultResolveSpace,
		Now:            time.Now,
		Hostname: func() (string, error) {
			h, err := os.Hostname()
			if err != nil {
				return "", err
			}
			if i := strings.Index(h, "."); i > 0 {
				return h[:i], nil
			}
			return h, nil
		},
	}
}

// Capture is NewCollector().Capture() convenience for production callers.
func Capture() (*Snapshot, []string, error) {
	return NewCollector().Capture()
}

// Capture runs hierarchy collection + process enrichment (fast defaults).
func (c *Collector) Capture() (*Snapshot, []string, error) {
	return c.CaptureWith(CaptureOpts{})
}

// CaptureWith runs Capture with options.
func (c *Collector) CaptureWith(opts CaptureOpts) (*Snapshot, []string, error) {
	return c.capture(nil, opts)
}

// CaptureProgressive is Capture with an optional per-window callback after that
// window's tabs/sessions are process-enriched (CLI streaming, save dry-run).
// Hierarchy is loaded once (single AppleScript when not fixture); callbacks run
// after each window is process-enriched.
func (c *Collector) CaptureProgressive(onWindowReady func(win SnapshotWindow) error) (*Snapshot, []string, error) {
	return c.capture(onWindowReady, CaptureOpts{})
}

// CaptureProgressiveWith is CaptureProgressive with CaptureOpts (e.g. SpaceAllow).
func (c *Collector) CaptureProgressiveWith(opts CaptureOpts, onWindowReady func(win SnapshotWindow) error) (*Snapshot, []string, error) {
	return c.capture(onWindowReady, opts)
}

func (c *Collector) capture(onWindowReady func(win SnapshotWindow) error, opts CaptureOpts) (*Snapshot, []string, error) {
	if c.ITermRunning != nil && !c.ITermRunning() {
		return nil, nil, fmt.Errorf("Error: iTerm2 is not running")
	}
	nowFn := c.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	hostFn := c.Hostname
	if hostFn == nil {
		hostFn = os.Hostname
	}
	now := nowFn()
	host, _ := hostFn()

	// Process table: one global ps when ListAllProcs is set (production fast path).
	// Progressive streaming (onWindowReady), fixtures, and SpaceAllow use phased
	// hierarchy so headers can be gated before expensive tabs/enrich.
	var warnings []string
	spaceGate := len(opts.SpaceAllow) > 0
	useBatch := !c.fixtureEnabled && onWindowReady == nil && !spaceGate

	procsByTTY, wProc := c.loadProcsByTTY(&warnings)
	warnings = append(warnings, wProc...)

	wantCwd := opts.IncludeCwd || c.fixtureEnabled
	cwdsByTTY := c.loadCwdsByTTY(procsByTTY, wantCwd, &warnings)

	var windows []SnapshotWindow
	var err error
	if useBatch {
		windows, warnings, err = c.captureBatch(now, procsByTTY, cwdsByTTY, warnings)
	} else {
		windows, warnings, err = c.capturePhased(now, procsByTTY, cwdsByTTY, onWindowReady, opts, warnings)
	}
	if err != nil {
		return nil, warnings, err
	}

	var nTabs, nSess, nIdle, nBusy, nUnknown int
	for _, win := range windows {
		for _, t := range win.Tabs {
			nTabs++
			nSess += len(t.Sessions)
			for _, s := range t.Sessions {
				if s.Idle == nil {
					nUnknown++
				} else if *s.Idle {
					nIdle++
				} else {
					nBusy++
				}
			}
		}
	}

	snap := &Snapshot{
		CapturedAt: now.Format("2006-01-02T15:04:05") + zoneOffset(now),
		Host:       host,
		Source:     "iterm2",
		Summary: SnapshotSummary{
			Windows:  len(windows),
			Tabs:     nTabs,
			Sessions: nSess,
			Idle:     nIdle,
			Busy:     nBusy,
			Unknown:  nUnknown,
		},
		Windows: windows,
	}
	return snap, warnings, nil
}

func (c *Collector) loadCwdsByTTY(procsByTTY map[string][]ProcRow, want bool, warnings *[]string) map[string]map[int]string {
	out := map[string]map[int]string{}
	if !want || c.ListCwds == nil {
		return out
	}
	for short, procs := range procsByTTY {
		var pids []int
		for _, p := range procs {
			pids = append(pids, p.PID)
		}
		if len(pids) == 0 {
			continue
		}
		m, err := c.ListCwds(pids)
		if err != nil {
			*warnings = append(*warnings, fmt.Sprintf("warning: cwd probe failed for %s: %v", short, err))
			continue
		}
		out[short] = m
	}
	return out
}

func enrichWindowSessions(win *SnapshotWindow, procsByTTY map[string][]ProcRow, cwdsByTTY map[string]map[int]string, now time.Time, getProcs func(short string) []ProcRow) {
	for ti := range win.Tabs {
		t := &win.Tabs[ti]
		for si := range t.Sessions {
			s := &t.Sessions[si]
			s.WindowIndex = win.Index
			s.TabIndex = t.Index
			short := strings.TrimPrefix(s.TTY, "/dev/")
			var procs []ProcRow
			if getProcs != nil {
				procs = getProcs(short)
			} else {
				procs = procsByTTY[short]
			}
			cwds := cwdsByTTY[short]
			if cwds == nil {
				cwds = map[int]string{}
			}
			idle, shellPID, chosen, cwd, snapProcs := EnrichFromProcs(procs, cwds, now)
			applyChosenToSession(s, idle, shellPID, chosen, cwd, snapProcs, now)
		}
	}
}

// captureBatch: one AppleScript hierarchy + preloaded process map (kck fast path).
// On batch AS failure, fall back to phased per-window queries (slower, more resilient).
func (c *Collector) captureBatch(now time.Time, procsByTTY map[string][]ProcRow, cwdsByTTY map[string]map[int]string, warnings []string) ([]SnapshotWindow, []string, error) {
	windows, w2, err := c.loadHierarchyBatch()
	warnings = append(warnings, w2...)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("warning: batch hierarchy failed, falling back to phased: %v", err))
		return c.capturePhased(now, procsByTTY, cwdsByTTY, nil, CaptureOpts{}, warnings)
	}
	for wi := range windows {
		c.stampApp(&windows[wi])
		enrichWindowSessions(&windows[wi], procsByTTY, cwdsByTTY, now, nil)
	}
	return windows, warnings, nil
}

// capturePhased: per-window ListTabs (fixtures + streaming + SpaceAllow gate).
func (c *Collector) capturePhased(now time.Time, procsByTTY map[string][]ProcRow, cwdsByTTY map[string]map[int]string, onWindowReady func(SnapshotWindow) error, opts CaptureOpts, warnings []string) ([]SnapshotWindow, []string, error) {
	headers, w2, err := c.ListWindows()
	warnings = append(warnings, w2...)
	if err != nil {
		return nil, warnings, err
	}

	// Lazy per-TTY fill when ListAllProcs was nil (fixtures).
	cache := map[string][]ProcRow{}
	for k, v := range procsByTTY {
		cache[k] = v
	}
	listProcs := c.ListProcs
	if listProcs == nil {
		listProcs = defaultListProcs
	}
	wantCwd := opts.IncludeCwd || c.fixtureEnabled
	getProcs := func(short string) []ProcRow {
		if short == "" {
			return nil
		}
		if p, ok := cache[short]; ok {
			return p
		}
		procs, err := listProcs(short)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("warning: ps failed for %s: %v", short, err))
			cache[short] = nil
			return nil
		}
		if len(procs) == 0 {
			warnings = append(warnings, fmt.Sprintf("warning: no processes on %s", short))
		}
		// Optional cwd for this tty when IncludeCwd/fixture and ListCwds set.
		if c.ListCwds != nil && wantCwd && cwdsByTTY[short] == nil && len(procs) > 0 {
			var pids []int
			for _, p := range procs {
				pids = append(pids, p.PID)
			}
			if m, err := c.ListCwds(pids); err == nil {
				cwdsByTTY[short] = m
			}
		}
		cache[short] = procs
		return procs
	}

	allowSet := spaceAllowSet(opts.SpaceAllow)
	spaceGate := len(allowSet) > 0
	spaceSkipped := 0

	windows := make([]SnapshotWindow, 0, len(headers))
	for _, hdr := range headers {
		c.stampApp(&hdr)
		hdrWin := SnapshotWindow{
			Index: hdr.Index, Name: hdr.Name, WindowID: hdr.WindowID,
			FixedSpace: hdr.FixedSpace, App: hdr.App,
		}
		spaceIdx, spaceWarn := c.resolveWindowSpace(hdrWin)
		if spaceGate && !spaceAllowed(spaceIdx, allowSet) {
			spaceSkipped++
			continue
		}
		// Pin resolved space so later callers stay consistent without re-resolve.
		sp := spaceIdx
		hdrWin.FixedSpace = &sp

		tabs, w3, err := c.ListTabsAndSessions(hdr.Index)
		if err != nil {
			return nil, append(warnings, w3...), err
		}
		warnings = append(warnings, w3...)
		if spaceWarn != "" {
			warnings = append(warnings, "warning: "+spaceWarn)
		}
		win := SnapshotWindow{
			Index: hdr.Index, Name: hdr.Name, WindowID: hdr.WindowID,
			FixedSpace: hdrWin.FixedSpace, App: hdr.App, Tabs: tabs,
		}
		enrichWindowSessions(&win, nil, cwdsByTTY, now, getProcs)
		if onWindowReady != nil {
			if err := onWindowReady(win); err != nil {
				return nil, warnings, err
			}
		}
		windows = append(windows, win)
	}
	if opts.SpaceSkipped != nil {
		*opts.SpaceSkipped = spaceSkipped
	}
	return windows, warnings, nil
}

func (c *Collector) loadHierarchyBatch() ([]SnapshotWindow, []string, error) {
	if c.OnListWindows != nil {
		c.OnListWindows()
	}
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	raw, err := runAS(listAllHierarchyAppleScript(c.AppTell))
	if err != nil {
		return nil, nil, fmt.Errorf("Error: failed to query iTerm2: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SnapshotWindow{}, nil, nil
	}
	wins, warnings := parseHierarchy(raw)
	for i := range wins {
		c.stampApp(&wins[i])
	}
	return wins, warnings, nil
}

func (c *Collector) loadHierarchyPhased() ([]SnapshotWindow, []string, error) {
	headers, warnings, err := c.ListWindows()
	if err != nil {
		return nil, warnings, err
	}
	windows := make([]SnapshotWindow, 0, len(headers))
	for _, hdr := range headers {
		tabs, w2, err := c.ListTabsAndSessions(hdr.Index)
		if err != nil {
			return nil, append(warnings, w2...), err
		}
		warnings = append(warnings, w2...)
		win := SnapshotWindow{
			Index: hdr.Index, Name: hdr.Name, WindowID: hdr.WindowID,
			FixedSpace: hdr.FixedSpace, App: hdr.App, Tabs: tabs,
		}
		c.stampApp(&win)
		windows = append(windows, win)
	}
	return windows, warnings, nil
}

func (c *Collector) loadProcsByTTY(warnings *[]string) (map[string][]ProcRow, []string) {
	var extra []string
	if c.ListAllProcs != nil {
		byTTY, err := c.ListAllProcs()
		if err != nil {
			extra = append(extra, fmt.Sprintf("warning: ps failed: %v", err))
			return map[string][]ProcRow{}, extra
		}
		if byTTY == nil {
			byTTY = map[string][]ProcRow{}
		}
		return byTTY, extra
	}
	// Per-TTY path (fixtures / inject).
	listProcs := c.ListProcs
	if listProcs == nil {
		listProcs = defaultListProcs
	}
	// Collect unique TTYs from fixtures only happens after hierarchy; callers
	// that use ListProcs alone still work via get-on-demand below.
	// For phased fixture, ListProcs is set and ListAllProcs cleared — build
	// lazily from sessions by scanning nothing here; enrich path needs map.
	// Return empty; enrich will call listProcs via helper... actually capture
	// already expects full map. Build from fixture windows if enabled.
	byTTY := map[string][]ProcRow{}
	seen := map[string]bool{}
	var sessions []string
	if c.fixtureEnabled {
		for _, w := range c.fixtureWindows {
			for _, t := range w.Tabs {
				for _, s := range t.Sessions {
					sessions = append(sessions, s.TTY)
				}
			}
		}
	}
	// Also walk is not available yet; for fixture ApplyPhasedFixture clones
	// into fixtureWindows. loadHierarchyPhased reads from fixtureWindows too.
	for _, w := range c.fixtureWindows {
		for _, t := range w.Tabs {
			for _, s := range t.Sessions {
				short := strings.TrimPrefix(s.TTY, "/dev/")
				if short == "" || seen[short] {
					continue
				}
				seen[short] = true
				procs, err := listProcs(short)
				if err != nil {
					extra = append(extra, fmt.Sprintf("warning: ps failed for %s: %v", s.TTY, err))
					continue
				}
				byTTY[short] = procs
				if len(procs) == 0 {
					extra = append(extra, fmt.Sprintf("warning: no processes on %s", s.TTY))
				}
			}
		}
	}
	_ = sessions
	return byTTY, extra
}

// PhasedFixtureOpts configures ApplyPhasedFixture (tests / inject path).
type PhasedFixtureOpts struct {
	Windows       []SnapshotWindow
	ITermRunning  bool
	IdleTTYs      []string          // short tty names classified idle (shell only)
	BusyTTYs      []string          // short tty names classified busy
	BusyLeafByTTY map[string]string // optional busy leaf command override
	CwdByTTY      map[string]string // cwd for processes on that short tty
	Now           time.Time
	Hostname      string
}

// ApplyPhasedFixture configures this Collector for fixture hierarchy + process
// enrich without real AppleScript/ps/lsof. Mutates c only (parallel-safe when
// each test owns its Collector). No process-global collector mutation.
func (c *Collector) ApplyPhasedFixture(opts PhasedFixtureOpts) {
	idleSet := map[string]bool{}
	for _, tty := range opts.IdleTTYs {
		idleSet[strings.TrimPrefix(tty, "/dev/")] = true
	}
	busySet := map[string]bool{}
	for _, tty := range opts.BusyTTYs {
		busySet[strings.TrimPrefix(tty, "/dev/")] = true
	}
	leafByTTY := map[string]string{}
	for k, v := range opts.BusyLeafByTTY {
		leafByTTY[strings.TrimPrefix(k, "/dev/")] = v
	}
	cwdByTTY := map[string]string{}
	for k, v := range opts.CwdByTTY {
		cwdByTTY[strings.TrimPrefix(k, "/dev/")] = v
	}
	// Map pid → tty short for ListCwds (fixture pids are unique per tty family).
	pidTTY := map[int]string{}
	now := opts.Now
	if now.IsZero() {
		now = time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	host := opts.Hostname
	if host == "" {
		host = "testhost"
	}
	// Clone windows so enrich mutations do not alter the caller's slice.
	fxWindows := cloneWindows(opts.Windows)

	c.fixtureEnabled = true
	c.fixtureWindows = fxWindows
	// Prefer per-TTY ListProcs inject; disable global batch for fixtures.
	c.ListAllProcs = nil
	c.ITermRunning = func() bool { return opts.ITermRunning }
	c.Now = func() time.Time { return now }
	c.Hostname = func() (string, error) { return host, nil }
	c.ListProcs = func(ttyShort string) ([]ProcRow, error) {
		short := strings.TrimPrefix(ttyShort, "/dev/")
		if idleSet[short] {
			pidTTY[1] = short
			pidTTY[2] = short
			return []ProcRow{
				{PID: 1, PPID: 0, Stat: "Ss", Etime: "1:00", RSSKB: 1000, Command: "login -fp u"},
				{PID: 2, PPID: 1, Stat: "S+", Etime: "0:59", RSSKB: 2000, Command: "-zsh"},
			}, nil
		}
		if busySet[short] {
			leaf := leafByTTY[short]
			if leaf == "" {
				leaf = "python train.py"
			}
			// Distinct pids per tty so cwd map can key by pid.
			base := 100 + len(pidTTY)*10
			pidTTY[base] = short
			pidTTY[base+1] = short
			pidTTY[base+2] = short
			return []ProcRow{
				{PID: base, PPID: 0, Stat: "Ss", Etime: "1:00", RSSKB: 1000, Command: "login -fp u"},
				{PID: base + 1, PPID: base, Stat: "S", Etime: "0:59", RSSKB: 2000, Command: "-zsh"},
				{PID: base + 2, PPID: base + 1, Stat: "R+", Etime: "0:30", RSSKB: 8000, Command: leaf},
			}, nil
		}
		return nil, nil
	}
	c.ListCwds = func(pids []int) (map[int]string, error) {
		m := map[int]string{}
		for _, p := range pids {
			cwd := "/tmp"
			if short, ok := pidTTY[p]; ok {
				if c, ok := cwdByTTY[short]; ok && c != "" {
					cwd = c
				}
			}
			m[p] = cwd
		}
		return m, nil
	}
	c.RunAppleScript = func(string) (string, error) {
		return "", fmt.Errorf("fixture collector: AppleScript not used")
	}
}

// ListWindows returns window index + name headers (tabs may be empty).
func (c *Collector) ListWindows() (windows []SnapshotWindow, warnings []string, err error) {
	if c.OnListWindows != nil {
		c.OnListWindows()
	}
	if c.fixtureEnabled {
		out := make([]SnapshotWindow, len(c.fixtureWindows))
		for i, w := range c.fixtureWindows {
			out[i] = SnapshotWindow{
				Index: w.Index, Name: w.Name, WindowID: w.WindowID,
				FixedSpace: w.FixedSpace, App: w.App,
			}
			c.stampApp(&out[i])
		}
		return out, nil, nil
	}
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	raw, err := runAS(listWindowsAppleScript(c.AppTell))
	if err != nil {
		return nil, nil, fmt.Errorf("Error: failed to query iTerm2: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		// No windows is valid.
		return []SnapshotWindow{}, nil, nil
	}
	wins, warnings := parseHierarchy(raw)
	// Ensure headers-only (strip any accidental tabs).
	for i := range wins {
		wins[i].Tabs = nil
		c.stampApp(&wins[i])
	}
	return wins, warnings, nil
}

// ListTabsAndSessions returns tabs and sessions for one window (by 1-based index).
func (c *Collector) ListTabsAndSessions(windowIndex int) (tabs []SnapshotTab, warnings []string, err error) {
	if c.OnListTabs != nil {
		c.OnListTabs(windowIndex)
	}
	if c.fixtureEnabled {
		for _, w := range c.fixtureWindows {
			if w.Index == windowIndex {
				return cloneTabs(w.Tabs), nil, nil
			}
		}
		return nil, nil, fmt.Errorf("Error: window %d not found", windowIndex)
	}
	runAS := c.RunAppleScript
	if runAS == nil {
		runAS = defaultRunAppleScript
	}
	script := listTabsAndSessionsAppleScript(windowIndex, c.AppTell)
	raw, err := runAS(script)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: failed to query iTerm2 window %d: %w", windowIndex, err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []SnapshotTab{}, nil, nil
	}
	// parseHierarchy expects optional ###W###; tabs may appear alone.
	// Prefix a synthetic window so session/tab parsing works.
	wrapped := fmt.Sprintf("###W###%d###\n%s", windowIndex, raw)
	wins, warnings := parseHierarchy(wrapped)
	if len(wins) == 0 {
		return []SnapshotTab{}, warnings, nil
	}
	return wins[0].Tabs, warnings, nil
}

func cloneWindows(in []SnapshotWindow) []SnapshotWindow {
	out := make([]SnapshotWindow, len(in))
	for i, w := range in {
		out[i] = SnapshotWindow{
			Index: w.Index, Name: w.Name, WindowID: w.WindowID,
			FixedSpace: cloneIntPtr(w.FixedSpace), App: w.App,
			Tabs: cloneTabs(w.Tabs),
		}
	}
	return out
}

func cloneIntPtr(p *int) *int {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func (c *Collector) stampApp(w *SnapshotWindow) {
	if w == nil {
		return
	}
	if w.App == "" && c.AppTag != "" {
		w.App = c.AppTag
	}
}

func spaceAllowSet(allow []int) map[int]struct{} {
	if len(allow) == 0 {
		return nil
	}
	set := make(map[int]struct{}, len(allow))
	for _, s := range allow {
		set[s] = struct{}{}
	}
	return set
}

func spaceAllowed(spaceIdx int, set map[int]struct{}) bool {
	if len(set) == 0 {
		return true
	}
	_, ok := set[spaceIdx]
	return ok
}

// resolveWindowSpace maps a snapshot window to (space index, warning without prefix).
// FixedSpace wins; missing id or resolve failure → space 0 + warning.
func (c *Collector) resolveWindowSpace(win SnapshotWindow) (spaceIdx int, warn string) {
	if win.FixedSpace != nil {
		idx := *win.FixedSpace
		if idx < 0 {
			return 0, fmt.Sprintf("invalid macOS Space index %d; using space 0", idx)
		}
		return idx, ""
	}
	if win.WindowID == 0 {
		return 0, "could not resolve macOS Space (missing iterm window id); using space 0"
	}
	resolve := c.ResolveSpace
	if resolve == nil {
		resolve = defaultResolveSpace
	}
	idx, err := resolve(win.WindowID)
	if err != nil {
		return 0, fmt.Sprintf("could not resolve macOS Space for window id %d: %v; using space 0", win.WindowID, err)
	}
	if idx < 0 {
		return 0, fmt.Sprintf("invalid macOS Space index %d for window id %d; using space 0", idx, win.WindowID)
	}
	return idx, ""
}

func defaultResolveSpace(windowID uint64) (int, error) {
	return space.SpaceIndexForWindow(windowID)
}

func cloneTabs(tabs []SnapshotTab) []SnapshotTab {
	if tabs == nil {
		return nil
	}
	out := make([]SnapshotTab, len(tabs))
	for i, t := range tabs {
		out[i] = SnapshotTab{Index: t.Index, Name: t.Name}
		if len(t.Sessions) > 0 {
			out[i].Sessions = make([]SnapshotSession, len(t.Sessions))
			copy(out[i].Sessions, t.Sessions)
		}
	}
	return out
}

// appleScriptAppLiteral returns the tell-application target string for AS.
// Empty → "iTerm2". Path → quoted absolute path (escaped for AS double quotes).
func appleScriptAppLiteral(appTell string) string {
	appTell = strings.TrimSpace(appTell)
	if appTell == "" || appTell == "iTerm2" {
		return `"iTerm2"`
	}
	if strings.HasPrefix(appTell, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			appTell = home + appTell[1:]
		}
	}
	esc := strings.ReplaceAll(appTell, `\`, `\\`)
	esc = strings.ReplaceAll(esc, `"`, `\"`)
	return `"` + esc + `"`
}

// listAllHierarchyAppleScript dumps every window/tab/session in one osascript.
// Format matches parseHierarchy (###W### / ###T### / ###S###).
func listAllHierarchyAppleScript(appTell string) string {
	return fmt.Sprintf(`
tell application %s
  set out to ""
  set wi to 0
  set winList to windows
  repeat with w in winList
    set wi to wi + 1
    try
      set wname to name of w
    on error
      set wname to ""
    end try
    try
      set wid to id of w
    on error
      set wid to 0
    end try
    set out to out & "###W###" & wi & "###" & wname & "###" & wid & linefeed
    try
      set tabList to tabs of w
      set ti to 0
      repeat with t in tabList
        set ti to ti + 1
        try
          set tname to name of current session of t
        on error
          set tname to ""
        end try
        set out to out & "###T###" & ti & "###" & tname & linefeed
        try
          set sessList to sessions of t
          set si to 0
          repeat with s in sessList
            set si to si + 1
            try
              set nm to name of s
            on error
              set nm to "?"
            end try
            try
              set ttyn to tty of s
            on error
              set ttyn to ""
            end try
            try
              set prof to profile name of s
            on error
              set prof to ""
            end try
            try
              set proc to is processing of s
            on error
              set proc to false
            end try
            try
              set uid to unique ID of s
            on error
              set uid to ""
            end try
            set out to out & "###S###" & si & "###" & ttyn & "###" & proc & "###" & prof & "###" & uid & "###" & nm & linefeed
          end repeat
        end try
      end repeat
    end try
  end repeat
  return out
end tell
`, appleScriptAppLiteral(appTell))
}

func listWindowsAppleScript(appTell string) string {
	return fmt.Sprintf(`
tell application %s
  set out to ""
  set wi to 0
  repeat with w in windows
    set wi to wi + 1
    try
      set wname to name of w
    on error
      set wname to ""
    end try
    try
      set wid to id of w
    on error
      set wid to 0
    end try
    set out to out & "###W###" & wi & "###" & wname & "###" & wid & linefeed
  end repeat
  return out
end tell
`, appleScriptAppLiteral(appTell))
}

func listTabsAndSessionsAppleScript(windowIndex int, appTell string) string {
	// Use numeric indexes throughout. Nested object references are unstable for
	// path-targeted iTerm instances.
	return fmt.Sprintf(`
tell application %s
  set out to ""
  set target to %d
  set windowCount to count of windows
  if target is less than or equal to windowCount then
    set tabCount to count of tabs of window target
    repeat with ti from 1 to tabCount
      try
        set tname to name of current session of tab ti of window target
      on error
        set tname to ""
      end try
      set out to out & "###T###" & ti & "###" & tname & linefeed
      set sessionCount to count of sessions of tab ti of window target
      repeat with si from 1 to sessionCount
        try
          set nm to name of session si of tab ti of window target
        on error
          set nm to "?"
        end try
        try
          set ttyn to tty of session si of tab ti of window target
        on error
          set ttyn to ""
        end try
        try
          set prof to profile name of session si of tab ti of window target
        on error
          set prof to ""
        end try
        try
          set proc to is processing of session si of tab ti of window target
        on error
          set proc to false
        end try
        try
          set uid to unique ID of session si of tab ti of window target
        on error
          set uid to ""
        end try
        set out to out & "###S###" & si & "###" & ttyn & "###" & proc & "###" & prof & "###" & uid & "###" & nm & linefeed
      end repeat
    end repeat
  end if
  return out
end tell
`, appleScriptAppLiteral(appTell), windowIndex)
}

func parseHierarchy(raw string) ([]SnapshotWindow, []string) {
	var warnings []string
	var windows []SnapshotWindow
	var curW *SnapshotWindow
	var curT *SnapshotTab

	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimRight(row, "\r")
		if row == "" {
			continue
		}
		switch {
		case strings.HasPrefix(row, "###W###"):
			rest := strings.TrimPrefix(row, "###W###")
			// Formats: "idx###name" or "idx###name###windowID"
			parts := strings.SplitN(rest, "###", 3)
			idx, _ := strconv.Atoi(parts[0])
			name := ""
			if len(parts) > 1 {
				name = parts[1]
			}
			var wid uint64
			if len(parts) > 2 {
				if v, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64); err == nil {
					wid = v
				}
			}
			windows = append(windows, SnapshotWindow{Index: idx, Name: name, WindowID: wid})
			curW = &windows[len(windows)-1]
			curT = nil
		case strings.HasPrefix(row, "###T###"):
			if curW == nil {
				warnings = append(warnings, "warning: tab before window in hierarchy")
				continue
			}
			rest := strings.TrimPrefix(row, "###T###")
			idxStr, name, _ := strings.Cut(rest, "###")
			idx, _ := strconv.Atoi(idxStr)
			curW.Tabs = append(curW.Tabs, SnapshotTab{Index: idx, Name: name})
			curT = &curW.Tabs[len(curW.Tabs)-1]
		case strings.HasPrefix(row, "###S###"):
			if curT == nil {
				warnings = append(warnings, "warning: session before tab in hierarchy")
				continue
			}
			rest := strings.TrimPrefix(row, "###S###")
			parts := strings.SplitN(rest, "###", 6)
			if len(parts) < 6 {
				warnings = append(warnings, "warning: malformed session line")
				continue
			}
			si, _ := strconv.Atoi(parts[0])
			proc := parts[2] == "true"
			curT.Sessions = append(curT.Sessions, SnapshotSession{
				Index:             si,
				TTY:               parts[1],
				ItermIsProcessing: proc,
				Profile:           parts[3],
				ID:                parts[4],
				Name:              parts[5],
			})
		}
	}
	return windows, warnings
}

func defaultITermRunning() bool {
	// System Events name check
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to (name of processes) contains "iTerm2"`)
	out, err := cmd.Output()
	if err == nil && strings.Contains(strings.ToLower(string(out)), "true") {
		return true
	}
	// process path fallback
	cmd = exec.Command("pgrep", "-f", "/Applications/iTerm.app/Contents/MacOS/iTerm2")
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}

func defaultRunAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return stdout.String(), nil
}

var psLineRe = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\S+)\s+(\S+)\s+(\d+)\s+([A-Z][a-z]{2}\s+[A-Z][a-z]{2}\s+\d+\s+\d+:\d+:\d+\s+\d+)\s+(.*)$`)

func defaultListProcs(ttyShort string) ([]ProcRow, error) {
	cmd := exec.Command("ps", "-t", ttyShort, "-o", "pid=,ppid=,stat=,etime=,rss=,lstart=,command=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		// no processes is ok
		if stdout.Len() == 0 {
			return nil, nil
		}
	}
	var out []ProcRow
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := psLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[1])
		ppid, _ := strconv.Atoi(m[2])
		rss, _ := strconv.ParseInt(m[5], 10, 64)
		out = append(out, ProcRow{
			PID:     pid,
			PPID:    ppid,
			Stat:    m[3],
			Etime:   m[4],
			RSSKB:   rss,
			Lstart:  m[6],
			Command: m[7],
		})
	}
	return out, nil
}

func defaultListTTYProcs(ttyShorts []string) ([]TTYProc, error) {
	if len(ttyShorts) == 0 {
		return nil, nil
	}
	cmd := exec.Command("ps", "-t", strings.Join(ttyShorts, ","), "-o", "pid=,ppid=,tty=,stat=,command=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var out []TTYProc
	for _, line := range strings.Split(stdout.String(), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 5 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		out = append(out, TTYProc{
			ProcRow: ProcRow{
				PID:     pid,
				PPID:    ppid,
				Stat:    fields[3],
				Command: strings.Join(fields[4:], " "),
			},
			TTY: fields[2],
		})
	}
	return out, nil
}

// defaultListAllProcs runs one global ps and partitions rows by short tty.
func defaultListAllProcs() (map[string][]ProcRow, error) {
	// tty pid ppid stat etime rss command… (lstart omitted — multi-token; etime is enough)
	cmd := exec.Command("ps", "-axo", "tty=,pid=,ppid=,stat=,etime=,rss=,command=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil && stdout.Len() == 0 {
		return nil, err
	}
	byTTY := map[string][]ProcRow{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		tty := fields[0]
		if tty == "??" || tty == "-" || tty == "" {
			continue
		}
		// Normalize /dev/ttys003 → ttys003
		tty = strings.TrimPrefix(tty, "/dev/")
		pid, err1 := strconv.Atoi(fields[1])
		ppid, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			continue
		}
		rss, _ := strconv.ParseInt(fields[5], 10, 64)
		row := ProcRow{
			PID:     pid,
			PPID:    ppid,
			Stat:    fields[3],
			Etime:   fields[4],
			RSSKB:   rss,
			Command: strings.Join(fields[6:], " "),
		}
		byTTY[tty] = append(byTTY[tty], row)
	}
	return byTTY, nil
}

func defaultListCwds(pids []int) (map[int]string, error) {
	if len(pids) == 0 {
		return map[int]string{}, nil
	}
	args := []string{"-a", "-d", "cwd", "-F", "pn"}
	// lsof -p a,b,c
	pidList := make([]string, len(pids))
	for i, p := range pids {
		pidList[i] = strconv.Itoa(p)
	}
	args = append(args, "-p", strings.Join(pidList, ","))
	cmd := exec.Command("lsof", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	_ = cmd.Run() // partial results ok
	result := map[int]string{}
	var cur int
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "p") {
			cur, _ = strconv.Atoi(strings.TrimPrefix(line, "p"))
		} else if strings.HasPrefix(line, "n") && cur != 0 {
			result[cur] = strings.TrimPrefix(line, "n")
		}
	}
	return result, nil
}
