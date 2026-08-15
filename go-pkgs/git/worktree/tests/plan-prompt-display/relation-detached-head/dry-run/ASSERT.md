## Expected

- Dry-run lists a fast-forward merge using the detached worktree commit SHA.
- Must not treat detached ahead commit as already-included (no empty command list).
- Output does **not** start with a leading blank line (`"\n"`).

## Errors

- None.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	mainRepo := filepath.Join(req.WorkRoot, "main")
	featureWT := filepath.Join(req.WorkRoot, "feature")
	targetBranch := readDefaultBranch(t, mainRepo)
	shortMain := shortPath(t, mainRepo)
	shortFeature := shortPath(t, featureWT)
	commit := revParseHEAD(t, featureWT)

	if strings.HasPrefix(resp.Output, "\n") {
		t.Fatal(`printDryRun output must not start with leading "\n"`)
	}

	if strings.Contains(strings.ToLower(resp.Output), "already included") {
		t.Fatalf("detached ahead commit must not be reported as already included, got:\n%s", resp.Output)
	}
	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	tmpl := fmt.Sprintf(`<contains>
  # %s: fast forward
  git -C %s merge --ff-only %s
  # worktree: remove
  git -C %s worktree remove %s
</contains>`, targetBranch, shortMain, commit, shortMain, shortFeature)
	assert.Output(t, resp.Output, tmpl)

	if strings.Contains(resp.Output, "merge --ff-only HEAD") {
		t.Fatalf("must not use symbolic HEAD ref in merge command, got:\n%s", resp.Output)
	}
	if strings.Contains(resp.Output, "merge --ff-only feature") {
		t.Fatalf("detached worktree must not use branch name when not checked out, got:\n%s", resp.Output)
	}
}
```
