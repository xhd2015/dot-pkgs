# shell/iterm2 — tab-set create, find, busy, orchestration (P1–P3)

## Version
0.0.3

Nested library doctests for **tab-set** features in
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2`. Does not inherit the parent
open-dir `Request`/`Run` from `../DOCTEST.md`.

| Phase | Status |
|-------|--------|
| **P1** create (`BuildTabSetNewWindowScript`) | Implemented — GREEN |
| **P2** find + busy | Implemented — GREEN |
| **P3** `RunTabSet` / `StatusTabSet` / `StopTabSet` + injectable `TabSetConfig` | Implemented — GREEN |
| **NoSubmit** `TabSpec.NoSubmit` → `write text "…" without newline` | Classic TDD — RED |

New NoSubmit leaves set `NoSubmit` via reflection (`setNoSubmit` in root SETUP)
so packages **compile** today; asserts fail RED until the field/behavior lands.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies `TabSetSpec` and optional run mode / injectors.
- **Tab-set script builder** — `BuildTabSetNewWindowScript` (create window + tabs).
- **Finder** — `BuildTabSetFindScript` + `ParseTabSetFindOutput` → `[]TabSessionRef`.
- **Busy classifier** — `ClassifyBusyFromComm` / injectable per-session busy.
- **Orchestrator** — `RunTabSet`, `StatusTabSet`, `StopTabSet` with injectable
  `TabSetConfig` (Find / Busy / Exec / FrontmostWindowID) so CI needs no live iTerm.
- **Session markers** — `user.koolTabSet`, `user.koolTabSetTab`.

### Behaviors

**P1 create / P2 find+busy** — unchanged (see create/find/busy leaves).

**TabSpec.NoSubmit** (this cycle)

- `NoSubmit bool` on `TabSpec` (default false).
- When true: command write is `write text "cmd" without newline` (stage, no Enter).
- When false/omit: `write text "cmd"` (submit with newline as today).
- **cwd / `cd`:** always write with newline / expression form that executes;
  only the **command** honors NoSubmit.
- Applies to all command write paths: `BuildTabSetNewWindowScript`,
  create-tab-in-window, resend idle.

**RunTabSet** (P3)

- `TabSetRunMode`: Smart (default), NewWindow, NoNewWindow.
- **NewWindow:** always `Exec(BuildTabSetNewWindowScript(spec))`; ignore find.
- **NoNewWindow:** requires `FrontmostWindowID` (or equivalent); else `ErrNoITermWindow`.
- **Smart:**
  - Find empty → create new window script; `CreatedWindow=true`.
  - Find non-empty → pick first WindowID in find order as most recent; if multiple
    distinct WindowIDs → non-empty `Warning` (e.g. mentions "2 windows").
  - Per config tab: missing → create tab + stamp + command (`created`);
    busy/unknown → skip (`skipped-busy` / `skipped-unknown`);
    idle → resend command only via write text (`resent`) — **no Ctrl+C**;
    resend honors NoSubmit.
- Mutating AppleScript goes through `cfg.Exec`; find via `cfg.Find`.

**StatusTabSet** (P3)

- For each config tab id: busy→`running`, idle→`idle`, unknown→`unknown`,
  missing→`missing`. `Instances` = distinct window count; multi-window Warning.

**StopTabSet** (P3)

- Find empty → nil error, zero closed counts, Warning (not running).
- Find non-empty → Exec close scripts (close window and/or tab); counts updated.

### Injection model (`TabSetConfig`)

```
Find(setName) ([]TabSessionRef, error)   // default: osascript find script + parse
Busy(ref TabSessionRef) BusyState          // default: tty probe + ClassifyBusyFromComm
Exec(script string) error                  // default: real osascript
FrontmostWindowID string                   // for NoNewWindow
```

Find order encodes recency: first-seen WindowID is most recent.

## Decision Tree

```
tab-set/
├── (P1 create — Phase=build-tab-set-script)
│   ├── new-window-four-tabs/
│   ├── single-tab/
│   ├── stamps-set-and-tab-vars/
│   ├── sets-session-names/
│   ├── optional-cwd/
│   ├── window-name/
│   ├── command-escape/
│   ├── no-submit-true/             NoSubmit → without newline              [RED]
│   ├── default-submit/             default submit form (no without newline)
│   └── cwd-and-no-submit/          cd executes; command without newline    [RED]
├── busy/                           [P2 classify-busy]
│   ├── idle-shell/
│   ├── busy-child/
│   └── unknown-empty/
├── find/                           [P2]
│   ├── script-scans-vars/
│   ├── parse-two-sessions/
│   └── parse-empty/
├── run/                            [P3 run-tab-set]
│   ├── smart-first-create/         empty find → create window script
│   ├── new-window-mode/            Mode NewWindow always create
│   ├── smart-skip-busy/            busy → skipped-busy, no resend
│   ├── smart-resend-idle/          idle → resent write text
│   ├── smart-resend-idle-no-submit/ resend without newline               [RED]
│   ├── smart-recreate-missing/     missing tab → create tab
│   ├── smart-recreate-no-submit/   create-tab without newline              [RED]
│   ├── smart-multi-window-warn/    2 windows → Warning; one window synced
│   ├── no-new-window-missing-front/ NoNewWindow + no front → ErrNoITermWindow
│   └── no-ctrl-c/                  no control-c / ctrl-c keystroke in scripts
├── status/                         [P3 status-tab-set]
│   └── mixed-states/               running/idle/missing/unknown
└── stop/                           [P3 stop-tab-set]
    ├── empty/                      not running warning, 0 closed
    └── closes-marked/              Exec close scripts for found sessions
```

## Test Index

| Leaf | Phase | Description | Expect |
|------|-------|-------------|--------|
| `new-window-four-tabs/` | build-tab-set-script | P1 create structure | GREEN |
| `single-tab/` | build-tab-set-script | P1 single tab | GREEN |
| `stamps-set-and-tab-vars/` | build-tab-set-script | P1 markers | GREEN |
| `sets-session-names/` | build-tab-set-script | P1 session names | GREEN |
| `optional-cwd/` | build-tab-set-script | P1 optional cwd | GREEN |
| `window-name/` | build-tab-set-script | P1 window name | GREEN |
| `command-escape/` | build-tab-set-script | P1 command escape | GREEN |
| `no-submit-true/` | build-tab-set-script | NoSubmit → without newline | RED |
| `default-submit/` | build-tab-set-script | default submit (no without newline) | GREEN |
| `cwd-and-no-submit/` | build-tab-set-script | cd executes; cmd without newline | RED |
| `busy/idle-shell/` | classify-busy | P2 idle shells | GREEN |
| `busy/busy-child/` | classify-busy | P2 busy non-shell | GREEN |
| `busy/unknown-empty/` | classify-busy | P2 unknown | GREEN |
| `find/script-scans-vars/` | build-find-script | P2 find script | GREEN |
| `find/parse-two-sessions/` | parse-find | P2 parse two | GREEN |
| `find/parse-empty/` | parse-find | P2 parse empty | GREEN |
| `run/smart-first-create/` | run-tab-set | empty find → create | GREEN |
| `run/new-window-mode/` | run-tab-set | always new window | GREEN |
| `run/smart-skip-busy/` | run-tab-set | skip busy tab | GREEN |
| `run/smart-resend-idle/` | run-tab-set | resend idle command | GREEN |
| `run/smart-resend-idle-no-submit/` | run-tab-set | resend without newline | RED |
| `run/smart-recreate-missing/` | run-tab-set | recreate missing tab | GREEN |
| `run/smart-recreate-no-submit/` | run-tab-set | create-tab without newline | RED |
| `run/smart-multi-window-warn/` | run-tab-set | multi-window warning | GREEN |
| `run/no-new-window-missing-front/` | run-tab-set | ErrNoITermWindow | GREEN |
| `run/no-ctrl-c/` | run-tab-set | no Ctrl+C in scripts | GREEN |
| `status/mixed-states/` | status-tab-set | mixed tab states | GREEN |
| `stop/empty/` | stop-tab-set | empty stop warning | GREEN |
| `stop/closes-marked/` | stop-tab-set | close marked sessions | GREEN |

## How to Run

```sh
# from go-pkgs module root (main repo or wrk --bring external path)
doctest vet ./shell/iterm2/tests/tab-set
doctest test ./shell/iterm2/tests/tab-set
```

Expect: existing leaves **GREEN**; new NoSubmit leaves **RED** (missing field /
without-newline until implementer).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// TabSpecInput mirrors iterm2.TabSpec for leaf Setup.
// NoSubmit is documented for leaves; root Run does not map it until product
// TabSpec.NoSubmit exists (new leaves set the field in Assert instead).
type TabSpecInput struct {
	ID       string
	Name     string
	Command  string
	Cwd      string
	NoSubmit bool
}

// SessionRefInput mirrors iterm2.TabSessionRef for injectable Find fixtures.
type SessionRefInput struct {
	SetName   string
	TabID     string
	WindowID  string
	SessionID string
	TTY       string
}

// Request selects Phase and inputs for create / find / busy / orchestration.
type Request struct {
	// Phase:
	//   build-tab-set-script | classify-busy | build-find-script | parse-find
	//   | run-tab-set | status-tab-set | stop-tab-set
	Phase string

	// --- P1 create / shared set name ---
	TabSetName string
	WindowName string
	Tabs       []TabSpecInput

	// --- P2 busy ---
	FgComm string
	FgOK   bool

	// --- P2 find parse ---
	FindOutput string

	// --- P3 orchestration ---
	// RunMode: "smart" | "new-window" | "no-new-window" (maps to TabSetRunMode).
	RunMode string
	// FindSessions is the injectable Find result (order = recency).
	FindSessions []SessionRefInput
	// BusyByTab maps TabID → "idle" | "busy" | "unknown".
	BusyByTab map[string]string
	// FrontmostWindowID for NoNewWindow mode.
	FrontmostWindowID string
}

// Response holds script from P1 Run; P2/P3 Asserts call product APIs directly.
type Response struct {
	Script string
}

// Run dispatches on req.Phase.
//
// P1 (implemented): BuildTabSetNewWindowScript
// P2 (implemented): ClassifyBusyFromComm, BuildTabSetFindScript, ParseTabSetFindOutput
//   (still invoked from P2 Asserts for historical package isolation).
//
// P3 public API to pin (Classic TDD — product calls live in run|status|stop Asserts):
//
//	type TabSetRunMode int
//	const (
//	    TabSetRunSmart TabSetRunMode = iota
//	    TabSetRunNewWindow
//	    TabSetRunNoNewWindow
//	)
//
//	type TabSetConfig struct {
//	    Find              func(setName string) ([]TabSessionRef, error)
//	    Busy              func(ref TabSessionRef) BusyState
//	    Exec              func(script string) error
//	    FrontmostWindowID string
//	}
//
//	type TabSetRunOptions struct { Mode TabSetRunMode }
//
//	type TabRunResult struct {
//	    TabID  string
//	    Action string // "created" | "resent" | "skipped-busy" | "skipped-unknown"
//	}
//	type TabSetRunResult struct {
//	    CreatedWindow bool
//	    FocusedWindow string
//	    Warning       string
//	    Tabs          []TabRunResult
//	}
//	func RunTabSet(spec TabSetSpec, opts TabSetRunOptions, cfg *TabSetConfig) (*TabSetRunResult, error)
//
//	type TabStatusEntry struct {
//	    TabID string
//	    State string // "running" | "idle" | "missing" | "unknown"
//	}
//	type TabSetStatus struct {
//	    SetName, WindowID, WindowName, Warning string
//	    Instances int
//	    Tabs []TabStatusEntry
//	}
//	func StatusTabSet(spec TabSetSpec, cfg *TabSetConfig) (*TabSetStatus, error)
//
//	type TabSetStopResult struct {
//	    ClosedWindows, ClosedTabs int
//	    Warning string
//	}
//	func StopTabSet(setName string, cfg *TabSetConfig) (*TabSetStopResult, error)
//
//	var ErrNoITermWindow error // NoNewWindow without frontmost window
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	resp := &Response{}
	switch req.Phase {
	case "build-tab-set-script":
		tabs := make([]iterm2.TabSpec, len(req.Tabs))
		for i, tab := range req.Tabs {
			tabs[i] = iterm2.TabSpec{
				ID:      tab.ID,
				Name:    tab.Name,
				Command: tab.Command,
				Cwd:     tab.Cwd,
			}
		}
		resp.Script = iterm2.BuildTabSetNewWindowScript(iterm2.TabSetSpec{
			Name:       req.TabSetName,
			WindowName: req.WindowName,
			Tabs:       tabs,
		})
		return resp, nil
	case "classify-busy", "build-find-script", "parse-find":
		// P2: product invoked in leaf Asserts
		return resp, nil
	case "run-tab-set", "status-tab-set", "stop-tab-set":
		// P3: product invoked in leaf Asserts (package isolation for Classic TDD)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```
