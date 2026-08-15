# Scenario

**Feature**: ahead + dirty → still errors "worktree not clean"

```
# feature branch ahead of main, worktree dirty → MergeBack rejects with clean error
ahead feat -> MergeBack -> IsClean fails -> error
```

## Steps

1. Create feature worktree, commit on it (ahead of main).
2. Make it dirty.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "feat.txt"), []byte("feat\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "feat.txt")
	runGit(t, featureWT, "commit", "-m", "feat commit")

	// ahead: feature has commit, main does not (no divergence)
	makeDirty(t, featureWT)

	req.SourcePath = featureWT
	req.TargetPath = ""
	req.Remove = false
	req.MakeDirty = true
	return nil
}
```
