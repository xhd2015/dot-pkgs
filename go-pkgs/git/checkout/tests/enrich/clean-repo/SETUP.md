# Scenario

**Feature**: committed repo yields full clean metadata

```
git init + commit on main -> Enrich -> branch, 7-char sha, msg, clean
```

## Steps

1. Create repo with branch `main` and commit message `backup fixture`.
2. Set `req.RepoPath` to repo directory.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "main")
	gitInitRepo(t, repoDir)
	gitInitialCommit(t, repoDir, "main", "backup fixture")
	req.RepoPath = repoDir
	return nil
}
```
