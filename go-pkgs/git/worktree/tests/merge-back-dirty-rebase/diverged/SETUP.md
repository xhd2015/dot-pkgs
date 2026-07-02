# Scenario

**Feature**: diverged branches — rebase required

```
# main and feature each have unique commits → relation "diverged"
commit on feature -> commit on main -> diverged
```

## Steps

1. Create main repo with base commit.
2. Create feature worktree, commit on it.
3. Commit on main (diverges from feature).
4. Each sub-scenario applies additional preconditions (clean/dirty, --rm/not).

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
	initRepo(t, mainRepo, "")

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "feature-only.txt"), []byte("f\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "feature-only.txt")
	runGit(t, featureWT, "commit", "-m", "feature only")

	// diverge: commit on main
	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("m\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "main only")

	req.MainRepo = mainRepo
	req.SourcePath = featureWT
	req.TargetPath = ""
	return nil
}
```
