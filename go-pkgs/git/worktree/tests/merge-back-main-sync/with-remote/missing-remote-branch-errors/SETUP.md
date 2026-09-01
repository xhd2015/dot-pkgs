# Scenario

**Feature**: origin exists but remote branch missing → fetch fails

```
empty bare origin (no master) -> fetch error
```

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Replace parent with-remote fixture: empty bare + local main/feature only.
	workRoot := req.WorkRoot
	mainRepo := filepath.Join(workRoot, "main2")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		return err
	}
	initRepo(t, mainRepo, "master")

	bare := filepath.Join(workRoot, "empty.git")
	runGit(t, workRoot, "init", "--bare", bare)
	runGit(t, mainRepo, "remote", "add", "origin", bare)
	// Do not push — origin/master does not exist.

	featureWT := filepath.Join(workRoot, "feature2")
	addWorktree(t, mainRepo, featureWT, "feature2")
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
