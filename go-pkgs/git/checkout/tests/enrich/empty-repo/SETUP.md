# Scenario

**Feature**: unborn HEAD yields enrichment error in Meta

```
git init only (no commits) -> Enrich -> Error: no commits (HEAD unborn)
```

## Steps

1. Create repo with `git init` but no add/commit.
2. Set `req.RepoPath` to repo directory.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "empty")
	gitInitRepo(t, repoDir)
	req.RepoPath = repoDir
	return nil
}
```
