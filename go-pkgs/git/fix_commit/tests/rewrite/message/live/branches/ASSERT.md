## Expected

- Exit 0. Stderr empty.
- `master`, `wrk-a`, `wrk-b` listed (lexicographic) and point at the new SHA.
- `other` still at the parent SHA.

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
		MessageFirstLine: "move tips",
		Branches:         []string{"master", "wrk-a", "wrk-b"},
	}))
	for _, br := range []string{"master", "wrk-a", "wrk-b"} {
		if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/"+br); got != newSHA {
			t.Fatalf("%s=%s want %s", br, got, newSHA)
		}
	}
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/other"); got != req.UnrelatedSHA {
		t.Fatalf("other=%s want %s", got, req.UnrelatedSHA)
	}
}
```
