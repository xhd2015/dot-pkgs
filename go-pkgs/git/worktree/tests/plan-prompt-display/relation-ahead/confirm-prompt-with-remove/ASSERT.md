## Expected

- `FormatPlanPrompt` output does **not** start with a leading blank line (`"\n"`).
- First line is exactly `branch feature is ahead, merge into <default-branch>?`.
- Relation question uses **actual** default branch at main (from `readDefaultBranch`), not a hardcoded wrong name.
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

(No leading `\n` before the question line.)

## Exit Code

- Success path: `Run` returns nil error; merge-back action is `aborted`.

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
	wantFirst := "branch feature is ahead, merge into " + targetBranch + "?"
	firstLine, _, _ := strings.Cut(resp.Output, "\n")
	if firstLine != wantFirst {
		t.Fatalf("first line = %q, want %q", firstLine, wantFirst)
	}

	if strings.Contains(resp.Output, "merge into main?") && targetBranch == "master" {
		t.Fatal("must not hardcode main when checkout is master")
	}

	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	tmpl := "<contains>\n" + `branch feature is ahead, merge into ` + targetBranch + `?
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
