# Scenario

**Feature**: already-included (ancestor) + dirty → still errors "worktree not clean"

```
# feature worktree is at commit that main already contains, dirty → rejected
ancestor feat -> MergeBack -> IsClean fails -> error
```

## Steps

1. Create feature worktree (no extra commits — same as main).
2. Make it dirty.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := req.MainRepo

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")

	// no extra commit on feature → feature is ancestor of main
	makeDirty(t, featureWT)

	req.SourcePath = featureWT
	req.TargetPath = ""
	req.Remove = false
	req.MakeDirty = true
	return nil
}
```
