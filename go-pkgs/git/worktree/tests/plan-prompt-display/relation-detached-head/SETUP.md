# Scenario

**Feature**: detached HEAD worktree ahead of target

```
# worktree checked out at commit ahead of target, not on a branch
main repo + feature worktree -> commit -> checkout --detach -> relation "ahead"
MergeBack must compare worktree commit SHA, not symbolic HEAD in target repo
```

## Steps

1. Init main repo and `feature` worktree.
2. Commit on feature so it is ahead of target default branch.
3. Detach HEAD in the feature worktree.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		return err
	}
	initRepo(t, mainRepo, req.DefaultBranch)

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")

	if err := os.WriteFile(filepath.Join(featureWT, "detached-ahead.txt"), []byte("x\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "detached-ahead.txt")
	runGit(t, featureWT, "commit", "-m", "detached ahead")
	runGit(t, featureWT, "checkout", "--detach")

	req.SourcePath = featureWT
	req.TargetPath = ""
	return nil
}
```