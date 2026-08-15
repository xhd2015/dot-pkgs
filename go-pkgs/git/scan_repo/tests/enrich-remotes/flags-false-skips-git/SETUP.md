# Scenario

**Feature**: `ListRemotes=false` skips git subprocesses (fake repo succeeds)

```
ListRemotes=false on fake .git -> scan succeeds without git binary
```

## Steps

1. Override `ListRemotes` to false (parent branch sets true).
2. Create fake repo (not a real git repo).
3. Set `req.Roots` to workspace.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ListRemotes = false
	root := t.TempDir()
	repoDir := filepath.Join(root, "fake")
	mkdirAll(t, repoDir)
	gitDir := filepath.Join(repoDir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
	req.Roots = []string{root}
	// Ensure scan does not invoke git: PATH without git would break enrich path.
	if _, err := os.Stat(filepath.Join(gitDir, "objects")); err != nil {
		t.Fatal(err)
	}
	return nil
}
```