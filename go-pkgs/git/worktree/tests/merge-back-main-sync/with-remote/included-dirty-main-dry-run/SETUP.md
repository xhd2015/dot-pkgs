# Scenario

**Feature**: already-included + dirty main dry-run lists remove only (no fetch/rebase)

```
origin present + feature merged into main + dirty main
  -> MergeBack DryRun+Remove
  -> plan: worktree remove + branch -D; no fetch/rebase; zero mutations
```

```go
import (
	"bytes"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	landFeatureThenDirtyMain(t, req)
	req.DryRun = true
	req.Stdout = &bytes.Buffer{}
	return nil
}
```
