# shell/lookpath — login shell environment resolve

## Version

0.0.2

Classic TDD doctests for login-shell environment dump APIs on package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`.

The symbols under test (`LoginEnvOptions`, `ResolveBashLoginEnvs`,
`ResolveZshLoginEnvs`, `ResolveBashLoginEnv`, `ResolveZshLoginEnv`) do **not**
exist yet on the production package. Root `Run` imports them so the suite is
**RED** (compile failure) until the implementer lands the public API.

**Default layer: L2** in-process library API with injectable `RunLogin`. No
process env/cwd mutation, no `t.Setenv` / `t.Chdir`, no real login-shell spawn.
Parallel-safe: `t.TempDir()` + injectors only.

**Out of scope this cycle:** `ResolveGoPath` (P2), `localbot.ResolveSplBinaryPath`
(P3), real-shell e2e, sealed suites `./tests/lookpath/` and `./tests/lookup-paths/`.

# DSN (Domain Specific Notion)

Capture a login interactive shell's environment (full dump or single variable)
without mutating process env/cwd — for thin GUI PATH / managed tool discovery.

### Participants

- **`ResolveBashLoginEnvs` / `ResolveZshLoginEnvs`** — full environ after login
  interactive shell; returns `[]string` like `os.Environ` (`KEY=value`).
- **`ResolveBashLoginEnv` / `ResolveZshLoginEnv`** — single variable from that
  dump; empty name → error; unset/empty value → `("", nil)`.
- **`LoginEnvOptions`** — `Home`, `Timeout`, `ShellBin` (optional; empty →
  bash/zsh by API), injectable `RunLogin(shell, command, env)`.
- **`RunLogin`** — production: `shell -lic` with minimal env (`PATH` system +
  `HOME`); dump via `env -0` (NUL-delimited). Tests always inject.

### Behaviors

**Full dump (`*LoginEnvs`)**

1. Call `RunLogin` with the target shell (`bash` / `zsh`, or `ShellBin` when set)
   and a dump command (production: `env -0`).
2. Parse NUL-delimited `KEY=value` stdout into `[]string` (`os.Environ` style).
3. `RunLogin` failure → non-nil error.

**Single variable (`*LoginEnv`)**

1. Empty `name` → error (no successful value).
2. Otherwise resolve dump (same path as Envs) and look up `name`.
3. Unset or empty value → `("", nil)` so callers can cascade.
4. Shell/run failure → non-nil error.

**Rules**

- Parallel-safe: isolation only via `LoginEnvOptions.RunLogin` / `Home`.
- Product never `Setenv` / `Chdir` / mutates process env or cwd.
- Harness never uses `t.Setenv` / `t.Chdir` / process PATH mutation.

## Decision Tree

```
shell/lookpath/tests/login-env/
├── DOCTEST.md
├── SETUP.md
├── envs/                                  # Resolve*LoginEnvs — full dump
│   ├── bash/
│   │   ├── multi-key/                     # FOO + GOPATH present as KEY=value
│   │   └── run-login-error/               # RunLogin fail → error
│   └── zsh/
│       ├── multi-key/
│       └── run-login-error/
└── env/                                   # Resolve*LoginEnv — single var
    ├── bash/
    │   ├── hit/                           # name present → value
    │   ├── miss-empty/                    # unset → ("", nil)
    │   ├── empty-name-error/              # "" → error
    │   └── run-login-error/               # RunLogin fail → error
    └── zsh/
        ├── hit/
        ├── miss-empty/
        ├── empty-name-error/
        └── run-login-error/
```

### Parameter significance (high → low)

1. **Operation** — envs (full dump) | env (single variable)
2. **Shell** — bash | zsh
3. **Outcome** — multi-key / hit | miss-empty | empty-name | run-login-error

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `envs/bash/multi-key` | ResolveBashLoginEnvs | Inject NUL dump with FOO + GOPATH → both present | RED |
| 2 | `envs/bash/run-login-error` | ResolveBashLoginEnvs | RunLogin failure → error | RED |
| 3 | `envs/zsh/multi-key` | ResolveZshLoginEnvs | Inject NUL dump; RunLogin shell=zsh | RED |
| 4 | `envs/zsh/run-login-error` | ResolveZshLoginEnvs | RunLogin failure → error | RED |
| 5 | `env/bash/hit` | ResolveBashLoginEnv | FOO present → `"1"` | RED |
| 6 | `env/bash/miss-empty` | ResolveBashLoginEnv | FOO unset → `("", nil)` | RED |
| 7 | `env/bash/empty-name-error` | ResolveBashLoginEnv | Empty name → error | RED |
| 8 | `env/bash/run-login-error` | ResolveBashLoginEnv | RunLogin failure → error | RED |
| 9 | `env/zsh/hit` | ResolveZshLoginEnv | FOO present → `"1"` | RED |
| 10 | `env/zsh/miss-empty` | ResolveZshLoginEnv | FOO unset → `("", nil)` | RED |
| 11 | `env/zsh/empty-name-error` | ResolveZshLoginEnv | Empty name → error | RED |
| 12 | `env/zsh/run-login-error` | ResolveZshLoginEnv | RunLogin failure → error | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/lookpath/tests/login-env
doctest test ./shell/lookpath/tests/login-env
doctest test -v ./shell/lookpath/tests/login-env/envs/bash/multi-key
doctest test -v ./shell/lookpath/tests/login-env/env
```

Classic TDD: expect **RED** (compile failure until `LoginEnvOptions` /
`Resolve*LoginEnv(s)` exist; then assert failures against incomplete
implementations) until implementer lands the API.

### Intended public API (implementer pins names)

```go
package lookpath

import "time"

type LoginEnvOptions struct {
	Home     string
	Timeout  time.Duration
	RunLogin func(shell, command string, env []string) (stdout string, err error)
	ShellBin string // optional; empty → "bash" / "zsh" by API
}

// Full environ after login interactive shell: []string "KEY=value" (os.Environ style).
func ResolveBashLoginEnvs(opts LoginEnvOptions) ([]string, error)
func ResolveZshLoginEnvs(opts LoginEnvOptions) ([]string, error)

// Single variable. Empty name → error.
// Unset or empty value → ("", nil) so callers can cascade.
// Shell/run failure → non-nil error.
func ResolveBashLoginEnv(name string, opts LoginEnvOptions) (string, error)
func ResolveZshLoginEnv(name string, opts LoginEnvOptions) (string, error)
```

```go
import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// Request is filled root→leaf. Operation + Shell select which public API Run calls.
type Request struct {
	Operation string // envs | env
	Shell     string // bash | zsh

	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// LoginEnvOptions inputs
	Home     string
	Timeout  time.Duration
	ShellBin string

	// EnvName is the single-variable name for Operation=env.
	EnvName string

	// RunLogin injectable fixtures.
	// LoginStdout is a production-style env -0 dump (NUL-delimited KEY=value).
	// LoginFail forces RunLogin to return an error.
	LoginStdout string
	LoginFail   bool
}

// Response observes API outputs and injection spies.
type Response struct {
	Envs  []string // full dump (Operation=envs)
	Value string   // single var (Operation=env)

	RunLoginCalls    []string   // shell names in call order
	RunLoginCommands []string   // command strings in call order
	RunLoginEnvs     [][]string // env slices passed to RunLogin
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}
	opts := buildLoginEnvOpts(req, resp)

	switch {
	case req.Operation == "envs" && req.Shell == "bash":
		envs, err := lookpath.ResolveBashLoginEnvs(opts)
		resp.Envs = envs
		return resp, err

	case req.Operation == "envs" && req.Shell == "zsh":
		envs, err := lookpath.ResolveZshLoginEnvs(opts)
		resp.Envs = envs
		return resp, err

	case req.Operation == "env" && req.Shell == "bash":
		val, err := lookpath.ResolveBashLoginEnv(req.EnvName, opts)
		resp.Value = val
		return resp, err

	case req.Operation == "env" && req.Shell == "zsh":
		val, err := lookpath.ResolveZshLoginEnv(req.EnvName, opts)
		resp.Value = val
		return resp, err

	default:
		return nil, fmt.Errorf("unknown Operation/Shell %q/%q", req.Operation, req.Shell)
	}
}

func buildLoginEnvOpts(req *Request, resp *Response) lookpath.LoginEnvOptions {
	opts := lookpath.LoginEnvOptions{
		Home:     req.Home,
		Timeout:  req.Timeout,
		ShellBin: req.ShellBin,
	}

	opts.RunLogin = func(shell, command string, env []string) (string, error) {
		resp.RunLoginCalls = append(resp.RunLoginCalls, shell)
		resp.RunLoginCommands = append(resp.RunLoginCommands, command)
		// Copy env so later mutation does not race with assertions.
		envCopy := append([]string(nil), env...)
		resp.RunLoginEnvs = append(resp.RunLoginEnvs, envCopy)

		if req.LoginFail {
			return "", fmt.Errorf("injected login shell failure: %s", shell)
		}
		return req.LoginStdout, nil
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

func assertRunLoginShell(t *testing.T, resp *Response, wantShell string) {
	t.Helper()
	if len(resp.RunLoginCalls) == 0 {
		t.Fatalf("RunLogin was not called; want shell %q", wantShell)
	}
	if resp.RunLoginCalls[0] != wantShell {
		t.Fatalf("RunLoginCalls[0] = %q, want %q (full=%#v)", resp.RunLoginCalls[0], wantShell, resp.RunLoginCalls)
	}
}

func assertRunLoginEnvHasHome(t *testing.T, resp *Response, home string) {
	t.Helper()
	if home == "" {
		return
	}
	if len(resp.RunLoginEnvs) == 0 {
		t.Fatal("RunLogin env not recorded")
	}
	want := "HOME=" + home
	for _, e := range resp.RunLoginEnvs[0] {
		if e == want {
			return
		}
	}
	t.Fatalf("RunLogin env missing %q; got %#v", want, resp.RunLoginEnvs[0])
}
```
