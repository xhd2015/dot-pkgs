## Expected

- Exit 0. Stderr empty.
- Message `fix typo` with `stripped Co-authored-by`. Author `Bob`, email
  still Alice. `master` moved.

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
		Stripped:         true,
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, "Bob", fixtureAuthorEmail, "fix typo")
	assertFullMessage(t, req.Dir, newSHA, "fix typo")
}
```
