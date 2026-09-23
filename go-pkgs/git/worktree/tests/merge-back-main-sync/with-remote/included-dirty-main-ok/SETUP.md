# Scenario

**Feature**: already-included worktree + dirty main → MergeBack --rm succeeds

```
origin present + feature merged into main + dirty main
  -> MergeBack Remove=true
  -> Action=removed; no origin rebase
```

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	landFeatureThenDirtyMain(t, req)
	return nil
}
```
