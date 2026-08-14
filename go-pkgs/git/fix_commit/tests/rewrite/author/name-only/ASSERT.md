## Expected

- Exit 0. Stderr empty.
- Author name `Bob`, email still `alice@example.com`, message `fix typo`.
- `master` at the new SHA.

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
		AuthorName:       "Bob",
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "fix typo",
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, "Bob", fixtureAuthorEmail, "fix typo")
	assertFullMessage(t, req.Dir, newSHA, "fix typo")
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != newSHA {
		t.Fatalf("master=%s want %s", got, newSHA)
	}
}
```
