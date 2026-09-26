# Scenario

**Feature**: `ReadBranch` names the branch on unborn HEAD

```
git init -b main (no commits) -> ReadBranch -> "main"
```

## Preconditions

- Isolated repo after `git init` only.

## Steps

1. Init unborn repo on `main`.
2. Set `req.Dir`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Dir = initUnbornRepo(t)
	return nil
}
```
