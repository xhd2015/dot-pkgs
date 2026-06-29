# Scenario

**Feature**: diverged branches — CASE C planned commands

```
# main and feature each have unique commits
commit on feature -> commit on main -> relation "diverged"
```

## Steps

1. Init repo + feature worktree.
2. Commit on feature, then commit on main (diverged).

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
	if err := os.WriteFile(filepath.Join(featureWT, "feature-only.txt"), []byte("f\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "feature-only.txt")
	runGit(t, featureWT, "commit", "-m", "feature only")

	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("m\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "main only")

	req.SourcePath = featureWT
	req.TargetPath = ""
	return nil
}
```