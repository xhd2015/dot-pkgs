# shell/codex/install — OpenAI Codex CLI install / update / ensure

## Version

0.0.2

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install`.

The production package does **not** exist yet. Root `Run` imports it so the suite
is **RED** (compile failure) until the implementer lands the public API under
`shell/codex/install/`.

**Default layer: L2** in-process library API with injectable HTTP, LookPath,
RunShell, RunVersion, and FetchLatest. No real network, no real `codex` binary,
no process env/cwd mutation. Parallel-safe: `t.TempDir()` only when needed.

**Out of scope this cycle:** spl landing wiring, iTerm changes, real npm network,
GitHub releases fallback, product-binary e2e.

## DSN (Domain Specific Notion)

### Participants

- **`ParseVersion(output)`** — extracts the first semver `X.Y.Z` from version
  command / package text (e.g. `codex-cli 0.147.0`, `@openai/codex, 0.122.0`).
  Empty or no match → error.
- **`NeedsUpdate(local, latest)`** — pure semver compare: true only when both
  parse as `X.Y.Z` and local < latest; false when equal, greater, empty, or
  unparseable.
- **`LatestVersion(ctx, opts)`** — GET `opts.URL` or `NPMLatestURL`
  (`https://registry.npmjs.org/@openai/codex/latest`); reads JSON `version`.
  Injectable `HTTPClient`. Non-2xx / bad JSON → error.
- **`LocalVersion(ctx, opts)`** — resolve bin via `opts.Bin` / `LookPath("codex")`,
  run version command via injectable `RunVersion`; returns raw command output.
- **`Install(ctx, opts)`** — runs `InstallCmd`
  (`curl -fsSL https://chatgpt.com/codex/install.sh | sh`) via injectable
  `RunShell`.
- **`Update(ctx, opts)`** — runs `UpdateCmd` (`codex update`) via injectable
  `RunShell` (path-qualified form allowed when `Bin` is set).
- **`Ensure(ctx, opts) (Result, error)`** — orchestrator:
  - missing bin → `Install`; `Action=install`
  - present + NeedsUpdate → `Update`; `Action=update`
  - present + !NeedsUpdate → noop; `Action=noop`
  - present + latest/local unknown → noop; `Action=noop` (no hard error required)
  - call `LatestVersion` / `FetchLatest` **only** when bin is present
- **Constants** — `InstallScriptURL`, `InstallCmd`, `UpdateCmd`, `NPMLatestURL`.

### Behaviors

**ParseVersion**

- `codex-cli 0.147.0` → `0.147.0`
- `@openai/codex, 0.122.0` → `0.122.0`
- empty / no `X.Y.Z` → error

**NeedsUpdate**

- local < latest → true; `==` / `>` / unparseable either side → false

**LatestVersion**

- 200 + `{"version":"0.147.0"}` → `0.147.0`
- HTTP 404 / transport failure → error

**LocalVersion**

- Injected runner success → raw stdout string
- Injected runner failure → error

**Install / Update**

- `Install` invokes `RunShell` exactly once with `InstallCmd`
- `Update` invokes `RunShell` exactly once with `UpdateCmd` (default)

**Ensure**

- LookPath miss → install path; `ShellCalls` has `InstallCmd` once; no latest fetch
- Present + local < latest → update path; `ShellCalls` has `UpdateCmd` once
- Present + local == latest → noop; no shell mutator
- Present + latest fetch fails → noop; no install/update shell call

## Decision Tree

```
shell/codex/install/tests/
├── DOCTEST.md
├── SETUP.md
├── parse-version/                         # ParseVersion (pure)
│   ├── codex-cli-prefix/                  # "codex-cli 0.147.0" → 0.147.0
│   ├── npm-package-form/                  # "@openai/codex, 0.122.0" → 0.122.0
│   ├── empty/                             # "" → error
│   └── garbage/                           # no semver → error
├── needs-update/                          # NeedsUpdate (pure)
│   ├── local-lt-latest/                   # true
│   ├── local-eq-latest/                   # false
│   ├── local-gt-latest/                   # false
│   └── unparseable/                       # empty/garbage → false
├── latest-version/                        # LatestVersion (fake HTTP)
│   ├── npm-json-ok/                       # 200 JSON version field
│   └── http-404/                          # 404 → error
├── local-version/                         # LocalVersion (inject runner)
│   ├── runner-success/                    # returns raw version stdout
│   └── runner-fail/                       # runner error → error
├── install-cmd/                           # Install (API Install; dir not "install" — avoids pkg name clash)
│   └── runs-install-cmd/                  # RunShell(InstallCmd) once
├── update/                                # Update
│   └── runs-update-cmd/                   # RunShell(UpdateCmd) once
└── ensure/                                # Ensure orchestration
    ├── missing-bin-installs/              # missing → install once; Action=install
    ├── present-outdated-updates/          # outdated → update once; Action=update
    ├── present-current-noop/              # current → no mutator; Action=noop
    └── present-latest-fail-noop/          # latest fail → noop; no install
```

### Parameter significance (high → low)

1. **Operation** — parse-version | needs-update | latest-version | local-version | install | update | ensure
2. **Outcome class** — happy vs error / install vs update vs noop
3. **Injection surface** — HTTP fixture, LookPath, RunShell, RunVersion, FetchLatest
4. **Version fixture strings** — local/latest pairs, version command output shapes

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `parse-version/codex-cli-prefix` | ParseVersion | `codex-cli 0.147.0` → `0.147.0` | RED |
| 2 | `parse-version/npm-package-form` | ParseVersion | `@openai/codex, 0.122.0` → `0.122.0` | RED |
| 3 | `parse-version/empty` | ParseVersion | empty → error | RED |
| 4 | `parse-version/garbage` | ParseVersion | no semver → error | RED |
| 5 | `needs-update/local-lt-latest` | NeedsUpdate | `0.1.0` < `0.2.0` → true | RED |
| 6 | `needs-update/local-eq-latest` | NeedsUpdate | equal → false | RED |
| 7 | `needs-update/local-gt-latest` | NeedsUpdate | local > latest → false | RED |
| 8 | `needs-update/unparseable` | NeedsUpdate | empty/garbage → false | RED |
| 9 | `latest-version/npm-json-ok` | LatestVersion | fake npm JSON → version | RED |
| 10 | `latest-version/http-404` | LatestVersion | HTTP 404 → error | RED |
| 11 | `local-version/runner-success` | LocalVersion | inject runner returns stdout | RED |
| 12 | `local-version/runner-fail` | LocalVersion | inject runner fails → error | RED |
| 13 | `install-cmd/runs-install-cmd` | Install | RunShell called once with InstallCmd | RED |
| 14 | `update/runs-update-cmd` | Update | RunShell called once with UpdateCmd | RED |
| 15 | `ensure/missing-bin-installs` | Ensure | missing bin → install; Action=install; no latest fetch | RED |
| 16 | `ensure/present-outdated-updates` | Ensure | present outdated → update; Action=update | RED |
| 17 | `ensure/present-current-noop` | Ensure | present current → noop; no shell mutator | RED |
| 18 | `ensure/present-latest-fail-noop` | Ensure | latest fetch fails → noop; no install | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/codex/install/tests
doctest test ./shell/codex/install/tests
doctest test -v ./shell/codex/install/tests/parse-version/codex-cli-prefix
doctest test -v ./shell/codex/install/tests/ensure
```

Classic TDD: expect **RED** (compile failure until package exists, then assert
failures against incomplete implementations) until implementer lands the API.

### Intended public API (implementer pins names)

```go
package install

const (
	InstallScriptURL = "https://chatgpt.com/codex/install.sh"
	InstallCmd       = `curl -fsSL https://chatgpt.com/codex/install.sh | sh`
	UpdateCmd        = "codex update"
	NPMLatestURL     = "https://registry.npmjs.org/@openai/codex/latest"
)

func ParseVersion(output string) (string, error)
func NeedsUpdate(local, latest string) bool
func LatestVersion(ctx context.Context, opts LatestVersionOpts) (string, error)
func LocalVersion(ctx context.Context, opts LocalVersionOpts) (string, error)
func Install(ctx context.Context, opts InstallOpts) error
func Update(ctx context.Context, opts UpdateOpts) error
func Ensure(ctx context.Context, opts EnsureOpts) (Result, error)

type LatestVersionOpts struct {
	URL        string // empty → NPMLatestURL
	HTTPClient *http.Client
}

type LocalVersionOpts struct {
	Bin        string
	LookPath   func(file string) (string, error)
	RunVersion func(ctx context.Context, bin string) (string, error)
}

type InstallOpts struct {
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

type UpdateOpts struct {
	Bin      string
	RunShell func(ctx context.Context, cmd string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

type EnsureOpts struct {
	Bin         string
	LookPath    func(file string) (string, error)
	RunShell    func(ctx context.Context, cmd string) error
	RunVersion  func(ctx context.Context, bin string) (string, error)
	FetchLatest func(ctx context.Context) (string, error) // nil → LatestVersion
	HTTPClient  *http.Client
	Stdout      io.Writer
	Stderr      io.Writer
}

type Result struct {
	Action        string // install | update | noop
	BinPath       string
	LocalVersion  string
	LatestVersion string
	NeedsUpdate   bool
}
```

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
)

// Request is filled root→leaf. Operation selects which public API Run calls.
type Request struct {
	Operation string // parse-version | needs-update | latest-version | local-version | install | update | ensure

	// WorkDir is an isolated temp root (root Setup). Used for fake bin paths.
	WorkDir string

	// ParseVersion
	VersionOutput string

	// NeedsUpdate
	LocalVer  string
	LatestVer string

	// LatestVersion HTTP fixture mode: npm-ok | http-404
	HTTPMode string
	// Optional absolute LatestURL override (empty → fake server root "/@openai/codex/latest").
	LatestURL string
	// JSON version field for npm-ok fixture (default 0.147.0).
	NPMVersion string

	// LocalVersion
	BinPath            string // injected resolved bin path
	VersionCmdOutput   string // injected RunVersion stdout
	VersionCmdFail     bool
	LookPathMiss       bool // LookPath returns not-found
	LookPathName       string

	// Install / Update / Ensure shell recording is always on in Run.
	// Ensure presence / versions:
	//   EnsurePresent=false → missing bin (LookPath miss)
	//   EnsurePresent=true  → LookPath returns BinPath
	EnsurePresent      bool
	EnsureLocalRaw     string // RunVersion stdout when present (e.g. "codex-cli 0.1.0")
	EnsureLatest       string // FetchLatest success value
	EnsureLatestFail   bool   // FetchLatest returns error
	EnsureRunVersionFail bool
}

// Response observes API outputs and injected side effects.
type Response struct {
	Version string // ParseVersion / LatestVersion / LocalVersion

	NeedsUpdate bool // NeedsUpdate pure result

	// Ensure Result fields
	Action              string
	BinPath             string
	LocalVersion        string
	LatestVersion       string
	ResultNeedsUpdate   bool

	// Injection spies (parallel-safe locals recorded into Response)
	ShellCalls       []string
	FetchLatestCalls int
	LookPathCalls    []string
	RunVersionCalls  []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}

	ctx := context.Background()
	resp := &Response{}

	switch req.Operation {
	case "parse-version":
		ver, err := install.ParseVersion(req.VersionOutput)
		resp.Version = ver
		return resp, err

	case "needs-update":
		resp.NeedsUpdate = install.NeedsUpdate(req.LocalVer, req.LatestVer)
		return resp, nil

	case "latest-version":
		srv, client := startNPMHTTPFixture(t, req)
		defer srv.Close()
		url := req.LatestURL
		if url == "" {
			url = srv.URL + "/@openai/codex/latest"
		}
		ver, err := install.LatestVersion(ctx, install.LatestVersionOpts{
			URL:        url,
			HTTPClient: client,
		})
		resp.Version = ver
		return resp, err

	case "local-version":
		bin := req.BinPath
		if bin == "" {
			bin = filepath.Join(req.WorkDir, "bin", "codex")
		}
		ver, err := install.LocalVersion(ctx, install.LocalVersionOpts{
			Bin: bin,
			LookPath: func(file string) (string, error) {
				resp.LookPathCalls = append(resp.LookPathCalls, file)
				if req.LookPathMiss {
					return "", fmt.Errorf("lookpath: %s: not found", file)
				}
				return bin, nil
			},
			RunVersion: func(ctx context.Context, b string) (string, error) {
				_ = ctx
				resp.RunVersionCalls = append(resp.RunVersionCalls, b)
				if req.VersionCmdFail {
					return "", fmt.Errorf("injected version command failure")
				}
				out := req.VersionCmdOutput
				if out == "" {
					out = "codex-cli 0.147.0"
				}
				return out, nil
			},
		})
		resp.Version = ver
		return resp, err

	case "install":
		err := install.Install(ctx, install.InstallOpts{
			RunShell: func(ctx context.Context, cmd string) error {
				_ = ctx
				resp.ShellCalls = append(resp.ShellCalls, cmd)
				return nil
			},
		})
		return resp, err

	case "update":
		err := install.Update(ctx, install.UpdateOpts{
			RunShell: func(ctx context.Context, cmd string) error {
				_ = ctx
				resp.ShellCalls = append(resp.ShellCalls, cmd)
				return nil
			},
		})
		return resp, err

	case "ensure":
		bin := req.BinPath
		if bin == "" {
			bin = filepath.Join(req.WorkDir, "bin", "codex")
		}
		result, err := install.Ensure(ctx, install.EnsureOpts{
			Bin: bin,
			LookPath: func(file string) (string, error) {
				resp.LookPathCalls = append(resp.LookPathCalls, file)
				if !req.EnsurePresent || req.LookPathMiss {
					return "", fmt.Errorf("lookpath: %s: not found", file)
				}
				return bin, nil
			},
			RunShell: func(ctx context.Context, cmd string) error {
				_ = ctx
				resp.ShellCalls = append(resp.ShellCalls, cmd)
				return nil
			},
			RunVersion: func(ctx context.Context, b string) (string, error) {
				_ = ctx
				resp.RunVersionCalls = append(resp.RunVersionCalls, b)
				if req.EnsureRunVersionFail {
					return "", fmt.Errorf("injected version command failure")
				}
				out := req.EnsureLocalRaw
				if out == "" {
					out = "codex-cli 0.147.0"
				}
				return out, nil
			},
			FetchLatest: func(ctx context.Context) (string, error) {
				_ = ctx
				resp.FetchLatestCalls++
				if req.EnsureLatestFail {
					return "", fmt.Errorf("injected latest fetch failure")
				}
				v := req.EnsureLatest
				if v == "" {
					v = "0.147.0"
				}
				return v, nil
			},
		})
		resp.Action = result.Action
		resp.BinPath = result.BinPath
		resp.LocalVersion = result.LocalVersion
		resp.LatestVersion = result.LatestVersion
		resp.ResultNeedsUpdate = result.NeedsUpdate
		return resp, err

	default:
		return nil, fmt.Errorf("unknown Operation %q", req.Operation)
	}
}

func startNPMHTTPFixture(t *testing.T, req *Request) (*httptest.Server, *http.Client) {
	t.Helper()
	ver := req.NPMVersion
	if ver == "" {
		ver = "0.147.0"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/@openai/codex/latest", func(w http.ResponseWriter, r *http.Request) {
		switch req.HTTPMode {
		case "http-404":
			http.NotFound(w, r)
			return
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"version": ver})
		}
	})
	srv := httptest.NewServer(mux)
	return srv, srv.Client()
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

func assertShellCalls(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ShellCalls = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ShellCalls[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
```
