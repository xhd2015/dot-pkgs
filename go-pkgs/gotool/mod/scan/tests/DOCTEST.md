# gotool/mod/scan — module tree scan with skip rules + streaming

## Version

0.0.2

Library doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan`. Each leaf
calls the scan package APIs directly via root `Run()` — no CLI. The `scan` package is
not yet implemented; these doctests are intentionally RED until the implementer adds it.

## DSN (Domain Specific Notion)

The **caller** hands `scan` a root directory (a Go workspace, possibly a git repo).
The **walker** does a `filepath.WalkDir` of root and, for every directory encountered:

- applies **name skips** — dirs literally named `.git`, `vendor`, or `testdata` are pruned;
- if the root is inside a git repo, applies **git skips**:
  - a dir that git (root's index/ignore rules) reports as ignored is pruned;
  - a dir that carries its own `.git` but is NOT a tracked submodule/gitlink of the root
    repo (a **nested separate repo**) is pruned wholesale — its subtree is never scanned;
- for every `go.mod` found in a non-skipped dir, the **reader** parses it into a `Module`
  (`Dir` relative to root with `.` for root, slash-joined, no `./`; `Path` = module path;
  `Requires`/`Replaces` from the file).

The caller chooses one of two consumption modes:

- **`Scan(root, opts)`** — collects all modules then sorts by `Dir` (lexical). Used by
  kool's tree command which needs a stable, sorted view.
- **`ScanStream(root, opts, fn)`** — calls `fn(module)` per module in walk order as found,
  no sort, no buffering. If `fn` returns an error the walk stops and that error is
  returned. Used by `kool go modules --list` which streams lines as it walks.

When the root is **not** a git repo, the git-based skips are disabled; only the name-based
skips (`.git`/`vendor`/`testdata`) apply, so an untracked-but-not-ignored directory with a
`go.mod` IS scanned.

## Decision Tree

The most significant factor is **consumption mode** (sorted batch vs. unordered stream),
because it picks which function is called and changes the observable ordering guarantee.
The next factor is **skip class** — the reason a subtree is excluded — which is the heart
of the scan behavior. Under skips, sibling branches are MECE over the skip rule exercised
(name-based / gitignore / nested-separate-repo / keeps-untracked-no-git).

```
scan tests
├── basic/
│   └── nested-modules/               # Scan returns all modules sorted by Dir
├── skips/
│   ├── testdata-dir/                 # name skip: testdata/ pruned
│   ├── vendor-dir/                   # name skip: vendor/ pruned
│   ├── gitignored-dir/               # git skip: .gitignore'd dir pruned
│   ├── nested-separate-repo/         # git skip: own .git, not a submodule → pruned
│   └── keeps-untracked-no-git/       # no git → untracked dir with go.mod IS scanned
└── stream/
    └── walk-order/                   # ScanStream emits walk order (unsorted) vs Scan sorts
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `basic/nested-modules` | `Scan` returns root + nested modules, sorted by Dir |
| 2 | `skips/testdata-dir` | `testdata/` subtree skipped, not in results |
| 3 | `skips/vendor-dir` | `vendor/` subtree skipped, not in results |
| 4 | `skips/gitignored-dir` | `.gitignore`'d dir skipped, not in results |
| 5 | `skips/nested-separate-repo` | dir with own `.git`, untracked by root → skipped |
| 6 | `skips/keeps-untracked-no-git` | no git at root → untracked dir with go.mod IS scanned |
| 7 | `stream/walk-order` | `ScanStream` emits walk order; `Scan` sorts — contrast proves stream unsorted |

## How to Run

```sh
doctest vet ./go-pkgs/gotool/mod/scan/tests
doctest test ./go-pkgs/gotool/mod/scan/tests
```

```go
import (
	"fmt"
	"sort"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

type Request struct {
	Operation string // "scan" -> scan.Scan, "stream" -> scan.ScanStream
	RootDir   string // populated by leaf Setup: workspace root passed to scan
}

type Response struct {
	Err      error
	Modules  []scan.Module // Scan result, sorted by Dir
	Streamed []scan.Module // ScanStream result, in walk (emission) order
}

// Run dispatches on req.Operation. For "scan" it calls scan.Scan and returns the
// sorted slice. For "stream" it calls scan.ScanStream, appending each module to
// resp.Streamed in emission order, then ALSO calls scan.Scan into resp.Modules so
// the leaf can contrast sorted vs. unsorted. Both call into the not-yet-implemented
// scan package, so leaves stay RED until the implementer lands scan.go.
func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}

	switch req.Operation {
	case "scan":
		modules, err := scan.Scan(req.RootDir, scan.Options{})
		resp.Err = err
		resp.Modules = modules
		if err != nil {
			return resp, nil
		}

	case "stream":
		err := scan.ScanStream(req.RootDir, scan.Options{}, func(m scan.Module) error {
			resp.Streamed = append(resp.Streamed, m)
			return nil
		})
		resp.Err = err
		if err != nil {
			return resp, nil
		}
		// Also run Scan so the leaf can contrast sorted vs. walk-order.
		modules, err := scan.Scan(req.RootDir, scan.Options{})
		if err != nil {
			return resp, fmt.Errorf("scan.Scan for contrast failed: %w", err)
		}
		resp.Modules = modules

	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	return resp, nil
}

// sortedDirs returns the Dir fields of resp.Modules (already sorted by Scan) and is
// available to leaves that want a defensive copy. Scan is contractually sorted by Dir.
var _ = sort.Strings
```
