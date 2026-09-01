# Scenario

**Feature**: no origin remote → remote sync skipped

```
no origin -> MergeBack uses local land only
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
	initRepo(t, mainRepo, "master")

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "feature.txt"), []byte("f\n"), 0644); err != nil {
		return err
	}
	runGit(t, featureWT, "add", "feature.txt")
	runGit(t, featureWT, "commit", "-m", "feature")

	req.MainRepo = mainRepo
	req.SourcePath = featureWT
	return nil
}
```
