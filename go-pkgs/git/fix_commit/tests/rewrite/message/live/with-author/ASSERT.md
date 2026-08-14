## Expected

- Exit 0. Stderr empty.
- New commit author is `Bob <bob@example.com>`; message `all three`.
- Committer and both dates unchanged. `master` at the new SHA.

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
	old, newSHA := parseRewrote(t, resp.Stdout)
	if old != req.OldSHA {
		t.Fatalf("rewrote old %s want %s", old, req.OldSHA)
	}
	assertOutput(t, resp.Stdout, formatSuccess(successReport{
		Old:              req.OldSHA,
		New:              newSHA,
		AuthorName:       "Bob",
		AuthorEmail:      "bob@example.com",
		MessageFirstLine: "all three",
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, "Bob", "bob@example.com", "all three")
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != newSHA {
		t.Fatalf("master=%s want %s", got, newSHA)
	}
}
```
