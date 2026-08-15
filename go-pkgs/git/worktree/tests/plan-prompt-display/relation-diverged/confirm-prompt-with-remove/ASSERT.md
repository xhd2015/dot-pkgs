## Expected

- `FormatPlanPrompt` output does **not** start with a leading blank line (`"\n"`).
- First line is exactly `branch feature has diverged, rebase and merge into <target>?`.
- Comment `# feature: rebase onto <target>` before rebase command at feature worktree.
- Merge, remove, and drop comments match CASE B pattern at main repo paths (shortened).

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
		t.Fatal(`FormatPlanPrompt output must not start with leading "\n"`)
	}
	wantFirst := "branch feature has diverged, rebase and merge into " + targetBranch + "?"
	firstLine, _, _ := strings.Cut(resp.Output, "\n")
	if firstLine != wantFirst {
		t.Fatalf("first line = %q, want %q", firstLine, wantFirst)
	}

	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	tmpl := "<contains>\n" + `branch feature has diverged, rebase and merge into ` + targetBranch + `?
  # feature: rebase onto ` + targetBranch + `
  git -C ` + shortFeature + ` rebase
  # ` + targetBranch + `: fast forward
  git -C ` + shortMain + ` merge --ff-only feature
  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortFeature + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
Proceed? [Y/n]:
</contains>`
	assert.Output(t, resp.Output, tmpl)
}
```
