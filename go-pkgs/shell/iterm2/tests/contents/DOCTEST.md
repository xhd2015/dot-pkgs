# shell/iterm2 — session contents dump

Library-level doctests for `BuildContentsScript` / `Contents`. Nested root
(does not inherit parent open-dir `Request`/`Run`).

## Version

0.0.1

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies iTerm session UUID (or `ITERM_SESSION_ID`).
- **Script builder** — `BuildContentsScript` emits AppleScript that locates
  the session by UUID and returns `contents of` that session.
- **Contents** — tries running installs (env path, home, system), first hit wins.

### Behaviors

- Script returns `contents of aSession`; no `activate` / `select`.
- Empty session id → error before osascript.
- Injected Exec body → Contents returns that body.
- Injected not-found on home then hit on system → system result + home tag skip.
- Not running home is skipped (no Exec for that path).
- All misses → `session not found`.

## Decision Tree

```
contents/
├── script/
│   ├── no-focus/
│   └── quote-escape/
├── api/
│   ├── empty-id/
│   ├── inject-body/
│   ├── home-miss-system-hit/
│   ├── skip-not-running/
│   └── not-found/
```

## How to Run

```sh
doctest test ./go-pkgs/shell/iterm2/tests/contents
```

```go
import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const defaultContentsSessionID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

type Request struct {
	Phase     string
	SessionID string
	AppPath   string
	// Injected stdout when Phase=contents and not not-found.
	Body string
	// NotFoundApps: tell-path substrings that should return session not found.
	NotFoundApps []string
	// Running: abs paths treated as running; nil → all running.
	Running []string
	// Home / env / existing apps for search.
	Home    string
	EnvPath string
	HomeApp bool
	SysApp  bool
}

type Response struct {
	Script    string
	SessionID string
	App       string
	Contents  string
	ExecN     int
	Scripts   []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	sid := req.SessionID
	if sid == "" {
		sid = defaultContentsSessionID
	}
	resp := &Response{}
	switch req.Phase {
	case "build-script":
		resp.Script = iterm2.BuildContentsScript(sid, req.AppPath)
		return resp, nil
	case "contents":
		home := req.Home
		if home == "" {
			home = "/Users/me"
		}
		var scripts []string
		cfg := &iterm2.ContentsConfig{
			Getenv: func(key string) string {
				if key == iterm2.EnvITerm2AppPath {
					return req.EnvPath
				}
				return ""
			},
			Home: func() string { return home },
			IsApp: func(path string) bool {
				if req.EnvPath != "" && path == req.EnvPath {
					return true
				}
				if req.HomeApp && strings.HasSuffix(path, "/Applications/iTerm.app") && strings.Contains(path, home) {
					return true
				}
				if req.SysApp && path == iterm2.AppPath {
					return true
				}
				return false
			},
			Running: func(abs string) bool {
				if req.Running == nil {
					return true
				}
				for _, p := range req.Running {
					if p == abs {
						return true
					}
				}
				return false
			},
			Exec: func(script string) (string, error) {
				scripts = append(scripts, script)
				for _, miss := range req.NotFoundApps {
					if miss != "" && strings.Contains(script, miss) {
						return "", fmt.Errorf("session not found: %s", sid)
					}
				}
				return req.Body, nil
			},
		}
		got, err := iterm2.Contents(req.SessionID, cfg)
		resp.Scripts = scripts
		resp.ExecN = len(scripts)
		resp.SessionID = got.SessionID
		resp.App = got.App
		resp.Contents = got.Contents
		return resp, err
	default:
		return nil, errors.New("unknown phase: " + req.Phase)
	}
}
```
