## Expected

- Exit 0. Stderr empty.
- Tag action: `v0.0.333  delete local+other, retag, push`.
- Local and `other` peel to the new SHA. `origin` tag still peels to the old SHA.

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
		MessageFirstLine: "retag other",
		Branches:         []string{"master"},
		Tags:             []string{"v0.0.333  delete local+other, retag, push"},
	}))
	if got := peeled(t, req.Dir, req.TagName); got != newSHA {
		t.Fatalf("local tag %s want %s", got, newSHA)
	}
	if got := peeled(t, req.OtherRemote, req.TagName); got != newSHA {
		t.Fatalf("other tag %s want %s", got, newSHA)
	}
	if got := peeled(t, req.BareRemote, req.TagName); got != req.OldSHA {
		t.Fatalf("origin tag %s want old %s", got, req.OldSHA)
	}
}
```
