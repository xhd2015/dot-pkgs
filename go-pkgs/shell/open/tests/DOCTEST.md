# shell/open — open a path, a URL, or a path with an application

## Version

0.0.1

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/open`.

**Default layer: L2** in-process library API with an injectable `GOOS` and
`Run`. No real `open`, `xdg-open`, `cmd start`, or GUI launch. Parallel-safe:
`t.TempDir()` + injectors only. No `t.Setenv` / `t.Chdir` / `os.Setenv` /
`os.Chdir`.

**Out of scope this cycle:** launching a terminal at a directory
(`shell/openterm2`), the VS Code CLI (`shell/vscode`), `open -R` reveal,
`open -b` bundle ids, Windows `start` escaping beyond the empty title.

# DSN (Domain Specific Notion)

Hand a target to the desktop so the operating system picks the handler. The
argv is a pure function of the target and the platform; the launch is one
injected runner call.

Targets are three:

- **path** — a file, directory, or `.app` bundle (default handler).
- **url** — a URL (the browser). Same argv shape as path; the name documents
  intent.
- **app** — a path opened with a named or bundled application
  (`open -a <app> [path]`). macOS only.

### Participants

- **`Args(target, goos)`** — pure argv for the default handler.
- **`URLArgs(url, goos)`** — `Args` for a URL.
- **`AppArgs(app, path, goos)`** — pure argv for `open -a`; empty `path`
  launches the app itself.
- **`PathConfig` / `URLConfig` / `AppConfig`** — build argv, then call the
  runner once; `Path` / `URL` / `App` are the nil-Config forms.
- **`Config`** — `GOOS` (empty → `runtime.GOOS`), `Run` (nil → exec and
  collect combined output).
- **`Result`** — `Args` (what ran) and `Out` (runner output).

### Behaviors

**Argv, by platform**

- `darwin` → `{"open", target}`.
- `linux` → `{"xdg-open", target}`.
- `windows` → `{"cmd", "/c", "start", "", target}` — the empty argument is the
  window title `start` would otherwise take from the target.
- Anything else → error; no runner call.

**Reject before running**

- Empty or whitespace-only target → error; runner not called.
- Empty app name → error; runner not called.
- `open -a` on a non-darwin platform → error; runner not called.

**Route**

- Valid target → exactly one runner call with the built argv; success returns
  `Result.Args` and `Result.Out`.
- Runner error → returned; no `Result`.

**Rules**

- Reuse `lookpath`-style injection: the platform and the runner live on
  `Config` only.
- Never mutate process env or cwd.
- Parallel-safe: no shared state, no fixed paths.

## Decision Tree

```
shell/open/tests/
├── DOCTEST.md
├── SETUP.md                               # WorkDir + target fixture
├── args/                                  # pure argv builders (no runner)
│   ├── path-darwin/                       # open <target>
│   ├── url-linux/                         # xdg-open <url>
│   ├── url-windows/                       # cmd /c start "" <url>
│   ├── unknown-platform/                  # error; no argv
│   ├── app-with-path/                     # open -a <app> <path>
│   ├── app-without-path/                  # open -a <app>
│   └── app-non-darwin/                    # error; no argv
└── run/                                   # Config.Run injection
    ├── reject-empty-path/                 # error; runner not called
    ├── reject-empty-app/                  # error; runner not called
    ├── path-success/                      # argv + Out
    ├── url-success/                       # argv + Out
    ├── app-success/                       # argv + Out
    └── runner-error/                      # error surfaces
```

### Parameter significance (high → low)

1. **Operation** — args (pure) | run (injected runner)
2. **Kind** — path | url | app (selects the builder)
3. **Platform** — darwin | linux | windows | unknown
4. **Validity** — empty target / empty app / non-darwin `open -a`
5. **Runner outcome** — success | injected error

## Test Index

| # | Leaf | API | Description |
|---|------|-----|-------------|
| 1 | `args/path-darwin` | Args | `open <target>` |
| 2 | `args/url-linux` | URLArgs | `xdg-open <url>` |
| 3 | `args/url-windows` | URLArgs | `cmd /c start "" <url>` |
| 4 | `args/unknown-platform` | Args | unknown GOOS → error |
| 5 | `args/app-with-path` | AppArgs | `open -a <app> <path>` |
| 6 | `args/app-without-path` | AppArgs | `open -a <app>` |
| 7 | `args/app-non-darwin` | AppArgs | `open -a` on linux → error |
| 8 | `run/reject-empty-path` | PathConfig | empty path → error; no runner |
| 9 | `run/reject-empty-app` | AppConfig | empty app → error; no runner |
| 10 | `run/path-success` | PathConfig | argv + Out |
| 11 | `run/url-success` | URLConfig | argv + Out |
| 12 | `run/app-success` | AppConfig | argv + Out |
| 13 | `run/runner-error` | PathConfig | runner error surfaces; no Result |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/open/tests
doctest test ./shell/open/tests
doctest test -v ./shell/open/tests/args
doctest test -v ./shell/open/tests/run/path-success
```

All leaves are unlabeled L2. Discovery runs the full tree. There is no `e2e`
leaf.

### Intended public API (implementer pins names)

```go
package open

type Result struct {
	Args []string
	Out  string
}

type Config struct {
	GOOS string
	Run  func(args []string) (string, error)
}

func Args(target, goos string) ([]string, error)
func URLArgs(url, goos string) ([]string, error)
func AppArgs(app, path, goos string) ([]string, error)

func Path(path string) (*Result, error)
func PathConfig(path string, cfg *Config) (*Result, error)
func URL(url string) (*Result, error)
func URLConfig(url string, cfg *Config) (*Result, error)
func App(app, path string) (*Result, error)
func AppConfig(app, path string, cfg *Config) (*Result, error)
```

```go
import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/open"
)

// Request is filled root→leaf. Kind selects which public API Run calls.
type Request struct {
	Operation string // args | run
	Kind      string // path | url | app

	// WorkDir is an isolated temp root; Target is a path inside it.
	WorkDir string
	Target  string

	URL     string
	AppName string
	AppPath string
	GOOS    string

	// RunnerErr non-empty → the injected runner returns errors.New(that).
	RunnerErr string
}

// Response observes API outputs and runner spies.
type Response struct {
	Args []string
	Out  string

	Runs [][]string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}

	switch req.Operation {
	case "args":
		argv, err := buildArgs(req)
		resp.Args = argv
		return resp, err

	case "run":
		cfg := &open.Config{
			GOOS: req.GOOS,
			Run: func(args []string) (string, error) {
				resp.Runs = append(resp.Runs, args)
				if req.RunnerErr != "" {
					return "", errors.New(req.RunnerErr)
				}
				return "launched", nil
			},
		}
		var res *open.Result
		var err error
		switch req.Kind {
		case "path":
			res, err = open.PathConfig(req.Target, cfg)
		case "url":
			res, err = open.URLConfig(req.URL, cfg)
		case "app":
			res, err = open.AppConfig(req.AppName, req.AppPath, cfg)
		default:
			t.Fatalf("unknown Kind %q for Operation run", req.Kind)
		}
		if res != nil {
			resp.Args = res.Args
			resp.Out = res.Out
		}
		return resp, err

	default:
		t.Fatalf("unknown Operation %q", req.Operation)
	}
	return resp, nil
}

func buildArgs(req *Request) ([]string, error) {
	switch req.Kind {
	case "path":
		return open.Args(req.Target, req.GOOS)
	case "url":
		return open.URLArgs(req.URL, req.GOOS)
	case "app":
		return open.AppArgs(req.AppName, req.AppPath, req.GOOS)
	default:
		return nil, errors.New("unknown Kind " + req.Kind)
	}
}

func wantGOOS(req *Request) string {
	if req.GOOS != "" {
		return req.GOOS
	}
	return "darwin"
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

func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args = %#v, want %#v", got, want)
		}
	}
}

func assertNoRun(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.Runs) != 0 {
		t.Fatalf("Runs = %#v, want none (validation must run before the runner)", resp.Runs)
	}
}

func assertRunOnce(t *testing.T, resp *Response, want []string) {
	t.Helper()
	if len(resp.Runs) != 1 {
		t.Fatalf("Runs = %#v, want exactly one call", resp.Runs)
	}
	assertArgs(t, resp.Runs[0], want)
}

func assertOut(t *testing.T, resp *Response, want string) {
	t.Helper()
	if resp.Out != want {
		t.Fatalf("Out = %q, want %q", resp.Out, want)
	}
}

func joinTarget(req *Request, name string) string {
	return filepath.Join(req.WorkDir, name)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
```
