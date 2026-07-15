# Scenario

**Feature**: enabling cache write does not change discovery `Result.Repos` shape

```
# same fixture as scan/single-repo
workspace/my-repo
  -> Scan(CacheRoot, NoCache=false)
  -> Repos: one main, Name=my-repo, GitDir=.../.git, empty Remotes/Worktrees
```

## Steps

1. Create workspace with one repo at `my-repo/` (identical layout to discovery single-repo).
2. Set `req.Roots` to the workspace; cache write remains enabled via parent.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	req.Roots = []string{root}
	return nil
}
```
