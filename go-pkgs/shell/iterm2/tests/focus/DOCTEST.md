# shell/iterm2 — FindByTTY + Focus (P1 pure iTerm)

## Version

0.0.2

Nested library doctests for **pure iTerm session scan / find-by-TTY / focus**
APIs in `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2`. Does **not** inherit
the parent open-dir `Request`/`Run` from `../DOCTEST.md`.

| Phase | Status |
|-------|--------|
| **P1** `SessionRef` + NormalizeTTY + ParseSessionListOutput + FindByTTY + Build*Script + Focus | Classic TDD — **RED** until implementer lands APIs |

**Out of scope:** agent-run, process tree, session store, CLI, name/title
session_id matching, live iTerm e2e (no L3).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies TTY query strings and/or a `SessionRef` to focus.
- **Session list script** — `BuildSessionListScript` dumps all iTerm sessions
  (window id/name, 1-based tab index, session id, tty, name).
- **Parser** — `ParseSessionListOutput` turns line-oriented TSV dump into
  `[]SessionRef`.
- **TTY normalizer** — `NormalizeTTY` so `ttys148` and `/dev/ttys148` match.
- **Finder** — pure `FindByTTY(refs, ttys)` filters refs (union over query TTYs;
  stable ref order).
- **Focus script builder** — `BuildFocusScript(ref)` activates iTerm, selects
  window by id, selects tab by 1-based index (no create window).
- **Focus runner** — `Focus(ref, cfg)` Execs the focus script via injectable
  `FocusConfig.Exec` (default real osascript).

### Behaviors

- **NormalizeTTY** — bare `ttysN` and `/dev/ttysN` normalize equal; empty stays empty.
- **Parse** — blank lines and `#` comments ignored; empty/whitespace → empty
  slice, nil error; TSV fields map to `SessionRef` (TabIndex 1-based int).
- **FindByTTY** — pure filter; no match → empty; multi query TTYs → union;
  order follows input `refs` (stable). Matching uses NormalizeTTY on both sides.
- **BuildSessionListScript** — AppleScript scans windows/tabs/sessions; emits
  tty; uses ASCII TAB field separator (not bare `tab` inside iTerm tell).
- **BuildFocusScript** — `activate`; select window by id; select tab by index;
  does **not** create a window.
- **Focus** — builds focus script and calls `cfg.Exec(script)`; Exec error
  propagates. Nil/missing Exec uses package default osascript.

### Dump format (session list)

Tab-separated, one session per line (ASCII TAB field sep):

```text
WindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName
```

Blank lines and `#` comments ignored. `SessionIndex` is present in the dump for
scanner completeness; parse may ignore it when filling `SessionRef`.

### Public API (Classic TDD — locked for implementer)

```go
// SessionRef is one iTerm2 session from a scan.
type SessionRef struct {
    WindowID   string // iTerm window id string
    WindowName string // optional
    TabIndex   int    // 1-based tab index in that window
    SessionID  string // iTerm session UUID (optional)
    TTY        string // e.g. /dev/ttys148
    Name       string // session name (optional)
}

func NormalizeTTY(s string) string
func ParseSessionListOutput(stdout string) ([]SessionRef, error)
func BuildSessionListScript() string
func FindByTTY(refs []SessionRef, ttys []string) []SessionRef
func BuildFocusScript(ref SessionRef) string

// FocusConfig injects Exec for Focus (tests + default osascript).
type FocusConfig struct {
    Exec func(script string) error // nil => default osascript
}

func Focus(ref SessionRef, cfg *FocusConfig) error
```

## Decision Tree

```
focus/
├── normalize-tty/                  [Phase=normalize-tty]
│   └── forms/                      bare ↔ /dev equal; empty stays empty
├── session-list/                   [Phase=parse-session-list | build-session-list-script]
│   ├── parse-empty/
│   ├── parse-multi/                fixture multi-line TSV → fields + TabIndex
│   └── build-script/               windows/tabs/sessions + tty + fieldSepAS
├── find-by-tty/                    [Phase=find-by-tty]
│   ├── no-match/
│   ├── one-match/
│   ├── multi-input/                multi query TTYs → union; stable ref order
│   └── normalization/              query ttys148 matches ref /dev/ttys148
└── focus/                          [Phase=build-focus-script | focus]
    ├── build-script/               activate + select window + select tab; no create
    ├── inject-exec/                Focus calls Exec with window id / tab
    └── exec-error/                 Exec error propagates
```

## Test Index

| Leaf | Phase | Description | Expect |
|------|-------|-------------|--------|
| `normalize-tty/forms/` | normalize-tty | bare/`/dev` equal; empty empty | RED |
| `session-list/parse-empty/` | parse-session-list | empty / whitespace / comments → [] | RED |
| `session-list/parse-multi/` | parse-session-list | multi-line fixture fields | RED |
| `session-list/build-script/` | build-session-list-script | scan structure + tty + ASCII TAB sep | RED |
| `find-by-tty/no-match/` | find-by-tty | no TTY match → empty | RED |
| `find-by-tty/one-match/` | find-by-tty | single matching ref | RED |
| `find-by-tty/multi-input/` | find-by-tty | multi TTY queries; union; stable order | RED |
| `find-by-tty/normalization/` | find-by-tty | bare query matches `/dev` ref | RED |
| `focus/build-script/` | build-focus-script | activate + window + tab; no create | RED |
| `focus/inject-exec/` | focus | injectable Exec receives focus script | RED |
| `focus/exec-error/` | focus | Exec error propagates | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/iterm2/tests/focus
doctest test ./shell/iterm2/tests/focus
```

Expect: **RED** (missing package symbols until implementer lands P1 APIs).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// SessionRefInput mirrors iterm2.SessionRef for Request fixtures / Response.
type SessionRefInput struct {
	WindowID   string
	WindowName string
	TabIndex   int
	SessionID  string
	TTY        string
	Name       string
}

// Request selects Phase and inputs for normalize / parse / find / scripts / focus.
type Request struct {
	// Phase:
	//   normalize-tty | parse-session-list | find-by-tty
	//   | build-session-list-script | build-focus-script | focus
	Phase string

	// --- normalize-tty ---
	TTY string

	// --- parse-session-list ---
	ListOutput string

	// --- find-by-tty ---
	Refs      []SessionRefInput
	QueryTTYs []string

	// --- build-focus-script / focus ---
	FocusRef SessionRefInput

	// --- focus inject error ---
	// When non-empty, inject-exec Assert leaves return this error from mock Exec.
	ExecError string
}

// Response holds pure-API results from Run.
type Response struct {
	Script     string
	Normalized string
	Refs       []SessionRefInput
}

// Run dispatches on req.Phase and calls product APIs (Classic TDD — missing
// symbols fail RED until implementer lands them).
//
// Phase "focus": product Focus invoked in leaf Asserts with injectable
// FocusConfig.Exec (package isolation; no live osascript).
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	resp := &Response{}
	switch req.Phase {
	case "normalize-tty":
		resp.Normalized = iterm2.NormalizeTTY(req.TTY)
		return resp, nil
	case "parse-session-list":
		refs, err := iterm2.ParseSessionListOutput(req.ListOutput)
		if err != nil {
			return resp, err
		}
		resp.Refs = sessionRefsToInputs(refs)
		return resp, nil
	case "find-by-tty":
		found := iterm2.FindByTTY(sessionRefsFromInputs(req.Refs), req.QueryTTYs)
		resp.Refs = sessionRefsToInputs(found)
		return resp, nil
	case "build-session-list-script":
		resp.Script = iterm2.BuildSessionListScript()
		return resp, nil
	case "build-focus-script":
		resp.Script = iterm2.BuildFocusScript(sessionRefFromInput(req.FocusRef))
		return resp, nil
	case "focus":
		// Assert drives Focus with injectable Exec
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```
