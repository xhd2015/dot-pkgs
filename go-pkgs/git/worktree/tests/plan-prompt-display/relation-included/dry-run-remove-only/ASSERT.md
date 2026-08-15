## Expected

- No `merge --ff-only` or `rebase` lines.
- `# worktree: remove` and `# worktree branch: drop` with shortened paths.
- Output does **not** start with a leading blank line (`"\n"`).

```go
import (
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
	shortMain := shortPath(t, mainRepo)
	shortFeature := shortPath(t, featureWT)

	if strings.HasPrefix(resp.Output, "\n") {
		t.Fatal(`printDryRun output must not start with leading "\n"`)
	}

	if strings.Contains(resp.Output, "merge --ff-only") || strings.Contains(resp.Output, " rebase") {
		t.Fatal("included relation must not list merge or rebase commands")
	}

	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	tmpl := "<contains>\n" + `  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortFeature + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
</contains>`
	assert.Output(t, resp.Output, tmpl)
}
```
