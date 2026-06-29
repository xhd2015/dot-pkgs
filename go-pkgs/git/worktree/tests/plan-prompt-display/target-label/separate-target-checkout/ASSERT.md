## Expected

- Question references `release`, not the main repo default branch name.
- Fast-forward comment uses `# release: fast forward`.
- Merge runs at shortened release worktree path (`git -C <short-release> merge ...`).

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
	releaseWT := filepath.Join(req.WorkRoot, "release")
	defaultBranch := readDefaultBranch(t, mainRepo)
	shortRelease := shortPath(t, releaseWT)

	if strings.Contains(resp.Output, "merge into "+defaultBranch+"?") && defaultBranch != "release" {
		t.Fatalf("expected merge into release, not default %q:\n%s", defaultBranch, resp.Output)
	}

	tmpl := `
<contains>
branch feature is ahead, merge into release?
  # release: fast forward
  git -C ` + shortRelease + ` merge --ff-only feature
</contains>`
	assert.Output(t, resp.Output, tmpl)
}
```