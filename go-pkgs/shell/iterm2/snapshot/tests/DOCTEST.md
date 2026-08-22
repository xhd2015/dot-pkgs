# shell/iterm2/snapshot — injectable iTerm2 hierarchy + process enrich

Classic TDD doctests for plan phase **P1**: package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot` — **RED** until the
implementer extracts core inventory + process enrich from kool `tools/iterm2`
(minus agent types).

Injectable capture of iTerm2 windows/tabs/sessions with process enrich
(idle/busy/unknown, pid, cwd). **No agent / procresolve / SessionAgent types.**

**Out of scope (later phases):** agent-pro attach, kool rewire, kck live list,
save/restore, CLI render (HTML/MD).

## Version

0.0.2

# DSN (Domain Specific Notion)

Library package for callers that need a full iTerm2 session inventory with
process enrichment, without agent tooling or process-global collectors.

### Participants

- **Caller** — library client that builds or injects a **Collector** and runs
  **Capture** / **CaptureWith** (or package-level **Capture** convenience).
- **Collector** — constructor-style holder of injectables: **RunAppleScript**,
  **ListProcs**, **ListCwds**, **ITermRunning**, **Now**, **Hostname**. Each
  test owns its instance (parallel-safe; no required process-global inject).
- **Phased fixture** — **ApplyPhasedFixture** on a Collector configures
  hierarchy windows + idle/busy tty process rows + cwd map without real
  `osascript` / `ps` / `lsof`.
- **Capture pipeline** — if iTerm not running → hard error; else list windows →
  tabs/sessions per window → enrich each session from procs/cwds → summary
  idle/busy/unknown counts; soft **warnings** on partial probe failures.
- **Snapshot model** — **Snapshot** / **SnapshotWindow** / **SnapshotTab** /
  **SnapshotSession** / **SnapshotProc** / **SnapshotSummary**. Session has
  **no** Agent field.

### Behaviors

- **iTerm gate** — `ITermRunning() == false` → error containing
  `iTerm2 is not running`; no snapshot.
- **Empty hierarchy** — zero windows → snapshot with zero summary counts;
  `Source == "iterm2"`; host/time from injects.
- **Idle** — shell-only (login ignored) on tty → `Idle=true`, summary Idle++.
- **Busy** — non-shell child present → `Idle=false`, summary Busy++; chosen
  process is the work/foreground leaf.
- **Unknown** — no processes on tty → `Idle=nil`, summary Unknown++; soft
  warning about no processes.
- **Cwd** — from **ListCwds** for chosen or shell pid.
- **Soft warnings** — ListProcs/ListCwds errors append warnings and continue
  (do not fail Capture).
- **CaptureWith** — supports **IncludeCwd** and **SpaceAllow** (space-first
  gate); zero opts match **Capture**. Agent `NoEnrich` lives in agent-pro.
- **AppTell / AppTag** — parameterize AppleScript tell target; stamp **App**
  when empty.
- **EnrichFromProcs / ListTTYProcs** — exported for live-scan style callers.

## Decision Tree

```text
snapshot/tests/
├── not-running/                 # ITermRunning=false → hard error
├── hierarchy/                   # window/tab/session structure
│   ├── empty-windows/           # no windows → zero summary
│   ├── single-idle/             # 1w/1t/1 idle shell session
│   └── multi-window-tab/        # multi window/tab indices preserved
├── enrich/                      # process classification + cwd
│   ├── busy-session/            # non-shell child → busy
│   ├── cwd-attached/            # ListCwds path on session
│   └── no-procs-unknown/        # empty procs → Idle=nil + warning
├── space-filter/                # SpaceAllow gate (FixedSpace; no live CGS)
│   ├── keeps-allowed/           # allowlist keeps matching window
│   └── skips-disallowed/        # other FixedSpace skipped + SpaceSkipped
├── app-tag/                     # AppTag stamped on fixture windows
│   └── stamps-empty-app/
├── warnings/
│   └── list-procs-error/        # ListProcs error → soft warning, continues
└── api/
    └── capture-with-zero-opts/  # CaptureWith({}) ≡ Capture on injects
```

Parameter ranking (most → least significant):

1. **iTerm availability** — not running (error) vs running (snapshot)
2. **SpaceAllow filter** — keep vs skip deep-capture (when set)
3. **Hierarchy shape** — empty / single / multi window-tab
4. **Enrich class** — idle / busy / unknown / cwd attach
5. **AppTag stamp** — empty App filled from collector
6. **Soft failure** — probe errors → warnings
7. **API surface** — Capture vs CaptureWith

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `not-running` | iTerm not running → error, nil snapshot |
| 2 | `hierarchy/empty-windows` | empty fixture → zero summary counts, Source=iterm2 |
| 3 | `hierarchy/single-idle` | one idle shell session; summary Idle=1 |
| 4 | `hierarchy/multi-window-tab` | multi window/tab hierarchy indices preserved |
| 5 | `enrich/busy-session` | non-shell child → Busy=1, Idle=false |
| 6 | `enrich/cwd-attached` | Cwd from ListCwds for session tty |
| 7 | `enrich/no-procs-unknown` | empty ListProcs → Unknown=1 + soft warning |
| 8 | `warnings/list-procs-error` | ListProcs error → soft warning; Capture succeeds |
| 9 | `api/capture-with-zero-opts` | CaptureWith(zero) same summary as Capture path |
| 10 | `space-filter/keeps-allowed` | SpaceAllow keeps FixedSpace-matching window |
| 11 | `space-filter/skips-disallowed` | non-matching FixedSpace skipped; SpaceSkipped=1 |
| 12 | `app-tag/stamps-empty-app` | AppTag fills empty Window.App |

## How to Run

```sh
# from go-pkgs module root
cd /Users/xhd2015/Projects/xhd2015/kck/external/dot-pkgs-master-2026-08-06-1/go-pkgs

doctest vet ./shell/iterm2/snapshot/tests
doctest test ./shell/iterm2/snapshot/tests

doctest test -v ./shell/iterm2/snapshot/tests/not-running
doctest test -v ./shell/iterm2/snapshot/tests/hierarchy/single-idle
doctest test -v ./shell/iterm2/snapshot/tests/enrich/busy-session
```

Classic TDD: expect **RED** (compile or assert failure) until
`go-pkgs/shell/iterm2/snapshot` production package exists and implements the
locked API below.

## Locked public API (implementer contract)

Import path: `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot`

```go
package snapshot

// --- model (no Agent / SessionAgent / AgentTreeNode) ---

type Snapshot struct {
	CapturedAt string
	Host       string
	Source     string // production: "iterm2"
	Summary    SnapshotSummary
	Windows    []SnapshotWindow
}

type SnapshotSummary struct {
	Windows  int
	Tabs     int
	Sessions int
	Idle     int
	Busy     int
	Unknown  int
}

type SnapshotWindow struct {
	Index      int
	Name       string
	WindowID   uint64
	FixedSpace *int  // json:"-"; fixture / space-first pin
	App        string // json:"-"; AppTag stamp / multi-app
	Tabs       []SnapshotTab
}

type SnapshotTab struct {
	Index    int
	Name     string
	Sessions []SnapshotSession
}

// SnapshotSession has no Agent field (agent enrich is agent-pro phase).
type SnapshotSession struct {
	Index             int
	ID                string
	Name              string
	TTY               string
	Profile           string
	ItermIsProcessing bool
	Idle              *bool // nil = unknown
	Cwd               *string
	ShellPID          *int
	PID               *int
	PPID              *int
	Stat              *string
	Command           *string
	CommandLine       *string
	StartTime         *string
	StartTimeUnix     *int64
	DurationSeconds   *int64
	Duration          *string
	Etime             *string
	RSSKB             *int64
	Processes         []SnapshotProc
	WindowIndex       int
	TabIndex          int
}

type SnapshotProc struct {
	PID             int
	PPID            int
	Stat            string
	Etime           string
	DurationSeconds int64
	Duration        string
	StartTime       *string
	StartTimeUnix   *int64
	RSSKB           int64
	Command         string
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

// CaptureOpts controls core inventory options (IncludeCwd, SpaceAllow).
// Agent-only flags (e.g. NoEnrich) live in agent-pro, not here.
type CaptureOpts struct {
	IncludeCwd   bool
	SpaceAllow   []int
	SpaceSkipped *int
}

// Collector gathers hierarchy + process enrichment. Fields may be overridden
// per instance (parallel-safe; no process-global required for production or tests).
type Collector struct {
	RunAppleScript func(script string) (string, error)
	ListProcs      func(ttyShort string) ([]ProcRow, error)
	ListAllProcs   func() (map[string][]ProcRow, error)
	ListTTYProcs   func(ttyShorts []string) ([]TTYProc, error)
	ListCwds       func(pids []int) (map[int]string, error)
	ITermRunning   func() bool
	Now            func() time.Time
	Hostname       func() (string, error)
	ResolveSpace   func(windowID uint64) (int, error)
	AppTell        string
	AppTag         string
}

// NewCollector returns a Collector with production defaults (live osascript/ps/lsof).
func NewCollector() *Collector

// Capture runs phased hierarchy collection + process enrichment.
func (c *Collector) Capture() (*Snapshot, []string /*warnings*/, error)

// CaptureWith runs Capture with options (IncludeCwd, SpaceAllow, …).
func (c *Collector) CaptureWith(opts CaptureOpts) (*Snapshot, []string, error)

// CaptureProgressiveWith is progressive Capture with CaptureOpts.
func (c *Collector) CaptureProgressiveWith(opts CaptureOpts, onWindowReady func(SnapshotWindow) error) (*Snapshot, []string, error)

// EnrichFromProcs classifies idle/busy and picks the primary process row.
func EnrichFromProcs(procs []ProcRow, cwds map[int]string, now time.Time) (
	idle *bool, shellPID *int, chosen *ProcRow, cwd *string, snapProcs []SnapshotProc,
)

// Capture is NewCollector().Capture() convenience for production callers.
func Capture() (*Snapshot, []string, error)

type TTYProc struct {
	ProcRow
	TTY string
}

// PhasedFixtureOpts configures ApplyPhasedFixture (tests / inject path).
type PhasedFixtureOpts struct {
	Windows       []SnapshotWindow
	ITermRunning  bool
	IdleTTYs      []string            // short tty names classified idle (shell only)
	BusyTTYs      []string            // short tty names classified busy
	BusyLeafByTTY map[string]string   // optional busy leaf command override
	CwdByTTY      map[string]string   // cwd for processes on that short tty
	Now           time.Time
	Hostname      string
}

// ApplyPhasedFixture configures this Collector for fixture hierarchy + process
// enrich without real AppleScript/ps/lsof. Mutates c only (parallel-safe when
// each test owns its Collector). No process-global collector mutation.
func (c *Collector) ApplyPhasedFixture(opts PhasedFixtureOpts)
```

Behavior parity notes (from kool core, minus agent):

- Error text for not-running includes `iTerm2 is not running`.
- Summary counts windows/tabs/sessions and Idle/Busy/Unknown over sessions.
- Soft warnings when ps/cwd probe fails or no processes on a tty.
- CapturedAt formatted from `Now`; Host from `Hostname` inject (short hostname OK).

```go
import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// Request is filled root→leaf. Each leaf owns its own Collector via Run
// (parallel-safe inject; no package globals).
type Request struct {
	// ITermRunning controls PhasedFixtureOpts.ITermRunning (default false until set).
	ITermRunning bool

	// Windows is the phased hierarchy fixture (tabs/sessions pre-structure).
	Windows []snapshot.SnapshotWindow

	// IdleTTYs / BusyTTYs are short tty names for process classification.
	IdleTTYs []string
	BusyTTYs []string
	// BusyLeafByTTY overrides default busy leaf command per short tty.
	BusyLeafByTTY map[string]string
	// CwdByTTY sets cwd returned for processes on that short tty.
	CwdByTTY map[string]string

	// Fixed clock/host for deterministic CapturedAt / Host.
	Now      time.Time
	Hostname string

	// UseCaptureWith: call CaptureWith instead of Capture (also required when
	// SpaceAllow / AppTag-only paths need CaptureOpts).
	UseCaptureWith bool

	// SpaceAllow / AppTag forwarded onto Collector + CaptureOpts.
	SpaceAllow []int
	AppTag     string

	// ListProcsMode overrides ListProcs after fixture apply:
	//   ""       — leave fixture ListProcs
	//   "empty"  — always return nil, nil
	//   "error"  — always return nil, fmt.Errorf(ListProcsErr)
	ListProcsMode string
	ListProcsErr  string
}

// Response observes Capture outcomes.
type Response struct {
	Snap         *snapshot.Snapshot
	Warnings     []string
	SpaceSkipped int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	now := req.Now
	if now.IsZero() {
		now = time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	host := req.Hostname
	if host == "" {
		host = "testhost"
	}

	c := snapshot.NewCollector()
	c.ApplyPhasedFixture(snapshot.PhasedFixtureOpts{
		Windows:       req.Windows,
		ITermRunning:  req.ITermRunning,
		IdleTTYs:      req.IdleTTYs,
		BusyTTYs:      req.BusyTTYs,
		BusyLeafByTTY: req.BusyLeafByTTY,
		CwdByTTY:      req.CwdByTTY,
		Now:           now,
		Hostname:      host,
	})
	if req.AppTag != "" {
		c.AppTag = req.AppTag
	}

	switch req.ListProcsMode {
	case "":
		// fixture ListProcs
	case "empty":
		c.ListProcs = func(ttyShort string) ([]snapshot.ProcRow, error) {
			return nil, nil
		}
	case "error":
		msg := req.ListProcsErr
		if msg == "" {
			msg = "ps failed"
		}
		c.ListProcs = func(ttyShort string) ([]snapshot.ProcRow, error) {
			return nil, fmt.Errorf("%s", msg)
		}
	default:
		t.Fatalf("unknown ListProcsMode %q", req.ListProcsMode)
	}

	var (
		snap    *snapshot.Snapshot
		warn    []string
		err     error
		skipped int
	)
	opts := snapshot.CaptureOpts{}
	if len(req.SpaceAllow) > 0 {
		opts.SpaceAllow = req.SpaceAllow
		opts.SpaceSkipped = &skipped
		req.UseCaptureWith = true
	}
	if req.UseCaptureWith {
		snap, warn, err = c.CaptureWith(opts)
	} else {
		snap, warn, err = c.Capture()
	}
	return &Response{Snap: snap, Warnings: warn, SpaceSkipped: skipped}, err
}

// --- assert helpers used by leaves ---

func mustSnap(t *testing.T, resp *Response, err error) *snapshot.Snapshot {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	return resp.Snap
}

func assertSummary(t *testing.T, got snapshot.SnapshotSummary, want snapshot.SnapshotSummary) {
	t.Helper()
	if got != want {
		t.Fatalf("summary got=%+v want=%+v", got, want)
	}
}

func boolVal(p *bool) (v bool, ok bool) {
	if p == nil {
		return false, false
	}
	return *p, true
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func warningHas(warns []string, substr string) bool {
	for _, w := range warns {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}
```
