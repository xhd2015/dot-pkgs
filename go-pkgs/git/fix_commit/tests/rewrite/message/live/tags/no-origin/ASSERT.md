## Expected

- Exit 0.
- Stderr: `warning: remote origin not found; skip remote tag operations\n`.
- Tag action is `delete local, retag`. Local tag peels to the new SHA.

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
	assertOutput(t, resp.Stderr, "warning: remote origin not found; skip remote tag operations\n")
	_, newSHA := parseRewrote(t, resp.Stdout)
	assertOutput(t, resp.Stdout, formatSuccess(successReport{
		Old:              req.OldSHA,
		New:              newSHA,
		AuthorName:       fixtureAuthorName,
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "retag no remote",
		Branches:         []string{"master"},
		Tags:             []string{"v0.0.333  delete local, retag"},
	}))
	if got := peeled(t, req.Dir, req.TagName); got != newSHA {
		t.Fatalf("local tag %s want %s", got, newSHA)
	}
}
```
