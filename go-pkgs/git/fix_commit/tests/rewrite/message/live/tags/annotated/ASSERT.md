## Expected

- Exit 0. Stderr empty.
- Local tag type is `tag`. Message still `release 333`. Tagger name/email/date
  match the captured original. Object peels to the new SHA. Origin peels to
  the new SHA.

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
		MessageFirstLine: "retag annotated",
		Branches:         []string{"master"},
		Tags:             []string{"v0.0.333  delete local+origin, retag, push"},
	}))
	if got := runGitOutput(t, req.Dir, "cat-file", "-t", req.TagName); got != "tag" {
		t.Fatalf("local tag type %q want tag", got)
	}
	gotMsg := runGitOutput(t, req.Dir, "tag", "-l", "--format=%(contents)", req.TagName)
	if strings.TrimSpace(gotMsg) != strings.TrimSpace(req.TagMessage) {
		t.Fatalf("tag message %q want %q", gotMsg, req.TagMessage)
	}
	if got := runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggername)", req.TagName); got != req.TaggerName {
		t.Fatalf("tagger name %q want %q", got, req.TaggerName)
	}
	if got := runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggeremail)", req.TagName); got != req.TaggerEmail {
		t.Fatalf("tagger email %q want %q", got, req.TaggerEmail)
	}
	if got := runGitOutput(t, req.Dir, "tag", "-l", "--format=%(taggerdate:unix)", req.TagName); got != req.TaggerUnix {
		t.Fatalf("tagger date %q want %q", got, req.TaggerUnix)
	}
	if got := peeled(t, req.Dir, req.TagName); got != newSHA {
		t.Fatalf("local tag %s want %s", got, newSHA)
	}
	if got := peeled(t, req.BareRemote, req.TagName); got != newSHA {
		t.Fatalf("origin tag %s want %s", got, newSHA)
	}
}
```
