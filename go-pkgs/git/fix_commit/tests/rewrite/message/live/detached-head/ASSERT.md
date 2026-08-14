## Expected

- Exit 0. Stderr empty.
- `HEAD` is the new SHA. `master` is still the parent.
- Success report has no `branches:` section (no exact-tip branch).

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
		MessageFirstLine: "detached rewrite",
	}))
	if got := runGitOutput(t, req.Dir, "rev-parse", "HEAD"); got != newSHA {
		t.Fatalf("HEAD=%s want %s", got, newSHA)
	}
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != req.ParentSHA {
		t.Fatalf("master=%s want parked parent %s", got, req.ParentSHA)
	}
}
```
