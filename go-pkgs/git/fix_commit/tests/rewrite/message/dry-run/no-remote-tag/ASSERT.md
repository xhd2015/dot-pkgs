## Expected

- Exit 0. Stderr empty.
- Plan includes `notice: remote origin has no tag v0.0.333` and no remote
  delete/push lines.
- `master`, local tag, and origin stay on the old SHA. Origin still has no tag.

## Expected Output

```
[dry-run] would rewrite <old>
[dry-run]   author:  Alice <alice@example.com>
[dry-run]   message: corrected message
[dry-run]   move branch master
[dry-run]   delete local tag v0.0.333
[dry-run]   notice: remote origin has no tag v0.0.333
[dry-run]   tag v0.0.333 at new commit
```

## Exit Code

- 0

```go
import (
	"fmt"
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
	want := fmt.Sprintf(`[dry-run] would rewrite %s
[dry-run]   author:  %s <%s>
[dry-run]   message: corrected message
[dry-run]   move branch master
[dry-run]   delete local tag v0.0.333
[dry-run]   notice: remote origin has no tag v0.0.333
[dry-run]   tag v0.0.333 at new commit
`, req.OldSHA, fixtureAuthorName, fixtureAuthorEmail)
	assertOutput(t, resp.Stdout, want)
	assertUnchangedSHA(t, req)
	if got := peeled(t, req.Dir, req.TagName); got != req.OldSHA {
		t.Fatalf("local tag moved to %s", got)
	}
	if got := runGitOutput(t, req.BareRemote, "tag", "-l", req.TagName); got != "" {
		t.Fatalf("origin unexpectedly has %s", got)
	}
}
```
