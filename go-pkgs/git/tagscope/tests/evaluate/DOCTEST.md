# tagscope/evaluate — Decide Next Tag from Owned-File Changes

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope` Phase 2.
Given collected tag inventory and owned-file trees at release vs HEAD, decide
per scope whether to skip or plan `NextTag`. Pure evaluation — no tag creation,
no wrk CLI.

Depends on Phase 1 types: `CollectedTags`, `TagScope`, `TagScopeKey`,
`ScopeLineage`, `ParsedTag`, `CollectFromNames`.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies collected tag names (or pre-built inventory), HEAD commit
  hash, per-scope release commit hashes, and injected `OwnedTreePair` maps.
- **Scope tree builder** — `BuildScopeTree` derives parent/child scope relationships
  from `Collected.Scopes` by path-prefix nesting.
- **Path owner** — `OwnedPathsForScope` returns path prefixes owned by a scope,
  excluding paths under nested child scopes.
- **Tree differ** — `DiffOwnedTrees` detects add/remove/rename (blob identity
  change) between two `OwnedTree` snapshots.
- **Tag incrementer** — `IncrementTag` bumps the trailing numeric segment of a
  release tag name (same semantics as kool `git_tag_next`).
- **Evaluator** — `Evaluate` walks scopes in `Collected.Scopes` order, applies
  gate rules, diffs owned trees, and emits `ChangePlan` with per-scope decisions.

### Behaviors

**Evaluate gates (in order, first match wins)**

- No `LatestRelease` in lineage → `SkipReason=no-baseline`.
- `HasPrereleaseHead` → `SkipReason=prerelease-head`.
- `LatestRelease` commit equals `HeadCommit` → `SkipReason=same-commit`.
- `DiffOwnedTrees(AtRelease, AtHead)` false → `SkipReason=no-changes`.
- Else → `NextTag = IncrementTag(LatestRelease.FullName)`, `SkipReason=""`.

**File ownership**

- Root scope `""`: all repo paths minus any path under a child scope prefix.
- Scope `sub/`: paths with prefix `sub/` minus paths under nested children
  (e.g. `sub/nested/`).

**DiffOwnedTrees**

- Returns true on any path added, removed, or same path with different blob id.

**IncrementTag**

- `v0.0.9` → `v0.0.10`; `sub/v0.2.9` → `sub/v0.2.10`.

## Decision Tree

```
evaluate
├── evaluate/                      [req.Op = "evaluate"]
│   ├── gates/                     [skip — gate triggered]
│   │   ├── no-baseline/
│   │   ├── prerelease-head/
│   │   ├── same-commit/
│   │   └── no-changes/
│   ├── bump/                      [plan NextTag]
│   │   ├── root-changed/
│   │   ├── sub-scope-changed/
│   │   └── multi-scope/
│   └── exclude/                   [nested ownership]
│       ├── child-only-change/
│       └── nested-scope/
├── diff/                          [req.Op = "diff"]
│   ├── identical/
│   ├── added-file/
│   ├── removed-file/
│   └── renamed-blob/
├── increment/                     [req.Op = "increment"]
│   ├── patch-bump/
│   └── subpath-bump/
└── scope-tree/                    [req.Op = "build-scope-tree" | "owned-paths"]
    ├── build/
    │   └── multi-nested/
    └── owned-paths/
        ├── root-excludes-child/
        └── nested-excludes-deeper/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `evaluate/gates/no-baseline` | Prerelease-only scope → `SkipReason=no-baseline` |
| `evaluate/gates/prerelease-head` | Newest prerelease head → `SkipReason=prerelease-head` |
| `evaluate/gates/same-commit` | Release commit equals HEAD → `SkipReason=same-commit` |
| `evaluate/gates/no-changes` | Identical owned trees → `SkipReason=no-changes` |
| `evaluate/bump/root-changed` | Root README differs → `NextTag=v0.0.3` |
| `evaluate/bump/sub-scope-changed` | `sub/` file differs → `NextTag=sub/v0.2.4` |
| `evaluate/bump/multi-scope` | Root and `sub/` both bump independently |
| `evaluate/exclude/child-only-change` | Nested change only → parent `sub/` skips, nested bumps |
| `evaluate/exclude/nested-scope` | `sub/` and `sub/nested/` scopes both evaluated in order |
| `diff/identical` | Same trees → `Changed=false` |
| `diff/added-file` | New path in AtHead → `Changed=true` |
| `diff/removed-file` | Path missing in AtHead → `Changed=true` |
| `diff/renamed-blob` | Same path, different oid → `Changed=true` |
| `increment/patch-bump` | `v0.0.9` → `v0.0.10` |
| `increment/subpath-bump` | `sub/v0.2.9` → `sub/v0.2.10` |
| `scope-tree/build/multi-nested` | Children map for root, `sub/`, `sub/nested/` |
| `scope-tree/owned-paths/root-excludes-child` | Root owns only non-child paths |
| `scope-tree/owned-paths/nested-excludes-deeper` | `sub/` excludes `sub/nested/` paths |

## How to Run

```sh
doctest vet ./git/tagscope/tests/evaluate
doctest test -v ./git/tagscope/tests/evaluate
```

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

type Request struct {
	Op              string // "evaluate" | "diff" | "increment" | "build-scope-tree" | "owned-paths"
	Names           []string
	HeadCommit      string
	ReleaseCommits  map[tagscope.TagScopeKey]string // commit at LatestRelease per scope
	OwnedTrees      map[tagscope.TagScopeKey]tagscope.OwnedTreePair
	OldTree         tagscope.OwnedTree
	NewTree         tagscope.OwnedTree
	Tag             string
	ScopePrefix     string   // for owned-paths
	AllPaths        []string // for owned-paths
}

type Response struct {
	Plan       tagscope.ChangePlan
	Changed    bool
	NextTag    string
	ScopeTree  tagscope.ScopeTree
	OwnedPaths []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Op {
	case "evaluate":
		collected := tagscope.CollectFromNames(req.Names)
		plan := tagscope.Evaluate(tagscope.ChangeCheckInput{
			Collected:      collected,
			HeadCommit:     req.HeadCommit,
			ReleaseCommits: req.ReleaseCommits,
			OwnedTrees:     req.OwnedTrees,
		})
		return &Response{Plan: plan}, nil
	case "diff":
		changed := tagscope.DiffOwnedTrees(req.OldTree, req.NewTree)
		return &Response{Changed: changed}, nil
	case "increment":
		next, err := tagscope.IncrementTag(req.Tag)
		if err != nil {
			return nil, err
		}
		return &Response{NextTag: next}, nil
	case "build-scope-tree":
		collected := tagscope.CollectFromNames(req.Names)
		tree := tagscope.BuildScopeTree(collected)
		return &Response{ScopeTree: tree}, nil
	case "owned-paths":
		collected := tagscope.CollectFromNames(req.Names)
		tree := tagscope.BuildScopeTree(collected)
		scope := scopeForPrefix(t, collected, req.ScopePrefix)
		paths := tagscope.OwnedPathsForScope(scope, tree, req.AllPaths)
		return &Response{OwnedPaths: paths}, nil
	default:
		t.Fatalf("unknown op %q", req.Op)
		return nil, nil
	}
}
```