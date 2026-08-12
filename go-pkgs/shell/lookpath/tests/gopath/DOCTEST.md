# shell/lookpath — ResolveGoPath cascade

## Version

0.0.2

Classic TDD doctests for `ResolveGoPath` / `ResolveGoPathWith` on package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`.

P1 login-env APIs (`LoginEnvOptions`, `ResolveBashLoginEnv`,
`ResolveZshLoginEnv`) already exist and are GREEN. The symbols under test for
this suite (`GoPathOptions`, `ResolveGoPath`, `ResolveGoPathWith`) do **not**
exist yet. Root `Run` imports them so the suite is **RED** (compile failure)
until the implementer lands the public API.

**Default layer: L2** in-process library API with injectables
(`LoginEnv.RunLogin`, `LookPath`, `RunGoEnv`, `Home`). No process env/cwd
mutation, no `t.Setenv` / `t.Chdir`, no real login-shell or `go env` spawn.
Parallel-safe: `t.TempDir()` + injectors only.

**Out of scope this cycle:** process `os.Getenv("GOPATH")` as a cascade step,
P3 localbot refactor, real-shell / real-go e2e, sealed suites
`./tests/lookpath/`, `./tests/lookup-paths/`, `./tests/login-env/`.

# DSN (Domain Specific Notion)

Resolve a usable GOPATH for managed tool discovery when the process env is
thin (GUI / Launch Services), without mutating process env/cwd.

### Participants

- **`ResolveGoPath` / `ResolveGoPathWith`** — cascade until a non-empty path
  (first multi-GOPATH segment) or final `~/go`.
- **`GoPathOptions`** — `LoginEnv` (login shell discovery), injectable
  `LookPath("go")`, injectable `RunGoEnv(goBin)` → trimmed `go env GOPATH`
  stdout, and `Home` for `~/go` fallback.
- **`ResolveBashLoginEnv` / `ResolveZshLoginEnv`** — P1; used for stages 1–2
  with shared `LoginEnvOptions` (inject `RunLogin` in tests).
- **`LookPath` / `RunGoEnv`** — stage 3; miss/error/empty → soft continue.
- **Home fallback** — stage 4: `filepath.Join(home, "go")`; hard error only if
  home cannot be resolved.

### Behaviors

**Cascade (locked order)**

1. Bash login `GOPATH` — non-empty after `TrimSpace` → return first segment.
2. Zsh login `GOPATH` — same; bash empty/fail is soft continue.
3. Resolve `go` binary (`LookPath` or production look) then `RunGoEnv` /
   `go env GOPATH` — non-empty → first segment.
4. Final `filepath.Join(home, "go")` when earlier stages empty/soft-fail.

**Soft vs hard**

- Login shell failure or empty value → **continue**.
- `go` miss / `RunGoEnv` error / empty → **continue**.
- Final `~/go` → error only if home cannot be resolved.

**Multi-GOPATH**

- `a:b` → return `filepath.Clean` of the first segment only.

**Rules**

- Parallel-safe: inject `RunLogin`, `LookPath`, `RunGoEnv`, `Home` only.
- Product and harness never `Setenv` / `Chdir` / process PATH mutation.
- Process `os.Getenv("GOPATH")` is **not** a cascade step.

## Decision Tree

```
shell/lookpath/tests/gopath/
├── DOCTEST.md
├── SETUP.md
├── via-bash-login/                      # stage 1 wins
│   ├── hit/                             # single path; zsh/go short-circuit
│   └── multi-segment/                   # a:b → Clean(first)
├── via-zsh-login/                       # stage 2 wins
│   ├── after-bash-empty/                # bash empty → zsh hit
│   └── after-bash-error/                # bash RunLogin fail soft → zsh hit
├── via-go-env/                          # stage 3 wins
│   └── hit/                             # both login empty → go env hit
└── via-home-fallback/                   # stage 4
    ├── all-soft-empty/                  # login empty + go env empty → ~/go
    ├── go-miss/                         # LookPath fail → ~/go
    └── go-env-error/                    # RunGoEnv fail → ~/go
```

### Parameter significance (high → low)

1. **Cascade winner** — bash-login | zsh-login | go-env | home-fallback
2. **Soft-fail flavor** (within zsh / home) — empty vs error vs go-miss
3. **Value shape** — single path | multi-segment first

Home-unresolvable hard error is **not** a leaf: `GoPathOptions` exposes `Home`
but no injectable home resolver for `UserHomeDir` failure (not L2-testable
without product inject surface).

## Test Index

| # | Leaf | Winner | Description | Classic |
|---|------|--------|-------------|---------|
| 1 | `via-bash-login/hit` | bash | Inject bash `GOPATH=/tmp/from-bash` → that path; no zsh/go | RED |
| 2 | `via-bash-login/multi-segment` | bash | `GOPATH=/tmp/a:/tmp/b` → `/tmp/a` | RED |
| 3 | `via-zsh-login/after-bash-empty` | zsh | Bash empty → zsh hit; no go | RED |
| 4 | `via-zsh-login/after-bash-error` | zsh | Bash RunLogin error soft → zsh hit | RED |
| 5 | `via-go-env/hit` | go-env | Both login empty → LookPath + RunGoEnv hit | RED |
| 6 | `via-home-fallback/all-soft-empty` | home | All empty → `filepath.Join(Home,"go")` | RED |
| 7 | `via-home-fallback/go-miss` | home | LookPath fails → `~/go` | RED |
| 8 | `via-home-fallback/go-env-error` | home | RunGoEnv fails → `~/go` | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/lookpath/tests/gopath
doctest test ./shell/lookpath/tests/gopath
doctest test -v ./shell/lookpath/tests/gopath/via-bash-login/hit
doctest test -v ./shell/lookpath/tests/gopath/via-home-fallback
```

Classic TDD: expect **RED** (compile failure until `GoPathOptions` /
`ResolveGoPathWith` exist; then assert failures against incomplete
implementations) until implementer lands the API.

### Intended public API (implementer pins names)

```go
package lookpath

type GoPathOptions struct {
	// Login shell discovery (Home, Timeout, RunLogin, ShellBin)
	LoginEnv LoginEnvOptions
	// LookPath for "go"; nil → lookpath.Look / exec via Options
	LookPath func(file string) (string, error)
	// RunGoEnv(goBin) returns trimmed stdout of `go env GOPATH`
	RunGoEnv func(goBin string) (string, error)
	// Home for ~/go fallback; empty → UserHomeDir (also may come from LoginEnv.Home)
	Home string
}

func ResolveGoPath() (string, error)
func ResolveGoPathWith(opts GoPathOptions) (string, error)
```

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

// Request is filled root→leaf. Fixtures drive injectables for the cascade.
type Request struct {
	// WorkDir is an isolated temp root (root Setup).
	WorkDir string

	// Home is injected into GoPathOptions.Home and LoginEnv.Home.
	Home string

	// Per-shell RunLogin inject fixtures (env -0 style dumps).
	BashStdout string
	BashFail   bool
	ZshStdout  string
	ZshFail    bool

	// LookPath("go") inject: GoBin path when LookPathFail is false.
	GoBin        string
	LookPathFail bool

	// RunGoEnv inject: stdout of go env GOPATH; GoEnvFail forces error.
	GoEnvStdout string
	GoEnvFail   bool
}

// Response observes ResolveGoPathWith output and injection spies.
type Response struct {
	Path string

	RunLoginCalls []string // shell names in call order
	LookPathFiles []string // files passed to LookPath
	GoEnvBins     []string // goBin args to RunGoEnv
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	resp := &Response{}
	opts := buildGoPathOpts(req, resp)

	path, err := lookpath.ResolveGoPathWith(opts)
	resp.Path = path
	return resp, err
}

func buildGoPathOpts(req *Request, resp *Response) lookpath.GoPathOptions {
	opts := lookpath.GoPathOptions{
		Home: req.Home,
		LoginEnv: lookpath.LoginEnvOptions{
			Home: req.Home,
		},
	}

	opts.LoginEnv.RunLogin = func(shell, command string, env []string) (string, error) {
		_ = command
		_ = env
		resp.RunLoginCalls = append(resp.RunLoginCalls, shell)
		switch shell {
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
			// ShellBin overrides are out of scope; treat unknown as empty success.
			return "", nil
		}
	}

	opts.LookPath = func(file string) (string, error) {
		resp.LookPathFiles = append(resp.LookPathFiles, file)
		if req.LookPathFail {
			return "", fmt.Errorf("injected lookpath miss: %s", file)
		}
		if req.GoBin == "" {
			return "", fmt.Errorf("injected lookpath empty: %s", file)
		}
		return req.GoBin, nil
	}

	opts.RunGoEnv = func(goBin string) (string, error) {
		resp.GoEnvBins = append(resp.GoEnvBins, goBin)
		if req.GoEnvFail {
			return "", fmt.Errorf("injected go env failure: %s", goBin)
		}
		return req.GoEnvStdout, nil
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

func assertPath(t *testing.T, resp *Response, want string) {
	t.Helper()
	assertEqual(t, "Path", resp.Path, want)
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

func assertNoLookPath(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.LookPathFiles) != 0 {
		t.Fatalf("LookPath was called with %#v; want no LookPath (cascade short-circuit)", resp.LookPathFiles)
	}
}

func assertNoGoEnv(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.GoEnvBins) != 0 {
		t.Fatalf("RunGoEnv was called with %#v; want no RunGoEnv (cascade short-circuit)", resp.GoEnvBins)
	}
}

func assertLookPathGo(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.LookPathFiles) == 0 {
		t.Fatal("LookPath was not called; want file \"go\"")
	}
	if resp.LookPathFiles[0] != "go" {
		t.Fatalf("LookPathFiles[0] = %q, want %q", resp.LookPathFiles[0], "go")
	}
}

func homeGo(req *Request) string {
	return filepath.Join(req.Home, "go")
}
```
