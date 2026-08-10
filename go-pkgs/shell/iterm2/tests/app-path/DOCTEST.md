# shell/iterm2 — path-bound tell + ResolveAppPath (ITERM2_APP_PATH)

## Version
0.0.1

Nested library doctests for **app path resolve** and **path-bound AppleScript
tell headers** in `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2`.

Does **not** inherit the parent open-dir `Request`/`Run` from `../DOCTEST.md`
(Classic TDD: new symbols would break parent leaves’ compile).

| Phase | Status |
|-------|--------|
| **P1** `ResolveAppPathWith` + `TellApplicationHeader` + path-bound Build\*App | Classic TDD — **RED** until implementer lands APIs |

**Out of scope (P2+):** localbot dedupe, go.mod publish, install package, live dual-app E2E.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — opens iTerm2 / builds scripts; may set `ITERM2_APP_PATH`.
- **Resolver** — `ResolveAppPath` / `ResolveAppPathWith` picks the preferred
  existing `iTerm.app` bundle path (or empty).
- **Tell header** — `TellApplicationHeader(appPath)` emits the AppleScript
  `tell application …` line (path-bound POSIX file vs bare `"iTerm2"`).
- **Script builders** — open / force-new / smoke (and other package tells)
  prefix scripts with the same header.

### Resolve order (locked)

1. **`ITERM2_APP_PATH`** — if env is set (non-empty after trim) **and** the path
   is a usable app dir → return that path.
2. Else if env is set but **unusable** (missing / not a dir) → **no fallthrough**
   (return `""`). Matches localbot `ResolveITerm2App` strictness: intentional
   sandbox / override must not silently pick host Applications.
3. Else (env unset/empty): **`~/Applications/iTerm.app`** if present.
4. Else: **`/Applications/iTerm.app`** if present.
5. Else: `""`.

`IsInstalled` / “installed” remains path-only (`ResolveAppPath() != ""`).
Bare name is **open script fallback only** when resolve is empty.

### Tell header (locked)

- Non-empty `appPath` → path-bound:
  `tell application (POSIX file "<escaped>" as text)`
  where path is escaped via `EscapePathForAppleScript`. Must **not** use bare
  `tell application "iTerm2"` as the tell target.
- Empty `appPath` → bare fallback: `tell application "iTerm2"`.

### Script builders (locked)

All package tells share the header helper. Production builders may call
`TellApplicationHeader(ResolveAppPath())`. For L2 exact-path asserts without
host FS coupling, export explicit-path builders (names may be refined if intent
holds):

```go
BuildForceNewWindowScriptApp(appPath, dirPath string, followUps ...string) string
BuildScriptApp(appPath, dirPath string, followUps ...string) string
BuildPathScanSmokeScriptApp(appPath string) string
```

Existing `BuildForceNewWindowScript` / `BuildScript` / `BuildPathScanSmokeScript`
should delegate to these (or equivalent) with `ResolveAppPath()`.

### Injection model (parallel-safe)

**No** `t.Setenv` / process-global env for resolve order tests. Inject via opts:

```go
const EnvITerm2AppPath = "ITERM2_APP_PATH" // optional export; value is the name

type ResolveAppPathOpts struct {
	// Getenv reads env; nil => os.Getenv. Tests inject a closure.
	Getenv func(key string) string
	// Home returns user home for ~/Applications candidate; nil => os.UserHomeDir.
	// Empty home skips the home candidate.
	Home func() string
	// IsApp reports whether path is a usable iTerm.app bundle directory.
	// nil => os.Stat + IsDir (same idea as resolveAppPathAmong).
	IsApp func(path string) bool
}

func ResolveAppPathWith(opts ResolveAppPathOpts) string
func ResolveAppPath() string // ResolveAppPathWith(zero opts) / production defaults

func TellApplicationHeader(appPath string) string
```

### Behaviors under test

| Area | Expect |
|------|--------|
| env-wins | Usable `ITERM2_APP_PATH` wins over home+system |
| prefer-home | Both home and system present, env unset → home |
| system-only | Only system present → `/Applications/iTerm.app` |
| empty | None present, env unset → `""` |
| env-missing | Env set to missing path → `""` (no home/system fallthrough) |
| path-bound header | Non-empty path → POSIX file tell; not bare `"iTerm2"` |
| bare-fallback header | Empty path → `tell application "iTerm2"` |
| force-new / smart-open / smoke App builders | Embed path-bound header for explicit app path |

## Decision Tree

```
app-path/
├── resolve/                        [Phase=resolve-app]
│   ├── env-wins/                   ITERM2_APP_PATH usable → that path
│   ├── prefer-home/                home + system → home
│   ├── system-only/                system only → /Applications/iTerm.app
│   ├── empty/                      none → ""
│   └── env-missing/                env set but missing → "" (no fallthrough)
├── tell-header/                    [Phase=tell-header]
│   ├── path-bound/                 POSIX file tell; not bare iTerm2
│   └── bare-fallback/              empty → tell application "iTerm2"
└── script/                         [Phase=build-*-app]
    ├── force-new-path-bound/       BuildForceNewWindowScriptApp + header
    ├── smart-open-path-bound/      BuildScriptApp + header
    └── smoke-path-bound/           BuildPathScanSmokeScriptApp + header
```

## Test Index

| Leaf | Phase | Description | Expect |
|------|-------|-------------|--------|
| `resolve/env-wins/` | resolve-app | Usable env wins over home+system | RED |
| `resolve/prefer-home/` | resolve-app | Home preferred over system | RED |
| `resolve/system-only/` | resolve-app | System when only system exists | RED |
| `resolve/empty/` | resolve-app | No candidates → empty | RED |
| `resolve/env-missing/` | resolve-app | Env set unusable → empty, no fallthrough | RED |
| `tell-header/path-bound/` | tell-header | POSIX file header | RED |
| `tell-header/bare-fallback/` | tell-header | Bare `"iTerm2"` when empty | RED |
| `script/force-new-path-bound/` | build-force-new-app | Force-new embeds path-bound tell | RED |
| `script/smart-open-path-bound/` | build-script-app | Smart-open embeds path-bound tell | RED |
| `script/smoke-path-bound/` | build-smoke-app | Smoke script embeds path-bound tell | RED |

## How to Run

```sh
# from go-pkgs module root (brought external path under consumer worktree)
doctest vet ./shell/iterm2/tests/app-path
doctest test ./shell/iterm2/tests/app-path
```

Expect: **RED** until implementer lands resolve opts, tell header, and path-bound
builders (compile fail on missing symbols and/or assert fail on bare tells).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// envITerm2AppPath is the discovery env name (product may export EnvITerm2AppPath).
const envITerm2AppPath = "ITERM2_APP_PATH"

// Request selects Phase and injectable resolve / script inputs.
// Parallel-safe: no process env mutation; Getenv/Home/IsApp built from fields.
type Request struct {
	// Phase:
	//   resolve-app | tell-header
	//   | build-force-new-app | build-script-app | build-smoke-app
	Phase string

	// --- resolve-app inject ---
	// EnvSet: when true, Getenv(ITERM2_APP_PATH) returns EnvValue (may be a
	// missing path for env-missing). When false, env is unset (Getenv returns "").
	EnvSet   bool
	EnvValue string
	// HomeDir is returned by Home(); empty skips ~/Applications candidate.
	HomeDir string
	// ExistingDirs are paths for which IsApp returns true (usable .app dirs).
	ExistingDirs []string

	// --- tell-header / build-*-app ---
	AppPath string
	Dir     string
}

// Response holds resolve / header / script outputs.
type Response struct {
	ResolvedPath string
	Header       string
	Script       string
}

// Run dispatches on req.Phase (Classic TDD — missing product symbols fail RED).
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Phase {
	case "resolve-app":
		opts := iterm2.ResolveAppPathOpts{
			Getenv: func(key string) string {
				if !req.EnvSet {
					return ""
				}
				// Product may export EnvITerm2AppPath == "ITERM2_APP_PATH".
				if key == envITerm2AppPath {
					return req.EnvValue
				}
				return ""
			},
			Home: func() string {
				return req.HomeDir
			},
			IsApp: func(path string) bool {
				for _, cand := range req.ExistingDirs {
					if cand == path {
						return true
					}
				}
				return false
			},
		}
		resp.ResolvedPath = iterm2.ResolveAppPathWith(opts)
		return resp, nil
	case "tell-header":
		resp.Header = iterm2.TellApplicationHeader(req.AppPath)
		return resp, nil
	case "build-force-new-app":
		dir := req.Dir
		if dir == "" {
			dir = "/tmp/iterm2-app-path-force-new"
		}
		resp.Script = iterm2.BuildForceNewWindowScriptApp(req.AppPath, dir)
		resp.Header = iterm2.TellApplicationHeader(req.AppPath)
		return resp, nil
	case "build-script-app":
		dir := req.Dir
		if dir == "" {
			dir = "/tmp/iterm2-app-path-smart-open"
		}
		resp.Script = iterm2.BuildScriptApp(req.AppPath, dir)
		resp.Header = iterm2.TellApplicationHeader(req.AppPath)
		return resp, nil
	case "build-smoke-app":
		resp.Script = iterm2.BuildPathScanSmokeScriptApp(req.AppPath)
		resp.Header = iterm2.TellApplicationHeader(req.AppPath)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```
