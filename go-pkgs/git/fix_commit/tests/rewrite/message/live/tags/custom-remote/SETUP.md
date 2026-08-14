# Scenario

**Feature**: `--remote other` deletes/pushes that remote; `origin` tag stays

```
origin + other both tagged at old -> --remote other -> other=new; origin=old
```

## Steps

1. Add remotes `origin` and `other`. Tag and push the tag to both.
2. Append `-m` `retag other` `--remote` `other`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BareRemote = addBareRemote(t, req.Dir, "origin")
	req.OtherRemote = addBareRemote(t, req.Dir, "other")
	runGit(t, req.Dir, "tag", req.TagName)
	runGit(t, req.Dir, "push", "-u", "origin", "master")
	runGit(t, req.Dir, "push", "origin", req.TagName)
	runGit(t, req.Dir, "push", "other", "master")
	runGit(t, req.Dir, "push", "other", req.TagName)
	req.Args = append(req.Args, "-m", "retag other", "--remote", "other")
	return nil
}
```
