# Scenario

**Feature**: annotated tag payload and tagger are preserved on the new commit

```
git tag -a v0.0.333 -m "release 333" + push -> rewrite -> annotated retag, same message/tagger
```

## Steps

1. Add bare `origin`. Create annotated tag `v0.0.333` with message
   `release 333`. Push branch and tag.
2. Record the tagger fields. Append `-m` `retag annotated`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BareRemote = addBareRemote(t, req.Dir, "origin")
	runGit(t, req.Dir, "tag", "-a", req.TagName, "-m", "release 333")
	req.TagMessage = runGitOutput(t, req.Dir, "tag", "-l", "--format=%(contents)", req.TagName)
	req.TaggerName = runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggername)", req.TagName)
	req.TaggerEmail = runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggeremail)", req.TagName)
	req.TaggerUnix = runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggerdate:unix)", req.TagName)
	runGit(t, req.Dir, "push", "-u", "origin", "master")
	runGit(t, req.Dir, "push", "origin", req.TagName)
	req.Args = append(req.Args, "-m", "retag annotated")
	return nil
}
```
