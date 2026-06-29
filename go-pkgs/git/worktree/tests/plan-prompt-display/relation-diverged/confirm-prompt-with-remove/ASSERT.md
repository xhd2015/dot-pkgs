## Expected

- Question: `branch feature has diverged, rebase and merge into <target>?`
- Comment `# feature: rebase onto <target>` before rebase command at feature worktree.
- Merge, remove, and drop comments match CASE B pattern at main repo paths (shortened).

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	mainRepo := filepath.Join(req.WorkRoot, "main")
	featureWT := filepath.Join(req.WorkRoot, "feature")
	targetBranch := readDefaultBranch(t, mainRepo)
	shortMain := shortPath(t, mainRepo)
	shortFeature := shortPath(t, featureWT)

	tmpl := `
<contains>
branch feature has diverged, rebase and merge into ` + targetBranch + `?
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