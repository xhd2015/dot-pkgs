# Scenario

**Feature**: `--ignore-dir` adds custom directory to ignore set

```
scratch/hidden-repo/.git ignored via --ignore-dir scratch -> empty stdout
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Set `req.Args` with `--root` and `--ignore-dir scratch`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	hidden := filepath.Join(root, "scratch", "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Args = []string{"--root", root, "--ignore-dir", "scratch"}
	return nil
}
```