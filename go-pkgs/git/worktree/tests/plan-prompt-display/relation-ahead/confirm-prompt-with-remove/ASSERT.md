## Expected

- `FormatPlanPrompt` output includes relation question with **actual** default branch at main (from `readDefaultBranch`), not a hardcoded wrong name.
- Comment `# <target>: fast forward` before merge line.
- Comments `# worktree: remove` and `# worktree branch: drop` before cleanup commands.
- `git -C` uses shortened main repo path; `worktree remove` uses shortened feature path.
- Contains `merge --ff-only feature`, `Proceed? [Y/n]:`.
## Expected Output

```
branch feature is ahead, merge into <default-branch>?
  # <default-branch>: fast forward
  git -C <short-main> merge --ff-only feature
  # worktree: remove
  git -C <short-main> worktree remove <short-feature>
  # worktree branch: drop
  git -C <short-main> branch -D feature
Proceed? [Y/n]:
```

## Exit Code

- Success path: `Run` returns nil error; merge-back action is `aborted`.

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

	if strings.Contains(resp.Output, "merge into main?") && targetBranch == "master" {
		t.Fatal("must not hardcode main when checkout is master")
	}

	tmpl := `
<contains>
branch feature is ahead, merge into ` + targetBranch + `?
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