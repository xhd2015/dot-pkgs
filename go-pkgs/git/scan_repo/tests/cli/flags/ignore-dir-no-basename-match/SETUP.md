# Scenario

**Feature**: `--ignore-dir` does not treat a relative path as a basename wildcard

```
# basename-only value does not match full-path ignore set
caller --ignore-dir scratch -> hidden-repo still discovered
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Pass `--ignore-dir scratch` (not an absolute path to `scratch`).

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