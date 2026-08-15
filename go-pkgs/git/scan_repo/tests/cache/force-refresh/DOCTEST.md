# scan_repo — Options.Refresh force cold (P5)

## Version
0.0.2

Nested doc tests for `Options.Refresh`. When true, `Scan` skips the warm path and
performs a cold full walk (and index rewrite when cache is enabled), finding
brand-new repos that warm soft incompleteness would omit.

This tree is nested (own `DOCTEST.md`) so force-refresh Run stays isolated; the
parent library tree also passes `Refresh` for P7 orphan-gc cold-rescan leaves.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies roots, explicit temp `CacheRoot`, and `Refresh=true`.
- **Cold seeder** — prior `Scan` with `NoCache=false` populates a complete root
  index so the workspace is warm-eligible.
- **Scan** — with `Refresh=true` must full-walk (not warm-serve) even when root
  cache is complete; returns brand-new repos planted after the seed.
- **Contrast** — same fixture without Refresh is covered by parent
  `cache/warm/serves-cached-omits-new` (omits brand-new).

### Behaviors

- `Refresh=true` + warm-eligible cache → cold full walk; Result includes known
  and brand-new mains, path-sorted.
- `NoCache` remains false (cache still enabled for rewrite); this leaf asserts
  discovery only, not index rewrite details.

## Decision Tree

```
force-refresh
└── finds-new/    # after warm seed + plant brand-new, Refresh finds both repos
```

## Test Index

| Leaf | Description |
|------|-------------|
| `finds-new` | `Refresh=true` force cold finds brand-new that warm would miss |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cache/force-refresh/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/force-refresh/
```

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Roots     []string
	CacheRoot string
	NoCache   bool
	Refresh   bool
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	result, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     req.Roots,
		CacheRoot: req.CacheRoot,
		NoCache:   req.NoCache,
		Refresh:   req.Refresh,
	})
	if err != nil {
		return nil, err
	}
	return &Response{Repos: result.Repos, RootErrors: result.RootErrors}, nil
}
```
