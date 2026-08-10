# shell/lookpath — batch LookupPaths + Dirs / DirsEnv

## Version

0.0.2

Classic TDD doctests for batch path resolution on package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`.

The symbols under test (`LookupPaths`, `LookupItem`, `LookupItems`, `Dirs`,
`DirsEnv`) do **not** exist yet on the production package (only `Look` /
`LookPath` / `DefaultDirs` / `IsExecutable` do). Root `Run` imports them so the
suite is **RED** (compile failure) until the implementer lands the API.

**Default layer: L2** in-process library API with injectable `LookPath`,
`IsExecutable`, and `RunLogin`. No process PATH mutation, no `t.Setenv` /
`t.Chdir`, no real login-shell spawn. Parallel-safe: `t.TempDir()` + injectors
only.

**Out of scope this cycle:** ai-critic `BuildManagedEnv` wiring, changing
existing `Look` / `Result.Via` strings, real bash `-lic` e2e, sealed suite
`./tests/lookpath/`.

# DSN (Domain Specific Notion)

Shared pure library: batch-resolve multiple bare CLI names under a thin PATH
and turn found paths into unique directory lists for managed process env.

### Participants

- **`LookupPaths(names, opts)`** — one `LookupItem` per input name, same order;
  best-effort (missing names are items, not errors). Empty name in list → error.
- **`LookupItem`** — `Name`, `Path` (absolute if found), `Missing`, `From`
  (`"bash"` | `"zsh"` when login wins; `""` for path/dirs/candidates).
- **`LookupItems`** — slice type with helpers `Dirs()` and `DirsEnv()`.
- **`Dirs()`** — unique `filepath.Dir(Path)` for non-missing items, first-seen
  order.
- **`DirsEnv()`** — join of `Dirs()` with `os.PathListSeparator`; empty string
  when no dirs.
- **`Options`** — same injectables as `Look`: `Home`, `ExtraDirs`,
  `ExtraCandidates`, `Shells`, `Timeout`, `LookPath`, `IsExecutable`,
  `RunLogin`.

### Behaviors

**Per-name resolution (cheap stages, then batch login)**

1. PATH (`opts.LookPath`) → found, `From=""`
2. ExtraDirs + name → `From=""`
3. DefaultDirs(home) + name → `From=""`
4. ExtraCandidates → `From=""`
5. Remaining names → login shells in `Shells` order (default bash, zsh) via
   `RunLogin` → `From` = shell basename (`bash`/`zsh`)
6. Still missing → `Missing=true`, `Path=""`, `From=""`

**Rules**

- Input order preserved; no name dedupe.
- Empty `names` → empty items, nil error.
- Empty string element in `names` → error (no partial items contract).
- Best-effort: missing names never make `LookupPaths` return error.
- Never `Missing && Path != ""`; never `!Missing && Path == ""`.
- Parallel-safe: injectables only; product never `Setenv`/`Chdir`.

## Decision Tree

```
shell/lookpath/tests/lookup-paths/
├── DOCTEST.md
├── SETUP.md
├── lookup-paths/                         # LookupPaths(names, opts)
│   ├── empty-names/                      # [] → empty items, nil err
│   ├── empty-name-error/                 # "" element → error
│   ├── order-preserved/                  # multi-name order + Name fields
│   ├── via-path/                         # injected LookPath; From=""
│   ├── via-extra-dir/                    # ExtraDirs hit; From=""
│   ├── mixed-found-missing/              # some Missing true, no error
│   └── via-login/
│       ├── bash-hit/                     # From=bash
│       ├── bash-fail-zsh-hit/            # From=zsh
│       └── batch-remaining/              # two names via login; From set
├── dirs/                                 # LookupItems.Dirs()
│   ├── unique-first-seen/                # two bins same dir → one Dir
│   └── all-missing-empty/                # all Missing → empty dirs
└── dirs-env/                             # LookupItems.DirsEnv()
    ├── join-separator/                   # PathListSeparator between two dirs
    └── empty-string/                     # no dirs → ""
```

### Parameter significance (high → low)

1. **Operation** — lookup-paths | dirs | dirs-env
2. **Outcome / stage** — empty | error | order | path | extra_dir | login |
   mixed | unique dirs | empty helpers
3. **Inject surface** — LookPath map, ExtraDirs layout, RunLogin shell maps
4. **Name multiplicity** — zero / one / multi / empty-element

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `lookup-paths/empty-names` | LookupPaths | Empty names → empty items, nil err | RED |
| 2 | `lookup-paths/empty-name-error` | LookupPaths | Empty string in names → error | RED |
| 3 | `lookup-paths/order-preserved` | LookupPaths | Multi-name order and Name fields preserved | RED |
| 4 | `lookup-paths/via-path` | LookupPaths | Injected LookPath hit; From="" | RED |
| 5 | `lookup-paths/via-extra-dir` | LookupPaths | ExtraDirs hit; From="" | RED |
| 6 | `lookup-paths/mixed-found-missing` | LookupPaths | Found + Missing; err nil | RED |
| 7 | `lookup-paths/via-login/bash-hit` | LookupPaths | RunLogin bash → From=bash | RED |
| 8 | `lookup-paths/via-login/bash-fail-zsh-hit` | LookupPaths | bash fail, zsh → From=zsh | RED |
| 9 | `lookup-paths/via-login/batch-remaining` | LookupPaths | Two remaining names via login | RED |
| 10 | `dirs/unique-first-seen` | Dirs | Two tools same bin dir → one Dir | RED |
| 11 | `dirs/all-missing-empty` | Dirs | All missing → empty slice | RED |
| 12 | `dirs-env/join-separator` | DirsEnv | Two dirs joined with PathListSeparator | RED |
| 13 | `dirs-env/empty-string` | DirsEnv | No found paths → "" | RED |

## How to Run

```sh
# from shell/lookpath package dir (or go-pkgs module root with path)
doctest vet ./tests/lookup-paths
doctest test ./tests/lookup-paths
doctest test -v ./tests/lookup-paths/lookup-paths/via-path
doctest test -v ./tests/lookup-paths/dirs
```

Classic TDD: expect **RED** (compile failure until `LookupPaths` / `LookupItem`
/ `LookupItems` / `Dirs` / `DirsEnv` exist; then assert failures against
incomplete implementations) until implementer lands the API.

### Intended public API (implementer pins names)

```go
package lookpath

func LookupPaths(names []string, opts Options) (LookupItems, error)

type LookupItem struct {
	Name    string // requested name
	Path    string // absolute binary path if found; "" if Missing
	Missing bool   // true if not found
	From    string // "bash" | "zsh" if found via login shell; "" otherwise
}

type LookupItems []LookupItem

func (items LookupItems) Dirs() []string
func (items LookupItems) DirsEnv() string
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
	Operation string // lookup-paths | dirs | dirs-env

	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// LookupPaths inputs
	Names           []string
	Home            string
	ExtraDirs       []string
	ExtraCandidates []string
	Shells          []string
	Timeout         time.Duration

	// LookPathHits: name → absolute path. Absent/empty = injectable miss.
	LookPathHits map[string]string

	// Optional override map for IsExecutable injectable (path → bool).
	// When nil, harness uses real mode bits via harnessIsExecutable.
	ExecOverride map[string]bool

	// RunLogin injectable fixtures.
	// LoginStdout: shell → stdout path string (single-path / whole stdout).
	// LoginStdoutByName: shell → name → path (for multi-name login probes).
	// LoginFail: shell → fail.
	LoginStdout       map[string]string
	LoginStdoutByName map[string]map[string]string
	LoginFail         map[string]bool
}

// Response observes API outputs and injection spies.
type Response struct {
	Items   lookpath.LookupItems
	Dirs    []string
	DirsEnv string

	LookPathCalls    []string
	IsExecCalls      []string
	RunLoginCalls    []string // shell names in call order
	RunLoginCommands []string // command strings in call order
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}
	opts := buildOpts(req, resp)

	switch req.Operation {
	case "lookup-paths":
		items, err := lookpath.LookupPaths(req.Names, opts)
		resp.Items = items
		return resp, err

	case "dirs":
		items, err := lookpath.LookupPaths(req.Names, opts)
		resp.Items = items
		if err != nil {
			return resp, err
		}
		resp.Dirs = items.Dirs()
		return resp, nil

	case "dirs-env":
		items, err := lookpath.LookupPaths(req.Names, opts)
		resp.Items = items
		if err != nil {
			return resp, err
		}
		resp.DirsEnv = items.DirsEnv()
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
		if req.LookPathHits != nil {
			if p, ok := req.LookPathHits[file]; ok && p != "" {
				return p, nil
			}
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
		if req.LoginStdoutByName != nil {
			if byName, ok := req.LoginStdoutByName[shell]; ok {
				// Prefer exact name match from "command -v <name>" style commands.
				for name, path := range byName {
					if name != "" && strings.Contains(command, name) {
						return path + "\n", nil
					}
				}
			}
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

func assertItemInvariants(t *testing.T, items lookpath.LookupItems) {
	t.Helper()
	for i, it := range items {
		if it.Missing && it.Path != "" {
			t.Fatalf("items[%d]: Missing && Path=%q", i, it.Path)
		}
		if !it.Missing && it.Path == "" {
			t.Fatalf("items[%d]: !Missing && Path empty (Name=%q)", i, it.Name)
		}
	}
}

func assertItemFound(t *testing.T, it lookpath.LookupItem, name, path, from string) {
	t.Helper()
	assertEqual(t, "Name", it.Name, name)
	assertEqual(t, "Missing", it.Missing, false)
	assertPathEqual(t, it.Path, path)
	assertEqual(t, "From", it.From, from)
}

func assertItemMissing(t *testing.T, it lookpath.LookupItem, name string) {
	t.Helper()
	assertEqual(t, "Name", it.Name, name)
	assertEqual(t, "Missing", it.Missing, true)
	assertEqual(t, "Path", it.Path, "")
	assertEqual(t, "From", it.From, "")
}
```
