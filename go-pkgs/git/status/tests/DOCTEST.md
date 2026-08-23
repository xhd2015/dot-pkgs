# git/status — Porcelain Parse and Status Formatting

## Version
0.0.3

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/status`. `ParsePorcelain`
aggregates `git status --porcelain` lines into backup `Counts`. `ParsePorcelainWrk`
applies the wrk five-bucket taxonomy (`??` → untracked; index change → staged; path-once).
`Format` renders backup-style strings; `FormatWrk` renders wrk `--status` values
(always five segments when dirty).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies porcelain text or pre-aggregated counts.
- **Backup parser** — `ParsePorcelain` counts modified, added, deleted, untracked,
  renamed, copied, and unmerged entries from porcelain lines.
- **Wrk parser** — `ParsePorcelainWrk` maps porcelain into five buckets: added,
  changed, renamed, deleted, untracked (`??` → **untracked**; index change → **staged**
  once; unstaged `M`/`D`/`R` → changed/deleted/renamed).
- **Backup formatter** — `Format(Counts, FormatBackup)` emits `clean` or
  `dirty (N modified, M added, …)` matching `machinebackup` output.
- **Wrk formatter** — `FormatWrk(WrkCounts)` emits `clean` or
  `dirty (N staged, N changed, N renamed, N deleted, N untracked)` with all five
  segments when dirty.

### Behaviors

- Empty porcelain → all counts zero → both formatters yield `"clean"`.
- Backup: non-zero counts → only non-zero labels appear in `dirty (...)` suffix;
  order: modified, added, deleted, untracked, renamed, copied, unmerged.
- Wrk: dirty output always lists staged, changed, renamed, deleted, untracked
  (zeros included).
- Wrk status includes untracked files by default (`??` → untracked via
  `ParsePorcelainWrk`). Callers may still pass `--untracked-files=no` via
  `checkout.Options.PorcelainUntracked: false`.

## Decision Tree

```
status
├── parse/
│   ├── clean/
│   ├── mixed/
│   ├── wrk-mixed/
│   └── wrk-am/
└── format/
    ├── backup-clean/
    ├── backup-dirty/
    ├── wrk-clean/
    ├── wrk-dirty/
    └── wrk-partial/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `parse/clean` | Empty porcelain → zero backup counts |
| `parse/mixed` | Two modified + one untracked |
| `parse/wrk-mixed` | Porcelain maps to wrk buckets (`??`→untracked, `M`→changed, `R`, `D`) |
| `parse/wrk-am` | `AM` path-once → Staged=1, Changed=0 |
| `format/backup-clean` | Zero backup counts → `"clean"` |
| `format/backup-dirty` | Backup counts → `"dirty (2 modified, 1 untracked)"` |
| `format/wrk-clean` | Zero WrkCounts → `"clean"` |
| `format/wrk-dirty` | `{1,1,1,1,1}` → full five-segment dirty line |
| `format/wrk-partial` | `{Changed:1}` → dirty with zero segments for other buckets |

## How to Run

```sh
doctest vet ./go-pkgs/git/status/tests/
doctest test -v ./go-pkgs/git/status/tests/
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

type Request struct {
	Op        string // "parse" | "format" | "parse-wrk" | "format-wrk"
	Porcelain string
	Counts    status.Counts
	WrkCounts status.WrkCounts
	Style     status.FormatStyle
}

type Response struct {
	Counts    status.Counts
	WrkCounts status.WrkCounts
	Formatted string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Op {
	case "parse":
		counts := status.ParsePorcelain(req.Porcelain)
		return &Response{Counts: counts}, nil
	case "parse-wrk":
		counts := status.ParsePorcelainWrk(req.Porcelain)
		return &Response{WrkCounts: counts}, nil
	case "format":
		formatted := status.Format(req.Counts, req.Style)
		return &Response{Formatted: formatted}, nil
	case "format-wrk":
		formatted := status.FormatWrk(req.WrkCounts)
		return &Response{Formatted: formatted}, nil
	default:
		t.Fatalf("unknown op %q", req.Op)
		return nil, nil
	}
}
```
