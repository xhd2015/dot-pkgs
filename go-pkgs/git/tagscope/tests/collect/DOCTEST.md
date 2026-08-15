# tagscope/collect — Parse Tags and Build Scope Inventory

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope` Phase 1.
`ParseTagName` recognizes scoped semver git tags. `CollectFromNames` builds
inventory and per-scope lineage from explicit tag names. `Collect` runs
`git tag -l` in a repo and delegates to `CollectFromNames`.

No HEAD comparison, owned-file diff, tag creation, or wrk CLI in this tree.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies a single tag name, an explicit tag name list, or a git
  repo root for collection.
- **Parser** — `ParseTagName` matches `{optional/path/}v{major}.{minor}.{patch}`
  with optional `-{prerelease}` suffix; non-matching names are rejected.
- **Collector** — `CollectFromNames` parses every name, buckets by `TagScope`,
  sorts scopes lexicographically by `VersionPrefix`, and orders tags within each
  scope newest-first using git `version:refname` semantics.
- **Lineage builder** — per scope derives `Newest`, `LatestRelease`, and
  `HasPrereleaseHead` from the sorted tag list.
- **Git collector** — `Collect` lists tags via `git tag -l` then reuses the
  name-list collector.

### Behaviors

**ParseTagName**

- Root numeric release `v0.0.2` → `PathPrefix=""`, `VersionPrefix="v"`,
  `Version="0.0.2"`, `Prerelease=""`, `IsNumericRelease=true`.
- Prerelease suffix `v0.0.2-alpha` → `Prerelease="alpha"`, `IsNumericRelease=false`.
- Scoped tags `sub/v0.2.3` → `PathPrefix="sub/"`, `VersionPrefix="sub/v"`.
- Deep paths `pkg/api/v1.0.0-dev` parse like shallow subpaths.
- Unrecognized patterns (`release-1.0`, `v0.0`, `v0.0.2.1`, `sub/v0.0.2/extra`)
  return `ok=false`.

**CollectFromNames**

- Parsed tags populate `All`, `Scopes`, and `ByScope`; unparsed names go to
  `Unparsed` only.
- Empty input yields empty inventory (no scopes, no unparsed).
- `Scopes` sorted by `VersionPrefix` ascending.
- Within each scope, `Tags` sorted newest-first; `Newest` is head;
  `LatestRelease` is newest numeric release only; `HasPrereleaseHead` when
  `Newest.Prerelease != ""`.

**Collect**

- Requires a git work tree; lists all tags and returns the same structure as
  `CollectFromNames`.

## Decision Tree

```
collect
├── parse/                         [req.Op = "parse"]
│   ├── root/
│   │   ├── numeric-release/
│   │   └── prerelease/
│   ├── subpath/
│   │   ├── numeric-release/
│   │   ├── prerelease/
│   │   └── deep-path/
│   └── unparsed/
│       ├── no-v-segment/
│       ├── two-components/
│       ├── four-components/
│       └── slash-after-version/
├── collect-names/                 [req.Op = "collect-names"]
│   ├── empty/
│   ├── all-unparsed/
│   ├── mixed/
│   ├── scope-order/
│   ├── lineage/
│   │   ├── newest-prerelease-head/
│   │   ├── newest-is-release/
│   │   ├── only-prerelease/
│   │   └── multi-scope/
│   └── sort/
│       ├── patch-order/
│       ├── major-minor-order/
│       └── same-version-prerelease/
└── collect/                       [req.Op = "collect"]
    └── from-git-repo/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `parse/root/numeric-release` | `v0.0.1` → root scope, numeric release |
| `parse/root/prerelease` | `v0.0.2-alpha` → prerelease, not numeric release |
| `parse/subpath/numeric-release` | `sub/v0.2.3` → scoped numeric release |
| `parse/subpath/prerelease` | `sub/v0.2.3-beta` → scoped prerelease |
| `parse/subpath/deep-path` | `pkg/api/v1.0.0-dev` → deep path + prerelease |
| `parse/unparsed/no-v-segment` | `release-1.0` → ok=false |
| `parse/unparsed/two-components` | `v0.0` → ok=false |
| `parse/unparsed/four-components` | `v0.0.2.1` → ok=false |
| `parse/unparsed/slash-after-version` | `sub/v0.0.2/extra` → ok=false |
| `collect-names/empty` | Empty list → empty inventory |
| `collect-names/all-unparsed` | Only invalid names → `Unparsed` only |
| `collect-names/mixed` | Parsed + unparsed names partitioned correctly |
| `collect-names/scope-order` | Scopes sorted by `VersionPrefix` lexicographically |
| `collect-names/lineage/newest-prerelease-head` | Prerelease head, older numeric `LatestRelease` |
| `collect-names/lineage/newest-is-release` | Newest equals `LatestRelease`, no prerelease head |
| `collect-names/lineage/only-prerelease` | Only prerelease tags → `LatestRelease=nil` |
| `collect-names/lineage/multi-scope` | Root and `sub/` scopes both in `ByScope` |
| `collect-names/sort/patch-order` | `v0.0.10` newer than `v0.0.2` (numeric patch) |
| `collect-names/sort/major-minor-order` | Major/minor boundaries sort correctly |
| `collect-names/sort/same-version-prerelease` | Same version: release beats prereleases |
| `collect/from-git-repo` | `Collect` matches tags created in temp git repo |

## How to Run

```sh
doctest vet ./git/tagscope/tests/collect
doctest test -v ./git/tagscope/tests/collect
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

type Request struct {
	Op       string   // "parse" | "collect-names" | "collect"
	Name     string   // for parse
	Names    []string // for collect-names
	RepoRoot string   // for collect
}

type Response struct {
	Parsed    tagscope.ParsedTag
	ParseOK   bool
	Collected tagscope.CollectedTags
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Op {
	case "parse":
		parsed, ok := tagscope.ParseTagName(req.Name)
		return &Response{Parsed: parsed, ParseOK: ok}, nil
	case "collect-names":
		collected := tagscope.CollectFromNames(req.Names)
		return &Response{Collected: collected}, nil
	case "collect":
		collected, err := tagscope.Collect(req.RepoRoot)
		if err != nil {
			return nil, err
		}
		return &Response{Collected: collected}, nil
	default:
		t.Fatalf("unknown op %q", req.Op)
		return nil, nil
	}
}
```