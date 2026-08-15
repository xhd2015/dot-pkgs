## Expected

- `printDryRun` stdout does **not** start with a leading blank line (`"\n"`).
- First line is exactly `  # <target>: fast forward` (indented planned-command comment).
- Stdout lists three command groups with the same `#` comments and Short paths as the confirm prompt body.
- No `Proceed?` line and no `branch ... is ahead` question line.

## Errors

- None.

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
	targetBranch := readDefaultBranch(t, mainRepo)
	shortMain := shortPath(t, mainRepo)
	shortFeature := shortPath(t, featureWT)

	if strings.HasPrefix(resp.Output, "\n") {
		t.Fatal(`printDryRun output must not start with leading "\n"`)
	}
	wantFirst := "  # " + targetBranch + ": fast forward"
	firstLine, _, _ := strings.Cut(resp.Output, "\n")
	if firstLine != wantFirst {
		t.Fatalf("first line = %q, want %q", firstLine, wantFirst)
	}

	if strings.Contains(resp.Output, "Proceed?") {
		t.Fatal("dry-run listing must not include confirmation prompt")
	}
	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	tmpl := "<contains>\n" + `  # ` + targetBranch + `: fast forward
  git -C ` + shortMain + ` merge --ff-only feature
  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortFeature + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
</contains>`
	assert.Output(t, resp.Output, tmpl)
}
```
