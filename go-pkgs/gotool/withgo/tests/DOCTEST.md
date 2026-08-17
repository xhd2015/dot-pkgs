# withgo — pin a Go toolchain and run under its GOROOT

## Version

0.0.2

Library doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo`. Each
leaf calls `withgo` APIs in-process via root `Run` — no CLI, no network, no
`go run download-go`.

**Classic TDD:** the `withgo` package does not exist yet. Leaves are compile-RED
until the implementer lands the library. Do not implement production code in
this tree.

# DSN (Domain Specific Notion)

Pins a go version to a known patch, optionally reads a module's `go` line,
resolves an on-disk GOROOT under an install dir, and execs a child with that
toolchain. Callers inject install dir, an install hook, and writers.

**Participants**

- **`PinPatch`** — maps a go version spelling to a pinned SDK directory name
  (`go1.19` → `go1.19.13`). Known major.minor rows come from kool
  `ResolveGoroot`. Naked `1.19` matches `go1.19`. Already-full patch is
  identity with a `go` prefix. Unknown major.minor is left unchanged but still
  `go`-prefixed.
- **`ModuleGoLine`** — reads `go.mod` in a module dir and returns major.minor
  as `go1.19`. Missing `go` line or missing `go.mod` is an error.
- **`DefaultInstallDir`** — `filepath.Join(home, "installed")` from
  `os.UserHomeDir()`. Read-only.
- **`ResolveGoroot`** — pin → dest `$InstallDir/<pin>`. Existing dest directory
  is returned and install is skipped. Missing dest with `Download: false` is an
  error. Missing dest with `Download: true` writes optional `Prompt` to Stderr
  then calls `Install` (test hook) or `downloadgo.Download`.
- **`Exec`** — child env `GOROOT=$abs` and `PATH=$abs/bin:$existingPATH` from
  `os.Getenv` (no process `Setenv`). ExtraEnv is appended. Bare `"go"` becomes
  `$GOROOT/bin/go` when that file exists. Empty args run `env`.
- **`Run`** — `ResolveGoroot` then `Exec`.
- **Install hook / writers / InstallDir** — test seams. Leaves always pass
  `t.TempDir()` as InstallDir and a recording hook; never real network.

**Behaviors**

- Pin table is the kool map (go1.14…go1.25). `go1.19` / `1.19` → `go1.19.13`.
- Dest exists as a directory → return it; do not call Install; do not write Prompt.
- Missing + `Download: false` → error; Install unused; Prompt unused.
- Missing + `Download: true` → write Prompt if set, then Install(ctx, pin, installDir).
- Exec child sees absolute GOROOT and a PATH that starts with `$GOROOT/bin`.
- No process-global `Setenv` / `Chdir` / stdio rewrite.

## Decision Tree

Root splits on **operation** (which public function is called). Under each
operation, the next split is the input class that changes the outcome.

```
gotool/withgo/tests/                    [op]
├── pin/                                [input class]
│   ├── known-table/                    full kool map + naked 1.19
│   ├── already-patched/                go1.19.13 / 1.19.13 identity
│   └── unknown/                        go1.99 unchanged
├── goline/                             [go.mod content]
│   ├── major-minor/                    go 1.19 → go1.19
│   ├── with-patch/                     go 1.19.13 → go1.19
│   ├── missing-go-line/                go.mod without go directive → error
│   └── missing-go-mod/                 no go.mod → error
├── resolve/                            [dest existence + Download]
│   ├── dest-exists/                    $InstallDir/go1.19.13 dir → that path
│   ├── missing-download-false/         missing + Download false → error
│   └── missing-download-true/          hook(pin=go1.19.13); Prompt on Stderr
├── exec/                               [args shape]
│   ├── bare-go/                        args=["go"] → $GOROOT/bin/go; env
│   └── empty-args/                     args=[] → env
├── run/                                Resolve existing dest + Exec fake go
└── installdir/                         DefaultInstallDir = $HOME/installed
```

### Parameter significance (high → low)

1. **Operation** — pin / goline / resolve / exec / run / installdir.
2. **Input class** — version shape, go.mod presence, dest existence, args shape.
3. **Download / Prompt / ExtraEnv** — only when the parent outcome uses them.

## Test Index

| # | Leaf | Description | Expected |
|---|------|-------------|----------|
| 1 | `pin/known-table` | Full kool pin map plus naked `1.19` | each input → listed pin |
| 2 | `pin/already-patched` | `go1.19.13` and `1.19.13` | `go1.19.13` |
| 3 | `pin/unknown` | `go1.99` not in the map | `go1.99` |
| 4 | `goline/major-minor` | `go 1.19` in go.mod | `go1.19` |
| 5 | `goline/with-patch` | `go 1.19.13` in go.mod | `go1.19` |
| 6 | `goline/missing-go-line` | go.mod with no `go` line | error |
| 7 | `goline/missing-go-mod` | directory has no go.mod | error |
| 8 | `resolve/dest-exists` | dest dir already present | dest path; hook unused; no Prompt |
| 9 | `resolve/missing-download-false` | dest missing, Download false | error; hook unused |
| 10 | `resolve/missing-download-true` | dest missing, Download true | hook(pin, dir); Prompt on Stderr |
| 11 | `exec/bare-go` | fake `$GOROOT/bin/go` | child GOROOT abs; PATH0=$goroot/bin |
| 12 | `exec/empty-args` | no args | runs `env`; GOROOT + PATH prefix |
| 13 | `run` | existing dest + fake go | compose resolve + exec |
| 14 | `installdir` | DefaultInstallDir | `$HOME/installed` |

## How to Run

From the go-pkgs module root:

```sh
doctest vet ./gotool/withgo/tests
doctest test ./gotool/withgo/tests
```

```go
import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
)

type Request struct {
	Op string // pin | goline | resolve | exec | run | installdir

	GoVersion string
	PinInputs []string // pin: when set, PinPatch each key into Pins

	ModDir string

	InstallDir    string
	Download      bool
	Prompt        string
	RecordInstall bool
	HookGoroot    string

	Goroot   string
	Args     []string
	ExtraEnv []string
	ExecDir  string
}

type Response struct {
	Pin         string
	Pins        map[string]string
	GoLine      string
	Goroot      string
	InstallDir  string
	Stdout      string
	Stderr      string
	HookCalled  bool
	HookVersion string
	HookDir     string
	HookCount   int
	Err         error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}
	var stdout, stderr bytes.Buffer

	switch req.Op {
	case "pin":
		if len(req.PinInputs) > 0 {
			resp.Pins = make(map[string]string, len(req.PinInputs))
			for _, in := range req.PinInputs {
				resp.Pins[in] = withgo.PinPatch(in)
			}
			break
		}
		resp.Pin = withgo.PinPatch(req.GoVersion)

	case "goline":
		line, err := withgo.ModuleGoLine(req.ModDir)
		resp.GoLine = line
		resp.Err = err

	case "resolve":
		goroot, err := withgo.ResolveGoroot(req.GoVersion, resolveOpts(req, resp, &stdout, &stderr))
		resp.Goroot = goroot
		resp.Err = err
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

	case "exec":
		resp.Err = withgo.Exec(req.Goroot, req.Args, withgo.ExecOptions{
			Dir:      req.ExecDir,
			ExtraEnv: req.ExtraEnv,
			Stdout:   &stdout,
			Stderr:   &stderr,
		})
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

	case "run":
		resp.Err = withgo.Run(req.GoVersion, req.Args, resolveOpts(req, resp, &stdout, &stderr), withgo.ExecOptions{
			Dir:      req.ExecDir,
			ExtraEnv: req.ExtraEnv,
			Stdout:   &stdout,
			Stderr:   &stderr,
		})
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()

	case "installdir":
		dir, err := withgo.DefaultInstallDir()
		resp.InstallDir = dir
		resp.Err = err

	default:
		return nil, fmt.Errorf("unknown op: %s", req.Op)
	}

	return resp, nil
}

func resolveOpts(req *Request, resp *Response, stdout, stderr *bytes.Buffer) withgo.ResolveOptions {
	opts := withgo.ResolveOptions{
		InstallDir: req.InstallDir,
		Download:   req.Download,
		Prompt:     req.Prompt,
		Stdout:     stdout,
		Stderr:     stderr,
	}
	if req.RecordInstall {
		opts.Install = func(ctx context.Context, version, installDir string) (string, error) {
			resp.HookCalled = true
			resp.HookCount++
			resp.HookVersion = version
			resp.HookDir = installDir
			if req.HookGoroot != "" {
				return req.HookGoroot, nil
			}
			return filepath.Join(installDir, version), nil
		}
	}
	return opts
}
```
