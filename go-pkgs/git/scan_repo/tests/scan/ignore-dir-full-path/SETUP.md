# Scenario

**Feature**: `Options.IgnoreDirs` matches normalized full directory paths only

```
# exact path ignore — not basename wildcard
caller IgnoreDirs=[abs/scratch] -> Walk skips scratch subtree -> no repos
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Set `req.IgnoreDirs` to the absolute path of `scratch`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	scratch := filepath.Join(root, "scratch")
	hidden := filepath.Join(scratch, "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Roots = []string{root}
	req.IgnoreDirs = []string{absPath(t, scratch)}
	return nil
}
```