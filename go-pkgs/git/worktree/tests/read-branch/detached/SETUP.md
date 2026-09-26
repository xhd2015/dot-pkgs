# Scenario

**Feature**: `ReadBranch` returns `"HEAD"` when detached

```
commit then git checkout --detach -> ReadBranch -> "HEAD"
```

## Steps

1. Init repo on `main` with one commit.
2. `git checkout --detach HEAD`.
3. Set `req.Dir`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	dir := initUnbornRepo(t)
	firstCommit(t, dir)
	runGit(t, dir, "checkout", "--detach", "HEAD")
	req.Dir = dir
	return nil
}
```
