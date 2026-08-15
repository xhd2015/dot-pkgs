# shell/iterm2 — title get/set scripts and API

Library-level doctests for iTerm2 session/window title helpers used by
`kool iterm2 set-title` / `get-title`. Nested root (does not inherit parent
open-dir `Request`/`Run` from `../DOCTEST.md`).

## Version

0.0.2

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies session identity (`ITERM_SESSION_ID` or UUID), title
  target (session vs window), and optional new title string.
- **Session detector** — treats empty/missing `ITERM_SESSION_ID` as not in
  iTerm2.
- **Script builder** — `BuildSetTitleScript` / `BuildGetTitleScript` emit
  AppleScript that locates the session by UUID and reads/writes session `name`
  or the containing window's name.
- **Escaper** — embeds titles with the same `\` / `"` escaping as path/command
  helpers (`EscapePathForAppleScript` or a shared title escaper).
- **SetTitle / GetTitle** — high-level API: env check, build script, run
  osascript (injectable), return old/new or current title.

### Behaviors

- **Build set (session)** — script contains session UUID and sets session name
  to the escaped title.
- **Build set (window)** — script contains UUID and sets window name.
- **Build get (session|window)** — script returns the corresponding name.
- **Escaping** — `"` → `\"`, `\` → `\\` inside AppleScript string literals.
- **API not-in-session** — error when `ITERM_SESSION_ID` empty.
- **API empty title on set** — error before osascript.

## Decision Tree

```
title/
├── script/                         [Phase=build-*]
│   ├── set-session/
│   ├── set-window/
│   ├── get-session/
│   └── get-window/
├── escaping/                       [Phase=escape-title]
│   └── title-quotes/
└── api/                            [Phase=set-title|get-title]
    ├── set-not-in-session/
    ├── set-empty-title/
    └── get-not-in-session/
```

## Test Index

| Leaf | Phase | Description |
|------|-------|-------------|
| `script/set-session/` | build-set-title | Script sets session name for UUID |
| `script/set-window/` | build-set-title | Script sets window name for UUID |
| `script/get-session/` | build-get-title | Script reads session name |
| `script/get-window/` | build-get-title | Script reads window name |
| `escaping/title-quotes/` | escape-title | Escapes `"` and `\` in titles |
| `api/set-not-in-session/` | set-title | Empty ITERM_SESSION_ID → error |
| `api/set-empty-title/` | set-title | Empty title → error |
| `api/get-not-in-session/` | get-title | Empty ITERM_SESSION_ID → error |

## How to Run

```sh
doctest vet ./external/dot-pkgs-master-2026-07-09/go-pkgs/shell/iterm2/tests/title
doctest test ./external/dot-pkgs-master-2026-07-09/go-pkgs/shell/iterm2/tests/title
```

```go
import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const defaultTitleSessionID = "w0t0p0:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

// Request selects library phase and inputs for title helpers.
type Request struct {
	// Phase: build-set-title | build-get-title | escape-title | set-title | get-title
	Phase string
	// SessionID is full ITERM_SESSION_ID (w…:UUID) when needed.
	SessionID string
	// Target is "session" (default) or "window".
	Target string
	// Title for set / escape inputs.
	Title string
	// ClearSessionEnv strips ITERM_SESSION_ID for API error leaves.
	ClearSessionEnv bool
	// EscapeInput for Phase=escape-title.
	EscapeInput string
}

// Response captures script/API results.
type Response struct {
	Script  string
	Escaped string
	Old     string
	New     string
	Title   string
}

func titleTarget(req *Request) iterm2.TitleTarget {
	if req.Target == "window" {
		return iterm2.TitleTargetWindow
	}
	return iterm2.TitleTargetSession
}

func sessionIDOrDefault(req *Request) string {
	if req.SessionID != "" {
		return req.SessionID
	}
	return defaultTitleSessionID
}

func sessionUUIDFromID(sid string) string {
	if i := strings.Index(sid, ":"); i >= 0 && i+1 < len(sid) {
		return sid[i+1:]
	}
	return sid
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Phase {
	case "build-set-title":
		sid := sessionIDOrDefault(req)
		resp.Script = iterm2.BuildSetTitleScript(sid, titleTarget(req), req.Title)
		return resp, nil
	case "build-get-title":
		sid := sessionIDOrDefault(req)
		resp.Script = iterm2.BuildGetTitleScript(sid, titleTarget(req))
		return resp, nil
	case "escape-title":
		// Prefer dedicated title escaper if present; path escaper has same rules.
		resp.Escaped = iterm2.EscapePathForAppleScript(req.EscapeInput)
		return resp, nil
	case "set-title":
		if req.ClearSessionEnv {
			prev, had := os.LookupEnv("ITERM_SESSION_ID")
			_ = os.Unsetenv("ITERM_SESSION_ID")
			t.Cleanup(func() {
				if had {
					_ = os.Setenv("ITERM_SESSION_ID", prev)
				}
			})
		} else {
			sid := sessionIDOrDefault(req)
			prev, had := os.LookupEnv("ITERM_SESSION_ID")
			_ = os.Setenv("ITERM_SESSION_ID", sid)
			t.Cleanup(func() {
				if had {
					_ = os.Setenv("ITERM_SESSION_ID", prev)
				} else {
					_ = os.Unsetenv("ITERM_SESSION_ID")
				}
			})
		}
		old, newTitle, err := iterm2.SetTitle(req.Title, titleTarget(req))
		resp.Old = old
		resp.New = newTitle
		return resp, err
	case "get-title":
		if req.ClearSessionEnv {
			prev, had := os.LookupEnv("ITERM_SESSION_ID")
			_ = os.Unsetenv("ITERM_SESSION_ID")
			t.Cleanup(func() {
				if had {
					_ = os.Setenv("ITERM_SESSION_ID", prev)
				}
			})
		} else {
			sid := sessionIDOrDefault(req)
			prev, had := os.LookupEnv("ITERM_SESSION_ID")
			_ = os.Setenv("ITERM_SESSION_ID", sid)
			t.Cleanup(func() {
				if had {
					_ = os.Setenv("ITERM_SESSION_ID", prev)
				} else {
					_ = os.Unsetenv("ITERM_SESSION_ID")
				}
			})
		}
		title, err := iterm2.GetTitle(titleTarget(req))
		resp.Title = title
		return resp, err
	default:
		return nil, errors.New("unknown phase: " + req.Phase)
	}
}
```
