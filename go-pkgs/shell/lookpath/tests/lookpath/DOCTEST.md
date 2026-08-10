# shell/lookpath — resolve CLI binaries under thin GUI PATH

## Version

0.0.2

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`.

The production package does **not** exist yet. Root `Run` imports it so the suite
is **RED** (compile failure) until the implementer lands the public API under
`shell/lookpath/`.

**Default layer: L2** in-process library API with injectable `LookPath`,
`IsExecutable`, and `RunLogin`. No process PATH mutation, no `t.Setenv` /
`t.Chdir`, no real login-shell spawn in the default suite. Parallel-safe:
`t.TempDir()` + injectors only.

**Out of scope this cycle:** Marcus wiring, agent-pro migration, codex/npm/nvm
candidate lists, process env mutation, `init()` discovery, real bash `-lic` e2e.

# DSN (Domain Specific Notion)

Shared pure library: resolve a CLI binary name when the process PATH is thin
(Launch Services / menu-bar apps).

### Participants

- **`Look(name, opts)`** — ordered resolution; returns `Result{Path, Via}` or
  error. Never mutates process env/cwd.
- **`LookPath(name, opts)`** — convenience: returns `Result.Path` only (same
  pipeline as `Look`).
- **`DefaultDirs(home)`** — ordered default search directories when `home` is
  set: `$HOME/.local/bin`, `$HOME/go/bin`, `/opt/homebrew/bin`,
  `/usr/local/bin`. Empty home → system dirs only (no `$HOME/…` entries).
- **`IsExecutable(path)`** — true only for an existing regular file with
  execute bit; false for missing, directories, or non-executable files.
- **`Options`** — `Home`, `ExtraDirs`, `ExtraCandidates`, `Shells` (default
  `bash`, `zsh`), `Timeout` (~5s per shell default), injectables `LookPath`,
  `IsExecutable`, `RunLogin` (nil → production defaults).
- **`Result.Via`** — which stage won: `direct` | `path` | `extra_dir` |
  `default_dir` | `candidate` | `login_shell:<shell>`.

### Behaviors

**Resolution order (`Look` / `LookPath`)**

1. Absolute path or name containing a path separator → check **only** that
   path (must be executable) → `Via=direct`. No PATH / dir / login fallthrough.
2. `opts.LookPath(name)` (default `exec.LookPath`) → `Via=path`
3. Each `ExtraDirs` + name → first executable → `Via=extra_dir`
4. `DefaultDirs(home)` + name → first executable → `Via=default_dir`
5. Each `ExtraCandidates` absolute path → first executable → `Via=candidate`
6. Login shells in order: `RunLogin(shell, "command -v <name>", minimal env)` →
   first non-empty path stdout → `Via=login_shell:<shell>`
7. Else error; message includes the binary name.

**Rules**

- Skip non-executable files and directories; continue later stages (except
  direct-path, which does not fall through).
- Parallel-safe: injectables only via `Options`; product never `Setenv`/`Chdir`.

## Decision Tree

```
shell/lookpath/tests/lookpath/
├── DOCTEST.md
├── SETUP.md
├── look/                                  # Look(name, opts) — stage that wins
│   ├── direct-path/                       # abs / path-separator form only
│   │   ├── absolute-executable/           # abs exec → Via=direct
│   │   ├── absolute-missing/              # abs missing → error; no fallthrough
│   │   ├── non-executable/                # abs non-exec file → error
│   │   └── is-directory/                  # abs directory → error
│   ├── via-path/                          # injected LookPath hit
│   │   └── hit/                           # Via=path
│   ├── via-extra-dir/
│   │   ├── hit/                           # file in ExtraDirs → Via=extra_dir
│   │   └── skips-non-executable/          # non-exec first dir, hit second
│   ├── via-default-dir/
│   │   └── hit-go-bin/                    # $HOME/go/bin → Via=default_dir
│   ├── via-candidate/
│   │   └── hit/                           # ExtraCandidates → Via=candidate
│   ├── via-login-shell/
│   │   ├── bash-hit/                      # RunLogin bash → login_shell:bash
│   │   └── bash-fail-zsh-hit/             # bash fail, zsh → login_shell:zsh
│   └── not-found/
│       └── all-miss/                      # all stages miss → error w/ name
├── look-path/                             # LookPath convenience
│   ├── returns-path/                      # hit → path string only
│   └── returns-error/                     # miss → error
├── default-dirs/                          # DefaultDirs pure
│   ├── with-home/                         # home-relative + system bins
│   └── empty-home/                        # system bins only
└── is-executable/                         # IsExecutable pure
    ├── executable-file/                   # 0755 regular → true
    ├── non-executable-file/               # 0644 regular → false
    └── directory/                         # directory → false
```

### Parameter significance (high → low)

1. **Operation** — look | look-path | default-dirs | is-executable
2. **Winning stage / outcome** — direct | path | extra_dir | default_dir |
   candidate | login_shell | not-found | pure helper edges
3. **Skip / inject surface** — non-executable vs directory; LookPath / RunLogin
   fixtures; Home / ExtraDirs / ExtraCandidates layout
4. **Name form** — absolute path vs bare binary name

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `look/direct-path/absolute-executable` | Look | Absolute executable → Path + Via=direct; no LookPath call | RED |
| 2 | `look/direct-path/absolute-missing` | Look | Absolute missing → error; no fallthrough | RED |
| 3 | `look/direct-path/non-executable` | Look | Absolute non-executable file → error | RED |
| 4 | `look/direct-path/is-directory` | Look | Absolute directory → error | RED |
| 5 | `look/via-path/hit` | Look | Injected LookPath hit → Via=path | RED |
| 6 | `look/via-extra-dir/hit` | Look | ExtraDirs hit → Via=extra_dir | RED |
| 7 | `look/via-extra-dir/skips-non-executable` | Look | Non-exec in first ExtraDir, hit second | RED |
| 8 | `look/via-default-dir/hit-go-bin` | Look | `$HOME/go/bin` → Via=default_dir | RED |
| 9 | `look/via-candidate/hit` | Look | ExtraCandidates → Via=candidate | RED |
| 10 | `look/via-login-shell/bash-hit` | Look | RunLogin bash → Via=login_shell:bash | RED |
| 11 | `look/via-login-shell/bash-fail-zsh-hit` | Look | bash fail, zsh success → Via=login_shell:zsh | RED |
| 12 | `look/not-found/all-miss` | Look | All stages miss → error includes name | RED |
| 13 | `look-path/returns-path` | LookPath | Hit → path string only | RED |
| 14 | `look-path/returns-error` | LookPath | Miss → error | RED |
| 15 | `default-dirs/with-home` | DefaultDirs | Home set → home bins + system bins | RED |
| 16 | `default-dirs/empty-home` | DefaultDirs | Empty home → system bins only | RED |
| 17 | `is-executable/executable-file` | IsExecutable | 0755 regular file → true | RED |
| 18 | `is-executable/non-executable-file` | IsExecutable | 0644 regular file → false | RED |
| 19 | `is-executable/directory` | IsExecutable | Directory → false | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/lookpath/tests/lookpath
doctest test ./shell/lookpath/tests/lookpath
doctest test -v ./shell/lookpath/tests/lookpath/look/via-path/hit
doctest test -v ./shell/lookpath/tests/lookpath/look
```

Classic TDD: expect **RED** (compile failure until package exists, then assert
failures against incomplete implementations) until implementer lands the API.

### Intended public API (implementer pins names)

```go
package lookpath

import "time"

type Result struct {
	Path string // absolute path to invocable binary
	Via  string // direct | path | extra_dir | default_dir | candidate | login_shell:<shell>
}

type Options struct {
	Home            string
	ExtraDirs       []string
	ExtraCandidates []string
	Shells          []string      // default {"bash","zsh"}
	Timeout         time.Duration // default ~5s per shell

	// Injectables (nil = production)
	LookPath     func(file string) (string, error)
	IsExecutable func(path string) bool
	RunLogin     func(shell, command string, env []string) (stdout string, err error)
}

func Look(name string, opts Options) (Result, error)
func LookPath(name string, opts Options) (string, error)
func DefaultDirs(home string) []string
func IsExecutable(path string) bool
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// Request is filled root→leaf. Operation selects which public API Run calls.
type Request struct {
	Operation string // look | look-path | default-dirs | is-executable

	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// Shared Look / LookPath inputs
	Name            string
	Home            string
	ExtraDirs       []string
	ExtraCandidates []string
	Shells          []string
	Timeout         time.Duration

	// Injected LookPath: if LookPathHit non-empty, injectable returns it;
	// otherwise injectable always misses (never touches real PATH).
	LookPathHit string

	// Optional override map for IsExecutable injectable (path → bool).
	// When nil, harness uses real mode bits via harnessIsExecutable.
	ExecOverride map[string]bool

	// RunLogin injectable fixtures (shell → stdout path or fail).
	LoginStdout map[string]string
	LoginFail   map[string]bool

	// ExpectNoLookPath: Assert may check LookPath was not called (direct-path).
	ExpectNoLookPath bool

	// DefaultDirs
	DefaultDirsHome string

	// IsExecutable pure API
	IsExecPath string
}

// Response observes API outputs and injection spies.
type Response struct {
	Path string
	Via  string

	DefaultDirs  []string
	IsExecutable bool

	LookPathCalls     []string
	IsExecCalls       []string
	RunLoginCalls     []string // shell names in call order
	RunLoginCommands  []string // command strings in call order
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}

	switch req.Operation {
	case "look":
		res, err := lookpath.Look(req.Name, buildOpts(req, resp))
		resp.Path = res.Path
		resp.Via = res.Via
		return resp, err

	case "look-path":
		p, err := lookpath.LookPath(req.Name, buildOpts(req, resp))
		resp.Path = p
		return resp, err

	case "default-dirs":
		resp.DefaultDirs = lookpath.DefaultDirs(req.DefaultDirsHome)
		return resp, nil

	case "is-executable":
		resp.IsExecutable = lookpath.IsExecutable(req.IsExecPath)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Operation %q", req.Operation)
	}
}

func buildOpts(req *Request, resp *Response) lookpath.Options {
	opts := lookpath.Options{
		Home:            req.Home,
		ExtraDirs:       append([]string(nil), req.ExtraDirs...),
		ExtraCandidates: append([]string(nil), req.ExtraCandidates...),
		Shells:          append([]string(nil), req.Shells...),
		Timeout:         req.Timeout,
	}

	opts.LookPath = func(file string) (string, error) {
		resp.LookPathCalls = append(resp.LookPathCalls, file)
		if req.LookPathHit != "" {
			return req.LookPathHit, nil
		}
		return "", fmt.Errorf("lookpath: %s: not found", file)
	}

	opts.IsExecutable = func(path string) bool {
		resp.IsExecCalls = append(resp.IsExecCalls, path)
		if req.ExecOverride != nil {
			if v, ok := req.ExecOverride[path]; ok {
				return v
			}
			return false
		}
		return harnessIsExecutable(path)
	}

	opts.RunLogin = func(shell, command string, env []string) (string, error) {
		_ = env
		resp.RunLoginCalls = append(resp.RunLoginCalls, shell)
		resp.RunLoginCommands = append(resp.RunLoginCommands, command)
		if req.LoginFail != nil && req.LoginFail[shell] {
			return "", fmt.Errorf("injected login shell failure: %s", shell)
		}
		if req.LoginStdout != nil {
			if out, ok := req.LoginStdout[shell]; ok {
				return out, nil
			}
		}
		return "", fmt.Errorf("injected login shell: %s: not found", shell)
	}

	return opts
}

func harnessIsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return false
	}
	return fi.Mode()&0o111 != 0
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatalf("write exec %s: %v", path, err)
	}
}

func writeNonExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("not executable\n"), 0o644); err != nil {
		t.Fatalf("write non-exec %s: %v", path, err)
	}
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

func assertPathEqual(t *testing.T, got, want string) {
	t.Helper()
	got = filepath.Clean(got)
	want = filepath.Clean(want)
	if got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}

func assertErrorContainsName(t *testing.T, err error, name string) {
	t.Helper()
	assertError(t, err)
	if !strings.Contains(err.Error(), name) {
		t.Fatalf("error %q does not contain binary name %q", err.Error(), name)
	}
}

func assertNoLookPathCalls(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.LookPathCalls) != 0 {
		t.Fatalf("LookPathCalls = %#v, want none (direct path must not fall through)", resp.LookPathCalls)
	}
}

func assertContainsAll(t *testing.T, got []string, want ...string) {
	t.Helper()
	set := make(map[string]bool, len(got))
	for _, g := range got {
		set[g] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Fatalf("DefaultDirs missing %q; got %#v", w, got)
		}
	}
}

func assertNotContainsPrefix(t *testing.T, got []string, prefix string) {
	t.Helper()
	for _, g := range got {
		if strings.HasPrefix(g, prefix) || strings.Contains(g, prefix) {
			// home-relative entries use the home path as prefix
			if prefix != "" && strings.HasPrefix(g, prefix) {
				t.Fatalf("DefaultDirs unexpectedly contains home-prefixed entry %q (home prefix %q)", g, prefix)
			}
		}
	}
}
```
