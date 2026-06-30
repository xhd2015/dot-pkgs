## Expected

- Dry-run merge command references branch name `feature`, not a raw commit hash.

## Errors

- None.

```go
import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

var fullCommitHash = regexp.MustCompile(`[0-9a-f]{40}`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	mainRepo := filepath.Join(req.WorkRoot, "main")
	targetBranch := readDefaultBranch(t, mainRepo)
	shortMain := shortPath(t, mainRepo)

	tmpl := `
<contains>
  # ` + targetBranch + `: fast forward
  git -C ` + shortMain + ` merge --ff-only feature
</contains>`
	assert.Output(t, resp.Output, tmpl)

	for _, line := range strings.Split(resp.Output, "\n") {
		if !strings.Contains(line, "merge --ff-only") {
			continue
		}
		if fullCommitHash.MatchString(line) {
			t.Fatalf("attached worktree must use branch name, not commit hash, got line:\n%s", line)
		}
	}
}
```