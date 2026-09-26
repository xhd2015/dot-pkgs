# Scenario

**Feature**: `ReadBranch` still returns the branch after the first commit

```
git init -b main + commit -> ReadBranch -> "main"
```

## Steps

1. Init unborn repo on `main`.
2. Create the first commit.
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
	req.Dir = dir
	return nil
}
```
