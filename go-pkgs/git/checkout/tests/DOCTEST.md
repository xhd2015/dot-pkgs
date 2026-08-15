# git/checkout — Checkout Enrichment

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/checkout`. `Enrich` runs
durable stepwise git enrichment (branch → sha → msg → status) and returns
`Meta` with partial fields plus `Error` on failure. Never returns error to caller.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies repo path and optional `ShortSHALength` (default 7).
- **Enricher** — runs git subprocesses via `git/cmd` in order; stops early on
  unborn HEAD or records partial `Meta` with `Error`.
- **Git** — external binary on PATH; tests skip when unavailable.

### Behaviors

- **Unborn HEAD** — `git init` only → `Error: "no commits (HEAD unborn)"`.
- **Clean repo** — after commit → `Branch`, 7-char `CommitSHA`, `CommitMsg`, `Status: "clean"`.
- **Dirty repo** — modified file → `Status: "dirty (N modified)"` with other fields populated.
- **Wrk style** — `StatusStyle: FormatWrk` uses `ParsePorcelainWrk` + `FormatWrk`
  for `Meta.Status`. Untracked files are included by default (`??` → added);
  `PorcelainUntracked: false` is an optional opt-out that appends
  `--untracked-files=no`.
- Enrich always returns `Meta`; caller inspects `Error` field.

## Decision Tree

```
checkout
└── enrich/
    ├── clean-repo/
    ├── empty-repo/
    ├── dirty-status/
    └── wrk-style/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `enrich/clean-repo` | Committed repo → branch, sha, msg, clean |
| `enrich/empty-repo` | No commits → `no commits (HEAD unborn)` |
| `enrich/dirty-status` | Modified file → dirty status with full meta |
| `enrich/wrk-style` | Clean repo + `FormatWrk` → `Status: clean` |

## How to Run

```sh
doctest vet ./go-pkgs/git/checkout/tests/
doctest test -v ./go-pkgs/git/checkout/tests/
```

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/checkout"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

type Request struct {
	RepoPath             string
	ShortSHALength       int
	StatusStyle          status.FormatStyle
	PorcelainUntracked   *bool
}

type Response struct {
	Meta checkout.Meta
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	opts := checkout.Options{}
	if req.ShortSHALength > 0 {
		opts.ShortSHALength = req.ShortSHALength
	}
	if req.StatusStyle != 0 {
		opts.StatusStyle = req.StatusStyle
	}
	if req.PorcelainUntracked != nil {
		opts.PorcelainUntracked = *req.PorcelainUntracked
	}
	meta := checkout.Enrich(context.Background(), req.RepoPath, opts)
	return &Response{Meta: meta}, nil
}
```