## Expected

- Exit 0. Stderr empty.
- Tag line: `v0.0.333  delete local+origin, retag, push`.
- Local tag is still lightweight (`cat-file -t` is `commit`) and peels to the
  new SHA. Bare origin tag peels to the new SHA.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	requireHarnessOK(t, err)
	requireExit(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("stderr=%q want empty", resp.Stderr)
	}
	_, newSHA := parseRewrote(t, resp.Stdout)
	assertOutput(t, resp.Stdout, formatSuccess(successReport{
		Old:              req.OldSHA,
		New:              newSHA,
		AuthorName:       fixtureAuthorName,
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "retag light",
		Branches:         []string{"master"},
		Tags:             []string{"v0.0.333  delete local+origin, retag, push"},
	}))
	if got := runGitOutput(t, req.Dir, "cat-file", "-t", req.TagName); got != "commit" {
		t.Fatalf("local tag type %q want commit (lightweight)", got)
	}
	if got := peeled(t, req.Dir, req.TagName); got != newSHA {
		t.Fatalf("local tag %s want %s", got, newSHA)
	}
	if got := peeled(t, req.BareRemote, req.TagName); got != newSHA {
		t.Fatalf("origin tag %s want %s", got, newSHA)
	}
}
```
