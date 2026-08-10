# go-pkgs worktree surface after gitops shim (P2)

Doc-style tests proving that **go-pkgs** public package paths still expose
list / WorktreesOnBranch / clean primitives after low-level implementations
delegate to (or re-export) gitops.

Import path under test (callers keep using go-pkgs, not gitops directly):

`github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree`

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — wrk / mvd / MergeBack (or this harness) importing **go-pkgs**
  `git/worktree` only.
- **Main checkout** — primary worktree (`.git` directory); always in `List`.
- **Linked worktree** — `git worktree add` checkout (`.git` file); in `List` and
  `ListLinked`.
- **Entry** — porcelain row: `Path`, `Branch` (empty when detached), `HEAD`,
  `IsMain`.
- **List / ListLinked** — inventory of registered worktrees via go-pkgs.
- **WorktreesOnBranch** — filter entries by branch name (re-export if missing);
  multi-checkout is data only, no refuse policy here.
- **IsClean** — go-pkgs: `error` when `git status --porcelain` is non-empty
  (untracked counts as dirty).
- **IsCleanWrk** — go-pkgs: wrk four-bucket clean (`??` → added → dirty).

**Behaviors**

- After P2 shim, go-pkgs package path still compiles and behaves for List,
  WorktreesOnBranch, and porcelain-aware clean.
- Two registered worktrees on the same branch → `WorktreesOnBranch` len=2,
  no policy error.
- Untracked-only dirt: `IsClean` returns error; `IsCleanWrk` is false.
- Clean repo: `IsClean` nil; `IsCleanWrk` true.
- Paths compared after clean / symlink resolution (macOS `/var` vs `/private/var`).

## Version

0.0.2

## Decision Tree

```
gitops-shim
├── list/
│   └── main-plus-linked/              (LEAF) List+ListLinked via go-pkgs
├── worktrees_on_branch/
│   └── two-linked-same-branch/        (LEAF) two linked on feature → len=2
└── clean/
    ├── clean-repo/                    (LEAF) IsClean nil; IsCleanWrk true
    └── untracked-dirty/               (LEAF) IsClean err; IsCleanWrk false
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/main-plus-linked` | Through go-pkgs `List`/`ListLinked`: main+linked inventory |
| 2 | `worktrees_on_branch/two-linked-same-branch` | Through go-pkgs `WorktreesOnBranch`: two linked same branch → len=2 |
| 3 | `clean/clean-repo` | Through go-pkgs clean: empty porcelain → IsClean nil, IsCleanWrk true |
| 4 | `clean/untracked-dirty` | Through go-pkgs clean: untracked → IsClean error, IsCleanWrk false |

## How to Run

```sh
cd external/dot-pkgs-master-2026-07-30/go-pkgs
doctest vet ./git/worktree/tests/gitops-shim
doctest test ./git/worktree/tests/gitops-shim/...
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// Op selects which go-pkgs worktree API Run exercises.
//   "list_and_linked" | "worktrees_on_branch" | "clean"
type Request struct {
	Op          string
	Dir         string // main repo (list / WorktreesOnBranch / clean)
	Branch      string // WorktreesOnBranch filter
	MainPath    string
	LinkedPath  string
	LinkedPath2 string
}

type Response struct {
	Entries []worktree.Entry
	Linked  []worktree.Entry

	// Clean surface: IsClean returns error on dirty porcelain (not a Run failure).
	IsCleanNil bool
	IsCleanErr string // empty when IsCleanNil
	IsCleanWrk bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Op {
	case "list_and_linked":
		entries, err := worktree.List(req.Dir)
		if err != nil {
			return nil, err
		}
		linked, err := worktree.ListLinked(req.Dir)
		if err != nil {
			return nil, err
		}
		return &Response{Entries: entries, Linked: linked}, nil
	case "worktrees_on_branch":
		entries, err := worktree.WorktreesOnBranch(req.Dir, req.Branch)
		if err != nil {
			return nil, err
		}
		return &Response{Entries: entries}, nil
	case "clean":
		cleanErr := worktree.IsClean(req.Dir)
		wrkOK, wrkErr := worktree.IsCleanWrk(req.Dir)
		if wrkErr != nil {
			return nil, wrkErr
		}
		resp := &Response{IsCleanNil: cleanErr == nil, IsCleanWrk: wrkOK}
		if cleanErr != nil {
			resp.IsCleanErr = cleanErr.Error()
		}
		return resp, nil
	default:
		t.Fatalf("unknown op %q", req.Op)
		return nil, nil
	}
}
```
