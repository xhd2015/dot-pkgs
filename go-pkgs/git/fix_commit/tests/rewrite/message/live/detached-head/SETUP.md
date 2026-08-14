# Scenario

**Feature**: detached HEAD on the old SHA moves; parked `master` stays

```
master parked at parent; checkout --detach old -> rewrite -> HEAD=new, master=parent
```

## Steps

1. Force `master` to the parent and detach HEAD at the target SHA.
2. Append `-m` `detached rewrite`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	runGit(t, req.Dir, "checkout", "--detach", req.OldSHA)
	runGit(t, req.Dir, "update-ref", "refs/heads/master", req.ParentSHA)
	req.Args = append(req.Args, "-m", "detached rewrite")
	return nil
}
```
