# Scenario

**Feature**: no `origin` remote → local retag; skip remote ops with a warning

```
local tag only, no remotes -> rewrite -> local retag; warning skip remote
```

## Steps

1. Create lightweight `v0.0.333`. Do not add a remote.
2. Append `-m` `retag no remote`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	runGit(t, req.Dir, "tag", req.TagName)
	req.Args = append(req.Args, "-m", "retag no remote")
	return nil
}
```
