## Expected

- Exit 0. Stderr empty.
- Plan lists stripped subject `fix typo` and `stripped Co-authored-by`.
- HEAD still the old SHA; message still contains `Co-authored-by`.

## Expected Output

```
[dry-run] would rewrite <old>
[dry-run]   author:  Alice <alice@example.com>
[dry-run]   message: fix typo
[dry-run]   stripped Co-authored-by
[dry-run]   move branch master
```

## Exit Code

- 0

```go
import (
	"fmt"
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
	want := fmt.Sprintf(`[dry-run] would rewrite %s
[dry-run]   author:  %s <%s>
[dry-run]   message: fix typo
[dry-run]   stripped Co-authored-by
[dry-run]   move branch master
`, req.OldSHA, fixtureAuthorName, fixtureAuthorEmail)
	assertOutput(t, resp.Stdout, want)
	assertUnchangedSHA(t, req)
	msg := runGitOutput(t, req.Dir, "log", "-1", "--format=%B")
	if !strings.Contains(strings.ToLower(msg), "co-authored-by:") {
		t.Fatalf("dry-run mutated message: %q", msg)
	}
}
```
