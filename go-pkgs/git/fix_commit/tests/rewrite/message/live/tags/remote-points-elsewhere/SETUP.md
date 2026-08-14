# Scenario

**Feature**: remote tag at a different commit is not deleted; local retag still happens

```
local tag at old; origin tag forced to parent -> rewrite -> warning; local=new; origin=parent
```

## Steps

1. Tag and push to origin, then point the **bare** tag at the parent commit.
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
	runGit(t, req.Dir, "push", "origin", req.TagName)
	runGit(t, req.BareRemote, "update-ref", "refs/tags/"+req.TagName, req.ParentSHA)
	req.Args = append(req.Args, "-m", "retag local only")
	return nil
}
```
