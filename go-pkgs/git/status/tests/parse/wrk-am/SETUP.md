# Scenario

**Feature**: staged-new then worktree-edit (`AM`) counts once as added

```
AM path -> ParsePorcelainWrk -> Staged=1, Changed=0
```

## Steps

1. Set `req.Op` to `"parse-wrk"`.
2. Set porcelain to a single `AM` line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "parse-wrk"
	req.Porcelain = "AM staged-then-edited.txt"
	return nil
}
```
