# Scenario

**Feature**: `--list-remotes` appends origin column on lines output

```
git remote origin -> line: path\tmain\torigin:xhd2015/lifelog@github.com
```

## Steps

1. Init repo with `origin` remote pointing at GitHub SSH URL.
2. Set `req.Args` to `["--root", <workspace>, "--list-remotes"]`.

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
	repoDir := filepath.Join(root, "proj")
	gitInitRepo(t, repoDir)
	gitRemoteAdd(t, repoDir, "origin", "git@github.com:xhd2015/lifelog.git")
	req.Args = []string{"--root", root, "--list-remotes"}
	return nil
}
```