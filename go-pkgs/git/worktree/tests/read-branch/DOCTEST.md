# worktree.ReadBranch — current branch including unborn HEAD

## Version
0.0.1

`ReadBranch` returns the checkout's current branch name. Detached HEAD is
`"HEAD"`. Unborn HEAD (`git init`, no commits) still has a symbolic ref, so
the branch name is returned — `git rev-parse --abbrev-ref HEAD` is not used
because it fatals on unborn HEAD.

## DSN (Domain Specific Notion)

- **ReadBranch** — `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree`.
- **Unborn HEAD** — repo after `git init` only; `HEAD` is a symbolic ref with
  no commit. Branch name (e.g. `main`) is still readable via `symbolic-ref`.
- **Detached** — `git checkout --detach`; ReadBranch returns `"HEAD"`.

## Decision Tree

```
read-branch
├── unborn-init/     (LEAF) git init -b main, no commits → "main"
├── after-commit/    (LEAF) same repo after first commit → "main"
└── detached/        (LEAF) checkout --detach → "HEAD"
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `unborn-init` | Unborn HEAD → `"main"` (not error) |
| 2 | `after-commit` | Attached after first commit → `"main"` |
| 3 | `detached` | Detached HEAD → `"HEAD"` |

## How to Run

```sh
cd go-pkgs
doctest vet ./git/worktree/tests/read-branch
doctest test -v ./git/worktree/tests/read-branch
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

type Request struct {
	Dir string
}

type Response struct {
	Branch string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = t
	_ = d
	branch, err := worktree.ReadBranch(req.Dir)
	if err != nil {
		return nil, err
	}
	return &Response{Branch: branch}, nil
}
```
