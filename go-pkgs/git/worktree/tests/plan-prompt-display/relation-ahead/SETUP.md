# Scenario

**Feature**: branch ahead of target — CASE B planned commands

```
# feature worktree commit ahead of target HEAD
main repo + feature worktree -> commit on feature -> relation "ahead"
MergeBack Remove=true -> merge --ff-only + worktree remove + branch -D
```

## Steps

1. Init main repo and `feature` worktree.
2. Commit on feature so it is ahead of target default branch.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		return err
	}
	initRepo(t, mainRepo, req.DefaultBranch)

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")

	if err := os.WriteFile(filepath.Join(featureWT, "ahead.txt"), []byte("x\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "ahead.txt")
	runGit(t, featureWT, "commit", "-m", "ahead")

	req.SourcePath = featureWT
	req.TargetPath = ""
	return nil
}
```