package snapshot

// Snapshot is a full inventory of iTerm2 windows, tabs, and sessions (panes).
type Snapshot struct {
	CapturedAt string           `json:"captured_at"`
	Host       string           `json:"host"`
	Source     string           `json:"source"`
	Summary    SnapshotSummary  `json:"summary"`
	Windows    []SnapshotWindow `json:"windows"`
}

// SnapshotSummary counts windows/tabs/sessions and idle/busy/unknown.
type SnapshotSummary struct {
	Windows  int `json:"windows"`
	Tabs     int `json:"tabs"`
	Sessions int `json:"sessions"`
	Idle     int `json:"idle"`
	Busy     int `json:"busy"`
	Unknown  int `json:"unknown"`
}

// SnapshotWindow is one iTerm2 window.
type SnapshotWindow struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	// WindowID is the iTerm/AppleScript window id (CG window number when available).
	// Zero means unknown.
	WindowID uint64 `json:"window_id,omitempty"`
	// FixedSpace, when non-nil, is the resolved 0-based Space for fixtures or
	// after a space-first capture pin. Prefer this over ResolveSpace so tests
	// need no live Mission Control / CGS.
	FixedSpace *int `json:"-"`
	// App is the canonical iTerm install that owns this window (AppTag stamp or
	// multi-app source). Not emitted on typical JSON inventories.
	App  string        `json:"-"`
	Tabs []SnapshotTab `json:"tabs"`
}

// SnapshotTab is one tab; sessions are panes within the tab.
type SnapshotTab struct {
	Index    int               `json:"index"`
	Name     string            `json:"name"`
	Sessions []SnapshotSession `json:"sessions"`
}

// SnapshotSession is one pane/session. ID is the iTerm2 session unique ID (UUID).
// No Agent field — agent enrich lives in a later agent-pro phase.
type SnapshotSession struct {
	Index             int            `json:"index"`
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	TTY               string         `json:"tty"`
	Profile           string         `json:"profile"`
	ItermIsProcessing bool           `json:"iterm_is_processing"`
	Idle              *bool          `json:"idle"` // nil = unknown
	Cwd               *string        `json:"cwd"`
	ShellPID          *int           `json:"shell_pid"`
	PID               *int           `json:"pid"`
	PPID              *int           `json:"ppid"`
	Stat              *string        `json:"stat"`
	Command           *string        `json:"command"`
	CommandLine       *string        `json:"command_line"`
	StartTime         *string        `json:"start_time"`
	StartTimeUnix     *int64         `json:"start_time_unix"`
	DurationSeconds   *int64         `json:"duration_seconds"`
	Duration          *string        `json:"duration"`
	Etime             *string        `json:"etime"`
	RSSKB             *int64         `json:"rss_kb"`
	Processes         []SnapshotProc `json:"processes"`
	// Layout hints (not required for id resolution).
	WindowIndex int `json:"window_index,omitempty"`
	TabIndex    int `json:"tab_index,omitempty"`
}

// SnapshotProc is one process observed on a session tty.
type SnapshotProc struct {
	PID             int     `json:"pid"`
	PPID            int     `json:"ppid"`
	Stat            string  `json:"stat"`
	Etime           string  `json:"etime"`
	DurationSeconds int64   `json:"duration_seconds"`
	Duration        string  `json:"duration"`
	StartTime       *string `json:"start_time"`
	StartTimeUnix   *int64  `json:"start_time_unix"`
	RSSKB           int64   `json:"rss_kb"`
	Command         string  `json:"command"`
}

// ProcRow is a raw process row returned by Collector.ListProcs (ps-like).
type ProcRow struct {
	PID     int
	PPID    int
	Stat    string
	Etime   string
	RSSKB   int64
	Lstart  string
	Command string
}

// CaptureOpts controls core inventory capture (not agent attach).
// Agent-only flags (e.g. NoEnrich) live in agent-pro, not here.
type CaptureOpts struct {
	// IncludeCwd runs lsof to fill session Cwd. Default false (fast path).
	IncludeCwd bool
	// SpaceAllow, when non-empty, enables space-first filtering: resolve each
	// window's Space (FixedSpace or ResolveSpace(WindowID)) and skip
	// ListTabsAndSessions + enrich when the index is not in the allowlist.
	SpaceAllow []int
	// SpaceSkipped, when non-nil, receives the count of window headers skipped
	// by SpaceAllow (not deep-captured).
	SpaceSkipped *int
}

func boolPtr(v bool) *bool    { return &v }
func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }
func strPtr(v string) *string { return &v }
