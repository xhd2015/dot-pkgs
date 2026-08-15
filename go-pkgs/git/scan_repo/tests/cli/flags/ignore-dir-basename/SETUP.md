# Scenario

**Feature**: `--ignore-dir-basename` skips directories by basename anywhere in the tree

```
scratch/hidden-repo/.git ignored via --ignore-dir-basename scratch -> empty stdout
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Set `req.Args` with `--root` and `--ignore-dir-basename scratch`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	hidden := filepath.Join(root, "scratch", "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Args = []string{"--root", root, "--ignore-dir-basename", "scratch"}
	return nil
}
```