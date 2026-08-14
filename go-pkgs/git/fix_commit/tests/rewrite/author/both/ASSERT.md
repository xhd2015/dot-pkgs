## Expected

- Exit 0. Stderr empty.
- Author `Bob <bob@example.com>`, message still `fix typo`.

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
		AuthorEmail:      "bob@example.com",
		MessageFirstLine: "fix typo",
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, "Bob", "bob@example.com", "fix typo")
	assertFullMessage(t, req.Dir, newSHA, "fix typo")
}
```
