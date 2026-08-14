## Expected

- Exit 0. Stderr empty.
- Success report includes `stripped Co-authored-by`.
- New message is exactly `fix typo` (blank trailer separator gone).
- Author unchanged. `master` moved.

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
		MessageFirstLine: "fix typo",
		Stripped:         true,
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, fixtureAuthorName, fixtureAuthorEmail, "fix typo")
	assertFullMessage(t, req.Dir, newSHA, "fix typo")
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != newSHA {
		t.Fatalf("master=%s want %s", got, newSHA)
	}
}
```
