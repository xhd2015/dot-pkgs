## Expected

- Four command blocks with comments; rebase line at feature checkout; no prompt trailer.

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

	if strings.Contains(resp.Output, "Proceed?") {
		t.Fatal("dry-run must not include Proceed prompt")
	}

	tmpl := `
<contains>
  # feature: rebase onto ` + targetBranch + `
  git -C ` + shortFeature + ` rebase
  # ` + targetBranch + `: fast forward
  git -C ` + shortMain + ` merge --ff-only feature
  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortFeature + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
</contains>`
	assert.Output(t, resp.Output, tmpl)
}
```