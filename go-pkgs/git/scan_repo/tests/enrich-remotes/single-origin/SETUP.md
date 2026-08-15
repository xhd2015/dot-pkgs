# Scenario

**Feature**: single origin remote is parsed with host, owner, repo

```
git remote add origin -> Remotes[{origin, github.com, xhd2015, lifelog}]
```

## Steps

1. Init repo and add `origin` remote.
2. Set `req.Roots` to workspace.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	repoDir := filepath.Join(root, "proj")
	gitInitRepo(t, repoDir)
	gitRemoteAdd(t, repoDir, "origin", "git@github.com:xhd2015/lifelog.git")
	req.Roots = []string{root}
	return nil
}
```