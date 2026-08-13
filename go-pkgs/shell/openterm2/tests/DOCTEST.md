# shell/openterm2 — open a directory in iTerm2 or Terminal.app

## Version

0.0.2

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/openterm2`.

The production package does **not** exist yet. Root `Run` imports it so the suite
is **RED** (compile failure) until the implementer lands the public API under
`shell/openterm2/`.

**Default layer: L2** in-process library API with injectable `ResolveITerm`,
`OpenITerm`, and `OpenTerminal`. No real iTerm2 or Terminal.app launch, no
`osascript`, no `open -a`. Parallel-safe: `t.TempDir()` + injectors only. No
`t.Setenv` / `t.Chdir` / `os.Setenv` / `os.Chdir`.

**Out of scope this cycle:** Marcus Workspace button / daemon
`POST /api/workspace/open-terminal`; changing `iterm2.Open`, `ResolveAppPath`,
or `ErrNotInstalled`; AppleScript / smart-reuse details; `ITERM2_APP_PATH`
order (owned by `shell/iterm2` app-path); Windows / Linux; real GUI launches.

# DSN (Domain Specific Notion)

Open a directory in a terminal on macOS: prefer a resolvable iTerm2 app, else
fall back to system Terminal. Never fall through to Terminal after a failed
iTerm2 open.

### Participants

- **`Open(dir)`** — `OpenConfig(dir, nil)`.
- **`OpenConfig(dir, cfg)`** — validate `dir`, resolve iTerm2, invoke one opener,
  return `Result{Via, AppPath}` or error.
- **`Config`** — `ResolveITerm` (nil → `iterm2.ResolveAppPath`), `OpenITerm`
  (nil → `iterm2.Open`), `OpenTerminal` (nil → `open -a Terminal.app <dir>`),
  `TerminalApp` (empty → `/Applications/Utilities/Terminal.app`).
- **`Result`** — `Via` is `iterm2` or `terminal`; `AppPath` is the `.app` path
  handed to the opener that ran.
- **`TerminalOpenArgs(app, dir)`** — pure argv for the Terminal fallback:
  `open -a <app> <absDir>`. Does not exec.

### Behaviors

**Validate first (before either opener)**

- Empty or whitespace-only `dir` → error; neither opener runs.
- Missing path or not a directory → error; neither opener runs.

**Route**

- `ResolveITerm()` non-empty → call `OpenITerm(dir)` only. Success:
  `Via=iterm2`, `AppPath` = resolve path. Error: return it (or wrap);
  **do not** call `OpenTerminal`.
- `ResolveITerm()` empty → call `OpenTerminal(dir)` only. Success:
  `Via=terminal`, `AppPath` = configured or default Terminal app.

**Argv helper**

- `TerminalOpenArgs(app, dir)` ≡ `[]string{"open", "-a", app, absDir}`.

**Rules**

- Reuse `iterm2.ResolveAppPath` / `iterm2.Open` as defaults; do not re-probe
  iTerm install paths here.
- Parallel-safe: inject resolvers and openers on `Config` only.

## Decision Tree

```
shell/openterm2/tests/
├── DOCTEST.md
├── SETUP.md
├── open/                                  # OpenConfig + injectables
│   ├── reject/                            # invalid dir; neither opener
│   │   ├── empty-dir/                     # dir=""
│   │   ├── whitespace-dir/                # whitespace-only dir
│   │   ├── not-directory/                 # dir is a file
│   │   └── missing-dir/                   # dir does not exist
│   ├── via-iterm2/                        # ResolveITerm non-empty
│   │   ├── success/                       # OpenITerm ok → Via=iterm2
│   │   ├── open-fails/                    # OpenITerm error; no Terminal
│   │   └── ignores-terminal-app/          # TerminalApp set; still iTerm
│   └── via-terminal/                      # ResolveITerm empty
│       ├── default-app/                   # AppPath = default Terminal.app
│       ├── configured-app/                # TerminalApp override
│       └── open-fails/                    # OpenTerminal error; no iTerm
└── terminal-open-args/                    # TerminalOpenArgs pure helper
    ├── default-app/                       # open -a default <abs dir>
    ├── configured-app/                    # open -a override <abs dir>
    └── relative-dir/                      # relative dir becomes abs
```

### Parameter significance (high → low)

1. **Operation** — open-config | terminal-open-args
2. **Dir validity** (open-config) — reject vs route
3. **ResolveITerm** — non-empty (iTerm) vs empty (Terminal)
4. **Opener outcome / TerminalApp** — success vs injected error; default vs override app
5. **Argv shape** — default app | override app | relative dir → abs

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `open/reject/empty-dir` | OpenConfig | Empty dir → error; neither opener | RED |
| 2 | `open/reject/whitespace-dir` | OpenConfig | Whitespace-only dir → error; neither opener | RED |
| 3 | `open/reject/not-directory` | OpenConfig | File path → error; neither opener | RED |
| 4 | `open/reject/missing-dir` | OpenConfig | Missing path → error; neither opener | RED |
| 5 | `open/via-iterm2/success` | OpenConfig | Resolve hit + OpenITerm ok → Via=iterm2; no Terminal | RED |
| 6 | `open/via-iterm2/open-fails` | OpenConfig | OpenITerm error returned; OpenTerminal not called | RED |
| 7 | `open/via-iterm2/ignores-terminal-app` | OpenConfig | TerminalApp set still routes to iTerm | RED |
| 8 | `open/via-terminal/default-app` | OpenConfig | Resolve empty → Via=terminal, default AppPath | RED |
| 9 | `open/via-terminal/configured-app` | OpenConfig | TerminalApp override in Result.AppPath | RED |
| 10 | `open/via-terminal/open-fails` | OpenConfig | OpenTerminal error; OpenITerm not called | RED |
| 11 | `terminal-open-args/default-app` | TerminalOpenArgs | `open -a` default Terminal.app + abs dir | RED |
| 12 | `terminal-open-args/configured-app` | TerminalOpenArgs | `open -a` override app + abs dir | RED |
| 13 | `terminal-open-args/relative-dir` | TerminalOpenArgs | Relative dir becomes `filepath.Abs` | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/openterm2/tests
doctest test ./shell/openterm2/tests
doctest test -v ./shell/openterm2/tests/open/via-iterm2/success
doctest test -v ./shell/openterm2/tests/open
```

Classic TDD: expect **RED** (compile failure until package exists, then assert
failures against incomplete implementations) until implementer lands the API.

All leaves are unlabeled L2. Discovery runs the full tree. There is no `e2e`
leaf.

### Intended public API (implementer pins names)

```go
package openterm2

const (
	ViaITerm2   = "iterm2"
	ViaTerminal = "terminal"
)

// Default Terminal.app when Config.TerminalApp is empty.
// /Applications/Utilities/Terminal.app

type Result struct {
	Via     string // "iterm2" | "terminal"
	AppPath string // resolved .app path passed to the opener
}

type Config struct {
	ResolveITerm func() string          // nil → iterm2.ResolveAppPath
	OpenITerm    func(dir string) error // nil → iterm2.Open
	OpenTerminal func(dir string) error // nil → open -a Terminal.app <dir>
	TerminalApp  string                 // empty → /Applications/Utilities/Terminal.app
}

func Open(dir string) (*Result, error)
func OpenConfig(dir string, cfg *Config) (*Result, error)
func TerminalOpenArgs(appPath, dir string) []string
```

```go
import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/openterm2"
)

const defaultTerminalApp = "/Applications/Utilities/Terminal.app"

// Request is filled root→leaf. Operation selects which public API Run calls.
type Request struct {
	Operation string // open-config | terminal-open-args

	// WorkDir is an isolated temp root (root Setup). ValidDir is an existing
	// subdirectory used as the default target directory.
	WorkDir  string
	ValidDir string

	// Dir is passed to OpenConfig / TerminalOpenArgs.
	Dir string

	// ITermAppPath is returned by the injected ResolveITerm hook.
	// Empty means "iTerm2 not resolvable".
	ITermAppPath string

	// OpenITermErr / OpenTerminalErr: empty → injected opener succeeds;
	// non-empty → opener returns errors.New(that message).
	OpenITermErr    string
	OpenTerminalErr string

	// TerminalApp is copied onto Config.TerminalApp (empty → product default).
	TerminalApp string

	// ArgsAppPath is the app argument to TerminalOpenArgs.
	ArgsAppPath string
}

// Response observes API outputs and injection spies.
type Response struct {
	Via     string
	AppPath string

	ResolveCalls      int
	OpenITermCalls    []string
	OpenTerminalCalls []string

	TerminalArgs []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}

	switch req.Operation {
	case "open-config":
		cfg := &openterm2.Config{
			TerminalApp: req.TerminalApp,
			ResolveITerm: func() string {
				resp.ResolveCalls++
				return req.ITermAppPath
			},
			OpenITerm: func(dir string) error {
				resp.OpenITermCalls = append(resp.OpenITermCalls, dir)
				if req.OpenITermErr != "" {
					return errors.New(req.OpenITermErr)
				}
				return nil
			},
			OpenTerminal: func(dir string) error {
				resp.OpenTerminalCalls = append(resp.OpenTerminalCalls, dir)
				if req.OpenTerminalErr != "" {
					return errors.New(req.OpenTerminalErr)
				}
				return nil
			},
		}
		res, err := openterm2.OpenConfig(req.Dir, cfg)
		if res != nil {
			resp.Via = res.Via
			resp.AppPath = res.AppPath
		}
		return resp, err

	case "terminal-open-args":
		resp.TerminalArgs = openterm2.TerminalOpenArgs(req.ArgsAppPath, req.Dir)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Operation %q", req.Operation)
	}
}

func wantTerminalApp(req *Request) string {
	if req.TerminalApp != "" {
		return req.TerminalApp
	}
	return defaultTerminalApp
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %#v, want %#v", field, got, want)
	}
}

func assertNoOpeners(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.OpenITermCalls) != 0 {
		t.Fatalf("OpenITermCalls = %#v, want none (validation must run before openers)", resp.OpenITermCalls)
	}
	if len(resp.OpenTerminalCalls) != 0 {
		t.Fatalf("OpenTerminalCalls = %#v, want none (validation must run before openers)", resp.OpenTerminalCalls)
	}
}

func assertOpenITermOnce(t *testing.T, resp *Response, dir string) {
	t.Helper()
	if len(resp.OpenITermCalls) != 1 {
		t.Fatalf("OpenITermCalls = %#v, want exactly [%q]", resp.OpenITermCalls, dir)
	}
	if resp.OpenITermCalls[0] != dir {
		t.Fatalf("OpenITermCalls[0] = %q, want %q", resp.OpenITermCalls[0], dir)
	}
}

func assertOpenTerminalOnce(t *testing.T, resp *Response, dir string) {
	t.Helper()
	if len(resp.OpenTerminalCalls) != 1 {
		t.Fatalf("OpenTerminalCalls = %#v, want exactly [%q]", resp.OpenTerminalCalls, dir)
	}
	if resp.OpenTerminalCalls[0] != dir {
		t.Fatalf("OpenTerminalCalls[0] = %q, want %q", resp.OpenTerminalCalls[0], dir)
	}
}

func assertNoOpenITerm(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.OpenITermCalls) != 0 {
		t.Fatalf("OpenITermCalls = %#v, want none", resp.OpenITermCalls)
	}
}

func assertNoOpenTerminal(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.OpenTerminalCalls) != 0 {
		t.Fatalf("OpenTerminalCalls = %#v, want none (must not fall through to Terminal)", resp.OpenTerminalCalls)
	}
}

func assertTerminalArgs(t *testing.T, got []string, app, dir string) {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", dir, err)
	}
	want := []string{"open", "-a", app, abs}
	if len(got) != len(want) {
		t.Fatalf("TerminalOpenArgs = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("TerminalOpenArgs = %#v, want %#v", got, want)
		}
	}
}

func fakeITermApp(req *Request) string {
	return filepath.Join(req.WorkDir, "Applications", "iTerm.app")
}
```
