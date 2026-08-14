# Scenario

**Feature**: descendants are not rewritten; exact-tip refs still move; warn

```
branch tip-here at old; commit child on master -> rewrite old -> child still parents old
```

## Steps

1. Point `tip-here` at the target SHA.
2. Add a child commit on `master` (so `master` is no longer an exact tip).
3. Rewrite the old SHA with `-m` `rewrite ancestor`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	runGit(t, req.Dir, "branch", "tip-here", req.OldSHA)
	req.ChildSHA = commitFile(t, req.Dir, "child.txt", "child\n", "child commit")
	req.Args = append(req.Args, "-m", "rewrite ancestor")
	return nil
}
```
