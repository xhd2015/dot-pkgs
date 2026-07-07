# file/analyse — Home Directory Scan and Format

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse`. `Scan` walks every
immediate child of a fake HOME directory, deep-aggregates sizes, enriches tool dirs
with semantic lines, and optionally streams each completed entry via `OnEntry`.
`FormatEntryBlock` and `FormatSummaryLines` render human-readable blocks unchanged
from ai-critic `machineanalyse`.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies `Options.Home` (temp fake HOME) and optional `OnEntry` callback.
- **Scan** — reads sorted top-level HOME children; builds one `EntryResult` per child.
- **Dir walker** — `deepSize` and `immediateChildSizes` for directory entries; children
  sorted alphabetically by name.
- **File scanner** — top-level files report `size` and `lines` (or `(binary)`).
- **Semantic enricher** — tool dirs (`.codex`, `.grok`, `.cursor`, …) get `SemanticLine`
  rollups (sessions rollouts, skills, plugins, etc.).
- **Git aggregator** — delegates to `git/scan_repo.Scan` per entry; counts repos and
  linked worktrees.
- **Node modules counter** — recursive `node_modules` directory count under each entry.
- **OnEntry hook** — invoked once per fully built entry in sorted order; non-nil error
  aborts scan and propagates from `Scan`.
- **Formatter** — `FormatEntryBlock` orders children → semantic → aggregates;
  `FormatSummaryLines` emits tool-specific summary lines only when indicator dirs exist.

### Behaviors

**Scan**

- Resolve and validate `Home` as an existing directory.
- Walk top-level children alphabetically by entry name.
- Directory entries: `EntryKindDir`, deep total bytes, immediate children with deep sizes.
- File entries: `EntryKindFile`, byte size, text line count or `(binary)`.
- `.codex` semantic: rollout session files, top-level skill dirs, plugins count (0 when absent).
- After each entry: call `OnEntry(result)` when set; abort on callback error.
- Return full `[]EntryResult` and `ScanSummary` even when `OnEntry` is nil.

**Format**

- Entry block: `> name`, child lines, blank line before semantic when both present,
  semantic lines, then optional `git-dirs`, `worktrees`, `node_modules` aggregates.
- Summary: `analyse-files summary` header; codex lines when `HasCodex`; grok lines only
  when `HasGrok`.

## Decision Tree

```
analyse
├── scan/                      [Mode=scan — direct analyse.Scan]
│   ├── basic-dir/             plain dir children sorted; deep sizes; EntryKindDir
│   ├── file-lines/            text lines N; binary lines (binary)
│   ├── codex-semantic/        rollout sessions; skills; plugins 0
│   ├── git-dirs/              Aggregates.GitRepos == 1 with git init fixture
│   ├── node-modules/          child node_modules + NodeModulesDirs recursive count
│   ├── entry-order/           results alphabetical by entry name
│   └── on-entry/              OnEntry order + abort-on-error
└── format/                    [Mode=format-* — pure formatters]
    ├── entry-block/           children before semantic before aggregates
    └── summary-topic/         codex when HasCodex; grok omitted when absent
```

## Test Index

| Leaf | Mode | Description |
|------|------|-------------|
| `scan/basic-dir` | Scan | Children sorted; deep sizes; dir entry kind |
| `scan/file-lines` | Scan | Text lines count; binary `lines (binary)` |
| `scan/codex-semantic` | Scan | Rollout sessions, top-level skill dirs, plugins 0 |
| `scan/git-dirs` | Scan | `Aggregates.GitRepos == 1` with git init fixture |
| `scan/node-modules` | Scan | Child `node_modules` + `NodeModulesDirs` recursive count |
| `scan/entry-order` | Scan | Results alphabetical by entry name |
| `scan/on-entry` | Scan | OnEntry once per entry in sorted order; abort on error |
| `format/entry-block` | Format | Children before semantic before aggregates |
| `format/summary-topic` | Format | Codex summary when HasCodex; grok omitted when absent |

## How to Run

```sh
cd dot-pkgs-with-critic/go-pkgs
doctest vet ./file/analyse/tests
doctest test -v ./file/analyse/tests/...
go test ./file/analyse/... -count=1
```

```go
import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

type Request struct {
	Home string

	// SeedProfile selects the fake HOME fixture (scan leaves).
	SeedProfile string

	// Mode selects Run behavior: "" or "scan", "format-entry", "format-summary".
	Mode string

	// On-entry test: collect callback names; abort after named entry.
	CollectOnEntry  bool
	AbortAfterEntry string

	// Format leaves populate these directly.
	Entry   analyse.EntryResult
	Summary analyse.ScanSummary
}

type Response struct {
	Entries      []analyse.EntryResult
	Summary      analyse.ScanSummary
	OnEntryOrder []string
	EntryBlock   string
	SummaryText  string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "format-entry":
		return &Response{
			EntryBlock: analyse.FormatEntryBlock(req.Entry),
		}, nil
	case "format-summary":
		return &Response{
			SummaryText: strings.Join(analyse.FormatSummaryLines(req.Summary), "\n"),
		}, nil
	default:
		var onEntryOrder []string
		opts := analyse.Options{Home: req.Home}
		if req.CollectOnEntry || req.AbortAfterEntry != "" {
			opts.OnEntry = func(entry analyse.EntryResult) error {
				onEntryOrder = append(onEntryOrder, entry.Name)
				if req.AbortAfterEntry != "" && entry.Name == req.AbortAfterEntry {
					return fmt.Errorf("abort after %s", entry.Name)
				}
				return nil
			}
		}
		entries, summary, err := analyse.Scan(context.Background(), opts)
		if err != nil {
			return &Response{OnEntryOrder: onEntryOrder}, err
		}
		return &Response{
			Entries:      entries,
			Summary:      summary,
			OnEntryOrder: onEntryOrder,
		}, nil
	}
}
```