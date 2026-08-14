## Expected

- Exit 0. Stderr empty (notice is not a warning).
- Tag action is `delete local, retag`. Local tag peels to the new SHA.
- Origin has no `v0.0.333`.
- Stdout ends with `notice: remote origin has no tag v0.0.333`.

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
	want := formatSuccess(successReport{
		Old:              req.OldSHA,
		New:              newSHA,
		AuthorName:       fixtureAuthorName,
		AuthorEmail:      fixtureAuthorEmail,
		MessageFirstLine: "retag local only",
		Branches:         []string{"master"},
		Tags:             []string{"v0.0.333  delete local, retag"},
	})
	want += "notice: remote origin has no tag v0.0.333\n"
	assertOutput(t, resp.Stdout, want)
	if got := peeled(t, req.Dir, req.TagName); got != newSHA {
		t.Fatalf("local tag %s want %s", got, newSHA)
	}
	if got := runGitOutput(t, req.BareRemote, "tag", "-l", req.TagName); got != "" {
		t.Fatalf("origin unexpectedly has %s", got)
	}
}
```
