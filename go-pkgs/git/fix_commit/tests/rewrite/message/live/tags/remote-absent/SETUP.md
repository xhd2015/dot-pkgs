# Scenario

**Feature**: origin exists but has no matching tag → local retag + stdout notice

```
local tag + origin without that tag -> rewrite -> local retag; notice skip remote
```

## Steps

1. Add bare `origin`, push `master` only. Lightweight `v0.0.333` stays local.
2. Append `-m` `retag local only`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BareRemote = addBareRemote(t, req.Dir, "origin")
	runGit(t, req.Dir, "tag", req.TagName)
	runGit(t, req.Dir, "push", "-u", "origin", "master")
	req.Args = append(req.Args, "-m", "retag local only")
	return nil
}
```
