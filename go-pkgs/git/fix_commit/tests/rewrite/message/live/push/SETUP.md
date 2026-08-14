# Scenario

**Feature**: `--push` force-with-lease updates remotes whose upstream is still the old SHA

```
origin/master tracks old; wrk has no upstream -> --push -> origin/master=new; wrk not on remote
```

## Steps

1. Add a local bare `origin`, push `-u master`.
2. Create local `wrk` at the old SHA with **no** upstream.
3. Append `-m` `pushed rewrite` `--push`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BareRemote = addBareRemote(t, req.Dir, "origin")
	runGit(t, req.Dir, "push", "-u", "origin", "master")
	runGit(t, req.Dir, "branch", "wrk", req.OldSHA)
	req.Args = append(req.Args, "-m", "pushed rewrite", "--push")
	return nil
}
```
