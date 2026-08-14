# Scenario

**Feature**: lightweight tag at the old SHA is deleted, retagged, and pushed

```
tag v0.0.333 + push origin -> rewrite -> local+origin tag peel to new SHA
```

## Steps

1. Add bare `origin`, lightweight tag `v0.0.333`, push branch and tag.
2. Append `-m` `retag light`.

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
	req.Args = append(req.Args, "-m", "retag light")
	return nil
}
```
