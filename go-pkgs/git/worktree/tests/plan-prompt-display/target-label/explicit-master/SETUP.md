# Scenario

**Feature**: repos with `master` default must not show hardcoded `main`

```
git init -b master -> target label "master" in prompt and comments
```

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DefaultBranch = "master"

	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		return err
	}
	initRepo(t, mainRepo, "master")

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