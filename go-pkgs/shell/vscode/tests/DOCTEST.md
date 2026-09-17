# shell/vscode — open a directory or file in Visual Studio Code

## Version

0.0.1

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/vscode`.

**Default layer: L2** in-process library API with an injectable resolver and
runner. No real `code`, no VS Code window, no `lookpath` login-shell probe.
Parallel-safe: `t.TempDir()` + injectors only. No `t.Setenv` / `t.Chdir` /
`os.Setenv` / `os.Chdir`.

**Out of scope this cycle:** the Marcus daemon endpoint
(`POST /api/workspace/open-vscode`), the ai-workshop button, `code --diff` /
`--add` / extension commands, Windows `.cmd` shim handling, `shell/open`
(the OS `open` primitive) and `shell/openterm2` (terminals).

# DSN (Domain Specific Notion)

Open a directory or a file in VS Code through the `code` CLI. The CLI is often
absent from a thin PATH (menu-bar apps, daemons), so resolution goes through
`shell/lookpath`: PATH, then the well-known app bundles, then login shells.

Opening a **directory** reuses the current window by default (`code -r`), which
is what an "Open in VS Code" button should do; `NewWindow` selects `-n`.

### Participants

- **`WellKnownCandidates(home)`** — pure: the `code` path inside the VS Code and
  VS Code Insiders bundles, system-wide then per-user.
- **`Resolve(opts)`** — `lookpath.Look("code", opts)` with the app-bundle
  candidates filled in when the caller supplies none.
- **`Args(codePath, dir, newWindow)`** — pure argv, CLI first.
- **`FileArgs(codePath, file, line)`** — pure argv; `line > 0` uses `--goto`.
- **`Open(dir)` / `OpenConfig(dir, cfg)`** — resolve, build argv, run once.
- **`OpenFile(file, line)` / `OpenFileConfig`** — the same for a file.
- **`Config`** — `Home`, `NewWindow`, `Resolve` (nil → `Resolve`), `Run` (nil →
  detached start).
- **`Result`** — `CodePath`, `Via`, `Args`, `Out`.

### Behaviors

**Resolve once, then run once**

- `Resolve` is called with `Config.Home` (empty → `os.UserHomeDir`).
- Exactly one runner call, with `Args[0]` = the resolved CLI.

**Reject before resolving**

- Empty or whitespace-only dir → error; `Resolve` and `Run` are not called.
- Empty or whitespace-only file → error; `Resolve` and `Run` are not called.

**Reuse by default**

- `NewWindow` false → `-r`; true → `-n`.

**File and line**

- `line > 0` → `{"--goto", "<file>:<line>"}`; otherwise `{<file>}`.

**Failure**

- Resolve error → returned wrapped, and the runner is not called.
- Runner error → returned; no `Result`.

**Rules**

- The path is not stat-ed: `code` opens a new buffer for a path that is not
  there yet, and the existing callers already hold a real one.
- Parallel-safe: all injection is on `Config`; no shared state.
- Never mutate process env or cwd.

## Decision Tree

```
shell/vscode/tests/
├── DOCTEST.md
├── SETUP.md                               # WorkDir + Home + Dir fixture
├── args/                                  # pure argv builders (no runner)
│   ├── reuse-window/                      # code -r <dir>
│   ├── new-window/                        # code -n <dir>
│   ├── file-with-line/                    # code --goto <file>:<line>
│   ├── file-without-line/                 # code <file>
│   └── file-empty/                        # error; no argv
├── candidates/                            # WellKnownCandidates (pure)
│   ├── system-and-home/                   # 4 paths: 2 system + 2 per-user
│   └── empty-home/                        # 2 system paths only
└── open/                                  # Config.Resolve + Config.Run injection
    ├── reject-empty-dir/                  # error; resolve + run not called
    ├── resolve-error/                     # error; run not called
    ├── runner-error/                      # error surfaces
    ├── success/                           # argv, Result.CodePath/Via
    ├── new-window-success/                # -n reaches the runner
    └── open-file-line/                    # --goto reaches the runner
```

### Parameter significance (high → low)

1. **Operation** — args (pure) | candidates (pure) | open (injected)
2. **Kind** — dir | file
3. **Window mode** — reuse (`-r`) | new (`-n`)
4. **Line** — positive (goto) | zero (plain file)
5. **Failure point** — none | resolve error | runner error

## Test Index

| # | Leaf | API | Description |
|---|------|-----|-------------|
| 1 | `args/reuse-window` | Args | `code -r <dir>` |
| 2 | `args/new-window` | Args | `code -n <dir>` |
| 3 | `args/file-with-line` | FileArgs | `code --goto <file>:<line>` |
| 4 | `args/file-without-line` | FileArgs | `code <file>` |
| 5 | `args/file-empty` | FileArgs | empty file → error |
| 6 | `candidates/system-and-home` | WellKnownCandidates | system + per-user bundles |
| 7 | `candidates/empty-home` | WellKnownCandidates | system bundles only |
| 8 | `open/reject-empty-dir` | OpenConfig | error; resolve/run not called |
| 9 | `open/resolve-error` | OpenConfig | resolve error; run not called |
| 10 | `open/runner-error` | OpenConfig | runner error surfaces; no Result |
| 11 | `open/success` | OpenConfig | argv + CodePath/Via |
| 12 | `open/new-window-success` | OpenConfig | `-n` reaches the runner |
| 13 | `open/open-file-line` | OpenFileConfig | `--goto` reaches the runner |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/vscode/tests
doctest test ./shell/vscode/tests
doctest test -v ./shell/vscode/tests/args
doctest test -v ./shell/vscode/tests/open/success
```

All leaves are unlabeled L2. Discovery runs the full tree. There is no `e2e`
leaf.

### Intended public API (implementer pins names)

```go
package vscode

type Result struct {
	CodePath string
	Via      string
	Args     []string
	Out      string
}

type Config struct {
	Home      string
	NewWindow bool
	Resolve   func(home string) (lookpath.Result, error)
	Run       func(args []string) (string, error)
}

func WellKnownCandidates(home string) []string
func Resolve(opts lookpath.Options) (lookpath.Result, error)
func Args(codePath, dir string, newWindow bool) []string
func FileArgs(codePath, file string, line int) ([]string, error)

func Open(dir string) (*Result, error)
func OpenConfig(dir string, cfg *Config) (*Result, error)
func OpenFile(file string, line int) (*Result, error)
func OpenFileConfig(file string, line int, cfg *Config) (*Result, error)

// Config.Run values: nil defaults to DetachRunner.
func DetachRunner(args []string) (string, error)
func WaitRunner(args []string) (string, error)
```

```go
import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/vscode"
)

// Request is filled root→leaf. Kind selects which public API Run calls.
type Request struct {
	Operation string // args | candidates | open
	Kind      string // dir | file

	// WorkDir is an isolated temp root; Dir and File are fixtures inside it.
	WorkDir string
	Dir     string
	File    string

	// Line is passed to the file APIs; > 0 means --goto.
	Line int
	// NewWindow selects -n instead of -r.
	NewWindow bool
	// Home is the per-user root; CodePath/Via are what the injected resolver
	// reports.
	Home     string
	CodePath string
	Via      string

	// ResolveErr non-empty → the injected resolver fails with that message.
	ResolveErr string
	// RunnerErr non-empty → the injected runner fails with that message.
	RunnerErr string
	// RunOut is the injected runner's output on success.
	RunOut string
}

// Response observes API outputs and injection spies.
type Response struct {
	Args       []string
	Out        string
	CodePath   string
	Via        string
	Candidates []string

	ResolveHomes []string
	Runs         [][]string
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
		switch req.Kind {
		case "dir":
			resp.Args = vscode.Args(req.CodePath, req.Dir, req.NewWindow)
			return resp, nil
		case "file":
			argv, err := vscode.FileArgs(req.CodePath, req.File, req.Line)
			resp.Args = argv
			return resp, err
		default:
			t.Fatalf("unknown Kind %q for Operation args", req.Kind)
		}

	case "candidates":
		resp.Candidates = vscode.WellKnownCandidates(req.Home)
		return resp, nil

	case "open":
		cfg := &vscode.Config{
			Home:      req.Home,
			NewWindow: req.NewWindow,
			Resolve: func(home string) (lookpath.Result, error) {
				resp.ResolveHomes = append(resp.ResolveHomes, home)
				if req.ResolveErr != "" {
					return lookpath.Result{}, errors.New(req.ResolveErr)
				}
				via := req.Via
				if via == "" {
					via = "path"
				}
				return lookpath.Result{Path: req.CodePath, Via: via}, nil
			},
			Run: func(args []string) (string, error) {
				resp.Runs = append(resp.Runs, args)
				if req.RunnerErr != "" {
					return "", errors.New(req.RunnerErr)
				}
				return req.RunOut, nil
			},
		}
		var res *vscode.Result
		var err error
		switch req.Kind {
		case "dir":
			res, err = vscode.OpenConfig(req.Dir, cfg)
		case "file":
			res, err = vscode.OpenFileConfig(req.File, req.Line, cfg)
		default:
			t.Fatalf("unknown Kind %q for Operation open", req.Kind)
		}
		if res != nil {
			resp.Args = res.Args
			resp.Out = res.Out
			resp.CodePath = res.CodePath
			resp.Via = res.Via
		}
		return resp, err

	default:
		t.Fatalf("unknown Operation %q", req.Operation)
	}
	return resp, nil
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
		t.Fatalf("Runs = %#v, want none", resp.Runs)
	}
}

func assertRunOnce(t *testing.T, resp *Response, want []string) {
	t.Helper()
	if len(resp.Runs) != 1 {
		t.Fatalf("Runs = %#v, want exactly one call", resp.Runs)
	}
	assertArgs(t, resp.Runs[0], want)
}

func assertNoResolve(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.ResolveHomes) != 0 {
		t.Fatalf("ResolveHomes = %#v, want none (validation must run before resolving)", resp.ResolveHomes)
	}
}

func assertResolvedHome(t *testing.T, resp *Response, want string) {
	t.Helper()
	if len(resp.ResolveHomes) != 1 {
		t.Fatalf("ResolveHomes = %#v, want exactly one call", resp.ResolveHomes)
	}
	if resp.ResolveHomes[0] != want {
		t.Fatalf("Resolve home = %q, want %q", resp.ResolveHomes[0], want)
	}
}

func assertDefaultRunner(t *testing.T, resp *Response, req *Request, want []string) {
	t.Helper()
	assertRunOnce(t, resp, want)
	if resp.CodePath != req.CodePath {
		t.Fatalf("Result.CodePath = %q, want %q", resp.CodePath, req.CodePath)
	}
}

func countSuffix(t *testing.T, paths []string, suffix string) int {
	t.Helper()
	n := 0
	for _, p := range paths {
		if strings.HasSuffix(p, suffix) {
			n++
		}
	}
	return n
}

func hasPrefix(paths []string, prefix string) bool {
	for _, p := range paths {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func systemCodePath(name string) string {
	return filepath.Join("/Applications", name, "Contents", "Resources", "app", "bin", "code")
}
```
