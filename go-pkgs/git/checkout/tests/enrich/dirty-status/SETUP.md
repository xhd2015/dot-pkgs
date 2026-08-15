# Scenario

**Feature**: dirty worktree reports backup-style status with other meta intact

```
committed repo + modified file -> Enrich -> dirty (1 modified) + branch/sha/msg
```

## Steps

1. Create repo with initial commit.
2. Modify tracked file without staging.
3. Set `req.RepoPath`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "dirty")
	gitInitRepo(t, repoDir)
	gitInitialCommit(t, repoDir, "main", "wip")
	readme := filepath.Join(repoDir, "README")
	writeFile(t, readme, "changed\n")
	req.RepoPath = repoDir
	return nil
}
```
