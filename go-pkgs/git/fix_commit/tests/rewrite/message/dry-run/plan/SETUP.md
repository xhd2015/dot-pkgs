# Scenario

**Feature**: dry-run prints the full plan including tag remote ops and mutates nothing

```
tag + origin + -m --dry-run -> [dry-run] lines -> master/tag/origin still old
```

## Steps

1. Add bare `origin`, lightweight `v0.0.333`, push branch and tag.
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
	runGit(t, req.Dir, "push", "origin", req.TagName)
	req.Args = append(req.Args, "-m", "corrected message", "--dry-run")
	return nil
}
```
