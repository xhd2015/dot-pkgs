# Scenario

**Feature**: main repo nests linked worktree with enriched checkout meta

```
main + feature-a linked worktree -> Scan -> Build -> one main Node with child worktree
```

## Preconditions

- `git` on PATH (skip otherwise).

## Steps

1. Create main repo with commit on `main`.
2. Add linked worktree `feature-a` on branch `feature/foo`.
3. Scan workspace and build snapshot with `BaseDir` = workspace root.

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
	gitInitialCommit(t, mainDir)
	runGit(t, mainDir, "branch", "-M", "main")
	gitWorktreeAdd(t, mainDir, wtDir, "feature/foo")
	req.Mode = "scan"
	req.BaseDir = root
	req.ScanRoots = []string{root}
	return nil
}
```
