# Scenario

**Feature**: multiple remotes are all listed and parsed

```
origin + upstream -> two Remotes entries with correct owner/repo
```

## Steps

1. Init repo with `origin` and `upstream` remotes.
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
	gitRemoteAdd(t, repoDir, "upstream", "https://github.com/golang/go.git")
	req.Roots = []string{root}
	return nil
}
```