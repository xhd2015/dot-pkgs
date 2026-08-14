## Expected

- Exit 0. Stderr empty.
- Stdout is the locked success report for `master`.
- New commit: same tree, same parent, same author/committer/dates; message
  `corrected message`.
- `refs/heads/master` is the new SHA.

## Expected Output

```
rewrote <old> -> <new>
  author:  Alice <alice@example.com>
  message: corrected message
  branches:
    master
```

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
		AuthorName:       fixtureAuthorName,
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "corrected message",
		Branches:         []string{"master"},
	}))
	assertNewCommit(t, req, newSHA, fixtureAuthorName, fixtureAuthorEmail, "corrected message")
	assertFullMessage(t, req.Dir, newSHA, "corrected message")
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != newSHA {
		t.Fatalf("master=%s want %s", got, newSHA)
	}
}
```
