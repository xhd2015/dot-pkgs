## Expected
- Exit code is 0 (validation passes — worktree is clean, branch is merged).
- Output contains planned `git -C` commands with `#` comments and shortened paths.
- Output contains `dry-run: would remove worktree`.
- Command paths match merge-back display formatting.
- The worktree directory still exists; history still has the worktree entry.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

func displayGitPath(path string) string {
	p := filepath.Clean(path)
	if strings.HasPrefix(p, "/private/var/") {
		p = "/var/" + strings.TrimPrefix(p, "/private/var/")
	}
	return pathfmt.Short(p)
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	mainRepo := filepath.Join(req.WorkRoot, "main")
	wtDir := filepath.Join(req.WorkRoot, "feature")
	shortMain := displayGitPath(mainRepo)
	shortWt := displayGitPath(wtDir)

	tmpl := `
<contains>
  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortWt + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
dry-run: would remove worktree
</contains>`
	assert.Output(t, resp.Output, tmpl)

	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertHistoryLen(t, req.ConfigHome, 1)
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```