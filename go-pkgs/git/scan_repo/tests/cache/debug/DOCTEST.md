# scan_repo — Options.Debug phase timing / warm-cold mode logs

## Version
0.0.2

Nested doc tests for `Options.Debug` (implemented). When true, `Scan` writes
structured `scan:` lines to `Options.Stderr` (default `os.Stderr`): cache root,
per-root `mode=warm|cold` + reason, warm serve counts/duration, refresh budget
summary, and root total duration. When false, zero `scan:` lines (Verbose skip
warnings remain separate).

This tree is nested (own `DOCTEST.md`) so stderr capture and Debug-focused Run
stay isolated from the parent library tree's Request/Response contract.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies roots, explicit temp `CacheRoot`, `Debug`, and optional
  `Stderr` capture buffer.
- **Scan** — discovery + warm/cold cache path; when `Debug=true`, emits phase
  timing and mode lines to stderr with prefix `scan:`.
- **Debug logger** — formats greppable lines (not Verbose permission/remote
  skip warnings). Default writer is `os.Stderr` when `Options.Stderr` is nil.
- **Cold path** — full walk when root cache is missing/incomplete, `NoCache`,
  or `Refresh`; logs `mode=cold` and a reason token.
- **Warm path** — complete root cache; logs `mode=warm`, serve candidate/live
  counts and duration, refresh budget/units summary, root total.
- **Stderr buffer (tests)** — `Run` always passes a `bytes.Buffer` as
  `Options.Stderr` so Assert can inspect captured text without polluting the
  test process stderr.

### Behaviors

- `Debug=true` + empty / missing root cache → cold walk; stderr contains
  `scan:` and `mode=cold` with a reason such as `missing_root_entry` (or
  `scan_complete_false` / `no_cache` when those apply).
- `Debug=true` + complete root cache (after cold seed) → warm serve; stderr
  contains `scan:`, `mode=warm`, and serve timing markers (candidates / live /
  duration). May also log refresh summary and root total.
- `Debug=false` → zero lines containing the substring `scan:` even when a cold
  or warm Scan runs with cache enabled.
- No per-directory cold-walk spam: debug volume stays phase-level (root /
  serve / refresh / total), not one line per visited directory.
- `Verbose` is orthogonal: skip warnings keep their existing format; they are
  not required to carry the `scan:` prefix.

## Decision Tree

```
debug                          [nested — Options.Debug + stderr capture]
├── on/                        [Debug=true]
│   ├── cold/                  # empty CacheRoot → mode=cold + scan:
│   └── warm/                  # cold seed then Scan → mode=warm + serve timing
└── off/                       [Debug=false — leaf]
                               # Scan with cache; zero scan: lines on stderr
```

## Test Index

| Leaf | Debug | Path | Description |
|------|-------|------|-------------|
| `on/cold` | true | cold (empty cache) | stderr has `scan:` + `mode=cold` (+ reason) |
| `on/warm` | true | warm (after seed) | stderr has `scan:` + `mode=warm` + serve timing |
| `off` | false | cold with cache | stderr has zero `scan:` markers |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cache/debug/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/debug/
```

From monorepo / worktree with nested external:

```sh
doctest vet ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/debug/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/debug/
```

```go
import (
	"bytes"
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Roots     []string
	CacheRoot string
	NoCache   bool
	Debug     bool // Options.Debug — phase-level scan: logs on Options.Stderr
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
	Stderr     string // captured Options.Stderr during Scan
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	var stderrBuf bytes.Buffer
	result, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     req.Roots,
		CacheRoot: req.CacheRoot,
		NoCache:   req.NoCache,
		Debug:     req.Debug,
		Stderr:    &stderrBuf,
	})
	stderr := stderrBuf.String()
	if err != nil {
		return &Response{Stderr: stderr}, err
	}
	return &Response{
		Repos:      result.Repos,
		RootErrors: result.RootErrors,
		Stderr:     stderr,
	}, nil
}
```
