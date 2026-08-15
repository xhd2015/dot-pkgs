# Scenario

**Feature**: merge target is a non-main worktree checkout

```
# release worktree is merge target; label is "release" not main default branch
main + release wt + feature wt -> ahead of release -> TargetPath=release
```

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

	releaseWT := filepath.Join(req.WorkRoot, "release")
	addWorktree(t, mainRepo, releaseWT, "release")

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "ahead.txt"), []byte("x\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "ahead.txt")
	runGit(t, featureWT, "commit", "-m", "ahead")

	req.SourcePath = featureWT
	req.TargetPath = releaseWT
	return nil
}
```