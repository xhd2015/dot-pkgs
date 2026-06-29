# Scenario

**Feature**: branch already included — dry-run lists remove-only commands

```
# feature at same commit as target (no merge/rebase needed)
main + feature worktree at same HEAD -> relation ancestor/same
```

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

	req.SourcePath = featureWT
	req.TargetPath = ""
	req.DryRun = true
	req.CapturePrompt = false
	return nil
}
```