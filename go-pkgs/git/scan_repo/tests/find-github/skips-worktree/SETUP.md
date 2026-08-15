# Scenario

**Feature**: linked worktree is skipped; main repo path is returned

## Steps

1. Init main repo with matching origin.
2. Add linked worktree `feature-a`.
3. Find by github identity returns main path only.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-a")
	gitInitRepo(t, mainDir)
	gitRemoteAdd(t, mainDir, "origin", "git@github.com:xhd2015/myproject.git")
	gitInitialCommit(t, mainDir)
	runGit(t, mainDir, "worktree", "add", wtDir)
	req.Roots = []string{root}
	req.FindGitHubOwner = "xhd2015"
	req.FindGitHubRepo = "myproject"
	return nil
}
```