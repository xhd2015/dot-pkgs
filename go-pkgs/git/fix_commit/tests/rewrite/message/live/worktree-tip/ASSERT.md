## Expected

- Exit 0. Stderr empty.
- `master` and `wrk` (lexicographic) move to the new SHA.
- `git -C <worktree> rev-parse HEAD` is the new SHA.

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
		MessageFirstLine: "move worktree tip",
		Branches:         []string{"master", "wrk"},
	}))
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/wrk"); got != newSHA {
		t.Fatalf("wrk=%s want %s", got, newSHA)
	}
	if got := runGitOutput(t, req.WorktreeDir, "rev-parse", "HEAD"); got != newSHA {
		t.Fatalf("worktree HEAD=%s want %s", got, newSHA)
	}
}
```
