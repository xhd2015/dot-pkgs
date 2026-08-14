## Expected

- Exit 0. Stderr empty.
- `master` and `wrk` move locally. `pushed:` lists only `master`.
- Bare `origin` `refs/heads/master` is the new SHA. `refs/heads/wrk` is absent.

## Exit Code

- 0

```go
import (
	"strings"
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
		MessageFirstLine: "pushed rewrite",
		Branches:         []string{"master", "wrk"},
		Pushed:           []string{"master"},
	}))
	if got := peeled(t, req.BareRemote, "refs/heads/master"); got != newSHA {
		t.Fatalf("origin/master=%s want %s", got, newSHA)
	}
	heads := runGitOutput(t, req.BareRemote, "show-ref", "--heads")
	if strings.Contains(heads, "refs/heads/wrk") {
		t.Fatalf("origin unexpectedly has wrk: %s", heads)
	}
}
```
