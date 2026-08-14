## Expected

- Exit 0.
- Stderr is
  `warning: commit has descendants; those commits still parent <oldsha>\n`.
- `tip-here` moves to the new SHA. `master` stays on the child.
- Child commit’s parent is still the **old** SHA.

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
	assertOutput(t, resp.Stderr, "warning: commit has descendants; those commits still parent "+req.OldSHA+"\n")
	_, newSHA := parseRewrote(t, resp.Stdout)
	assertOutput(t, resp.Stdout, formatSuccess(successReport{
		Old:              req.OldSHA,
		New:              newSHA,
		AuthorName:       fixtureAuthorName,
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "rewrite ancestor",
		Branches:         []string{"tip-here"},
	}))
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/tip-here"); got != newSHA {
		t.Fatalf("tip-here=%s want %s", got, newSHA)
	}
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != req.ChildSHA {
		t.Fatalf("master=%s want child %s", got, req.ChildSHA)
	}
	if got := commitField(t, req.Dir, req.ChildSHA, "%P"); got != req.OldSHA {
		t.Fatalf("child parent %s want old %s", got, req.OldSHA)
	}
}
```
