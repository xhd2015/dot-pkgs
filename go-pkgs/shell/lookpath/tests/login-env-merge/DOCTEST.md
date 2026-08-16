# shell/lookpath — ResolveLoginEnvs + MergeEnvs

## Version

0.0.2

Classic TDD doctests for detected-shell login env dump and pure env merge on
package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`.

The symbols under test do **not** exist yet:

- `ResolveLoginEnvs(opts LoginEnvOptions) (shell string, envs []string, err error)`
- `MergeEnvs(envs ...[]string) []string`
- `LoginEnvOptions.DetectShell` (injectable; nil → `detect.Shell()`)

Root `Run` calls them so the suite is **RED** (compile failure) until the
implementer lands the API.

**Default layer: L2** in-process library API. Injectable `RunLogin` +
`DetectShell`. No process env/cwd mutation, no `t.Setenv` / `t.Chdir`, no real
login shell. Parallel-safe: `t.TempDir()` + injectors only.

**Sealed (do not modify):** `./tests/login-env/` (P1 GREEN), P1 pathfmt
`ShortEnv`, `./tests/gopath/`, `./tests/lookpath/`, `./tests/lookup-paths/`.

# DSN (Domain Specific Notion)

Detect the user shell, dump one login interactive environ (or cascade bash→zsh
when unknown), then optionally merge env slices with last-wins semantics —
without mutating process env/cwd.

### Participants

- **`ResolveLoginEnvs`** — detect shell → dump login environ via `RunLogin` +
  `env -0` path shared with `Resolve*LoginEnvs`. Returns `(shellName, envs, err)`.
- **`MergeEnvs`** — pure merge of `[]string` KEY=value slices; later wins;
  first-seen key order. No I/O; no bash/zsh policy baked in.
- **`LoginEnvOptions`** — `Home`, `Timeout`, `ShellBin`, injectable `RunLogin`,
  injectable **`DetectShell func() string`** (nil → `detect.Shell()`).
- **`DetectShell`** — returns `"bash"` | `"zsh"` | `""` (or other → treat as
  unknown). Tests always inject; never `t.Setenv("SHELL")`.
- **`RunLogin`** — production: `shell -lic` + `env -0`; tests inject per-shell
  stdout/error for bash/zsh cascade.

### Behaviors

**ResolveLoginEnvs**

1. Detect via `opts.DetectShell` if non-nil, else `detect.Shell()`.
2. **bash / zsh detected** — dump that shell only (`ShellBin` overrides binary;
   logical shell name still bash/zsh). Return `(shellName, envs, err)`.
3. **unknown** (`""` or anything else) — try **bash**, then **zsh** on dump
   **error or empty** `envs`. First non-empty success wins. If both fail,
   return the last error (or bash error if zsh also empty/fail).
4. Never mutate process env/cwd.

**MergeEnvs**

- Pure. Later slice wins on same key; empty value overwrites.
- Within one slice, last occurrence wins.
- Key order: first-seen; overwrite in place; append keys only seen later.
- Nil / omitted slices skipped. `MergeEnvs()` → empty result (`nil` or
  zero-length; asserts use `len == 0`).
- Caller policy chooses order (e.g. `MergeEnvs(bash, zsh)` ⇒ zsh wins).

## Decision Tree

```
shell/lookpath/tests/login-env-merge/
├── resolve/                              # ResolveLoginEnvs
│   ├── detected-bash/
│   │   ├── hit/                          # Detect=bash; FOO+GOPATH dump
│   │   └── run-login-error/              # Detect=bash; RunLogin fail → error
│   ├── detected-zsh/
│   │   └── hit/                          # Detect=zsh; dump → shell=zsh
│   └── unknown/                          # Detect="" or other → bash then zsh
│       ├── bash-nonempty/                # bash keys; single RunLogin
│       ├── bash-empty-then-zsh/          # bash empty → zsh keys
│       ├── bash-error-then-zsh/          # bash err → zsh keys
│       ├── both-error/                   # both fail → error
│       └── detect-other/                 # Detect=fish → same cascade as ""
└── merge/                                # MergeEnvs pure
    ├── later-wins/                       # FOO=1 then FOO=2 → FOO=2
    ├── union-append/                     # FOO then BAR; order preserved
    ├── empty-overwrites/                 # FOO= then empty value
    ├── last-in-slice/                    # FOO=1,FOO=2 in one slice
    ├── first-seen-order/                 # overwrite keeps first-seen index
    ├── skip-nil-slice/                   # nil middle slice ignored
    └── no-args/                          # MergeEnvs() → empty
```

### Parameter significance (high → low)

1. **Op** — resolve (`ResolveLoginEnvs`) | merge (`MergeEnvs`)
2. **Detect result** (resolve) — bash | zsh | unknown/other
3. **Dump outcome** — hit | empty cascade | error cascade
4. **Merge rule** — overwrite | union | empty | order | nil | no-args

## Test Index

| # | Leaf | Op | Description |
|---|------|-----|-------------|
| 1 | `resolve/detected-bash/hit` | resolve | Detect bash; NUL dump FOO+GOPATH |
| 2 | `resolve/detected-bash/run-login-error` | resolve | Detect bash; RunLogin fail → error |
| 3 | `resolve/detected-zsh/hit` | resolve | Detect zsh; dump → shell=zsh |
| 4 | `resolve/unknown/bash-nonempty` | resolve | Detect ""; bash nonempty; one call |
| 5 | `resolve/unknown/bash-empty-then-zsh` | resolve | Bash empty → zsh keys; shell=zsh |
| 6 | `resolve/unknown/bash-error-then-zsh` | resolve | Bash err → zsh keys; shell=zsh |
| 7 | `resolve/unknown/both-error` | resolve | Both fail → error |
| 8 | `resolve/unknown/detect-other` | resolve | Detect fish → bash then zsh cascade |
| 9 | `merge/later-wins` | merge | Later slice overwrites FOO |
| 10 | `merge/union-append` | merge | Disjoint keys; first-seen order |
| 11 | `merge/empty-overwrites` | merge | Empty value overwrites |
| 12 | `merge/last-in-slice` | merge | Last occurrence in one slice wins |
| 13 | `merge/first-seen-order` | merge | Overwrite keeps first-seen key order |
| 14 | `merge/skip-nil-slice` | merge | Nil slice skipped |
| 15 | `merge/no-args` | merge | `MergeEnvs()` empty result |

## How to Run

From the **my** worktree root:

```sh
doctest vet ./external/dot-pkgs-master-2026-08-16/go-pkgs/shell/lookpath/tests/login-env-merge/
doctest test ./external/dot-pkgs-master-2026-08-16/go-pkgs/shell/lookpath/tests/login-env-merge/
```

From the go-pkgs module root:

```sh
doctest vet ./shell/lookpath/tests/login-env-merge/
doctest test ./shell/lookpath/tests/login-env-merge/
```

### Intended public API (implementer pins names)

```go
package lookpath

type LoginEnvOptions struct {
	Home     string
	Timeout  time.Duration
	RunLogin func(shell, command string, env []string) (stdout string, err error)
	ShellBin string
	// DetectShell, when non-nil, returns "bash"|"zsh"|"" (or other → unknown).
	// nil → detect.Shell() from $SHELL basename.
	DetectShell func() string
}

func ResolveLoginEnvs(opts LoginEnvOptions) (shell string, envs []string, err error)
func MergeEnvs(envs ...[]string) []string
```

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// Request is filled root→leaf. Op selects ResolveLoginEnvs vs MergeEnvs.
type Request struct {
	Op string // "resolve" | "merge"

	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// LoginEnvOptions inputs (resolve).
	Home     string
	Timeout  time.Duration
	ShellBin string

	// DetectShellResult is returned by the injected DetectShell func.
	DetectShellResult string

	// Per-shell RunLogin inject fixtures (env -0 style dumps).
	BashStdout string
	BashFail   bool
	ZshStdout  string
	ZshFail    bool

	// MergeInputs is the variadic argument list for MergeEnvs.
	// When MergeNoArgs is true, Run calls MergeEnvs() with zero args.
	MergeInputs [][]string
	MergeNoArgs bool
}

// Response observes API outputs and injection spies.
type Response struct {
	// resolve
	Shell string
	Envs  []string

	// merge
	Merged []string

	RunLoginCalls    []string
	DetectShellCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}

	switch req.Op {
	case "merge":
		if req.MergeNoArgs {
			resp.Merged = lookpath.MergeEnvs()
			return resp, nil
		}
		resp.Merged = lookpath.MergeEnvs(req.MergeInputs...)
		return resp, nil

	case "resolve":
		opts := buildLoginEnvOpts(req, resp)
		shell, envs, err := lookpath.ResolveLoginEnvs(opts)
		resp.Shell = shell
		resp.Envs = envs
		return resp, err

	default:
		return nil, fmt.Errorf("unknown Op %q", req.Op)
	}
}

func buildLoginEnvOpts(req *Request, resp *Response) lookpath.LoginEnvOptions {
	opts := lookpath.LoginEnvOptions{
		Home:     req.Home,
		Timeout:  req.Timeout,
		ShellBin: req.ShellBin,
	}

	// Always inject DetectShell so tests never depend on process $SHELL.
	opts.DetectShell = func() string {
		resp.DetectShellCalls++
		return req.DetectShellResult
	}

	opts.RunLogin = func(shell, command string, env []string) (string, error) {
		_ = command
		_ = env
		resp.RunLoginCalls = append(resp.RunLoginCalls, shell)
		// ShellBin may replace the binary name; map via basename for fixtures.
		base := filepath.Base(shell)
		base = strings.TrimSuffix(base, ".exe")
		switch base {
		case "bash":
			if req.BashFail {
				return "", fmt.Errorf("injected bash login failure")
			}
			return req.BashStdout, nil
		case "zsh":
			if req.ZshFail {
				return "", fmt.Errorf("injected zsh login failure")
			}
			return req.ZshStdout, nil
		default:
			return "", fmt.Errorf("injected unexpected shell %q", shell)
		}
	}

	return opts
}

// nulEnvDump builds an env -0 style stdout: KEY=value pairs joined by NUL,
// trailing NUL when non-empty (matches common env -0 output).
func nulEnvDump(pairs ...string) string {
	if len(pairs) == 0 {
		return ""
	}
	return strings.Join(pairs, "\x00") + "\x00"
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

func assertEnvsContain(t *testing.T, envs []string, want ...string) {
	t.Helper()
	set := make(map[string]bool, len(envs))
	for _, e := range envs {
		set[e] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Fatalf("Envs missing %q; got %#v", w, envs)
		}
	}
}

func assertRunLoginOrder(t *testing.T, resp *Response, want ...string) {
	t.Helper()
	if len(resp.RunLoginCalls) != len(want) {
		t.Fatalf("RunLoginCalls = %#v, want %#v", resp.RunLoginCalls, want)
	}
	for i := range want {
		if resp.RunLoginCalls[i] != want[i] {
			t.Fatalf("RunLoginCalls = %#v, want %#v", resp.RunLoginCalls, want)
		}
	}
}

func assertShell(t *testing.T, resp *Response, want string) {
	t.Helper()
	assertEqual(t, "Shell", resp.Shell, want)
}

func assertMergedEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("Merged = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Merged = %#v, want %#v", got, want)
		}
	}
}

func assertMergedEmpty(t *testing.T, got []string) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("Merged = %#v, want empty (nil or len 0)", got)
	}
}
```
