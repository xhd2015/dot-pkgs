# Scenario

**Feature**: every exact-tip branch moves; unrelated branch stays

```
branch wrk-a, wrk-b at old; other at parent -> rewrite -> wrk-a/wrk-b/master move, other stays
```

## Steps

1. Create `wrk-a` and `wrk-b` at the target SHA and `other` at the parent.
2. Append `-m` `move tips`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	runGit(t, req.Dir, "branch", "wrk-a", req.OldSHA)
	runGit(t, req.Dir, "branch", "wrk-b", req.OldSHA)
	runGit(t, req.Dir, "branch", "other", req.ParentSHA)
	req.UnrelatedBranch = "other"
	req.UnrelatedSHA = req.ParentSHA
	req.Args = append(req.Args, "-m", "move tips")
	return nil
}
```
