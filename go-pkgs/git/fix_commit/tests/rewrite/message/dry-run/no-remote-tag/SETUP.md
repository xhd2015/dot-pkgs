# Scenario

**Feature**: dry-run notices a missing remote tag and does not plan remote delete/push

```
local tag + origin without that tag + --dry-run -> notice; no remote ops; no mutations
```

## Steps

1. Add bare `origin`, push `master` only. Lightweight `v0.0.333` stays local.
2. Append `-m` `corrected message` `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.TagName = "v0.0.333"
	req.BareRemote = addBareRemote(t, req.Dir, "origin")
	runGit(t, req.Dir, "tag", req.TagName)
	runGit(t, req.Dir, "push", "-u", "origin", "master")
	req.Args = append(req.Args, "-m", "corrected message", "--dry-run")
	return nil
}
```
