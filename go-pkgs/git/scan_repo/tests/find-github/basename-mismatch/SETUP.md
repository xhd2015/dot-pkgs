# Scenario

**Feature**: local clone basename differs from GitHub repo name

```
~/Projects/myproject-clone origin xhd2015/myproject -> match myproject
```

## Steps

1. Init `myproject-clone` with origin `git@github.com:xhd2015/myproject.git`.
2. Set find target `xhd2015` / `myproject`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	cloneDir := filepath.Join(root, "myproject-clone")
	gitInitRepo(t, cloneDir)
	gitRemoteAdd(t, cloneDir, "origin", "git@github.com:xhd2015/myproject.git")
	gitInitialCommit(t, cloneDir)
	req.Roots = []string{root}
	req.FindGitHubOwner = "xhd2015"
	req.FindGitHubRepo = "myproject"
	return nil
}
```