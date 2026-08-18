## Expected
- Exit code 0 (operation aborted cleanly after decline).
- Output lists planned git commands with `#` comments and shortened paths.
- Contains rebase, merge, remove, drop lines and `Proceed? [Y/n]`.
- Worktree still exists; neither side's unique file merged away.

## Exit Code
- 0

```go
import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

func displayGitPath(path, baseDir string) string {
	p := filepath.Clean(path)
	if strings.HasPrefix(p, "/private/var/") {
		p = "/var/" + strings.TrimPrefix(p, "/private/var/")
	}
	return pathfmt.ShortFrom(p, baseDir)
}

func defaultBranchAt(t *testing.T, repo string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "branch", "--show-current").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")
	target := defaultBranchAt(t, mainRepo)
	shortMain := displayGitPath(mainRepo, req.WorkRoot)
	shortWt := displayGitPath(wtDir, req.WorkRoot)

	// Avoid leading "" from raw-string newline after `: ` — FormatPlanPrompt has no leading \n (P1).
	tmpl := "<contains>\n" + `branch feature has diverged, rebase and merge into ` + target + `?
  # feature: rebase onto ` + target + `
  git -C ` + shortWt + ` rebase
  # ` + target + `: fast forward
  git -C ` + shortMain + ` merge --ff-only feature
  # worktree: remove
  git -C ` + shortMain + ` worktree remove ` + shortWt + `
  # worktree branch: drop
  git -C ` + shortMain + ` branch -D feature
Proceed? [Y/n]:
</contains>`
	assert.Output(t, resp.Output, tmpl)

	assertFileExists(t, wtDir)
	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertFileNotExists(t, filepath.Join(mainRepo, "feature-work"))
	assertFileExists(t, filepath.Join(mainRepo, "main-work"))
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```